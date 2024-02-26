package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"

	"github.com/docker/docker/api/types/container"

	commoncmd "github.com/KYVENetwork/kyve-rdk/common/goutils/cmd"

	pooltypes "github.com/KYVENetwork/chain/x/pool/types"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/chain"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/config"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/utils"

	"github.com/KYVENetwork/kyve-rdk/common/goutils/docker"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-version"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const (
	// globalCleanupLabel labels all containers and images created by kysor. It can be used to remove all kysor containers and images
	globalCleanupLabel = "kysor-all"
	protocolPath       = "protocol/core"
	runtimePath        = "runtime"
)

type Runtime struct {
	RuntimeVersion  string
	ProtocolVersion string
	RepoDir         string
}

// getHigherVersion returns the higher version of the two given versions or nil if the old version is higher
// If constraints are given, the new version must match them
func getHigherVersion(old *kyveRef, ref *plumbing.Reference, path string, constraints version.Constraints) *kyveRef {
	var oldVersion *version.Version
	if old != nil {
		oldVersion = old.ver
	}
	split := strings.Split(ref.Name().Short(), path)
	if len(split) == 2 {
		newVersion, err := version.NewVersion(split[1])
		if err != nil {
			// Ignore invalid versions
			return old
		}
		if newVersion.Prerelease() != "" {
			// Ignore prerelease versions
			return old
		}
		if oldVersion != nil && newVersion.LessThan(oldVersion) {
			// Ignore lower versions
			return old
		}
		if constraints != nil && !constraints.Check(newVersion) {
			// Ignore versions which don't match the constraints
			return old
		}
		return &kyveRef{
			ver: newVersion,
			ref: ref,
		}
	}
	return old
}

type kyveRef struct {
	ver  *version.Version
	ref  *plumbing.Reference
	path string
	name string
}

// getIntegrationVersions returns the required protocol and runtime versions for the given pool
// protocol version: Latest patch version that is defined on-chain (ex: v1.1.0 -> v1.1.3)
// runtime version: Latest version (no constraints) -> TODO: save constraints on-chain and use them
func getIntegrationVersions(repo *git.Repository, pool *pooltypes.Pool, repoDir string, wantedProtocolVers *version.Version, wantedRuntimeVers *version.Version) (*kyveRef, *kyveRef, error) {
	tagrefs, err := repo.Tags()
	if err != nil {
		return nil, nil, err
	}

	expectedRuntime := pool.Runtime
	split := strings.Split(expectedRuntime, "@kyvejs/")
	if len(split) != 2 {
		return nil, nil, fmt.Errorf("invalid runtime name: %s", expectedRuntime)
	}
	expectedRuntimeDir := split[1]

	// TODO: How should we name the runtime?
	expectedRuntime = fmt.Sprintf("@kyvejs/integration/%s", expectedRuntimeDir)

	pVersion, err := version.NewVersion(pool.Protocol.Version)
	if err != nil {
		return nil, nil, err
	}

	// Protocol must be at least the same major and minor version as defined in the pool
	protocolVersContraint, err := version.NewConstraint(fmt.Sprintf(">=%s, < %d.%d.0", pVersion.String(), pVersion.Segments()[0], pVersion.Segments()[1]+1))
	if err != nil {
		return nil, nil, err
	}

	var latestRuntimeVersion *kyveRef
	var latestProtocolVersion *kyveRef
	err = tagrefs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsTag() && strings.HasPrefix(ref.Name().Short(), "@kyvejs/protocol@") {
			if wantedProtocolVers != nil {
				if ref.Name().Short() == fmt.Sprintf("@kyvejs/protocol@%s", wantedProtocolVers.String()) && ref.Target().IsTag() {
					latestProtocolVersion = &kyveRef{
						ver: wantedProtocolVers,
						ref: ref,
					}
				}
			} else {
				latestProtocolVersion = getHigherVersion(latestProtocolVersion, ref, "@kyvejs/protocol@", protocolVersContraint)
			}
		} else if ref.Name().IsTag() && strings.HasPrefix(ref.Name().Short(), expectedRuntime) {
			if wantedRuntimeVers != nil {
				if ref.Name().Short() == fmt.Sprintf("%s@%s", expectedRuntime, wantedRuntimeVers.String()) && ref.Target().IsTag() {
					latestRuntimeVersion = &kyveRef{
						ver: wantedRuntimeVers,
						ref: ref,
					}
				}
			} else {
				latestRuntimeVersion = getHigherVersion(latestRuntimeVersion, ref, fmt.Sprintf("%s@", expectedRuntime), nil)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if latestProtocolVersion == nil {
		if wantedProtocolVers != nil {
			return nil, nil, fmt.Errorf("no protocol found for kyvejs/protocol@%s", wantedProtocolVers)
		}
		return nil, nil, fmt.Errorf("no protocol found for kyvejs/protocol@")
	}
	if latestRuntimeVersion == nil {
		if wantedRuntimeVers != nil {
			return nil, nil, fmt.Errorf("no runtime found for %s@%s", expectedRuntime, wantedRuntimeVers)
		}
		return nil, nil, fmt.Errorf("no runtime found for %s", expectedRuntime)
	}

	latestProtocolVersion.path = filepath.Join(repoDir, protocolPath)
	latestRuntimeVersion.path = filepath.Join(repoDir, runtimePath, expectedRuntimeDir)
	latestProtocolVersion.name = "protocol"
	latestRuntimeVersion.name = fmt.Sprintf("runtime-%s", expectedRuntimeDir)

	return latestProtocolVersion, latestRuntimeVersion, nil
}

type kyveRepo struct {
	name string
	dir  string
	repo *git.Repository
}

// getMainBranch returns the main branch of the given repository
func getMainBranch(repo *git.Repository) (*plumbing.Reference, error) {
	var main *plumbing.Reference
	refs, _ := repo.References()
	err := refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference {
			if ref.Name().Short() == "main" {
				main = ref
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get main branch: %v", err)
	}
	if main == nil {
		return nil, fmt.Errorf("no main branch found")
	}
	return main, nil
}

// pullRepo clones or pulls the kyvejs repository
func pullRepo(repoDir string, silent bool) (*kyveRepo, error) {
	//TODO: change this branch
	repoName := "github.com/shifty11/kyvejs"
	repoUrl := fmt.Sprintf("https://%s.git", repoName)

	var repo *git.Repository
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		// Clone the given repository to the given directory
		if !silent {
			fmt.Printf("📥  Cloning %s\n", repoUrl)
		}
		repo, err = git.PlainClone(repoDir, false, &git.CloneOptions{
			URL:      repoUrl,
			Progress: os.Stdout,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Otherwise open the existing repository
		repo, err = git.PlainOpen(repoDir)
		if err != nil {
			return nil, err
		}

		// Get the main branch
		main, err := getMainBranch(repo)
		if err != nil {
			return nil, err
		}

		w, err := repo.Worktree()
		if err != nil {
			return nil, err
		}

		// Reset the worktree to the latest commit, discarding any local changes
		// If we don't do this, the pull will fail if there are local changes
		err = w.Reset(&git.ResetOptions{Commit: main.Hash(), Mode: git.HardReset})
		if err != nil {
			return nil, fmt.Errorf("failed to reset worktree: %v", err)
		}

		// Pull the latest changes
		if !silent {
			fmt.Println("⬇️   Pulling latest changes")
		}
		err = w.Pull(&git.PullOptions{ReferenceName: main.Name(), Force: true})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) && !errors.Is(err, git.ErrNonFastForwardUpdate) {
			return nil, fmt.Errorf("failed to pull latest changes: %v", err)
		}
	}

	return &kyveRepo{
		repo: repo,
		name: repoName,
		dir:  repoDir,
	}, nil
}

func buildImage(worktree *git.Worktree, ref *plumbing.Reference, cli *client.Client, image docker.Image, verbose bool) error {
	fmt.Printf("📦  Checkout %s\n", ref.Name().Short())
	err := worktree.Checkout(&git.CheckoutOptions{
		Branch: ref.Name(),
		Force:  true,
	})
	if err != nil {
		return err
	}

	showOnlyProgress := true
	var printFn func(string)
	if verbose {
		showOnlyProgress = false
		printFn = func(text string) {
			fmt.Print(text)
		}
	}

	fmt.Printf("🏗️   Building %s ...\n", image.Tags[0])
	return docker.BuildImage(context.Background(), cli, image, docker.OutputOptions{ShowOnlyProgress: showOnlyProgress, PrintFn: printFn})
}

// buildImages builds the protocol and runtime images
func buildImages(kr *kyveRepo, cli *client.Client, pool *pooltypes.Pool, label string, protocolVersion *version.Version, runtimeVersion *version.Version, verbose bool) (*docker.Image, *docker.Image, error) {
	w, err := kr.repo.Worktree()
	if err != nil {
		return nil, nil, err
	}

	protocol, runtime, err := getIntegrationVersions(kr.repo, pool, kr.dir, protocolVersion, runtimeVersion)
	if err != nil {
		return nil, nil, err
	}

	protocolImage := docker.Image{
		Path:   protocol.path,
		Tags:   []string{fmt.Sprintf("%s/%s:%s", strings.ToLower(kr.name), protocol.name, protocol.ver.String())},
		Labels: map[string]string{globalCleanupLabel: "", label: ""},
	}
	runtimeImage := docker.Image{
		Path:   runtime.path,
		Tags:   []string{fmt.Sprintf("%s/%s:%s", strings.ToLower(kr.name), runtime.name, runtime.ver.String())},
		Labels: map[string]string{globalCleanupLabel: "", label: ""},
	}

	err = buildImage(w, protocol.ref, cli, protocolImage, verbose)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("🏗️   Finished bulding image: " + protocolImage.Tags[0])

	err = buildImage(w, runtime.ref, cli, runtimeImage, verbose)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("🏗️   Finished bulding image " + runtimeImage.Tags[0])
	return &protocolImage, &runtimeImage, nil
}

type StartResult struct {
	Name string
	ID   string
}

// startContainers starts the protocol and runtime containers
func startContainers(cli *client.Client, valConfig config.ValaccountConfig, pool *pooltypes.Pool, debug bool, protocol *docker.Image, runtime *docker.Image, label string, runtimeEnv []string) (*StartResult, *StartResult, error) {
	protocolName := fmt.Sprintf("%s-%s", label, protocol.TagsLastPartWithoutVersion()[0])
	runtimeName := fmt.Sprintf("%s-%s", label, runtime.TagsLastPartWithoutVersion()[0])

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	env, err := docker.CreateProtocolEnv(docker.ProtocolEnv{
		Valaccount:  valConfig.Valaccount,
		RpcAddress:  config.GetConfigX().RPC,
		RestAddress: config.GetConfigX().REST,
		Host:        runtimeName,
		PoolId:      pool.Id,
		Debug:       debug,
		ChainId:     config.GetConfigX().ChainID,
	})
	if err != nil {
		return nil, nil, err
	}

	err = docker.CreateNetwork(ctx, cli, docker.NetworkConfig{
		Name:   label,
		Labels: map[string]string{globalCleanupLabel: "", label: ""},
	})
	if err != nil {
		return nil, nil, err
	}

	pConfig := docker.ContainerConfig{
		Image:   protocol.Tags[0],
		Name:    protocolName,
		Network: label,
		Env:     env,
		Labels:  map[string]string{globalCleanupLabel: "", label: ""},
	}

	rConfig := docker.ContainerConfig{
		Image:      runtime.Tags[0],
		Name:       runtimeName,
		Network:    label,
		Env:        runtimeEnv,
		Labels:     map[string]string{globalCleanupLabel: "", label: ""},
		ExtraHosts: []string{"host.docker.internal:host-gateway"},
	}

	protocolId, err := docker.StartContainer(ctx, cli, pConfig)
	if err != nil {
		return nil, nil, err
	}
	fmt.Print("🚀  Started container ")
	utils.PrintlnItalic(protocolName)
	protocolResult := &StartResult{
		Name: protocolName,
		ID:   protocolId,
	}

	runtimeId, err := docker.StartContainer(ctx, cli, rConfig)
	if err != nil {
		return nil, nil, err
	}
	fmt.Print("🚀  Started container ")
	utils.PrintlnItalic(runtimeName)
	runtimeResult := &StartResult{
		Name: runtimeName,
		ID:   runtimeId,
	}

	return protocolResult, runtimeResult, nil
}

func getRuntimeEnv(cmd *cobra.Command) ([]string, error) {
	var env []string
	envFile, err := commoncmd.GetStringFromPromptOrFlag(cmd, flagStartEnvFile)
	if err != nil {
		return nil, err
	}
	if envFile != "" {
		path, err := homedir.Expand(envFile)
		if err != nil {
			return nil, err
		}
		k := koanf.New(".")
		if err := k.Load(file.Provider(path), dotenv.Parser()); err != nil {
			return nil, fmt.Errorf("failed to load env file: %v", err)
		}
		for key, value := range k.All() {
			env = append(env, fmt.Sprintf("%s=%v", key, value))
		}
	}
	return env, nil
}

// printLogs prints the logs of the given container (stdout and stderr)
// Errors are sent to the errChan and the name of the container is sent to the endChan when the logs end
// This function is blocking
func printLogs(ctx context.Context, cli *client.Client, cont *StartResult, colorAttr color.Attribute, errChan chan error) {
	logs, err := cli.ContainerLogs(context.Background(), cont.ID,
		container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true, Details: false})
	if err != nil {
		errChan <- err
		return
	}

	reader := bufio.NewReader(logs)
	for {
		// Discard the 8-byte header
		_, err := reader.Discard(8)
		if err != nil {
			if err == io.EOF {
				break
			}
			errChan <- err
			return
		}

		// Read one line
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			errChan <- err
			return
		}

		// Print the line
		color.Set(colorAttr)
		fmt.Printf("%s: ", cont.Name)
		color.Unset()
		fmt.Print(line)

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}

	select {
	case <-ctx.Done():
		return
	default:
		// If the context has not been canceled, the logs ended unexpectedly (which means the container died)
		errChan <- fmt.Errorf("container %s stopped unexpectedly (ID: %s)", cont.Name, cont.ID)
	}
}

// start (or restart) the protocol and runtime containers
func start(
	ctx context.Context,
	cmd *cobra.Command,
	kyveClient *chain.KyveClient,
	cli *client.Client,
	valConfig config.ValaccountConfig,
	runtimeEnv []string,
	protocolVersion *version.Version,
	runtimeVersion *version.Version,
	debug bool,
	detached bool,
	errChan chan error,
	newVersionChan chan interface{},
) (string, error) {
	response, err := kyveClient.QueryPool(valConfig.Pool)
	if err != nil {
		return "", fmt.Errorf("failed to query pool: %v", err)
	}
	pool := response.GetPool().Data

	if detached {
		fmt.Println("    Starting KYSOR (detached)...")
		fmt.Println("    Auto update during runtime is disabled in detached mode!")
	} else {
		fmt.Println("    Starting KYSOR...")
	}
	fmt.Printf("    Running on platform and architecture: %s - %s\n\n", goruntime.GOOS, goruntime.GOARCH)

	homeDir, err := config.GetHomeDir(cmd)
	if err != nil {
		return "", err
	}

	// Clone or pull the kyvejs repository
	repoDir := filepath.Join(homeDir, "kyvejs")
	repo, err := pullRepo(repoDir, false)
	if err != nil {
		return "", err
	}

	// Build images
	label := valConfig.GetContainerLabel()
	protocol, runtime, err := buildImages(repo, cli, pool, label, protocolVersion, runtimeVersion, debug)
	if err != nil {
		return "", fmt.Errorf("failed to build images: %v", err)
	}

	// Stop and remove existing containers
	err = tearDownContainers(cli, label)
	if err != nil {
		return "", err
	}

	// Start containers
	protocolContainer, runtimeContainer, err := startContainers(cli, valConfig, pool, debug, protocol, runtime, label, runtimeEnv)
	if err != nil {
		return "", err
	}

	if detached {
		fmt.Println()
		fmt.Println("🔍  Use following commands to view the logs:")
		fmt.Print("    ")
		utils.PrintlnItalic(fmt.Sprintf("docker logs -f %s", runtimeContainer.Name))
		fmt.Print("    ")
		utils.PrintlnItalic(fmt.Sprintf("docker logs -f %s", protocolContainer.Name))
	} else {
		// Print protocol logs
		go printLogs(ctx, cli, protocolContainer, color.FgGreen, errChan)

		// Print runtime logs
		go printLogs(ctx, cli, runtimeContainer, color.FgBlue, errChan)

		// Check for new versions only if versions are not pinned
		if protocolVersion == nil && runtimeVersion == nil {
			fmt.Println("🔄  Auto update of docker container's is enabled")
			go checkNewVersion(ctx, kyveClient, valConfig.Pool, repo, newVersionChan)
		} else {
			fmt.Println("🔄  Auto update of docker container's is disabled")
		}
		fmt.Println()
	}
	return label, nil
}

// checkNewVersion checks if a new version is available and sends a signal to the newVersionChan if it is
// It also updates the local repository and pulls the latest changes
// This function is blocking
func checkNewVersion(ctx context.Context, kyveClient *chain.KyveClient, poolId uint64, kr *kyveRepo, newVersionChan chan interface{}) {
	var currentProtocol, currentRuntime *version.Version
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		_, err := pullRepo(kr.dir, true)
		if err != nil {
			fmt.Println("failed to update repository: ", err)
			continue
		}

		response, err := kyveClient.QueryPool(poolId)
		if err != nil {
			fmt.Printf("failed to query pool: %v\n", err)
			continue
		}

		protocolRef, runtimeRef, err := getIntegrationVersions(kr.repo, response.GetPool().Data, kr.dir, nil, nil)
		if err != nil {
			fmt.Println("failed to get runtime versions: ", err)
			continue
		}
		if currentProtocol == nil {
			currentProtocol = protocolRef.ver
		}
		if currentRuntime == nil {
			currentRuntime = runtimeRef.ver
		}

		if protocolRef.ver.String() != currentProtocol.String() || runtimeRef.ver.String() != currentRuntime.String() {
			newVersionChan <- nil
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Continue the loop
		}
	}
}

func validateVersion(s string) error {
	if s == "" {
		return nil
	}
	_, err := version.NewVersion(s)
	return err
}

var (
	flagStartValaccount = commoncmd.OptionFlag[config.ValaccountConfig]{
		Name:             "valaccount",
		Short:            "v",
		Usage:            "Name of the valaccount to run",
		Required:         true,
		MaxSelectionSize: 10,
	}
	flagStartEnvFile = commoncmd.StringFlag{
		Name:       "env-file",
		Short:      "e",
		Usage:      "Specify the path to an .env file which should be used when starting a binary",
		Required:   false,
		ValidateFn: commoncmd.ValidatePathExistsOrEmpty,
	}
	flagStartProtocolVersion = commoncmd.StringFlag{
		Name:       "protocol-version",
		Usage:      "Specify the version of the protocol to run",
		Prompt:     "Specify the version of the protocol to run (leave empty for latest)",
		Required:   false,
		ValidateFn: validateVersion,
	}
	flagStartRuntimeVersion = commoncmd.StringFlag{
		Name:       "runtime-version",
		Usage:      "Specify the version of the runtime to run",
		Prompt:     "Specify the version of the runtime to run (leave empty for latest)",
		Required:   false,
		ValidateFn: validateVersion,
	}
	flagStartDebug = commoncmd.BoolFlag{
		Name:         "debug",
		Short:        "",
		Usage:        "Run the validator node in debug mode",
		DefaultValue: false,
	}
	flagStartDetached = commoncmd.BoolFlag{
		Name:         "detached",
		Short:        "d",
		Usage:        "Run the validator node in detached mode (no auto update)",
		DefaultValue: false,
	}
)

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Start data validator",
		PreRunE: commoncmd.CombineFuncs(utils.CheckDockerInstalled, config.LoadConfigs, commoncmd.SetupInteractiveMode),
		RunE: func(cmd *cobra.Command, args []string) error {
			kyveClient, err := chain.NewKyveClient(config.GetConfigX(), config.ValaccountConfigs)
			if err != nil {
				return err
			}

			// Return if no valaccount exists
			flagStartValaccount.Options = config.ValaccountConfigOptions
			if len(flagStartValaccount.Options) == 0 {
				fmt.Println("No valaccount found. Create one with 'kysor valaccounts create'")
				return nil
			}

			// Valaccount config
			valaccOption, err := commoncmd.GetOptionFromPromptOrFlag(cmd, flagStartValaccount)
			if err != nil {
				return err
			}
			valConfig := valaccOption.Value()

			// Runtime env
			runtimeEnv, err := getRuntimeEnv(cmd)
			if err != nil {
				return err
			}

			// Protocol version
			var protocolVersion *version.Version
			protocolVersionStr, err := commoncmd.GetStringFromPromptOrFlag(cmd, flagStartProtocolVersion)
			if err != nil {
				return err
			}
			if protocolVersionStr != "" {
				protocolVersion, err = version.NewVersion(protocolVersionStr)
				if err != nil {
					return err
				}
			}

			// Runtime version
			var runtimeVersion *version.Version
			runtimeVersionStr, err := commoncmd.GetStringFromPromptOrFlag(cmd, flagStartRuntimeVersion)
			if err != nil {
				return err
			}
			if runtimeVersionStr != "" {
				runtimeVersion, err = version.NewVersion(runtimeVersionStr)
				if err != nil {
					return err
				}
			}

			// Debug
			debug, err := commoncmd.GetBoolFromPromptOrFlag(cmd, flagStartDebug)
			if err != nil {
				return err
			}

			// Detached
			detached, err := commoncmd.GetBoolFromPromptOrFlag(cmd, flagStartDetached)
			if err != nil {
				return err
			}

			cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
			if err != nil {
				return fmt.Errorf("failed to create docker client: %v", err)
			}
			//goland:noinspection GoUnhandledErrorResult
			defer cli.Close()

			errChan := make(chan error)              // async error channel
			newVersionChan := make(chan interface{}) // new version is available

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Detached 	-> start containers and forget about them
			// Not detached -> listen to signals and stop containers on signal
			//              -> listen to new version and restart containers on new version
			//   			-> listen to log end and throw error if log ends unexpectedly (which means the container died)
			label, err := start(
				ctx,
				cmd,
				kyveClient,
				cli,
				valConfig,
				runtimeEnv,
				protocolVersion,
				runtimeVersion,
				debug,
				detached,
				errChan,
				newVersionChan,
			)
			if err != nil {
				return err
			}
			if !detached {
				sigc := make(chan os.Signal, 1)
				signal.Notify(sigc,
					syscall.SIGHUP,
					syscall.SIGINT,
					syscall.SIGTERM,
					syscall.SIGQUIT,
				)

				// Cleanup containers on exit
				defer func() {
					cancel()

					// Cleanup containers
					if err := tearDownContainers(cli, label); err != nil {
						fmt.Printf("failed to stop containers: %v\n", err)
					}
				}()

				// Enter loop
				for {
					select {
					case <-sigc:
						// Stop signal received, stop containers
						fmt.Println("\n🛑  Stopping KYSOR...")
						return nil
					case <-newVersionChan:
						// New version available, restart containers
						fmt.Println("🔄  New version available, restarting KYSOR...")

						cancel()
						newCtx, newCancel := context.WithCancel(context.Background())
						cancel = newCancel

						label, err = start(
							newCtx,
							cmd,
							kyveClient,
							cli,
							valConfig,
							runtimeEnv,
							protocolVersion,
							runtimeVersion,
							debug,
							detached,
							errChan,
							newVersionChan,
						)
						if err != nil {
							return err
						}
					case err := <-errChan:
						// Error received, throw error
						if err != nil {
							return err
						}
					}
				}
			}
			return nil
		},
	}
	commoncmd.AddOptionFlags(cmd, []commoncmd.OptionFlag[config.ValaccountConfig]{flagStartValaccount})
	commoncmd.AddStringFlags(cmd, []commoncmd.StringFlag{flagStartEnvFile, flagStartProtocolVersion, flagStartRuntimeVersion})
	commoncmd.AddBoolFlags(cmd, []commoncmd.BoolFlag{flagStartDebug, flagStartDetached})
	return cmd
}

func init() {
	rootCmd.AddCommand(startCmd())
}

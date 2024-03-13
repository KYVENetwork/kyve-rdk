package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/types"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/docker/go-connections/nat"

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
	// globalContainerLabel labels all containers and images created by kysor.
	// It can be used to address all containers and images created by kysor.
	globalContainerLabel = "kysor-all"
	protocolPath         = "protocol/core"
	runtimePath          = "runtime"
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

	protocolPrefix := "protocol/core@"

	// TODO: after chain-upgrade 1.5.0, remove this and get the runtime from the pool
	split := strings.Split(pool.Runtime, "@kyvejs/")
	if len(split) != 2 {
		return nil, nil, fmt.Errorf("invalid runtime name: %s", pool.Runtime)
	}
	expectedRuntimeDir := split[1]
	runtimePrefix := fmt.Sprintf("runtime/%s@", expectedRuntimeDir)

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
		if ref.Name().IsTag() && strings.HasPrefix(ref.Name().Short(), protocolPrefix) {
			if wantedProtocolVers != nil {
				if ref.Name().Short() == fmt.Sprintf("%s%s", protocolPrefix, wantedProtocolVers.String()) {
					latestProtocolVersion = &kyveRef{
						ver: wantedProtocolVers,
						ref: ref,
					}
				}
			} else {
				latestProtocolVersion = getHigherVersion(latestProtocolVersion, ref, protocolPrefix, protocolVersContraint)
			}
		} else if ref.Name().IsTag() && strings.HasPrefix(ref.Name().Short(), runtimePrefix) {
			if wantedRuntimeVers != nil {
				if ref.Name().Short() == fmt.Sprintf("%s%s", runtimePrefix, wantedRuntimeVers.String()) {
					latestRuntimeVersion = &kyveRef{
						ver: wantedRuntimeVers,
						ref: ref,
					}
				}
			} else {
				latestRuntimeVersion = getHigherVersion(latestRuntimeVersion, ref, runtimePrefix, nil)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if latestProtocolVersion == nil {
		if wantedProtocolVers != nil {
			return nil, nil, fmt.Errorf("no protocol found for %s%s", protocolPrefix, wantedProtocolVers)
		}
		return nil, nil, fmt.Errorf("no protocol found for %s", protocolPrefix)
	}
	if latestRuntimeVersion == nil {
		if wantedRuntimeVers != nil {
			return nil, nil, fmt.Errorf("no runtime found for %s%s", runtimePrefix, wantedRuntimeVers)
		}
		return nil, nil, fmt.Errorf("no runtime found for %s", runtimePrefix)
	}

	latestProtocolVersion.path = filepath.Join(repoDir, protocolPath)
	latestRuntimeVersion.path = filepath.Join(repoDir, runtimePath, expectedRuntimeDir)
	latestProtocolVersion.name = "protocol-core"
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

// pullRepo clones or pulls the kyve-rdk repository
func pullRepo(repoDir string, silent bool) (*kyveRepo, error) {
	var repo *git.Repository
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		// Clone the given repository to the given directory
		if !silent {
			fmt.Printf("📥  Cloning %s\n", types.RepoUrl)
		}
		repo, err = git.PlainClone(repoDir, false, &git.CloneOptions{
			URL:      types.RepoUrl,
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
			return nil, fmt.Errorf("failed to reset worktree: %v\nTry to delete the repo folder '%s'", err, repoDir)
		}

		// Pull the latest changes
		if !silent {
			fmt.Println("⬇️   Pulling latest changes")
		}
		err = w.Pull(&git.PullOptions{ReferenceName: main.Name(), Force: true})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) && !errors.Is(err, git.ErrNonFastForwardUpdate) {
			return nil, fmt.Errorf("failed to pull latest changes: %v\nTry to delete the repo folder '%s'", err, repoDir)
		}
	}

	return &kyveRepo{
		repo: repo,
		name: types.RepoName,
		dir:  repoDir,
	}, nil
}

func buildImage(worktree *git.Worktree, ref *plumbing.Reference, cli *client.Client, image docker.Image, verbose bool) error {
	if ref != nil {
		fmt.Printf("📦  Checkout %s\n", ref.Name().Short())
		err := worktree.Checkout(&git.CheckoutOptions{
			Branch: ref.Name(),
			Force:  true,
		})
		if err != nil {
			return err
		}
	}

	showOnlyProgress := true
	var printFn func(string)
	if verbose {
		showOnlyProgress = false
		printFn = func(text string) {
			fmt.Print(text)
		}
	}

	fmt.Printf("🐳  Building %s ...\n", image.Tags[0])
	err := docker.BuildImage(context.Background(), cli, image, docker.OutputOptions{ShowOnlyProgress: showOnlyProgress, PrintFn: printFn})
	if err == nil {
		fmt.Printf("✅  Finished bulding image: %s\n", image.Tags[0])
	}
	return err
}

// buildImages builds the protocol and runtime images
func buildImages(
	kr *kyveRepo,
	cli *client.Client,
	pool *pooltypes.Pool,
	label string,
	options AdvancedOptions,
) (*docker.Image, *docker.Image, error) {
	w, err := kr.repo.Worktree()
	if err != nil {
		return nil, nil, err
	}

	var protocolImage docker.Image
	var protocolRef *plumbing.Reference
	var runtimeImage docker.Image
	var runtimeRef *plumbing.Reference

	// TODO: split runtime and protocol into separate functions
	protocol, runtime, err := getIntegrationVersions(kr.repo, pool, kr.dir, options.ProtocolVersion, options.RuntimeVersion)
	if err != nil {
		return nil, nil, err
	}

	if options.ProtocolBuildDir != "" {
		// If protocolBuildDir is set, use it as the build directory
		vers := "0.0.0-local"
		protocolImage = docker.Image{
			Path:      options.ProtocolBuildDir,
			Tags:      []string{fmt.Sprintf("%s/%s:%s", strings.ToLower(kr.name), protocol.name, "local")},
			Labels:    map[string]string{globalContainerLabel: "", label: ""},
			BuildArgs: map[string]*string{"VERSION": &vers},
		}
	} else {
		// Otherwise, use the version from the repository
		protocolRef = protocol.ref
		vers := protocol.ver.String()
		protocolImage = docker.Image{
			Path:      protocol.path,
			Tags:      []string{fmt.Sprintf("%s/%s:%s", strings.ToLower(kr.name), protocol.name, protocol.ver.String())},
			Labels:    map[string]string{globalContainerLabel: "", label: ""},
			BuildArgs: map[string]*string{"VERSION": &vers},
		}
	}

	if options.RuntimeBuildDir != "" {
		// If runtimeBuildDir is set, use it as the build directory
		vers := "0.0.0-local"
		runtimeImage = docker.Image{
			Path:      options.RuntimeBuildDir,
			Tags:      []string{fmt.Sprintf("%s/%s:%s", strings.ToLower(kr.name), runtime.name, "local")},
			Labels:    map[string]string{globalContainerLabel: "", label: ""},
			BuildArgs: map[string]*string{"VERSION": &vers},
		}
	} else {
		// Otherwise, use the version from the repository
		runtimeRef = runtime.ref
		vers := runtime.ver.String()
		runtimeImage = docker.Image{
			Path:      runtime.path,
			Tags:      []string{fmt.Sprintf("%s/%s:%s", strings.ToLower(kr.name), runtime.name, runtime.ver.String())},
			Labels:    map[string]string{globalContainerLabel: "", label: ""},
			BuildArgs: map[string]*string{"VERSION": &vers},
		}
	}

	err = buildImage(w, protocolRef, cli, protocolImage, options.Debug)
	if err != nil {
		return nil, nil, err
	}

	err = buildImage(w, runtimeRef, cli, runtimeImage, options.Debug)
	if err != nil {
		return nil, nil, err
	}
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

	rpc, err := config.GetConfigX().GetWorkingRPC()
	if err != nil {
		return nil, nil, err
	}

	rest, err := config.GetConfigX().GetWorkingREST()
	if err != nil {
		return nil, nil, err
	}

	env, err := docker.CreateProtocolEnv(docker.ProtocolEnv{
		Valaccount:  valConfig.Valaccount,
		RpcAddress:  rpc,
		RestAddress: rest,
		Host:        runtimeName,
		PoolId:      pool.Id,
		Debug:       debug,
		ChainId:     config.GetConfigX().ChainID,
		Metrics:     valConfig.Metrics,
		MetricsPort: valConfig.MetricsPort,
	})
	if err != nil {
		return nil, nil, err
	}

	err = docker.CreateNetwork(ctx, cli, docker.NetworkConfig{
		Name:   label,
		Labels: map[string]string{globalContainerLabel: "", label: ""},
	})
	if err != nil {
		return nil, nil, err
	}

	var exposedPorts nat.PortSet
	if valConfig.Metrics {
		port, err := nat.NewPort("tcp", strconv.FormatUint(valConfig.MetricsPort, 10))
		if err != nil {
			return nil, nil, err
		}
		exposedPorts = nat.PortSet{port: {}}
	}

	pConfig := docker.ContainerConfig{
		Image:        protocol.Tags[0],
		Name:         protocolName,
		Network:      label,
		Env:          env,
		Labels:       map[string]string{globalContainerLabel: "", label: ""},
		ExposedPorts: exposedPorts,
	}

	rConfig := docker.ContainerConfig{
		Image:      runtime.Tags[0],
		Name:       runtimeName,
		Network:    label,
		Env:        runtimeEnv,
		Labels:     map[string]string{globalContainerLabel: "", label: ""},
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
	options AdvancedOptions,
	errChan chan error,
	newVersionChan chan interface{},
) (string, error) {
	response, err := kyveClient.QueryPool(valConfig.Pool)
	if err != nil {
		return "", fmt.Errorf("failed to query pool: %v", err)
	}
	pool := response.GetPool().Data

	if options.Detached {
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

	// Clone or pull the kyve-rdk repository
	repoDir := filepath.Join(homeDir, "kyve-rdk")
	repo, err := pullRepo(repoDir, false)
	if err != nil {
		return "", err
	}

	// Build images
	label := valConfig.GetContainerLabel()
	protocol, runtime, err := buildImages(repo, cli, pool, label, options)
	if err != nil {
		return "", fmt.Errorf("failed to build images: %v", err)
	}

	// Stop and remove existing containers
	err = tearDownContainers(cli, label)
	if err != nil {
		return "", err
	}

	// Start containers
	protocolContainer, runtimeContainer, err := startContainers(cli, valConfig, pool, options.Debug, protocol, runtime, label, runtimeEnv)
	if err != nil {
		return "", err
	}

	if options.Detached {
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

		// If protocol and runtime are custom, there is no need to check for new versions
		if options.HasCustomProtocol() && options.HasCustomRuntime() {
			fmt.Println("🔄  Auto update of docker containers are disabled")
		} else {
			fmt.Println("🔄  Auto update of docker containers are enabled")
			go checkNewVersion(ctx, kyveClient, valConfig.Pool, repo, newVersionChan)
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

func validateVersionOrEmpty(s string) error {
	if s == "" {
		return nil
	}
	_, err := version.NewVersion(s)
	return err
}

var (
	flagStartValaccount = commoncmd.OptionFlag[config.ValaccountConfig]{
		Name:              "valaccount",
		Short:             "v",
		Usage:             "Name of the valaccount to run",
		Required:          true,
		MaxSelectionSize:  10,
		StartInSearchMode: true,
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
		Usage:      "Specify the protocol version",
		Prompt:     "Specify the protocol version (leave empty for latest)",
		Required:   false,
		ValidateFn: validateVersionOrEmpty,
	}
	flagStartProtocolBuildDir = commoncmd.StringFlag{
		Name:       "protocol-build-dir",
		Usage:      "Specify the directory to build the protocol image from",
		Prompt:     "Specify the directory to build the protocol image from (leave empty to not build)",
		Required:   false,
		ValidateFn: commoncmd.ValidatePathExistsOrEmpty,
	}
	flagStartRuntimeVersion = commoncmd.StringFlag{
		Name:       "runtime-version",
		Usage:      "Specify the runtime version",
		Prompt:     "Specify the runtime version (leave empty for latest)",
		Required:   false,
		ValidateFn: validateVersionOrEmpty,
	}
	flagStartRuntimeBuildDir = commoncmd.StringFlag{
		Name:       "runtime-build-dir",
		Usage:      "Specify the directory to build the runtime image from",
		Prompt:     "Specify the directory to build the runtime image from (leave empty to not build)",
		Required:   false,
		ValidateFn: commoncmd.ValidatePathExistsOrEmpty,
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

type AdvancedOptions struct {
	ProtocolVersion  *version.Version
	ProtocolBuildDir string
	RuntimeVersion   *version.Version
	RuntimeBuildDir  string
	Debug            bool
	Detached         bool
}

func (o AdvancedOptions) HasCustomProtocol() bool {
	return o.ProtocolVersion != nil || o.ProtocolBuildDir != ""
}

func (o AdvancedOptions) HasCustomRuntime() bool {
	return o.RuntimeVersion != nil || o.RuntimeBuildDir != ""
}

func getAdvanceOptions(cmd *cobra.Command) (options AdvancedOptions, err error) {
	// Prompt to show advanced options
	showAdvanced := false
	if commoncmd.IsInteractive(cmd) {
		showAdvanced, err = commoncmd.PromptYesNo("Show advanced options?", commoncmd.No)
		if err != nil {
			return options, err
		}
	}

	// Protocol version & build dir
	var protocolVersionStr string
	if showAdvanced {
		protocolVersionStr, err = commoncmd.GetStringFromPromptOrFlag(cmd, flagStartProtocolVersion)
		if err != nil {
			return options, err
		}
	} else {
		protocolVersionStr, err = commoncmd.GetStringFromFlag(cmd, flagStartProtocolVersion)
		if err != nil {
			return options, err
		}
	}
	if protocolVersionStr != "" {
		options.ProtocolVersion, err = version.NewVersion(protocolVersionStr)
		if err != nil {
			return options, err
		}
	} else {
		// Protocol build dir (only ask if protocol version is not set)
		if showAdvanced {
			options.ProtocolBuildDir, err = commoncmd.GetStringFromPromptOrFlag(cmd, flagStartProtocolBuildDir)
			if err != nil {
				return options, err
			}
		} else {
			options.ProtocolBuildDir, err = commoncmd.GetStringFromFlag(cmd, flagStartProtocolBuildDir)
			if err != nil {
				return options, err
			}
		}
	}

	// Runtime version & build dir
	var runtimeVersionStr string
	if showAdvanced {
		runtimeVersionStr, err = commoncmd.GetStringFromPromptOrFlag(cmd, flagStartRuntimeVersion)
		if err != nil {
			return options, err
		}
	} else {
		runtimeVersionStr, err = commoncmd.GetStringFromFlag(cmd, flagStartRuntimeVersion)
		if err != nil {
			return options, err
		}
	}
	if runtimeVersionStr != "" {
		options.RuntimeVersion, err = version.NewVersion(runtimeVersionStr)
		if err != nil {
			return options, err
		}
	} else {
		// Runtime build dir (only ask if runtime version is not set)
		if showAdvanced {
			options.RuntimeBuildDir, err = commoncmd.GetStringFromPromptOrFlag(cmd, flagStartRuntimeBuildDir)
			if err != nil {
				return options, err
			}
		} else {
			options.RuntimeBuildDir, err = commoncmd.GetStringFromFlag(cmd, flagStartRuntimeBuildDir)
			if err != nil {
				return options, err
			}
		}
	}

	if showAdvanced {
		// Debug
		options.Debug, err = commoncmd.GetBoolFromPromptOrFlag(cmd, flagStartDebug)
		if err != nil {
			return options, err
		}

		// Detached
		options.Detached, err = commoncmd.GetBoolFromPromptOrFlag(cmd, flagStartDetached)
		if err != nil {
			return options, err
		}
	} else {
		// Debug
		options.Debug, err = cmd.Flags().GetBool(flagStartDebug.Name)
		if err != nil {
			return options, err
		}

		// Detached
		options.Detached, err = cmd.Flags().GetBool(flagStartDetached.Name)
		if err != nil {
			return options, err
		}
	}

	return options, nil
}

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Start data validator",
		PreRunE: commoncmd.CombineFuncs(utils.CheckDockerInstalled, utils.CheckUpdateAvailable, config.LoadConfigs, commoncmd.SetupInteractiveMode),
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

			options, err := getAdvanceOptions(cmd)
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
				options,
				errChan,
				newVersionChan,
			)
			if err != nil {
				return err
			}
			if !options.Detached {
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
							options,
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
	commoncmd.AddStringFlags(cmd, []commoncmd.StringFlag{
		flagStartEnvFile,
		flagStartProtocolVersion,
		flagStartRuntimeVersion,
		flagStartProtocolBuildDir,
		flagStartRuntimeBuildDir,
	})
	commoncmd.AddBoolFlags(cmd, []commoncmd.BoolFlag{flagStartDebug, flagStartDetached})

	// Only protocol-version or protocol-build-dir can be set
	cmd.MarkFlagsMutuallyExclusive(flagStartProtocolVersion.Name, flagStartProtocolBuildDir.Name)
	// Only runtime-version or runtime-build-dir can be set
	cmd.MarkFlagsMutuallyExclusive(flagStartRuntimeVersion.Name, flagStartRuntimeBuildDir.Name)
	return cmd
}

func init() {
	rootCmd.AddCommand(startCmd())
}

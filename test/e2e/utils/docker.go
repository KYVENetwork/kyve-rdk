package utils

import (
	"context"
	"errors"
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"

	"github.com/KYVENetwork/kyve-rdk/common/goutils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

const (
	// kyveStorageName is the name of the volume used for kyve storage
	kyveStorageName = "kyvestorage"
	// kyveStorageMountProtocol is the mount path for the kyve storage volume in the protocol container
	kyveStorageMountProtocol = "/tmp/kyvestorage"
	// kyveStorageMountApi is the mount path for the kyve storage volume in the testapi container
	kyveStorageMountApi = "/app/api"
	// cleanupLabel is the label used to identify containers and volumes that should be cleaned up
	cleanupLabel = "kyve-e2e-test"
	// interchainCleanupLabel is the label used to identify containers and volumes that should be cleaned up from interchaintest
	interchainCleanupLabel = "ibc-test"
	// runtimeImagePrefix is the prefix used for the runtime image name
	runtimeImagePrefix = "runtime"
	// kystrapTemplatesDir is the path to the templates folder
	kystrapTemplatesDir = rootPath + "tools/kystrap/templates"
	// kystrapMount is the mount path for the output folder in the kystrap container
	kystrapMount = "/app/runtime"
)

var defaultVersion = "1.0.0-test"
var (
	protocolImage = docker.Image{
		Path:      rootPath + "protocol/core",
		Tags:      []string{"protocol"},
		Labels:    map[string]string{cleanupLabel: ""},
		BuildArgs: map[string]*string{"VERSION": &defaultVersion},
	}
	testapiImage = docker.Image{
		Path:   "testapi",
		Tags:   []string{"testapi"},
		Labels: map[string]string{cleanupLabel: ""},
	}
	kystrapImage = docker.Image{
		Path:   rootPath + "tools/kystrap",
		Tags:   []string{"kystrap-e2etest"},
		Labels: map[string]string{cleanupLabel: ""},
	}
)

type IntegrationBuilder struct {
	testName string
	log      *zap.Logger
}

func NewIntegrationBuilder(testName string, log *zap.Logger) *IntegrationBuilder {
	return &IntegrationBuilder{
		testName: testName,
		log:      log,
	}
}

func (pc *IntegrationBuilder) printToDebugLog(text string) {
	pc.log.Debug(text)
}

// BuildCoreAndDeps builds the protocol, testapi and kystrap images
func (pc *IntegrationBuilder) BuildCoreAndDeps() error {
	// First, cleanup any old containers and volumes
	err := pc.Cleanup()
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer cli.Close()

	// Build all the images concurrently
	configs := []docker.Image{protocolImage, testapiImage, kystrapImage}
	errChs := make([]chan error, len(configs))
	for i, img := range configs {
		errChs[i] = make(chan error)
		docker.BuildImageAsync(context.Background(), cli, img, errChs[i], docker.OutputOptions{PrintFn: pc.printToDebugLog})
	}

	for _, errCh := range errChs {
		err := <-errCh
		if err != nil {
			return err
		}
	}
	return nil
}

func (pc *IntegrationBuilder) BuildRuntimes(testConfigs []*TestConfig) error {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer cli.Close()

	var runtimeConfigs []docker.Image

	// Find all runtimes and build them
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5*time.Duration(len(testConfigs)))
	defer cancel()

	for _, cfg := range testConfigs {
		// Create the volumes that will be shared between the runtimes and testapi containers
		_, err = cli.VolumeCreate(ctx, volume.CreateOptions{
			Name:   kyveStorageVolumeName(cfg.Runtime),
			Labels: map[string]string{cleanupLabel: ""},
		})
		if err != nil {
			return err
		}
		runtimeConfigs = append(runtimeConfigs, docker.Image{
			Path:      cfg.Runtime.Path,
			Tags:      []string{runtimeImage(cfg.Runtime)},
			Labels:    map[string]string{cleanupLabel: ""},
			BuildArgs: map[string]*string{"VERSION": &defaultVersion},
		})
	}

	// Build all the images concurrently
	errChs := make([]chan error, len(runtimeConfigs))
	for i, img := range runtimeConfigs {
		errChs[i] = make(chan error)
		docker.BuildImageAsync(context.Background(), cli, img, errChs[i], docker.OutputOptions{PrintFn: pc.printToDebugLog})
	}

	for i, errCh := range errChs {
		err := <-errCh
		if err != nil {
			image := runtimeConfigs[i]
			hint := ""
			if strings.Contains(image.Path, "/tmp-e2e") {
				hint = fmt.Sprintf("\nHint: Make sure the the templates in '%s' are up to date and the runtime is built correctly.",
					kystrapTemplatesDir)
			}
			return fmt.Errorf("failed to build image %s: %v%s", image.Tags[0], err, hint)
		}
	}
	return nil
}

func (pc *IntegrationBuilder) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer cli.Close()

	labelFilters := []string{fmt.Sprintf("%s=", cleanupLabel), fmt.Sprintf("%s=%s", interchainCleanupLabel, pc.testName)}

	var containers []types.Container
	for _, label := range labelFilters {
		c, err := cli.ContainerList(ctx, container.ListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("label", label)),
		})
		if err != nil {
			return err
		}
		containers = append(containers, c...)
	}
	for _, cont := range containers {
		err = cli.ContainerRemove(ctx, cont.ID, container.RemoveOptions{
			Force: true,
		})
		if err != nil {
			return err
		}
	}

	var volumes []*volume.Volume
	for _, label := range labelFilters {
		v, err := cli.VolumeList(ctx, volume.ListOptions{Filters: filters.NewArgs(
			filters.Arg("label", label),
		)})
		if err != nil {
			return err
		}
		volumes = append(volumes, v.Volumes...)
	}
	for _, vol := range volumes {
		err = cli.VolumeRemove(ctx, vol.Name, true)
		if err != nil {
			return err
		}
	}
	return nil
}

type KystrapRunner struct {
	kystrapConfig docker.ContainerConfig
}

func NewKystrapRunner() *KystrapRunner {
	path, err := filepath.Abs(runtimePathRelative)
	if err != nil {
		panic(fmt.Errorf("failed to get absolute path for runtime folder: %v", err))
	}
	binds := []string{fmt.Sprintf("%s:%s", path, kystrapMount)}
	return &KystrapRunner{
		kystrapConfig: docker.ContainerConfig{
			Image: kystrapImage.Tags[0],
			Name:  kystrapImage.Tags[0],
			Binds: binds,
		},
	}
}

func runDockerAndRemove(ctx context.Context, cli *client.Client, config docker.ContainerConfig, timeout time.Duration) error {
	id, err := docker.StartContainer(ctx, cli, config)
	if err != nil {
		return err
	}

	// Try to remove the container until it succeeds or the timeout is reached
	for {
		ctx, cancel := context.WithTimeout(ctx, timeout)

		err = cli.ContainerRemove(ctx, id, container.RemoveOptions{})
		if err != nil {
			// check if err is a timeout error
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				cancel()
				return fmt.Errorf("failed to remove container: %v", err)
			}
		} else {
			cancel()
			return nil
		}
	}
}

// BootstrapTmpRuntimes creates runtimes for the provided TmpRuntime's
// The runtimes will be created in the runtime folder
func (kr *KystrapRunner) BootstrapTmpRuntimes(tmpRuntimes []TmpRuntime) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer cli.Close()

	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	for _, tmp := range tmpRuntimes {
		config := kr.kystrapConfig
		config.Cmd = []string{"create", "-n", tmp.Name, "-l", tmp.Language, "-y"}
		config.User = fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid)

		err := runDockerAndRemove(ctx, cli, config, time.Second*10)
		if err != nil {
			return err
		}
	}
	return nil
}

type ProtocolRunner struct {
	testapiConfig   docker.ContainerConfig
	runtimeConfig   docker.ContainerConfig
	protocolConfigs []docker.ContainerConfig
	testConfig      TestConfig
	sharedVolume    string
	networkId       string
	restAddress     string
	rpcAddress      string
	label           string
}

func NewProtocolRunner(testConfig TestConfig, networkId string, restAddress string, rpcAddress string) *ProtocolRunner {
	runtimeImageName := runtimeImage(testConfig.Runtime)
	runtimeName := fmt.Sprintf("%s-%s-%s", cleanupLabel, runtimeImagePrefix, testConfig.Runtime.Name)
	label := fmt.Sprintf("%s-%s", cleanupLabel, testConfig.Runtime.Name)

	var protocolConfigs []docker.ContainerConfig
	for _, cfg := range testConfig.GetProtocolConfigs() {
		name := fmt.Sprintf("%s-%s-%s-%s", cleanupLabel, protocolImage.Tags[0], runtimeImageName, cfg.ProtocolNode.KeyName())
		env, err := docker.CreateProtocolEnv(docker.ProtocolEnv{
			Valaccount:  cfg.Valaccount.Mnemonic(),
			RpcAddress:  rpcAddress,
			RestAddress: restAddress,
			Host:        runtimeName,
			PoolId:      testConfig.PoolId,
			Debug:       true,
		})
		if err != nil {
			panic(fmt.Errorf("programming error! This should never happen! Error: %s", err))
		}
		binds := []string{fmt.Sprintf("%s:%s", kyveStorageVolumeName(testConfig.Runtime), kyveStorageMountProtocol)}
		protocolConfigs = append(protocolConfigs, docker.ContainerConfig{
			Image:   protocolImage.Tags[0],
			Name:    name,
			Network: networkId,
			Env:     env,
			Binds:   binds,
			Labels:  map[string]string{cleanupLabel: "", label: ""},
		})
	}
	return &ProtocolRunner{
		testapiConfig: docker.ContainerConfig{
			Image:   testapiImage.Tags[0],
			Name:    fmt.Sprintf("%s-%s-%s", cleanupLabel, testapiImage.Tags[0], runtimeImageName),
			Network: networkId,
			Binds:   []string{fmt.Sprintf("%s:%s:ro", testConfig.Runtime.TestDataApiPath, kyveStorageMountApi)},
			Labels:  map[string]string{cleanupLabel: "", label: ""},
		},
		runtimeConfig: docker.ContainerConfig{
			Image:   runtimeImageName,
			Name:    fmt.Sprintf("%s-%s", cleanupLabel, runtimeImageName),
			Network: networkId,
			Labels:  map[string]string{cleanupLabel: "", label: ""},
		},
		protocolConfigs: protocolConfigs,
		testConfig:      testConfig,
		sharedVolume:    kyveStorageVolumeName(testConfig.Runtime),
		networkId:       networkId,
		restAddress:     restAddress,
		rpcAddress:      rpcAddress,
		label:           label,
	}
}

func runtimeImage(runtime Runtime) string {
	return fmt.Sprintf("%s-%s", runtimeImagePrefix, runtime.Name)
}

func kyveStorageVolumeName(runtime Runtime) string {
	return fmt.Sprintf("%s-%s", kyveStorageName, runtimeImage(runtime))
}

func (pc *ProtocolRunner) RunProtocolContainers() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer cli.Close()

	// Run testapi
	_, err = docker.StartContainer(ctx, cli, pc.testapiConfig)
	if err != nil {
		return err
	}

	// Run runtime
	_, err = docker.StartContainer(ctx, cli, pc.runtimeConfig)
	if err != nil {
		return err
	}

	// Run protocol with multiple nodes
	for _, protocol := range pc.protocolConfigs {
		_, err = docker.StartContainer(ctx, cli, protocol)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pc *ProtocolRunner) StopProtocolContainers() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer cli.Close()

	// Stop containers
	_, err = docker.StopContainers(ctx, cli, pc.label)
	if err != nil {
		return err
	}

	// Remove containers
	return docker.RemoveContainers(ctx, cli, pc.label)
}

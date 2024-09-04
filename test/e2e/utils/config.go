package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"gopkg.in/yaml.v3"
)

// runtimePathRelative is the path to the runtime folder (relative to the root of the e2e test folder)
const (
	rootPath            = "../../"
	runtimePathRelative = rootPath + "runtime"
	testdataPath        = "%s/testdata"
	testdataApiPath     = testdataPath + "/api"
)

type poolConfigYml struct {
	StartKey string                 `default:"1" yaml:"startKey"`
	Config   map[string]interface{} `default:"{}" yaml:"config"`
}

type ProtocolConfig struct {
	ProtocolNode ibc.Wallet
	Valaccount   ibc.Wallet
}

type PoolConfig struct {
	StartKey string `yaml:"startKey"`
	Config   string `yaml:"config"`
}

type Runtime struct {
	Name            string
	Path            string
	TestDataPath    string
	TestDataApiPath string
}

type TestConfig struct {
	Alice      ProtocolConfig
	Bob        ProtocolConfig
	Viktor     ProtocolConfig
	PoolId     uint64
	PoolConfig PoolConfig
	Runtime    Runtime
}

func getPoolConfig(runtime Runtime) (*PoolConfig, error) {
	path, err := filepath.Abs(fmt.Sprintf("%s/config.yml", runtime.TestDataPath))
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist", path)
	}

	// Read the config.yml file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var obj poolConfigYml

	// Set default values from struct tags
	err = defaults.Set(&obj)
	if err != nil {
		return nil, err
	}

	// Unmarshal the yaml
	err = yaml.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}

	// Convert to json
	jsonData, err := json.Marshal(obj.Config)
	if err != nil {
		return nil, err
	}

	return &PoolConfig{
		StartKey: obj.StartKey,
		Config:   string(jsonData),
	}, nil
}

func ensurePathExists(template string, basePath string) (string, error) {
	path, err := filepath.Abs(fmt.Sprintf(template, basePath))
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("%s does not exist", path)
	}
	return path, nil
}

// getRuntimes returns a list of all runtime folder names
func getRuntimes() ([]Runtime, error) {
	path, err := filepath.Abs(runtimePathRelative)
	if err != nil {
		return nil, err
	}
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var runtimeDirs []Runtime
	for _, entry := range dirEntries {
		if entry.IsDir() {
			runtime := filepath.Join(path, entry.Name())
			testDataPath, err := ensurePathExists(testdataPath, runtime)
			if err != nil {
				return nil, err
			}
			testDataApiPath, err := ensurePathExists(testdataApiPath, runtime)
			if err != nil {
				return nil, err
			}
			runtimeDirs = append(runtimeDirs, Runtime{
				Path:            runtime,
				Name:            entry.Name(),
				TestDataPath:    testDataPath,
				TestDataApiPath: testDataApiPath,
			})
		}
	}
	return runtimeDirs, nil
}

func GetTestConfigs() ([]*TestConfig, error) {
	var testConfigs []*TestConfig
	runtimes, err := getRuntimes()
	if err != nil {
		return nil, err
	}
	for _, runtime := range runtimes {
		poolConfig, err := getPoolConfig(runtime)
		if err != nil {
			return nil, err
		}
		testConfigs = append(testConfigs, &TestConfig{
			PoolConfig: *poolConfig,
			Runtime:    runtime,
		})
	}
	return testConfigs, nil
}

type TmpRuntime struct {
	Name     string
	Path     string
	Language string
}

func GetTmpRuntimeDirectories() ([]TmpRuntime, error) {
	// Read Languages from directories in templates folder
	// Every directory is a language
	var languages []string
	dirs, err := os.ReadDir(kystrapTemplatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %v", err)
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			languages = append(languages, dir.Name())
		}
	}

	iPath, err := filepath.Abs(runtimePathRelative)
	if err != nil {
		return nil, err
	}

	// Build the tmp folder names
	var tmpRuntimes []TmpRuntime
	for _, language := range languages {
		name := fmt.Sprintf("tmp-e2e-%s", language)
		tmpRuntimes = append(tmpRuntimes, TmpRuntime{
			Name:     name,
			Path:     fmt.Sprintf("%s/%s", iPath, name),
			Language: language,
		})
	}
	return tmpRuntimes, nil
}

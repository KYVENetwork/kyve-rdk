package config

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"

	commoncmd "github.com/KYVENetwork/kyve-rdk/common/goutils/cmd"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

const configFileName = "config.toml"

var FlagHome = commoncmd.StringFlag{
	Name:         "home",
	DefaultValue: "~/.kysor", // Overwritten in init() to set the path as absolute
	Usage:        "The location of the .kysor home directory where binaries and configs are stored.",
}

var config *KysorConfig

type KysorConfig struct {
	ChainID string `koanf:"chainId"`
	RPC     string `koanf:"rpc"`
	REST    string `koanf:"rest"`
}

func (c KysorConfig) GetDenom() string {
	if strings.HasPrefix(c.ChainID, "kaon") || strings.HasPrefix(c.ChainID, "korellia") {
		return "tkyve"
	}
	return "ukyve"
}

func (c KysorConfig) GetChainPrettyName() string {
	if strings.HasPrefix(c.ChainID, "kyve") {
		return "kyve"
	}
	if strings.HasPrefix(c.ChainID, "kaon") {
		return "kaon"
	}
	if strings.HasPrefix(c.ChainID, "korellia") {
		return "korellia"
	}
	return c.ChainID
}

func (c KysorConfig) GetWorkingRPC() (string, error) {
	var client = http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
		},
	}

	// Ping the rpc to check if it is working
	// If it is not working, try the next one
	for _, rpc := range strings.Split(c.RPC, ",") {
		_, err := client.Get(rpc + "/status")
		if err == nil {
			return rpc, nil
		}
	}
	return "", fmt.Errorf("no working rpc found")
}

func (c KysorConfig) GetWorkingREST() (string, error) {
	var client = http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
		},
	}

	// Ping the rest to check if it is working
	// If it is not working, try the next one
	for _, rest := range strings.Split(c.REST, ",") {
		_, err := client.Get(rest)
		if err == nil {
			return rest, nil
		}
	}
	return "", fmt.Errorf("no working rest found")
}

func (c KysorConfig) Save(path string) error {
	return save(c, path)
}

func save(s interface{}, path string) error {
	if DoesConfigExist(path) {
		return fmt.Errorf("config file already exists at %s", path)
	}

	k := koanf.New(".")

	// Load the config into koanf
	err := k.Load(structs.Provider(s, "koanf"), nil)
	if err != nil {
		return err
	}

	b, err := k.Marshal(toml.Parser())
	if err != nil {
		return err
	}

	// Create the config directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	// Save the config file
	err = os.WriteFile(path, b, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func GetHomeDir(cmd *cobra.Command) (string, error) {
	// Go up the command tree until the root command is reached
	if cmd.Parent() != nil {
		return GetHomeDir(cmd.Parent())
	}
	// Get the home directory from the flags
	homeDir, err := cmd.Flags().GetString(FlagHome.Name)
	if err != nil {
		return "", err
	}
	// Expand the home directory
	homeDir, err = homedir.Expand(homeDir)
	if err != nil {
		return "", err
	}
	return filepath.Abs(homeDir)
}

func GetConfigFilePath(cmd *cobra.Command) (string, error) {
	homeDir, err := GetHomeDir(cmd)
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, configFileName), nil
}

func DoesConfigExist(configFilePath string) bool {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return false
	} else if err != nil {
		cobra.CheckErr(err)
	}
	return true
}

// GetConfigX returns the current config, assuming it has been loaded
// It panics if the config has not been loaded
func GetConfigX() KysorConfig {
	if config == nil {
		cobra.CheckErr(fmt.Errorf("config has not been initialized"))
	}
	return *config
}

func LoadConfigs(cmd *cobra.Command, _ []string) error {
	err := loadKysorConfig(cmd, nil)
	if err != nil {
		return err
	}
	return loadValaccountConfigs(cmd, nil)
}

func loadKysorConfig(cmd *cobra.Command, _ []string) error {
	configFilePath, err := GetConfigFilePath(cmd)
	if err != nil {
		return err
	}

	// Check if the config file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist at %s. Run `kysor init` to create a new config", configFilePath)
	}

	k := koanf.New(".")

	// Load the config file
	if err := k.Load(file.Provider(configFilePath), toml.Parser()); err != nil {
		return fmt.Errorf("error loading config file: %s", err)
	}

	// Unmarshal the config file into the config struct
	err = k.Unmarshal("", &config)
	if err != nil {
		return fmt.Errorf("error unmarshalling config file: %s", err)
	}

	// Add port to the RPC URLs if they are missing
	var rpcs []string
	for _, rpc := range strings.Split(config.RPC, ",") {
		if !regexp.MustCompile(`:\d+$`).MatchString(rpc) {
			if strings.HasPrefix(rpc, "https://") {
				rpc += ":443"
			} else if strings.HasPrefix(rpc, "http://") {
				rpc += ":80"
			}
		}
		rpcs = append(rpcs, rpc)
	}
	config.RPC = strings.Join(rpcs, ",")

	// Add port to the REST URLs if they are missing
	var rests []string
	for _, rest := range strings.Split(config.REST, ",") {
		if !regexp.MustCompile(`:\d+$`).MatchString(rest) {
			if strings.HasPrefix(rest, "https://") {
				rest += ":443"
			} else if strings.HasPrefix(rest, "http://") {
				rest += ":80"
			}
		}
		rests = append(rests, rest)
	}
	config.REST = strings.Join(rests, ",")

	return nil
}

func init() {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	FlagHome.DefaultValue = filepath.Join(homeDir, ".kysor")
}

package cmd

import (
	"context"
	"fmt"
	querytypes "github.com/KYVENetwork/chain/x/query/types"
	commoncmd "github.com/KYVENetwork/kyve-rdk/common/goutils/cmd"
	"github.com/KYVENetwork/kyve-rdk/common/goutils/docker"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/chain"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/config"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/utils"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"strings"
)

type runStatus struct {
	containers []string
	valConfig  config.ValaccountConfig
	pool       querytypes.PoolResponse
}

func getRunStatus(cli *client.Client, kyveClient *chain.KyveClient) (runStatusList []runStatus, err error) {
	containers, err := docker.ListContainers(context.Background(), cli, globalContainerLabel)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	for _, valConfig := range config.ValaccountConfigOptions {
		rs := runStatus{valConfig: valConfig.Value()}
		for _, cont := range containers {
			label := valConfig.Value().GetContainerLabel()
			if _, ok := cont.Labels[label]; ok {
				rs.containers = append(rs.containers, strings.TrimPrefix(cont.Names[0], "/"))
			}
		}

		// Add only if there are running containers
		if len(rs.containers) != 0 {
			pool, err := kyveClient.QueryPool(rs.valConfig.Pool)
			if err != nil {
				return nil, err
			}
			rs.pool = pool.GetPool()

			runStatusList = append(runStatusList, rs)
		}
	}

	return runStatusList, nil
}

func getBaseUrl() string {
	chainId := config.GetConfigX().ChainID
	chainPrefix := chainId[:strings.LastIndex(chainId, "-")]
	return fmt.Sprintf("https://app.%s.kyve.network", chainPrefix)
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show KYSOR status",
		PreRunE: commoncmd.CombineFuncs(utils.CheckDockerInstalled, utils.CheckUpdateAvailable, config.LoadConfigs, commoncmd.SetupInteractiveMode),
		RunE: func(cmd *cobra.Command, args []string) error {
			kyveClient, err := chain.NewKyveClient(config.GetConfigX(), config.ValaccountConfigs)
			if err != nil {
				return err
			}

			cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
			if err != nil {
				return fmt.Errorf("failed to create docker client: %v", err)
			}
			//goland:noinspection GoUnhandledErrorResult
			defer cli.Close()

			runStatusList, err := getRunStatus(cli, kyveClient)
			if err != nil {
				return fmt.Errorf("failed to get run status: %v", err)
			}

			if len(runStatusList) == 0 {
				fmt.Println("No containers are running")
				return nil
			}

			for _, rs := range runStatusList {
				baseUrl := getBaseUrl()
				poolUlr := fmt.Sprintf("%s/#/pools/%d", baseUrl, rs.pool.GetData().Id)
				fmt.Printf("Valaccount: %s\n", rs.valConfig.Name())
				fmt.Printf("  Pool: %s (ID: %d) -> %s\n", rs.pool.GetData().Name, rs.pool.GetData().Id, poolUlr)
				fmt.Print("  Running docker containers:\n")
				for _, cont := range rs.containers {
					fmt.Printf("    - %s\n", cont)
				}
				fmt.Printf("  Log commands:\n")
				for _, cont := range rs.containers {
					fmt.Print("    ")
					utils.PrintlnItalic(fmt.Sprintf("docker logs -f %s", cont))
				}
				fmt.Println()
			}

			return nil
		},
	}
}

func init() {
	rootCmd.AddCommand(statusCmd())
}

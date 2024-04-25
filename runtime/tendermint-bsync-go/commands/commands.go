package commands

import (
	"fmt"
	"github.com/KYVENetwork/kyve-rdk/runtime/tendermint-bsync-go/server"
	"github.com/spf13/cobra"
)

var (
	port  int32
	debug bool
)

var rootCmd = &cobra.Command{
	Use:   "runtime",
	Short: "Runtime",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of the runtime",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(server.RuntimeVersion)
	},
}

func init() {
	startCmd.Flags().Int32VarP(&port, "port", "p", 50051, "port")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "debug")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the runtime",
	Run: func(cmd *cobra.Command, args []string) {
		server.StartServer(port, debug)
	},
}

func Execute() {
	versionCmd.Flags()
	startCmd.Flags()

	if err := rootCmd.Execute(); err != nil {
		panic(fmt.Errorf("failed to execute root command: %w", err))
	}
}

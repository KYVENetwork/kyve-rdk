package cmd

import (
	"fmt"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/types"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/utils"
	"runtime"

	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Show KYSOR version",
		PreRunE: utils.CheckUpdateAvailable,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Kysor version: %s\n", types.Version)
			if types.Tag != "" {
				fmt.Printf("Git tag: %s\n", types.Tag)
			}
			if types.Commit != "" {
				fmt.Printf("Git commit: %s\n", types.Commit)
			}
			fmt.Println()
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Println()
			fmt.Printf("Platform: %s\n", runtime.GOOS)
			fmt.Printf("Arch: %s\n", runtime.GOARCH)
		},
	}
}

func init() {
	rootCmd.AddCommand(versionCmd())
}

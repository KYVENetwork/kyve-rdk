package utils

import (
	"fmt"
	"github.com/KYVENetwork/kyve-rdk/tools/kysor/cmd/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/go-version"
	"strings"

	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/savioxavier/termlink"

	commoncmd "github.com/KYVENetwork/kyve-rdk/common/goutils/cmd"
	"github.com/spf13/cobra"
)

var hasInteractiveInfoBeenShown = false

// ShowInteractiveInfo prints a message to the user that the command is running in interactive mode.
func ShowInteractiveInfo() {
	if !hasInteractiveInfoBeenShown {
		fmt.Println("KYSOR is running in interactive mode.")
		fmt.Println("Add '-y' to your command to disable interactive mode.")
		fmt.Println()
		hasInteractiveInfoBeenShown = true
	}
}

func RunPromptCommandE(cmd *cobra.Command, args []string) error {
	// Check if the interactive flag is set
	// -> if so ask the user what to do
	if commoncmd.IsInteractive(cmd) {
		ShowInteractiveInfo()

		// Prompt for the next command
		nextCmd, err := commoncmd.PromptCmd(cmd.Commands())
		if err != nil {
			return err
		}

		// Run persistent pre run functions
		if nextCmd.PersistentPreRunE != nil {
			err = nextCmd.PersistentPreRunE(nextCmd, args)
			if err != nil {
				return err
			}
		} else if nextCmd.PersistentPreRun != nil {
			nextCmd.PersistentPreRun(nextCmd, args)
		}

		// Run pre run functions
		if nextCmd.PreRunE != nil {
			err = nextCmd.PreRunE(nextCmd, args)
			if err != nil {
				return err
			}
		} else if nextCmd.PreRun != nil {
			nextCmd.PreRun(nextCmd, args)
		}

		// Run the next command
		if nextCmd.RunE != nil {
			return nextCmd.RunE(nextCmd, args)
		} else if nextCmd.Run != nil {
			nextCmd.Run(nextCmd, args)
			return nil
		}
		return fmt.Errorf("no run function defined for command: %s", nextCmd.Name())
	}
	// Otherwise show the help
	return cmd.Help()
}

func CheckDockerInstalled(_ *cobra.Command, _ []string) error {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		cyan := color.New(color.FgHiMagenta).SprintFunc()
		hyperlink := cyan(termlink.Link("Install Docker", "https://docs.docker.com/engine/install", true))
		return fmt.Errorf("failed to connect to the docker daemon: %w\n"+
			"- Did you install docker? (%s)\n"+
			"- Is the docker deamon running?", err, hyperlink)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer cli.Close()
	return nil
}

func CheckUpdateAvailable(_ *cobra.Command, _ []string) error {
	currentVersion, err := version.NewVersion(types.Version)
	if err != nil {
		// Silently ignore the error
		return nil
	}

	// Create the remote with repository URL
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{types.RepoUrl},
	})

	// We can then use every Remote functions to retrieve wanted information
	refs, err := rem.List(&git.ListOptions{
		// Returns all references, including peeled references.
		PeelingOption: git.AppendPeeled,
	})
	if err != nil {
		return err
	}

	// Only check for tags that are newer than the current version
	var latestTag *version.Version
	for _, ref := range refs {
		if ref.Name().IsTag() && strings.HasPrefix(ref.Name().Short(), "tools/kysor@") {
			v, err := version.NewVersion(strings.TrimPrefix(ref.Name().Short(), "tools/kysor@"))
			if err != nil {
				continue
			}
			if v.GreaterThan(currentVersion) {
				if latestTag == nil || v.GreaterThan(latestTag) {
					latestTag = v
				}
			}
		}
	}

	if latestTag != nil {
		readmeLink := "/tree/main/tools/kysor#installationupdate"
		updateLink := types.RepoUrl + readmeLink
		fmt.Printf("🎉  A new version of KYSOR is available: %s\n    Update guide: %s\n\n", latestTag.Original(), updateLink)
	}
	return nil
}

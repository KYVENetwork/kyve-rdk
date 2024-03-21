package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	commoncmd "github.com/KYVENetwork/kyve-rdk/common/goutils/cmd"

	"github.com/KYVENetwork/kyve-rdk/tools/kystrap/bootstrap"
	"github.com/KYVENetwork/kyve-rdk/tools/kystrap/types"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var regexpAlphaNumericAndDash = regexp.MustCompile(`^[a-z0-9-]+$`)

var (
	flagLanguage = commoncmd.OptionFlag[types.Language]{
		Name:     "language",
		Short:    "l",
		Usage:    fmt.Sprintf("Language for your runtime (%s)", strings.Join(types.LanguagesStringSlice(), ", ")),
		Prompt:   "Select the Language for your runtime",
		Required: true,
		ValidateFn: func(input string) error {
			if commoncmd.ValidateNotEmpty(input) != nil {
				return fmt.Errorf("language must not be empty")
			}
			for _, language := range types.Languages {
				if language.StringValue() == input {
					return nil
				}
			}
			return fmt.Errorf("invalid language. Please choose one from '%s'", strings.Join(types.LanguagesStringSlice(), ", "))
		},
	}
	flagName = commoncmd.StringFlag{
		Name:     "name",
		Short:    "n",
		Usage:    "Name for your runtime",
		Prompt:   "Set a name for your runtime",
		Required: true,
		ValidateFn: func(input string) error {
			if len(input) < 3 {
				return errors.New("name must be at least 3 characters long")
			}
			if !regexpAlphaNumericAndDash.MatchString(input) {
				return errors.New("name must only contain lowercase alphanumeric characters and dashes")
			}
			return nil
		},
	}
	flagOutput = commoncmd.StringFlag{
		Name:         "output",
		Short:        "o",
		Usage:        "Output directory for your runtime",
		DefaultValue: "runtime",
	}
)

func promptLinkToGithub() error {
	const githubUrl = "https://github.com/KYVENetwork/kyve-rdk/issues/new"
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Please create an issue on %s", githubUrl),
		IsConfirm: true,
		Default:   "y",
	}
	_, err := prompt.Run()
	return err
}

func CmdCreateRuntime() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create runtime",
		PreRunE: commoncmd.SetupInteractiveMode,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Output directory
			outputDir, err := cmd.Flags().GetString(flagOutput.Name)
			if err != nil {
				return err
			}

			// Language
			languageOption, err := commoncmd.GetOptionFromPromptOrFlag(cmd, flagLanguage)
			if err != nil {
				return err
			}
			language := languageOption.Value()
			if language.IsRequestOtherLanguage() {
				return promptLinkToGithub()
			}

			// Name
			name, err := commoncmd.GetStringFromPromptOrFlag(cmd, flagName)
			if err != nil {
				return err
			}

			// Create runtime
			err = bootstrap.CreateRuntime(outputDir, language, name)
			if err != nil {
				return err
			}

			err = bootstrap.UpdateReleasePleaseConfig(language, name)
			if err != nil {
				return err
			}

			fmt.Printf("✅ Successfully created runtime in `%s`\n", name)
			return nil
		},
	}
	flagLanguage.Options = types.Languages
	commoncmd.AddOptionFlags(cmd, []commoncmd.OptionFlag[types.Language]{flagLanguage})
	commoncmd.AddStringFlags(cmd, []commoncmd.StringFlag{flagName, flagOutput})
	return cmd
}

func init() {
	rootCmd.AddCommand(CmdCreateRuntime())
}

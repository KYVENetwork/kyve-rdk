package cmd

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type noBellStdout struct{}

func (n *noBellStdout) Write(p []byte) (int, error) {
	if len(p) == 1 && p[0] == readline.CharBell {
		return 0, nil
	}
	return readline.Stdout.Write(p)
}

func (n *noBellStdout) Close() error {
	return readline.Stdout.Close()
}

var NoBellStdout = &noBellStdout{}

func getPromptString(cmd *cobra.Command) string {
	return fmt.Sprintf("%s - %s", cmd.Name(), cmd.Short)
}

// PromptCmd prompts the user to select one of the given options.
func PromptCmd(options []*cobra.Command) (*cobra.Command, error) {
	var items []string

	// Commands that will not be shown in the list
	blacklist := []string{"completion", "help"}
	for _, option := range options {
		if !slices.Contains(blacklist, option.Name()) {
			items = append(items, getPromptString(option))
		}
	}

	prompt := promptui.Select{
		Label:  "What do you want to do?",
		Items:  items,
		Stdout: NoBellStdout,
		Size:   len(options),
	}

	_, result, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	for _, option := range options {
		if getPromptString(option) == result {
			return option, nil
		}
	}
	return nil, fmt.Errorf("invalid option: %s", result)
}

type YesNoOption string

const (
	Yes YesNoOption = "Yes"
	No  YesNoOption = "No"
)

func PromptYesNo(label string, defaultValue YesNoOption) (bool, error) {
	cursorPos := 0
	if defaultValue == No {
		cursorPos = 1
	}
	prompt := promptui.Select{
		Label:     label,
		Items:     []string{string(Yes), string(No)},
		CursorPos: cursorPos,
		Stdout:    NoBellStdout,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return result == string(Yes), nil
}

// GetStringFromFlag returns the string value from the given flag
func GetStringFromFlag(cmd *cobra.Command, flag StringFlag) (string, error) {
	value, err := cmd.Flags().GetString(flag.Name)
	if err != nil {
		return "", err
	}

	// If the flag was set and a validation function exists, we need to validate the value
	if cmd.Flags().Changed(flag.Name) && flag.ValidateFn != nil {
		err = flag.ValidateFn(value)
		if err != nil {
			return "", err
		}
	}
	return value, nil
}

// GetStringFromPromptOrFlag returns the string value from
// 1. the given flag
// 2. prompts the user for the value if the flag was not set
func GetStringFromPromptOrFlag(cmd *cobra.Command, flag StringFlag) (string, error) {
	if IsInteractive(cmd) && !cmd.Flags().Changed(flag.Name) {
		// Only prompt if we are in interactive mode and the flag was not set
		label := flag.Prompt
		if label == "" {
			label = flag.Usage
		}

		prompt := promptui.Prompt{
			Label:    label,
			Validate: flag.ValidateFn,
			Default:  flag.DefaultValue,
			Stdout:   NoBellStdout,
		}
		return prompt.Run()
	} else {
		return GetStringFromFlag(cmd, flag)
	}
}

// GetBoolFromPromptOrFlag returns the bool value from
// 1. the given flag
// 2. prompts the user for the value if the flag was not set
func GetBoolFromPromptOrFlag(cmd *cobra.Command, flag BoolFlag) (bool, error) {
	value, err := cmd.Flags().GetBool(flag.Name)
	if err != nil {
		return false, err
	}

	if IsInteractive(cmd) && !cmd.Flags().Changed(flag.Name) {
		// Only prompt if we are in interactive mode and the flag was not set
		defaultValue := Yes
		if !value {
			defaultValue = No
		}

		label := flag.Prompt
		if label == "" {
			label = flag.Usage
		}

		return PromptYesNo(label, defaultValue)
	}
	return value, nil
}

// GetIntFromPromptOrFlag returns the int value from
// 1. the given flag
// 2. prompts the user for the value if the flag was not set
func GetIntFromPromptOrFlag(cmd *cobra.Command, flag IntFlag) (int64, error) {
	value, err := cmd.Flags().GetInt64(flag.Name)
	if err != nil {
		return 0, err
	}

	if IsInteractive(cmd) && !cmd.Flags().Changed(flag.Name) {
		// Only prompt if we are in interactive mode and the flag was not set
		label := flag.Prompt
		if label == "" {
			label = flag.Usage
		}

		prompt := promptui.Prompt{
			Label:    label,
			Validate: flag.ValidateFn,
			Default:  strconv.FormatInt(flag.DefaultValue, 10),
			Stdout:   NoBellStdout,
		}
		result, err := prompt.Run()
		if err != nil {
			return 0, err
		}

		parsed, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	} else if cmd.Flags().Changed(flag.Name) {
		// If the flag was set we need to validate it (if a validation function is set)
		if flag.ValidateFn != nil {
			err = flag.ValidateFn(strconv.FormatInt(value, 10))
			if err != nil {
				return 0, err
			}
		}
	}
	return value, nil
}

// GetOptionFromPrompt returns the option value from a select prompt
func GetOptionFromPrompt[T any](flag OptionFlag[T]) (Option[T], error) {
	label := flag.Prompt
	if label == "" {
		label = flag.Usage
	}

	cursorPos := 0
	var items []string
	for i, option := range flag.Options {
		items = append(items, option.Name())
		if option == flag.DefaultValue {
			cursorPos = i
		}
	}

	size := len(items)
	if flag.MaxSelectionSize > 0 {
		size = int(flag.MaxSelectionSize)
	}

	prompt := promptui.Select{
		Label:             label,
		Items:             items,
		Stdout:            NoBellStdout,
		Size:              size,
		CursorPos:         cursorPos,
		StartInSearchMode: flag.StartInSearchMode,
		Searcher: func(input string, index int) bool {
			return strings.Contains(strings.ToLower(items[index]), input)
		},
	}
	_, result, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	for _, option := range flag.Options {
		if option.Name() == result {
			return option, nil
		}
	}
	return nil, fmt.Errorf("invalid option: %s", result)
}

// GetOptionFromPromptOrFlag returns the option value from
// 1. the given flag
// 2. prompts the user for the value if the flag was not set
func GetOptionFromPromptOrFlag[T any](cmd *cobra.Command, flag OptionFlag[T]) (Option[T], error) {
	value, err := cmd.Flags().GetString(flag.Name)
	if err != nil {
		return nil, err
	}

	if len(flag.Options) == 0 {
		return nil, fmt.Errorf("no options available")
	}

	if IsInteractive(cmd) && !cmd.Flags().Changed(flag.Name) {
		// Only prompt if we are in interactive mode and the flag was not set
		return GetOptionFromPrompt(flag)
	} else if cmd.Flags().Changed(flag.Name) {
		// If the flag was set we need to validate it (if a validation function is set)
		if flag.ValidateFn != nil {
			err = flag.ValidateFn(value)
			if err != nil {
				return nil, err
			}
		}
	}
	for _, option := range flag.Options {
		if option.StringValue() == value || option.Name() == value {
			return option, nil
		}
	}
	return nil, fmt.Errorf("invalid option: %s", value)
}

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var FlagNonInteractive = BoolFlag{Name: "yes", Short: "y", DefaultValue: false, Usage: "Non-interactive mode: Skips all prompts (default false)", Required: false}

// AddStringFlags adds the given string flags to the given command.
// If a flag is required it will be marked as required.
func AddStringFlags(cmd *cobra.Command, flags []StringFlag) {
	for _, f := range flags {
		cmd.Flags().StringP(f.Name, f.Short, f.DefaultValue, f.Usage)
		if f.Required {
			err := cmd.MarkFlagRequired(f.Name)
			if err != nil {
				panic(err)
			}
		}
	}
}

// AddBoolFlags adds the given bool flags to the given command.
// If a flag is required it will be marked as required.
func AddBoolFlags(cmd *cobra.Command, flags []BoolFlag) {
	for _, f := range flags {
		cmd.Flags().BoolP(f.Name, f.Short, f.DefaultValue, f.Usage)
		if f.Required {
			err := cmd.MarkFlagRequired(f.Name)
			if err != nil {
				panic(err)
			}
		}
	}
}

// AddIntFlags adds the given int flags to the given command.
// If a flag is required it will be marked as required.
func AddIntFlags(cmd *cobra.Command, flags []IntFlag) {
	for _, f := range flags {
		cmd.Flags().Int64P(f.Name, f.Short, f.DefaultValue, f.Usage)
		if f.Required {
			err := cmd.MarkFlagRequired(f.Name)
			if err != nil {
				panic(err)
			}
		}
	}
}

// AddOptionFlags adds the given option flags to the given command.
// If a flag is required it will be marked as required.
func AddOptionFlags[T any](cmd *cobra.Command, flags []OptionFlag[T]) {
	for _, f := range flags {
		var defaultValue string
		if f.DefaultValue != nil {
			defaultValue = f.DefaultValue.Name()
		}
		cmd.Flags().StringP(f.Name, f.Short, defaultValue, f.Usage)
		if f.Required {
			err := cmd.MarkFlagRequired(f.Name)
			if err != nil {
				panic(err)
			}
		}
	}
}

// IsInteractive returns true if the non-interactive flag was not set.
func IsInteractive(cmd *cobra.Command) bool {
	return !cmd.Flags().Changed(FlagNonInteractive.Name)
}

// SetupInteractiveMode sets up the interactive mode for the given command.
// This means that all flags are not required anymore.
// Load the config file before running this function.
func SetupInteractiveMode(cmd *cobra.Command, _ []string) error {
	if IsInteractive(cmd) {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			for val, annotation := range f.Annotations {
				if val == cobra.BashCompOneRequiredFlag {
					annotation[0] = "false"
				}
			}
		})
	}
	return nil
}

func AddPersistentStringFlags(cmd *cobra.Command, flags []StringFlag) {
	for _, f := range flags {
		cmd.PersistentFlags().StringP(f.Name, f.Short, f.DefaultValue, f.Usage)
	}
}

func AddPersistentBoolFlags(cmd *cobra.Command, flags []BoolFlag) {
	for _, f := range flags {
		cmd.PersistentFlags().BoolP(f.Name, f.Short, f.DefaultValue, f.Usage)
	}
}

func CombineFuncs(funcs ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range funcs {
			err := f(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

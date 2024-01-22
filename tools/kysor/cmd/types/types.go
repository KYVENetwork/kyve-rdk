package types

import "fmt"

type StringFlag struct {
	Name         string
	Short        string
	DefaultValue string
	Usage        string
	Prompt       string
	Required     bool
	ValidateFn   func(input string) error
}

type BoolFlag struct {
	Name         string
	Short        string
	DefaultValue bool
	Usage        string
	Prompt       string
	Required     bool
}

type IntFlag struct {
	Name         string
	Short        string
	DefaultValue int64
	Usage        string
	Prompt       string
	Required     bool
	ValidateFn   func(input string) error
}

type CmdConfig struct {
	Name  string
	Short string
}

func (c CmdConfig) ActionString() string {
	return fmt.Sprintf("%s (%s)", c.Short, c.Name)
}

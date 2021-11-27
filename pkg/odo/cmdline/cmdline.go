// cmdline package provides an abstration of a cmdline utility
package cmdline

import (
	"context"

	"github.com/spf13/cobra"
)

type Cmdline interface {
	// GetWorkingdirectory returns tehe directory on which the command should execute
	GetWorkingDirectory() (string, error)

	// FlagValueIfSet returns the value for a flag
	FlagValueIfSet(flagName string) string

	// IsFlagSet returns true if the flag is explicitely set
	IsFlagSet(flagName string) bool

	// CheckIfConfigurationNeeded checks against a set of commands that do *NOT* need configuration.
	CheckIfConfigurationNeeded() (bool, error)

	// Context returns the context attached to the command
	Context() context.Context

	// GetArgsAfterDashes returns the sub-array of args after `--`
	// returns an error if no args were passed after --
	GetArgsAfterDashes(args []string) ([]string, error)

	// GetParentName returns an empty string is there is no parent of the name of the parent
	GetParentName() string

	GetRootName() string

	GetName() string

	// TODO temporary, to be removed
	GetCmd() *cobra.Command
}

package cmdutil

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// RequireArgs returns an Args validator that allows missing args for interactive mode
// When args are missing and we're in an interactive terminal, it returns nil
// so the command can handle interactive input in RunE
func RequireArgs(n int, argNames string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			// If running in interactive terminal, allow the command to handle it
			if IsInteractive() {
				return nil
			}
			// Non-interactive: show help and error
			cmd.Help()
			fmt.Println()
			if n == 1 {
				return fmt.Errorf("missing required argument: %s", argNames)
			}
			return fmt.Errorf("missing required arguments: %s", argNames)
		}
		if len(args) > n {
			return fmt.Errorf("too many arguments provided")
		}
		return nil
	}
}

// IsInteractive returns true if stdin is a terminal (interactive mode)
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// NeedsMissingArg returns true if we need to prompt for a missing argument
func NeedsMissingArg(args []string, required int) bool {
	return len(args) < required && IsInteractive()
}

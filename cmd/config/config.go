package config

import "github.com/spf13/cobra"

// Cmd is the config command
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage hevycli configuration",
	Long: `View and modify hevycli configuration settings.

Configuration is stored in ~/.hevycli/config.yaml and can be overridden
by environment variables (HEVYCLI_*) or command-line flags.

Examples:
  hevycli config init          # Interactive setup
  hevycli config show          # Display current configuration
  hevycli config set api-key   # Set your API key`,
}

func init() {
	Cmd.AddCommand(initCmd)
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(setCmd)
}

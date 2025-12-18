package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize hevycli configuration",
	Long: `Interactive setup for hevycli configuration.

This command creates the configuration file and prompts for:
- API key (required, from https://hevy.com/settings?developer)
- Display preferences (output format, units)

The configuration is saved to ~/.hevycli/config.yaml

Note: Hevy Pro subscription is required for API access.`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	cfg := config.DefaultConfig()

	fmt.Println("Welcome to hevycli configuration!")
	fmt.Println()

	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config already exists
	existingCfg, err := config.Load("")
	if err == nil && existingCfg.API.Key != "" {
		fmt.Printf("Existing configuration found at %s\n", config.ConfigPath())
		fmt.Print("Do you want to overwrite it? [y/N]: ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Configuration unchanged.")
			return nil
		}
		fmt.Println()
	}

	// Prompt for API key
	fmt.Println("Get your API key at: https://hevy.com/settings?developer")
	fmt.Println("(Requires Hevy Pro subscription)")
	fmt.Println()
	fmt.Print("Enter your Hevy API key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Validate the API key
	fmt.Print("Validating API key... ")
	client := api.NewClient(apiKey)
	if err := client.ValidateAuth(); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("invalid API key: %w", err)
	}
	fmt.Println("OK")
	cfg.API.Key = apiKey
	fmt.Println()

	// Prompt for units
	fmt.Print("Preferred units (metric/imperial) [metric]: ")
	units, _ := reader.ReadString('\n')
	units = strings.TrimSpace(strings.ToLower(units))
	if units == "imperial" {
		cfg.Display.Units = "imperial"
	} else {
		cfg.Display.Units = "metric"
	}

	// Prompt for default output format
	fmt.Print("Default output format (json/table/plain) [table]: ")
	outputFmt, _ := reader.ReadString('\n')
	outputFmt = strings.TrimSpace(strings.ToLower(outputFmt))
	if outputFmt == "json" || outputFmt == "plain" {
		cfg.Display.OutputFormat = outputFmt
	} else {
		cfg.Display.OutputFormat = "table"
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("Configuration saved to %s\n", config.ConfigPath())
	fmt.Println()
	fmt.Println("You're all set! Try these commands:")
	fmt.Println("  hevycli config show     # View your configuration")
	fmt.Println("  hevycli workout list    # List your workouts (coming soon)")

	return nil
}

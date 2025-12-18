package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
)

var validateKey bool

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  api-key          Your Hevy API key
  default-output   Default output format (json, table, plain)
  units            Measurement units (metric, imperial)
  color            Enable/disable colored output (true, false)
  date-format      Date format (Go format string, e.g., "2006-01-02")
  time-format      Time format (Go format string, e.g., "15:04")

Examples:
  hevycli config set api-key your-api-key-here
  hevycli config set units imperial
  hevycli config set default-output json
  hevycli config set color false`,
	Args: cobra.ExactArgs(2),
	RunE: runSet,
}

func init() {
	setCmd.Flags().BoolVar(&validateKey, "validate", true,
		"validate API key before saving (for api-key only)")
}

func runSet(cmd *cobra.Command, args []string) error {
	key := strings.ToLower(args[0])
	value := args[1]

	// Load existing config or create default
	cfg, err := config.Load("")
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Map key to config field
	switch key {
	case "api-key", "apikey", "key":
		if validateKey && value != "" {
			fmt.Print("Validating API key... ")
			client := api.NewClient(value)
			if err := client.ValidateAuth(); err != nil {
				fmt.Println("FAILED")
				return fmt.Errorf("invalid API key: %w", err)
			}
			fmt.Println("OK")
		}
		cfg.API.Key = value

	case "default-output", "output-format", "output":
		value = strings.ToLower(value)
		if value != "json" && value != "table" && value != "plain" {
			return fmt.Errorf("invalid output format: %s (must be json, table, or plain)", value)
		}
		cfg.Display.OutputFormat = value

	case "units":
		value = strings.ToLower(value)
		if value != "metric" && value != "imperial" {
			return fmt.Errorf("invalid units: %s (must be metric or imperial)", value)
		}
		cfg.Display.Units = value

	case "color":
		value = strings.ToLower(value)
		cfg.Display.Color = value == "true" || value == "1" || value == "yes" || value == "on"

	case "date-format":
		cfg.Display.DateFormat = value

	case "time-format":
		cfg.Display.TimeFormat = value

	case "base-url", "baseurl":
		cfg.API.BaseURL = value

	default:
		return fmt.Errorf("unknown configuration key: %s\n\nAvailable keys: api-key, default-output, units, color, date-format, time-format", key)
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s successfully\n", key)
	return nil
}

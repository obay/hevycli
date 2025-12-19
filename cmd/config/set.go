package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/cmdutil"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/tui/prompt"
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
	Args: cmdutil.RequireArgs(2, "<key> <value>"),
	RunE: runSet,
}

func init() {
	setCmd.Flags().BoolVar(&validateKey, "validate", true,
		"validate API key before saving (for api-key only)")
}

func runSet(cmd *cobra.Command, args []string) error {
	var key, value string

	if len(args) >= 2 {
		key = strings.ToLower(args[0])
		value = args[1]
	} else {
		// Interactive mode - let user select key and enter value
		keyOptions := []prompt.SelectOption{
			{ID: "api-key", Title: "API Key", Description: "Your Hevy API key"},
			{ID: "default-output", Title: "Default Output", Description: "Output format (json, table, plain)"},
			{ID: "units", Title: "Units", Description: "Measurement units (metric, imperial)"},
			{ID: "color", Title: "Color", Description: "Enable/disable colored output"},
			{ID: "date-format", Title: "Date Format", Description: "Date format (Go format string)"},
			{ID: "time-format", Title: "Time Format", Description: "Time format (Go format string)"},
		}

		selected, err := prompt.Select("Select configuration key", keyOptions, "Choose a setting to configure...")
		if err != nil {
			return err
		}
		key = selected.ID

		// Get value based on key type
		switch key {
		case "default-output":
			outputOptions := []prompt.SelectOption{
				{ID: "table", Title: "Table", Description: "Formatted table output (default)"},
				{ID: "json", Title: "JSON", Description: "Raw JSON output"},
				{ID: "plain", Title: "Plain", Description: "Plain text output"},
			}
			selectedOutput, err := prompt.Select("Select output format", outputOptions, "Choose a format...")
			if err != nil {
				return err
			}
			value = selectedOutput.ID

		case "units":
			unitsOptions := []prompt.SelectOption{
				{ID: "metric", Title: "Metric", Description: "Kilograms and kilometers"},
				{ID: "imperial", Title: "Imperial", Description: "Pounds and miles"},
			}
			selectedUnits, err := prompt.Select("Select measurement units", unitsOptions, "Choose units...")
			if err != nil {
				return err
			}
			value = selectedUnits.ID

		case "color":
			colorOptions := []prompt.SelectOption{
				{ID: "true", Title: "Enabled", Description: "Show colored output"},
				{ID: "false", Title: "Disabled", Description: "Plain monochrome output"},
			}
			selectedColor, err := prompt.Select("Enable colored output?", colorOptions, "Choose...")
			if err != nil {
				return err
			}
			value = selectedColor.ID

		default:
			// Text input for other keys
			placeholder := "Enter value..."
			if key == "api-key" {
				placeholder = "Enter your Hevy API key..."
			} else if key == "date-format" {
				placeholder = "e.g., 2006-01-02"
			} else if key == "time-format" {
				placeholder = "e.g., 15:04"
			}

			value, err = prompt.TextInput("Enter "+key, placeholder, "enter to confirm")
			if err != nil {
				return err
			}
		}
	}

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

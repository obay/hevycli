package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/obay/hevycli/internal/config"
)

var showSecrets bool

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long: `Display the current hevycli configuration values.

By default, sensitive values like the API key are masked.
Use --show-secrets to reveal them.`,
	RunE: runShow,
}

func init() {
	showCmd.Flags().BoolVar(&showSecrets, "show-secrets", false,
		"show sensitive values like API key")
}

func runShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create a copy for display
	displayCfg := *cfg

	// Mask API key unless --show-secrets is used
	if !showSecrets && displayCfg.API.Key != "" {
		keyLen := len(displayCfg.API.Key)
		if keyLen > 4 {
			displayCfg.API.Key = "***" + displayCfg.API.Key[keyLen-4:]
		} else {
			displayCfg.API.Key = "****"
		}
	}

	out, err := yaml.Marshal(displayCfg)
	if err != nil {
		return fmt.Errorf("failed to format config: %w", err)
	}

	fmt.Printf("# Configuration file: %s\n", config.ConfigPath())
	fmt.Println()
	fmt.Print(string(out))

	return nil
}

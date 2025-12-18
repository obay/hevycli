package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config holds all configuration values
type Config struct {
	API     APIConfig     `mapstructure:"api" yaml:"api"`
	Display DisplayConfig `mapstructure:"display" yaml:"display"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	Key     string `mapstructure:"key" yaml:"key,omitempty"`
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
}

// DisplayConfig holds display-related configuration
type DisplayConfig struct {
	OutputFormat string `mapstructure:"output_format" yaml:"output_format"`
	Color        bool   `mapstructure:"color" yaml:"color"`
	Units        string `mapstructure:"units" yaml:"units"`
	DateFormat   string `mapstructure:"date_format" yaml:"date_format"`
	TimeFormat   string `mapstructure:"time_format" yaml:"time_format"`
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			BaseURL: "https://api.hevyapp.com/v1",
		},
		Display: DisplayConfig{
			OutputFormat: "table",
			Color:        true,
			Units:        "metric",
			DateFormat:   "2006-01-02",
			TimeFormat:   "15:04",
		},
	}
}

// ConfigDir returns the configuration directory path
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hevycli")
}

// ConfigPath returns the full path to config file
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// EnsureConfigDir creates config directory if it doesn't exist
func EnsureConfigDir() error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return nil
}

// Load reads configuration from file, environment, and returns a Config struct
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	// Set config file location
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(ConfigDir())
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Set defaults
	v.SetDefault("api.base_url", "https://api.hevyapp.com/v1")
	v.SetDefault("display.output_format", "table")
	v.SetDefault("display.color", true)
	v.SetDefault("display.units", "metric")
	v.SetDefault("display.date_format", "2006-01-02")
	v.SetDefault("display.time_format", "15:04")

	// Environment variable support
	v.SetEnvPrefix("HEVYCLI")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables
	_ = v.BindEnv("api.key", "HEVYCLI_API_KEY")
	_ = v.BindEnv("display.output_format", "HEVYCLI_OUTPUT_FORMAT")
	_ = v.BindEnv("display.units", "HEVYCLI_UNITS")
	_ = v.BindEnv("display.color", "HEVYCLI_COLOR")

	// Read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only return error if it's not a "file not found" error
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config: %w", err)
			}
		}
	}

	// Unmarshal into Config struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

// LoadFrom loads configuration from a specific file path
func LoadFrom(cfgFile string) (*Config, error) {
	return Load(cfgFile)
}

// Save writes the configuration to the default config path
func Save(cfg *Config) error {
	return SaveTo(cfg, ConfigPath())
}

// SaveTo writes the configuration to a specific file path
func SaveTo(cfg *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file with restricted permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetAPIKey returns the API key with precedence: env > config
func (c *Config) GetAPIKey() string {
	// Check environment variable first
	if key := os.Getenv("HEVYCLI_API_KEY"); key != "" {
		return key
	}
	return c.API.Key
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate output format
	switch c.Display.OutputFormat {
	case "json", "table", "plain":
		// Valid
	default:
		return fmt.Errorf("invalid output format: %s (must be json, table, or plain)", c.Display.OutputFormat)
	}

	// Validate units
	switch c.Display.Units {
	case "metric", "imperial":
		// Valid
	default:
		return fmt.Errorf("invalid units: %s (must be metric or imperial)", c.Display.Units)
	}

	return nil
}

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "https://api.hevyapp.com/v1", cfg.API.BaseURL)
	assert.Equal(t, "", cfg.API.Key)
	assert.Equal(t, "table", cfg.Display.OutputFormat)
	assert.Equal(t, "metric", cfg.Display.Units)
	assert.True(t, cfg.Display.Color)
	assert.Equal(t, "2006-01-02", cfg.Display.DateFormat)
	assert.Equal(t, "15:04", cfg.Display.TimeFormat)
}

func TestSaveAndLoad(t *testing.T) {
	// Create temp directory for test config
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	// Create config to save
	cfg := DefaultConfig()
	cfg.API.Key = "test-api-key-12345"
	cfg.Display.Units = "imperial"
	cfg.Display.OutputFormat = "json"

	// Save config
	err := SaveTo(cfg, cfgFile)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(cfgFile)
	require.NoError(t, err)

	// Load config
	loaded, err := LoadFrom(cfgFile)
	require.NoError(t, err)

	// Verify loaded values
	assert.Equal(t, "test-api-key-12345", loaded.API.Key)
	assert.Equal(t, "imperial", loaded.Display.Units)
	assert.Equal(t, "json", loaded.Display.OutputFormat)
	assert.Equal(t, "https://api.hevyapp.com/v1", loaded.API.BaseURL)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "invalid output format",
			config: &Config{
				Display: DisplayConfig{
					OutputFormat: "xml",
					Units:        "metric",
				},
			},
			expectError: true,
			errorMsg:    "invalid output format",
		},
		{
			name: "invalid units",
			config: &Config{
				Display: DisplayConfig{
					OutputFormat: "table",
					Units:        "stones",
				},
			},
			expectError: true,
			errorMsg:    "invalid units",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	// Test config key
	cfg := DefaultConfig()
	cfg.API.Key = "config-key"
	assert.Equal(t, "config-key", cfg.GetAPIKey())

	// Test env var takes precedence
	os.Setenv("HEVYCLI_API_KEY", "env-key")
	defer os.Unsetenv("HEVYCLI_API_KEY")

	assert.Equal(t, "env-key", cfg.GetAPIKey())
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	assert.Contains(t, dir, ".hevycli")
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	assert.Contains(t, path, ".hevycli")
	assert.Contains(t, path, "config.yaml")
}

func TestEnsureConfigDir(t *testing.T) {
	// This test uses the actual home directory
	// In a real test suite, you'd mock this
	err := EnsureConfigDir()
	assert.NoError(t, err)

	// Verify directory exists
	_, err = os.Stat(ConfigDir())
	assert.NoError(t, err)
}

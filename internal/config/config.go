package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	Provider   string            `yaml:"provider"`
	Model      string            `yaml:"model"`
	APIKeys    map[string]string `yaml:"api_keys"`
	Budgets    Budgets           `yaml:"budgets"`
	StorageDir string            `yaml:"storage_dir"`
}

// Budgets holds budget limits.
type Budgets struct {
	Monthly float64 `yaml:"monthly"`
	Daily   float64 `yaml:"daily"`
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	return &Config{
		Provider:   "openai",
		Model:      "gpt-3.5-turbo",
		APIKeys:    make(map[string]string),
		Budgets:    Budgets{Monthly: 10.0, Daily: 1.0},
		StorageDir: filepath.Join(home, ".mycli"),
	}
}

// Load loads the configuration from the default location (~/.config/mycli/config.yaml).
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(home, ".config", "mycli", "config.yaml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults for fields that are zero-value in the file
	// This ensures backward compatibility with older configs
	// and handles missing fields gracefully.
	home, err = os.UserHomeDir()
	if err != nil {
		home = ""
	}
	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-3.5-turbo"
	}
	if cfg.APIKeys == nil {
		cfg.APIKeys = make(map[string]string)
	}
	if cfg.StorageDir == "" {
		cfg.StorageDir = filepath.Join(home, ".mycli")
	}
	if cfg.Budgets.Monthly == 0 {
		cfg.Budgets.Monthly = 10.0
	}
	if cfg.Budgets.Daily == 0 {
		cfg.Budgets.Daily = 1.0
	}

	return &cfg, nil
}

// Save saves the configuration to the default location.
func (c *Config) Save() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".config", "mycli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
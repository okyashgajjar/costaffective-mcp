package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Provider != "openai" {
		t.Errorf("expected provider 'openai', got %s", cfg.Provider)
	}
	if cfg.Model != "gpt-3.5-turbo" {
		t.Errorf("expected model 'gpt-3.5-turbo', got %s", cfg.Model)
	}
	if cfg.APIKeys == nil {
		t.Error("expected APIKeys map to be initialized")
	}
	if cfg.Budgets.Monthly != 10.0 {
		t.Errorf("expected monthly budget 10.0, got %f", cfg.Budgets.Monthly)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.StorageDir = tmpDir
	cfg.Provider = "anthropic"
	cfg.Model = "claude-3-opus-20240229"
	cfg.APIKeys["anthropic"] = "sk-ant-test-key"

	// Override config path for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, ".config", "mycli", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config file was not created at %s", configPath)
	}

	// Load and verify
	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loadedCfg.Provider != "anthropic" {
		t.Errorf("expected provider 'anthropic', got %s", loadedCfg.Provider)
	}
	if loadedCfg.Model != "claude-3-opus-20240229" {
		t.Errorf("expected model 'claude-3-opus-20240229', got %s", loadedCfg.Model)
	}
	if loadedCfg.APIKeys["anthropic"] != "sk-ant-test-key" {
		t.Errorf("expected API key for anthropic, got %s", loadedCfg.APIKeys["anthropic"])
	}
}

func TestLoadNonExistentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load non-existent config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config to be returned, got nil")
	}
}

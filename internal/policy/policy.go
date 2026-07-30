package policy

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ValidationRule struct {
	Level   string   `yaml:"level"`   // e.g., "error", "warn", "info"
	Match   []string `yaml:"match"`   // Globs for files
	Require []string `yaml:"require"` // RegEx patterns that must be present
	Deny    []string `yaml:"deny"`    // RegEx patterns that must NOT be present
	Message string   `yaml:"message"`
}

type Category struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Patterns    []string         `yaml:"patterns"` // Globs to define this category
	Rules       []ValidationRule `yaml:"rules"`
}

type Policy struct {
	Version     string     `yaml:"version"`
	GlobalRules []string   `yaml:"global_rules"` // Path to shared global rule files (optional)
	Categories  []Category `yaml:"categories"`
}

// ParsePolicy looks for .costwise.yaml in the repo root and parses it.
func ParsePolicy(repoRoot string) (*Policy, error) {
	policyPath := filepath.Join(repoRoot, ".costwise.yaml")
	data, err := os.ReadFile(policyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No policy found, not an error
		}
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse policy yaml: %w", err)
	}

	return &p, nil
}

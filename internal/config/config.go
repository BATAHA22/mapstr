package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Language    string   `yaml:"language"`
	Output      []string `yaml:"output"`
	AI          AIConfig `yaml:"ai"`
	Depth       int      `yaml:"depth"`
	Incremental bool     `yaml:"incremental"`
	Ignore      []string `yaml:"ignore"`
}

type AIConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	Fallback string `yaml:"fallback"`
	NoAI     bool   `yaml:"no_ai"`
}

func DefaultConfig() *Config {
	return &Config{
		Language: "en",
		Output:   []string{"md", "mermaid", "json"},
		AI: AIConfig{
			Provider: "",
			Model:    "",
			Fallback: "ollama",
			NoAI:     false,
		},
		Depth:       3,
		Incremental: true,
		Ignore: []string{
			// Version control & dependencies
			".git",
			"node_modules",
			"vendor",
			// Python environments
			"__pycache__",
			"venv",
			"env",
			".venv",
			// Build output
			"dist",
			"build",
			".next",
			".nuxt",
			// Framework-specific
			"__init__.py",
			"manage.py",
			// Mapstr own output
			"mapstr",
			".mapstr-cache.json",
			// Minified & lock files
			"*.min.js",
			"*.min.css",
			"*.lock",
			// Django migrations (auto-generated)
			"[0-9][0-9][0-9][0-9]_*.py",
		},
	}
}

// Load reads a .mapstr.yml from the given directory or its parents.
// If no config file is found, it returns the default config.
func Load(dir string) (*Config, error) {
	cfg := DefaultConfig()

	path, err := findConfigFile(dir)
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	return cfg, nil
}

func findConfigFile(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(abs, ".mapstr.yml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}

	return "", fmt.Errorf("config: .mapstr.yml not found")
}

// ShouldIgnore checks if a path matches any ignore pattern.
func (c *Config) ShouldIgnore(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range c.Ignore {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		if base == pattern {
			return true
		}
	}
	return false
}

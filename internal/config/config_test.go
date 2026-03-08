package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Language != "en" {
		t.Errorf("default language = %q, want %q", cfg.Language, "en")
	}

	if cfg.Depth != 3 {
		t.Errorf("default depth = %d, want 3", cfg.Depth)
	}

	if len(cfg.Output) != 3 {
		t.Errorf("default output formats = %d, want 3", len(cfg.Output))
	}

	if len(cfg.Ignore) == 0 {
		t.Error("default ignore list should not be empty")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()

	yamlContent := `language: ar
output:
  - md
ai:
  provider: openai
  model: gpt-4o
depth: 5
ignore:
  - node_modules
  - dist
`
	os.WriteFile(filepath.Join(dir, ".mapstr.yml"), []byte(yamlContent), 0644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Language != "ar" {
		t.Errorf("language = %q, want %q", cfg.Language, "ar")
	}

	if cfg.AI.Provider != "openai" {
		t.Errorf("provider = %q, want %q", cfg.AI.Provider, "openai")
	}

	if cfg.AI.Model != "gpt-4o" {
		t.Errorf("model = %q, want %q", cfg.AI.Model, "gpt-4o")
	}

	if cfg.Depth != 5 {
		t.Errorf("depth = %d, want 5", cfg.Depth)
	}
}

func TestLoadNoFile(t *testing.T) {
	dir := t.TempDir()

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Should return defaults
	if cfg.Language != "en" {
		t.Errorf("language = %q, want default %q", cfg.Language, "en")
	}
}

func TestShouldIgnore(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		path   string
		ignore bool
	}{
		{"node_modules", true},
		{".git", true},
		{"dist", true},
		{"vendor", true},
		{"__pycache__", true},
		{"src", false},
		{"main.go", false},
		{"internal/parser", false},
	}

	for _, tt := range tests {
		got := cfg.ShouldIgnore(tt.path)
		if got != tt.ignore {
			t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.ignore)
		}
	}
}

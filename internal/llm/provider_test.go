package llm

import (
	"context"
	"testing"
)

func TestResolveExplicitProvider(t *testing.T) {
	providers := []string{"claude", "openai", "gemini", "deepseek", "mistral", "ollama"}

	for _, name := range providers {
		p, err := Resolve(name, "")
		if err != nil {
			t.Errorf("Resolve(%q) failed: %v", name, err)
			continue
		}
		if p.Name() != name {
			t.Errorf("Resolve(%q).Name() = %q", name, p.Name())
		}
	}
}

func TestResolveUnknownProvider(t *testing.T) {
	_, err := Resolve("nonexistent", "")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestSetModel(t *testing.T) {
	p, err := Resolve("openai", "gpt-4-turbo")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	// We can't easily check the model was set without exposing it,
	// but we can verify the provider was created without error.
	if p.Name() != "openai" {
		t.Errorf("expected openai, got %s", p.Name())
	}
}

func TestProviderAvailability(t *testing.T) {
	// Without API keys set, cloud providers should not be available.
	// We don't unset env vars here to avoid interfering with a real environment,
	// but we verify the Available() method returns a bool.
	p := NewClaude()
	_ = p.Available() // should not panic

	p = NewOpenAI()
	_ = p.Available()

	p = NewGemini()
	_ = p.Available()
}

func TestBuildPrompt(t *testing.T) {
	prompt := BuildPrompt("test-project", nil, "no graph", "en")

	if prompt == "" {
		t.Error("prompt should not be empty")
	}

	if !contains(prompt, "test-project") {
		t.Error("prompt should contain project name")
	}

	if !contains(prompt, "en") {
		t.Error("prompt should contain language")
	}
}

func TestSummarizeWithFallbackPrimaryFails(t *testing.T) {
	// Test that fallback to empty name produces an error
	primary := &mockProvider{
		name: "mock",
		err:  context.DeadlineExceeded,
	}

	_, _, err := SummarizeWithFallback(context.Background(), primary, "", "test prompt")
	if err == nil {
		t.Error("expected error when primary fails and no fallback")
	}
}

type mockProvider struct {
	name   string
	result string
	err    error
}

func (m *mockProvider) Name() string                                     { return m.name }
func (m *mockProvider) Available() bool                                  { return true }
func (m *mockProvider) Summarize(_ context.Context, _ string) (string, *Usage, error) {
	return m.result, nil, m.err
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

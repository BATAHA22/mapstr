package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BATAHA22/mapstr/internal/parser"
)

// Provider is the interface all LLM backends must implement.
type Provider interface {
	Name() string
	Summarize(ctx context.Context, prompt string) (string, *Usage, error)
	Available() bool
}

// registry of provider constructors in auto-detection order.
var providerOrder = []struct {
	name    string
	factory func() Provider
}{
	{"claude", NewClaude},
	{"openai", NewOpenAI},
	{"gemini", NewGemini},
	{"deepseek", NewDeepSeek},
	{"mistral", NewMistral},
	{"ollama", NewOllama},
}

// Resolve returns the appropriate provider based on explicit choice or auto-detection.
func Resolve(providerName, model string) (Provider, error) {
	if providerName != "" {
		return resolveExplicit(providerName, model)
	}
	return autoDetect(model)
}

func resolveExplicit(name, model string) (Provider, error) {
	for _, entry := range providerOrder {
		if entry.name == name {
			p := entry.factory()
			if model != "" {
				if setter, ok := p.(interface{ SetModel(string) }); ok {
					setter.SetModel(model)
				}
			}
			return p, nil
		}
	}
	return nil, fmt.Errorf("llm: unknown provider %q", name)
}

func autoDetect(model string) (Provider, error) {
	for _, entry := range providerOrder {
		p := entry.factory()
		if p.Available() {
			if model != "" {
				if setter, ok := p.(interface{ SetModel(string) }); ok {
					setter.SetModel(model)
				}
			}
			return p, nil
		}
	}
	return nil, fmt.Errorf("llm: no provider available — set an API key or use --no-ai")
}

// SummarizeWithFallback tries the primary provider, then the fallback.
func SummarizeWithFallback(ctx context.Context, primary Provider, fallbackName, prompt string) (string, *Usage, error) {
	result, usage, err := primary.Summarize(ctx, prompt)
	if err == nil && strings.TrimSpace(result) != "" {
		return result, usage, nil
	}

	if fallbackName == "" {
		return "", nil, fmt.Errorf("llm: primary provider %s failed: %w", primary.Name(), err)
	}

	fb, fbErr := resolveExplicit(fallbackName, "")
	if fbErr != nil {
		return "", nil, fmt.Errorf("llm: primary %s failed (%w), fallback %q not found", primary.Name(), err, fallbackName)
	}

	result, usage, fbErr = fb.Summarize(ctx, prompt)
	if fbErr != nil {
		return "", nil, fmt.Errorf("llm: primary %s and fallback %s both failed: %w", primary.Name(), fb.Name(), fbErr)
	}

	return result, usage, nil
}

// DefaultTimeout for LLM API calls.
const DefaultTimeout = 60 * time.Second

// BuildPrompt constructs the summarization prompt from parsed data.
func BuildPrompt(projectName string, nodes []*parser.FileNode, graphSummary, lang string) string {
	var b strings.Builder

	languages := map[string]int{}
	var entryPoints []string
	var funcSummary strings.Builder

	for _, n := range nodes {
		languages[n.Language]++
		if isEntryPoint(n.Path) {
			entryPoints = append(entryPoints, n.Path)
		}
		for _, f := range n.Functions {
			if f.Exported {
				fmt.Fprintf(&funcSummary, "- %s: %s()\n", n.Path, f.Name)
			}
		}
	}

	var langList []string
	for l, count := range languages {
		langList = append(langList, fmt.Sprintf("%s (%d files)", l, count))
	}

	fmt.Fprintf(&b, "You are analyzing a software project called %q.\n", projectName)
	fmt.Fprintf(&b, "Below is the structural analysis of the codebase:\n\n")
	fmt.Fprintf(&b, "Files analyzed: %d\n", len(nodes))
	fmt.Fprintf(&b, "Languages: %s\n", strings.Join(langList, ", "))
	if len(entryPoints) > 0 {
		fmt.Fprintf(&b, "Entry points: %s\n", strings.Join(entryPoints, ", "))
	}
	fmt.Fprintf(&b, "\nFile structure and relationships:\n%s\n", graphSummary)
	fmt.Fprintf(&b, "\nKey functions and modules:\n%s\n", funcSummary.String())
	fmt.Fprintf(&b, `Generate a clear, concise architecture overview with these sections:
1. Overview (what this project does, 2-3 sentences)
2. Architecture (how it's structured, main layers)
3. Entry Points (where to start reading the code)
4. Data Flow (how data moves through the system)
5. Key Dependencies (main external libraries and their purpose)

Be specific, technical, and helpful to a developer who is new to this codebase.
Output language: %s
`, lang)

	return b.String()
}

func isEntryPoint(path string) bool {
	base := strings.ToLower(path)
	entryNames := []string{
		"main.go", "main.py", "index.js", "index.ts", "app.go",
		"app.js", "app.ts", "server.go", "server.js", "server.ts",
		"cmd/", "src/main",
	}
	for _, name := range entryNames {
		if strings.Contains(base, name) {
			return true
		}
	}
	return false
}

// envKey returns an environment variable value or empty string.
func envKey(key string) string {
	return os.Getenv(key)
}

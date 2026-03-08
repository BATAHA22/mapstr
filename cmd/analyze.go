package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/BATAHA22/mapstr/internal/config"
	gitutil "github.com/BATAHA22/mapstr/internal/git"
	"github.com/BATAHA22/mapstr/internal/graph"
	"github.com/BATAHA22/mapstr/internal/llm"
	"github.com/BATAHA22/mapstr/internal/output"
	"github.com/BATAHA22/mapstr/internal/parser"
	"github.com/BATAHA22/mapstr/internal/ui"
)

func runAnalyze(cmd *cobra.Command, args []string) error {
	if flagMCP {
		return runMCPServer(cmd, args)
	}

	totalStart := time.Now()

	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Validate the path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path %q: %w", absPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", absPath)
	}

	projectName := filepath.Base(absPath)

	// Load config
	cfg, err := config.Load(absPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	// Apply CLI flag overrides
	applyOverrides(cfg)

	// ── Pre-flight: Validate LLM provider before doing any work ───────
	if !cfg.AI.NoAI {
		_, err := llm.CheckAvailability(cfg.AI.Provider, cfg.AI.Model)
		if setupErr, ok := err.(*llm.SetupError); ok {
			printSetupHelp(setupErr, cfg.AI.Provider)
			return fmt.Errorf("provider not configured")
		} else if err != nil {
			return err
		}
	}

	fmt.Printf("\n  Analyzing %s...\n\n", projectName)

	// ── Phase 1: Parse project (with spinner) ──────────────────────────
	sp := ui.NewSpinner("Scanning project structure...")
	sp.Start()

	nodes, err := parseProject(absPath, cfg)
	sp.Stop()
	if err != nil {
		return err
	}
	fmt.Printf("  ✓ Parsed %d files\n", len(nodes))

	if len(nodes) == 0 {
		return fmt.Errorf("no supported source files found in %s", absPath)
	}

	// ── Phase 2: Build dependency graph (with spinner) ─────────────────
	sp = ui.NewSpinner("Building dependency graph...")
	sp.Start()

	g := graph.Build(nodes)
	sp.Stop()
	fmt.Printf("  ✓ Built graph: %d nodes, %d edges\n", len(g.Nodes), len(g.Edges))

	// ── Phase 3: AI summarization (with spinner) ───────────────────────
	var aiSummary string
	var usage *llm.Usage
	llmFailed := false
	if !cfg.AI.NoAI {
		aiSummary, usage, err = summarize(cfg, projectName, nodes, g)
		if err != nil {
			llmFailed = true
			fmt.Fprintf(os.Stderr, "\n  ⚠ LLM call failed: %v\n", err)
			fmt.Fprintln(os.Stderr, "  Generating structural-only output (no AI summary).")
		} else {
			fmt.Printf("  ✓ AI summary generated (%s)\n", usage.Provider)
		}
	}

	// ── Phase 4: Generate outputs ──────────────────────────────────────
	outDir := flagOutDir
	if outDir == "" {
		outDir = filepath.Join(absPath, "mapstr")
	}
	if err := os.MkdirAll(outDir, 0750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	outRel, _ := filepath.Rel(absPath, outDir)
	if outRel == "" {
		outRel = outDir
	}

	var outputFiles []string

	outputFormats := cfg.Output
	for _, format := range outputFormats {
		switch strings.ToLower(format) {
		case "md", "markdown":
			content := output.GenerateMarkdown(projectName, nodes, g, aiSummary)
			path := filepath.Join(outDir, "CONTEXT.md")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				return fmt.Errorf("write CONTEXT.md: %w", err)
			}
			outputFiles = append(outputFiles, "CONTEXT.md")

		case "mermaid", "mmd":
			content := output.GenerateMermaid(g)
			path := filepath.Join(outDir, "GRAPH.mmd")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				return fmt.Errorf("write GRAPH.mmd: %w", err)
			}
			outputFiles = append(outputFiles, "GRAPH.mmd")

		case "json":
			content, jsonErr := output.GenerateJSON(projectName, nodes, g, aiSummary)
			if jsonErr != nil {
				return fmt.Errorf("generate JSON: %w", jsonErr)
			}
			path := filepath.Join(outDir, "context.json")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				return fmt.Errorf("write context.json: %w", err)
			}
			outputFiles = append(outputFiles, "context.json")

		default:
			fmt.Fprintf(os.Stderr, "  ⚠ Unknown output format %q\n", format)
		}
	}

	// Save cache for incremental mode
	if cfg.Incremental && gitutil.IsGitRepo(absPath) {
		if err := gitutil.SaveCache(absPath, nodes); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ Could not save cache: %v\n", err)
		}
	}

	// ── Print summary ──────────────────────────────────────────────────
	if llmFailed {
		// Partial success: LLM failed but structural output was generated
		fmt.Printf("  ⚠ Partial output written to %s/ (no AI summary)\n", outRel)
		fmt.Println("  Tip: check your API key or use --no-ai to skip LLM.")
	} else {
		costStr := "N/A"
		if usage != nil {
			costStr = usage.CostString()
		}

		ui.PrintSummary(ui.Summary{
			Duration:    time.Since(totalStart),
			CostStr:     costStr,
			Provider:    providerName(usage),
			OutputDir:   outRel,
			OutputFiles: outputFiles,
			FileCount:   len(nodes),
			NodeCount:   len(g.Nodes),
			EdgeCount:   len(g.Edges),
			NoAI:        cfg.AI.NoAI,
		})
	}

	return nil
}

func providerName(u *llm.Usage) string {
	if u == nil {
		return ""
	}
	return u.Provider
}

func printSetupHelp(e *llm.SetupError, requestedProvider string) {
	fmt.Fprintln(os.Stderr)

	if e.Provider == "" {
		// Auto-detect failed: no providers configured at all
		fmt.Fprintln(os.Stderr, "  ⚠️  No LLM provider configured")
		fmt.Fprintln(os.Stderr, "  ℹ️  Set an API key for any supported provider:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "    export ANTHROPIC_API_KEY=\"sk-...\"   # Claude")
		fmt.Fprintln(os.Stderr, "    export OPENAI_API_KEY=\"sk-...\"      # OpenAI")
		fmt.Fprintln(os.Stderr, "    export GEMINI_API_KEY=\"...\"         # Gemini")
		fmt.Fprintln(os.Stderr, "    export MISTRAL_API_KEY=\"...\"        # Mistral")
		fmt.Fprintln(os.Stderr, "    export DEEPSEEK_API_KEY=\"...\"       # DeepSeek")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Or run without AI:")
		fmt.Fprintln(os.Stderr, "    mapstr --no-ai")
		fmt.Fprintln(os.Stderr)
		return
	}

	// Specific provider missing
	providerFlag := ""
	if requestedProvider != "" {
		providerFlag = " --provider " + requestedProvider
	}

	fmt.Fprintf(os.Stderr, "  ⚠️  No API key found for %s", e.Provider)
	if requestedProvider != "" {
		fmt.Fprintf(os.Stderr, " (--provider %s)", requestedProvider)
	}
	fmt.Fprintln(os.Stderr)

	fmt.Fprintf(os.Stderr, "  ℹ️  Set %s or use --no-ai\n", e.EnvVar)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  Quick start:")
	fmt.Fprintf(os.Stderr, "    export %s=\"your-key-here\"\n", e.EnvVar)
	fmt.Fprintf(os.Stderr, "    mapstr%s\n", providerFlag)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  Or run without AI:")
	fmt.Fprintln(os.Stderr, "    mapstr --no-ai")
	fmt.Fprintln(os.Stderr)
}

func parseProject(absPath string, cfg *config.Config) ([]*parser.FileNode, error) {
	// Check for incremental mode
	if cfg.Incremental && gitutil.IsGitRepo(absPath) {
		changed, err := gitutil.ChangedFiles(absPath)
		if err == nil && len(changed) > 0 {
			cached, cacheErr := gitutil.LoadCache(absPath)
			if cacheErr == nil && len(cached) > 0 {
				fmt.Printf("  Incremental mode: %d changed files\n", len(changed))
				freshNodes, parseErr := parser.ParseProject(absPath, cfg)
				if parseErr != nil {
					return nil, fmt.Errorf("parse: %w", parseErr)
				}
				return gitutil.MergeNodes(cached, freshNodes), nil
			}
		}
	}

	// Full parse
	nodes, err := parser.ParseProject(absPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return nodes, nil
}

func summarize(cfg *config.Config, projectName string, nodes []*parser.FileNode, g *graph.DependencyGraph) (string, *llm.Usage, error) {
	provider, err := llm.Resolve(cfg.AI.Provider, cfg.AI.Model)
	if err != nil {
		return "", nil, err
	}

	sp := ui.NewSpinner(fmt.Sprintf("Generating AI summary (%s)...", provider.Name()))
	sp.Start()
	defer sp.Stop()

	prompt := llm.BuildPrompt(projectName, nodes, g.Summary(), cfg.Language)

	ctx, cancel := context.WithTimeout(context.Background(), llm.DefaultTimeout)
	defer cancel()

	result, usage, err := llm.SummarizeWithFallback(ctx, provider, cfg.AI.Fallback, prompt)
	if err != nil {
		return "", nil, err
	}

	return result, usage, nil
}

func applyOverrides(cfg *config.Config) {
	if flagLang != "en" || cfg.Language == "" {
		cfg.Language = flagLang
	}

	if len(flagOutput) > 0 {
		cfg.Output = flagOutput
	}

	if flagProvider != "" {
		cfg.AI.Provider = flagProvider
	}
	if flagModel != "" {
		cfg.AI.Model = flagModel
	}
	if flagNoAI {
		cfg.AI.NoAI = true
	}
	if flagDepth != 3 || cfg.Depth == 0 {
		cfg.Depth = flagDepth
	}
}

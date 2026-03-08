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
	if !cfg.AI.NoAI {
		aiSummary, usage, err = summarize(cfg, projectName, nodes, g)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ LLM summarization failed: %v\n", err)
			fmt.Fprintln(os.Stderr, "  Falling back to structural-only output.")
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

	// ── Print summary box ──────────────────────────────────────────────
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

	return nil
}

func providerName(u *llm.Usage) string {
	if u == nil {
		return ""
	}
	return u.Provider
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

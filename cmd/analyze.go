package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mapstr/mapstr/internal/config"
	gitutil "github.com/mapstr/mapstr/internal/git"
	"github.com/mapstr/mapstr/internal/graph"
	"github.com/mapstr/mapstr/internal/llm"
	"github.com/mapstr/mapstr/internal/output"
	"github.com/mapstr/mapstr/internal/parser"
)

func runAnalyze(cmd *cobra.Command, args []string) error {
	if flagMCP {
		return runMCPServer(cmd, args)
	}

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

	fmt.Printf("Analyzing %s...\n", projectName)

	// Parse the project
	start := time.Now()
	nodes, err := parseProject(absPath, cfg)
	if err != nil {
		return err
	}
	fmt.Printf("  Parsed %d files in %s\n", len(nodes), time.Since(start).Round(time.Millisecond))

	if len(nodes) == 0 {
		return fmt.Errorf("no supported source files found in %s", absPath)
	}

	// Build dependency graph
	g := graph.Build(nodes)
	fmt.Printf("  Built graph: %d nodes, %d edges\n", len(g.Nodes), len(g.Edges))

	// AI summarization
	var aiSummary string
	if !cfg.AI.NoAI {
		aiSummary, err = summarize(cfg, projectName, nodes, g)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: LLM summarization failed: %v\n", err)
			fmt.Fprintln(os.Stderr, "  Falling back to structural-only output.")
		}
	}

	// Generate outputs
	outDir := flagOutDir
	if outDir == "" {
		outDir = filepath.Join(absPath, "mapstr")
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	outputFormats := cfg.Output
	for _, format := range outputFormats {
		switch strings.ToLower(format) {
		case "md", "markdown":
			content := output.GenerateMarkdown(projectName, nodes, g, aiSummary)
			path := filepath.Join(outDir, "CONTEXT.md")
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("write CONTEXT.md: %w", err)
			}
			fmt.Printf("  Written: %s\n", path)

		case "mermaid", "mmd":
			content := output.GenerateMermaid(g)
			path := filepath.Join(outDir, "GRAPH.mmd")
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("write GRAPH.mmd: %w", err)
			}
			fmt.Printf("  Written: %s\n", path)

		case "json":
			content, jsonErr := output.GenerateJSON(projectName, nodes, g, aiSummary)
			if jsonErr != nil {
				return fmt.Errorf("generate JSON: %w", jsonErr)
			}
			path := filepath.Join(outDir, "context.json")
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("write context.json: %w", err)
			}
			fmt.Printf("  Written: %s\n", path)

		default:
			fmt.Fprintf(os.Stderr, "  Warning: unknown output format %q\n", format)
		}
	}

	// Save cache for incremental mode
	if cfg.Incremental && gitutil.IsGitRepo(absPath) {
		if err := gitutil.SaveCache(absPath, nodes); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not save cache: %v\n", err)
		}
	}

	fmt.Println("Done.")
	return nil
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

func summarize(cfg *config.Config, projectName string, nodes []*parser.FileNode, g *graph.DependencyGraph) (string, error) {
	provider, err := llm.Resolve(cfg.AI.Provider, cfg.AI.Model)
	if err != nil {
		return "", err
	}

	fmt.Printf("  Using LLM: %s\n", provider.Name())

	prompt := llm.BuildPrompt(projectName, nodes, g.Summary(), cfg.Language)

	ctx, cancel := context.WithTimeout(context.Background(), llm.DefaultTimeout)
	defer cancel()

	result, err := llm.SummarizeWithFallback(ctx, provider, cfg.AI.Fallback, prompt)
	if err != nil {
		return "", err
	}

	return result, nil
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

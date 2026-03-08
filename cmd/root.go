package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagLang     string
	flagOutput   []string
	flagProvider string
	flagModel    string
	flagNoAI     bool
	flagDepth    int
	flagWatch    bool
	flagMCP      bool
	flagConfig   string
	flagOutDir   string

	Version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "mapstr [path]",
	Short: "AI-Powered Codebase Navigator",
	Long: `Mapstr analyzes any software project and generates a human-readable
map of the codebase using structural parsing and multi-LLM summarization.

  mapstr ./my-project
  mapstr ./my-project --provider openai --model gpt-4o
  mapstr ./my-project --no-ai

Output files (written to <project>/mapstr/):
  CONTEXT.md   — Natural-language architecture overview
  GRAPH.mmd    — Mermaid dependency diagram
  context.json — Structured data for AI assistants`,
	Args:    cobra.MaximumNArgs(1),
	Version: Version,
	RunE:    runAnalyze,
}

func init() {
	rootCmd.Flags().StringVarP(&flagLang, "lang", "l", "en", "Output language (en, ar, es, fr, de, zh, ja, ...)")
	rootCmd.Flags().StringSliceVarP(&flagOutput, "output", "o", []string{"md", "mermaid", "json"}, "Output formats")
	rootCmd.Flags().StringVarP(&flagProvider, "provider", "p", "", "LLM provider: claude, openai, gemini, ollama, deepseek, mistral")
	rootCmd.Flags().StringVarP(&flagModel, "model", "m", "", "Model name (uses provider default if not set)")
	rootCmd.Flags().BoolVar(&flagNoAI, "no-ai", false, "Skip LLM — structural analysis + graph only")
	rootCmd.Flags().IntVarP(&flagDepth, "depth", "d", 3, "Dependency tree depth")
	rootCmd.Flags().BoolVarP(&flagWatch, "watch", "w", false, "Watch mode — regenerate on file changes")
	rootCmd.Flags().BoolVar(&flagMCP, "mcp", false, "Start as MCP server")
	rootCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "Path to .mapstr.yml")
	rootCmd.Flags().StringVar(&flagOutDir, "out-dir", "", "Output directory (default: <project>/mapstr/)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

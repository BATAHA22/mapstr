package cmd

import (
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
	Args:          cobra.MaximumNArgs(1),
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runAnalyze,
}

const versionTemplate = `
  __  __    _    ____  ____ _____ ____
 |  \/  |  / \  |  _ \/ ___|_   _|  _ \
 | |\/| | / _ \ | |_) \___ \ | | | |_) |
 | |  | |/ ___ \|  __/ ___) || | |  _ <
 |_|  |_/_/   \_\_|   |____/ |_| |_| \_\

 Version : {{.Version}}
 Author  : Taha Ben Ali
 License : MIT
 Repo    : github.com/BATAHA22/mapstr
 Tagline : "Don't read the code. Map it."
`

func init() {
	rootCmd.SetVersionTemplate(versionTemplate)
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
		// Errors already printed by runAnalyze; just exit with code 1.
		os.Exit(1)
	}
}

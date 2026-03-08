# Mapstr

> AI-Powered Codebase Navigator — Understand any project in seconds.

**"Don't read the code. Map it."**

Mapstr is a single-binary CLI that analyzes any software project and generates a human-readable map of the codebase using structural parsing and multi-LLM summarization.

## Quick Start

```bash
# Analyze current directory (auto-detects LLM provider)
mapstr .

# Use a specific provider
mapstr ./my-project --provider claude
mapstr ./my-project --provider openai --model gpt-4o
mapstr ./my-project --provider gemini --model gemini-2.5-pro

# Use a local model (fully offline)
mapstr ./my-project --provider ollama --model llama3

# Structural analysis only (no AI, no API calls)
mapstr ./my-project --no-ai
```

## Output

One command generates three files:

| File | Purpose |
|------|---------|
| `CONTEXT.md` | Natural-language architecture overview |
| `GRAPH.mmd` | Mermaid dependency diagram |
| `context.json` | Structured data for AI assistants |

## Installation

```bash
# Go
go install github.com/mapstr/mapstr@latest

# Homebrew
brew install mapstr/tap/mapstr

# Pre-built binaries
# Download from GitHub Releases
```

## Supported Languages

| Language | Status |
|----------|--------|
| Go | Stable |
| JavaScript | Stable |
| TypeScript | Stable |
| Python | Stable |

## Supported LLM Providers

| Provider | Env Variable | Default Model |
|----------|-------------|---------------|
| Claude (Anthropic) | `ANTHROPIC_API_KEY` | claude-sonnet-4-5 |
| OpenAI | `OPENAI_API_KEY` | gpt-4o |
| Gemini (Google) | `GEMINI_API_KEY` | gemini-2.5-pro |
| DeepSeek | `DEEPSEEK_API_KEY` | deepseek-chat |
| Mistral | `MISTRAL_API_KEY` | mistral-large-latest |
| Ollama (local) | — | llama3 |

Provider auto-detection: if no `--provider` flag is set, Mapstr checks for API keys in the order above and uses the first available provider.

## Configuration

Create a `.mapstr.yml` in your project root:

```yaml
language: en
output:
  - md
  - mermaid
  - json
ai:
  provider: claude
  model: claude-sonnet-4-5
  fallback: ollama
depth: 3
incremental: true
ignore:
  - node_modules
  - .git
  - dist
  - vendor
```

## CLI Flags

```
Usage:
  mapstr [path] [flags]

Flags:
  -l, --lang string        Output language (default "en")
  -o, --output strings     Output formats (default [md,mermaid,json])
  -p, --provider string    LLM provider: claude, openai, gemini, ollama, deepseek, mistral
  -m, --model string       Model name
      --no-ai              Structural analysis only (no LLM calls)
  -d, --depth int          Dependency tree depth (default 3)
  -w, --watch              Watch mode
      --mcp                Start as MCP server
  -c, --config string      Config file path
      --out-dir string     Output directory (default ".")
  -h, --help               Help
  -v, --version            Version
```

## MCP Server

Mapstr can run as an MCP server for Claude Code, Cursor, or other MCP-compatible tools:

```json
{
  "mcpServers": {
    "mapstr": {
      "command": "mapstr",
      "args": ["--mcp"]
    }
  }
}
```

## GitHub Actions

```yaml
- name: Generate Codebase Context
  uses: mapstr/mapstr-action@v1
  with:
    lang: en
    output: md
```

## How It Works

```
codebase
   |
   v
[Language Parsers]     -- Go / JS / TS / Python AST parsing
   |
   v
[Dependency Resolver]  -- Resolves imports and cross-file references
   |
   v
[Graph Builder]        -- Builds relationship graph
   |
   v
[LLM Summarizer]      -- Claude / OpenAI / Gemini / Ollama / ...
   |
   v
[Output Engine]
   +-- CONTEXT.md
   +-- GRAPH.mmd
   +-- context.json
```

## License

MIT

# Mapstr

> AI-Powered Codebase Navigator — Understand any project in seconds.

---

## The Problem

Every developer wastes hours — sometimes days — trying to understand a new codebase before writing a single line of code. Existing tools either serve only AI assistants (Repomix, kit), are overly complex (CodePrism in Rust), or are limited in language support and output formats.

**There is no simple, universal tool that gives a developer an instant, human-readable map of any codebase.**

---

## The Solution

A single-binary CLI written in Go that analyzes any software project and generates a complete understanding in seconds — using any LLM provider you choose:

```bash
# Use Claude (default)
mapstr ./my-project --lang en --output md

# Use OpenAI
mapstr ./my-project --provider openai --model gpt-4o

# Use Gemini
mapstr ./my-project --provider gemini --model gemini-2.5-pro

# Use a local model via Ollama
mapstr ./my-project --provider ollama --model llama3

# No AI — just structural analysis + graph (zero API calls)
mapstr ./my-project --no-ai
```

### Three outputs, one command:

| Output         | Purpose                                                  |
|----------------|----------------------------------------------------------|
| `CONTEXT.md`   | A natural-language narrative explaining the project's architecture, modules, and data flow |
| `GRAPH.mmd`    | A Mermaid diagram visualizing relationships between files, modules, and functions |
| `context.json` | Structured data optimized for AI assistants (Claude, Cursor, Copilot) |

---

## How It Works

```
codebase
   |
   v
[Tree-sitter Parser]        <-- Parses JS / TS / Go / Python / Rust / Java / ...
   |
   v
[Dependency Resolver]        <-- Resolves imports, exports, and cross-file references
   |
   v
[Graph Builder]              <-- Builds a relationship graph of functions / modules / APIs
   |
   v
[LLM Summarizer]             <-- Claude / OpenAI / Gemini / Ollama / DeepSeek / Mistral
   |
   v
[Output Engine]
   +-- CONTEXT.md
   +-- GRAPH.mmd (Mermaid)
   +-- context.json
```

### Key Technical Details

- **Tree-sitter** provides language-agnostic AST parsing — no custom parsers per language
- **Graph construction** identifies: function calls, module imports, API routes, type dependencies
- **Unified LLM interface** — a single provider abstraction supports multiple backends:
  - **Claude** (Anthropic) — default, best for nuanced architecture summaries
  - **GPT-4o / GPT-4.1** (OpenAI) — widely available, strong general performance
  - **Gemini 2.5 Pro** (Google) — large context window, good for monorepos
  - **DeepSeek R1 / V3** — cost-effective, strong reasoning
  - **Mistral Large** — EU-hosted option, good multilingual support
  - **Ollama** (local) — fully offline, zero API costs, privacy-first
- **`--no-ai` mode**: generates structural analysis and dependency graphs without any LLM calls — useful for air-gapped environments or when you just need the graph
- **Incremental mode**: only re-analyzes changed files (via git diff integration)
- **Provider auto-detection**: reads `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`, etc. — picks the first available provider if none is specified

---

## Competitive Advantage

| Feature                            | Mapstr  | Repomix   | CodePrism  |
|------------------------------------|-------------|-----------|------------|
| Multi-LLM support                 | Yes (6+)    | No        | No         |
| Offline / no-AI mode              | Yes         | No        | No         |
| Visual Mermaid dependency graph   | Yes         | No        | Complex    |
| Natural-language summary (i18n)   | Yes         | No        | No         |
| Simple single-binary CLI          | Yes (Go)    | Yes (Node)| No (Rust)  |
| Serves the developer directly     | Yes         | No        | No         |
| MCP Server (Claude/Cursor)        | Yes         | No        | Yes        |
| GitHub App integration            | Yes         | No        | No         |
| Incremental analysis (git-aware)  | Yes         | No        | No         |

**Core differentiator:** Mapstr is built *for developers*, not just for AI tools. The output is meant to be read, shared, and used by humans first.

---

## Tech Stack

| Layer          | Technology                            |
|----------------|---------------------------------------|
| Language       | Go (Cobra CLI)                        |
| Code Parsing   | Tree-sitter (via go bindings)         |
| AI Summary     | Claude, OpenAI, Gemini, DeepSeek, Mistral, Ollama |
| Graph Output   | Mermaid.js                            |
| Config File    | `.mapstr.yml`                     |
| Deployment     | Goreleaser -> GitHub Releases         |
| Testing        | Go stdlib + golden file tests         |

---

## Installation

```bash
# Go
go install github.com/mapstr/mapstr@latest

# Homebrew
brew install mapstr

# NPM (thin wrapper)
npm install -g mapstr

# Docker
docker run --rm -v $(pwd):/app mapstr /app

# Pre-built binaries
# Download from GitHub Releases — Linux, macOS, Windows
```

---

## Integrations

### VS Code Extension
One-click button that runs `mapstr` on the open project and renders the output in a side panel — Mermaid graph included.

### GitHub Actions
```yaml
- name: Generate Codebase Context
  uses: mapstr/mapstr-action@v1
  with:
    lang: en
    output: md
```
Automatically generates and commits `CONTEXT.md` on every push — keeps documentation always up to date.

### MCP Server (Claude Code / Cursor)
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
Exposes codebase context as an MCP tool, so AI assistants can query project structure on demand.

### Git Hook (post-clone)
```bash
git config --global init.templateDir ~/.git-templates
# Automatically runs mapstr after every git clone
```

---

## Configuration — `.mapstr.yml`

```yaml
language: en          # Output language (en, ar, es, fr, de, zh, ja, ...)
output:
  - md
  - mermaid
  - json
ai:
  provider: claude    # claude | openai | gemini | deepseek | mistral | ollama
  model: claude-sonnet-4-5
  # provider: openai
  # model: gpt-4o
  # provider: gemini
  # model: gemini-2.5-pro
  # provider: ollama
  # model: llama3
  fallback: ollama    # If primary provider fails, fall back to local model
  no_ai: false        # Set to true for structural-only analysis (no LLM calls)
depth: 3              # How deep to traverse the dependency tree
incremental: true     # Only re-analyze changed files
ignore:
  - node_modules
  - .git
  - dist
  - vendor
  - __pycache__
```

---

## Build Plan (5 Days)

| Day   | Milestone                                              |
|-------|--------------------------------------------------------|
| Day 1 | Go CLI scaffold (Cobra) + Tree-sitter parsing for JS/Go |
| Day 2 | Graph builder + Mermaid output generation              |
| Day 3 | LLM provider abstraction + Claude/OpenAI/Gemini/Ollama integration |
| Day 4 | JSON output + `.mapstr.yml` config + README        |
| Day 5 | Goreleaser setup + GitHub Release + launch prep        |

---

## Launch Strategy

### Day 1 — Launch
- **Hacker News**: *"Show HN: I built a CLI that maps any codebase in seconds with AI"*
- **Reddit**: r/programming, r/golang, r/webdev, r/devtools
- Animated GIF/video showing live before-and-after analysis

### Week 1 — Adoption
- Share real-world examples: mapping popular open-source repos (Express, Gin, FastAPI)
- Publish `CONTEXT.md` files for well-known projects as proof of value
- Engage with early adopters, ship fixes fast

### Month 1 — Growth
- **GitHub App** that auto-adds `CONTEXT.md` to any repo
- **README badge**: `![Mapstr](badge-url)` — social proof in every repo
- Complete docs + examples for every supported language
- VS Code extension on the marketplace

### Targets
```
Week 1  ->   500 stars
Week 2  ->  2,000 stars
Month 1 ->  5,000 stars  <- Qualifies for Claude for Open Source
```

---

## Why This Qualifies for Claude for Open Source

- Integrates **Claude API** as the default and recommended provider — not a wrapper, a real use case
- Multi-LLM support drives broader adoption, but Claude remains the flagship experience
- Designed for daily use by thousands of developers
- Critical infrastructure in the developer workflow (onboarding, code review, documentation)
- Direct, measurable impact on developer productivity
- Grows the Claude ecosystem by making codebases more accessible to AI tools

---

## Future Roadmap

- **Interactive mode**: TUI with navigable dependency graph
- **PR context**: Auto-generate context diffs for pull requests
- **Team mode**: Shared context maps with annotations
- **Plugin system**: Custom analyzers for frameworks (React, Django, Rails)
- **Embedding export**: Vector embeddings of codebase structure for RAG pipelines
- **Provider benchmarks**: Auto-compare summary quality across providers for a given codebase
- **Custom/self-hosted LLMs**: Support for any OpenAI-compatible API endpoint (vLLM, LMStudio, etc.)

---

## License

MIT — Free forever.

---

> **"Don't read the code. Map it."**

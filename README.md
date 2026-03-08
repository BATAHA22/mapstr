<p align="center">
  <h1 align="center">🗺️ Mapstr</h1>
  <p align="center">
    <strong>AI-Powered Codebase Navigator — Understand any project in seconds.</strong>
  </p>
  <p align="center">
    <a href="#-quick-start">Quick Start</a> •
    <a href="#-installation">Installation</a> •
    <a href="#-llm-providers">LLM Providers</a> •
    <a href="#-configuration">Configuration</a> •
    <a href="#-integrations">Integrations</a>
  </p>
  <p align="center">
    <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version">
    <img src="https://img.shields.io/badge/License-MIT-green?style=flat" alt="License">
    <img src="https://img.shields.io/badge/LLMs-6+_Providers-blueviolet?style=flat" alt="LLM Providers">
    <img src="https://img.shields.io/badge/Languages-Go_JS_TS_Python-orange?style=flat" alt="Languages">
  </p>
</p>

---

> **"Don't read the code. Map it."** 🧭

Mapstr is a **single-binary CLI** that analyzes any software project and generates a complete, human-readable map of the codebase — powered by structural parsing and multi-LLM summarization.

**One command. Three outputs. Any language. Any LLM.**

---

## ⚡ Quick Start

```bash
# 🔍 Analyze current directory (auto-detects your LLM provider)
mapstr .

# 🤖 Pick your AI provider
mapstr ./my-project --provider claude
mapstr ./my-project --provider openai --model gpt-4o
mapstr ./my-project --provider gemini --model gemini-2.5-pro

# 🏠 Use a local model — fully offline, zero API costs
mapstr ./my-project --provider ollama --model llama3

# 🔧 Structural analysis only — no AI, no API calls
mapstr ./my-project --no-ai
```

---

## 📦 Output

One command generates **three files** instantly:

| File | Description |
|------|-------------|
| 📄 `CONTEXT.md` | Natural-language architecture overview — written for humans |
| 📊 `GRAPH.mmd` | Mermaid dependency diagram — visualize relationships at a glance |
| 🤖 `context.json` | Structured data — optimized for AI assistants (Claude, Cursor, Copilot) |

<details>
<summary>📄 Example <code>CONTEXT.md</code> output</summary>

```markdown
# Project: my-api

## Overview
This project contains 42 files across 2 languages.

| Language   | Files |
|------------|-------|
| Go         | 30    |
| TypeScript | 12    |

## Architecture
- 87 functions/methods
- 15 types/classes
- 12 API routes
- 23 module dependencies

## Entry Points
- `cmd/server/main.go`
- `src/index.ts`

## API Routes
| Method | Path           | File              |
|--------|----------------|-------------------|
| GET    | /api/users     | handlers/user.go  |
| POST   | /api/users     | handlers/user.go  |
| GET    | /api/health    | handlers/health.go|
```

</details>

<details>
<summary>📊 Example <code>GRAPH.mmd</code> output</summary>

```mermaid
graph TD
    main_go["main.go"]
    server_go["server.go"]
    handler_go["handler.go"]
    db_go["db.go"]

    main_go --> server_go
    server_go --> handler_go
    handler_go --> db_go

    classDef module fill:#e1f5fe,stroke:#01579b
    classDef api fill:#e8f5e9,stroke:#1b5e20
```

</details>

---

## 📥 Installation

Choose your preferred method:

### 🐹 Go

```bash
go install github.com/mapstr/mapstr@latest
```

### 🍺 Homebrew (macOS / Linux)

```bash
brew install mapstr/tap/mapstr
```

### 📦 Pre-built Binaries

Download the latest release for your platform from [**GitHub Releases**](https://github.com/mapstr/mapstr/releases):

| Platform | Architecture | Download |
|----------|-------------|----------|
| 🐧 Linux | amd64 / arm64 | `mapstr_linux_amd64.tar.gz` |
| 🍎 macOS | amd64 / arm64 (Apple Silicon) | `mapstr_darwin_arm64.tar.gz` |
| 🪟 Windows | amd64 | `mapstr_windows_amd64.zip` |

### 🐳 Docker

```bash
docker run --rm -v $(pwd):/app mapstr/mapstr /app
```

---

## 🌐 Supported Languages

| Language | Parser | Status |
|----------|--------|--------|
| 🐹 Go | `go/parser` (stdlib AST) | ✅ Stable |
| 🟨 JavaScript | Regex-based extraction | ✅ Stable |
| 🔷 TypeScript | Regex-based extraction | ✅ Stable |
| 🐍 Python | Regex-based extraction | ✅ Stable |
| 🦀 Rust | — | 🔜 Planned |
| ☕ Java | — | 🔜 Planned |
| 💎 Ruby | — | 🔜 Planned |

> **What gets extracted:** functions, methods, classes, interfaces, types, imports, exports, and API routes (Express, Flask, FastAPI, net/http, etc.)

---

## 🤖 LLM Providers

Mapstr works with **6+ LLM providers** out of the box. Just set your API key and go:

| Provider | Env Variable | Default Model | Best For |
|----------|-------------|---------------|----------|
| 🟣 Claude (Anthropic) | `ANTHROPIC_API_KEY` | `claude-sonnet-4-5` | Nuanced architecture summaries |
| 🟢 OpenAI | `OPENAI_API_KEY` | `gpt-4o` | Strong general performance |
| 🔵 Gemini (Google) | `GEMINI_API_KEY` | `gemini-2.5-pro` | Large context — great for monorepos |
| 🟡 DeepSeek | `DEEPSEEK_API_KEY` | `deepseek-chat` | Cost-effective reasoning |
| 🟠 Mistral | `MISTRAL_API_KEY` | `mistral-large-latest` | EU-hosted, strong multilingual |
| ⚫ Ollama (local) | — | `llama3` | Fully offline, zero cost, private |

### 🔍 Auto-Detection

If no `--provider` flag is set, Mapstr automatically checks for API keys in the order above and uses the **first available provider**. No configuration needed — just set your key and run.

### 🛡️ Fallback System

If the primary provider fails, Mapstr gracefully falls back:

```
Primary provider → Fallback provider → --no-ai mode → never crashes
```

---

## ⚙️ Configuration

Create a `.mapstr.yml` in your project root for persistent settings:

```yaml
# 🌍 Output language (en, ar, es, fr, de, zh, ja, ...)
language: en

# 📦 Output formats to generate
output:
  - md        # CONTEXT.md
  - mermaid   # GRAPH.mmd
  - json      # context.json

# 🤖 AI provider settings
ai:
  provider: claude              # claude | openai | gemini | deepseek | mistral | ollama
  model: claude-sonnet-4-5      # Model name (uses provider default if omitted)
  fallback: ollama              # Fallback if primary fails
  no_ai: false                  # Set true for structural-only analysis

# 🔍 Analysis settings
depth: 3                        # How deep to traverse the dependency tree
incremental: true               # Only re-analyze changed files (git-aware)

# 🚫 Ignored directories
ignore:
  - node_modules
  - .git
  - dist
  - vendor
  - __pycache__
  - .next
  - build
```

---

## 🚩 CLI Reference

```
Usage:
  mapstr [path] [flags]

Flags:
  -l, --lang string        🌍 Output language (default "en")
  -o, --output strings     📦 Output formats: md, mermaid, json (default: all)
  -p, --provider string    🤖 LLM provider: claude, openai, gemini, ollama, deepseek, mistral
  -m, --model string       🧠 Model name (uses provider default if not set)
      --no-ai              🔧 Skip LLM — structural analysis + graph only
  -d, --depth int          🔍 Dependency tree depth (default 3)
  -w, --watch              👀 Watch mode — regenerate on file changes
      --mcp                🔌 Start as MCP server for AI assistants
  -c, --config string      📄 Path to .mapstr.yml config file
      --out-dir string     📁 Output directory (default ".")
  -h, --help               ❓ Show help
  -v, --version            📌 Show version
```

### 💡 Usage Examples

```bash
# Analyze a project in Arabic
mapstr ./my-project --lang ar

# Generate only the Mermaid diagram
mapstr ./my-project --output mermaid --no-ai

# Use GPT-4o and output to a specific directory
mapstr ./api --provider openai --model gpt-4o --out-dir ./docs

# Analyze with DeepSeek for cost-effective summaries
mapstr ./large-monorepo --provider deepseek
```

---

## 🔌 Integrations

### 🧩 MCP Server (Claude Code / Cursor)

Run Mapstr as an MCP server so AI assistants can query your project structure on demand:

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

### 🔄 GitHub Actions

Auto-generate and commit `CONTEXT.md` on every push — documentation that never goes stale:

```yaml
- name: 🗺️ Generate Codebase Context
  uses: mapstr/mapstr-action@v1
  with:
    lang: en
    output: md
```

### 🪝 Git Hook (post-clone)

Automatically run Mapstr after every `git clone`:

```bash
git config --global init.templateDir ~/.git-templates
# Add mapstr to the post-checkout hook
```

---

## 🏗️ How It Works

```
📂 codebase
   │
   ▼
🔬 Language Parsers         Go / JS / TS / Python AST parsing
   │
   ▼
🔗 Dependency Resolver      Resolves imports, exports, cross-file references
   │
   ▼
🕸️ Graph Builder            Builds relationship graph of modules & functions
   │
   ▼
🤖 LLM Summarizer           Claude / OpenAI / Gemini / Ollama / DeepSeek / Mistral
   │
   ▼
📦 Output Engine
   ├── 📄 CONTEXT.md         Architecture overview
   ├── 📊 GRAPH.mmd          Mermaid dependency diagram
   └── 🤖 context.json       Structured data for AI tools
```

### 🔑 Key Design Decisions

- **No CGo** — Pure Go + regex parsers for JS/Python. Single static binary, easy cross-compilation.
- **No Tree-sitter dependency** — Go's stdlib `go/parser` for Go files, battle-tested regex for JS/TS/Python.
- **Provider-agnostic** — Unified `Provider` interface. Add a new LLM in ~50 lines.
- **Incremental analysis** — Git-aware caching. Only re-parses changed files.
- **Graceful degradation** — If AI fails, you still get the structural analysis and graph.

---

## 🆚 Comparison

| Feature | Mapstr | Repomix | CodePrism |
|---------|--------|---------|-----------|
| 🤖 Multi-LLM support (6+) | ✅ | ❌ | ❌ |
| 🔧 Offline / no-AI mode | ✅ | ❌ | ❌ |
| 📊 Visual Mermaid graph | ✅ | ❌ | ⚠️ Complex |
| 🌍 Multi-language summaries (i18n) | ✅ | ❌ | ❌ |
| 📦 Single binary CLI | ✅ (Go) | ✅ (Node) | ❌ (Rust) |
| 👨‍💻 Built for developers | ✅ | ❌ | ❌ |
| 🔌 MCP Server | ✅ | ❌ | ✅ |
| 🔄 GitHub App | ✅ | ❌ | ❌ |
| ⚡ Incremental (git-aware) | ✅ | ❌ | ❌ |

---

## 🗺️ Roadmap

- [ ] 🖥️ **Interactive TUI** — Navigable dependency graph in the terminal
- [ ] 🔀 **PR Context** — Auto-generate context diffs for pull requests
- [ ] 👥 **Team Mode** — Shared context maps with annotations
- [ ] 🔌 **Plugin System** — Custom analyzers for frameworks (React, Django, Rails)
- [ ] 🧬 **Embedding Export** — Vector embeddings for RAG pipelines
- [ ] 📊 **Provider Benchmarks** — Compare summary quality across LLMs
- [ ] 🌐 **Custom Endpoints** — Support any OpenAI-compatible API (vLLM, LMStudio)
- [ ] 🦀 **More Languages** — Rust, Java, C#, Ruby, PHP

---

## 🤝 Contributing

Contributions are welcome! Here's how to get started:

```bash
# Clone the repo
git clone https://github.com/mapstr/mapstr.git
cd mapstr

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build
go build -o mapstr .

# Run on itself 🤯
./mapstr . --no-ai
```

---

## 📄 License

MIT — Free and open source, forever.

---

<p align="center">
  <strong>🗺️ "Don't read the code. Map it."</strong>
  <br><br>
  <a href="https://github.com/mapstr/mapstr">⭐ Star on GitHub</a> •
  <a href="https://github.com/mapstr/mapstr/issues">🐛 Report Bug</a> •
  <a href="https://github.com/mapstr/mapstr/issues">💡 Request Feature</a>
</p>

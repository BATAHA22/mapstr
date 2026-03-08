package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mapstr/mapstr/internal/config"
)

// FunctionDef represents a function or method definition.
type FunctionDef struct {
	Name       string   `json:"name"`
	Receiver   string   `json:"receiver,omitempty"`
	Parameters []string `json:"parameters,omitempty"`
	Returns    []string `json:"returns,omitempty"`
	Line       int      `json:"line"`
	Exported   bool     `json:"exported"`
}

// TypeDef represents a type definition (struct, interface, class, etc).
type TypeDef struct {
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Fields   []string `json:"fields,omitempty"`
	Line     int      `json:"line"`
	Exported bool     `json:"exported"`
}

// APIRoute represents an HTTP route definition.
type APIRoute struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
	Line    int    `json:"line"`
}

// FileNode is the parsed representation of a single source file.
type FileNode struct {
	Path      string        `json:"path"`
	Language  string        `json:"language"`
	Functions []FunctionDef `json:"functions,omitempty"`
	Imports   []string      `json:"imports,omitempty"`
	Exports   []string      `json:"exports,omitempty"`
	APIRoutes []APIRoute    `json:"api_routes,omitempty"`
	Types     []TypeDef     `json:"types,omitempty"`
}

// Parser is the interface that language-specific parsers must implement.
type Parser interface {
	Language() string
	Extensions() []string
	Parse(path string, content []byte) (*FileNode, error)
}

// registry holds all registered parsers keyed by extension.
var (
	registry   = map[string]Parser{}
	registryMu sync.RWMutex
)

// Register adds a parser to the global registry for each of its extensions.
func Register(p Parser) {
	registryMu.Lock()
	defer registryMu.Unlock()
	for _, ext := range p.Extensions() {
		registry[ext] = p
	}
}

// ForExtension returns the parser registered for the given file extension.
func ForExtension(ext string) (Parser, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := registry[ext]
	return p, ok
}

// ParseProject walks a project directory and parses all recognized source files.
func ParseProject(root string, cfg *config.Config) ([]*FileNode, error) {
	var nodes []*FileNode
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 32) // limit concurrent file reads

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip files we can't read
		}

		// Get path relative to root for ignore checking.
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}

		if info.IsDir() {
			if cfg.ShouldIgnore(rel) || cfg.ShouldIgnore(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if cfg.ShouldIgnore(rel) || cfg.ShouldIgnore(info.Name()) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		p, ok := ForExtension(ext)
		if !ok {
			return nil
		}

		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return
			}

			node, parseErr := p.Parse(rel, content)
			if parseErr != nil {
				return
			}

			mu.Lock()
			nodes = append(nodes, node)
			mu.Unlock()
		}()

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("parser: walk %s: %w", root, err)
	}

	wg.Wait()
	return nodes, nil
}

// SupportedLanguages returns a deduplicated list of registered languages.
func SupportedLanguages() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	seen := map[string]bool{}
	var langs []string
	for _, p := range registry {
		lang := p.Language()
		if !seen[lang] {
			seen[lang] = true
			langs = append(langs, lang)
		}
	}
	return langs
}

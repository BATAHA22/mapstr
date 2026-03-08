package graph

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BATAHA22/mapstr/internal/parser"
)

// GraphNode represents a node in the dependency graph.
type GraphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"` // "module", "function", "api", "type"
	File  string `json:"file"`
}

// GraphEdge represents a directed edge in the dependency graph.
type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "imports", "calls", "extends", "implements"
}

// DependencyGraph is the complete graph of a project's dependencies.
type DependencyGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// Build constructs a dependency graph from parsed file nodes.
func Build(nodes []*parser.FileNode) *DependencyGraph {
	g := &DependencyGraph{}

	// Index: file path -> node ID
	fileIndex := map[string]string{}
	// Index: export name -> file path
	exportIndex := map[string]string{}

	// Phase 1: Create module nodes for each file
	for _, node := range nodes {
		id := sanitizeID(node.Path)
		label := filepath.Base(node.Path)

		g.Nodes = append(g.Nodes, GraphNode{
			ID:    id,
			Label: label,
			Type:  "module",
			File:  node.Path,
		})

		fileIndex[node.Path] = id

		// Index exports
		for _, exp := range node.Exports {
			exportIndex[exp] = node.Path
		}
	}

	// Phase 2: Create edges from imports
	for _, node := range nodes {
		fromID := fileIndex[node.Path]

		for _, imp := range node.Imports {
			// Try to resolve the import to a local file
			targetPath := resolveImport(imp, node.Path, fileIndex)
			if targetPath != "" {
				toID := fileIndex[targetPath]
				g.Edges = append(g.Edges, GraphEdge{
					From: fromID,
					To:   toID,
					Type: "imports",
				})
			}
		}
	}

	// Phase 3: Add type and function nodes for key exports
	for _, node := range nodes {
		for _, t := range node.Types {
			if !t.Exported {
				continue
			}
			id := sanitizeID(node.Path + ":" + t.Name)
			g.Nodes = append(g.Nodes, GraphNode{
				ID:    id,
				Label: t.Name,
				Type:  "type",
				File:  node.Path,
			})
		}

		for _, route := range node.APIRoutes {
			id := sanitizeID(node.Path + ":" + route.Method + ":" + route.Path)
			g.Nodes = append(g.Nodes, GraphNode{
				ID:    id,
				Label: route.Method + " " + route.Path,
				Type:  "api",
				File:  node.Path,
			})
		}
	}

	return g
}

// Summary returns a human-readable text summary of the graph.
func (g *DependencyGraph) Summary() string {
	var b strings.Builder

	moduleCount := 0
	apiCount := 0
	typeCount := 0

	for _, n := range g.Nodes {
		switch n.Type {
		case "module":
			moduleCount++
		case "api":
			apiCount++
		case "type":
			typeCount++
		}
	}

	b.WriteString("Graph summary:\n")
	b.WriteString("  Modules: " + itoa(moduleCount) + "\n")
	b.WriteString("  Types: " + itoa(typeCount) + "\n")
	b.WriteString("  API routes: " + itoa(apiCount) + "\n")
	b.WriteString("  Dependencies: " + itoa(len(g.Edges)) + "\n\n")

	b.WriteString("Key relationships:\n")
	for _, e := range g.Edges {
		b.WriteString("  " + e.From + " --" + e.Type + "--> " + e.To + "\n")
	}

	return b.String()
}

// resolveImport attempts to match an import string to a known local file.
func resolveImport(imp, fromPath string, fileIndex map[string]string) string {
	// Handle relative imports (./foo, ../bar)
	if strings.HasPrefix(imp, ".") {
		dir := filepath.Dir(fromPath)
		resolved := filepath.Join(dir, imp)
		resolved = filepath.ToSlash(resolved)

		// Try exact match, then with common extensions
		extensions := []string{"", ".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".dart", "/index.js", "/index.ts"}
		for _, ext := range extensions {
			candidate := resolved + ext
			if _, ok := fileIndex[candidate]; ok {
				return candidate
			}
		}
	}

	// Handle Go package imports — match by directory name
	for path := range fileIndex {
		dir := filepath.Dir(path)
		if strings.HasSuffix(imp, filepath.Base(dir)) || strings.Contains(imp, filepath.ToSlash(dir)) {
			return path
		}
	}

	return ""
}

func sanitizeID(s string) string {
	r := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		".", "_",
		":", "_",
		" ", "_",
		"-", "_",
	)
	return r.Replace(s)
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

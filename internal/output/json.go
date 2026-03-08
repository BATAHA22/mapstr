package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mapstr/mapstr/internal/graph"
	"github.com/mapstr/mapstr/internal/parser"
)

// ContextJSON is the structured output for AI assistants.
type ContextJSON struct {
	Project     string                `json:"project"`
	GeneratedAt string               `json:"generated_at"`
	Summary     string               `json:"summary"`
	Files       []*parser.FileNode   `json:"files"`
	Graph       *graph.DependencyGraph `json:"graph"`
	EntryPoints []string             `json:"entry_points"`
}

// GenerateJSON produces the context.json content.
func GenerateJSON(projectName string, nodes []*parser.FileNode, g *graph.DependencyGraph, aiSummary string) (string, error) {
	var entryPoints []string
	for _, n := range nodes {
		if isEntryPoint(n.Path) {
			entryPoints = append(entryPoints, n.Path)
		}
	}

	summary := aiSummary
	if summary == "" {
		summary = buildStructuralSummary(nodes, g)
	}

	out := ContextJSON{
		Project:     projectName,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Summary:     summary,
		Files:       nodes,
		Graph:       g,
		EntryPoints: entryPoints,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("output: json marshal: %w", err)
	}

	return string(data), nil
}

func buildStructuralSummary(nodes []*parser.FileNode, g *graph.DependencyGraph) string {
	languages := map[string]int{}
	for _, n := range nodes {
		languages[n.Language]++
	}

	var parts []string
	for lang, count := range languages {
		parts = append(parts, fmt.Sprintf("%d %s files", count, lang))
	}

	return fmt.Sprintf("Project with %s, %d dependencies.",
		strings.Join(parts, ", "), len(g.Edges))
}

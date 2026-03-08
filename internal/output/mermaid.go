package output

import (
	"github.com/BATAHA22/mapstr/internal/graph"
)

// GenerateMermaid produces the GRAPH.mmd content.
func GenerateMermaid(g *graph.DependencyGraph) string {
	return g.ToMermaid()
}

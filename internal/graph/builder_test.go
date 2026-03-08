package graph

import (
	"strings"
	"testing"

	"github.com/BATAHA22/mapstr/internal/parser"
)

func TestBuild(t *testing.T) {
	nodes := []*parser.FileNode{
		{
			Path:     "main.go",
			Language: "Go",
			Imports:  []string{"./internal/server"},
			Functions: []parser.FunctionDef{
				{Name: "main", Line: 5, Exported: false},
			},
		},
		{
			Path:     "internal/server/server.go",
			Language: "Go",
			Imports:  []string{"net/http"},
			Functions: []parser.FunctionDef{
				{Name: "Start", Line: 10, Exported: true},
			},
			Types: []parser.TypeDef{
				{Name: "Server", Kind: "struct", Exported: true},
			},
			Exports: []string{"Start", "Server"},
		},
	}

	g := Build(nodes)

	if len(g.Nodes) < 2 {
		t.Errorf("expected at least 2 nodes, got %d", len(g.Nodes))
	}

	// Check that exported types appear as type nodes
	hasTypeNode := false
	for _, n := range g.Nodes {
		if n.Type == "type" && n.Label == "Server" {
			hasTypeNode = true
		}
	}
	if !hasTypeNode {
		t.Error("expected a type node for Server")
	}
}

func TestBuildWithAPIRoutes(t *testing.T) {
	nodes := []*parser.FileNode{
		{
			Path:     "routes.go",
			Language: "Go",
			APIRoutes: []parser.APIRoute{
				{Method: "GET", Path: "/users", Handler: "getUsers", Line: 10},
				{Method: "POST", Path: "/users", Handler: "createUser", Line: 15},
			},
		},
	}

	g := Build(nodes)

	apiNodes := 0
	for _, n := range g.Nodes {
		if n.Type == "api" {
			apiNodes++
		}
	}

	if apiNodes != 2 {
		t.Errorf("expected 2 api nodes, got %d", apiNodes)
	}
}

func TestMermaid(t *testing.T) {
	g := &DependencyGraph{
		Nodes: []GraphNode{
			{ID: "main_go", Label: "main.go", Type: "module"},
			{ID: "server_go", Label: "server.go", Type: "module"},
		},
		Edges: []GraphEdge{
			{From: "main_go", To: "server_go", Type: "imports"},
		},
	}

	mermaid := g.ToMermaid()

	if !strings.Contains(mermaid, "graph TD") {
		t.Error("mermaid output should start with graph TD")
	}
	if !strings.Contains(mermaid, "main_go") {
		t.Error("mermaid should contain main_go node")
	}
	if !strings.Contains(mermaid, "main_go --> server_go") {
		t.Error("mermaid should contain the edge")
	}
}

func TestSummary(t *testing.T) {
	g := &DependencyGraph{
		Nodes: []GraphNode{
			{ID: "a", Label: "a.go", Type: "module"},
			{ID: "b", Label: "b.go", Type: "module"},
			{ID: "T", Label: "MyType", Type: "type"},
		},
		Edges: []GraphEdge{
			{From: "a", To: "b", Type: "imports"},
		},
	}

	summary := g.Summary()

	if !strings.Contains(summary, "Modules: 2") {
		t.Errorf("summary should mention 2 modules: %s", summary)
	}
	if !strings.Contains(summary, "Types: 1") {
		t.Errorf("summary should mention 1 type: %s", summary)
	}
	if !strings.Contains(summary, "Dependencies: 1") {
		t.Errorf("summary should mention 1 dependency: %s", summary)
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"main.go", "main_go"},
		{"internal/parser/go.go", "internal_parser_go_go"},
		{"file:method", "file_method"},
	}

	for _, tt := range tests {
		got := sanitizeID(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

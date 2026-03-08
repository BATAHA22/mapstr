package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/BATAHA22/mapstr/internal/graph"
	"github.com/BATAHA22/mapstr/internal/parser"
)

func sampleNodes() []*parser.FileNode {
	return []*parser.FileNode{
		{
			Path:     "main.go",
			Language: "Go",
			Functions: []parser.FunctionDef{
				{Name: "main", Line: 3, Exported: false},
			},
			Imports: []string{"./server"},
		},
		{
			Path:     "server/server.go",
			Language: "Go",
			Functions: []parser.FunctionDef{
				{Name: "Start", Line: 10, Exported: true},
				{Name: "Stop", Line: 20, Exported: true},
			},
			Types: []parser.TypeDef{
				{Name: "Server", Kind: "struct", Exported: true},
			},
			APIRoutes: []parser.APIRoute{
				{Method: "GET", Path: "/health", Handler: "healthCheck", Line: 30},
			},
			Exports: []string{"Start", "Stop", "Server"},
		},
	}
}

func sampleGraph() *graph.DependencyGraph {
	return &graph.DependencyGraph{
		Nodes: []graph.GraphNode{
			{ID: "main_go", Label: "main.go", Type: "module"},
			{ID: "server_go", Label: "server.go", Type: "module"},
		},
		Edges: []graph.GraphEdge{
			{From: "main_go", To: "server_go", Type: "imports"},
		},
	}
}

func TestGenerateMarkdownWithAI(t *testing.T) {
	md := GenerateMarkdown("test-project", sampleNodes(), sampleGraph(), "This is an AI summary.")

	if !strings.Contains(md, "# Project: test-project") {
		t.Error("should contain project header")
	}
	if !strings.Contains(md, "This is an AI summary.") {
		t.Error("should contain AI summary")
	}
	if !strings.Contains(md, "Mapstr") {
		t.Error("should contain Mapstr attribution")
	}
}

func TestGenerateMarkdownNoAI(t *testing.T) {
	md := GenerateMarkdown("test-project", sampleNodes(), sampleGraph(), "")

	if !strings.Contains(md, "## Overview") {
		t.Error("should contain structural overview")
	}
	if !strings.Contains(md, "2 files") {
		t.Error("should mention file count")
	}
	if !strings.Contains(md, "GET") {
		t.Error("should list API routes")
	}
	if !strings.Contains(md, "Server") {
		t.Error("should list key types")
	}
}

func TestGenerateJSON(t *testing.T) {
	nodes := sampleNodes()
	g := sampleGraph()

	jsonStr, err := GenerateJSON("test-project", nodes, g, "AI summary here")
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	var result ContextJSON
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if result.Project != "test-project" {
		t.Errorf("project = %q, want %q", result.Project, "test-project")
	}

	if result.Summary != "AI summary here" {
		t.Errorf("summary mismatch")
	}

	if len(result.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(result.Files))
	}

	if result.Graph == nil {
		t.Error("graph should not be nil")
	}

	if len(result.EntryPoints) == 0 {
		t.Error("should detect entry points")
	}
}

func TestGenerateJSONNoAI(t *testing.T) {
	jsonStr, err := GenerateJSON("proj", sampleNodes(), sampleGraph(), "")
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	var result ContextJSON
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if result.Summary == "" {
		t.Error("should have a structural summary when no AI")
	}
}

func TestGenerateMermaid(t *testing.T) {
	g := sampleGraph()
	mmd := GenerateMermaid(g)

	if !strings.Contains(mmd, "graph TD") {
		t.Error("should contain graph TD")
	}
	if !strings.Contains(mmd, "main_go") {
		t.Error("should contain main_go node")
	}
}

package tests

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mapstr/mapstr/internal/config"
	"github.com/mapstr/mapstr/internal/graph"
	"github.com/mapstr/mapstr/internal/output"
	"github.com/mapstr/mapstr/internal/parser"

	// Register parsers
	_ "github.com/mapstr/mapstr/internal/parser"
)

func TestEndToEndNoAI(t *testing.T) {
	fixtureDir, err := filepath.Abs("fixtures/sample")
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.AI.NoAI = true

	// Parse
	nodes, err := parser.ParseProject(fixtureDir, cfg)
	if err != nil {
		t.Fatalf("ParseProject: %v", err)
	}

	if len(nodes) == 0 {
		t.Fatal("no files parsed")
	}

	// Verify we got files from all three languages
	languages := map[string]bool{}
	for _, n := range nodes {
		languages[n.Language] = true
	}

	for _, lang := range []string{"Go", "JavaScript", "Python"} {
		if !languages[lang] {
			t.Errorf("expected %s files to be parsed", lang)
		}
	}

	// Build graph
	g := graph.Build(nodes)
	if len(g.Nodes) == 0 {
		t.Error("graph should have nodes")
	}

	// Test markdown output
	md := output.GenerateMarkdown("sample", nodes, g, "")
	if !strings.Contains(md, "# Project: sample") {
		t.Error("markdown should contain project header")
	}
	if !strings.Contains(md, "## Overview") {
		t.Error("markdown should contain overview section")
	}

	// Test JSON output
	jsonStr, err := output.GenerateJSON("sample", nodes, g, "")
	if err != nil {
		t.Fatalf("GenerateJSON: %v", err)
	}

	var result output.ContextJSON
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal: %v", err)
	}

	if result.Project != "sample" {
		t.Errorf("project = %q, want %q", result.Project, "sample")
	}

	if len(result.Files) != len(nodes) {
		t.Errorf("JSON files count = %d, parsed = %d", len(result.Files), len(nodes))
	}

	// Test mermaid output
	mmd := output.GenerateMermaid(g)
	if !strings.Contains(mmd, "graph TD") {
		t.Error("mermaid should start with graph TD")
	}
}

func TestNoAIProducesValidOutput(t *testing.T) {
	fixtureDir, err := filepath.Abs("fixtures/sample")
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.AI.NoAI = true

	nodes, err := parser.ParseProject(fixtureDir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	g := graph.Build(nodes)

	// Verify --no-ai generates all three outputs without errors
	md := output.GenerateMarkdown("test", nodes, g, "")
	if md == "" {
		t.Error("markdown should not be empty")
	}

	jsonStr, err := output.GenerateJSON("test", nodes, g, "")
	if err != nil {
		t.Errorf("JSON generation failed: %v", err)
	}
	if jsonStr == "" {
		t.Error("JSON should not be empty")
	}

	mmd := output.GenerateMermaid(g)
	if mmd == "" {
		t.Error("mermaid should not be empty")
	}
}

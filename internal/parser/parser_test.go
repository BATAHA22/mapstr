package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mapstr/mapstr/internal/config"
)

func TestGoParser(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantFuncs []string
		wantTypes []string
		wantImps  []string
	}{
		{
			name: "basic function",
			code: `package main

import "fmt"

func Hello(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
`,
			wantFuncs: []string{"Hello"},
			wantImps:  []string{"fmt"},
		},
		{
			name: "struct and method",
			code: `package service

type UserService struct {
	DB string
}

func (s *UserService) Create(name string) error {
	return nil
}

func (s *UserService) delete(id int) error {
	return nil
}
`,
			wantFuncs: []string{"Create", "delete"},
			wantTypes: []string{"UserService"},
		},
		{
			name: "interface",
			code: `package store

type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
}
`,
			wantTypes: []string{"Store"},
		},
		{
			name: "multiple imports",
			code: `package main

import (
	"context"
	"fmt"
	"net/http"
)

func main() {}
`,
			wantFuncs: []string{"main"},
			wantImps:  []string{"context", "fmt", "net/http"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &GoParser{}
			node, err := p.Parse("test.go", []byte(tt.code))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if tt.wantFuncs != nil {
				gotFuncs := make(map[string]bool)
				for _, f := range node.Functions {
					gotFuncs[f.Name] = true
				}
				for _, want := range tt.wantFuncs {
					if !gotFuncs[want] {
						t.Errorf("missing function %q", want)
					}
				}
			}

			if tt.wantTypes != nil {
				gotTypes := make(map[string]bool)
				for _, typ := range node.Types {
					gotTypes[typ.Name] = true
				}
				for _, want := range tt.wantTypes {
					if !gotTypes[want] {
						t.Errorf("missing type %q", want)
					}
				}
			}

			if tt.wantImps != nil {
				gotImps := make(map[string]bool)
				for _, imp := range node.Imports {
					gotImps[imp] = true
				}
				for _, want := range tt.wantImps {
					if !gotImps[want] {
						t.Errorf("missing import %q", want)
					}
				}
			}
		})
	}
}

func TestJSParser(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		code      string
		wantFuncs []string
		wantImps  []string
		wantTypes []string
	}{
		{
			name: "ES module imports and exports",
			file: "app.js",
			code: `import express from 'express';
import { Router } from 'express';
import config from './config';

export function createApp() {
  const app = express();
  return app;
}

export default createApp;
`,
			wantFuncs: []string{"createApp"},
			wantImps:  []string{"express", "./config"},
		},
		{
			name: "CommonJS",
			file: "server.js",
			code: `const http = require('http');
const app = require('./app');

function startServer(port) {
  http.createServer(app).listen(port);
}

module.exports = startServer;
`,
			wantFuncs: []string{"startServer"},
			wantImps:  []string{"http", "./app"},
		},
		{
			name: "TypeScript class and interface",
			file: "service.ts",
			code: `import { User } from './models';

export interface UserService {
  getUser(id: string): Promise<User>;
}

export class UserServiceImpl {
  constructor(private db: Database) {}
}

export type UserID = string;
`,
			wantTypes: []string{"UserService", "UserServiceImpl", "UserID"},
			wantImps:  []string{"./models"},
		},
		{
			name: "Express routes",
			file: "routes.js",
			code: `const express = require('express');
const router = express.Router();

router.get('/users', getUsers);
router.post('/users', createUser);
app.delete('/users/:id', deleteUser);
`,
			wantImps: []string{"express"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &JSParser{}
			node, err := p.Parse(tt.file, []byte(tt.code))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if tt.wantFuncs != nil {
				gotFuncs := make(map[string]bool)
				for _, f := range node.Functions {
					gotFuncs[f.Name] = true
				}
				for _, want := range tt.wantFuncs {
					if !gotFuncs[want] {
						t.Errorf("missing function %q", want)
					}
				}
			}

			if tt.wantImps != nil {
				gotImps := make(map[string]bool)
				for _, imp := range node.Imports {
					gotImps[imp] = true
				}
				for _, want := range tt.wantImps {
					if !gotImps[want] {
						t.Errorf("missing import %q", want)
					}
				}
			}

			if tt.wantTypes != nil {
				gotTypes := make(map[string]bool)
				for _, typ := range node.Types {
					gotTypes[typ.Name] = true
				}
				for _, want := range tt.wantTypes {
					if !gotTypes[want] {
						t.Errorf("missing type %q", want)
					}
				}
			}
		})
	}
}

func TestPythonParser(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantFuncs []string
		wantImps  []string
		wantTypes []string
	}{
		{
			name: "basic module",
			code: `import os
from pathlib import Path

def read_file(path):
    with open(path) as f:
        return f.read()

class FileReader:
    def __init__(self, base_dir):
        self.base_dir = base_dir
`,
			wantFuncs: []string{"read_file"},
			wantImps:  []string{"os", "pathlib"},
			wantTypes: []string{"FileReader"},
		},
		{
			name: "private functions",
			code: `def public_func():
    pass

def _private_func():
    pass
`,
			wantFuncs: []string{"public_func", "_private_func"},
		},
		{
			name: "__all__ exports",
			code: `__all__ = ['create_app', 'Config']

def create_app():
    pass

def helper():
    pass

class Config:
    pass
`,
			wantFuncs: []string{"create_app", "helper"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PythonParser{}
			node, err := p.Parse("test.py", []byte(tt.code))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if tt.wantFuncs != nil {
				gotFuncs := make(map[string]bool)
				for _, f := range node.Functions {
					gotFuncs[f.Name] = true
				}
				for _, want := range tt.wantFuncs {
					if !gotFuncs[want] {
						t.Errorf("missing function %q", want)
					}
				}
			}

			if tt.wantImps != nil {
				gotImps := make(map[string]bool)
				for _, imp := range node.Imports {
					gotImps[imp] = true
				}
				for _, want := range tt.wantImps {
					if !gotImps[want] {
						t.Errorf("missing import %q", want)
					}
				}
			}

			if tt.wantTypes != nil {
				gotTypes := make(map[string]bool)
				for _, typ := range node.Types {
					gotTypes[typ.Name] = true
				}
				for _, want := range tt.wantTypes {
					if !gotTypes[want] {
						t.Errorf("missing type %q", want)
					}
				}
			}
		})
	}
}

func TestParseProject(t *testing.T) {
	dir := t.TempDir()

	// Create a Go file
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(`package main

func main() {}
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a JS file
	jsFile := filepath.Join(dir, "index.js")
	if err := os.WriteFile(jsFile, []byte(`export function hello() {}
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a file that should be ignored
	if err := os.Mkdir(filepath.Join(dir, "node_modules"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "dep.js"), []byte(`module.exports = {};`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	nodes, err := ParseProject(dir, cfg)
	if err != nil {
		t.Fatalf("ParseProject failed: %v", err)
	}

	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}

	// Verify node_modules was ignored
	for _, n := range nodes {
		if filepath.Base(filepath.Dir(n.Path)) == "node_modules" {
			t.Error("node_modules should be ignored")
		}
	}
}

func TestForExtension(t *testing.T) {
	tests := []struct {
		ext    string
		exists bool
	}{
		{".go", true},
		{".js", true},
		{".ts", true},
		{".py", true},
		{".rs", false},
		{".java", false},
	}

	for _, tt := range tests {
		_, ok := ForExtension(tt.ext)
		if ok != tt.exists {
			t.Errorf("ForExtension(%q) = %v, want %v", tt.ext, ok, tt.exists)
		}
	}
}

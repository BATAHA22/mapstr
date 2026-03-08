package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mapstr/mapstr/internal/config"
)

func BenchmarkParseProject(b *testing.B) {
	dir := b.TempDir()

	// Create 100 Go files
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf(`package pkg%d

import "fmt"

type Service%d struct {
	Name string
}

func (s *Service%d) Run() error {
	fmt.Println("running")
	return nil
}

func Helper%d() string {
	return "helper"
}
`, i, i, i, i)
		path := filepath.Join(dir, fmt.Sprintf("file%d.go", i))
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			b.Fatal(err)
		}
	}

	// Create 100 JS files
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf(`import { something } from './file%d';

export function handler%d(req, res) {
  res.json({ ok: true });
}

export class Service%d {
  constructor() {}
}
`, (i+1)%100, i, i)
		path := filepath.Join(dir, fmt.Sprintf("file%d.js", i))
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			b.Fatal(err)
		}
	}

	cfg := config.DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseProject(dir, cfg)
	}
}

func BenchmarkGoParser(b *testing.B) {
	code := []byte(`package main

import (
	"context"
	"fmt"
	"net/http"
)

type Server struct {
	Addr string
	Handler http.Handler
}

func NewServer(addr string) *Server {
	return &Server{Addr: addr}
}

func (s *Server) Start(ctx context.Context) error {
	fmt.Printf("Starting server on %s\n", s.Addr)
	return http.ListenAndServe(s.Addr, s.Handler)
}

func (s *Server) Stop() error {
	return nil
}
`)

	p := &GoParser{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse("server.go", code)
	}
}

func BenchmarkJSParser(b *testing.B) {
	code := []byte(`import express from 'express';
import { Router } from 'express';
import cors from 'cors';
import { UserController } from './controllers/user';

const app = express();
app.use(cors());

const router = Router();
router.get('/api/users', UserController.list);
router.post('/api/users', UserController.create);
router.get('/api/users/:id', UserController.get);
router.put('/api/users/:id', UserController.update);
router.delete('/api/users/:id', UserController.delete);

export class Application {
  constructor() {
    this.app = app;
  }

  start(port) {
    this.app.listen(port);
  }
}

export default app;
`)

	p := &JSParser{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse("app.js", code)
	}
}

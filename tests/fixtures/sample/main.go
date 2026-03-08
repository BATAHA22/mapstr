package main

import (
	"fmt"
	"net/http"
	"time"
)

type App struct {
	Name string
	Port int
}

func NewApp(name string, port int) *App {
	return &App{Name: name, Port: port}
}

func (a *App) Start() error {
	http.HandleFunc("/health", a.healthHandler)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.Port),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return srv.ListenAndServe()
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK: %s", a.Name)
}

func main() {
	app := NewApp("sample", 8080)
	_ = app.Start()
}

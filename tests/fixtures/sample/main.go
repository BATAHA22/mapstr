package main

import (
	"fmt"
	"net/http"
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
	return http.ListenAndServe(fmt.Sprintf(":%d", a.Port), nil)
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK: %s", a.Name)
}

func main() {
	app := NewApp("sample", 8080)
	_ = app.Start()
}

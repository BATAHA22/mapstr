package main

import (
	"github.com/mapstr/mapstr/cmd"

	// Register language parsers
	_ "github.com/mapstr/mapstr/internal/parser"
)

func main() {
	cmd.Execute()
}

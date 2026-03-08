package main

import (
	"github.com/BATAHA22/mapstr/cmd"

	// Register language parsers
	_ "github.com/BATAHA22/mapstr/internal/parser"
)

func main() {
	cmd.Execute()
}

package ui

import (
	"fmt"
	"strings"
	"time"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

var (
	bold    = color.New(color.Bold).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	dim     = color.New(color.Faint).SprintFunc()
)

// Spinner wraps briandowns/spinner with convenience methods.
type Spinner struct {
	s *spinner.Spinner
}

// NewSpinner creates a styled spinner with the given suffix message.
func NewSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 120*time.Millisecond)
	s.Suffix = "  " + message
	s.Color("cyan")
	return &Spinner{s: s}
}

// Start begins the spinner animation.
func (sp *Spinner) Start() { sp.s.Start() }

// Stop halts the spinner and clears the line.
func (sp *Spinner) Stop() { sp.s.Stop() }

// Update changes the spinner suffix message.
func (sp *Spinner) Update(message string) { sp.s.Suffix = "  " + message }

// Summary holds the data for the final summary box.
type Summary struct {
	Duration    time.Duration
	CostStr     string
	Provider    string
	OutputDir   string
	OutputFiles []string
	FileCount   int
	NodeCount   int
	EdgeCount   int
	NoAI        bool
}

// PrintSummary prints the styled completion summary box.
func PrintSummary(s Summary) {
	files := strings.Join(s.OutputFiles, ", ")
	duration := s.Duration.Round(time.Millisecond).String()
	outputLine := fmt.Sprintf("%s/ (%s)", s.OutputDir, files)
	statsLine := fmt.Sprintf("%s files | %s nodes | %s edges",
		bold(fmt.Sprint(s.FileCount)),
		bold(fmt.Sprint(s.NodeCount)),
		bold(fmt.Sprint(s.EdgeCount)),
	)

	// Build content lines (plain text for width, styled for display)
	// Build plain-text rows to compute widths
	type row struct {
		plain  string // for width calculation (no ANSI, ASCII only)
		styled string // actual output with colors
	}

	makeRow := func(icon, label, styledVal, plainVal string) row {
		return row{
			plain:  fmt.Sprintf("  %s  %s : %s", icon, label, plainVal),
			styled: fmt.Sprintf("  %s  %s : %s", icon, label, styledVal),
		}
	}

	var rows []row
	rows = append(rows, makeRow("T", "Duration", bold(duration), duration))
	if !s.NoAI {
		rows = append(rows, makeRow("$", "Cost    ", bold(s.CostStr), s.CostStr))
	}
	rows = append(rows, makeRow("F", "Outputs ", outputLine, outputLine))
	plainStats := fmt.Sprintf("%d files | %d nodes | %d edges", s.FileCount, s.NodeCount, s.EdgeCount)
	rows = append(rows, makeRow("S", "Stats   ", statsLine, plainStats))

	// Find max plain width
	maxLen := len("  Analysis Complete!")
	for _, r := range rows {
		if len(r.plain) > maxLen {
			maxLen = len(r.plain)
		}
	}
	boxWidth := maxLen + 4

	border := strings.Repeat("─", boxWidth)
	emptyLine := dim("  │") + strings.Repeat(" ", boxWidth) + dim("│")

	// Icons mapped back
	icons := []string{"⏱ ", "💰", "📁", "📊"}
	if s.NoAI {
		icons = []string{"⏱ ", "📁", "📊"}
	}

	fmt.Println()
	fmt.Println(dim("  ┌" + border + "┐"))
	fmt.Println(emptyLine)

	titlePad := boxWidth - len("  Analysis Complete!")
	fmt.Println(dim("  │") + "  " + green("Analysis Complete!") + strings.Repeat(" ", titlePad) + dim("│"))
	fmt.Println(emptyLine)

	for i, r := range rows {
		// Replace ASCII placeholder with emoji in styled output
		icon := icons[i]
		pad := boxWidth - len(r.plain)
		if pad < 0 {
			pad = 0
		}
		// Emoji takes ~2 display cols vs 1 ASCII char, so subtract 1 from padding
		pad -= 1
		if pad < 0 {
			pad = 0
		}
		// Swap the placeholder char in styled with the emoji
		styledLine := "  " + icon + r.styled[4:]
		fmt.Println(dim("  │") + styledLine + strings.Repeat(" ", pad) + dim("│"))
	}

	fmt.Println(emptyLine)
	fmt.Println(dim("  └" + border + "┘"))
	fmt.Println()
}


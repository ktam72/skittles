package main

import (
	"fmt"
	"os"

	"github.com/ktam/skittles/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	switchToEnglishInput()

	leftDir := ""
	rightDir := ""

	args := os.Args[1:]
	switch len(args) {
	case 1:
		leftDir = args[0]
		rightDir = args[0]
	case 2:
		leftDir = args[0]
		rightDir = args[1]
	default:
		var err error
		leftDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		rightDir = leftDir
	}

	m := ui.NewModel(leftDir, rightDir)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

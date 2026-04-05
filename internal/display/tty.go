package display

import (
	"os"

	"golang.org/x/term"
)

// isTerminal reports whether f is connected to a terminal.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

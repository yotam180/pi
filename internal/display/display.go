package display

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ANSI escape sequences for styling terminal output.
const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	red       = "\033[31m"
	green     = "\033[32m"
	yellow    = "\033[33m"
	boldRed   = "\033[1;31m"
	boldGreen = "\033[1;32m"
)

// StatusKind represents the semantic status of a display line (success,
// failure, in-progress, etc.). Methods like InstallStatus and PackageFetch
// accept StatusKind instead of raw icon strings, giving compile-time safety
// and centralised icon→style mapping.
type StatusKind string

const (
	StatusSuccess       StatusKind = "success"        // ✓ — bold green (newly installed, newly fetched)
	StatusSuccessCached StatusKind = "success_cached"  // ✓ — dim (already installed, already cached)
	StatusInProgress    StatusKind = "in_progress"     // → or ↓ — plain
	StatusFailed        StatusKind = "failed"          // ✗ — bold red
	StatusWarning       StatusKind = "warning"         // ⚠ — yellow
)

// statusIcon returns the Unicode icon for a StatusKind.
func statusIcon(kind StatusKind) string {
	switch kind {
	case StatusSuccess, StatusSuccessCached:
		return "✓"
	case StatusInProgress:
		return "→"
	case StatusFailed:
		return "✗"
	case StatusWarning:
		return "⚠"
	default:
		return "?"
	}
}

// Printer writes styled output to a writer. When color is disabled
// (non-TTY, NO_COLOR set, or explicitly disabled), all style methods
// produce plain text.
type Printer struct {
	w     io.Writer
	color bool
}

// New creates a Printer that writes to w. Color is enabled only when
// w is a TTY and the NO_COLOR environment variable is not set.
func New(w io.Writer) *Printer {
	return &Printer{w: w, color: shouldColor(w)}
}

// NewForWriter creates a Printer for an arbitrary io.Writer. If the writer
// is a *os.File backed by a terminal, color is auto-detected (same as New).
// Otherwise, color is disabled. Use this when the concrete type of the writer
// is not known at the call site.
func NewForWriter(w io.Writer) *Printer {
	if f, ok := w.(*os.File); ok {
		return New(f)
	}
	return &Printer{w: w, color: false}
}

// NewWithColor creates a Printer with an explicit color toggle.
// Useful for testing or when the caller has already determined the mode.
func NewWithColor(w io.Writer, color bool) *Printer {
	return &Printer{w: w, color: color}
}

// Plain writes unformatted text (respects Printf-style format).
func (p *Printer) Plain(format string, a ...any) {
	fmt.Fprintf(p.w, format, a...)
}

// Dim writes text in dim/grey style.
func (p *Printer) Dim(format string, a ...any) {
	p.styled(dim, format, a...)
}

// Green writes text in bold green style.
func (p *Printer) Green(format string, a ...any) {
	p.styled(boldGreen, format, a...)
}

// Red writes text in bold red style.
func (p *Printer) Red(format string, a ...any) {
	p.styled(boldRed, format, a...)
}

// Bold writes text in bold style.
func (p *Printer) Bold(format string, a ...any) {
	p.styled(bold, format, a...)
}

// Warn writes text in yellow style (for non-fatal warnings).
func (p *Printer) Warn(format string, a ...any) {
	p.styled(yellow, format, a...)
}

// InstallStatus prints a formatted installer status line styled according
// to the StatusKind:
//   - StatusSuccessCached → dim (already installed)
//   - StatusSuccess → bold green (newly installed)
//   - StatusInProgress → plain (installing...)
//   - StatusFailed → bold red
//   - StatusWarning → yellow
func (p *Printer) InstallStatus(kind StatusKind, name, status, version string) {
	icon := statusIcon(kind)
	var line string
	if version != "" {
		line = fmt.Sprintf("  %s  %-25s %s (%s)\n", icon, name, status, version)
	} else {
		line = fmt.Sprintf("  %s  %-25s %s\n", icon, name, status)
	}

	p.printStatusLine(kind, line)
}

// PackageFetch prints a package fetch status line styled according to
// the StatusKind:
//   - StatusSuccessCached → dim (already cached / found)
//   - StatusSuccess → bold green (newly fetched)
//   - StatusInProgress → plain (fetching...)
//   - StatusFailed → bold red
//   - StatusWarning → yellow (source not found)
func (p *Printer) PackageFetch(kind StatusKind, source, status, detail string) {
	icon := statusIcon(kind)
	if kind == StatusInProgress {
		icon = "↓"
	}
	var line string
	if detail != "" {
		line = fmt.Sprintf("  %s  %-35s %s (%s)\n", icon, source, status, detail)
	} else {
		line = fmt.Sprintf("  %s  %-35s %s\n", icon, source, status)
	}

	p.printStatusLine(kind, line)
}

// printStatusLine routes a pre-formatted line to the appropriate style method
// based on StatusKind. Shared by InstallStatus and PackageFetch.
func (p *Printer) printStatusLine(kind StatusKind, line string) {
	switch kind {
	case StatusFailed:
		p.Red("%s", line)
	case StatusSuccessCached:
		p.Dim("%s", line)
	case StatusSuccess:
		p.Green("%s", line)
	case StatusWarning:
		p.Warn("%s", line)
	default:
		p.Plain("%s", line)
	}
}

// SetupHeader prints a setup entry header line (e.g. "==> setup[0]: docker/up")
// in dim style.
func (p *Printer) SetupHeader(format string, a ...any) {
	p.Dim(format, a...)
}

// StepTrace prints a step execution trace line (e.g. "  → bash: echo hello")
// in dim style.
func (p *Printer) StepTrace(stepType, value string) {
	truncated := truncateTrace(value, 80)
	p.Dim("  → %s: %s\n", stepType, truncated)
}

// truncateTrace shortens a value to maxLen, collapsing newlines.
func truncateTrace(s string, maxLen int) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i] + "..."
	}
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// styled wraps text in ANSI codes when color is enabled.
func (p *Printer) styled(code, format string, a ...any) {
	if !p.color {
		fmt.Fprintf(p.w, format, a...)
		return
	}
	text := fmt.Sprintf(format, a...)
	fmt.Fprintf(p.w, "%s%s%s", code, text, reset)
}

// shouldColor returns true when the writer supports color output.
// Color is disabled when:
//   - NO_COLOR environment variable is set (any value)
//   - The writer is not a *os.File backed by a terminal
func shouldColor(w io.Writer) bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isTerminal(f)
}

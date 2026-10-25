// Package ui provides terminal styling, TTY detection, NO_COLOR compliance,
// and formatted tabular/diagnostic presentations for the pkt-git CLI.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dustin/go-humanize"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
)

var (
	mu         sync.RWMutex
	forcedTTY  *bool
	forcedNoColor *bool
)

// SetForceTTY overrides TTY detection for testing. Call ResetOverrides to clear.
func SetForceTTY(v bool) {
	mu.Lock()
	defer mu.Unlock()
	forcedTTY = &v
}

// SetForceNoColor overrides NO_COLOR detection for testing. Call ResetOverrides to clear.
func SetForceNoColor(v bool) {
	mu.Lock()
	defer mu.Unlock()
	forcedNoColor = &v
}

// ResetOverrides clears any forced TTY or NO_COLOR settings.
func ResetOverrides() {
	mu.Lock()
	defer mu.Unlock()
	forcedTTY = nil
	forcedNoColor = nil
}

// IsTTY reports whether w is an interactive terminal.
func IsTTY(w io.Writer) bool {
	mu.RLock()
	if forcedTTY != nil {
		defer mu.RUnlock()
		return *forcedTTY
	}
	mu.RUnlock()

	if f, ok := w.(*os.File); ok {
		fd := f.Fd()
		return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
	}
	return false
}

// ColorEnabled reports whether ANSI color and styling should be used on w.
func ColorEnabled(w io.Writer) bool {
	mu.RLock()
	if forcedNoColor != nil {
		noColor := *forcedNoColor
		mu.RUnlock()
		if noColor {
			return false
		}
		return IsTTY(w)
	}
	mu.RUnlock()

	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	return IsTTY(w)
}

// Success writes a success message with green ✔ on TTY or [ok] in plain text.
func Success(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if ColorEnabled(w) {
		r := rendererFor(w)
		sym := r.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("✔")
		fmt.Fprintf(w, "%s %s\n", sym, msg)
	} else {
		fmt.Fprintf(w, "[ok] %s\n", msg)
	}
}

// Warning writes a warning message with yellow ⚠ on TTY or [warn] in plain text.
func Warning(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if ColorEnabled(w) {
		r := rendererFor(w)
		sym := r.NewStyle().Foreground(lipgloss.Color("11")).Bold(true).Render("⚠")
		fmt.Fprintf(w, "%s %s\n", sym, msg)
	} else {
		fmt.Fprintf(w, "[warn] %s\n", msg)
	}
}

// Error writes an error message with red ✖ on TTY or pkt: in plain text.
func Error(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if ColorEnabled(w) {
		r := rendererFor(w)
		sym := r.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("✖")
		fmt.Fprintf(w, "%s %s\n", sym, msg)
	} else {
		fmt.Fprintf(w, "pkt: %s\n", msg)
	}
}

// Info writes an informational message with blue ℹ on TTY or [info] in plain text.
func Info(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if ColorEnabled(w) {
		r := rendererFor(w)
		sym := r.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render("ℹ")
		fmt.Fprintf(w, "%s %s\n", sym, msg)
	} else {
		fmt.Fprintf(w, "[info] %s\n", msg)
	}
}

// FormatBytes returns a human-readable representation of a byte count.
func FormatBytes(n int64) string {
	if n < 0 {
		return "0 B"
	}
	return humanize.Bytes(uint64(n))
}

// DeviceBadge returns a color-coded device category tag or plain text.
func DeviceBadge(w io.Writer, deviceType string) string {
	if !ColorEnabled(w) {
		return deviceType
	}
	r := rendererFor(w)
	style := r.NewStyle().Bold(true)
	switch strings.ToLower(deviceType) {
	case "router":
		return style.Foreground(lipgloss.Color("39")).Render(deviceType)
	case "switch", "multilayerswitch":
		return style.Foreground(lipgloss.Color("42")).Render(deviceType)
	case "server":
		return style.Foreground(lipgloss.Color("207")).Render(deviceType)
	case "accesspoint":
		return style.Foreground(lipgloss.Color("214")).Render(deviceType)
	case "pc", "laptop", "pda":
		return style.Foreground(lipgloss.Color("248")).Render(deviceType)
	default:
		return deviceType
	}
}

// RenderTable formats a 2D grid into an aligned table.
func RenderTable(w io.Writer, headers []string, rows [][]string) string {
	t := table.New().
		Headers(headers...).
		Rows(rows...)

	if ColorEnabled(w) {
		r := rendererFor(w)
		t.Border(lipgloss.RoundedBorder()).
			BorderStyle(r.NewStyle().Foreground(lipgloss.Color("240"))).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return r.NewStyle().
						Bold(true).
						Foreground(lipgloss.Color("39")).
						Padding(0, 1)
				}
				return r.NewStyle().Padding(0, 1)
			})
	} else {
		t.Border(lipgloss.NormalBorder()).
			StyleFunc(func(row, col int) lipgloss.Style {
				return lipgloss.NewStyle().Padding(0, 1)
			})
	}

	return t.Render()
}

func rendererFor(w io.Writer) *lipgloss.Renderer {
	r := lipgloss.NewRenderer(w)
	if !ColorEnabled(w) {
		r.SetColorProfile(termenv.Ascii)
	}
	return r
}

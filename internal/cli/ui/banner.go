// Package ui provides terminal styling, banners, spinners, and diagnostics for pkt-git.
package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const asciiBanner = `▄▄▄▄  ▄  ▄ ▄▄▄▄▄   ▄▄▄▄▄ ▄ ▄▄▄▄▄
█   █ █▄▀    █     █     █   █
█▀▀▀▀ █ ▀▄   █     █ ▀▀█ █   █
▀     ▀  ▀   ▀     ▀▀▀▀▀ ▀ ▀▀▀▀▀`

const bannerSubtitle = "Cisco Packet Tracer ⇄ Git Version Control"

// Banner writes the styled or plain pkt-git ASCII art banner to w.
func Banner(w io.Writer) {
	if ColorEnabled(w) {
		r := rendererFor(w)
		pktStyle := r.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		subStyle := r.NewStyle().Foreground(lipgloss.Color("245"))
		fmt.Fprintln(w, pktStyle.Render(asciiBanner))
		fmt.Fprintf(w, "  %s\n\n", subStyle.Render(bannerSubtitle))
	} else {
		fmt.Fprintln(w, asciiBanner)
		fmt.Fprintf(w, "  %s\n\n", bannerSubtitle)
	}
}

// BannerString returns the banner and subtitle as a formatted string.
func BannerString(color bool) string {
	var sb strings.Builder
	if color {
		r := rendererFor(&sb)
		pktStyle := r.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		subStyle := r.NewStyle().Foreground(lipgloss.Color("245"))
		fmt.Fprintln(&sb, pktStyle.Render(asciiBanner))
		fmt.Fprintf(&sb, "  %s\n", subStyle.Render(bannerSubtitle))
	} else {
		fmt.Fprintln(&sb, asciiBanner)
		fmt.Fprintf(&sb, "  %s\n", bannerSubtitle)
	}
	return sb.String()
}

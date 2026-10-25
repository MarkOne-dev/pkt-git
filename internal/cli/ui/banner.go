// Package ui provides terminal styling, banners, spinners, and diagnostics for pkt-git.
package ui

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/MarkOne-dev/pkt-git/internal/config"
)

const asciiBanner = `▄▄▄▄  ▄  ▄ ▄▄▄▄▄   ▄▄▄▄▄ ▄ ▄▄▄▄▄
█   █ █▄▀    █     █     █   █
█▀▀▀▀ █ ▀▄   █     █ ▀▀█ █   █
▀     ▀  ▀   ▀     ▀▀▀▀▀ ▀ ▀▀▀▀▀`

const bannerSubtitle = "Cisco Packet Tracer ⇄ Git Version Control"

func centerLine(s string, width int) string {
	runes := []rune(s)
	runeLen := len(runes)
	if runeLen >= width {
		return s
	}
	padding := (width - runeLen) / 2
	return strings.Repeat(" ", padding) + s
}

func metadataLine() string {
	ver := config.Version
	if ver == "" {
		ver = "dev"
	}
	if !strings.HasPrefix(ver, "v") && ver != "dev" {
		ver = "v" + ver
	}
	return fmt.Sprintf("%s • %s • %s/%s", ver, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func targetWidth(meta string) int {
	w := 46
	if runeLen := len([]rune(meta)); runeLen > w {
		w = runeLen + 4
	}
	return w
}

// Banner writes the centered pkt-git banner, subtitle, and runtime metadata to w.
func Banner(w io.Writer) {
	lines := strings.Split(asciiBanner, "\n")
	meta := metadataLine()
	wWidth := targetWidth(meta)

	if ColorEnabled(w) {
		r := rendererFor(w)
		pktStyle := r.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		subStyle := r.NewStyle().Foreground(lipgloss.Color("248"))
		metaStyle := r.NewStyle().Foreground(lipgloss.Color("243"))

		for _, line := range lines {
			fmt.Fprintln(w, pktStyle.Render(centerLine(line, wWidth)))
		}
		fmt.Fprintln(w, subStyle.Render(centerLine(bannerSubtitle, wWidth)))
		fmt.Fprintf(w, "%s\n\n", metaStyle.Render(centerLine(meta, wWidth)))
	} else {
		for _, line := range lines {
			fmt.Fprintln(w, centerLine(line, wWidth))
		}
		fmt.Fprintln(w, centerLine(bannerSubtitle, wWidth))
		fmt.Fprintf(w, "%s\n\n", centerLine(meta, wWidth))
	}
}

// BannerString returns the centered banner, subtitle, and metadata as a formatted string.
func BannerString(color bool) string {
	var sb strings.Builder
	lines := strings.Split(asciiBanner, "\n")
	meta := metadataLine()
	wWidth := targetWidth(meta)

	if color {
		r := rendererFor(&sb)
		pktStyle := r.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		subStyle := r.NewStyle().Foreground(lipgloss.Color("248"))
		metaStyle := r.NewStyle().Foreground(lipgloss.Color("243"))

		for _, line := range lines {
			fmt.Fprintln(&sb, pktStyle.Render(centerLine(line, wWidth)))
		}
		fmt.Fprintln(&sb, subStyle.Render(centerLine(bannerSubtitle, wWidth)))
		fmt.Fprintf(&sb, "%s\n", metaStyle.Render(centerLine(meta, wWidth)))
	} else {
		for _, line := range lines {
			fmt.Fprintln(&sb, centerLine(line, wWidth))
		}
		fmt.Fprintln(&sb, centerLine(bannerSubtitle, wWidth))
		fmt.Fprintf(&sb, "%s\n", centerLine(meta, wWidth))
	}
	return sb.String()
}

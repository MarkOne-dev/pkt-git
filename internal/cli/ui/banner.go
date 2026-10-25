// Package ui provides terminal styling, banners, spinners, and diagnostics for pkt-git.
package ui

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/MarkOne-dev/pkt-git/internal/config"
	"github.com/MarkOne-dev/pkt-git/internal/i18n"
)

var asciiBannerLines = []string{
	"▄▄▄▄  ▄  ▄ ▄▄▄▄▄   ▄▄▄▄▄ ▄ ▄▄▄▄▄",
	"█   █ █▄▀    █     █     █   █  ",
	"█▀▀▀▀ █ ▀▄   █     █ ▀▀█ █   █  ",
	"▀     ▀  ▀   ▀     ▀▀▀▀▀ ▀   ▀  ",
}

func subtitleText() string {
	return i18n.T("BannerSubtitle")
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

func calculateLayout(meta, subtitle string) (artIndent, subIndent, metaIndent string) {
	artWidth := len([]rune(asciiBannerLines[0]))
	subWidth := len([]rune(subtitle))
	metaWidth := len([]rune(meta))

	containerWidth := 46
	if subWidth > containerWidth {
		containerWidth = subWidth
	}
	if metaWidth > containerWidth {
		containerWidth = metaWidth
	}

	artPad := (containerWidth - artWidth) / 2
	subPad := (containerWidth - subWidth) / 2
	metaPad := (containerWidth - metaWidth) / 2

	if artPad < 0 {
		artPad = 0
	}
	if subPad < 0 {
		subPad = 0
	}
	if metaPad < 0 {
		metaPad = 0
	}

	return strings.Repeat(" ", artPad), strings.Repeat(" ", subPad), strings.Repeat(" ", metaPad)
}

// Banner writes the centered pkt-git banner, subtitle, and runtime metadata to w.
func Banner(w io.Writer) {
	meta := metadataLine()
	sub := subtitleText()
	artIndent, subIndent, metaIndent := calculateLayout(meta, sub)

	if ColorEnabled(w) {
		r := rendererFor(w)
		pktStyle := r.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		subStyle := r.NewStyle().Foreground(lipgloss.Color("248"))
		metaStyle := r.NewStyle().Foreground(lipgloss.Color("243"))

		for _, line := range asciiBannerLines {
			fmt.Fprintln(w, pktStyle.Render(artIndent+line))
		}
		fmt.Fprintln(w, subStyle.Render(subIndent+sub))
		fmt.Fprintf(w, "%s\n\n", metaStyle.Render(metaIndent+meta))
	} else {
		for _, line := range asciiBannerLines {
			fmt.Fprintln(w, artIndent+line)
		}
		fmt.Fprintln(w, subIndent+sub)
		fmt.Fprintf(w, "%s\n\n", metaIndent+meta)
	}
}

// BannerString returns the centered banner, subtitle, and metadata as a formatted string.
func BannerString(color bool) string {
	var sb strings.Builder
	meta := metadataLine()
	sub := subtitleText()
	artIndent, subIndent, metaIndent := calculateLayout(meta, sub)

	if color {
		r := rendererFor(&sb)
		pktStyle := r.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		subStyle := r.NewStyle().Foreground(lipgloss.Color("248"))
		metaStyle := r.NewStyle().Foreground(lipgloss.Color("243"))

		for _, line := range asciiBannerLines {
			fmt.Fprintln(&sb, pktStyle.Render(artIndent+line))
		}
		fmt.Fprintln(&sb, subStyle.Render(subIndent+sub))
		fmt.Fprintf(&sb, "%s\n", metaStyle.Render(metaIndent+meta))
	} else {
		for _, line := range asciiBannerLines {
			fmt.Fprintln(&sb, artIndent+line)
		}
		fmt.Fprintln(&sb, subIndent+sub)
		fmt.Fprintf(&sb, "%s\n", metaIndent+meta)
	}
	return sb.String()
}

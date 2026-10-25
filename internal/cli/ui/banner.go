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

var asciiBannerLines = []string{
	"▄▄▄▄  ▄  ▄ ▄▄▄▄▄   ▄▄▄▄▄ ▄ ▄▄▄▄▄",
	"█   █ █▄▀    █     █     █   █  ",
	"█▀▀▀▀ █ ▀▄   █     █ ▀▀█ █   █  ",
	"▀     ▀  ▀   ▀     ▀▀▀▀▀ ▀ ▀▀▀▀▀",
}

const bannerSubtitle = "Cisco Packet Tracer ⇄ Git Version Control"

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

func calculateLayout(meta string) (artIndent, subIndent, metaIndent string) {
	artWidth := len([]rune(asciiBannerLines[0]))
	subWidth := len([]rune(bannerSubtitle))
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
	artIndent, subIndent, metaIndent := calculateLayout(meta)

	if ColorEnabled(w) {
		r := rendererFor(w)
		pktStyle := r.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		subStyle := r.NewStyle().Foreground(lipgloss.Color("248"))
		metaStyle := r.NewStyle().Foreground(lipgloss.Color("243"))

		for _, line := range asciiBannerLines {
			fmt.Fprintln(w, pktStyle.Render(artIndent+line))
		}
		fmt.Fprintln(w, subStyle.Render(subIndent+bannerSubtitle))
		fmt.Fprintf(w, "%s\n\n", metaStyle.Render(metaIndent+meta))
	} else {
		for _, line := range asciiBannerLines {
			fmt.Fprintln(w, artIndent+line)
		}
		fmt.Fprintln(w, subIndent+bannerSubtitle)
		fmt.Fprintf(w, "%s\n\n", metaIndent+meta)
	}
}

// BannerString returns the centered banner, subtitle, and metadata as a formatted string.
func BannerString(color bool) string {
	var sb strings.Builder
	meta := metadataLine()
	artIndent, subIndent, metaIndent := calculateLayout(meta)

	if color {
		r := rendererFor(&sb)
		pktStyle := r.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		subStyle := r.NewStyle().Foreground(lipgloss.Color("248"))
		metaStyle := r.NewStyle().Foreground(lipgloss.Color("243"))

		for _, line := range asciiBannerLines {
			fmt.Fprintln(&sb, pktStyle.Render(artIndent+line))
		}
		fmt.Fprintln(&sb, subStyle.Render(subIndent+bannerSubtitle))
		fmt.Fprintf(&sb, "%s\n", metaStyle.Render(metaIndent+meta))
	} else {
		for _, line := range asciiBannerLines {
			fmt.Fprintln(&sb, artIndent+line)
		}
		fmt.Fprintln(&sb, subIndent+bannerSubtitle)
		fmt.Fprintf(&sb, "%s\n", metaIndent+meta)
	}
	return sb.String()
}

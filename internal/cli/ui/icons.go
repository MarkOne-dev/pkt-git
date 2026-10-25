package ui

import "io"

// Standard Nerd Font glyphs used across the pkt-git CLI.
const (
	IconRocket   = "\uf135" // 󱄵 nf-md-rocket
	IconTerminal = "\uf120" //  nf-fa-terminal
	IconGit      = "\ue702" //  nf-dev-git
	IconFolder   = "\uf07b" //  nf-fa-folder
	IconCheck    = "\uf00c" //  nf-fa-check
	IconCross    = "\uf00d" //  nf-fa-times
	IconWarn     = "\uf071" //  nf-fa-warning
	IconInfo     = "\uf05a" //  nf-fa-info_circle
)

// Glyph returns the specified Nerd Font glyph when color and TTY are enabled,
// or the fallback string when operating in plain-text or non-interactive mode.
func Glyph(w io.Writer, nfGlyph, fallback string) string {
	if ColorEnabled(w) {
		return nfGlyph
	}
	return fallback
}

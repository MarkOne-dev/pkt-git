package cli

import (
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/config"
)

func newWelcomeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "welcome",
		Short: "Display branded welcome banner and onboarding guide",
		Args:  maxArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()

			// 1. Display ASCII Banner
			ui.Banner(out)

			// 2. Animated environment check
			spinner := ui.NewSpinner(out, "Verifying environment prerequisites...")
			spinner.Start()

			// Check Git
			var gitReady bool
			if _, err := exec.LookPath("git"); err == nil {
				gitReady = true
			}

			// Brief pause for visual feedback when interactive
			if ui.ColorEnabled(out) {
				time.Sleep(120 * time.Millisecond)
			}

			if gitReady {
				spinner.Success("Environment ready (Git installed, pkt-git v%s)", config.Version)
			} else {
				spinner.Fail("Git not found on PATH. Run 'pkt doctor' for remediation instructions.")
			}

			// 3. Quickstart guide
			renderQuickstart(out)

			return nil
		},
	}
}

func renderQuickstart(w io.Writer) {
	fmt.Fprintln(w)
	if ui.ColorEnabled(w) {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		stepNumStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
		codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		commentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

		fmt.Fprintln(w, titleStyle.Render("🚀 Quickstart Guide:"))
		fmt.Fprintf(w, "  %s %s\n", stepNumStyle.Render("1."), "Initialize a Packet Tracer project repository:")
		fmt.Fprintf(w, "     %s %s\n", codeStyle.Render("pkt init"), commentStyle.Render("# wires git hooks & exports existing topology"))
		fmt.Fprintf(w, "  %s %s\n", stepNumStyle.Render("2."), "Normal Git commits:")
		fmt.Fprintf(w, "     %s %s\n", codeStyle.Render("git add . && git commit -m \"update lab\""), commentStyle.Render("# auto-syncs XML and .cfg files"))
		fmt.Fprintf(w, "  %s %s\n", stepNumStyle.Render("3."), "Diagnostics & AI Agent Skills:")
		fmt.Fprintf(w, "     %s %s\n", codeStyle.Render("pkt doctor"), commentStyle.Render("# inspect environment health"))
		fmt.Fprintf(w, "     %s %s\n", codeStyle.Render("pkt skills"), commentStyle.Render("# enable AI assistant skill in Claude/Gemini"))
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  Run %s for all commands.\n", codeStyle.Render("pkt --help"))
	} else {
		fmt.Fprintln(w, "Quickstart Guide:")
		fmt.Fprintln(w, "  1. Initialize a Packet Tracer project repository:")
		fmt.Fprintln(w, "     pkt init          # wires git hooks & exports existing topology")
		fmt.Fprintln(w, "  2. Normal Git commits:")
		fmt.Fprintln(w, "     git add . && git commit -m \"update lab\"  # auto-syncs XML and .cfg files")
		fmt.Fprintln(w, "  3. Diagnostics & AI Agent Skills:")
		fmt.Fprintln(w, "     pkt doctor        # inspect environment health")
		fmt.Fprintln(w, "     pkt skills        # enable AI assistant skill in Claude/Gemini")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  Run 'pkt --help' for all commands.")
	}
}

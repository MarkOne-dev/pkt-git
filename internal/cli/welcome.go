package cli

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/config"
	"github.com/MarkOne-dev/pkt-git/internal/i18n"
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
			spinner := ui.NewSpinner(out, i18n.T("WelcomeVerifying"))
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
				ver := config.Version
				if ver == "" {
					ver = "dev"
				}
				if !strings.HasPrefix(ver, "v") && ver != "dev" {
					ver = "v" + ver
				}
				spinner.Success("%s", i18n.T("WelcomeReady", map[string]interface{}{"Version": ver}))
			} else {
				spinner.Fail("%s", i18n.T("DoctorGitMissing"))
			}

			// 3. Quickstart guide
			renderQuickstart(out)

			return nil
		},
	}
}

func renderQuickstart(w io.Writer) {
	fmt.Fprintln(w)
	title := i18n.T("WelcomeQuickstartTitle")
	step1Title := i18n.T("WelcomeStep1Title")
	step1Comment := i18n.T("WelcomeStep1Comment")
	step2Title := i18n.T("WelcomeStep2Title")
	step2Comment := i18n.T("WelcomeStep2Comment")
	step3Title := i18n.T("WelcomeStep3Title")
	step3Doctor := i18n.T("WelcomeStep3Doctor")
	step3Skills := i18n.T("WelcomeStep3Skills")

	if ui.ColorEnabled(w) {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		stepNumStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
		codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		commentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

		fmt.Fprintln(w, titleStyle.Render(ui.IconRocket+" "+title))
		fmt.Fprintf(w, "  %s %s\n", stepNumStyle.Render("1."), step1Title)
		fmt.Fprintf(w, "     %s %s\n", codeStyle.Render("pkt init"), commentStyle.Render(step1Comment))
		fmt.Fprintf(w, "  %s %s\n", stepNumStyle.Render("2."), step2Title)
		fmt.Fprintf(w, "     %s %s\n", codeStyle.Render("git add . && git commit -m \"update lab\""), commentStyle.Render(step2Comment))
		fmt.Fprintf(w, "  %s %s\n", stepNumStyle.Render("3."), step3Title)
		fmt.Fprintf(w, "     %s %s\n", codeStyle.Render("pkt doctor"), commentStyle.Render(step3Doctor))
		fmt.Fprintf(w, "     %s %s\n", codeStyle.Render("pkt skills"), commentStyle.Render(step3Skills))
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s\n", i18n.T("WelcomeRunHelp", map[string]interface{}{"Command": codeStyle.Render("pkt --help")}))
	} else {
		fmt.Fprintln(w, title)
		fmt.Fprintf(w, "  1. %s\n", step1Title)
		fmt.Fprintf(w, "     pkt init          %s\n", step1Comment)
		fmt.Fprintf(w, "  2. %s\n", step2Title)
		fmt.Fprintf(w, "     git add . && git commit -m \"update lab\"  %s\n", step2Comment)
		fmt.Fprintf(w, "  3. %s\n", step3Title)
		fmt.Fprintf(w, "     pkt doctor        %s\n", step3Doctor)
		fmt.Fprintf(w, "     pkt skills        %s\n", step3Skills)
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s\n", i18n.T("WelcomeRunHelp", map[string]interface{}{"Command": "'pkt --help'"}))
	}
}

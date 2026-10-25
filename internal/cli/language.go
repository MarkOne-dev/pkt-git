package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/config"
	"github.com/MarkOne-dev/pkt-git/internal/i18n"
)

type langOption struct {
	Code string
	Name string
	Desc string
}

var supportedLanguages = []langOption{
	{Code: "en", Name: "English", Desc: "Default English interface"},
	{Code: "es", Name: "Español", Desc: "Interfaz predeterminada en español"},
}

// StdinReader allows mocking stdin in unit tests.
var StdinReader io.Reader = os.Stdin

// IsInteractiveFunc allows mocking interactive TTY checks in unit tests.
var IsInteractiveFunc = func(w io.Writer) bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && ui.ColorEnabled(w)
}

func newLanguageCmd() *cobra.Command {
	var localScope bool

	cmd := &cobra.Command{
		Use:     "language [code]",
		Aliases: []string{"lang"},
		Short:   "View or set the preferred interface language (en, es)",
		Long: `View or set the preferred interface language for pkt-git.
Language preferences are persisted in Git configuration (pkt.language).
Supported languages: en (English), es (Español).`,
		Example: `  pkt language          # Interactive language selection
  pkt language es       # Set language to Spanish globally
  pkt lang en --local   # Set language to English for current repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			current := i18n.CurrentLanguage()

			if len(args) == 1 {
				targetCode := strings.ToLower(strings.TrimSpace(args[0]))
				return setLanguage(out, targetCode, localScope)
			}

			if IsInteractiveFunc(out) {
				return runInteractiveMenu(out, current, localScope)
			}

			renderNonInteractive(out, current)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&localScope, "local", "l", false, "save language preference in current repository (.git/config)")
	return cmd
}

func setLanguage(w io.Writer, code string, local bool) error {
	var matched *langOption
	for i := range supportedLanguages {
		if supportedLanguages[i].Code == code {
			matched = &supportedLanguages[i]
			break
		}
	}
	if matched == nil {
		return usagef("%s", i18n.T("LanguageInvalidCode", map[string]interface{}{"Code": code}))
	}

	// Persist to git config
	args := []string{"config"}
	scopeStr := "global (~/.gitconfig)"
	if local {
		args = append(args, "--local")
		scopeStr = "local (.git/config)"
	} else {
		args = append(args, "--global")
	}
	args = append(args, config.GitConfigLanguage, code)

	cmd := exec.Command("git", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to save git config %s: %s (%w)", config.GitConfigLanguage, strings.TrimSpace(string(out)), err)
	}

	// Update active localizer
	i18n.Init(code)

	msg := i18n.T("LanguageUpdated", map[string]interface{}{
		"Language": matched.Name,
		"Code":     matched.Code,
	})
	scopeMsg := i18n.T("LanguagePersistedScope", map[string]interface{}{
		"Target": "git config " + scopeStr,
	})

	ui.Success(w, "%s", msg)
	fmt.Fprintf(w, "  %s\n", scopeMsg)
	return nil
}

func runInteractiveMenu(w io.Writer, current string, local bool) error {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	numStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	commentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	fmt.Fprintln(w)
	fmt.Fprintln(w, titleStyle.Render(ui.Glyph(w, "\uf1ab ", "")+i18n.T("LanguageMenuTitle")))
	fmt.Fprintln(w)

	for i, opt := range supportedLanguages {
		isActive := opt.Code == current
		prefix := "[ ]"
		if isActive {
			prefix = activeStyle.Render("[✔]")
		}
		num := numStyle.Render(fmt.Sprintf("%d.", i+1))
		fmt.Fprintf(w, "  %s %s %-8s %s\n", num, prefix, opt.Name, commentStyle.Render("("+opt.Code+") - "+opt.Desc))
	}

	fmt.Fprintln(w)
	fmt.Fprint(w, i18n.T("LanguageMenuPrompt"))

	reader := bufio.NewReader(StdinReader)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}

	choice := strings.ToLower(strings.TrimSpace(input))
	if choice == "" {
		fmt.Fprintln(w)
		return nil
	}

	targetCode := ""
	switch choice {
	case "1", "en", "english":
		targetCode = "en"
	case "2", "es", "español", "espanol", "spanish":
		targetCode = "es"
	default:
		return usagef("%s", i18n.T("LanguageInvalidCode", map[string]interface{}{"Code": choice}))
	}

	fmt.Fprintln(w)
	return setLanguage(w, targetCode, local)
}

func renderNonInteractive(w io.Writer, current string) {
	currentName := "English"
	for _, opt := range supportedLanguages {
		if opt.Code == current {
			currentName = opt.Name
			break
		}
	}

	fmt.Fprintln(w, i18n.T("LanguageCurrent", map[string]interface{}{
		"Language": currentName,
		"Code":     current,
	}))
	fmt.Fprintln(w)
	fmt.Fprintln(w, i18n.T("LanguageAvailable"))
	for _, opt := range supportedLanguages {
		fmt.Fprintf(w, "  • %-4s %s\n", opt.Code, opt.Name)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, i18n.T("LanguageUsage"))
	fmt.Fprintln(w, "  pkt language <en|es> [--local]")
}

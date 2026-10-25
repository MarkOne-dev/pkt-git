package cli

import (
	"io"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/i18n"
)

func TestLanguageNonInteractive(t *testing.T) {
	oldInteractive := IsInteractiveFunc
	defer func() { IsInteractiveFunc = oldInteractive }()
	IsInteractiveFunc = func(w io.Writer) bool { return false }

	code, stdout, stderr := runCLI("--lang", "en", "language")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "Current language:") {
		t.Fatalf("stdout missing 'Current language:': %q", stdout)
	}
	if !strings.Contains(stdout, "Available languages:") {
		t.Fatalf("stdout missing 'Available languages:': %q", stdout)
	}
	if !strings.Contains(stdout, "en") || !strings.Contains(stdout, "es") {
		t.Fatalf("stdout missing language list: %q", stdout)
	}
}

func TestLanguageAliasLang(t *testing.T) {
	oldInteractive := IsInteractiveFunc
	defer func() { IsInteractiveFunc = oldInteractive }()
	IsInteractiveFunc = func(w io.Writer) bool { return false }

	code, stdout, stderr := runCLI("--lang", "en", "lang")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "Current language:") {
		t.Fatalf("stdout missing 'Current language:': %q", stdout)
	}
}

func TestLanguageInvalidCode(t *testing.T) {
	code, _, stderr := runCLI("language", "xyz")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr=%q", code, stderr)
	}
	if !strings.Contains(stderr, "xyz") {
		t.Fatalf("stderr missing invalid code: %q", stderr)
	}
}

func TestLanguageDirectSwitch(t *testing.T) {
	t.Cleanup(func() {
		// Restore default language
		runCLI("language", "es", "--local")
	})

	// Switch to English locally
	code, stdout, stderr := runCLI("language", "en", "--local")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Language updated to English") && !strings.Contains(stdout, "Idioma actualizado a English") {
		t.Fatalf("stdout missing confirmation: %q", stdout)
	}
	if i18n.CurrentLanguage() != "en" {
		t.Errorf("CurrentLanguage() = %q, want 'en'", i18n.CurrentLanguage())
	}

	// Switch to Spanish locally
	code, stdout, stderr = runCLI("lang", "es", "--local")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Idioma actualizado a Español") && !strings.Contains(stdout, "Language updated to Español") {
		t.Fatalf("stdout missing confirmation: %q", stdout)
	}
	if i18n.CurrentLanguage() != "es" {
		t.Errorf("CurrentLanguage() = %q, want 'es'", i18n.CurrentLanguage())
	}
}

func TestLanguageInteractiveMenu(t *testing.T) {
	oldInteractive := IsInteractiveFunc
	oldStdin := StdinReader
	defer func() {
		IsInteractiveFunc = oldInteractive
		StdinReader = oldStdin
	}()

	IsInteractiveFunc = func(w io.Writer) bool { return true }
	StdinReader = strings.NewReader("1\n") // Select option 1 (English)

	code, stdout, stderr := runCLI("language", "--local")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "English") {
		t.Fatalf("stdout missing English in interactive output:\n%s", stdout)
	}
}

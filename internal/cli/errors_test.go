package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/codec"
	"github.com/MarkOne-dev/pkt-git/internal/hooks"
	"github.com/MarkOne-dev/pkt-git/internal/i18n"
)

func TestStructuredErrorTypes(t *testing.T) {
	inner := errors.New("underlying failure")

	repoErr := &RepoError{Message: "repo failed", Hint: "init repo", Err: inner}
	if !errors.Is(repoErr, ErrRepo) {
		t.Errorf("RepoError does not match ErrRepo")
	}
	if !errors.Is(repoErr, inner) {
		t.Errorf("RepoError does not unwrap to inner error")
	}
	if repoErr.Error() != "repo failed" {
		t.Errorf("RepoError.Error() = %q, want 'repo failed'", repoErr.Error())
	}

	codecErr := &CodecError{Message: "codec broken", Path: "test.pkt", Err: inner}
	if !errors.Is(codecErr, ErrCodec) {
		t.Errorf("CodecError does not match ErrCodec")
	}
	if !errors.Is(codecErr, inner) {
		t.Errorf("CodecError does not unwrap to inner error")
	}

	snapErr := &SnapshotError{Message: "xml damaged", Device: "R1", Err: inner}
	if !errors.Is(snapErr, ErrSnapshot) {
		t.Errorf("SnapshotError does not match ErrSnapshot")
	}
	if !errors.Is(snapErr, inner) {
		t.Errorf("SnapshotError does not unwrap to inner error")
	}

	gitErr := &GitOpError{Message: "git push failed", Command: "push", Err: inner}
	if !errors.Is(gitErr, ErrGitOp) {
		t.Errorf("GitOpError does not match ErrGitOp")
	}
	if !errors.Is(gitErr, inner) {
		t.Errorf("GitOpError does not unwrap to inner error")
	}

	uErr := usagef("bad args: %s", "x")
	if !errors.Is(uErr, ErrUsage) {
		t.Errorf("usageError does not match ErrUsage")
	}
	if !strings.Contains(uErr.Error(), "bad args: x") {
		t.Errorf("usageError.Error() = %q", uErr.Error())
	}
}

func TestLocalizeCLIErrors(t *testing.T) {
	// English verification
	i18n.Init("en")

	msg, hint := localizeCLIError(hooks.ErrNotRepository)
	if msg != "not a git repository" {
		t.Errorf("en msg = %q, want 'not a git repository'", msg)
	}
	if !strings.Contains(hint, "run 'git init'") {
		t.Errorf("en hint = %q, want to contain 'run git init'", hint)
	}

	msg, _ = localizeCLIError(codec.ErrTruncated)
	if !strings.Contains(msg, "truncated") {
		t.Errorf("en truncated msg = %q", msg)
	}

	msg, _ = localizeCLIError(codec.ErrInvalidXML)
	if !strings.Contains(msg, "invalid XML") {
		t.Errorf("en xml msg = %q", msg)
	}

	// Spanish verification
	i18n.Init("es")

	msg, hint = localizeCLIError(hooks.ErrNotRepository)
	if msg != "no es un repositorio git" {
		t.Errorf("es msg = %q, want 'no es un repositorio git'", msg)
	}
	if !strings.Contains(hint, "git init") {
		t.Errorf("es hint = %q, want to contain 'git init'", hint)
	}

	msg, _ = localizeCLIError(codec.ErrTruncated)
	if !strings.Contains(msg, "truncado") {
		t.Errorf("es truncated msg = %q", msg)
	}

	msg, _ = localizeCLIError(codec.ErrInvalidXML)
	if !strings.Contains(msg, "XML inválido o malformado") {
		t.Errorf("es xml msg = %q", msg)
	}
}

func TestCLIExecutionLocalizedErrors(t *testing.T) {
	dir := t.TempDir()

	// 1. Error in English via --lang en
	code, _, stderrEn := runCLI("--lang", "en", "init", dir)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderrEn, "not a git repository") {
		t.Errorf("stderrEn missing 'not a git repository': %q", stderrEn)
	}
	if !strings.Contains(stderrEn, "Hint:") {
		t.Errorf("stderrEn missing 'Hint:': %q", stderrEn)
	}

	// 2. Error in Spanish via --lang es
	code, _, stderrEs := runCLI("--lang", "es", "init", dir)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderrEs, "no es un repositorio git") {
		t.Errorf("stderrEs missing 'no es un repositorio git': %q", stderrEs)
	}
	if !strings.Contains(stderrEs, "Consejo:") {
		t.Errorf("stderrEs missing 'Consejo:': %q", stderrEs)
	}
}

func TestCLIExecutionNoColorErrorFallback(t *testing.T) {
	ui.SetForceNoColor(true)
	defer ui.ResetOverrides()

	dir := t.TempDir()
	code, _, stderr := runCLI("--lang", "en", "init", dir)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.HasPrefix(stderr, "pkt: not a git repository") {
		t.Errorf("stderr want prefix 'pkt: not a git repository', got: %q", stderr)
	}
}

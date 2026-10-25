package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/gitx"
	"github.com/MarkOne-dev/pkt-git/internal/hooks"
)

func TestInitCommand(t *testing.T) {
	dir := initCLIRepo(t)
	code, stdout, stderr := runCLI("init", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "[ok] initialized") || !strings.Contains(stdout, "pre-commit") {
		t.Fatalf("stdout = %q", stdout)
	}
	for _, name := range []string{"pre-commit", "post-merge", "post-checkout"} {
		info, err := os.Stat(filepath.Join(dir, hooks.HookDir, name))
		if err != nil {
			t.Fatalf("missing hook %s: %v", name, err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
			t.Fatalf("hook %s is not executable", name)
		}
	}
	repo := gitx.New(dir)
	if got, err := repo.ConfigGet(hooks.HooksPathKey); err != nil || got != hooks.HookDir {
		t.Fatalf("core.hooksPath = %q, %v", got, err)
	}
	bin, err := repo.ConfigGet(hooks.BinConfigKey)
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(bin) {
		t.Fatalf("pktgit.bin = %q, want an absolute path", bin)
	}
}

func TestInitCommandAutoExportsTopologies(t *testing.T) {
	dir := initCLIRepo(t)
	pktPath := filepath.Join(dir, "network.pkt")
	writeCLIFile(t, pktPath, encodeCLI(t, minimalCLIXML))

	code, stdout, stderr := runCLI("init", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "[ok] initialized") {
		t.Fatalf("stdout missing initialized: %q", stdout)
	}
	if !strings.Contains(stdout, "auto-detected network.pkt -> exported to snapshot/network") {
		t.Fatalf("stdout missing auto-detected export: %q", stdout)
	}
	snapXML := filepath.Join(dir, "snapshot", "network", "base.snapshot.xml")
	if _, err := os.Stat(snapXML); err != nil {
		t.Fatalf("missing auto-exported snapshot XML: %v", err)
	}
}

func TestInitCommandRejectsNonRepository(t *testing.T) {
	dir := t.TempDir()
	code, _, stderr := runCLI("init", dir)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "not a git repository") {
		t.Fatalf("stderr = %q", stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, hooks.HookDir)); !os.IsNotExist(err) {
		t.Fatalf("non-repository got hooks: %v", err)
	}
}

func TestInitCommandTooManyArguments(t *testing.T) {
	code, _, stderr := runCLI("init", "a", "b")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected at most 1 argument") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestHelpHidesHookCommands(t *testing.T) {
	code, stdout, stderr := runCLI("--help")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	for _, want := range []string{"init", "export", "build", "inspect", "edit", "sync", "resolve", "verify", "version"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("help does not list %q:\n%s", want, stdout)
		}
	}
	for _, hidden := range []string{"pre-commit", "post-merge", "post-checkout"} {
		if strings.Contains(stdout, hidden) {
			t.Fatalf("help advertises hidden command %q:\n%s", hidden, stdout)
		}
	}
}

func TestVersionOutsideRepository(t *testing.T) {
	t.Chdir(t.TempDir())
	code, stdout, stderr := runCLI("version")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if strings.TrimSpace(stdout) != versionString() {
		t.Fatalf("version output = %q, want %q", strings.TrimSpace(stdout), versionString())
	}
	code, flagOutput, stderr := runCLI("--version")
	if code != 0 {
		t.Fatalf("--version exit code = %d, stderr=%q", code, stderr)
	}
	if strings.TrimSpace(flagOutput) != versionString() {
		t.Fatalf("--version output = %q, want %q", strings.TrimSpace(flagOutput), versionString())
	}
}

func TestHookCommandArguments(t *testing.T) {
	dir := initCLIRepo(t)
	t.Chdir(dir)

	tests := [][]string{
		{"pre-commit", "extra"},
		{"post-merge", "HEAD"},
		{"post-checkout", "old", "new"},
		{"post-checkout", "old", "new", "2"},
	}
	for _, args := range tests {
		code, _, stderr := runCLI(args...)
		if code != 2 {
			t.Fatalf("%q exit code = %d, want 2 (stderr=%q)", args, code, stderr)
		}
	}

	code, stdout, stderr := runCLI("pre-commit")
	if code != 0 {
		t.Fatalf("empty pre-commit exit code = %d, stderr=%q", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("empty pre-commit output = %q", stdout)
	}
}

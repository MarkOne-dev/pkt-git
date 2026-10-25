package cli

import (
	"strings"
	"testing"
)

func TestIntegrationHelpListsPublicCommands(t *testing.T) {
	code, stdout, stderr := runCLI("--help")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	for _, command := range []string{
		"init", "export", "build", "inspect", "edit",
		"sync", "resolve", "verify", "skills", "version",
	} {
		if !strings.Contains(stdout, command) {
			t.Fatalf("help does not list %q:\n%s", command, stdout)
		}
	}
	for _, hidden := range []string{"pre-commit", "post-merge", "post-checkout"} {
		if strings.Contains(stdout, hidden) {
			t.Fatalf("help advertises hidden command %q:\n%s", hidden, stdout)
		}
	}
}

func TestIntegrationHelpSubcommand(t *testing.T) {
	code, stdout, stderr := runCLI("help")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Usage:") || !strings.Contains(stdout, "skills") {
		t.Fatalf("help output is incomplete:\n%s", stdout)
	}
}

func TestIntegrationMissingArguments(t *testing.T) {
	tests := [][]string{
		{"export"},
		{"build"},
		{"inspect"},
		{"sync"},
		{"verify"},
		{"init", "a", "b"},
		{"skills", "extra"},
		{"resolve", "--ours", "--theirs"},
	}
	for _, args := range tests {
		code, _, stderr := runCLI(args...)
		if code != 2 {
			t.Fatalf("%q exit code = %d, want 2 (stderr=%q)", args, code, stderr)
		}
		if strings.TrimSpace(stderr) == "" {
			t.Fatalf("%q produced no usage error", args)
		}
	}
}

func TestIntegrationUnknownCommand(t *testing.T) {
	code, _, stderr := runCLI("does-not-exist")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestIntegrationVersionOutsideRepository(t *testing.T) {
	t.Chdir(t.TempDir())
	code, stdout, stderr := runCLI("version")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("version output is empty")
	}
	code, flagOutput, stderr := runCLI("--version")
	if code != 0 {
		t.Fatalf("--version exit code = %d, stderr=%q", code, stderr)
	}
	if strings.TrimSpace(flagOutput) != strings.TrimSpace(stdout) {
		t.Fatalf("version outputs differ: %q vs %q", stdout, flagOutput)
	}
}

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelp(t *testing.T) {
	code, stdout, stderr := runCLI("--help")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Usage:") || !strings.Contains(stdout, "version") {
		t.Fatalf("help output is incomplete:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestRootWithoutArgumentsShowsHelpAndExitsTwo(t *testing.T) {
	code, stdout, stderr := runCLI()
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Fatalf("help output is missing:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestUnknownCommandIsUsageError(t *testing.T) {
	code, _, stderr := runCLI("frobnicate")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestVersion(t *testing.T) {
	code, stdout, stderr := runCLI("version")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("version output is empty")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestVersionRejectsExtraArguments(t *testing.T) {
	code, _, stderr := runCLI("version", "extra")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected 0 argument") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestUnknownFlagIsUsageError(t *testing.T) {
	code, _, stderr := runCLI("--frobnicate")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "unknown flag") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func runCLI(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := ExecuteArgs(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

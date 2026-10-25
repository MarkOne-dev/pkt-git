package cli

import (
	"strings"
	"testing"
)

func TestWelcomeCommand(t *testing.T) {
	code, stdout, stderr := runCLI("welcome")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "Cisco Packet Tracer ⇄ Git Version Control") {
		t.Fatalf("stdout missing banner: %q", stdout)
	}
	if !strings.Contains(stdout, "Quickstart Guide:") {
		t.Fatalf("stdout missing quickstart guide: %q", stdout)
	}
	if !strings.Contains(stdout, "pkt init") {
		t.Fatalf("stdout missing pkt init step: %q", stdout)
	}
}

func TestWelcomeCommandRejectsExtraArguments(t *testing.T) {
	code, _, stderr := runCLI("welcome", "extra")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected at most 0 argument") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestHelpListsWelcomeCommand(t *testing.T) {
	code, stdout, stderr := runCLI("--help")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "welcome") {
		t.Fatalf("help missing welcome command: %q", stdout)
	}
}

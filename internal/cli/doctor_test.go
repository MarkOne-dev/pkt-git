package cli

import (
	"strings"
	"testing"
)

func TestDoctorCommand(t *testing.T) {
	code, stdout, stderr := runCLI("doctor")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Git:") {
		t.Fatalf("expected Git check in doctor output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Git Identity:") {
		t.Fatalf("expected Git Identity check in doctor output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Binary:") {
		t.Fatalf("expected Binary check in doctor output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Cisco Packet Tracer:") {
		t.Fatalf("expected Cisco Packet Tracer check in doctor output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Go SDK:") {
		t.Fatalf("expected Go SDK check in doctor output:\n%s", stdout)
	}
}

func TestDoctorCommandRejectsExtraArguments(t *testing.T) {
	code, _, stderr := runCLI("doctor", "extra")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected 0 argument") {
		t.Fatalf("stderr = %q", stderr)
	}
}

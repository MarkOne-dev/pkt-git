package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestUIPlainFallback(t *testing.T) {
	ResetOverrides()
	defer ResetOverrides()

	var buf bytes.Buffer
	Success(&buf, "exported %s", "project.pkt")
	if !strings.Contains(buf.String(), "[ok] exported project.pkt") {
		t.Fatalf("expected [ok] prefix in non-TTY, got %q", buf.String())
	}

	buf.Reset()
	Warning(&buf, "stale binary")
	if !strings.Contains(buf.String(), "[warn] stale binary") {
		t.Fatalf("expected [warn] prefix in non-TTY, got %q", buf.String())
	}

	buf.Reset()
	Error(&buf, "cannot read file")
	if !strings.Contains(buf.String(), "pkt: cannot read file") {
		t.Fatalf("expected pkt: prefix in non-TTY, got %q", buf.String())
	}

	buf.Reset()
	Info(&buf, "hook installed")
	if !strings.Contains(buf.String(), "[info] hook installed") {
		t.Fatalf("expected [info] prefix in non-TTY, got %q", buf.String())
	}
}

func TestUIForcedTTY(t *testing.T) {
	ResetOverrides()
	defer ResetOverrides()

	SetForceTTY(true)
	SetForceNoColor(false)
	var buf bytes.Buffer

	Success(&buf, "operation completed")
	if !strings.Contains(buf.String(), "✔") {
		t.Fatalf("expected ✔ symbol in TTY, got %q", buf.String())
	}

	buf.Reset()
	Warning(&buf, "caution message")
	if !strings.Contains(buf.String(), "⚠") {
		t.Fatalf("expected ⚠ symbol in TTY, got %q", buf.String())
	}

	buf.Reset()
	Error(&buf, "failure message")
	if !strings.Contains(buf.String(), "✖") {
		t.Fatalf("expected ✖ symbol in TTY, got %q", buf.String())
	}

	buf.Reset()
	Info(&buf, "notice message")
	if !strings.Contains(buf.String(), "ℹ") {
		t.Fatalf("expected ℹ symbol in TTY, got %q", buf.String())
	}
}

func TestUINoColorOverride(t *testing.T) {
	ResetOverrides()
	defer ResetOverrides()

	SetForceTTY(true)
	SetForceNoColor(true)

	var buf bytes.Buffer
	Success(&buf, "plain operation")
	if strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("did not expect ANSI escape codes when NoColor is forced: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "[ok] plain operation") {
		t.Fatalf("expected plain [ok] prefix, got %q", buf.String())
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{1024, "1.0 kB"},
		{1048576, "1.0 MB"},
		{5894851, "5.9 MB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestDeviceBadge(t *testing.T) {
	ResetOverrides()
	defer ResetOverrides()

	var buf bytes.Buffer
	// In non-TTY, badge returns plain name
	if got := DeviceBadge(&buf, "Router"); got != "Router" {
		t.Errorf("expected plain badge 'Router', got %q", got)
	}

	// In TTY, badge contains ANSI color
	SetForceTTY(true)
	SetForceNoColor(false)
	if got := DeviceBadge(&buf, "Router"); !strings.Contains(got, "Router") {
		t.Errorf("expected badge containing 'Router', got %q", got)
	}
}

func TestRenderTable(t *testing.T) {
	ResetOverrides()
	defer ResetOverrides()

	var buf bytes.Buffer
	headers := []string{"NAME", "TYPE", "CONFIG"}
	rows := [][]string{
		{"R1", "Router", "120"},
		{"SW1", "Switch", "80"},
	}

	tbl := RenderTable(&buf, headers, rows)
	for _, term := range []string{"NAME", "TYPE", "CONFIG", "R1", "Router", "120", "SW1", "Switch", "80"} {
		if !strings.Contains(tbl, term) {
			t.Errorf("rendered table missing %q:\n%s", term, tbl)
		}
	}
}

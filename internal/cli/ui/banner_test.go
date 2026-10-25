package ui

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/i18n"
)

func TestBannerPlain(t *testing.T) {
	i18n.Init("en")
	SetForceTTY(false)
	SetForceNoColor(true)
	defer ResetOverrides()

	var buf bytes.Buffer
	Banner(&buf)
	got := buf.String()

	if !strings.Contains(got, "▄▄▄▄  ▄  ▄ ▄▄▄▄▄") {
		t.Fatalf("Banner missing ASCII art: %q", got)
	}
	if !strings.Contains(got, "Cisco Packet Tracer ⇄ Git Version Control") {
		t.Fatalf("Banner missing subtitle: %q", got)
	}
	if !strings.Contains(got, runtime.GOOS) {
		t.Fatalf("Banner missing OS in metadata line: %q", got)
	}
	if !strings.Contains(got, runtime.Version()) {
		t.Fatalf("Banner missing Go version in metadata line: %q", got)
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("Banner has ANSI escapes when NO_COLOR is forced: %q", got)
	}
}

func TestBannerColor(t *testing.T) {
	i18n.Init("en")
	SetForceTTY(true)
	SetForceNoColor(false)
	defer ResetOverrides()

	var buf bytes.Buffer
	Banner(&buf)
	got := buf.String()

	if !strings.Contains(got, "Cisco Packet Tracer ⇄ Git Version Control") {
		t.Fatalf("Banner missing subtitle: %q", got)
	}
	if !strings.Contains(got, runtime.GOOS) {
		t.Fatalf("Banner missing OS metadata in colored output: %q", got)
	}
}

func TestBannerString(t *testing.T) {
	plain := BannerString(false)
	if !strings.Contains(plain, "Cisco Packet Tracer") || strings.Contains(plain, "\x1b[") {
		t.Fatalf("BannerString(false) invalid: %q", plain)
	}
	if !strings.Contains(plain, runtime.GOOS) {
		t.Fatalf("BannerString(false) missing OS: %q", plain)
	}

	colored := BannerString(true)
	if !strings.Contains(colored, "Cisco Packet Tracer") {
		t.Fatalf("BannerString(true) missing text: %q", colored)
	}
}

func TestBannerSpanish(t *testing.T) {
	i18n.Init("es")
	defer i18n.Init("en")

	var buf bytes.Buffer
	Banner(&buf)
	got := buf.String()

	if !strings.Contains(got, "Cisco Packet Tracer ⇄ Control de versiones Git") {
		t.Fatalf("Banner missing Spanish subtitle: %q", got)
	}
}

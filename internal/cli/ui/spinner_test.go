package ui

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func TestSpinnerInteractiveSuccess(t *testing.T) {
	var buf safeBuffer
	s := NewSpinner(&buf, "Testing operation")
	s.SetInteractive(true)
	s.SetInterval(10 * time.Millisecond)

	s.Start()
	time.Sleep(35 * time.Millisecond)
	s.UpdateMessage("Still running")
	time.Sleep(20 * time.Millisecond)
	s.Success("Operation finished")

	got := buf.String()
	if !strings.Contains(got, "Operation finished") {
		t.Fatalf("stdout missing success message: %q", got)
	}
	if !strings.Contains(got, "✔") {
		t.Fatalf("stdout missing checkmark symbol: %q", got)
	}
}

func TestSpinnerInteractiveFail(t *testing.T) {
	var buf safeBuffer
	s := NewSpinner(&buf, "Testing fail")
	s.SetInteractive(true)
	s.SetInterval(10 * time.Millisecond)

	s.Start()
	time.Sleep(25 * time.Millisecond)
	s.Fail("Operation failed")

	got := buf.String()
	if !strings.Contains(got, "Operation failed") {
		t.Fatalf("stdout missing fail message: %q", got)
	}
	if !strings.Contains(got, "✖") {
		t.Fatalf("stdout missing cross symbol: %q", got)
	}
}

func TestSpinnerNonInteractiveFallback(t *testing.T) {
	var buf safeBuffer
	s := NewSpinner(&buf, "Static task")
	s.SetInteractive(false)

	s.Start()
	s.Success("Static task complete")

	got := buf.String()
	if !strings.Contains(got, "[..] Static task...") {
		t.Fatalf("missing static start line: %q", got)
	}
	if !strings.Contains(got, "[ok] Static task complete") {
		t.Fatalf("missing static ok line: %q", got)
	}
	if strings.Contains(got, "\r") || strings.Contains(got, "\x1b[") {
		t.Fatalf("non-interactive emitted ANSI control characters: %q", got)
	}
}

func TestSpinnerStopIdempotent(t *testing.T) {
	var buf safeBuffer
	s := NewSpinner(&buf, "Stop test")
	s.SetInteractive(true)
	s.Start()
	s.Stop()
	// Second stop should not panic or hang
	s.Stop()
}

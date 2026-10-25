package ptdetect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanWindowsProgramDirs(t *testing.T) {
	tmpDir := t.TempDir()
	ptDir := filepath.Join(tmpDir, "Cisco Packet Tracer 8.2.2", "bin")
	if err := os.MkdirAll(ptDir, 0o755); err != nil {
		t.Fatalf("failed to create dummy directory: %v", err)
	}

	exePath := filepath.Join(ptDir, "PacketTracer.exe")
	if err := os.WriteFile(exePath, []byte("dummy exe"), 0o755); err != nil {
		t.Fatalf("failed to create dummy exe: %v", err)
	}

	inst := scanWindowsProgramDirs([]string{tmpDir})
	if inst == nil {
		t.Fatal("expected installation found in Windows program dir scan, got nil")
	}

	if inst.Path != exePath {
		t.Errorf("Path = %q, want %q", inst.Path, exePath)
	}
	if inst.Version != "8.2.2" {
		t.Errorf("Version = %q, want 8.2.2", inst.Version)
	}
	if inst.Method != "Versioned Program Files Scan" {
		t.Errorf("Method = %q, want Versioned Program Files Scan", inst.Method)
	}
}

func TestScanWindowsProgramDirsNoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	inst := scanWindowsProgramDirs([]string{tmpDir})
	if inst != nil {
		t.Errorf("expected nil for empty directory, got %+v", inst)
	}
}

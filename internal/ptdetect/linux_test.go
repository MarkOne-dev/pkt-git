package ptdetect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDesktopFile(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "packettracer.AppImage")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("failed to create dummy binary: %v", err)
	}

	desktopContent := `[Desktop Entry]
Name=Cisco Packet Tracer 8.2.2
Comment=Cisco Packet Tracer Network Simulator
Exec=` + binPath + ` %f
Icon=/usr/share/pixmaps/packettracer.png
Terminal=false
Type=Application
Categories=Network;Education;
`
	desktopPath := filepath.Join(tmpDir, "packettracer.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0o644); err != nil {
		t.Fatalf("failed to create dummy desktop file: %v", err)
	}

	inst := ParseDesktopFile(desktopPath)
	if inst == nil {
		t.Fatal("expected non-nil Installation from ParseDesktopFile")
	}

	if inst.Path != binPath {
		t.Errorf("Path = %q, want %q", inst.Path, binPath)
	}
	if inst.Version != "8.2.2" {
		t.Errorf("Version = %q, want 8.2.2", inst.Version)
	}
	if inst.Method != "XDG Desktop Entry (packettracer.desktop)" {
		t.Errorf("Method = %q, want XDG Desktop Entry (packettracer.desktop)", inst.Method)
	}
}

func TestParseDesktopFileRejectsMissingBinary(t *testing.T) {
	tmpDir := t.TempDir()
	desktopContent := `[Desktop Entry]
Name=Cisco Packet Tracer
Exec=/nonexistent/path/PacketTracer %f
Type=Application
`
	desktopPath := filepath.Join(tmpDir, "pt.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0o644); err != nil {
		t.Fatalf("failed to create dummy desktop file: %v", err)
	}

	inst := ParseDesktopFile(desktopPath)
	if inst != nil {
		t.Errorf("expected nil for nonexistent binary, got %+v", inst)
	}
}

func TestScanDesktopDirs(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "PacketTracer")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("failed to create dummy binary: %v", err)
	}

	desktopContent := `[Desktop Entry]
Name=Cisco Packet Tracer
Exec=` + binPath + `
Type=Application
`
	desktopPath := filepath.Join(tmpDir, "cisco-pt.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0o644); err != nil {
		t.Fatalf("failed to create desktop file: %v", err)
	}

	inst := scanDesktopDirs([]string{tmpDir})
	if inst == nil {
		t.Fatal("expected non-nil Installation from scanDesktopDirs")
	}
	if inst.Path != binPath {
		t.Errorf("Path = %q, want %q", inst.Path, binPath)
	}
}

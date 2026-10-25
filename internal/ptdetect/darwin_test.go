package ptdetect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectAppBundle(t *testing.T) {
	tmpDir := t.TempDir()
	bundleDir := filepath.Join(tmpDir, "Cisco Packet Tracer.app")
	contentsDir := filepath.Join(bundleDir, "Contents")
	macosDir := filepath.Join(contentsDir, "MacOS")
	if err := os.MkdirAll(macosDir, 0o755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	execPath := filepath.Join(macosDir, "PacketTracer")
	if err := os.WriteFile(execPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("failed to write dummy exec: %v", err)
	}

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleShortVersionString</key>
	<string>8.2.1</string>
</dict>
</plist>`
	if err := os.WriteFile(filepath.Join(contentsDir, "Info.plist"), []byte(plistContent), 0o644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	inst := inspectAppBundle(bundleDir)
	if inst == nil {
		t.Fatal("expected non-nil installation from inspectAppBundle")
	}

	if inst.Path != execPath {
		t.Errorf("Path = %q, want %q", inst.Path, execPath)
	}
	if inst.Version != "8.2.1" {
		t.Errorf("Version = %q, want 8.2.1", inst.Version)
	}
	if inst.Method != "macOS Application Bundle" {
		t.Errorf("Method = %q, want macOS Application Bundle", inst.Method)
	}
}

func TestMdfindSpotlight(t *testing.T) {
	oldMdfind := MdfindFunc
	defer func() { MdfindFunc = oldMdfind }()

	tmpDir := t.TempDir()
	bundleDir := filepath.Join(tmpDir, "Cisco Packet Tracer 8.0.app")
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatalf("failed to create bundle: %v", err)
	}

	MdfindFunc = func(query string) ([]string, error) {
		return []string{bundleDir}, nil
	}

	inst := detectDarwin()
	if inst == nil {
		t.Fatal("expected non-nil installation from detectDarwin")
	}
	if inst.Path != bundleDir {
		t.Errorf("Path = %q, want %q", inst.Path, bundleDir)
	}
	if inst.Method != "macOS Spotlight (mdfind)" {
		t.Errorf("Method = %q, want macOS Spotlight (mdfind)", inst.Method)
	}
}

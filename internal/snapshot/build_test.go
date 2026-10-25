package snapshot

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
)

func TestBuildFromExportedSnapshot(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "snapshot")
	if _, err := Export(specimenPath(t), outputDir); err != nil {
		t.Fatalf("export: %v", err)
	}
	xmlPath := filepath.Join(outputDir, "base.snapshot.xml")
	before := mustRead(t, xmlPath)

	outputPath := filepath.Join(t.TempDir(), "topologia", "rebuilt.pkt")
	result, err := Build(xmlPath, outputPath)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	data := mustRead(t, outputPath)
	if result.Size != len(data) {
		t.Fatalf("Size = %d, want %d", result.Size, len(data))
	}

	decoded, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode rebuilt binary: %v", err)
	}
	if !bytes.Equal(decoded, before) {
		t.Fatal("rebuilt binary does not decode to the source XML")
	}

	after := mustRead(t, xmlPath)
	if !bytes.Equal(before, after) {
		t.Fatal("Build modified the source snapshot")
	}
}

func TestBuildInvalidXMLLeavesBinaryUntouched(t *testing.T) {
	xmlPath := filepath.Join(t.TempDir(), "base.snapshot.xml")
	if err := os.WriteFile(xmlPath, []byte("<PACKETTRACER5>"), 0o644); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(t.TempDir(), "existing.pkt")
	sentinel := []byte("previous binary")
	if err := os.WriteFile(outputPath, sentinel, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(xmlPath, outputPath); err == nil {
		t.Fatal("Build accepted malformed XML")
	}
	if got := mustRead(t, outputPath); !bytes.Equal(got, sentinel) {
		t.Fatalf("existing binary changed: %q", got)
	}
}

func TestBuildInvalidXMLDoesNotCreateOutput(t *testing.T) {
	xmlPath := filepath.Join(t.TempDir(), "base.snapshot.xml")
	if err := os.WriteFile(xmlPath, []byte("<PACKETTRACER5>"), 0o644); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(t.TempDir(), "new.pkt")

	if _, err := Build(xmlPath, outputPath); err == nil {
		t.Fatal("Build accepted malformed XML")
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output was created on failure: %v", err)
	}
}

func TestBuildCreatesParentDirectory(t *testing.T) {
	xmlPath := filepath.Join(t.TempDir(), "base.snapshot.xml")
	xmlData := []byte("<PACKETTRACER5><NETWORK><DEVICES/><LINKS/></NETWORK></PACKETTRACER5>")
	if err := os.WriteFile(xmlPath, xmlData, 0o644); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(t.TempDir(), "nested", "out.pkt")

	if _, err := Build(xmlPath, outputPath); err != nil {
		t.Fatalf("Build: %v", err)
	}
	decoded, err := codec.Decode(mustRead(t, outputPath))
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if !bytes.Equal(decoded, xmlData) {
		t.Fatal("output does not decode to the source XML")
	}
}

package editor

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/xmledit"
)

func TestSetInterfaceAppliesAndSyncsDerivedFiles(t *testing.T) {
	xmlPath := writeSnapshot(t, setifaceXML)

	if err := SetInterface(xmlPath, "R1", "Fa0/1", []string{
		" description New",
		" switchport access vlan 20",
	}); err != nil {
		t.Fatalf("SetInterface: %v", err)
	}

	data := mustReadFile(t, xmlPath)
	index, err := xmledit.BuildIndex(data)
	if err != nil {
		t.Fatalf("updated XML is invalid: %v", err)
	}
	device, err := index.FindDevice("R1")
	if err != nil {
		t.Fatal(err)
	}
	block, err := device.InterfaceBlock("Fa0/1")
	if err != nil {
		t.Fatal(err)
	}
	if len(block.Lines) != 2 || block.Lines[0].Text != " description New" {
		t.Fatalf("block lines = %+v", block.Lines)
	}

	cfg := mustReadFile(t, filepath.Join(filepath.Dir(xmlPath), "configs", "R1.cfg"))
	if !bytes.Contains(cfg, []byte("description New")) {
		t.Fatalf("config was not refreshed:\n%s", cfg)
	}
	if bytes.Contains(cfg, []byte("description Old")) {
		t.Fatalf("config still contains the old line:\n%s", cfg)
	}

	metadata := mustReadFile(t, filepath.Join(filepath.Dir(xmlPath), "metadata", "topology.json"))
	if !json.Valid(metadata) {
		t.Fatal("metadata/topology.json is invalid")
	}
}

func TestChangeLineAppliesAndSyncsDerivedFiles(t *testing.T) {
	xmlPath := writeSnapshot(t, chglineXML)

	if err := ChangeLine(xmlPath, "R1", "Fa0/1", " description Old", " description New"); err != nil {
		t.Fatalf("ChangeLine: %v", err)
	}

	cfg := mustReadFile(t, filepath.Join(filepath.Dir(xmlPath), "configs", "R1.cfg"))
	if !bytes.Contains(cfg, []byte("description New")) {
		t.Fatalf("config was not refreshed:\n%s", cfg)
	}
	if bytes.Contains(cfg, []byte("description Old")) {
		t.Fatalf("config still contains the old line:\n%s", cfg)
	}
}

func TestInsertAfterAppliesAndSyncsDerivedFiles(t *testing.T) {
	xmlPath := writeSnapshot(t, insertXML)

	if err := InsertAfter(xmlPath, "R1", "hostname", []string{"ntp server 1.1.1.1"}); err != nil {
		t.Fatalf("InsertAfter: %v", err)
	}

	cfg := mustReadFile(t, filepath.Join(filepath.Dir(xmlPath), "configs", "R1.cfg"))
	if !bytes.Contains(cfg, []byte("hostname R1\nntp server 1.1.1.1\n")) {
		t.Fatalf("config does not contain the inserted line:\n%s", cfg)
	}
}

func TestApplyInvalidResultLeavesOriginalBytes(t *testing.T) {
	xmlPath := writeSnapshot(t, setifaceXML)
	before := mustReadFile(t, xmlPath)

	err := Apply(xmlPath, func([]byte) ([]byte, error) {
		return []byte("<PACKETTRACER5>"), nil
	})
	if err == nil {
		t.Fatal("Apply accepted invalid XML")
	}

	after := mustReadFile(t, xmlPath)
	if !bytes.Equal(before, after) {
		t.Fatal("invalid edit changed the snapshot bytes")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(xmlPath), "configs")); !os.IsNotExist(err) {
		t.Fatalf("invalid edit created derived files: %v", err)
	}
}

func TestApplyTransformErrorLeavesOriginalBytes(t *testing.T) {
	xmlPath := writeSnapshot(t, setifaceXML)
	before := mustReadFile(t, xmlPath)

	if err := SetInterface(xmlPath, "R1", "Fa0/9", nil); err == nil {
		t.Fatal("SetInterface accepted a missing interface")
	}

	after := mustReadFile(t, xmlPath)
	if !bytes.Equal(before, after) {
		t.Fatal("failed transform changed the snapshot bytes")
	}
}

func TestApplyReportsSyncFailure(t *testing.T) {
	xmlPath := writeSnapshot(t, setifaceXML)
	configPath := filepath.Join(filepath.Dir(xmlPath), "configs")
	if err := os.WriteFile(configPath, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := SetInterface(xmlPath, "R1", "Fa0/1", []string{" description New"})
	if err == nil {
		t.Fatal("SetInterface succeeded despite a sync failure")
	}
	if !strings.Contains(err.Error(), "derived files may be stale") {
		t.Fatalf("error does not report stale derived files: %v", err)
	}

	updated := mustReadFile(t, xmlPath)
	if !bytes.Contains(updated, []byte("description New")) {
		t.Fatal("snapshot write did not happen before the sync failure")
	}
}

func writeSnapshot(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "base.snapshot.xml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

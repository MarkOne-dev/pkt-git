package snapshot

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/topology"
)

func TestSyncRefreshesDerivedFilesAndPrunesStaleConfigs(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "snapshot")
	if _, err := Export(specimenPath(t), outputDir); err != nil {
		t.Fatalf("export: %v", err)
	}
	xmlPath := filepath.Join(outputDir, "base.snapshot.xml")
	beforeXML := mustRead(t, xmlPath)

	stale := filepath.Join(outputDir, "configs", "STALE.cfg")
	if err := os.WriteFile(stale, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Sync(outputDir)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if result.DeviceCount != 107 || result.LinkCount != 106 || result.ConfigCount != 34 {
		t.Fatalf("counts = %+v", result)
	}
	if len(result.Removed) != 1 || result.Removed[0] != "STALE.cfg" {
		t.Fatalf("removed = %v, want [STALE.cfg]", result.Removed)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale config was not removed: %v", err)
	}

	afterXML := mustRead(t, xmlPath)
	if !bytes.Equal(beforeXML, afterXML) {
		t.Fatal("Sync modified base.snapshot.xml")
	}

	metadata := mustRead(t, filepath.Join(outputDir, "metadata", "topology.json"))
	var catalog topology.Catalog
	if err := json.Unmarshal(metadata, &catalog); err != nil {
		t.Fatal(err)
	}
	if catalog.TotalDevices != 107 || catalog.TotalLinks != 106 {
		t.Fatalf("metadata totals = %d/%d", catalog.TotalDevices, catalog.TotalLinks)
	}
	if catalog.SourceFile != "Caso1_Grupo09-2.pkt" {
		t.Fatalf("source_file = %q", catalog.SourceFile)
	}
}

func TestPruneConfigsHandlesCaseOnlyNames(t *testing.T) {
	dir := t.TempDir()
	lower := filepath.Join(dir, "r1.cfg")
	upper := filepath.Join(dir, "R1.cfg")
	if err := os.WriteFile(lower, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(upper, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lowerInfo, err := os.Stat(lower)
	if err != nil {
		t.Fatal(err)
	}
	upperInfo, err := os.Stat(upper)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(lowerInfo, upperInfo) {
		t.Skip("filesystem is case-insensitive")
	}

	removed, err := pruneConfigs(dir, map[string][]byte{"r1.cfg": []byte("new\n")})
	if err != nil {
		t.Fatalf("pruneConfigs: %v", err)
	}
	if len(removed) != 1 || removed[0] != "R1.cfg" {
		t.Fatalf("removed = %q, want [R1.cfg]", removed)
	}
	if _, err := os.Stat(upper); !os.IsNotExist(err) {
		t.Fatalf("case-only stale config still exists: %v", err)
	}
	if _, err := os.Stat(lower); err != nil {
		t.Fatalf("expected config was removed: %v", err)
	}
}

func TestSyncAcceptsXMLPath(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "snapshot")
	if _, err := Export(specimenPath(t), outputDir); err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := Sync(filepath.Join(outputDir, "base.snapshot.xml")); err != nil {
		t.Fatalf("Sync with XML path: %v", err)
	}
}

func TestSyncCreatesMissingDirectories(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "snapshot")
	if _, err := Export(specimenPath(t), outputDir); err != nil {
		t.Fatalf("export: %v", err)
	}
	xmlData := mustRead(t, filepath.Join(outputDir, "base.snapshot.xml"))

	projectDir := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "base.snapshot.xml"), xmlData, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Sync(projectDir); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	for _, path := range []string{
		filepath.Join(projectDir, "configs"),
		filepath.Join(projectDir, "metadata", "topology.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
}

func TestSyncInvalidXMLPreservesDerivedFiles(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "snapshot")
	if _, err := Export(specimenPath(t), outputDir); err != nil {
		t.Fatalf("export: %v", err)
	}
	before := captureTree(t, outputDir)

	xmlPath := filepath.Join(outputDir, "base.snapshot.xml")
	if err := os.WriteFile(xmlPath, []byte("<PACKETTRACER5>"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The corrupt XML itself is the only expected change.
	before["base.snapshot.xml"] = []byte("<PACKETTRACER5>")

	if _, err := Sync(outputDir); err == nil {
		t.Fatal("Sync accepted malformed XML")
	}
	after := captureTree(t, outputDir)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("failed sync modified derived files")
	}
}

func TestSyncPreservesExistingSourceFile(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "custom-directory")
	if _, err := Export(specimenPath(t), outputDir); err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := Sync(outputDir); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	metadata := mustRead(t, filepath.Join(outputDir, "metadata", "topology.json"))
	if !strings.Contains(string(metadata), `"source_file": "Caso1_Grupo09-2.pkt"`) {
		t.Fatalf("source_file was not preserved:\n%s", metadata)
	}
}

package snapshot

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
	"github.com/MarkOne-dev/pkt-git/internal/topology"
)

func TestDefaultSnapshotDir(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"network.pkt", filepath.Join("snapshot", "network")},
		{"topologia/network.pkt", filepath.Join("snapshot", "network")},
		{"lab_v1.pka", filepath.Join("snapshot", "lab_v1")},
		{"/abs/path/to/my_topology.PKT", filepath.Join("snapshot", "my_topology")},
	}
	for _, tt := range tests {
		got := DefaultSnapshotDir(tt.input)
		if got != tt.want {
			t.Errorf("DefaultSnapshotDir(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFindTopologies(t *testing.T) {
	dir := t.TempDir()

	// Create root topology
	if err := os.WriteFile(filepath.Join(dir, "root.pkt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create subfolder topology
	topoDir := filepath.Join(dir, "topologia")
	if err := os.MkdirAll(topoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(topoDir, "branch.pka"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create ignored files/dirs
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "ignored.pkt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	snapDir := filepath.Join(dir, "snapshot")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapDir, "ignored.pkt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "temp.rebuilt.pkt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := FindTopologies(dir)
	if err != nil {
		t.Fatalf("FindTopologies: %v", err)
	}
	want := []string{"root.pkt", filepath.Join("topologia", "branch.pka")}
	if !reflect.DeepEqual(found, want) {
		t.Fatalf("FindTopologies(%q) = %v, want %v", dir, found, want)
	}
}

func TestExportSpecimen(t *testing.T) {
	pktPath := specimenPath(t)
	outputDir := filepath.Join(t.TempDir(), "snapshot")

	result, err := Export(pktPath, outputDir)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if result.DeviceCount != 107 || result.LinkCount != 106 {
		t.Fatalf("counts = %d devices, %d links", result.DeviceCount, result.LinkCount)
	}
	if result.ConfigCount != 34 {
		t.Fatalf("config count = %d, want 34", result.ConfigCount)
	}
	if result.XMLSize != 5894851 {
		t.Fatalf("XML size = %d, want 5894851", result.XMLSize)
	}

	xmlPath := filepath.Join(outputDir, "base.snapshot.xml")
	gotXML, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatalf("read snapshot XML: %v", err)
	}
	wantXML, err := codec.Decode(mustRead(t, pktPath))
	if err != nil {
		t.Fatalf("decode specimen: %v", err)
	}
	if !bytes.Equal(gotXML, wantXML) {
		t.Fatal("exported XML differs from the decoded XML")
	}

	entries, err := os.ReadDir(filepath.Join(outputDir, "configs"))
	if err != nil {
		t.Fatalf("read configs: %v", err)
	}
	cfgCount := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".cfg") {
			cfgCount++
		}
	}
	if cfgCount != 34 {
		t.Fatalf("config files = %d, want 34", cfgCount)
	}

	cfg, err := os.ReadFile(filepath.Join(outputDir, "configs", "SW_ACC_ADMIN_PUNO.cfg"))
	if err != nil {
		t.Fatalf("read device config: %v", err)
	}
	if !bytes.HasSuffix(cfg, []byte("\n")) {
		t.Fatal("device config does not end with a newline")
	}
	if !bytes.Contains(cfg, []byte("hostname SW_ACC_ADMIN_PUNO")) {
		t.Fatal("device config does not contain its hostname")
	}

	metadata, err := os.ReadFile(filepath.Join(outputDir, "metadata", "topology.json"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	var catalog topology.Catalog
	if err := json.Unmarshal(metadata, &catalog); err != nil {
		t.Fatalf("parse metadata: %v", err)
	}
	if catalog.SourceFile != "Caso1_Grupo09-2.pkt" {
		t.Fatalf("source_file = %q", catalog.SourceFile)
	}
	if catalog.TotalDevices != 107 || catalog.TotalLinks != 106 {
		t.Fatalf("metadata totals = %d/%d", catalog.TotalDevices, catalog.TotalLinks)
	}
}

func TestExportPrunesStaleConfigs(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "snapshot")
	xmlData := []byte(`<PACKETTRACER5><NETWORK><DEVICES><DEVICE><ENGINE><TYPE model="2960">Switch</TYPE><NAME translate="true">R1</NAME><RUNNINGCONFIG><LINE>hostname R1</LINE><LINE>!</LINE></RUNNINGCONFIG></ENGINE></DEVICE></DEVICES><LINKS/></NETWORK></PACKETTRACER5>`)
	raw, err := codec.Encode(xmlData)
	if err != nil {
		t.Fatalf("encode test binary: %v", err)
	}
	pktPath := filepath.Join(dir, "project.pkt")
	if err := os.WriteFile(pktPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Export(pktPath, outputDir)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if !slices.Contains(result.Files, "base.snapshot.xml") || !slices.Contains(result.Files, "configs/R1.cfg") {
		t.Fatalf("Files = %q", result.Files)
	}

	stale := filepath.Join(outputDir, "configs", "OLD.cfg")
	if err := os.WriteFile(stale, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err = Export(pktPath, outputDir)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale config still exists: %v", err)
	}
	if !slices.Contains(result.Removed, "OLD.cfg") || !slices.Contains(result.Files, "configs/OLD.cfg") {
		t.Fatalf("result = %+v", result)
	}
}

func TestExportFailedDecodePreservesSnapshot(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "snapshot")
	if _, err := Export(specimenPath(t), outputDir); err != nil {
		t.Fatalf("initial export: %v", err)
	}
	before := captureTree(t, outputDir)

	badPkt := filepath.Join(t.TempDir(), "bad.pkt")
	raw := mustRead(t, specimenPath(t))
	if err := os.WriteFile(badPkt, raw[:len(raw)-1], 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Export(badPkt, outputDir); err == nil {
		t.Fatal("Export accepted a truncated binary")
	}

	after := captureTree(t, outputDir)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("failed decode modified the previous snapshot")
	}
}

func TestExportFailedParsePreservesSnapshot(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "snapshot")
	if _, err := Export(specimenPath(t), outputDir); err != nil {
		t.Fatalf("initial export: %v", err)
	}
	before := captureTree(t, outputDir)

	badPkt := filepath.Join(t.TempDir(), "bad.pkt")
	encoded, err := codec.Encode([]byte(`<PACKETTRACER5/>`))
	if err != nil {
		t.Fatalf("encode invalid topology: %v", err)
	}
	if err := os.WriteFile(badPkt, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Export(badPkt, outputDir); err == nil {
		t.Fatal("Export accepted a binary without NETWORK")
	}

	after := captureTree(t, outputDir)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("failed parse modified the previous snapshot")
	}
}

func TestExportDoesNotCreateOutputOnFailure(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "snapshot")
	badPkt := filepath.Join(t.TempDir(), "bad.pkt")
	if err := os.WriteFile(badPkt, []byte("not a container"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Export(badPkt, outputDir); err == nil {
		t.Fatal("Export accepted an invalid binary")
	}
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		t.Fatalf("output directory was created on failure: %v", err)
	}
}

func specimenPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "specimens", "Caso1_Grupo09-2.pkt")
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func captureTree(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			files[rel+"/"] = nil
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = data
		return nil
	})
	if err != nil {
		t.Fatalf("capture tree: %v", err)
	}
	return files
}

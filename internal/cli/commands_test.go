package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
)

const minimalCLIXML = `<PACKETTRACER5><NETWORK><DEVICES><DEVICE><ENGINE><TYPE model="2960">Switch</TYPE><NAME translate="true">R1</NAME><RUNNINGCONFIG><LINE>hostname R1</LINE><LINE>!</LINE></RUNNINGCONFIG></ENGINE></DEVICE></DEVICES></NETWORK></PACKETTRACER5>`

func TestExportCommand(t *testing.T) {
	dir := t.TempDir()
	pktPath := filepath.Join(dir, "project with space.pkt")
	outputDir := filepath.Join(dir, "snapshot with space")
	writeCLIFile(t, pktPath, encodeCLI(t, minimalCLIXML))

	code, stdout, stderr := runCLI("export", pktPath, outputDir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "[ok] exported") {
		t.Fatalf("stdout = %q", stdout)
	}
	for _, path := range []string{
		filepath.Join(outputDir, "base.snapshot.xml"),
		filepath.Join(outputDir, "configs", "R1.cfg"),
		filepath.Join(outputDir, "metadata", "topology.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
}

func TestExportCommandUsage(t *testing.T) {
	code, _, stderr := runCLI("export")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected 2 argument") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestExportCommandFailure(t *testing.T) {
	dir := t.TempDir()
	pktPath := filepath.Join(dir, "broken.pkt")
	outputDir := filepath.Join(dir, "out")
	writeCLIFile(t, pktPath, []byte("not a container"))

	code, _, stderr := runCLI("export", pktPath, outputDir)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "broken.pkt") {
		t.Fatalf("stderr = %q", stderr)
	}
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		t.Fatalf("failed export created output: %v", err)
	}
}

func TestBuildCommand(t *testing.T) {
	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "base.snapshot.xml")
	outputPath := filepath.Join(dir, "project.pkt")
	writeCLIFile(t, xmlPath, []byte(minimalCLIXML))

	code, stdout, stderr := runCLI("build", xmlPath, outputPath)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "[ok] built") {
		t.Fatalf("stdout = %q", stdout)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := codec.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != minimalCLIXML {
		t.Fatalf("decoded XML = %q", decoded)
	}
}

func TestBuildCommandFailure(t *testing.T) {
	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "base.snapshot.xml")
	outputPath := filepath.Join(dir, "project.pkt")
	writeCLIFile(t, xmlPath, []byte("<PACKETTRACER5>"))

	code, _, stderr := runCLI("build", xmlPath, outputPath)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "invalid XML") {
		t.Fatalf("stderr = %q", stderr)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("failed build created output: %v", err)
	}
}

func TestInspectCommand(t *testing.T) {
	dir := t.TempDir()
	pktPath := filepath.Join(dir, "project.pkt")
	writeCLIFile(t, pktPath, encodeCLI(t, minimalCLIXML))

	code, stdout, stderr := runCLI("inspect", pktPath)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	for _, want := range []string{"1 device", "R1", "Switch", "2960"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout)
		}
	}
}

func TestInspectCommandUsage(t *testing.T) {
	code, _, stderr := runCLI("inspect")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected 1 argument") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestInspectCommandFilter(t *testing.T) {
	dir := t.TempDir()
	pktPath := filepath.Join(dir, "project.pkt")
	writeCLIFile(t, pktPath, encodeCLI(t, minimalCLIXML))

	code, stdout, stderr := runCLI("inspect", pktPath, "--type", "Switch")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "R1") || !strings.Contains(stdout, "Switch") {
		t.Fatalf("expected R1 and Switch in filtered output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "matching \"Switch\"") {
		t.Fatalf("expected filter notice in stdout:\n%s", stdout)
	}

	code, stdout, stderr = runCLI("inspect", pktPath, "--type", "Router")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if strings.Contains(stdout, "R1") {
		t.Fatalf("did not expect R1 in Router-filtered output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "0 device(s) matching \"Router\"") {
		t.Fatalf("expected 0 matching devices notice:\n%s", stdout)
	}
}

func encodeCLI(t *testing.T, xml string) []byte {
	t.Helper()
	data, err := codec.Encode([]byte(xml))
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	return data
}

func writeCLIFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

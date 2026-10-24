package hooks

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
	"github.com/MarkOne-dev/pkt-git/internal/gitx"
)

const minimalXML = `<PACKETTRACER5><NETWORK><DEVICES><DEVICE><ENGINE><TYPE model="2960">Switch</TYPE><NAME translate="true">R1</NAME><RUNNINGCONFIG><LINE>hostname R1</LINE><LINE>!</LINE></RUNNINGCONFIG></ENGINE></DEVICE></DEVICES></NETWORK></PACKETTRACER5>`

func TestPreCommitExportsStagedBinaries(t *testing.T) {
	dir := initRepo(t)
	binary := encodeBinary(t, minimalXML)
	writeFile(t, dir, "project with space.pkt", binary)
	writeFile(t, dir, "activity.pka", binary)

	repo := gitx.New(dir)
	if err := repo.Add([]string{"project with space.pkt", "activity.pka"}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	result, err := PreCommit(dir, &out)
	if err != nil {
		t.Fatalf("PreCommit: %v", err)
	}
	if len(result.Exported) != 2 {
		t.Fatalf("Exported = %q, want two snapshots", result.Exported)
	}
	for _, project := range []string{"project with space", "activity"} {
		projectDir := filepath.Join(dir, "snapshot", project)
		if got := readFile(t, filepath.Join(projectDir, "base.snapshot.xml")); got != minimalXML {
			t.Fatalf("%s XML = %q", project, got)
		}
		if cfg := readFile(t, filepath.Join(projectDir, "configs", "R1.cfg")); !strings.Contains(cfg, "hostname R1") {
			t.Fatalf("%s config = %q", project, cfg)
		}
		metadata := readFile(t, filepath.Join(projectDir, "metadata", "topology.json"))
		if !strings.Contains(metadata, `"source_file": "`+project) {
			t.Fatalf("%s metadata does not name its source: %s", project, metadata)
		}
	}

	staged, err := repo.CachedChangedFiles("A", "snapshot")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"snapshot/project with space/base.snapshot.xml",
		"snapshot/project with space/configs/R1.cfg",
		"snapshot/project with space/metadata/topology.json",
		"snapshot/activity/base.snapshot.xml",
	} {
		if !contains(staged, want) {
			t.Fatalf("staged files do not include %q: %q", want, staged)
		}
	}
	if !strings.Contains(out.String(), "binary changed: project with space.pkt") {
		t.Fatalf("output does not report the export:\n%s", out.String())
	}
}

func TestPreCommitStagesRemovedConfig(t *testing.T) {
	dir := initRepo(t)
	twoDevices := strings.Replace(minimalXML, "</DEVICES>", `<DEVICE><ENGINE><TYPE model="2960">Switch</TYPE><NAME translate="true">R2</NAME><RUNNINGCONFIG><LINE>hostname R2</LINE><LINE>!</LINE></RUNNINGCONFIG></ENGINE></DEVICE></DEVICES>`, 1)
	writeFile(t, dir, "a.pkt", encodeBinary(t, twoDevices))
	repo := gitx.New(dir)
	if err := repo.Add([]string{"a.pkt"}); err != nil {
		t.Fatal(err)
	}
	if _, err := PreCommit(dir, nil); err != nil {
		t.Fatalf("baseline PreCommit: %v", err)
	}
	runGit(t, dir, "commit", "-q", "-m", "baseline")

	writeFile(t, dir, "a.pkt", encodeBinary(t, minimalXML))
	if err := repo.Add([]string{"a.pkt"}); err != nil {
		t.Fatal(err)
	}
	if _, err := PreCommit(dir, nil); err != nil {
		t.Fatalf("PreCommit: %v", err)
	}
	staged, err := repo.CachedChangedFiles("D", "snapshot/a/configs/R2.cfg")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(staged, "snapshot/a/configs/R2.cfg") {
		t.Fatalf("removed config deletion is not staged: %q", staged)
	}
	if _, err := os.Stat(filepath.Join(dir, "snapshot", "a", "configs", "R2.cfg")); !os.IsNotExist(err) {
		t.Fatalf("stale config still exists: %v", err)
	}
}

func TestPreCommitExportFailureAbortsWithoutStaging(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "bad.pkt", "this is not a packet tracer container")

	repo := gitx.New(dir)
	if err := repo.Add([]string{"bad.pkt"}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if _, err := PreCommit(dir, &out); err == nil {
		t.Fatal("PreCommit succeeded for an invalid binary")
	} else if !strings.Contains(err.Error(), "bad.pkt") {
		t.Fatalf("error does not name the binary: %v", err)
	}

	staged, err := repo.CachedChangedFiles("A", "snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if len(staged) != 0 {
		t.Fatalf("failed export staged a partial snapshot: %q", staged)
	}
	if _, err := os.Stat(filepath.Join(dir, "snapshot")); !os.IsNotExist(err) {
		t.Fatalf("failed export created snapshot/: %v", err)
	}
	if strings.Contains(out.String(), "included automatically") {
		t.Fatalf("failure output claimed success:\n%s", out.String())
	}
}

func TestPreCommitTextOnlySnapshotWarns(t *testing.T) {
	dir := initRepo(t)
	xmlPath := filepath.Join(dir, "snapshot", "project", "base.snapshot.xml")
	writeFile(t, dir, "snapshot/project/base.snapshot.xml", minimalXML)
	commitAll(t, dir, "snapshot")
	writeFile(t, dir, "snapshot/project/base.snapshot.xml", minimalXML+"\n")

	repo := gitx.New(dir)
	if err := repo.Add([]string{"snapshot/project/base.snapshot.xml"}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	result, err := PreCommit(dir, &out)
	if err != nil {
		t.Fatalf("PreCommit: %v", err)
	}
	if len(result.TextOnlySnapshots) != 1 || result.TextOnlySnapshots[0] != "snapshot/project/base.snapshot.xml" {
		t.Fatalf("TextOnlySnapshots = %q", result.TextOnlySnapshots)
	}
	text := out.String()
	if !strings.Contains(text, "textual snapshot changed without its .pkt") {
		t.Fatalf("missing warning:\n%s", text)
	}
	if !strings.Contains(text, "pkt build snapshot/project/base.snapshot.xml topologia/project.pkt") {
		t.Fatalf("missing rebuild command:\n%s", text)
	}
	if readFile(t, xmlPath) == minimalXML {
		t.Fatal("text-only check modified the staged snapshot")
	}
}

func TestPreCommitWarnsForOtherTextOnlySnapshot(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "a.pkt", encodeBinary(t, minimalXML))
	writeFile(t, dir, "snapshot/other/base.snapshot.xml", minimalXML)
	commitAll(t, dir, "baseline")

	writeFile(t, dir, "a.pkt", encodeBinary(t, strings.Replace(minimalXML, "hostname R1", "hostname A2", 1)))
	writeFile(t, dir, "snapshot/other/base.snapshot.xml", minimalXML+"\n")
	writeFile(t, dir, "snapshot/other/README.md", "unrelated snapshot note\n")
	repo := gitx.New(dir)
	if err := repo.Add([]string{"a.pkt", "snapshot/other/base.snapshot.xml"}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	result, err := PreCommit(dir, &out)
	if err != nil {
		t.Fatalf("PreCommit: %v", err)
	}
	if !contains(result.Exported, "snapshot/a") {
		t.Fatalf("Exported = %q", result.Exported)
	}
	if len(result.TextOnlySnapshots) != 1 || result.TextOnlySnapshots[0] != "snapshot/other/base.snapshot.xml" {
		t.Fatalf("TextOnlySnapshots = %q", result.TextOnlySnapshots)
	}
	if !strings.Contains(out.String(), "textual snapshot changed without its .pkt") {
		t.Fatalf("missing warning for the other project:\n%s", out.String())
	}
	staged, err := repo.CachedChangedFiles("A", "snapshot/other/README.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(staged) != 0 {
		t.Fatalf("pre-commit staged an unrelated snapshot file: %q", staged)
	}
}

func TestPreCommitIgnoresUnrelatedChanges(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "notes.txt", "unrelated")
	repo := gitx.New(dir)
	if err := repo.Add([]string{"notes.txt"}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	result, err := PreCommit(dir, &out)
	if err != nil {
		t.Fatalf("PreCommit: %v", err)
	}
	if len(result.Exported) != 0 || len(result.TextOnlySnapshots) != 0 {
		t.Fatalf("result = %+v, want empty", result)
	}
	if out.Len() != 0 {
		t.Fatalf("unrelated commit produced output:\n%s", out.String())
	}
}

func encodeBinary(t *testing.T, xml string) []byte {
	t.Helper()
	data, err := codec.Encode([]byte(xml))
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	return data
}

func writeFile(t *testing.T, dir, name string, data any) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	var raw []byte
	switch value := data.(type) {
	case []byte:
		raw = value
	case string:
		raw = []byte(value)
	default:
		t.Fatalf("writeFile: unsupported type %T", data)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitAll(t *testing.T, dir, message string) {
	t.Helper()
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", message, "--no-gpg-sign")
}

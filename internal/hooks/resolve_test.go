package hooks

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/gitx"
)

func TestResolveReportMode(t *testing.T) {
	dir := createTextConflict(t)
	xmlPath := filepath.Join(dir, "snapshot", "project", "base.snapshot.xml")
	before, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	result, err := Resolve(dir, "", &out)
	if !errors.Is(err, ErrReportOnly) {
		t.Fatalf("error = %v, want ErrReportOnly", err)
	}
	if len(result.Text) != 1 || result.Text[0] != "snapshot/project/base.snapshot.xml" {
		t.Fatalf("Text = %q", result.Text)
	}
	if len(result.Binary) != 0 {
		t.Fatalf("Binary = %q", result.Binary)
	}
	text := out.String()
	if !strings.Contains(text, "conflict! current branch vs") {
		t.Fatalf("missing conflict header:\n%s", text)
	}
	if !strings.Contains(text, "first marker at line") {
		t.Fatalf("missing marker location:\n%s", text)
	}
	if !strings.Contains(text, "pkt resolve --ours") || !strings.Contains(text, "pkt resolve --theirs") {
		t.Fatalf("missing resolution choices:\n%s", text)
	}

	after, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("report mode modified the conflicted file")
	}
	repo := gitx.New(dir)
	staged, err := repo.CachedChangedFiles("ACMR")
	if err != nil {
		t.Fatal(err)
	}
	if len(staged) != 0 {
		t.Fatalf("report mode staged paths: %q", staged)
	}
}

func TestResolvePolicy(t *testing.T) {
	tests := []struct {
		policy string
		want   string
	}{
		{policy: "ours", want: "hostname MAIN"},
		{policy: "theirs", want: "hostname SIDE"},
	}
	for _, tt := range tests {
		t.Run(tt.policy, func(t *testing.T) {
			dir := createTextConflict(t)
			repo := gitx.New(dir)
			writeFile(t, dir, "notes.txt", "local unrelated work")
			if err := repo.Add([]string{"notes.txt"}); err != nil {
				t.Fatal(err)
			}
			writeFile(t, dir, "snapshot/project/README.md", "unrelated snapshot note\n")
			writeFile(t, dir, "snapshot/project/configs/UNTRACKED.cfg", "untracked\n")

			var out bytes.Buffer
			result, err := Resolve(dir, tt.policy, &out)
			if err != nil {
				t.Fatalf("Resolve %s: %v", tt.policy, err)
			}
			xml := readFile(t, filepath.Join(dir, "snapshot", "project", "base.snapshot.xml"))
			if !strings.Contains(xml, tt.want) {
				t.Fatalf("resolved XML does not contain %q:\n%s", tt.want, xml)
			}
			if hasConflictMarker(xml) {
				t.Fatalf("resolved XML still has markers:\n%s", xml)
			}
			cfg := readFile(t, filepath.Join(dir, "snapshot", "project", "configs", "R1.cfg"))
			if !strings.Contains(cfg, tt.want) {
				t.Fatalf("derived config does not contain %q:\n%s", tt.want, cfg)
			}
			if !contains(result.Refreshed, "snapshot/project") {
				t.Fatalf("Refreshed = %q", result.Refreshed)
			}
			if !contains(result.Staged, "snapshot/project/configs/R1.cfg") {
				t.Fatalf("Staged = %q", result.Staged)
			}
			if !strings.Contains(out.String(), "resolution staged") {
				t.Fatalf("missing completion message:\n%s", out.String())
			}

			staged, err := repo.CachedChangedFiles("ACMR")
			if err != nil {
				t.Fatal(err)
			}
			if !contains(result.Staged, "snapshot/project/configs/R1.cfg") {
				t.Fatalf("generated config is not staged: %q", result.Staged)
			}
			if !contains(staged, "snapshot/project/configs/R1.cfg") {
				t.Fatalf("derived config is not staged: %q", staged)
			}
			if !contains(staged, "notes.txt") {
				t.Fatalf("unrelated staged file was lost: %q", staged)
			}
			if contains(staged, "snapshot/project/README.md") {
				t.Fatalf("resolve staged an unrelated snapshot file: %q", staged)
			}
			if contains(staged, "snapshot/project/configs/UNTRACKED.cfg") {
				t.Fatalf("resolve staged an untracked stale config: %q", staged)
			}
			if !strings.Contains(out.String(), "resolution staged") {
				t.Fatalf("missing completion message:\n%s", out.String())
			}
		})
	}
}

func TestResolveBinaryOnlyConflict(t *testing.T) {
	dir := initRepo(t)
	xmlPath := "snapshot/project/base.snapshot.xml"
	binaryPath := "topologia/project.pkt"
	writeFile(t, dir, xmlPath, minimalXML)
	writeFile(t, dir, binaryPath, encodeBinary(t, minimalXML))
	commitAll(t, dir, "base")

	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, binaryPath, encodeBinary(t, strings.Replace(minimalXML, "hostname R1", "hostname SIDE", 1)))
	commitAll(t, dir, "side binary")

	runGit(t, dir, "checkout", "-q", "main")
	writeFile(t, dir, binaryPath, encodeBinary(t, strings.Replace(minimalXML, "hostname R1", "hostname MAIN", 1)))
	commitAll(t, dir, "main binary")

	runGitExpectFailure(t, dir, "merge", "side")

	var out bytes.Buffer
	result, err := Resolve(dir, "theirs", &out)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(result.Binary) != 1 || result.Binary[0] != binaryPath {
		t.Fatalf("Binary = %q", result.Binary)
	}
	if !contains(result.Rebuilt, binaryPath) {
		t.Fatalf("Rebuilt = %q", result.Rebuilt)
	}
	if got := decodeFile(t, filepath.Join(dir, binaryPath)); got != minimalXML {
		t.Fatalf("rebuilt binary decodes to %q, want the snapshot XML", got)
	}
	repo := gitx.New(dir)
	staged, err := repo.CachedChangedFiles("ACMR")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(staged, binaryPath) {
		t.Fatalf("rebuilt binary is not staged: %q", staged)
	}
	if !strings.Contains(out.String(), "rebuilt from the resolved snapshot") {
		t.Fatalf("missing rebuild report:\n%s", out.String())
	}
}

func TestResolveBinaryConflictMissingSnapshot(t *testing.T) {
	dir := initRepo(t)
	binaryPath := "topologia/project.pkt"
	writeFile(t, dir, binaryPath, encodeBinary(t, minimalXML))
	commitAll(t, dir, "base")

	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, binaryPath, encodeBinary(t, strings.Replace(minimalXML, "hostname R1", "hostname SIDE", 1)))
	commitAll(t, dir, "side")
	runGit(t, dir, "checkout", "-q", "main")
	writeFile(t, dir, binaryPath, encodeBinary(t, strings.Replace(minimalXML, "hostname R1", "hostname MAIN", 1)))
	commitAll(t, dir, "main")
	runGitExpectFailure(t, dir, "merge", "side")

	if _, err := Resolve(dir, "ours", nil); err == nil || !strings.Contains(err.Error(), "snapshot") {
		t.Fatalf("error = %v, want a missing-snapshot error", err)
	}
	repo := gitx.New(dir)
	staged, err := repo.CachedChangedFiles("ACMR")
	if err != nil {
		t.Fatal(err)
	}
	if len(staged) != 0 {
		t.Fatalf("missing-snapshot resolution staged paths: %q", staged)
	}
}

func TestResolveRejectsRemainingMarkers(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "notes.txt", "base\n")
	commitAll(t, dir, "base")

	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, "notes.txt", "<<<<<<< literal side marker\n")
	commitAll(t, dir, "side")
	runGit(t, dir, "checkout", "-q", "main")
	writeFile(t, dir, "notes.txt", "main\n")
	commitAll(t, dir, "main")
	runGitExpectFailure(t, dir, "merge", "side")

	if _, err := Resolve(dir, "theirs", nil); err == nil || !strings.Contains(err.Error(), "conflict markers") {
		t.Fatalf("error = %v, want a remaining-marker error", err)
	}
	repo := gitx.New(dir)
	staged, err := repo.CachedChangedFiles("ACMR")
	if err != nil {
		t.Fatal(err)
	}
	if len(staged) != 0 {
		t.Fatalf("marker failure staged paths: %q", staged)
	}
}

func TestHasConflictMarkerOnlyMatchesLineStart(t *testing.T) {
	if hasConflictMarker(" banner ======= text\n") {
		t.Fatal("hasConflictMarker matched a substring instead of a marker line")
	}
	for _, marker := range []string{"<<<<<<< HEAD\n", "=======\n", ">>>>>>> side\n"} {
		if !hasConflictMarker(marker) {
			t.Fatalf("hasConflictMarker did not match %q", marker)
		}
	}
}

func TestResolveNoMerge(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "file.txt", "data")
	commitAll(t, dir, "commit")

	if _, err := Resolve(dir, "ours", nil); !errors.Is(err, ErrNoMerge) {
		t.Fatalf("error = %v, want ErrNoMerge", err)
	}
}

func createTextConflict(t *testing.T) string {
	t.Helper()
	dir := initRepo(t)
	xmlPath := "snapshot/project/base.snapshot.xml"
	writeFile(t, dir, xmlPath, minimalXML)
	commitAll(t, dir, "base")

	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, xmlPath, strings.Replace(minimalXML, "hostname R1", "hostname SIDE", 1))
	commitAll(t, dir, "side")

	runGit(t, dir, "checkout", "-q", "main")
	writeFile(t, dir, xmlPath, strings.Replace(minimalXML, "hostname R1", "hostname MAIN", 1))
	commitAll(t, dir, "main")
	runGitExpectFailure(t, dir, "merge", "side")
	return dir
}

func runGitExpectFailure(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	if err := cmd.Run(); err == nil {
		t.Fatalf("git %s succeeded unexpectedly", strings.Join(args, " "))
	}
}

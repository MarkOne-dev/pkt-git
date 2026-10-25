package gitx

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigGetSet(t *testing.T) {
	repo, _ := initRepo(t)

	if err := repo.ConfigSet("pktgit.bin", "/opt/pkt"); err != nil {
		t.Fatalf("ConfigSet: %v", err)
	}
	got, err := repo.ConfigGet("pktgit.bin")
	if err != nil {
		t.Fatalf("ConfigGet: %v", err)
	}
	if got != "/opt/pkt" {
		t.Fatalf("ConfigGet = %q, want /opt/pkt", got)
	}

	if _, err := repo.ConfigGet("pktgit.missing"); !errors.Is(err, ErrConfigMissing) {
		t.Fatalf("missing key error = %v, want ErrConfigMissing", err)
	}
}

func TestChangedFilesPreservesSpacesAndRevisions(t *testing.T) {
	repo, dir := initRepo(t)
	path := "project with space.pkt"
	writeFile(t, dir, path, "one")
	commitAll(t, dir, "first")

	writeFile(t, dir, path, "two")
	commitAll(t, dir, "second")

	files, err := repo.ChangedFiles([]string{"HEAD~1", "HEAD"}, "*.pkt")
	if err != nil {
		t.Fatalf("ChangedFiles: %v", err)
	}
	if len(files) != 1 || files[0] != path {
		t.Fatalf("changed files = %q, want [%q]", files, path)
	}

	writeFile(t, dir, path, "working tree")
	files, err = repo.ChangedFiles(nil, "*.pkt")
	if err != nil {
		t.Fatalf("ChangedFiles working tree: %v", err)
	}
	if len(files) != 1 || files[0] != path {
		t.Fatalf("working-tree files = %q, want [%q]", files, path)
	}
}

func TestCachedChangedFilesPreservesSpaces(t *testing.T) {
	repo, dir := initRepo(t)
	path := "staged file with space.pkt"
	writeFile(t, dir, path, "staged")

	if err := repo.Add([]string{path}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	files, err := repo.CachedChangedFiles("ACMR", "*.pkt")
	if err != nil {
		t.Fatalf("CachedChangedFiles: %v", err)
	}
	if len(files) != 1 || files[0] != path {
		t.Fatalf("staged files = %q, want [%q]", files, path)
	}
}

func TestDiffTreeRootFiles(t *testing.T) {
	repo, dir := initRepo(t)
	path := "root file.txt"
	writeFile(t, dir, path, "root")
	commitAll(t, dir, "root")

	files, err := repo.DiffTreeRootFiles("HEAD", "*.txt")
	if err != nil {
		t.Fatalf("DiffTreeRootFiles: %v", err)
	}
	if len(files) != 1 || files[0] != path {
		t.Fatalf("root files = %q, want [%q]", files, path)
	}
}

func TestConflictedFilesCheckoutAndAdd(t *testing.T) {
	repo, dir := initRepo(t)
	path := "conflict file.txt"
	writeFile(t, dir, path, "base")
	commitAll(t, dir, "base")

	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, path, "side")
	commitAll(t, dir, "side")

	runGit(t, dir, "checkout", "-q", "main")
	writeFile(t, dir, path, "main")
	commitAll(t, dir, "main")

	runGitExpectFailure(t, dir, "merge", "side")

	conflicts, err := repo.ConflictedFiles()
	if err != nil {
		t.Fatalf("ConflictedFiles: %v", err)
	}
	if len(conflicts) != 1 || conflicts[0] != path {
		t.Fatalf("conflicts = %q, want [%q]", conflicts, path)
	}

	if err := repo.Checkout("theirs", conflicts); err != nil {
		t.Fatalf("Checkout theirs: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, path))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "side" {
		t.Fatalf("resolved content = %q, want side", data)
	}

	if err := repo.Add(conflicts); err != nil {
		t.Fatalf("Add: %v", err)
	}
	staged, err := repo.CachedChangedFiles("M", "*.txt")
	if err != nil {
		t.Fatalf("CachedChangedFiles: %v", err)
	}
	if len(staged) != 1 || staged[0] != path {
		t.Fatalf("staged after resolve = %q, want [%q]", staged, path)
	}

	if err := repo.Checkout("invalid", conflicts); err == nil {
		t.Fatal("Checkout accepted an invalid policy")
	}
}

func TestRevisionHelpers(t *testing.T) {
	repo, dir := initRepo(t)
	writeFile(t, dir, "file.txt", "data")
	commitAll(t, dir, "commit")

	hash, err := repo.RevParse("HEAD")
	if err != nil {
		t.Fatalf("RevParse: %v", err)
	}
	if len(hash) != 40 {
		t.Fatalf("RevParse = %q, want a 40-character hash", hash)
	}
	if name := repo.NameRev("HEAD"); name == "" {
		t.Fatal("NameRev returned an empty name")
	}
	if root, err := repo.Toplevel(); err != nil || root == "" {
		t.Fatalf("Toplevel = %q, %v", root, err)
	}
	if gitDir, err := repo.GitDir(); err != nil || gitDir == "" {
		t.Fatalf("GitDir = %q, %v", gitDir, err)
	}
}

func initRepo(t *testing.T) (*Repo, string) {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	return New(dir), dir
}

func writeFile(t *testing.T, dir, path, content string) {
	t.Helper()
	full := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitAll(t *testing.T, dir, message string) {
	t.Helper()
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", message, "--no-gpg-sign")
}

func runGit(t *testing.T, dir string, args ...string) string {
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
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
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

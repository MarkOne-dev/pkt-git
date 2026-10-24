package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncCommand(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "base.snapshot.xml"), []byte(editCLIXML))
	writeCLIFile(t, filepath.Join(dir, "configs", "stale.cfg"), []byte("old\n"))

	code, stdout, stderr := runCLI("sync", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "[ok] refreshed") || !strings.Contains(stdout, "removed stale configs: stale.cfg") {
		t.Fatalf("stdout = %q", stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "configs", "R1.cfg")); err != nil {
		t.Fatalf("config was not generated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "configs", "stale.cfg")); !os.IsNotExist(err) {
		t.Fatalf("stale config was not pruned: %v", err)
	}
}

func TestSyncCommandUsage(t *testing.T) {
	code, _, stderr := runCLI("sync")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected 1 argument") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestResolveRejectsBothPolicies(t *testing.T) {
	code, _, stderr := runCLI("resolve", "--ours", "--theirs")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestResolveWithoutMerge(t *testing.T) {
	dir := initCLIRepo(t)
	t.Chdir(dir)

	code, _, stderr := runCLI("resolve", "--ours")
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "no merge in progress") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestResolveReportMode(t *testing.T) {
	dir := initCLIRepo(t)
	xmlPath := filepath.Join(dir, "snapshot", "project", "base.snapshot.xml")
	writeCLIFile(t, xmlPath, []byte(editCLIXML))
	runGitCLI(t, dir, "add", ".")
	runGitCLI(t, dir, "commit", "-q", "-m", "base")

	runGitCLI(t, dir, "checkout", "-q", "-b", "side")
	writeCLIFile(t, xmlPath, []byte(strings.Replace(editCLIXML, "description Old", "description SIDE", 1)))
	runGitCLI(t, dir, "add", "snapshot/project/base.snapshot.xml")
	runGitCLI(t, dir, "commit", "-q", "-m", "side")
	runGitCLI(t, dir, "checkout", "-q", "main")
	writeCLIFile(t, xmlPath, []byte(strings.Replace(editCLIXML, "description Old", "description MAIN", 1)))
	runGitCLI(t, dir, "add", "snapshot/project/base.snapshot.xml")
	runGitCLI(t, dir, "commit", "-q", "-m", "main")
	runGitCLIExpectFailure(t, dir, "merge", "side")
	t.Chdir(dir)

	code, stdout, stderr := runCLI("resolve")
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stdout, "conflict! current branch") {
		t.Fatalf("stdout = %q", stdout)
	}
	if !strings.Contains(stderr, "conflict report only") {
		t.Fatalf("stderr = %q", stderr)
	}
	if !strings.Contains(readEditFile(t, xmlPath), "<<<<<<<") {
		t.Fatal("report mode changed the conflicted file")
	}
}

func TestVerifyCommand(t *testing.T) {
	dir := t.TempDir()
	pktPath := filepath.Join(dir, "project.pkt")
	writeCLIFile(t, pktPath, encodeCLI(t, minimalCLIXML))

	code, stdout, stderr := runCLI("verify", pktPath)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "XML is a document  : true") || !strings.Contains(stdout, "XML round trip     : true") {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestVerifyCommandFailure(t *testing.T) {
	dir := t.TempDir()
	pktPath := filepath.Join(dir, "broken.pkt")
	writeCLIFile(t, pktPath, []byte("not a container"))

	code, _, stderr := runCLI("verify", pktPath)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "verify:") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestVerifyCommandUsage(t *testing.T) {
	code, _, stderr := runCLI("verify")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected 1 argument") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func initCLIRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGitCLI(t, dir, "init", "-q", "-b", "main")
	runGitCLI(t, dir, "config", "user.email", "test@example.com")
	runGitCLI(t, dir, "config", "user.name", "Test User")
	return dir
}

func runGitCLI(t *testing.T, dir string, args ...string) {
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
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func runGitCLIExpectFailure(t *testing.T, dir string, args ...string) {
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

package hooks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain lets the generated hooks invoke this test binary as a stand-in for
// the pkt executable. The wrapper created by installHooks sets the environment
// variable, so normal test runs are unaffected.
func TestMain(m *testing.M) {
	if os.Getenv("PKTG_TEST_HOOK") == "1" {
		os.Exit(runTestHook(os.Args[1:]))
	}
	os.Exit(m.Run())
}

func runTestHook(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "test hook: missing command")
		return 2
	}
	var err error
	switch args[0] {
	case "pre-commit":
		_, err = PreCommit(".", os.Stdout)
	case "post-merge":
		_, err = PostMerge(".", args[1:], os.Stdout)
	case "post-checkout":
		_, err = PostCheckout(".", args[1:], os.Stdout)
	default:
		err = fmt.Errorf("test hook: unknown command %q", args[0])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func TestIntegrationInitAndCommit(t *testing.T) {
	dir := initRepo(t)
	installHooks(t, dir)
	writeFile(t, dir, "project.pkt", encodeBinary(t, minimalXML))

	runGit(t, dir, "add", "project.pkt")
	runGit(t, dir, "commit", "-q", "-m", "add project")

	files := runGit(t, dir, "show", "--name-only", "--pretty=format:", "HEAD")
	for _, want := range []string{
		"snapshot/project/base.snapshot.xml",
		"snapshot/project/configs/R1.cfg",
		"snapshot/project/metadata/topology.json",
	} {
		if !strings.Contains(files, want) {
			t.Fatalf("commit does not contain %q:\n%s", want, files)
		}
	}
}

func TestIntegrationMergeRebuildsBinary(t *testing.T) {
	dir := setupHookRepo(t)
	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, "snapshot/project/base.snapshot.xml",
		strings.Replace(minimalXML, "hostname R1", "hostname MERGE_SIDE", 1))
	runGit(t, dir, "add", "snapshot/project/base.snapshot.xml")
	runGit(t, dir, "commit", "-q", "-m", "side snapshot")

	runGit(t, dir, "checkout", "-q", "main")
	runGit(t, dir, "merge", "--no-ff", "-m", "merge side", "side")

	xml := decodeFile(t, filepath.Join(dir, "topologia", "project.pkt"))
	if !strings.Contains(xml, "hostname MERGE_SIDE") {
		t.Fatalf("post-merge did not rebuild from the merged snapshot:\n%s", xml)
	}
}

func TestIntegrationCheckoutRebuildsBinary(t *testing.T) {
	dir := setupHookRepo(t)
	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, "snapshot/project/base.snapshot.xml",
		strings.Replace(minimalXML, "hostname R1", "hostname CHECKOUT_SIDE", 1))
	runGit(t, dir, "add", "snapshot/project/base.snapshot.xml")
	runGit(t, dir, "commit", "-q", "-m", "side snapshot")

	runGit(t, dir, "checkout", "-q", "main")
	runGit(t, dir, "checkout", "-q", "side")

	xml := decodeFile(t, filepath.Join(dir, "topologia", "project.pkt"))
	if !strings.Contains(xml, "hostname CHECKOUT_SIDE") {
		t.Fatalf("post-checkout did not rebuild from the checked-out snapshot:\n%s", xml)
	}
}

func TestIntegrationSquashMergeRebuildsBinary(t *testing.T) {
	dir := setupHookRepo(t)
	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, "snapshot/project/base.snapshot.xml",
		strings.Replace(minimalXML, "hostname R1", "hostname SQUASH_SIDE", 1))
	runGit(t, dir, "add", "snapshot/project/base.snapshot.xml")
	runGit(t, dir, "commit", "-q", "-m", "side snapshot")

	runGit(t, dir, "checkout", "-q", "main")
	runGit(t, dir, "merge", "--squash", "side")

	xml := decodeFile(t, filepath.Join(dir, "topologia", "project.pkt"))
	if !strings.Contains(xml, "hostname SQUASH_SIDE") {
		t.Fatalf("squash post-merge did not rebuild from the squash snapshot:\n%s", xml)
	}
}

func TestIntegrationConflictResolution(t *testing.T) {
	dir := createTextConflict(t)
	installHooks(t, dir)

	var out bytes.Buffer
	result, err := Resolve(dir, "theirs", &out)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(result.Text) != 1 {
		t.Fatalf("Text = %q", result.Text)
	}
	xml := readFile(t, filepath.Join(dir, "snapshot", "project", "base.snapshot.xml"))
	if !strings.Contains(xml, "hostname SIDE") || hasConflictMarker(xml) {
		t.Fatalf("conflict was not resolved to the incoming side:\n%s", xml)
	}
	if !strings.Contains(readFile(t, filepath.Join(dir, "snapshot", "project", "configs", "R1.cfg")), "hostname SIDE") {
		t.Fatal("conflict resolution did not refresh the derived config")
	}
}

func setupHookRepo(t *testing.T) string {
	t.Helper()
	dir := initRepo(t)
	installHooks(t, dir)
	writeFile(t, dir, "snapshot/project/base.snapshot.xml", minimalXML)
	writeFile(t, dir, "topologia/project.pkt", encodeBinary(t, minimalXML))
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "baseline")
	return dir
}

func installHooks(t *testing.T, dir string) {
	t.Helper()
	testBinary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	wrapperDir := t.TempDir()
	wrapper := filepath.Join(wrapperDir, "pkt-wrapper")
	body := "#!/bin/sh\nPKTG_TEST_HOOK=1 exec " + shellQuote(filepath.ToSlash(testBinary)) + " \"$@\"\n"
	if err := os.WriteFile(wrapper, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(dir, wrapper); err != nil {
		t.Fatalf("Init: %v", err)
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

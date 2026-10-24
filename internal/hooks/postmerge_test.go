package hooks

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
)

func TestPostCheckoutRebuildsChangedSnapshot(t *testing.T) {
	dir := initRepo(t)
	xmlPath := "snapshot/project/base.snapshot.xml"
	writeFile(t, dir, xmlPath, minimalXML)
	commitAll(t, dir, "snapshot")
	oldRev := revParse(t, dir, "HEAD")

	modified := strings.Replace(minimalXML, "hostname R1", "hostname R2", 1)
	writeFile(t, dir, xmlPath, modified)
	commitAll(t, dir, "change")
	newRev := revParse(t, dir, "HEAD")

	var out bytes.Buffer
	result, err := PostCheckout(dir, []string{oldRev, newRev, "0"}, &out)
	if err != nil {
		t.Fatalf("PostCheckout flag 0: %v", err)
	}
	if len(result.Rebuilt) != 0 || out.Len() != 0 {
		t.Fatalf("non-branch checkout rebuilt: %+v\n%s", result, out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "topologia", "project.pkt")); !os.IsNotExist(err) {
		t.Fatalf("non-branch checkout created a binary: %v", err)
	}

	result, err = PostCheckout(dir, []string{oldRev, newRev, "1"}, &out)
	if err != nil {
		t.Fatalf("PostCheckout: %v", err)
	}
	if len(result.Rebuilt) != 1 || result.Rebuilt[0] != "topologia/project.pkt" {
		t.Fatalf("Rebuilt = %q", result.Rebuilt)
	}
	if got := decodeFile(t, filepath.Join(dir, "topologia", "project.pkt")); got != modified {
		t.Fatalf("rebuilt XML = %q, want %q", got, modified)
	}
	if !strings.Contains(out.String(), "topologia/project.pkt") {
		t.Fatalf("output does not report the rebuild:\n%s", out.String())
	}
}

func TestPostCheckoutRootCommit(t *testing.T) {
	dir := initRepo(t)
	xmlPath := "snapshot/root/base.snapshot.xml"
	writeFile(t, dir, xmlPath, minimalXML)
	commitAll(t, dir, "root")
	newRev := revParse(t, dir, "HEAD")

	var out bytes.Buffer
	result, err := PostCheckout(dir, []string{strings.Repeat("0", 40), newRev, "1"}, &out)
	if err != nil {
		t.Fatalf("PostCheckout: %v", err)
	}
	if len(result.Rebuilt) != 1 || result.Rebuilt[0] != "topologia/root.pkt" {
		t.Fatalf("Rebuilt = %q", result.Rebuilt)
	}
	if got := decodeFile(t, filepath.Join(dir, "topologia", "root.pkt")); got != minimalXML {
		t.Fatalf("rebuilt XML = %q", got)
	}
}

func TestPostCheckoutUnchangedSnapshot(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "snapshot/project/base.snapshot.xml", minimalXML)
	commitAll(t, dir, "snapshot")
	rev := revParse(t, dir, "HEAD")

	var out bytes.Buffer
	result, err := PostCheckout(dir, []string{rev, rev, "1"}, &out)
	if err != nil {
		t.Fatalf("PostCheckout: %v", err)
	}
	if len(result.Rebuilt) != 0 || out.Len() != 0 {
		t.Fatalf("unchanged snapshot was rebuilt: %+v\n%s", result, out.String())
	}
}

func TestPostMergeRebuildsChangedSnapshot(t *testing.T) {
	dir := initRepo(t)
	xmlPath := "snapshot/project/base.snapshot.xml"
	writeFile(t, dir, xmlPath, minimalXML)
	commitAll(t, dir, "snapshot")

	runGit(t, dir, "checkout", "-q", "-b", "side")
	modified := strings.Replace(minimalXML, "hostname R1", "hostname SIDE", 1)
	writeFile(t, dir, xmlPath, modified)
	commitAll(t, dir, "side change")

	runGit(t, dir, "checkout", "-q", "main")
	runGit(t, dir, "merge", "--no-ff", "-m", "merge side", "side")

	var out bytes.Buffer
	result, err := PostMerge(dir, []string{"1"}, &out)
	if err != nil {
		t.Fatalf("PostMerge: %v", err)
	}
	if len(result.Rebuilt) != 1 || result.Rebuilt[0] != "topologia/project.pkt" {
		t.Fatalf("Rebuilt = %q", result.Rebuilt)
	}
	if got := decodeFile(t, filepath.Join(dir, "topologia", "project.pkt")); got != modified {
		t.Fatalf("rebuilt XML = %q, want %q", got, modified)
	}
}

func TestPostMergeFallsBackWithoutOrigHead(t *testing.T) {
	dir := initRepo(t)
	xmlPath := "snapshot/project/base.snapshot.xml"
	writeFile(t, dir, xmlPath, minimalXML)
	commitAll(t, dir, "snapshot")
	modified := strings.Replace(minimalXML, "hostname R1", "hostname FALLBACK", 1)
	writeFile(t, dir, xmlPath, modified)
	commitAll(t, dir, "change")

	_ = os.Remove(filepath.Join(dir, ".git", "ORIG_HEAD"))

	var out bytes.Buffer
	result, err := PostMerge(dir, nil, &out)
	if err != nil {
		t.Fatalf("PostMerge: %v", err)
	}
	if len(result.Rebuilt) != 1 {
		t.Fatalf("Rebuilt = %q", result.Rebuilt)
	}
	if got := decodeFile(t, filepath.Join(dir, "topologia", "project.pkt")); got != modified {
		t.Fatalf("rebuilt XML = %q, want %q", got, modified)
	}
}

func TestPostMergeRejectsRevisionAsSquashFlag(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "file.txt", "data")
	commitAll(t, dir, "commit")

	if _, err := PostMerge(dir, []string{"HEAD"}, nil); err == nil {
		t.Fatal("PostMerge accepted a revision as the squash flag")
	}
}

func TestRebuildFailureReturnsError(t *testing.T) {
	dir := initRepo(t)
	xmlPath := "snapshot/project/base.snapshot.xml"
	writeFile(t, dir, xmlPath, minimalXML)
	commitAll(t, dir, "valid")
	oldRev := revParse(t, dir, "HEAD")

	writeFile(t, dir, xmlPath, "<PACKETTRACER5>")
	commitAll(t, dir, "invalid")
	newRev := revParse(t, dir, "HEAD")

	var out bytes.Buffer
	result, err := PostCheckout(dir, []string{oldRev, newRev, "1"}, &out)
	if !errors.Is(err, ErrRebuildFailed) {
		t.Fatalf("error = %v, want ErrRebuildFailed", err)
	}
	if len(result.Failed) != 1 || result.Failed[0] != "topologia/project.pkt" {
		t.Fatalf("Failed = %q", result.Failed)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "topologia", "project.pkt")); !os.IsNotExist(statErr) {
		t.Fatalf("invalid snapshot produced a binary: %v", statErr)
	}
	if !strings.Contains(out.String(), "ERROR rebuilding topologia/project.pkt") {
		t.Fatalf("output does not report the failure:\n%s", out.String())
	}
}

func revParse(t *testing.T, dir, rev string) string {
	t.Helper()
	return strings.TrimSpace(runGit(t, dir, "rev-parse", rev))
}

func decodeFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	xml, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return string(xml)
}

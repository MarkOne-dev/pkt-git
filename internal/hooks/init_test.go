package hooks

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/gitx"
)

func TestInitCreatesHooksAndConfig(t *testing.T) {
	dir := initRepo(t)
	result, err := Init(dir, "/opt/pkt/bin/pkt")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if !filepath.IsAbs(result.Executable) || result.Executable != "/opt/pkt/bin/pkt" {
		t.Fatalf("Executable = %q", result.Executable)
	}

	for _, name := range hookNames {
		path := filepath.Join(dir, HookDir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if info.Mode().Perm()&0o111 == 0 {
			t.Fatalf("%s is not executable: %v", path, info.Mode())
		}
		body := readFile(t, path)
		if !strings.Contains(body, "pktgit.bin") || !strings.Contains(body, name) {
			t.Fatalf("%s has unexpected body:\n%s", path, body)
		}
	}

	repo := gitx.New(dir)
	if got, err := repo.ConfigGet(HooksPathKey); err != nil || got != HookDir {
		t.Fatalf("%s = %q, %v; want %q", HooksPathKey, got, err, HookDir)
	}
	if got, err := repo.ConfigGet(BinConfigKey); err != nil || got != "/opt/pkt/bin/pkt" {
		t.Fatalf("%s = %q, %v; want /opt/pkt/bin/pkt", BinConfigKey, got, err)
	}

	attrs := readFile(t, filepath.Join(dir, ".gitattributes"))
	for _, rule := range requiredAttributes {
		if strings.Count(attrs, rule+"\n") != 1 {
			t.Fatalf("attribute %q appears %d times:\n%s", rule, strings.Count(attrs, rule+"\n"), attrs)
		}
	}
	if !result.AttributesCreated || !result.AttributesChanged {
		t.Fatalf("result = %+v, want created attributes", result)
	}
}

func TestInitIsIdempotent(t *testing.T) {
	dir := initRepo(t)
	if _, err := Init(dir, "/opt/pkt"); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	attrsBefore := readFile(t, filepath.Join(dir, ".gitattributes"))
	hookBefore := readFile(t, filepath.Join(dir, HookDir, "pre-commit"))

	result, err := Init(dir, "/opt/pkt")
	if err != nil {
		t.Fatalf("second Init: %v", err)
	}
	if result.AttributesCreated || result.AttributesChanged || len(result.AttributesAdded) != 0 {
		t.Fatalf("second Init changed attributes: %+v", result)
	}
	if got := readFile(t, filepath.Join(dir, ".gitattributes")); got != attrsBefore {
		t.Fatalf(".gitattributes changed:\nbefore:\n%s\nafter:\n%s", attrsBefore, got)
	}
	if got := readFile(t, filepath.Join(dir, HookDir, "pre-commit")); got != hookBefore {
		t.Fatalf("pre-commit changed:\nbefore:\n%s\nafter:\n%s", hookBefore, got)
	}
}

func TestInitMergesAttributes(t *testing.T) {
	dir := initRepo(t)
	existing := "# keep this comment\n*.txt text\n*.pkt binary\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitattributes"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Init(dir, "/opt/pkt")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	attrs := readFile(t, filepath.Join(dir, ".gitattributes"))
	for _, want := range []string{"# keep this comment", "*.txt text"} {
		if !strings.Contains(attrs, want) {
			t.Fatalf("unrelated attribute %q was lost:\n%s", want, attrs)
		}
	}
	if strings.Count(attrs, "*.pkt binary\n") != 1 {
		t.Fatalf("*.pkt binary was duplicated:\n%s", attrs)
	}
	for _, rule := range requiredAttributes {
		if strings.Count(attrs, rule+"\n") != 1 {
			t.Fatalf("attribute %q appears %d times:\n%s", rule, strings.Count(attrs, rule+"\n"), attrs)
		}
	}
	if result.AttributesCreated || !result.AttributesChanged {
		t.Fatalf("result = %+v, want changed existing attributes", result)
	}
	if contains(result.AttributesAdded, "*.pkt binary") {
		t.Fatalf("existing rule was reported as added: %+v", result.AttributesAdded)
	}
	if !contains(result.AttributesAdded, "*.pka binary") {
		t.Fatalf("missing rule was not reported as added: %+v", result.AttributesAdded)
	}
}

func TestInitRejectsNonRepository(t *testing.T) {
	dir := t.TempDir()
	_, err := Init(dir, "/opt/pkt")
	if !errors.Is(err, ErrNotRepository) {
		t.Fatalf("Init error = %v, want ErrNotRepository", err)
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("error is not actionable: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, HookDir)); !os.IsNotExist(statErr) {
		t.Fatalf("non-repository got hooks: %v", statErr)
	}
}

func TestInitResolvesExecutablePath(t *testing.T) {
	dir := initRepo(t)
	result, err := Init(dir, "relative-pkt")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if !filepath.IsAbs(result.Executable) {
		t.Fatalf("Executable = %q, want an absolute path", result.Executable)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	return dir
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

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

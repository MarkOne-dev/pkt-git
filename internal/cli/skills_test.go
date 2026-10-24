package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillsCommand(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	code, stdout, stderr := runCLI("skills")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	for _, agent := range []string{"opencode", "claude", "gemini"} {
		if !strings.Contains(stdout, "[ok] "+agent+": installed") {
			t.Fatalf("stdout does not report %s:\n%s", agent, stdout)
		}
	}
	for _, path := range []string{
		filepath.Join(home, ".config", "opencode", "skills", "pkt-git", "SKILL.md"),
		filepath.Join(home, ".claude", "skills", "pkt-git", "SKILL.md"),
		filepath.Join(home, ".gemini", "skills", "pkt-git", "SKILL.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
}

func TestSkillsRemoveCommand(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if code, _, stderr := runCLI("skills"); code != 0 {
		t.Fatalf("install exit code = %d, stderr=%q", code, stderr)
	}

	code, stdout, stderr := runCLI("skills", "--remove")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	for _, agent := range []string{"opencode", "claude", "gemini"} {
		if !strings.Contains(stdout, "[ok] "+agent+": removed") {
			t.Fatalf("stdout does not report %s removal:\n%s", agent, stdout)
		}
	}
	for _, path := range []string{
		filepath.Join(home, ".config", "opencode", "skills", "pkt-git"),
		filepath.Join(home, ".claude", "skills", "pkt-git"),
		filepath.Join(home, ".gemini", "skills", "pkt-git"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s still exists: %v", path, err)
		}
	}
}

func TestSkillsCommandUsage(t *testing.T) {
	code, _, stderr := runCLI("skills", "extra")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected 0 argument") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestHelpListsSkills(t *testing.T) {
	code, stdout, stderr := runCLI("--help")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "skills") {
		t.Fatalf("help does not list skills:\n%s", stdout)
	}
}

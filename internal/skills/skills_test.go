package skills

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTargets(t *testing.T) {
	home := t.TempDir()
	targets, err := Targets(home)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"opencode": filepath.Join(home, ".config", "opencode", "skills", "pkt-git"),
		"claude":   filepath.Join(home, ".claude", "skills", "pkt-git"),
		"gemini":   filepath.Join(home, ".gemini", "skills", "pkt-git"),
	}
	if len(targets) != len(want) {
		t.Fatalf("targets = %+v", targets)
	}
	for _, target := range targets {
		if want[target.Agent] != target.Dir {
			t.Fatalf("%s dir = %q, want %q", target.Agent, target.Dir, want[target.Agent])
		}
	}
}

func TestInstallAllAgents(t *testing.T) {
	home := t.TempDir()
	other := filepath.Join(home, ".claude", "skills", "other", "SKILL.md")
	writeTestFile(t, other, "other skill")

	results, err := Install(home)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("results = %+v", results)
	}
	for _, result := range results {
		if result.Err != nil {
			t.Fatalf("%s: %v", result.Target.Agent, result.Err)
		}
		if result.Action != "installed" {
			t.Fatalf("%s action = %q, want installed", result.Target.Agent, result.Action)
		}
		path := filepath.Join(result.Target.Dir, "SKILL.md")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: %v", result.Target.Agent, err)
		}
		if string(data) != string(Asset()) {
			t.Fatalf("%s skill content differs from the embedded asset", result.Target.Agent)
		}
		if !strings.Contains(string(data), "name: pkt-git") {
			t.Fatalf("%s skill has no frontmatter", result.Target.Agent)
		}
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("unrelated skill was removed: %v", err)
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	home := t.TempDir()
	if _, err := Install(home); err != nil {
		t.Fatal(err)
	}
	results, err := Install(home)
	if err != nil {
		t.Fatalf("second Install: %v", err)
	}
	for _, result := range results {
		if result.Action != "unchanged" {
			t.Fatalf("%s action = %q, want unchanged", result.Target.Agent, result.Action)
		}
	}
}

func TestRemoveOnlyPktGit(t *testing.T) {
	home := t.TempDir()
	if _, err := Install(home); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(home, ".gemini", "skills", "other", "SKILL.md")
	writeTestFile(t, other, "other skill")

	results, err := Remove(home)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	for _, result := range results {
		if result.Action != "removed" {
			t.Fatalf("%s action = %q, want removed", result.Target.Agent, result.Action)
		}
		if _, err := os.Stat(result.Target.Dir); !os.IsNotExist(err) {
			t.Fatalf("%s skill directory still exists: %v", result.Target.Agent, err)
		}
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("unrelated skill was removed: %v", err)
	}

	results, err = Remove(home)
	if err != nil {
		t.Fatalf("second Remove: %v", err)
	}
	for _, result := range results {
		if result.Action != "absent" {
			t.Fatalf("%s action = %q, want absent", result.Target.Agent, result.Action)
		}
	}
}

func TestMissingAssetPreservesExistingInstall(t *testing.T) {
	home := t.TempDir()
	if _, err := Install(home); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".config", "opencode", "skills", "pkt-git", "SKILL.md")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := InstallContent(home, nil); !errors.Is(err, ErrMissingAsset) {
		t.Fatalf("error = %v, want ErrMissingAsset", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("existing skill was deleted: %v", err)
	}
	if string(before) != string(after) {
		t.Fatal("missing-asset failure changed the existing skill")
	}
}

func TestSkillGuardrailInvariants(t *testing.T) {
	asset := string(Asset())
	requiredInvariants := []string{
		"Operator-only role & zero code modification",
		"Never edit Go, Python, or",
		"Never write ad-hoc scripts",
		"pkt resolve --theirs",
		"pkt resolve --ours",
		"Plain-Language Human Mediation",
		"The text is the source of truth; the binary is an artifact",
		"Never global find-replace or `sed` over snapshot XML",
	}

	for _, req := range requiredInvariants {
		if !strings.Contains(asset, req) {
			t.Errorf("Asset() missing required invariant or directive: %q", req)
		}
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

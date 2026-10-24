// Package skills installs the bundled pkt-git agent skill into the global
// skill directories used by OpenCode, Claude Code, and Gemini CLI.
package skills

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MarkOne-dev/pkt-git/internal/fsx"
)

//go:embed assets/pkt-git/SKILL.md
var skillContent []byte

// ErrMissingAsset reports that the embedded skill is unavailable. Callers must
// not remove an existing installation in that case.
var ErrMissingAsset = errors.New("skills: bundled pkt-git skill is unavailable; reinstall pkt-git")

// Target is one agent's global skill directory for pkt-git.
type Target struct {
	Agent string
	Dir   string
}

// Result reports one agent's outcome.
type Result struct {
	Target Target
	Action string
	Err    error
}

// Asset returns a copy of the embedded skill for inspection.
func Asset() []byte {
	return append([]byte(nil), skillContent...)
}

// Targets returns the supported agents in display order.
func Targets(home string) ([]Target, error) {
	if home == "" {
		return nil, errors.New("skills: home directory is empty")
	}
	return []Target{
		{Agent: "opencode", Dir: filepath.Join(home, ".config", "opencode", "skills", "pkt-git")},
		{Agent: "claude", Dir: filepath.Join(home, ".claude", "skills", "pkt-git")},
		{Agent: "gemini", Dir: filepath.Join(home, ".gemini", "skills", "pkt-git")},
	}, nil
}

// Install writes the embedded skill into every supported agent directory.
func Install(home string) ([]Result, error) {
	return InstallContent(home, skillContent)
}

// InstallContent is Install with an explicit payload, used by tests to model a
// missing bundled asset.
func InstallContent(home string, content []byte) ([]Result, error) {
	if len(content) == 0 {
		return nil, ErrMissingAsset
	}
	targets, err := Targets(home)
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(targets))
	var failures []error
	for _, target := range targets {
		result := Result{Target: target}
		path := filepath.Join(target.Dir, "SKILL.md")
		existing, readErr := os.ReadFile(path)
		switch {
		case readErr == nil && bytes.Equal(existing, content):
			result.Action = "unchanged"
		case readErr != nil && !errors.Is(readErr, os.ErrNotExist):
			result.Err = fmt.Errorf("skills: read %s: %w", path, readErr)
		default:
			if err := fsx.WriteFileAtomic(path, content, 0o644); err != nil {
				result.Err = fmt.Errorf("skills: install %s: %w", path, err)
			} else {
				result.Action = "installed"
			}
		}
		if result.Err != nil {
			failures = append(failures, result.Err)
		}
		results = append(results, result)
	}
	return results, errors.Join(failures...)
}

// Remove deletes only the pkt-git skill directory for each supported agent.
func Remove(home string) ([]Result, error) {
	targets, err := Targets(home)
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(targets))
	var failures []error
	for _, target := range targets {
		result := Result{Target: target}
		_, statErr := os.Lstat(target.Dir)
		switch {
		case errors.Is(statErr, os.ErrNotExist):
			result.Action = "absent"
		case statErr != nil:
			result.Err = fmt.Errorf("skills: inspect %s: %w", target.Dir, statErr)
		default:
			if err := os.RemoveAll(target.Dir); err != nil {
				result.Err = fmt.Errorf("skills: remove %s: %w", target.Dir, err)
			} else {
				result.Action = "removed"
			}
		}
		if result.Err != nil {
			failures = append(failures, result.Err)
		}
		results = append(results, result)
	}
	return results, errors.Join(failures...)
}

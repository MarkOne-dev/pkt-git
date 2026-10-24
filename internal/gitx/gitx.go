// Package gitx wraps the git command line for pkt-git's hooks and resolver.
package gitx

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrConfigMissing reports a Git configuration key that is not set.
var ErrConfigMissing = errors.New("gitx: configuration key is not set")

// Repo runs Git commands in a working directory.
type Repo struct {
	Dir string
}

// New returns a repository wrapper for dir. An empty dir uses the current
// process working directory.
func New(dir string) *Repo {
	return &Repo{Dir: dir}
}

// CommandError reports a failed Git command without hiding its stderr.
type CommandError struct {
	Args     []string
	Stderr   string
	ExitCode int
	Err      error
}

func (e *CommandError) Error() string {
	command := "git " + strings.Join(e.Args, " ")
	if e.Stderr != "" {
		return fmt.Sprintf("%s failed: %s", command, strings.TrimSpace(e.Stderr))
	}
	return fmt.Sprintf("%s failed: %v", command, e.Err)
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

// Run executes git and returns stdout.
func (r *Repo) Run(args ...string) (string, error) {
	out, err := r.run(args...)
	return string(out), err
}

// RunBytes executes git and returns raw stdout, preserving NUL separators.
func (r *Repo) RunBytes(args ...string) ([]byte, error) {
	return r.run(args...)
}

func (r *Repo) run(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		exitCode := 0
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return nil, &CommandError{
			Args:     args,
			Stderr:   stderr.String(),
			ExitCode: exitCode,
			Err:      err,
		}
	}
	return stdout.Bytes(), nil
}

// ConfigGet reads one repository-local Git configuration value.
func (r *Repo) ConfigGet(key string) (string, error) {
	out, err := r.Run("config", "--get", key)
	if err != nil {
		var commandErr *CommandError
		if errors.As(err, &commandErr) && commandErr.ExitCode == 1 {
			return "", fmt.Errorf("%w: %s", ErrConfigMissing, key)
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ConfigSet writes one repository-local Git configuration value.
func (r *Repo) ConfigSet(key, value string) error {
	_, err := r.Run("config", key, value)
	return err
}

// RevParse resolves a revision to a commit hash.
func (r *Repo) RevParse(rev string) (string, error) {
	out, err := r.Run("rev-parse", rev)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Toplevel returns the repository root.
func (r *Repo) Toplevel() (string, error) {
	out, err := r.Run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GitDir returns the repository's Git directory.
func (r *Repo) GitDir() (string, error) {
	out, err := r.Run("rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ChangedFiles lists paths changed between revisions, or against the working
// tree when no revisions are supplied. Output is NUL-separated, so paths with
// spaces are preserved.
func (r *Repo) ChangedFiles(revisions []string, pathspecs ...string) ([]string, error) {
	args := []string{"diff", "--name-only", "-z"}
	args = append(args, revisions...)
	args = appendPathspecs(args, pathspecs)
	out, err := r.RunBytes(args...)
	if err != nil {
		return nil, err
	}
	return splitNUL(out), nil
}

// CachedChangedFiles lists staged paths matching a diff filter.
func (r *Repo) CachedChangedFiles(filter string, pathspecs ...string) ([]string, error) {
	args := []string{"diff", "--cached", "--name-only", "-z", "--diff-filter=" + filter}
	args = appendPathspecs(args, pathspecs)
	out, err := r.RunBytes(args...)
	if err != nil {
		return nil, err
	}
	return splitNUL(out), nil
}

// ConflictedFiles lists unmerged paths.
func (r *Repo) ConflictedFiles() ([]string, error) {
	out, err := r.RunBytes("diff", "--name-only", "--diff-filter=U", "-z")
	if err != nil {
		return nil, err
	}
	return splitNUL(out), nil
}

// DiffTreeRootFiles lists every path in a root commit.
func (r *Repo) DiffTreeRootFiles(revision string, pathspecs ...string) ([]string, error) {
	args := []string{"diff-tree", "--root", "-r", "--name-only", "-z", "--no-commit-id", revision}
	args = appendPathspecs(args, pathspecs)
	out, err := r.RunBytes(args...)
	if err != nil {
		return nil, err
	}
	return splitNUL(out), nil
}

// Checkout applies one side of a conflict to the supplied paths.
func (r *Repo) Checkout(policy string, paths []string) error {
	if policy != "ours" && policy != "theirs" {
		return fmt.Errorf("gitx: unsupported checkout policy %q", policy)
	}
	if len(paths) == 0 {
		return nil
	}
	args := []string{"checkout", "--" + policy, "--"}
	args = append(args, paths...)
	_, err := r.Run(args...)
	return err
}

// TrackedFiles returns the subset of paths that exist in the index.
func (r *Repo) TrackedFiles(paths []string) (map[string]bool, error) {
	if len(paths) == 0 {
		return map[string]bool{}, nil
	}
	args := []string{"ls-files", "-z", "--"}
	args = append(args, paths...)
	out, err := r.RunBytes(args...)
	if err != nil {
		return nil, err
	}
	tracked := make(map[string]bool)
	for _, path := range splitNUL(out) {
		tracked[path] = true
	}
	return tracked, nil
}

// Add stages the supplied paths, including deletions and renames under a
// directory, without touching unrelated repository paths.
func (r *Repo) Add(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	args := []string{"add", "-A", "--"}
	args = append(args, paths...)
	_, err := r.Run(args...)
	return err
}

// NameRev returns a human-readable name for a revision, falling back to the
// revision itself when Git cannot name it.
func (r *Repo) NameRev(revision string) string {
	out, err := r.Run("name-rev", "--name-only", revision)
	if err != nil {
		return revision
	}
	name := strings.TrimSpace(out)
	if name == "" {
		return revision
	}
	return name
}

func appendPathspecs(args []string, pathspecs []string) []string {
	if len(pathspecs) == 0 {
		return args
	}
	args = append(args, "--")
	return append(args, pathspecs...)
}

func splitNUL(out []byte) []string {
	if len(out) == 0 {
		return nil
	}
	parts := bytes.Split(out, []byte{0})
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) > 0 {
			paths = append(paths, string(part))
		}
	}
	return paths
}

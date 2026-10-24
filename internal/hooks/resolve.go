package hooks

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/gitx"
	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

var (
	// ErrNoMerge reports that pkt resolve was run without an active merge.
	ErrNoMerge = errors.New("hooks: no merge in progress; nothing to resolve")
	// ErrReportOnly reports a successful conflict report. The caller should
	// exit non-zero because no resolution policy was selected.
	ErrReportOnly = errors.New("hooks: conflict report only; choose --ours or --theirs")
	// ErrNoConflicts reports a merge state without conflicted paths.
	ErrNoConflicts = errors.New("hooks: merge state has no conflicted files")
)

// ConflictInfo describes one conflicted path.
type ConflictInfo struct {
	Path        string
	Binary      bool
	Blocks      int
	FirstMarker int
}

// ResolveResult describes a report or a completed resolution.
type ResolveResult struct {
	Conflicts []ConflictInfo
	Text      []string
	Binary    []string
	Refreshed []string
	Rebuilt   []string
	Staged    []string
}

// Resolve reports merge conflicts or applies a policy to them. policy must be
// empty (report only), "ours", or "theirs".
func Resolve(dir, policy string, out io.Writer) (ResolveResult, error) {
	if out == nil {
		out = io.Discard
	}
	if policy != "" && policy != "ours" && policy != "theirs" {
		return ResolveResult{}, fmt.Errorf("hooks: unsupported resolve policy %q", policy)
	}

	repo := gitx.New(dir)
	root, err := repo.Toplevel()
	if err != nil {
		return ResolveResult{}, fmt.Errorf("resolve: %w", ErrNotRepository)
	}
	if _, err := repo.RevParse("MERGE_HEAD"); err != nil {
		return ResolveResult{}, ErrNoMerge
	}

	conflicts, err := repo.ConflictedFiles()
	if err != nil {
		return ResolveResult{}, fmt.Errorf("resolve: list conflicts: %w", err)
	}
	if len(conflicts) == 0 {
		return ResolveResult{}, ErrNoConflicts
	}

	result := ResolveResult{}
	for _, conflict := range conflicts {
		info := ConflictInfo{Path: conflict, Binary: isBinaryPath(conflict)}
		if info.Binary {
			result.Binary = append(result.Binary, conflict)
		} else {
			result.Text = append(result.Text, conflict)
			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(conflict)))
			if err != nil {
				return ResolveResult{}, fmt.Errorf("resolve: read %s: %w", conflict, err)
			}
			info.Blocks = strings.Count(string(content), "<<<<<<<")
			info.FirstMarker = firstMarkerLine(string(content))
		}
		result.Conflicts = append(result.Conflicts, info)
	}

	if policy == "" {
		incoming := repo.NameRev("MERGE_HEAD")
		fmt.Fprintf(out, "[pkt-git] conflict! current branch vs '%s':\n", incoming)
		for _, info := range result.Conflicts {
			if info.Binary {
				fmt.Fprintf(out, "  %s  (binary — rebuilt automatically from the snapshot)\n", info.Path)
				continue
			}
			fmt.Fprintf(out, "  %s  (%d conflict block(s))\n", info.Path, info.Blocks)
			if info.FirstMarker > 0 {
				fmt.Fprintf(out, "    first marker at line %d: (ours) vs (theirs from %s)\n", info.FirstMarker, incoming)
			}
		}
		fmt.Fprintln(out, "[pkt-git] fix the <<<<<<< markers (device-scoped, see docs/workflow.md), or run:")
		fmt.Fprintf(out, "    pkt resolve --theirs   # accept incoming (%s)\n", incoming)
		fmt.Fprintln(out, "    pkt resolve --ours     # keep current")
		return result, ErrReportOnly
	}

	if len(result.Text) > 0 {
		if err := repo.Checkout(policy, result.Text); err != nil {
			return ResolveResult{}, fmt.Errorf("resolve: checkout %s: %w", policy, err)
		}
	}
	for _, textPath := range result.Text {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(textPath)))
		if err != nil {
			return ResolveResult{}, fmt.Errorf("resolve: read %s: %w", textPath, err)
		}
		if hasConflictMarker(string(content)) {
			return ResolveResult{}, fmt.Errorf("resolve: %s still has conflict markers", textPath)
		}
	}

	refreshed := make(map[string]bool)
	refreshedFiles := make(map[string]bool)
	removedFiles := make(map[string]bool)
	rebuildFromSnapshot := func(project, xmlPath string) error {
		if refreshed[project] {
			return nil
		}
		xmlAbs := filepath.Join(root, filepath.FromSlash(xmlPath))
		if _, err := os.Stat(xmlAbs); err != nil {
			return fmt.Errorf("resolve: snapshot %s: %w", xmlPath, err)
		}
		syncResult, err := snapshot.Sync(xmlAbs)
		if err != nil {
			return fmt.Errorf("resolve: refresh %s: %w", xmlPath, err)
		}
		refreshed[project] = true
		projectRel := path.Join("snapshot", project)
		result.Refreshed = append(result.Refreshed, projectRel)
		for _, file := range syncResult.Files {
			refreshedFiles[path.Join(projectRel, file)] = true
		}
		for _, file := range syncResult.Removed {
			removedFiles[path.Join(projectRel, "configs", file)] = true
		}
		return nil
	}

	for _, textPath := range result.Text {
		parts := strings.Split(filepath.ToSlash(textPath), "/")
		if len(parts) < 3 || parts[0] != "snapshot" {
			continue
		}
		project := parts[1]
		xmlPath := path.Join("snapshot", project, "base.snapshot.xml")
		if err := rebuildFromSnapshot(project, xmlPath); err != nil {
			return ResolveResult{}, err
		}
	}

	for _, binaryPath := range result.Binary {
		stem := strings.TrimSuffix(filepath.Base(filepath.FromSlash(binaryPath)), filepath.Ext(binaryPath))
		if stem == "" || stem == "." {
			return ResolveResult{}, fmt.Errorf("resolve: invalid binary path %q", binaryPath)
		}
		xmlPath := path.Join("snapshot", stem, "base.snapshot.xml")
		if err := rebuildFromSnapshot(stem, xmlPath); err != nil {
			return ResolveResult{}, err
		}
		xmlAbs := filepath.Join(root, filepath.FromSlash(xmlPath))
		binaryAbs := filepath.Join(root, filepath.FromSlash(binaryPath))
		if _, err := snapshot.Build(xmlAbs, binaryAbs); err != nil {
			return ResolveResult{}, fmt.Errorf("resolve: rebuild %s: %w", binaryPath, err)
		}
		result.Rebuilt = append(result.Rebuilt, binaryPath)
		fmt.Fprintf(out, "[pkt-git] binary '%s' rebuilt from the resolved snapshot.\n", binaryPath)
	}

	stageSet := make(map[string]bool)
	for _, textPath := range result.Text {
		stageSet[filepath.ToSlash(textPath)] = true
	}
	for stagePath := range refreshedFiles {
		stageSet[stagePath] = true
	}
	if len(removedFiles) > 0 {
		paths := make([]string, 0, len(removedFiles))
		for stagePath := range removedFiles {
			paths = append(paths, stagePath)
		}
		tracked, err := repo.TrackedFiles(paths)
		if err != nil {
			return ResolveResult{}, fmt.Errorf("resolve: check removed configs: %w", err)
		}
		for stagePath := range removedFiles {
			if tracked[stagePath] {
				stageSet[stagePath] = true
			}
		}
	}
	for _, binaryPath := range result.Rebuilt {
		stageSet[filepath.ToSlash(binaryPath)] = true
	}
	stagePaths := make([]string, 0, len(stageSet))
	for stagePath := range stageSet {
		stagePaths = append(stagePaths, stagePath)
	}
	sort.Strings(stagePaths)
	if err := repo.Add(stagePaths); err != nil {
		return ResolveResult{}, fmt.Errorf("resolve: stage resolved paths: %w", err)
	}
	result.Staged = stagePaths
	fmt.Fprintln(out, "[pkt-git] resolution staged — finish with: git commit")
	return result, nil
}

func isBinaryPath(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".pkt" || ext == ".pka"
}

func firstMarkerLine(content string) int {
	for index, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "<<<<<<<") {
			return index + 1
		}
	}
	return 0
}

func hasConflictMarker(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "<<<<<<<") ||
			strings.HasPrefix(line, "=======") ||
			strings.HasPrefix(line, ">>>>>>>") {
			return true
		}
	}
	return false
}

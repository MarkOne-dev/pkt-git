package hooks

import (
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/gitx"
	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

// PreCommitResult summarizes a pre-commit run.
type PreCommitResult struct {
	Exported          []string
	TextOnlySnapshots []string
}

// PreCommit exports every staged .pkt/.pka into snapshot/<project>/ and stages
// the resulting snapshot directories. It stages nothing if any export fails.
// A text-only snapshot change is allowed but reported as a stale binary.
func PreCommit(dir string, out io.Writer) (PreCommitResult, error) {
	if out == nil {
		out = io.Discard
	}
	repo := gitx.New(dir)
	root, err := repo.Toplevel()
	if err != nil {
		return PreCommitResult{}, fmt.Errorf("pre-commit: %w", ErrNotRepository)
	}

	staged, err := repo.CachedChangedFiles("ACMR", "*.pkt", "*.pka")
	if err != nil {
		return PreCommitResult{}, fmt.Errorf("pre-commit: list staged binaries: %w", err)
	}

	result := PreCommitResult{}
	stagedProjects := make(map[string]bool, len(staged))
	stagePaths := make([]string, 0, len(staged)*3)
	removedPaths := make([]string, 0)
	seen := make(map[string]string, len(staged))
	for _, binary := range staged {
		rel := filepath.FromSlash(binary)
		stem := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
		if stem == "" || stem == "." {
			return PreCommitResult{}, fmt.Errorf("pre-commit: invalid binary path %q", binary)
		}
		key := strings.ToLower(stem)
		if previous, ok := seen[key]; ok {
			return PreCommitResult{}, fmt.Errorf(
				"pre-commit: %s and %s both map to snapshot/%s", previous, binary, stem)
		}
		seen[key] = binary
		stagedProjects[key] = true

		snapshotRel := path.Join("snapshot", stem)
		snapshotAbs := filepath.Join(root, filepath.FromSlash(snapshotRel))
		binaryAbs := filepath.Join(root, rel)
		fmt.Fprintf(out, "[pre-commit] binary changed: %s -> exporting snapshot to %s\n",
			binary, snapshotRel)
		exportResult, err := snapshot.Export(binaryAbs, snapshotAbs)
		if err != nil {
			return PreCommitResult{}, fmt.Errorf("pre-commit: export %s: %w", binary, err)
		}
		result.Exported = append(result.Exported, snapshotRel)
		removed := make(map[string]bool, len(exportResult.Removed))
		for _, name := range exportResult.Removed {
			removed[path.Join(snapshotRel, "configs", name)] = true
		}
		for _, file := range exportResult.Files {
			stagePath := path.Join(snapshotRel, file)
			if removed[stagePath] {
				removedPaths = append(removedPaths, stagePath)
				continue
			}
			stagePaths = append(stagePaths, stagePath)
		}
	}

	if len(removedPaths) > 0 {
		tracked, err := repo.TrackedFiles(removedPaths)
		if err != nil {
			return PreCommitResult{}, fmt.Errorf("pre-commit: check removed configs: %w", err)
		}
		for _, stagePath := range removedPaths {
			if tracked[stagePath] {
				stagePaths = append(stagePaths, stagePath)
			}
		}
	}

	textOnly, err := repo.CachedChangedFiles("ACMR", "snapshot/*/base.snapshot.xml")
	if err != nil {
		return PreCommitResult{}, fmt.Errorf("pre-commit: list staged snapshots: %w", err)
	}
	for _, xml := range textOnly {
		project := path.Base(path.Dir(filepath.ToSlash(xml)))
		if stagedProjects[strings.ToLower(project)] {
			continue
		}
		result.TextOnlySnapshots = append(result.TextOnlySnapshots, xml)
	}

	if len(stagePaths) > 0 {
		if err := repo.Add(stagePaths); err != nil {
			return PreCommitResult{}, fmt.Errorf("pre-commit: stage exported snapshots: %w", err)
		}
		fmt.Fprintln(out, "[pre-commit] textual snapshot included automatically.")
	}
	if len(result.TextOnlySnapshots) > 0 {
		fmt.Fprintln(out, "[pre-commit] Warning: textual snapshot changed without its .pkt.")
		fmt.Fprintln(out, "  Valid (textual edit/merge); rebuild the binary before opening PT:")
		for _, xml := range result.TextOnlySnapshots {
			project := path.Base(path.Dir(filepath.ToSlash(xml)))
			fmt.Fprintf(out, "    pkt build %s topologia/%s.pkt\n", filepath.ToSlash(xml), project)
		}
	}
	return result, nil
}

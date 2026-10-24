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
	if len(staged) == 0 {
		textOnly, err := repo.CachedChangedFiles("ACMR", "snapshot/*/base.snapshot.xml")
		if err != nil {
			return PreCommitResult{}, fmt.Errorf("pre-commit: list staged snapshots: %w", err)
		}
		if len(textOnly) == 0 {
			return result, nil
		}
		result.TextOnlySnapshots = textOnly
		fmt.Fprintln(out, "[pre-commit] Warning: textual snapshot changed without its .pkt.")
		fmt.Fprintln(out, "  Valid (textual edit/merge); rebuild the binary before opening PT:")
		for _, xml := range textOnly {
			project := path.Base(path.Dir(filepath.ToSlash(xml)))
			fmt.Fprintf(out, "    pkt build %s topologia/%s.pkt\n", filepath.ToSlash(xml), project)
		}
		return result, nil
	}

	seen := make(map[string]string, len(staged))
	dirs := make([]string, 0, len(staged))
	for _, binary := range staged {
		rel := filepath.FromSlash(binary)
		stem := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
		if stem == "" || stem == "." {
			return PreCommitResult{}, fmt.Errorf("pre-commit: invalid binary path %q", binary)
		}
		if previous, ok := seen[stem]; ok {
			return PreCommitResult{}, fmt.Errorf(
				"pre-commit: %s and %s both map to snapshot/%s", previous, binary, stem)
		}
		seen[stem] = binary

		snapshotRel := filepath.Join("snapshot", stem)
		snapshotAbs := filepath.Join(root, snapshotRel)
		binaryAbs := filepath.Join(root, rel)
		fmt.Fprintf(out, "[pre-commit] binary changed: %s -> exporting snapshot to %s\n",
			binary, filepath.ToSlash(snapshotRel))
		if _, err := snapshot.Export(binaryAbs, snapshotAbs); err != nil {
			return PreCommitResult{}, fmt.Errorf("pre-commit: export %s: %w", binary, err)
		}
		result.Exported = append(result.Exported, filepath.ToSlash(snapshotRel))
		dirs = append(dirs, filepath.ToSlash(snapshotRel))
	}

	if err := repo.Add(dirs); err != nil {
		return PreCommitResult{}, fmt.Errorf("pre-commit: stage exported snapshots: %w", err)
	}
	fmt.Fprintln(out, "[pre-commit] textual snapshot included automatically.")
	return result, nil
}

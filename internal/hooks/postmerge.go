package hooks

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/gitx"
	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

// ErrRebuildFailed reports one or more binary rebuild failures after a
// history change. The hook still completes; callers should exit non-zero.
var ErrRebuildFailed = errors.New("hooks: one or more binaries could not be rebuilt")

// RebuildResult lists the binaries rebuilt and the snapshots that failed.
type RebuildResult struct {
	Rebuilt []string
	Failed  []string
}

// PostMerge rebuilds binaries for snapshots changed by a merge or pull. Git
// passes the squash flag as the only argument; it is never treated as a
// revision. A squash merge has no new HEAD commit, so its changes are read
// from the index; a normal merge uses ORIG_HEAD..HEAD, falling back to
// HEAD^..HEAD and then to the root commit.
func PostMerge(dir string, args []string, out io.Writer) (RebuildResult, error) {
	if out == nil {
		out = io.Discard
	}
	if len(args) > 1 {
		return RebuildResult{}, fmt.Errorf("hooks: post-merge expects at most one squash flag")
	}
	squash := len(args) == 1 && args[0] == "1"
	if len(args) == 1 && args[0] != "0" && args[0] != "1" {
		return RebuildResult{}, fmt.Errorf("hooks: post-merge expected squash flag 0 or 1, got %q", args[0])
	}

	repo := gitx.New(dir)
	root, err := repo.Toplevel()
	if err != nil {
		return RebuildResult{}, fmt.Errorf("post-merge: %w", ErrNotRepository)
	}

	var changed []string
	if squash {
		changed, err = repo.CachedChangedFiles("ACMR", "snapshot/*/base.snapshot.xml")
	} else {
		changed, err = mergeChangedSnapshots(repo)
	}
	if err != nil {
		return RebuildResult{}, fmt.Errorf("post-merge: list changed snapshots: %w", err)
	}
	return rebuildSnapshots(root, changed, out)
}

// PostCheckout rebuilds binaries for snapshots changed by a branch checkout.
// Git passes old revision, new revision, and a branch flag; only branch
// checkouts (flag 1) trigger a rebuild.
func PostCheckout(dir string, args []string, out io.Writer) (RebuildResult, error) {
	if out == nil {
		out = io.Discard
	}
	if len(args) != 3 {
		return RebuildResult{}, fmt.Errorf("hooks: post-checkout expects old, new, and branch flag")
	}
	oldRev, newRev, branchFlag := args[0], args[1], args[2]
	if branchFlag != "1" {
		return RebuildResult{}, nil
	}
	if isZeroRevision(newRev) {
		return RebuildResult{}, fmt.Errorf("hooks: post-checkout received a zero new revision")
	}

	repo := gitx.New(dir)
	root, err := repo.Toplevel()
	if err != nil {
		return RebuildResult{}, fmt.Errorf("post-checkout: %w", ErrNotRepository)
	}

	var changed []string
	if isZeroRevision(oldRev) {
		changed, err = repo.DiffTreeRootFiles(newRev, "snapshot/*/base.snapshot.xml")
	} else {
		changed, err = repo.ChangedFiles([]string{oldRev, newRev}, "snapshot/*/base.snapshot.xml")
	}
	if err != nil {
		return RebuildResult{}, fmt.Errorf("post-checkout: list changed snapshots: %w", err)
	}
	return rebuildSnapshots(root, changed, out)
}

func mergeChangedSnapshots(repo *gitx.Repo) ([]string, error) {
	if _, err := repo.RevParse("ORIG_HEAD"); err == nil {
		return repo.ChangedFiles([]string{"ORIG_HEAD", "HEAD"}, "snapshot/*/base.snapshot.xml")
	}
	if _, err := repo.RevParse("HEAD^"); err == nil {
		return repo.ChangedFiles([]string{"HEAD^", "HEAD"}, "snapshot/*/base.snapshot.xml")
	}
	return repo.DiffTreeRootFiles("HEAD", "snapshot/*/base.snapshot.xml")
}

func rebuildSnapshots(root string, xmlPaths []string, out io.Writer) (RebuildResult, error) {
	result := RebuildResult{}
	for _, xmlPath := range xmlPaths {
		rel := filepath.FromSlash(xmlPath)
		abs := filepath.Join(root, rel)
		if _, err := os.Stat(abs); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			result.Failed = append(result.Failed, xmlPath)
			fmt.Fprintf(out, "[pkt-git] ERROR reading %s: %v\n", filepath.ToSlash(xmlPath), err)
			continue
		}

		project := path.Base(path.Dir(filepath.ToSlash(xmlPath)))
		if project == "" || project == "." || project == "/" {
			result.Failed = append(result.Failed, xmlPath)
			fmt.Fprintf(out, "[pkt-git] ERROR: cannot derive a project name from %s\n", filepath.ToSlash(xmlPath))
			continue
		}
		pktRel := path.Join("topologia", project+".pkt")
		pktAbs := filepath.Join(root, filepath.FromSlash(pktRel))
		fmt.Fprintf(out, "[pkt-git] snapshot of '%s' updated; rebuilding %s ...\n", project, pktRel)
		if _, err := snapshot.Build(abs, pktAbs); err != nil {
			result.Failed = append(result.Failed, pktRel)
			fmt.Fprintf(out, "[pkt-git] ERROR rebuilding %s: %v\n", pktRel, err)
			continue
		}
		fmt.Fprintf(out, "[pkt-git] '%s' rebuilt; open it in Packet Tracer.\n", pktRel)
		result.Rebuilt = append(result.Rebuilt, pktRel)
	}

	if len(result.Failed) > 0 {
		return result, fmt.Errorf("%w: %s", ErrRebuildFailed, strings.Join(result.Failed, ", "))
	}
	return result, nil
}

func isZeroRevision(rev string) bool {
	if rev == "" {
		return false
	}
	return strings.Trim(rev, "0") == ""
}

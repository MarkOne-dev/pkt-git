package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/hooks"
	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [dir]",
		Short: "Wire pkt-git hooks into a Git repository",
		Args:  maxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			result, err := hooks.Init(dir, "")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			displayRoot := ui.TruncatePath(filepath.ToSlash(result.Root), 48)
			ui.Success(out, "initialized %s", displayRoot)
			fmt.Fprintf(out, "  hooks: %s (%s)\n", hooks.HookDir, strings.Join(result.HooksWritten, ", "))
			fmt.Fprintf(out, "  binary: %s\n", result.Executable)
			switch {
			case result.AttributesCreated:
				fmt.Fprintln(out, "  attributes: created .gitattributes")
			case len(result.AttributesAdded) > 0:
				fmt.Fprintf(out, "  attributes: added %s\n", strings.Join(result.AttributesAdded, ", "))
			default:
				fmt.Fprintln(out, "  attributes: unchanged")
			}

			topos, err := snapshot.FindTopologies(result.Root)
			if err != nil {
				return err
			}
			for _, t := range topos {
				fullPktPath := filepath.Join(result.Root, t)
				destRel := snapshot.DefaultSnapshotDir(t)
				destDir := filepath.Join(result.Root, destRel)
				expResult, expErr := snapshot.Export(fullPktPath, destDir)
				if expErr != nil {
					ui.Warning(out, "failed to auto-export %s: %v", t, expErr)
					continue
				}
				ui.Success(out, "auto-detected %s -> exported to %s", filepath.ToSlash(t), filepath.ToSlash(destRel))
				fmt.Fprintf(out, "  XML: %d bytes (%s), devices: %d, links: %d, configs: %d\n",
					expResult.XMLSize, ui.FormatBytes(int64(expResult.XMLSize)), expResult.DeviceCount, expResult.LinkCount, expResult.ConfigCount)
			}

			return nil
		},
	}
}

package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

func newExportCmd() *cobra.Command {
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "export [project.pkt] [output_dir]",
		Short: "Export a Packet Tracer binary to a textual snapshot",
		Args:  maxArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			var targets [][2]string

			switch len(args) {
			case 2:
				targets = append(targets, [2]string{args[0], args[1]})
			case 1:
				dest := snapshot.DefaultSnapshotDir(args[0])
				targets = append(targets, [2]string{args[0], dest})
			case 0:
				topos, err := snapshot.FindTopologies(".")
				if err != nil {
					return err
				}
				if len(topos) == 0 {
					return usagef("export: no .pkt or .pka files found in repository; specify a file")
				}
				if len(topos) > 1 && !allFlag {
					return usagef("export: multiple topologies found (%s); specify one or pass --all", strings.Join(topos, ", "))
				}
				for _, t := range topos {
					dest := snapshot.DefaultSnapshotDir(t)
					targets = append(targets, [2]string{t, dest})
				}
			}

			for _, target := range targets {
				src, dest := target[0], target[1]
				result, err := snapshot.Export(src, dest)
				if err != nil {
					return err
				}
				ui.Success(out, "exported %s -> %s", filepath.ToSlash(src), filepath.ToSlash(dest))
				fmt.Fprintf(out, "  XML: %d bytes (%s), devices: %d, links: %d, configs: %d\n",
					result.XMLSize, ui.FormatBytes(int64(result.XMLSize)), result.DeviceCount, result.LinkCount, result.ConfigCount)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&allFlag, "all", false, "export all discovered topologies when multiple are present")
	return cmd
}

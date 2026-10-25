package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

func newExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <project.pkt> <output_dir>",
		Short: "Export a Packet Tracer binary to a textual snapshot",
		Args:  exactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := snapshot.Export(args[0], args[1])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			ui.Success(out, "exported %s -> %s", args[0], args[1])
			fmt.Fprintf(out, "  XML: %d bytes (%s), devices: %d, links: %d, configs: %d\n",
				result.XMLSize, ui.FormatBytes(int64(result.XMLSize)), result.DeviceCount, result.LinkCount, result.ConfigCount)
			return nil
		},
	}
}

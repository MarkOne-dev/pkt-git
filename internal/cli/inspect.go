package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

func newInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <project.pkt>",
		Short: "Show a read-only topology summary for a binary",
		Args:  exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := snapshot.Inspect(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s: %d device(s), %d link(s)\n", result.Name, len(result.Devices), result.Links)
			writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(writer, "NAME\tTYPE\tMODEL\tCONFIG LINES")
			for _, device := range result.Devices {
				fmt.Fprintf(writer, "%s\t%s\t%s\t%d\n", device.Name, device.Type, device.Model, device.ConfigLines)
			}
			return writer.Flush()
		},
	}
}

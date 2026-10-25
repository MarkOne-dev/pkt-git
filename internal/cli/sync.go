package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync <snapshot-dir|snapshot.xml>",
		Short: "Refresh configs and metadata from a snapshot XML",
		Args:  exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := snapshot.Sync(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			ui.Success(out, "refreshed %s", args[0])
			fmt.Fprintf(out, "  XML: %d bytes (%s), devices: %d, links: %d, configs: %d\n",
				result.XMLSize, ui.FormatBytes(int64(result.XMLSize)), result.DeviceCount, result.LinkCount, result.ConfigCount)
			if len(result.Removed) > 0 {
				fmt.Fprintf(out, "  removed stale configs: %s\n", strings.Join(result.Removed, ", "))
			}
			return nil
		},
	}
}

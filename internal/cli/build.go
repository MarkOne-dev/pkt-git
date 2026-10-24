package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

func newBuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build <snapshot.xml> <output.pkt>",
		Short: "Build a Packet Tracer binary from a textual snapshot",
		Args:  exactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := snapshot.Build(args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[ok] built %s (%d bytes)\n", args[1], result.Size)
			return nil
		},
	}
}

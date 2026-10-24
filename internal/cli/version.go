package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/config"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the pkt-git version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), config.Version)
		},
	}
}

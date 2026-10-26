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
		Args:  exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), versionString())
			return nil
		},
	}
}

func versionString() string {
	return config.Version
}

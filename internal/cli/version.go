package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/config"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the pkt-git version and build metadata",
		Args:  exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), versionString())
			return nil
		},
	}
}

func versionString() string {
	version := config.Version
	var metadata []string
	if config.Commit != "" && config.Commit != "none" {
		metadata = append(metadata, "commit "+config.Commit)
	}
	if config.Date != "" && config.Date != "unknown" {
		metadata = append(metadata, "built "+config.Date)
	}
	if len(metadata) == 0 {
		return version
	}
	return fmt.Sprintf("%s (%s)", version, strings.Join(metadata, ", "))
}

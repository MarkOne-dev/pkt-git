// Package cli wires the pkt-git command tree.
package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the root command and all subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "pkt",
		Short:         "Git-native version control for Cisco Packet Tracer projects",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newVersionCmd(),
	)

	return root
}

// Execute runs the CLI and returns a process exit code.
func Execute() int {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "pkt:", err)
		if errors.Is(err, errUsage) {
			return 2
		}
		return 1
	}
	return 0
}

// Package cli wires the pkt-git command tree.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the root command and all subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "pkt",
		Short:         "Git-native version control for Cisco Packet Tracer projects",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          rootArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = cmd.Help()
			return &usageError{}
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return usagef("%v", err)
	})

	root.AddCommand(
		newExportCmd(),
		newBuildCmd(),
		newInspectCmd(),
		newVersionCmd(),
	)

	return root
}

// Execute runs the CLI with the process arguments and returns an exit code.
func Execute() int {
	return ExecuteArgs(os.Args[1:], os.Stdout, os.Stderr)
}

// ExecuteArgs runs the CLI with explicit arguments and streams. It maps usage
// failures to exit code 2, runtime failures to 1, and success to 0.
func ExecuteArgs(args []string, out, errOut io.Writer) int {
	if out == nil {
		out = io.Discard
	}
	if errOut == nil {
		errOut = io.Discard
	}

	root := NewRootCmd()
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(errOut)

	err := root.Execute()
	if err == nil {
		return 0
	}

	var usage *usageError
	if errors.As(err, &usage) {
		if usage.message != "" {
			fmt.Fprintln(errOut, "pkt:", usage.message)
		}
		return 2
	}
	if strings.HasPrefix(err.Error(), "unknown command") {
		fmt.Fprintln(errOut, "pkt:", err)
		return 2
	}
	fmt.Fprintln(errOut, "pkt:", err)
	return 1
}

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/hooks"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [dir]",
		Short: "Wire pkt-git hooks into a Git repository",
		Args:  maxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			result, err := hooks.Init(dir, "")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "[ok] initialized %s\n", result.Root)
			fmt.Fprintf(out, "  hooks: %s (%s)\n", hooks.HookDir, strings.Join(result.HooksWritten, ", "))
			fmt.Fprintf(out, "  binary: %s\n", result.Executable)
			switch {
			case result.AttributesCreated:
				fmt.Fprintln(out, "  attributes: created .gitattributes")
			case len(result.AttributesAdded) > 0:
				fmt.Fprintf(out, "  attributes: added %s\n", strings.Join(result.AttributesAdded, ", "))
			default:
				fmt.Fprintln(out, "  attributes: unchanged")
			}
			return nil
		},
	}
}

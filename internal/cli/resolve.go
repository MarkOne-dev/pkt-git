package cli

import (
	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/hooks"
)

func newResolveCmd() *cobra.Command {
	var ours, theirs bool
	cmd := &cobra.Command{
		Use:   "resolve [--ours|--theirs]",
		Short: "Report or resolve snapshot merge conflicts",
		Args:  exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if ours && theirs {
				return usagef("resolve: --ours and --theirs are mutually exclusive")
			}
			policy := ""
			if ours {
				policy = "ours"
			} else if theirs {
				policy = "theirs"
			}
			_, err := hooks.Resolve(".", policy, cmd.OutOrStdout())
			return err
		},
	}
	cmd.Flags().BoolVar(&ours, "ours", false, "accept the current branch's side")
	cmd.Flags().BoolVar(&theirs, "theirs", false, "accept the incoming branch's side")
	return cmd
}

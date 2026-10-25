package cli

import (
	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/hooks"
)

func newPreCommitCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "pre-commit",
		Short:  "Internal pre-commit hook entry point",
		Hidden: true,
		Args:   exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := hooks.PreCommit(".", cmd.OutOrStdout())
			return err
		},
	}
}

func newPostMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "post-merge [squash-flag]",
		Short:  "Internal post-merge hook entry point",
		Hidden: true,
		Args:   maxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && args[0] != "0" && args[0] != "1" {
				return usagef("post-merge: expected squash flag 0 or 1, got %q", args[0])
			}
			_, err := hooks.PostMerge(".", args, cmd.OutOrStdout(), cmd.ErrOrStderr())
			return err
		},
	}
}

func newPostCheckoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "post-checkout <old> <new> <branch-flag>",
		Short:  "Internal post-checkout hook entry point",
		Hidden: true,
		Args:   exactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[2] != "0" && args[2] != "1" {
				return usagef("post-checkout: expected branch flag 0 or 1, got %q", args[2])
			}
			_, err := hooks.PostCheckout(".", args, cmd.OutOrStdout(), cmd.ErrOrStderr())
			return err
		},
	}
}

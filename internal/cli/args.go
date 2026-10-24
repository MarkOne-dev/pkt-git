package cli

import "github.com/spf13/cobra"

func exactArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return usagef("%s: expected %d argument(s), got %d", cmd.CommandPath(), n, len(args))
		}
		return nil
	}
}

func rootArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usagef("unknown command %q for %q", args[0], cmd.CommandPath())
	}
	return nil
}

func minArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return usagef("%s: expected at least %d argument(s), got %d", cmd.CommandPath(), n, len(args))
		}
		return nil
	}
}

func maxArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > n {
			return usagef("%s: expected at most %d argument(s), got %d", cmd.CommandPath(), n, len(args))
		}
		return nil
	}
}

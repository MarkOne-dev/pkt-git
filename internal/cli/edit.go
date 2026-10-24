package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/editor"
)

func newEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <xml> <show|setiface|chgline|insertafter> ...",
		Short: "Inspect or edit one device's running config",
		Long: "Device-scoped edits on a snapshot XML. The legacy order\n" +
			"`pkt edit <xml> <command> ...` and the subcommand order\n" +
			"`pkt edit <command> <xml> ...` are both accepted.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return usagef("edit: expected <xml> <command> ...")
			}
			xmlPath, command, rest := args[0], args[1], args[2:]
			switch command {
			case "show":
				return runEditShow(cmd, xmlPath, rest)
			case "setiface":
				return runEditSetiface(cmd, xmlPath, rest)
			case "chgline":
				return runEditChgline(cmd, xmlPath, rest)
			case "insertafter":
				return runEditInsertAfter(cmd, xmlPath, rest)
			default:
				return usagef("edit: unknown command %q", command)
			}
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "show <xml> <DEVICE>",
			Short: "Print one device's redacted running config",
			Args:  exactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runEditShow(cmd, args[0], args[1:])
			},
		},
		&cobra.Command{
			Use:   "setiface <xml> <DEVICE> <IFACE> <line>...",
			Short: "Replace an interface block",
			Args:  minArgs(4),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runEditSetiface(cmd, args[0], args[1:])
			},
		},
		&cobra.Command{
			Use:   "chgline <xml> <DEVICE> <IFACE> <old_line> <new_line>",
			Short: "Replace one exact line inside an interface block",
			Args:  exactArgs(5),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runEditChgline(cmd, args[0], args[1:])
			},
		},
		&cobra.Command{
			Use:   "insertafter <xml> <DEVICE> <phrase> <line>...",
			Short: "Insert lines after the first config line containing a phrase",
			Args:  minArgs(4),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runEditInsertAfter(cmd, args[0], args[1:])
			},
		},
	)
	return cmd
}

func runEditShow(cmd *cobra.Command, xmlPath string, args []string) error {
	if err := exactArgs(1)(cmd, args); err != nil {
		return err
	}
	return editor.Show(xmlPath, args[0], cmd.OutOrStdout())
}

func runEditSetiface(cmd *cobra.Command, xmlPath string, args []string) error {
	if err := minArgs(3)(cmd, args); err != nil {
		return err
	}
	if err := editor.SetInterface(xmlPath, args[0], args[1], args[2:]); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[ok] updated %s\n", xmlPath)
	return nil
}

func runEditChgline(cmd *cobra.Command, xmlPath string, args []string) error {
	if err := exactArgs(4)(cmd, args); err != nil {
		return err
	}
	if err := editor.ChangeLine(xmlPath, args[0], args[1], args[2], args[3]); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[ok] updated %s\n", xmlPath)
	return nil
}

func runEditInsertAfter(cmd *cobra.Command, xmlPath string, args []string) error {
	if err := minArgs(3)(cmd, args); err != nil {
		return err
	}
	if err := editor.InsertAfter(xmlPath, args[0], args[1], args[2:]); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[ok] updated %s\n", xmlPath)
	return nil
}

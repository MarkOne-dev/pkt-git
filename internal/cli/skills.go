package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/skills"
)

func newSkillsCmd() *cobra.Command {
	var remove bool
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Install or remove the bundled agent skill",
		Args:  exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("skills: locate home directory: %w", err)
			}
			var results []skills.Result
			if remove {
				results, err = skills.Remove(home)
			} else {
				results, err = skills.Install(home)
			}
			out := cmd.OutOrStdout()
			for _, result := range results {
				if result.Err != nil {
					fmt.Fprintf(out, "[error] %s: %v\n", result.Target.Agent, result.Err)
					continue
				}
				fmt.Fprintf(out, "[ok] %s: %s %s\n", result.Target.Agent, result.Action, result.Target.Dir)
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&remove, "remove", false, "remove the pkt-git skill from every supported agent")
	return cmd
}

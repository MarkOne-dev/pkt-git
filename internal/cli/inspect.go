package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

func newInspectCmd() *cobra.Command {
	var typeFilter string

	cmd := &cobra.Command{
		Use:   "inspect <project.pkt>",
		Short: "Show a read-only topology summary for a binary",
		Args:  exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := snapshot.Inspect(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			var filtered []snapshot.InspectDevice
			typeCounts := make(map[string]int)
			for _, d := range result.Devices {
				typeCounts[d.Type]++
				if typeFilter == "" || strings.EqualFold(d.Type, typeFilter) || strings.Contains(strings.ToLower(d.Type), strings.ToLower(typeFilter)) {
					filtered = append(filtered, d)
				}
			}

			if typeFilter != "" {
				fmt.Fprintf(out, "%s: %d device(s) matching %q (of %d total), %d link(s)\n",
					result.Name, len(filtered), typeFilter, len(result.Devices), result.Links)
			} else {
				fmt.Fprintf(out, "%s: %d device(s), %d link(s)\n", result.Name, len(result.Devices), result.Links)
			}

			headers := []string{"NAME", "TYPE", "MODEL", "CONFIG LINES"}
			rows := make([][]string, 0, len(filtered))
			for _, device := range filtered {
				rows = append(rows, []string{
					device.Name,
					ui.DeviceBadge(out, device.Type),
					device.Model,
					strconv.Itoa(device.ConfigLines),
				})
			}

			fmt.Fprintln(out, ui.RenderTable(out, headers, rows))

			// Category Rollup
			var breakdown []string
			for _, cat := range []string{"Router", "Switch", "MultiLayerSwitch", "Server", "AccessPoint", "Pc", "Laptop"} {
				if count := typeCounts[cat]; count > 0 {
					breakdown = append(breakdown, fmt.Sprintf("%s: %d", cat, count))
				}
			}
			if len(breakdown) > 0 {
				fmt.Fprintf(out, "Breakdown: %s\n", strings.Join(breakdown, ", "))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&typeFilter, "type", "t", "", "Filter devices by type (e.g. Router, Switch, Server)")
	return cmd
}

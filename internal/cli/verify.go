package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/codec"
)

func newVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <file.pkt>",
		Short: "Verify the container round trip for a binary",
		Args:  exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("verify: read %s: %w", args[0], err)
			}
			result, err := codec.Verify(raw)
			if err != nil {
				return fmt.Errorf("verify: %s: %w; check the producing Packet Tracer version (tested: %s)", args[0], err, codec.SupportedVersion)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "binary size        : %d (%s)\n", result.BinarySize, ui.FormatBytes(int64(result.BinarySize)))
			fmt.Fprintf(out, "extracted XML      : %d (%s)\n", result.XMLSize, ui.FormatBytes(int64(result.XMLSize)))
			fmt.Fprintf(out, "rebuilt binary     : %d (%s)\n", result.RebuiltSize, ui.FormatBytes(int64(result.RebuiltSize)))
			fmt.Fprintf(out, "XML is a document  : %t\n", result.XMLIsDocument)
			fmt.Fprintf(out, "XML round trip     : %t\n", result.XMLRoundTrip)
			fmt.Fprintf(out, "byte-identical     : %t\n", result.ByteIdentical)
			if !result.XMLIsDocument || !result.XMLRoundTrip {
				return errors.New("verify: XML document or round trip check failed")
			}
			ui.Success(out, "container round trip verified")
			return nil
		},
	}
}

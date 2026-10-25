package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

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
			fmt.Fprintf(out, "binary size        : %d\n", result.BinarySize)
			fmt.Fprintf(out, "extracted XML      : %d\n", result.XMLSize)
			fmt.Fprintf(out, "rebuilt binary     : %d\n", result.RebuiltSize)
			fmt.Fprintf(out, "XML is a document  : %t\n", result.XMLIsDocument)
			fmt.Fprintf(out, "XML round trip     : %t\n", result.XMLRoundTrip)
			fmt.Fprintf(out, "byte-identical     : %t\n", result.ByteIdentical)
			if !result.XMLIsDocument || !result.XMLRoundTrip {
				return errors.New("verify: XML document or round trip check failed")
			}
			return nil
		},
	}
}

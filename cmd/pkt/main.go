// Command pkt is the pkt-git CLI entrypoint.
package main

import (
	"os"

	"github.com/MarkOne-dev/pkt-git/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}

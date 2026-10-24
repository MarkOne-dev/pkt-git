package editor

import (
	"fmt"
	"os"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
	"github.com/MarkOne-dev/pkt-git/internal/fsx"
	"github.com/MarkOne-dev/pkt-git/internal/snapshot"
)

// Transform computes new XML bytes from the current bytes.
type Transform func(data []byte) ([]byte, error)

// Apply validates a transformed document, writes it atomically, and refreshes
// the derived snapshot files. Validation failures leave the original intact.
func Apply(xmlPath string, transform Transform) error {
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return fmt.Errorf("editor: read %s: %w", xmlPath, err)
	}

	updated, err := transform(data)
	if err != nil {
		return err
	}
	if err := codec.ValidateXML(updated); err != nil {
		return fmt.Errorf("editor: resulting XML is invalid: %w", err)
	}
	if err := fsx.WriteFileAtomic(xmlPath, updated, 0); err != nil {
		return err
	}

	if _, err := snapshot.Sync(xmlPath); err != nil {
		return fmt.Errorf("editor: snapshot was changed but derived files may be stale: %w", err)
	}
	return nil
}

// SetInterface replaces an interface block and refreshes derived files.
func SetInterface(xmlPath, deviceName, iface string, lines []string) error {
	return Apply(xmlPath, func(data []byte) ([]byte, error) {
		return SetInterfaceData(data, deviceName, iface, lines)
	})
}

// ChangeLine replaces one exact config line and refreshes derived files.
func ChangeLine(xmlPath, deviceName, iface, oldLine, newLine string) error {
	return Apply(xmlPath, func(data []byte) ([]byte, error) {
		return ChangeLineData(data, deviceName, iface, oldLine, newLine)
	})
}

// InsertAfter inserts lines after the first matching config line and refreshes
// derived files.
func InsertAfter(xmlPath, deviceName, phrase string, lines []string) error {
	return Apply(xmlPath, func(data []byte) ([]byte, error) {
		return InsertAfterData(data, deviceName, phrase, lines)
	})
}

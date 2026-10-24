// Package editor implements read-only and mutating operations on snapshot XML.
package editor

import (
	"fmt"
	"io"
	"os"

	"github.com/MarkOne-dev/pkt-git/internal/security"
	"github.com/MarkOne-dev/pkt-git/internal/xmledit"
)

// Show prints one device's running config with credential values redacted.
// It never writes to the snapshot.
func Show(xmlPath, deviceName string, out io.Writer) error {
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return fmt.Errorf("editor: read %s: %w", xmlPath, err)
	}
	lines, err := ShowData(data, deviceName)
	if err != nil {
		return err
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(out, line); err != nil {
			return fmt.Errorf("editor: write show output: %w", err)
		}
	}
	return nil
}

// ShowData returns one device's redacted running config without touching disk.
func ShowData(data []byte, deviceName string) ([]string, error) {
	index, err := xmledit.BuildIndex(data)
	if err != nil {
		return nil, err
	}
	device, err := index.FindDevice(deviceName)
	if err != nil {
		return nil, err
	}
	if device.RunningConfig == nil {
		return nil, fmt.Errorf("%w: %q", xmledit.ErrNoRunningConfig, deviceName)
	}

	lines := make([]string, 0, len(device.RunningConfig.Lines))
	for _, line := range device.RunningConfig.Lines {
		lines = append(lines, security.RedactLine(line.Text))
	}
	return lines, nil
}

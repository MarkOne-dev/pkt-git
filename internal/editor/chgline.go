package editor

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/xmledit"
)

var (
	// ErrLineNotFound reports an exact config line that is absent.
	ErrLineNotFound = errors.New("editor: line not found")
	// ErrLineAmbiguous reports more than one exact config line match.
	ErrLineAmbiguous = errors.New("editor: line is ambiguous")
)

// ChangeLineData replaces one complete LINE value inside an interface block.
func ChangeLineData(data []byte, deviceName, iface, oldLine, newLine string) ([]byte, error) {
	if strings.ContainsAny(newLine, "\r\n") {
		return nil, fmt.Errorf("editor: replacement line %q contains a line break", newLine)
	}

	index, err := xmledit.BuildIndex(data)
	if err != nil {
		return nil, err
	}
	device, err := index.FindDevice(deviceName)
	if err != nil {
		return nil, err
	}
	block, err := device.InterfaceBlock(iface)
	if err != nil {
		return nil, err
	}

	var matches []xmledit.Line
	for _, line := range block.Lines {
		if line.Text == oldLine {
			matches = append(matches, line)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %q in %s of device %q", ErrLineNotFound, oldLine, iface, deviceName)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("%w: %q in %s of device %q", ErrLineAmbiguous, oldLine, iface, deviceName)
	}

	line := matches[0]
	var escaped bytes.Buffer
	if err := xml.EscapeText(&escaped, []byte(newLine)); err != nil {
		return nil, fmt.Errorf("editor: escape replacement line: %w", err)
	}

	var out bytes.Buffer
	if line.SelfClosing {
		out.Write(data[:line.Start])
		out.WriteString("<LINE>")
		out.Write(escaped.Bytes())
		out.WriteString("</LINE>")
		out.Write(data[line.End:])
		return out.Bytes(), nil
	}

	out.Write(data[:line.ContentStart])
	out.Write(escaped.Bytes())
	out.Write(data[line.ContentEnd:])
	return out.Bytes(), nil
}

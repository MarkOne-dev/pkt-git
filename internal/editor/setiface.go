package editor

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/xmledit"
)

// SetInterfaceData replaces the content between an interface line and its next
// "!" line. It returns new bytes and leaves the input untouched.
func SetInterfaceData(data []byte, deviceName, iface string, lines []string) ([]byte, error) {
	for _, line := range lines {
		if strings.ContainsAny(line, "\r\n") {
			return nil, fmt.Errorf("editor: config line %q contains a line break", line)
		}
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

	var replacement bytes.Buffer
	replacement.WriteString(block.Newline)
	for _, line := range lines {
		replacement.WriteString(block.Indent)
		replacement.WriteString("<LINE>")
		if err := xml.EscapeText(&replacement, []byte(line)); err != nil {
			return nil, fmt.Errorf("editor: escape config line: %w", err)
		}
		replacement.WriteString("</LINE>")
		replacement.WriteString(block.Newline)
	}
	replacement.WriteString(block.Close.Indent)

	var out bytes.Buffer
	out.Write(data[:block.ContentStart])
	out.Write(replacement.Bytes())
	out.Write(data[block.ContentEnd:])
	return out.Bytes(), nil
}

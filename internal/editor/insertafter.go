package editor

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/xmledit"
)

// ErrPhraseNotFound reports a phrase that no config line contains.
var ErrPhraseNotFound = errors.New("editor: phrase not found")

// InsertAfterData inserts lines immediately after the first config line that
// contains phrase. It returns new bytes and leaves the input untouched.
func InsertAfterData(data []byte, deviceName, phrase string, lines []string) ([]byte, error) {
	for _, line := range lines {
		if strings.ContainsAny(line, "\r\n") {
			return nil, fmt.Errorf("editor: inserted line %q contains a line break", line)
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
	if device.RunningConfig == nil {
		return nil, fmt.Errorf("%w: %q", xmledit.ErrNoRunningConfig, deviceName)
	}

	var match *xmledit.Line
	for i := range device.RunningConfig.Lines {
		if strings.Contains(device.RunningConfig.Lines[i].Text, phrase) {
			match = &device.RunningConfig.Lines[i]
			break
		}
	}
	if match == nil {
		return nil, fmt.Errorf("%w: %q in device %q", ErrPhraseNotFound, phrase, deviceName)
	}

	var inserted bytes.Buffer
	for _, line := range lines {
		inserted.WriteString(match.Indent)
		inserted.WriteString("<LINE>")
		if err := xml.EscapeText(&inserted, []byte(line)); err != nil {
			return nil, fmt.Errorf("editor: escape inserted line: %w", err)
		}
		inserted.WriteString("</LINE>")
		inserted.WriteString(match.Newline)
	}

	var out bytes.Buffer
	insertAt := match.LineEnd
	suffix := data[match.End:insertAt]
	if len(bytes.TrimSpace(bytes.TrimRight(suffix, "\r\n"))) != 0 {
		insertAt = match.End
	}
	out.Write(data[:insertAt])
	if insertAt == match.End && insertAt < len(data) && data[insertAt] != '\n' && data[insertAt] != '\r' {
		out.WriteString(match.Newline)
	}
	out.Write(inserted.Bytes())
	out.Write(data[insertAt:])
	return out.Bytes(), nil
}

package editor

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/xmledit"
)

const insertXML = `<PACKETTRACER5><NETWORK><DEVICES>
  <DEVICE><ENGINE>
    <TYPE model="2811">Router</TYPE>
    <NAME translate="true">R1</NAME>
    <RUNNINGCONFIG>
      <LINE>!</LINE>
      <LINE>hostname R1</LINE>
      <LINE>!</LINE>
      <LINE>interface Fa0/1</LINE>
      <LINE> description Old</LINE>
      <LINE>!</LINE>
      <LINE> description Dup</LINE>
    </RUNNINGCONFIG>
  </ENGINE></DEVICE>
</DEVICES></NETWORK></PACKETTRACER5>`

func TestInsertAfterFirstMatchingLine(t *testing.T) {
	data := []byte(insertXML)
	index, err := xmledit.BuildIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	device, _ := index.FindDevice("R1")
	var matched xmledit.Line
	for _, line := range device.RunningConfig.Lines {
		if strings.Contains(line.Text, "hostname") {
			matched = line
			break
		}
	}
	if matched.Text == "" {
		t.Fatal("hostname line not found")
	}

	result, err := InsertAfterData(data, "R1", "hostname", []string{
		"ntp server 1.1.1.1",
		"logging host 2.2.2.2",
	})
	if err != nil {
		t.Fatalf("InsertAfterData: %v", err)
	}
	if !bytes.HasPrefix(result, data[:matched.LineEnd]) {
		t.Fatal("prefix before the insertion point changed")
	}
	if !bytes.HasSuffix(result, data[matched.LineEnd:]) {
		t.Fatal("suffix after the insertion point changed")
	}
	if string(data) != insertXML {
		t.Fatal("InsertAfterData modified the input slice")
	}

	index2, err := xmledit.BuildIndex(result)
	if err != nil {
		t.Fatalf("result is not valid XML: %v", err)
	}
	device2, _ := index2.FindDevice("R1")
	texts := make([]string, len(device2.RunningConfig.Lines))
	for i, line := range device2.RunningConfig.Lines {
		texts[i] = line.Text
	}
	want := []string{"!", "hostname R1", "ntp server 1.1.1.1", "logging host 2.2.2.2", "!", "interface Fa0/1", " description Old", "!", " description Dup"}
	if strings.Join(texts, "|") != strings.Join(want, "|") {
		t.Fatalf("lines = %q, want %q", texts, want)
	}
}

func TestInsertAfterMinifiedXML(t *testing.T) {
	data := []byte(`<PACKETTRACER5><NETWORK><DEVICES><DEVICE><ENGINE><TYPE model="2811">Router</TYPE><NAME translate="true">R1</NAME><RUNNINGCONFIG><LINE>hostname R1</LINE><LINE>!</LINE></RUNNINGCONFIG></ENGINE></DEVICE></DEVICES></NETWORK></PACKETTRACER5>`)
	result, err := InsertAfterData(data, "R1", "hostname", []string{"ntp server 1.1.1.1"})
	if err != nil {
		t.Fatalf("InsertAfterData: %v", err)
	}
	index, err := xmledit.BuildIndex(result)
	if err != nil {
		t.Fatalf("result is not valid XML: %v", err)
	}
	device, _ := index.FindDevice("R1")
	texts := make([]string, len(device.RunningConfig.Lines))
	for i, line := range device.RunningConfig.Lines {
		texts[i] = line.Text
	}
	want := []string{"hostname R1", "ntp server 1.1.1.1", "!"}
	if strings.Join(texts, "|") != strings.Join(want, "|") {
		t.Fatalf("lines = %q, want %q", texts, want)
	}
}

func TestInsertAfterOnlyFirstMatch(t *testing.T) {
	result, err := InsertAfterData([]byte(insertXML), "R1", "description", []string{" description Inserted"})
	if err != nil {
		t.Fatalf("InsertAfterData: %v", err)
	}
	index, err := xmledit.BuildIndex(result)
	if err != nil {
		t.Fatal(err)
	}
	device, _ := index.FindDevice("R1")
	count := 0
	for _, line := range device.RunningConfig.Lines {
		if line.Text == " description Inserted" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("inserted line count = %d, want 1", count)
	}
	oldCount := 0
	for _, line := range device.RunningConfig.Lines {
		if line.Text == " description Dup" {
			oldCount++
		}
	}
	if oldCount != 1 {
		t.Fatalf("original duplicate count = %d, want 1", oldCount)
	}
}

func TestInsertAfterEscapesXML(t *testing.T) {
	result, err := InsertAfterData([]byte(insertXML), "R1", "hostname", []string{" description A&B <x>"})
	if err != nil {
		t.Fatalf("InsertAfterData: %v", err)
	}
	if !bytes.Contains(result, []byte("A&amp;B &lt;x&gt;")) {
		t.Fatalf("escaped line not found:\n%s", result)
	}
}

func TestInsertAfterPreservesCRLF(t *testing.T) {
	data := []byte(strings.ReplaceAll(insertXML, "\n", "\r\n"))
	result, err := InsertAfterData(data, "R1", "hostname", []string{"ntp server 1.1.1.1"})
	if err != nil {
		t.Fatalf("InsertAfterData: %v", err)
	}
	if !bytes.Contains(result, []byte("      <LINE>ntp server 1.1.1.1</LINE>\r\n")) {
		t.Fatalf("inserted line does not use CRLF:\n%q", result)
	}
}

func TestInsertAfterErrors(t *testing.T) {
	if _, err := InsertAfterData([]byte(insertXML), "R1", "missing-phrase", nil); !errors.Is(err, ErrPhraseNotFound) {
		t.Fatalf("error = %v, want ErrPhraseNotFound", err)
	}
	if _, err := InsertAfterData([]byte(insertXML), "R1", "hostname", []string{"bad\nline"}); err == nil {
		t.Fatal("InsertAfterData accepted a line break")
	}
}

package editor

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/xmledit"
)

const setifaceXML = `<PACKETTRACER5><NETWORK><DEVICES>
  <DEVICE><ENGINE>
    <TYPE model="2811">Router</TYPE>
    <NAME translate="true">R1</NAME>
    <RUNNINGCONFIG>
      <LINE>!</LINE>
      <LINE>interface Fa0/1</LINE>
      <LINE> description Old</LINE>
      <LINE> switchport access vlan 10</LINE>
      <LINE>!</LINE>
      <LINE>interface Fa0/2</LINE>
      <LINE>!</LINE>
    </RUNNINGCONFIG>
  </ENGINE></DEVICE>
</DEVICES></NETWORK></PACKETTRACER5>`

func TestSetInterfaceReplacesBlockAndPreservesSurroundings(t *testing.T) {
	data := []byte(setifaceXML)
	index, err := xmledit.BuildIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	device, _ := index.FindDevice("R1")
	block, err := device.InterfaceBlock("Fa0/1")
	if err != nil {
		t.Fatal(err)
	}

	result, err := SetInterfaceData(data, "R1", "Fa0/1", []string{
		" description New",
		" switchport access vlan 20",
		" switchport mode access",
	})
	if err != nil {
		t.Fatalf("SetInterfaceData: %v", err)
	}

	if !bytes.HasPrefix(result, data[:block.ContentStart]) {
		t.Fatal("prefix before the interface block changed")
	}
	if !bytes.HasSuffix(result, data[block.ContentEnd:]) {
		t.Fatal("suffix after the interface block changed")
	}
	if bytes.Equal(result, data) {
		t.Fatal("SetInterfaceData returned the input unchanged")
	}
	if string(data) != setifaceXML {
		t.Fatal("SetInterfaceData modified the input slice")
	}

	index2, err := xmledit.BuildIndex(result)
	if err != nil {
		t.Fatalf("result is not valid XML: %v", err)
	}
	device2, _ := index2.FindDevice("R1")
	block2, err := device2.InterfaceBlock("Fa0/1")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{" description New", " switchport access vlan 20", " switchport mode access"}
	if len(block2.Lines) != len(want) {
		t.Fatalf("block lines = %d, want %d", len(block2.Lines), len(want))
	}
	for i, line := range block2.Lines {
		if line.Text != want[i] {
			t.Fatalf("line %d = %q, want %q", i, line.Text, want[i])
		}
	}
	if strings.Contains(string(result), "description Old") || strings.Contains(string(result), "vlan 10") {
		t.Fatal("old interface lines remain")
	}
}

func TestSetInterfaceMinifiedXML(t *testing.T) {
	data := []byte(`<PACKETTRACER5><NETWORK><DEVICES><DEVICE><ENGINE><TYPE model="2811">Router</TYPE><NAME translate="true">R1</NAME><RUNNINGCONFIG><LINE>interface Fa0/1</LINE><LINE> description Old</LINE><LINE>!</LINE></RUNNINGCONFIG></ENGINE></DEVICE></DEVICES></NETWORK></PACKETTRACER5>`)
	result, err := SetInterfaceData(data, "R1", "Fa0/1", []string{" description New"})
	if err != nil {
		t.Fatalf("SetInterfaceData: %v", err)
	}
	index, err := xmledit.BuildIndex(result)
	if err != nil {
		t.Fatalf("result is not valid XML: %v", err)
	}
	device, _ := index.FindDevice("R1")
	block, err := device.InterfaceBlock("Fa0/1")
	if err != nil {
		t.Fatal(err)
	}
	if len(block.Lines) != 1 || block.Lines[0].Text != " description New" {
		t.Fatalf("lines = %+v", block.Lines)
	}
	if strings.Contains(string(result), "description Old") {
		t.Fatal("old interface line remains")
	}
}

func TestSetInterfaceEmptyLinesRemovesBlockContent(t *testing.T) {
	result, err := SetInterfaceData([]byte(setifaceXML), "R1", "Fa0/2", nil)
	if err != nil {
		t.Fatalf("SetInterfaceData: %v", err)
	}
	index, err := xmledit.BuildIndex(result)
	if err != nil {
		t.Fatal(err)
	}
	device, _ := index.FindDevice("R1")
	block, err := device.InterfaceBlock("Fa0/2")
	if err != nil {
		t.Fatal(err)
	}
	if len(block.Lines) != 0 {
		t.Fatalf("block lines = %d, want 0", len(block.Lines))
	}
}

func TestSetInterfaceEscapesXML(t *testing.T) {
	line := " description A&B <x>"
	result, err := SetInterfaceData([]byte(setifaceXML), "R1", "Fa0/1", []string{line})
	if err != nil {
		t.Fatalf("SetInterfaceData: %v", err)
	}
	if !bytes.Contains(result, []byte("A&amp;B &lt;x&gt;")) {
		t.Fatalf("escaped line not found:\n%s", result)
	}
	index, err := xmledit.BuildIndex(result)
	if err != nil {
		t.Fatal(err)
	}
	device, _ := index.FindDevice("R1")
	block, err := device.InterfaceBlock("Fa0/1")
	if err != nil {
		t.Fatal(err)
	}
	if block.Lines[0].Text != line {
		t.Fatalf("decoded line = %q, want %q", block.Lines[0].Text, line)
	}
}

func TestSetInterfacePreservesCRLF(t *testing.T) {
	data := []byte(strings.ReplaceAll(setifaceXML, "\n", "\r\n"))
	result, err := SetInterfaceData(data, "R1", "Fa0/1", []string{" description CRLF"})
	if err != nil {
		t.Fatalf("SetInterfaceData: %v", err)
	}
	if !bytes.Contains(result, []byte("      <LINE> description CRLF</LINE>\r\n")) {
		t.Fatalf("generated line does not use CRLF:\n%q", result)
	}
}

func TestSetInterfaceErrors(t *testing.T) {
	if _, err := SetInterfaceData([]byte(setifaceXML), "R1", "Fa0/9", nil); !errors.Is(err, xmledit.ErrInterfaceNotFound) {
		t.Fatalf("error = %v, want ErrInterfaceNotFound", err)
	}
	if _, err := SetInterfaceData([]byte(setifaceXML), "R1", "Fa0/1", []string{"line\nbreak"}); err == nil {
		t.Fatal("SetInterfaceData accepted a line break")
	}
}

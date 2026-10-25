package editor

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/xmledit"
)

const chglineXML = `<PACKETTRACER5><NETWORK><DEVICES>
  <DEVICE><ENGINE>
    <TYPE model="2811">Router</TYPE>
    <NAME translate="true">R1</NAME>
    <RUNNINGCONFIG>
      <LINE>!</LINE>
      <LINE>interface Fa0/1</LINE>
      <LINE> description Old</LINE>
      <LINE> switchport access vlan 10</LINE>
      <LINE> description Dup</LINE>
      <LINE> description Dup</LINE>
      <LINE>!</LINE>
      <LINE>interface Fa0/2</LINE>
      <LINE/>
      <LINE>!</LINE>
    </RUNNINGCONFIG>
  </ENGINE></DEVICE>
</DEVICES></NETWORK></PACKETTRACER5>`

func TestChangeLineReplacesOneCompleteLine(t *testing.T) {
	data := []byte(chglineXML)
	index, err := xmledit.BuildIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	device, _ := index.FindDevice("R1")
	block, err := device.InterfaceBlock("Fa0/1")
	if err != nil {
		t.Fatal(err)
	}

	result, err := ChangeLineData(data, "R1", "Fa0/1", " description Old", " description New")
	if err != nil {
		t.Fatalf("ChangeLineData: %v", err)
	}
	if !bytes.HasPrefix(result, data[:block.ContentStart]) {
		t.Fatal("prefix before the interface block changed")
	}
	if !bytes.HasSuffix(result, data[block.ContentEnd:]) {
		t.Fatal("suffix after the interface block changed")
	}
	if string(data) != chglineXML {
		t.Fatal("ChangeLineData modified the input slice")
	}

	index2, err := xmledit.BuildIndex(result)
	if err != nil {
		t.Fatal(err)
	}
	device2, _ := index2.FindDevice("R1")
	block2, err := device2.InterfaceBlock("Fa0/1")
	if err != nil {
		t.Fatal(err)
	}
	if block2.Lines[0].Text != " description New" {
		t.Fatalf("changed line = %q", block2.Lines[0].Text)
	}
	if block2.Lines[1].Text != " switchport access vlan 10" {
		t.Fatalf("neighbor line changed: %q", block2.Lines[1].Text)
	}
}

func TestChangeLineRejectsSubstringAndAmbiguousMatches(t *testing.T) {
	if _, err := ChangeLineData([]byte(chglineXML), "R1", "Fa0/1", "description", "x"); !errors.Is(err, ErrLineNotFound) {
		t.Fatalf("substring error = %v, want ErrLineNotFound", err)
	}
	if _, err := ChangeLineData([]byte(chglineXML), "R1", "Fa0/1", " description Dup", "x"); !errors.Is(err, ErrLineAmbiguous) {
		t.Fatalf("ambiguous error = %v, want ErrLineAmbiguous", err)
	}
}

func TestChangeLineEscapesXML(t *testing.T) {
	result, err := ChangeLineData([]byte(chglineXML), "R1", "Fa0/1", " description Old", " description A&B <x>")
	if err != nil {
		t.Fatalf("ChangeLineData: %v", err)
	}
	if !bytes.Contains(result, []byte("A&amp;B &lt;x&gt;")) {
		t.Fatalf("escaped replacement not found:\n%s", result)
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
	if block.Lines[0].Text != " description A&B <x>" {
		t.Fatalf("decoded line = %q", block.Lines[0].Text)
	}
}

func TestChangeLinePreservesCRLF(t *testing.T) {
	data := []byte(strings.ReplaceAll(chglineXML, "\n", "\r\n"))
	result, err := ChangeLineData(data, "R1", "Fa0/1", " description Old", " description CRLF")
	if err != nil {
		t.Fatalf("ChangeLineData: %v", err)
	}
	if !bytes.Contains(result, []byte("<LINE> description CRLF</LINE>\r\n")) {
		t.Fatalf("changed line does not keep CRLF:\n%q", result)
	}
}

func TestChangeLineHandlesSelfClosingLine(t *testing.T) {
	result, err := ChangeLineData([]byte(chglineXML), "R1", "Fa0/2", "", "hostname R1")
	if err != nil {
		t.Fatalf("ChangeLineData: %v", err)
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
	if len(block.Lines) != 1 || block.Lines[0].Text != "hostname R1" {
		t.Fatalf("lines = %+v", block.Lines)
	}
	if block.Lines[0].SelfClosing {
		t.Fatal("replaced line is still self-closing")
	}
}

func TestChangeLineErrors(t *testing.T) {
	if _, err := ChangeLineData([]byte(chglineXML), "R1", "Fa0/9", "x", "y"); !errors.Is(err, xmledit.ErrInterfaceNotFound) {
		t.Fatalf("error = %v, want ErrInterfaceNotFound", err)
	}
	if _, err := ChangeLineData([]byte(chglineXML), "R1", "Fa0/1", " description Old", "bad\nline"); err == nil {
		t.Fatal("ChangeLineData accepted a line break")
	}
}

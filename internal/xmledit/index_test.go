package xmledit

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
)

const sampleXML = `<PACKETTRACER5>
  <NETWORK>
    <DEVICES>
      <DEVICE>
        <ENGINE>
          <TYPE model="2811">Router</TYPE>
          <NAME translate="true">R1</NAME>
          <RUNNINGCONFIG>
            <LINE>!</LINE>
            <LINE>interface Fa0/1</LINE>
            <LINE> description Uplink</LINE>
            <LINE> switchport access vlan 10</LINE>
            <LINE>!</LINE>
          </RUNNINGCONFIG>
        </ENGINE>
      </DEVICE>
      <DEVICE>
        <ENGINE>
          <TYPE model="PC-PT">PC</TYPE>
          <NAME translate="true">PC1</NAME>
          <RUNNINGCONFIG/>
        </ENGINE>
      </DEVICE>
      <DEVICE>
        <ENGINE>
          <TYPE model="2960">Switch</TYPE>
          <NAME translate="true">SW1</NAME>
        </ENGINE>
      </DEVICE>
    </DEVICES>
  </NETWORK>
</PACKETTRACER5>`

func TestBuildIndexDeviceSpans(t *testing.T) {
	data := []byte(sampleXML)
	index, err := BuildIndex(data)
	if err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}
	if len(index.Devices) != 3 {
		t.Fatalf("devices = %d, want 3", len(index.Devices))
	}

	r1, err := index.FindDevice("R1")
	if err != nil {
		t.Fatalf("FindDevice: %v", err)
	}
	if !bytes.HasPrefix(data[r1.Start:r1.End], []byte("<DEVICE>")) {
		t.Fatalf("device span does not start at DEVICE: %q", data[r1.Start:r1.End])
	}
	if !bytes.HasSuffix(data[r1.Start:r1.End], []byte("</DEVICE>")) {
		t.Fatalf("device span does not end at DEVICE: %q", data[r1.Start:r1.End])
	}
	if got := string(data[r1.NameStart:r1.NameEnd]); got != `<NAME translate="true">R1</NAME>` {
		t.Fatalf("name span = %q", got)
	}
	if r1.RunningConfig == nil || len(r1.RunningConfig.Lines) != 5 {
		t.Fatalf("running config = %+v", r1.RunningConfig)
	}

	pc, err := index.FindDevice("PC1")
	if err != nil {
		t.Fatal(err)
	}
	if pc.RunningConfig == nil {
		t.Fatal("self-closing RUNNINGCONFIG was not indexed")
	}
	if len(pc.RunningConfig.Lines) != 0 {
		t.Fatalf("self-closing RUNNINGCONFIG has %d lines", len(pc.RunningConfig.Lines))
	}

	sw, err := index.FindDevice("SW1")
	if err != nil {
		t.Fatal(err)
	}
	if sw.RunningConfig != nil {
		t.Fatalf("device without RUNNINGCONFIG got %+v", sw.RunningConfig)
	}
}

func TestBuildIndexLineSpans(t *testing.T) {
	data := []byte(sampleXML)
	index, err := BuildIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	device, err := index.FindDevice("R1")
	if err != nil {
		t.Fatal(err)
	}

	var target *Line
	for i := range device.RunningConfig.Lines {
		if device.RunningConfig.Lines[i].Text == "interface Fa0/1" {
			target = &device.RunningConfig.Lines[i]
			break
		}
	}
	if target == nil {
		t.Fatal("interface line not found")
	}

	if got := string(data[target.Start:target.End]); got != "<LINE>interface Fa0/1</LINE>" {
		t.Fatalf("element span = %q", got)
	}
	if got := string(data[target.ContentStart:target.ContentEnd]); got != "interface Fa0/1" {
		t.Fatalf("content span = %q", got)
	}
	if target.LineStart >= target.Start || target.LineEnd <= target.End {
		t.Fatalf("physical line bounds = %d..%d for element %d..%d", target.LineStart, target.LineEnd, target.Start, target.End)
	}
	if got := string(data[target.LineStart:target.Start]); got != "            " {
		t.Fatalf("indent = %q", got)
	}
	if target.Newline != "\n" {
		t.Fatalf("newline = %q", target.Newline)
	}
}

func TestBuildIndexInterfaceBlock(t *testing.T) {
	data := []byte(sampleXML)
	index, err := BuildIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	device, err := index.FindDevice("R1")
	if err != nil {
		t.Fatal(err)
	}

	block, err := device.InterfaceBlock("Fa0/1")
	if err != nil {
		t.Fatalf("InterfaceBlock: %v", err)
	}
	if block.Interface.Text != "interface Fa0/1" {
		t.Fatalf("interface line = %q", block.Interface.Text)
	}
	if block.Close.Text != "!" {
		t.Fatalf("close line = %q", block.Close.Text)
	}
	if len(block.Lines) != 2 {
		t.Fatalf("block lines = %d, want 2", len(block.Lines))
	}
	if block.Lines[0].Text != " description Uplink" {
		t.Fatalf("first block line = %q", block.Lines[0].Text)
	}
	if block.ContentStart != block.Interface.LineEnd || block.ContentEnd != block.Close.LineStart {
		t.Fatalf("block bounds = %d..%d", block.ContentStart, block.ContentEnd)
	}
	if got := string(data[block.ContentStart:block.ContentEnd]); !strings.Contains(got, "description Uplink") {
		t.Fatalf("block content = %q", got)
	}
	if block.Newline != "\n" {
		t.Fatalf("block newline = %q", block.Newline)
	}
}

func TestBuildIndexCRLF(t *testing.T) {
	data := []byte(strings.ReplaceAll(sampleXML, "\n", "\r\n"))
	index, err := BuildIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	device, err := index.FindDevice("R1")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range device.RunningConfig.Lines {
		if line.Newline != "\r\n" {
			t.Fatalf("line %q newline = %q", line.Text, line.Newline)
		}
		if !bytes.HasSuffix(data[line.LineStart:line.LineEnd], []byte("\r\n")) {
			t.Fatalf("line %q does not include CRLF", line.Text)
		}
	}

	block, err := device.InterfaceBlock("Fa0/1")
	if err != nil {
		t.Fatal(err)
	}
	if block.Newline != "\r\n" {
		t.Fatalf("block newline = %q, want CRLF", block.Newline)
	}
}

func TestBuildIndexSelfClosingLine(t *testing.T) {
	data := []byte(`<PACKETTRACER5><NETWORK><DEVICES><DEVICE><ENGINE>
		<NAME translate="true">R1</NAME>
		<RUNNINGCONFIG><LINE/></RUNNINGCONFIG>
	</ENGINE></DEVICE></DEVICES></NETWORK></PACKETTRACER5>`)
	index, err := BuildIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	device, err := index.FindDevice("R1")
	if err != nil {
		t.Fatal(err)
	}
	if len(device.RunningConfig.Lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(device.RunningConfig.Lines))
	}
	line := device.RunningConfig.Lines[0]
	if !line.SelfClosing {
		t.Fatal("self-closing LINE was not marked")
	}
	if line.Text != "" {
		t.Fatalf("text = %q, want empty", line.Text)
	}
	if line.ContentStart != line.ContentEnd {
		t.Fatalf("content span = %d..%d", line.ContentStart, line.ContentEnd)
	}
}

func TestBuildIndexErrors(t *testing.T) {
	index, err := BuildIndex([]byte(sampleXML))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := index.FindDevice("missing"); !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("error = %v, want ErrDeviceNotFound", err)
	}

	duplicate := []byte(`<PACKETTRACER5><NETWORK><DEVICES>
		<DEVICE><ENGINE><NAME translate="true">D</NAME></ENGINE></DEVICE>
		<DEVICE><ENGINE><NAME translate="true">D</NAME></ENGINE></DEVICE>
	</DEVICES></NETWORK></PACKETTRACER5>`)
	dupIndex, err := BuildIndex(duplicate)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dupIndex.FindDevice("D"); !errors.Is(err, ErrDeviceAmbiguous) {
		t.Fatalf("error = %v, want ErrDeviceAmbiguous", err)
	}

	device, _ := index.FindDevice("SW1")
	if _, err := device.InterfaceBlock("Fa0/1"); !errors.Is(err, ErrNoRunningConfig) {
		t.Fatalf("error = %v, want ErrNoRunningConfig", err)
	}
	r1, _ := index.FindDevice("R1")
	if _, err := r1.InterfaceBlock("Fa0/9"); !errors.Is(err, ErrInterfaceNotFound) {
		t.Fatalf("error = %v, want ErrInterfaceNotFound", err)
	}

	if _, err := BuildIndex([]byte(`<PACKETTRACER5>`)); err == nil {
		t.Fatal("BuildIndex accepted malformed XML")
	}
}

func TestBuildIndexSpecimen(t *testing.T) {
	data := mustDecodeSpecimen(t)
	index, err := BuildIndex(data)
	if err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}
	if len(index.Devices) != 107 {
		t.Fatalf("devices = %d, want 107", len(index.Devices))
	}
	isp1, err := index.FindDevice("ISP1")
	if err != nil {
		t.Fatalf("FindDevice ISP1: %v", err)
	}
	if isp1.RunningConfig == nil || len(isp1.RunningConfig.Lines) == 0 {
		t.Fatal("ISP1 has no indexed running config")
	}
	if _, err := isp1.InterfaceBlock("FastEthernet0/0"); err != nil {
		t.Fatalf("ISP1 FastEthernet0/0: %v", err)
	}
}

func mustDecodeSpecimen(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "specimens", "Caso1_Grupo09-2.pkt")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := codec.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

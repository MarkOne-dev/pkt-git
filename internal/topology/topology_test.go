package topology

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
)

func TestParseSpecimen(t *testing.T) {
	doc := loadSpecimen(t)

	if doc.Root != "PACKETTRACER5" {
		t.Fatalf("Root = %q, want PACKETTRACER5", doc.Root)
	}
	if len(doc.Devices) != 107 {
		t.Fatalf("devices = %d, want 107", len(doc.Devices))
	}
	if len(doc.Links) != 106 {
		t.Fatalf("links = %d, want 106", len(doc.Links))
	}

	first := doc.Devices[0]
	if first.Name != "Power Distribution Device0" {
		t.Fatalf("first device = %q", first.Name)
	}
	if first.Type != "Power Distribution Device" || first.Model != "Power Distribution Device" {
		t.Fatalf("first device type/model = %q/%q", first.Type, first.Model)
	}

	isp1 := findDevice(t, doc, "ISP1")
	if isp1.Type != "Router" || isp1.Model != "2811" {
		t.Fatalf("ISP1 type/model = %q/%q", isp1.Type, isp1.Model)
	}
	if !containsLine(isp1.RunningConfig, "hostname ISP1") {
		t.Fatal("ISP1 running config does not contain its hostname")
	}

	link := doc.Links[0]
	if link.LinkType != "eSerial" {
		t.Fatalf("first link type = %q, want eSerial", link.LinkType)
	}
	if link.From.Name != "ISP1" || link.From.Port != "Serial0/2/1" {
		t.Fatalf("first link from = %+v", link.From)
	}
	if link.To.Name != "ISP2" || link.To.Port != "Serial0/2/1" {
		t.Fatalf("first link to = %+v", link.To)
	}
}

func TestParseUnresolvedEndpointUsesFallback(t *testing.T) {
	doc := parseString(t, `<PACKETTRACER5><NETWORK>
		<DEVICES><DEVICE><ENGINE>
			<TYPE model="2811">Router</TYPE>
			<NAME>A</NAME>
			<SAVE_REF_ID>save-ref-id:1</SAVE_REF_ID>
		</ENGINE></DEVICE></DEVICES>
		<LINKS><LINK><TYPE>eCopper</TYPE><CABLE>
			<FROM>save-ref-id:1</FROM><PORT>Fa0</PORT>
			<TO>save-ref-id:123456789012</TO><PORT>Fa1</PORT>
		</CABLE></LINK></LINKS>
	</NETWORK></PACKETTRACER5>`)

	if len(doc.Links) != 1 {
		t.Fatalf("links = %d, want 1", len(doc.Links))
	}
	link := doc.Links[0]
	if link.From.Name != "A" || link.From.Port != "Fa0" {
		t.Fatalf("from = %+v", link.From)
	}
	if link.To.Name != "56789012" || link.To.Port != "Fa1" {
		t.Fatalf("to = %+v, want fallback 56789012/Fa1", link.To)
	}
}

func TestParsePreservesConfigOrder(t *testing.T) {
	doc := parseString(t, `<PACKETTRACER5><NETWORK>
		<DEVICES><DEVICE><ENGINE>
			<TYPE model="2960">Switch</TYPE><NAME>SW1</NAME>
			<RUNNINGCONFIG><LINE>!</LINE><LINE>hostname SW1</LINE><LINE>!</LINE></RUNNINGCONFIG>
		</ENGINE></DEVICE></DEVICES>
	</NETWORK></PACKETTRACER5>`)

	if len(doc.Devices) != 1 {
		t.Fatalf("devices = %d, want 1", len(doc.Devices))
	}
	want := []string{"!", "hostname SW1", "!"}
	if strings.Join(doc.Devices[0].RunningConfig, "|") != strings.Join(want, "|") {
		t.Fatalf("config = %v, want %v", doc.Devices[0].RunningConfig, want)
	}
}

func TestParseSkipsDeviceWithoutEngine(t *testing.T) {
	doc := parseString(t, `<PACKETTRACER5><NETWORK>
		<DEVICES><DEVICE/><DEVICE><ENGINE><NAME>D1</NAME></ENGINE></DEVICE></DEVICES>
	</NETWORK></PACKETTRACER5>`)
	if len(doc.Devices) != 1 || doc.Devices[0].Name != "D1" {
		t.Fatalf("devices = %+v", doc.Devices)
	}
}

func TestParseMissingNetwork(t *testing.T) {
	if _, err := Parse([]byte(`<PACKETTRACER5/>`)); !errors.Is(err, ErrNoNetwork) {
		t.Fatalf("error = %v, want ErrNoNetwork", err)
	}
}

func TestParseMalformedXML(t *testing.T) {
	if _, err := Parse([]byte(`<PACKETTRACER5>`)); err == nil {
		t.Fatal("Parse accepted malformed XML")
	}
}

func loadSpecimen(t *testing.T) *Document {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "specimens", "Caso1_Grupo09-2.pkt")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read specimen: %v", err)
	}
	xmlData, err := decodeSpecimen(raw)
	if err != nil {
		t.Fatalf("decode specimen: %v", err)
	}
	doc, err := Parse(xmlData)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return doc
}

func parseString(t *testing.T, xmlData string) *Document {
	t.Helper()
	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return doc
}

func decodeSpecimen(raw []byte) ([]byte, error) {
	return codec.Decode(raw)
}

func findDevice(t *testing.T, doc *Document, name string) Device {
	t.Helper()
	for _, dev := range doc.Devices {
		if dev.Name == name {
			return dev
		}
	}
	t.Fatalf("device %q not found", name)
	return Device{}
}

func containsLine(lines []string, want string) bool {
	for _, line := range lines {
		if line == want {
			return true
		}
	}
	return false
}

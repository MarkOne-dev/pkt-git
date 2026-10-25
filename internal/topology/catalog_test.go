package topology

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestCatalogForSpecimen(t *testing.T) {
	doc := loadSpecimen(t)
	catalog := CatalogFor("Caso1_Grupo09-2.pkt", doc)

	if catalog.SourceFile != "Caso1_Grupo09-2.pkt" {
		t.Fatalf("SourceFile = %q", catalog.SourceFile)
	}
	if catalog.DocumentRoot != "PACKETTRACER5" {
		t.Fatalf("DocumentRoot = %q", catalog.DocumentRoot)
	}
	if catalog.TotalDevices != 107 || catalog.TotalLinks != 106 {
		t.Fatalf("totals = %d/%d", catalog.TotalDevices, catalog.TotalLinks)
	}
	if len(catalog.Devices) != catalog.TotalDevices || len(catalog.Links) != catalog.TotalLinks {
		t.Fatalf("catalog lengths do not match totals")
	}

	data, err := catalog.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("catalog JSON is invalid")
	}
	if bytes.HasSuffix(data, []byte("\n")) {
		t.Fatal("catalog JSON has a trailing newline")
	}

	text := string(data)
	keys := []string{`"source_file"`, `"document_root"`, `"total_devices"`, `"total_links"`, `"devices"`, `"links"`}
	last := -1
	for _, key := range keys {
		index := strings.Index(text, key)
		if index < 0 {
			t.Fatalf("catalog JSON is missing %s", key)
		}
		if index <= last {
			t.Fatalf("catalog JSON key %s is out of order", key)
		}
		last = index
	}
	if strings.Contains(text, "runningconfig") || strings.Contains(text, "startupconfig") {
		t.Fatal("catalog JSON leaked device configuration fields")
	}
}

func TestCatalogFieldOrderAndFallback(t *testing.T) {
	doc := parseString(t, `<PACKETTRACER5><NETWORK>
		<DEVICES><DEVICE><ENGINE>
			<TYPE model="2811">Router</TYPE><NAME>A&amp;B</NAME><SAVE_REF_ID>save-ref-id:1</SAVE_REF_ID>
		</ENGINE></DEVICE></DEVICES>
		<LINKS><LINK><TYPE>eCopper</TYPE><CABLE>
			<FROM>save-ref-id:1</FROM><PORT>Fa0</PORT>
			<TO>save-ref-id:123456789012</TO><PORT>Fa1</PORT>
		</CABLE></LINK></LINKS>
	</NETWORK></PACKETTRACER5>`)

	data, err := CatalogFor("demo.pkt", doc).JSON()
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, `\u0026`) {
		t.Fatal("catalog JSON HTML-escaped the device name")
	}
	if !strings.Contains(text, `"name": "A&B"`) {
		t.Fatalf("catalog JSON does not contain the raw device name:\n%s", text)
	}
	if !strings.Contains(text, `"name": "56789012"`) {
		t.Fatalf("catalog JSON does not contain the fallback endpoint:\n%s", text)
	}
}

func TestCatalogEmptySlices(t *testing.T) {
	data, err := CatalogFor("empty.pkt", &Document{Root: "PACKETTRACER5"}).JSON()
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, `"devices": []`) || !strings.Contains(text, `"links": []`) {
		t.Fatalf("empty catalog did not render empty arrays:\n%s", text)
	}
}

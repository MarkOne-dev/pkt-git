package topology

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Catalog is the deterministic metadata/topology.json document.
type Catalog struct {
	SourceFile   string   `json:"source_file"`
	DocumentRoot string   `json:"document_root"`
	TotalDevices int      `json:"total_devices"`
	TotalLinks   int      `json:"total_links"`
	Devices      []Device `json:"devices"`
	Links        []Link   `json:"links"`
}

// CatalogFor builds the metadata catalog for a parsed document.
func CatalogFor(sourceFile string, doc *Document) Catalog {
	catalog := Catalog{
		SourceFile:   sourceFile,
		DocumentRoot: "PACKETTRACER5",
		Devices:      []Device{},
		Links:        []Link{},
	}
	if doc == nil {
		return catalog
	}

	catalog.DocumentRoot = doc.Root
	catalog.TotalDevices = len(doc.Devices)
	catalog.TotalLinks = len(doc.Links)
	catalog.Devices = make([]Device, len(doc.Devices))
	copy(catalog.Devices, doc.Devices)
	catalog.Links = make([]Link, len(doc.Links))
	copy(catalog.Links, doc.Links)
	return catalog
}

// JSON renders the catalog with stable field order and no trailing newline.
func (c Catalog) JSON() ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return nil, fmt.Errorf("topology: encode catalog: %w", err)
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

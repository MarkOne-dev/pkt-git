package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
	"github.com/MarkOne-dev/pkt-git/internal/topology"
)

// InspectDevice is one row in a read-only topology summary.
type InspectDevice struct {
	Name        string
	Type        string
	Model       string
	ConfigLines int
}

// InspectResult is the read-only summary of a binary.
type InspectResult struct {
	Name    string
	Devices []InspectDevice
	Links   int
}

// Inspect decodes a binary and returns a summary without writing any file.
func Inspect(pktPath string) (InspectResult, error) {
	raw, err := os.ReadFile(pktPath)
	if err != nil {
		return InspectResult{}, fmt.Errorf("snapshot: read %s: %w", pktPath, err)
	}
	xmlData, err := codec.Decode(raw)
	if err != nil {
		return InspectResult{}, err
	}
	doc, err := topology.Parse(xmlData)
	if err != nil {
		return InspectResult{}, err
	}

	result := InspectResult{
		Name:    filepath.Base(pktPath),
		Devices: make([]InspectDevice, 0, len(doc.Devices)),
		Links:   len(doc.Links),
	}
	for _, device := range doc.Devices {
		result.Devices = append(result.Devices, InspectDevice{
			Name:        device.Name,
			Type:        device.Type,
			Model:       device.Model,
			ConfigLines: len(device.RunningConfig),
		})
	}
	return result, nil
}

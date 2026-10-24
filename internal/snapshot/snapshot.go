// Package snapshot owns the filesystem contract for textual Packet Tracer
// snapshots: export, build, sync, and inspect.
package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
	"github.com/MarkOne-dev/pkt-git/internal/fsx"
	"github.com/MarkOne-dev/pkt-git/internal/topology"
)

// ExportResult summarizes a completed export.
type ExportResult struct {
	XMLSize     int
	ConfigCount int
	DeviceCount int
	LinkCount   int
}

type exportFiles struct {
	xml      []byte
	configs  map[string][]byte
	metadata []byte
	devices  int
	links    int
}

// Export decodes a binary, parses it, computes every output in memory, and
// only then refreshes the snapshot directory. Decode and parse failures leave
// the previous snapshot untouched.
func Export(pktPath, outputDir string) (ExportResult, error) {
	raw, err := os.ReadFile(pktPath)
	if err != nil {
		return ExportResult{}, fmt.Errorf("snapshot: read %s: %w", pktPath, err)
	}

	files, err := prepareExport(filepath.Base(pktPath), raw)
	if err != nil {
		return ExportResult{}, err
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return ExportResult{}, fmt.Errorf("snapshot: create %s: %w", outputDir, err)
	}
	configDir := filepath.Join(outputDir, "configs")
	metadataDir := filepath.Join(outputDir, "metadata")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return ExportResult{}, fmt.Errorf("snapshot: create %s: %w", configDir, err)
	}
	if err := os.MkdirAll(metadataDir, 0o755); err != nil {
		return ExportResult{}, fmt.Errorf("snapshot: create %s: %w", metadataDir, err)
	}

	if err := fsx.WriteFileAtomic(filepath.Join(outputDir, "base.snapshot.xml"), files.xml, 0o644); err != nil {
		return ExportResult{}, err
	}

	names := make([]string, 0, len(files.configs))
	for name := range files.configs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := fsx.WriteFileAtomic(filepath.Join(configDir, name), files.configs[name], 0o644); err != nil {
			return ExportResult{}, err
		}
	}

	if err := fsx.WriteFileAtomic(filepath.Join(metadataDir, "topology.json"), files.metadata, 0o644); err != nil {
		return ExportResult{}, err
	}

	return ExportResult{
		XMLSize:     len(files.xml),
		ConfigCount: len(files.configs),
		DeviceCount: files.devices,
		LinkCount:   files.links,
	}, nil
}

func prepareExport(sourceFile string, raw []byte) (exportFiles, error) {
	xmlData, err := codec.Decode(raw)
	if err != nil {
		return exportFiles{}, err
	}

	doc, err := topology.Parse(xmlData)
	if err != nil {
		return exportFiles{}, err
	}
	if err := topology.ValidateUniqueConfigNames(doc.Devices); err != nil {
		return exportFiles{}, err
	}

	configs := make(map[string][]byte)
	for _, device := range doc.Devices {
		if len(device.RunningConfig) == 0 {
			continue
		}
		configs[topology.ConfigFileName(device)] = []byte(strings.Join(device.RunningConfig, "\n") + "\n")
	}

	metadata, err := topology.CatalogFor(sourceFile, doc).JSON()
	if err != nil {
		return exportFiles{}, err
	}

	return exportFiles{
		xml:      xmlData,
		configs:  configs,
		metadata: metadata,
		devices:  len(doc.Devices),
		links:    len(doc.Links),
	}, nil
}

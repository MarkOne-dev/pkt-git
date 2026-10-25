// Package snapshot owns the filesystem contract for textual Packet Tracer
// snapshots: export, build, sync, and inspect.
package snapshot

import (
	"fmt"
	"os"
	"path"
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
	// Files lists generated paths relative to the output directory, using
	// forward slashes. Removed lists stale config files that were deleted.
	Files   []string
	Removed []string
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
		return ExportResult{}, fmt.Errorf("snapshot: export %s: %w", pktPath, err)
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

	removed, err := pruneConfigs(configDir, files.configs)
	if err != nil {
		return ExportResult{}, err
	}
	if err := fsx.WriteFileAtomic(filepath.Join(metadataDir, "topology.json"), files.metadata, 0o644); err != nil {
		return ExportResult{}, err
	}

	generated := make([]string, 0, 2+len(names)+len(removed))
	generated = append(generated, "base.snapshot.xml")
	for _, name := range names {
		generated = append(generated, path.Join("configs", name))
	}
	generated = append(generated, path.Join("metadata", "topology.json"))
	for _, name := range removed {
		generated = append(generated, path.Join("configs", name))
	}

	return ExportResult{
		XMLSize:     len(files.xml),
		ConfigCount: len(files.configs),
		DeviceCount: files.devices,
		LinkCount:   files.links,
		Files:       generated,
		Removed:     removed,
	}, nil
}

func prepareExport(sourceFile string, raw []byte) (exportFiles, error) {
	xmlData, err := codec.Decode(raw)
	if err != nil {
		return exportFiles{}, err
	}
	return prepareSnapshot(sourceFile, xmlData)
}

func prepareSnapshot(sourceFile string, xmlData []byte) (exportFiles, error) {
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

// DefaultSnapshotDir returns the conventional snapshot directory for a .pkt file:
// "snapshot/<stem>".
func DefaultSnapshotDir(pktPath string) string {
	base := filepath.Base(pktPath)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	return filepath.Join("snapshot", stem)
}

// FindTopologies searches root for .pkt and .pka files, skipping ignored directories
// such as .git, .githooks, snapshot, bin, dist, node_modules, and skills.
// It returns sorted relative paths to discovered topology files.
func FindTopologies(root string) ([]string, error) {
	if root == "" {
		root = "."
	}
	var topologies []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "snapshot" || name == "bin" || name == "dist" || name == "node_modules" || name == "skills" {
				if p != root {
					return filepath.SkipDir
				}
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext == ".pkt" || ext == ".pka" {
			if strings.HasSuffix(d.Name(), ".rebuilt.pkt") {
				return nil
			}
			rel, relErr := filepath.Rel(root, p)
			if relErr != nil {
				topologies = append(topologies, filepath.Clean(p))
			} else {
				topologies = append(topologies, filepath.Clean(rel))
			}
		}
		return nil
	})
	sort.Strings(topologies)
	return topologies, err
}

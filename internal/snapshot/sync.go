package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/fsx"
)

// SyncResult summarizes a completed derived-file refresh.
type SyncResult struct {
	XMLSize     int
	ConfigCount int
	DeviceCount int
	LinkCount   int
	Removed     []string
}

// Sync regenerates configs/*.cfg and metadata/topology.json from a snapshot
// XML, creating missing directories and pruning stale generated configs. The
// XML itself is never written.
func Sync(snapshotPath string) (SyncResult, error) {
	xmlPath, err := resolveSnapshotXML(snapshotPath)
	if err != nil {
		return SyncResult{}, err
	}
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		return SyncResult{}, fmt.Errorf("snapshot: read %s: %w", xmlPath, err)
	}

	projectDir := filepath.Dir(xmlPath)
	files, err := prepareSnapshot(sourceFileFor(projectDir), xmlData)
	if err != nil {
		return SyncResult{}, err
	}

	configDir := filepath.Join(projectDir, "configs")
	metadataDir := filepath.Join(projectDir, "metadata")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return SyncResult{}, fmt.Errorf("snapshot: create %s: %w", configDir, err)
	}
	if err := os.MkdirAll(metadataDir, 0o755); err != nil {
		return SyncResult{}, fmt.Errorf("snapshot: create %s: %w", metadataDir, err)
	}

	names := make([]string, 0, len(files.configs))
	for name := range files.configs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := fsx.WriteFileAtomic(filepath.Join(configDir, name), files.configs[name], 0o644); err != nil {
			return SyncResult{}, err
		}
	}

	removed, err := pruneConfigs(configDir, files.configs)
	if err != nil {
		return SyncResult{}, err
	}
	if err := fsx.WriteFileAtomic(filepath.Join(metadataDir, "topology.json"), files.metadata, 0o644); err != nil {
		return SyncResult{}, err
	}

	return SyncResult{
		XMLSize:     len(xmlData),
		ConfigCount: len(files.configs),
		DeviceCount: files.devices,
		LinkCount:   files.links,
		Removed:     removed,
	}, nil
}

func resolveSnapshotXML(snapshotPath string) (string, error) {
	info, err := os.Stat(snapshotPath)
	if err != nil {
		return "", fmt.Errorf("snapshot: stat %s: %w", snapshotPath, err)
	}
	if info.IsDir() {
		return filepath.Join(snapshotPath, "base.snapshot.xml"), nil
	}
	return snapshotPath, nil
}

func sourceFileFor(projectDir string) string {
	metadataPath := filepath.Join(projectDir, "metadata", "topology.json")
	var existing struct {
		SourceFile string `json:"source_file"`
	}
	if data, err := os.ReadFile(metadataPath); err == nil {
		if json.Unmarshal(data, &existing) == nil && existing.SourceFile != "" {
			return existing.SourceFile
		}
	}

	name := filepath.Base(projectDir)
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "snapshot"
	}
	return name + ".pkt"
}

func pruneConfigs(configDir string, expected map[string][]byte) ([]string, error) {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, fmt.Errorf("snapshot: read %s: %w", configDir, err)
	}

	removed := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cfg") {
			continue
		}
		if _, ok := expected[entry.Name()]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(configDir, entry.Name())); err != nil {
			return nil, fmt.Errorf("snapshot: remove stale config %s: %w", entry.Name(), err)
		}
		removed = append(removed, entry.Name())
	}
	sort.Strings(removed)
	return removed, nil
}

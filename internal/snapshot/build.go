package snapshot

import (
	"fmt"
	"os"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
	"github.com/MarkOne-dev/pkt-git/internal/fsx"
)

// BuildResult summarizes a completed rebuild.
type BuildResult struct {
	Size int
}

// Build validates a snapshot XML and atomically writes the encoded binary.
// Validation and encoding happen before the destination is touched.
func Build(xmlPath, outputPath string) (BuildResult, error) {
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		return BuildResult{}, fmt.Errorf("snapshot: read %s: %w", xmlPath, err)
	}
	if err := codec.ValidateXML(xmlData); err != nil {
		return BuildResult{}, fmt.Errorf("snapshot: invalid XML in %s: %w", xmlPath, err)
	}

	data, err := codec.Encode(xmlData)
	if err != nil {
		return BuildResult{}, fmt.Errorf("snapshot: encode %s: %w", xmlPath, err)
	}
	if err := fsx.WriteFileAtomic(outputPath, data, 0o644); err != nil {
		return BuildResult{}, err
	}
	return BuildResult{Size: len(data)}, nil
}

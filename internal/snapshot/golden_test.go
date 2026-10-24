package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"testing"
)

const (
	goldenXMLHash      = "d6ee8f3a2851dfe20d570f5f3f842e26b27ce0d860fce096bc7c074a64aa7c68"
	goldenMetadataHash = "839a5a113b93127762c8c87f5e2822ceb7ae86efcf4aba4e39c078b6d3cd3ca7"
	goldenConfigHash   = "fba6d53300a03b02483cdbd6d7a8dd6d92f6d2ddd159a5ac519eccc07c90f0f3"
)

func TestGoldenSpecimen(t *testing.T) {
	files, err := prepareExport("Caso1_Grupo09-2.pkt", mustRead(t, specimenPath(t)))
	if err != nil {
		t.Fatalf("prepareExport: %v", err)
	}

	if got := sha256Hex(files.xml); got != goldenXMLHash {
		t.Fatalf("XML hash = %s, want %s", got, goldenXMLHash)
	}
	if got := sha256Hex(files.metadata); got != goldenMetadataHash {
		t.Fatalf("metadata hash = %s, want %s", got, goldenMetadataHash)
	}
	if got := configSetHash(files.configs); got != goldenConfigHash {
		t.Fatalf("config set hash = %s, want %s", got, goldenConfigHash)
	}
	if len(files.configs) != 34 {
		t.Fatalf("config count = %d, want 34", len(files.configs))
	}
	if files.devices != 107 || files.links != 106 {
		t.Fatalf("topology totals = %d/%d, want 107/106", files.devices, files.links)
	}
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func configSetHash(configs map[string][]byte) string {
	names := make([]string, 0, len(configs))
	for name := range configs {
		names = append(names, name)
	}
	sort.Strings(names)

	hash := sha256.New()
	for _, name := range names {
		fmt.Fprintf(hash, "%s\x00", name)
		hash.Write(configs[name])
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

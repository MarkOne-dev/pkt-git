package snapshot

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInspectSpecimen(t *testing.T) {
	result, err := Inspect(specimenPath(t))
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if result.Name != "Caso1_Grupo09-2.pkt" {
		t.Fatalf("Name = %q", result.Name)
	}
	if len(result.Devices) != 107 || result.Links != 106 {
		t.Fatalf("summary = %d devices, %d links", len(result.Devices), result.Links)
	}
	if result.Devices[0].Name != "Power Distribution Device0" {
		t.Fatalf("first device = %q", result.Devices[0].Name)
	}

	isp1 := findInspectDevice(t, result, "ISP1")
	if isp1.Type != "Router" || isp1.Model != "2811" {
		t.Fatalf("ISP1 = %+v", isp1)
	}
	if isp1.ConfigLines == 0 {
		t.Fatal("ISP1 has no CLI config lines")
	}
}

func TestInspectWritesNoFiles(t *testing.T) {
	dir := t.TempDir()
	pktPath := filepath.Join(dir, "sample.pkt")
	if err := os.WriteFile(pktPath, mustRead(t, specimenPath(t)), 0o644); err != nil {
		t.Fatal(err)
	}
	before := captureTree(t, dir)

	if _, err := Inspect(pktPath); err != nil {
		t.Fatalf("Inspect: %v", err)
	}

	after := captureTree(t, dir)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("Inspect modified the filesystem")
	}
}

func TestInspectInvalidBinary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.pkt")
	if err := os.WriteFile(path, []byte("not a container"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(path); err == nil {
		t.Fatal("Inspect accepted an invalid binary")
	}
}

func findInspectDevice(t *testing.T, result InspectResult, name string) InspectDevice {
	t.Helper()
	for _, device := range result.Devices {
		if device.Name == name {
			return device
		}
	}
	t.Fatalf("device %q not found", name)
	return InspectDevice{}
}

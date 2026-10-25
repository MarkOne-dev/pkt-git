package ptdetect

import (
	"errors"
	"testing"
)

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Cisco Packet Tracer 8.2.2", "8.2.2"},
		{"/opt/pt/bin/PacketTracer8.1", "8.1"},
		{"C:\\Program Files\\Cisco Packet Tracer 8.0.1\\bin\\PacketTracer.exe", "8.0.1"},
		{"Cisco Packet Tracer v7.3.0", "7.3.0"},
		{"packettracer.AppImage", ""},
		{"random string without version", ""},
	}

	for _, tt := range tests {
		got := ExtractVersion(tt.input)
		if got != tt.want {
			t.Errorf("ExtractVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFindSystemPath(t *testing.T) {
	oldLookPath := LookPathFunc
	defer func() { LookPathFunc = oldLookPath }()

	LookPathFunc = func(file string) (string, error) {
		if file == "PacketTracer" {
			return "/usr/local/bin/PacketTracer", nil
		}
		return "", errors.New("not found")
	}

	inst := Find()
	if inst == nil {
		t.Fatal("expected installation found in PATH, got nil")
	}
	if inst.Path != "/usr/local/bin/PacketTracer" {
		t.Errorf("Path = %q, want /usr/local/bin/PacketTracer", inst.Path)
	}
	if inst.Method != "System PATH" {
		t.Errorf("Method = %q, want System PATH", inst.Method)
	}
}

func TestHostEnvironmentDetection(t *testing.T) {
	inst := Find()
	if inst != nil {
		t.Logf("Host detected Packet Tracer: Path=%s Version=%s Method=%s", inst.Path, inst.Version, inst.Method)
	} else {
		t.Log("Host detected Packet Tracer: nil")
	}
}

package topology

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"SW_ACC_ADMIN_PUNO", "SW_ACC_ADMIN_PUNO"},
		{"Router Lima", "Router_Lima"},
		{"Router/Lima", "Router_Lima"},
		{"a@@b", "a_b"},
		{"a..b", "a..b"},
		{"...", "device"},
		{"", "device"},
		{"-device-", "-device-"},
		{"device.v1", "device.v1"},
		{"MKT LOG", "MKT_LOG"},
		{"CON", "_CON"},
		{"NUL.cfg", "_NUL.cfg"},
		{"con", "_con"},
	}
	for _, tt := range tests {
		if got := SanitizeName(tt.in); got != tt.want {
			t.Errorf("SanitizeName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestConfigFileName(t *testing.T) {
	device := Device{Name: "SW_ACC_ADMIN_PUNO"}
	if got := ConfigFileName(device); got != "SW_ACC_ADMIN_PUNO.cfg" {
		t.Fatalf("ConfigFileName = %q", got)
	}
}

func TestValidateUniqueConfigNamesAcceptsDistinctDevices(t *testing.T) {
	devices := []Device{{Name: "R1"}, {Name: "R2"}, {Name: "R 3"}}
	if err := ValidateUniqueConfigNames(devices); err != nil {
		t.Fatalf("ValidateUniqueConfigNames: %v", err)
	}
}

func TestValidateUniqueConfigNamesRejectsCaseInsensitiveCollision(t *testing.T) {
	devices := []Device{{Name: "R1"}, {Name: "r1"}}
	err := ValidateUniqueConfigNames(devices)
	if err == nil {
		t.Fatal("ValidateUniqueConfigNames accepted a case-insensitive collision")
	}
	var collision *NameCollisionError
	if !errors.As(err, &collision) {
		t.Fatalf("error type = %T, want *NameCollisionError", err)
	}
	if collision.First != "R1" || collision.Second != "r1" {
		t.Fatalf("names = %q/%q", collision.First, collision.Second)
	}
}

func TestValidateUniqueConfigNamesRejectsCollision(t *testing.T) {
	devices := []Device{{Name: "a b"}, {Name: "a@b"}}
	err := ValidateUniqueConfigNames(devices)
	if err == nil {
		t.Fatal("ValidateUniqueConfigNames accepted a collision")
	}

	var collision *NameCollisionError
	if !errors.As(err, &collision) {
		t.Fatalf("error type = %T, want *NameCollisionError", err)
	}
	if collision.FileName != "a_b.cfg" {
		t.Fatalf("FileName = %q", collision.FileName)
	}
	if collision.First != "a b" || collision.Second != "a@b" {
		t.Fatalf("names = %q/%q", collision.First, collision.Second)
	}
	for _, want := range []string{"a b", "a@b", "a_b.cfg"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not mention %q", err, want)
		}
	}
}

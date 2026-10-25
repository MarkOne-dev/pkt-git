package editor

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/xmledit"
)

const showXML = `<PACKETTRACER5><NETWORK><DEVICES>
  <DEVICE><ENGINE>
    <TYPE model="2811">Router</TYPE>
    <NAME translate="true">R1</NAME>
    <RUNNINGCONFIG>
      <LINE>!</LINE>
      <LINE>enable secret 5 $1$abc</LINE>
      <LINE>username admin password cisco</LINE>
      <LINE>snmp-server community public RO</LINE>
      <LINE>no service password-encryption</LINE>
      <LINE>interface Fa0/1</LINE>
      <LINE>!</LINE>
    </RUNNINGCONFIG>
  </ENGINE></DEVICE>
  <DEVICE><ENGINE>
    <TYPE model="PC-PT">PC</TYPE>
    <NAME translate="true">PC1</NAME>
  </ENGINE></DEVICE>
</DEVICES></NETWORK></PACKETTRACER5>`

func TestShowRedactsSecretsAndLeavesFileUntouched(t *testing.T) {
	path := filepath.Join(t.TempDir(), "base.snapshot.xml")
	if err := os.WriteFile(path, []byte(showXML), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Show(path, "R1", &out); err != nil {
		t.Fatalf("Show: %v", err)
	}

	text := out.String()
	for _, want := range []string{
		"enable secret <redacted>",
		"username admin password <redacted>",
		"snmp-server community <redacted>",
		"no service password-encryption",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("show output does not contain %q:\n%s", want, text)
		}
	}
	for _, secret := range []string{"$1$abc", "cisco", "public"} {
		if strings.Contains(text, secret) {
			t.Fatalf("show output leaked %q:\n%s", secret, text)
		}
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("Show modified the snapshot")
	}
}

func TestShowPreservesLineOrder(t *testing.T) {
	lines, err := ShowData([]byte(showXML), "R1")
	if err != nil {
		t.Fatalf("ShowData: %v", err)
	}
	want := []string{
		"!",
		"enable secret <redacted>",
		"username admin password <redacted>",
		"snmp-server community <redacted>",
		"no service password-encryption",
		"interface Fa0/1",
		"!",
	}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Fatalf("lines = %q, want %q", lines, want)
	}
}

func TestShowMissingDevice(t *testing.T) {
	if _, err := ShowData([]byte(showXML), "missing"); !errors.Is(err, xmledit.ErrDeviceNotFound) {
		t.Fatalf("error = %v, want ErrDeviceNotFound", err)
	}
}

func TestShowDeviceWithoutConfig(t *testing.T) {
	if _, err := ShowData([]byte(showXML), "PC1"); !errors.Is(err, xmledit.ErrNoRunningConfig) {
		t.Fatalf("error = %v, want ErrNoRunningConfig", err)
	}
}

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const editCLIXML = `<PACKETTRACER5>
<NETWORK>
<DEVICES>
<DEVICE>
<ENGINE>
<TYPE model="2960">Switch</TYPE>
<NAME translate="true">R1</NAME>
<RUNNINGCONFIG>
<LINE>!</LINE>
<LINE>hostname R1</LINE>
<LINE>enable secret 5 hash</LINE>
<LINE>!</LINE>
<LINE>interface FastEthernet0/1</LINE>
<LINE> description Old</LINE>
<LINE> switchport access vlan 50</LINE>
<LINE>!</LINE>
</RUNNINGCONFIG>
</ENGINE>
</DEVICE>
</DEVICES>
</NETWORK>
</PACKETTRACER5>`

func TestEditShowLegacyOrder(t *testing.T) {
	xmlPath := writeEditSnapshot(t)
	before, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runCLI("edit", xmlPath, "show", "R1")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "hostname R1") || !strings.Contains(stdout, "<redacted>") {
		t.Fatalf("stdout = %q", stdout)
	}
	after, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("show modified the snapshot")
	}
}

func TestEditShowSubcommandOrder(t *testing.T) {
	xmlPath := writeEditSnapshot(t)
	code, stdout, stderr := runCLI("edit", "show", xmlPath, "R1")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "hostname R1") {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestEditSetInterface(t *testing.T) {
	xmlPath := writeEditSnapshot(t)
	code, stdout, stderr := runCLI("edit", xmlPath, "setiface", "R1", "FastEthernet0/1",
		" description New", " switchport access vlan 60")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "[ok] updated") {
		t.Fatalf("stdout = %q", stdout)
	}
	xml := readEditFile(t, xmlPath)
	if !strings.Contains(xml, "description New") || strings.Contains(xml, "description Old") {
		t.Fatalf("XML was not updated:\n%s", xml)
	}
	cfg := readEditFile(t, filepath.Join(filepath.Dir(xmlPath), "configs", "R1.cfg"))
	if !strings.Contains(cfg, "description New") {
		t.Fatalf("derived config was not refreshed:\n%s", cfg)
	}
}

func TestEditChangeLine(t *testing.T) {
	xmlPath := writeEditSnapshot(t)
	code, _, stderr := runCLI("edit", xmlPath, "chgline", "R1", "FastEthernet0/1",
		" switchport access vlan 50", " switchport access vlan 60")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	cfg := readEditFile(t, filepath.Join(filepath.Dir(xmlPath), "configs", "R1.cfg"))
	if !strings.Contains(cfg, "switchport access vlan 60") || strings.Contains(cfg, "vlan 50") {
		t.Fatalf("config = %q", cfg)
	}
}

func TestEditInsertAfter(t *testing.T) {
	xmlPath := writeEditSnapshot(t)
	code, _, stderr := runCLI("edit", xmlPath, "insertafter", "R1", "hostname",
		" ntp server 1.1.1.1")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	cfg := readEditFile(t, filepath.Join(filepath.Dir(xmlPath), "configs", "R1.cfg"))
	if !strings.Contains(cfg, "hostname R1\n ntp server 1.1.1.1") {
		t.Fatalf("config = %q", cfg)
	}
}

func TestEditUsageErrors(t *testing.T) {
	xmlPath := writeEditSnapshot(t)
	tests := [][]string{
		{"edit"},
		{"edit", xmlPath, "show"},
		{"edit", "show", xmlPath},
		{"edit", xmlPath, "nope"},
		{"edit", xmlPath, "chgline", "R1", "FastEthernet0/1", "old"},
		{"edit", xmlPath, "setiface", "R1", "FastEthernet0/1"},
		{"edit", xmlPath, "insertafter", "R1", "hostname"},
	}
	for _, args := range tests {
		code, _, stderr := runCLI(args...)
		if code != 2 {
			t.Fatalf("%q exit code = %d, want 2 (stderr=%q)", args, code, stderr)
		}
	}
}

func TestEditRuntimeError(t *testing.T) {
	xmlPath := writeEditSnapshot(t)
	code, _, stderr := runCLI("edit", xmlPath, "show", "missing")
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "device not found") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestEditHelpListsSubcommands(t *testing.T) {
	code, stdout, stderr := runCLI("edit", "--help")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%q", code, stderr)
	}
	for _, want := range []string{"show", "setiface", "chgline", "insertafter"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("help does not mention %q:\n%s", want, stdout)
		}
	}
}

func writeEditSnapshot(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "base.snapshot.xml")
	writeCLIFile(t, path, []byte(editCLIXML))
	return path
}

func readEditFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

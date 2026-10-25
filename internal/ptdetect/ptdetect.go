// Package ptdetect provides multi-platform discovery of Cisco Packet Tracer installations.
package ptdetect

import (
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// Installation holds metadata for a discovered Cisco Packet Tracer installation.
type Installation struct {
	Path    string // Absolute path or command to launch Packet Tracer
	Version string // Detected version (e.g. "8.2.2") or empty if unknown
	Method  string // Discovery mechanism (e.g. "System PATH", "XDG Desktop Entry", "Windows Registry")
}

var versionRegex = regexp.MustCompile(`(?i)(?:packet\s*tracer|v)?\s*([0-9]+(?:\.[0-9]+)+(?:[.-][0-9a-z]+)?)`)

// ExtractVersion extracts a semantic version number from a path, description, or comment.
func ExtractVersion(s string) string {
	matches := versionRegex.FindStringSubmatch(s)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// LookPathFunc allows overriding exec.LookPath in tests.
var LookPathFunc = exec.LookPath

// Find searches for a Cisco Packet Tracer installation using a multi-tiered pipeline:
// 1. System PATH
// 2. OS-specific desktop entries and application registries
// 3. Versioned directory globbing
// 4. Container packages (Flatpak / Snap)
func Find() *Installation {
	for _, bin := range []string{"PacketTracer", "packettracer", "packet-tracer", "pt"} {
		if path, err := LookPathFunc(bin); err == nil && path != "" {
			ver := ExtractVersion(path)
			return &Installation{
				Path:    path,
				Version: ver,
				Method:  "System PATH",
			}
		}
	}

	switch runtime.GOOS {
	case "linux":
		return detectLinux()
	case "windows":
		return detectWindows()
	case "darwin":
		return detectDarwin()
	default:
		return nil
	}
}

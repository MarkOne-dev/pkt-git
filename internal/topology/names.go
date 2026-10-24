package topology

import (
	"fmt"
	"regexp"
	"strings"
)

var unsafeNameChars = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

// NameCollisionError reports two devices that would write the same config file.
type NameCollisionError struct {
	FileName string
	First    string
	Second   string
}

func (e *NameCollisionError) Error() string {
	return fmt.Sprintf("topology: device names %q and %q both map to config file %q", e.First, e.Second, e.FileName)
}

// SanitizeName converts a device name into a safe config file stem.
func SanitizeName(name string) string {
	safe := unsafeNameChars.ReplaceAllString(name, "_")
	safe = strings.Trim(safe, "._")
	if safe == "" {
		return "device"
	}
	return safe
}

// ConfigFileName returns the config file name for a device.
func ConfigFileName(device Device) string {
	return SanitizeName(device.Name) + ".cfg"
}

// ValidateUniqueConfigNames rejects devices that would overwrite each other's
// config file.
func ValidateUniqueConfigNames(devices []Device) error {
	seen := make(map[string]string, len(devices))
	for _, device := range devices {
		fileName := ConfigFileName(device)
		if first, ok := seen[fileName]; ok {
			return &NameCollisionError{
				FileName: fileName,
				First:    first,
				Second:   device.Name,
			}
		}
		seen[fileName] = device.Name
	}
	return nil
}

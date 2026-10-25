package topology

import (
	"fmt"
	"regexp"
	"strings"
)

var unsafeNameChars = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

var windowsReservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

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
	// Windows reserves device names even when they have an extension (for
	// example CON.cfg). Prefix them so the generated file is portable.
	stem := strings.SplitN(safe, ".", 2)[0]
	if windowsReservedNames[strings.ToUpper(stem)] {
		safe = "_" + safe
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
		key := strings.ToLower(fileName)
		if first, ok := seen[key]; ok {
			return &NameCollisionError{
				FileName: fileName,
				First:    first,
				Second:   device.Name,
			}
		}
		seen[key] = device.Name
	}
	return nil
}

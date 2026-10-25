//go:build windows

package ptdetect

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func queryWindowsRegistry() *Installation {
	// 1. App Paths check (64-bit and 32-bit views)
	appPathKeys := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\PacketTracer.exe`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\App Paths\PacketTracer.exe`,
	}

	for _, keyPath := range appPathKeys {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
		if err == nil {
			val, _, err := k.GetStringValue("")
			k.Close()
			if err == nil && val != "" {
				cleanVal := strings.Trim(val, `"'`)
				if fi, err := os.Stat(cleanVal); err == nil && !fi.IsDir() {
					return &Installation{
						Path:    cleanVal,
						Version: ExtractVersion(cleanVal),
						Method:  "Windows Registry (App Paths)",
					}
				}
			}
		}
	}

	// 2. Uninstall keys check
	uninstallBases := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, baseKey := range uninstallBases {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, baseKey, registry.ENUMERATE_SUB_KEYS)
		if err != nil {
			continue
		}
		subkeys, err := k.ReadSubKeyNames(-1)
		k.Close()
		if err != nil {
			continue
		}

		for _, sub := range subkeys {
			subPath := baseKey + `\` + sub
			sk, err := registry.OpenKey(registry.LOCAL_MACHINE, subPath, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			dispName, _, _ := sk.GetStringValue("DisplayName")
			installLoc, _, _ := sk.GetStringValue("InstallLocation")
			dispVer, _, _ := sk.GetStringValue("DisplayVersion")
			dispIcon, _, _ := sk.GetStringValue("DisplayIcon")
			sk.Close()

			if strings.Contains(strings.ToLower(dispName), "packet tracer") {
				target := ""
				if installLoc != "" {
					cand := filepath.Join(installLoc, "bin", "PacketTracer.exe")
					if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
						target = cand
					}
				}
				if target == "" && dispIcon != "" {
					iconPath := strings.Trim(dispIcon, `"'`)
					if fi, err := os.Stat(iconPath); err == nil && !fi.IsDir() && strings.HasSuffix(strings.ToLower(iconPath), ".exe") {
						target = iconPath
					}
				}
				if target != "" {
					ver := dispVer
					if ver == "" {
						ver = ExtractVersion(dispName)
					}
					return &Installation{
						Path:    target,
						Version: ver,
						Method:  "Windows Registry (Uninstall)",
					}
				}
			}
		}
	}

	return nil
}

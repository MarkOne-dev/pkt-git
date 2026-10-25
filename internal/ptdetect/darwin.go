package ptdetect

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MdfindFunc allows overriding mdfind in tests.
var MdfindFunc = func(query string) ([]string, error) {
	out, err := exec.Command("mdfind", query).Output()
	if err != nil {
		return nil, err
	}
	var results []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			results = append(results, line)
		}
	}
	return results, nil
}

func detectDarwin() *Installation {
	// 1. Query Spotlight via mdfind
	if results, err := MdfindFunc("kMDItemCFBundleIdentifier == 'com.cisco.PacketTracer*'"); err == nil && len(results) > 0 {
		for _, r := range results {
			if inst := inspectAppBundle(r); inst != nil {
				inst.Method = "macOS Spotlight (mdfind)"
				return inst
			}
		}
	}

	// 2. Scan standard Application directories
	home, _ := os.UserHomeDir()
	appGlobs := []string{
		"/Applications/Cisco Packet Tracer*.app",
		"/Applications/PacketTracer*.app",
	}
	if home != "" {
		appGlobs = append(appGlobs,
			filepath.Join(home, "Applications", "Cisco Packet Tracer*.app"),
			filepath.Join(home, "Applications", "PacketTracer*.app"),
		)
	}

	for _, pattern := range appGlobs {
		matches, err := filepath.Glob(pattern)
		if err == nil {
			for _, m := range matches {
				if inst := inspectAppBundle(m); inst != nil {
					return inst
				}
			}
		}
	}

	return nil
}

func inspectAppBundle(bundlePath string) *Installation {
	fi, err := os.Stat(bundlePath)
	if err != nil || !fi.IsDir() {
		return nil
	}

	// Check for executable inside Contents/MacOS
	macosDir := filepath.Join(bundlePath, "Contents", "MacOS")
	execTarget := bundlePath
	if entries, err := os.ReadDir(macosDir); err == nil {
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Name()), "packettracer") {
				execTarget = filepath.Join(macosDir, e.Name())
				break
			}
		}
	}

	ver := parsePlistVersion(filepath.Join(bundlePath, "Contents", "Info.plist"))
	if ver == "" {
		ver = ExtractVersion(bundlePath)
	}

	return &Installation{
		Path:    execTarget,
		Version: ver,
		Method:  "macOS Application Bundle",
	}
}

func parsePlistVersion(plistPath string) string {
	content, err := os.ReadFile(plistPath)
	if err != nil {
		return ""
	}
	text := string(content)
	keyIdx := strings.Index(text, "<key>CFBundleShortVersionString</key>")
	if keyIdx == -1 {
		keyIdx = strings.Index(text, "<key>CFBundleVersion</key>")
	}
	if keyIdx != -1 {
		rest := text[keyIdx:]
		start := strings.Index(rest, "<string>")
		end := strings.Index(rest, "</string>")
		if start != -1 && end != -1 && end > start+8 {
			return strings.TrimSpace(rest[start+8 : end])
		}
	}
	return ""
}

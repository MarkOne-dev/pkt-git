package ptdetect

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func detectLinux() *Installation {
	// 1. Scan XDG Desktop Entries (local user first, then system)
	home, _ := os.UserHomeDir()
	desktopDirs := []string{}
	if home != "" {
		desktopDirs = append(desktopDirs, filepath.Join(home, ".local", "share", "applications"))
	}
	desktopDirs = append(desktopDirs,
		"/usr/share/applications",
		"/usr/local/share/applications",
	)

	if inst := scanDesktopDirs(desktopDirs); inst != nil {
		return inst
	}

	// 2. Direct AppImage and filesystem directory patterns
	appImageGlobs := []string{
		"/usr/lib/packettracer/*.AppImage",
		"/usr/lib/packettracer/PacketTracer*",
		"/usr/lib/packettracer/packettracer*",
		"/opt/pt*/bin/PacketTracer*",
		"/opt/pt*/bin/packettracer*",
		"/opt/packettracer*/bin/PacketTracer*",
		"/opt/packettracer*/bin/packettracer*",
		"/opt/CiscoPacketTracer*/bin/PacketTracer*",
	}

	for _, pattern := range appImageGlobs {
		matches, err := filepath.Glob(pattern)
		if err == nil {
			for _, m := range matches {
				if fi, err := os.Stat(m); err == nil && !fi.IsDir() {
					ver := ExtractVersion(m)
					return &Installation{
						Path:    m,
						Version: ver,
						Method:  "Directory Scan",
					}
				}
			}
		}
	}

	// 3. Flatpak detection
	flatpakDirs := []string{
		"/var/lib/flatpak/app/com.cisco.PacketTracer",
	}
	if home != "" {
		flatpakDirs = append(flatpakDirs, filepath.Join(home, ".local", "share", "flatpak", "app", "com.cisco.PacketTracer"))
	}
	for _, fd := range flatpakDirs {
		if fi, err := os.Stat(fd); err == nil && fi.IsDir() {
			return &Installation{
				Path:    "flatpak run com.cisco.PacketTracer",
				Version: "",
				Method:  "Flatpak Container",
			}
		}
	}

	// 4. Snap detection
	for _, snapPath := range []string{"/snap/bin/packet-tracer", "/snap/bin/packettracer"} {
		if fi, err := os.Stat(snapPath); err == nil && !fi.IsDir() {
			return &Installation{
				Path:    snapPath,
				Version: "",
				Method:  "Snap Package",
			}
		}
	}

	return nil
}

func scanDesktopDirs(dirs []string) *Installation {
	for _, d := range dirs {
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".desktop") {
				continue
			}
			lowerName := strings.ToLower(e.Name())
			if strings.Contains(lowerName, "packet") || strings.Contains(lowerName, "pt") || strings.Contains(lowerName, "cisco") {
				desktopPath := filepath.Join(d, e.Name())
				if inst := ParseDesktopFile(desktopPath); inst != nil {
					return inst
				}
			}
		}
	}
	return nil
}

// ParseDesktopFile inspects a freedesktop .desktop entry and returns an Installation if it references Cisco Packet Tracer.
func ParseDesktopFile(path string) *Installation {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var execCmd, name, comment string
	isPT := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Name=") {
			name = strings.TrimPrefix(line, "Name=")
			if strings.Contains(strings.ToLower(name), "packet tracer") {
				isPT = true
			}
		} else if strings.HasPrefix(line, "Comment=") {
			comment = strings.TrimPrefix(line, "Comment=")
			if strings.Contains(strings.ToLower(comment), "packet tracer") {
				isPT = true
			}
		} else if strings.HasPrefix(line, "Exec=") {
			execCmd = strings.TrimPrefix(line, "Exec=")
		}
	}

	// Also check filename itself
	if strings.Contains(strings.ToLower(filepath.Base(path)), "packettracer") {
		isPT = true
	}

	if !isPT || execCmd == "" {
		return nil
	}

	// Clean up Exec command (strip quotes, args like %f, %u, etc.)
	fields := strings.Fields(execCmd)
	if len(fields) == 0 {
		return nil
	}
	cleanPath := strings.Trim(fields[0], `"'`)

	// Verify target executable exists
	if fi, err := os.Stat(cleanPath); err != nil || fi.IsDir() {
		// If cleanPath is just a command name in PATH, check LookPathFunc
		if p, err := LookPathFunc(cleanPath); err == nil && p != "" {
			cleanPath = p
		} else {
			return nil
		}
	}

	// Try extracting version from Name, Comment, or Path
	ver := ExtractVersion(name)
	if ver == "" {
		ver = ExtractVersion(comment)
	}
	if ver == "" {
		ver = ExtractVersion(cleanPath)
	}

	return &Installation{
		Path:    cleanPath,
		Version: ver,
		Method:  "XDG Desktop Entry (" + filepath.Base(path) + ")",
	}
}

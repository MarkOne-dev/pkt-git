package ptdetect

import (
	"os"
	"path/filepath"
)

func detectWindows() *Installation {
	// 1. Check registry entries
	if inst := queryWindowsRegistry(); inst != nil {
		return inst
	}

	// 2. Scan versioned Program Files directories
	programDirs := []string{}
	for _, envVar := range []string{"ProgramFiles", "ProgramFiles(x86)", "ProgramW6432"} {
		if val := os.Getenv(envVar); val != "" {
			programDirs = append(programDirs, val)
		}
	}
	if len(programDirs) == 0 {
		programDirs = []string{`C:\Program Files`, `C:\Program Files (x86)`}
	}

	return scanWindowsProgramDirs(programDirs)
}

func scanWindowsProgramDirs(dirs []string) *Installation {
	for _, baseDir := range dirs {
		patterns := []string{
			filepath.Join(baseDir, "Cisco Packet Tracer*", "bin", "PacketTracer.exe"),
			filepath.Join(baseDir, "Cisco Packet Tracer*", "bin", "PacketTracer"),
			filepath.Join(baseDir, "Cisco Packet Tracer*", "PacketTracer.exe"),
		}
		for _, pattern := range patterns {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				continue
			}
			for _, m := range matches {
				if fi, err := os.Stat(m); err == nil && !fi.IsDir() {
					ver := ExtractVersion(m)
					return &Installation{
						Path:    m,
						Version: ver,
						Method:  "Versioned Program Files Scan",
					}
				}
			}
		}
	}
	return nil
}

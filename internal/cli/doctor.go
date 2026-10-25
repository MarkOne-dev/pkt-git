package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MarkOne-dev/pkt-git/internal/cli/ui"
	"github.com/MarkOne-dev/pkt-git/internal/skills"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Inspect host environment prerequisites and tool configuration",
		Args:  exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			hasWarnings := false

			// 1. Git Installation & Version
			gitPath, err := exec.LookPath("git")
			if err != nil {
				ui.Error(out, "Git: not found in PATH (required)")
				fmt.Fprintln(out, "  Remediation: install Git via your system package manager (e.g. apt, dnf, brew, winget).")
				return fmt.Errorf("doctor: git is required but not installed")
			}

			gitVerBytes, err := exec.Command("git", "--version").Output()
			gitVer := strings.TrimSpace(string(gitVerBytes))
			if err != nil || gitVer == "" {
				gitVer = "installed"
			}
			ui.Success(out, "Git: %s (%s)", gitVer, gitPath)

			// 2. Git User Identity
			nameBytes, _ := exec.Command("git", "config", "user.name").Output()
			emailBytes, _ := exec.Command("git", "config", "user.email").Output()
			userName := strings.TrimSpace(string(nameBytes))
			userEmail := strings.TrimSpace(string(emailBytes))

			if userName != "" && userEmail != "" {
				ui.Success(out, "Git Identity: %s <%s>", userName, userEmail)
			} else {
				hasWarnings = true
				ui.Warning(out, "Git Identity: unconfigured user.name or user.email")
				fmt.Fprintln(out, "  Remediation: configure Git identity to prevent commit errors:")
				fmt.Fprintln(out, "    git config --global user.name \"Your Name\"")
				fmt.Fprintln(out, "    git config --global user.email \"you@example.com\"")
			}

			// 3. Binary & PATH
			exePath, err := os.Executable()
			if err == nil {
				exeDir := filepath.Dir(exePath)
				inPath := isDirInPATH(exeDir)
				if inPath {
					ui.Success(out, "Binary: %s (in PATH)", exePath)
				} else {
					hasWarnings = true
					ui.Warning(out, "Binary: %s (directory not in PATH)", exePath)
					fmt.Fprintf(out, "  Remediation: add %s to your PATH in ~/.bashrc or ~/.zshrc.\n", exeDir)
				}
			}

			// 4. Agent Skills
			home, err := os.UserHomeDir()
			if err == nil && home != "" {
				targets, _ := skills.Targets(home)
				var installed []string
				for _, t := range targets {
					if _, err := os.Stat(filepath.Join(t.Dir, "SKILL.md")); err == nil {
						installed = append(installed, t.Agent)
					}
				}
				if len(installed) > 0 {
					ui.Success(out, "Agent Skills: installed for %s", strings.Join(installed, ", "))
				} else {
					ui.Info(out, "Agent Skills: none installed (run 'pkt skills' to install for OpenCode, Claude, Gemini)")
				}
			}

			// 5. Cisco Packet Tracer detection
			ptPath := findPacketTracer()
			if ptPath != "" {
				ui.Success(out, "Cisco Packet Tracer: detected at %s", ptPath)
			} else {
				ui.Info(out, "Cisco Packet Tracer: not detected in common paths (optional for CLI; required to open/edit topologies)")
			}

			// 6. Go compiler availability
			goPath, err := exec.LookPath("go")
			if err == nil {
				goVerBytes, _ := exec.Command("go", "version").Output()
				goVer := strings.TrimSpace(string(goVerBytes))
				ui.Info(out, "Go SDK: %s (%s, optional for runtime)", goVer, goPath)
			} else {
				ui.Info(out, "Go SDK: not installed (optional; prebuilt binaries require no Go runtime)")
			}

			fmt.Fprintln(out)
			if hasWarnings {
				ui.Warning(out, "Environment has advisory warnings; see above for remediation.")
			} else {
				ui.Success(out, "All checks passed! Your environment is ready.")
			}

			return nil
		},
	}
}

func isDirInPATH(dir string) bool {
	pathEnv := os.Getenv("PATH")
	for _, p := range filepath.SplitList(pathEnv) {
		if filepath.Clean(p) == filepath.Clean(dir) {
			return true
		}
	}
	return false
}

func findPacketTracer() string {
	if p, err := exec.LookPath("PacketTracer"); err == nil {
		return p
	}
	if p, err := exec.LookPath("packettracer"); err == nil {
		return p
	}

	var candidates []string
	switch runtime.GOOS {
	case "linux":
		candidates = []string{
			"/opt/pt/bin/PacketTracer",
			"/usr/local/bin/PacketTracer",
			"/usr/local/bin/packettracer",
		}
	case "darwin":
		candidates = []string{
			"/Applications/Cisco Packet Tracer.app",
		}
	case "windows":
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")
		if programFiles != "" {
			candidates = append(candidates,
				filepath.Join(programFiles, "Cisco Packet Tracer", "bin", "PacketTracer.exe"),
			)
		}
		if programFilesX86 != "" {
			candidates = append(candidates,
				filepath.Join(programFilesX86, "Cisco Packet Tracer", "bin", "PacketTracer.exe"),
			)
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

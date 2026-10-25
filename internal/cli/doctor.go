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
	"github.com/MarkOne-dev/pkt-git/internal/ptdetect"
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
					if runtime.GOOS == "windows" {
						fmt.Fprintf(out, "  Remediation: add %s to your user PATH via System Properties or PowerShell.\n", exeDir)
					} else {
						fmt.Fprintf(out, "  Remediation: add 'export PATH=\"$PATH:%s\"' to your ~/.bashrc or ~/.zshrc.\n", exeDir)
					}
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
			ptInst := ptdetect.Find()
			if ptInst != nil {
				if ptInst.Version != "" {
					ui.Success(out, "Cisco Packet Tracer: v%s (%s via %s)", ptInst.Version, ptInst.Path, ptInst.Method)
				} else {
					ui.Success(out, "Cisco Packet Tracer: %s (via %s)", ptInst.Path, ptInst.Method)
				}
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

			// 7. Nerd Font Iconography
			fontName := detectNerdFont()
			if fontName != "" {
				ui.Success(out, "Nerd Font: %s (detected)", fontName)
			} else {
				ui.Info(out, "Nerd Font: not detected (optional; enables rich icons in welcome banner & terminal)")
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

func detectNerdFont() string {
	if fcPath, err := exec.LookPath("fc-list"); err == nil && fcPath != "" {
		out, err := exec.Command("fc-list", ":", "family").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.Contains(strings.ToLower(line), "nerd font") {
					parts := strings.Split(line, ",")
					return strings.TrimSpace(parts[0])
				}
			}
		}
	}

	home, _ := os.UserHomeDir()
	var searchDirs []string
	if home != "" {
		switch runtime.GOOS {
		case "linux":
			searchDirs = append(searchDirs,
				filepath.Join(home, ".local", "share", "fonts"),
				filepath.Join(home, ".fonts"),
				"/usr/share/fonts",
				"/usr/local/share/fonts",
			)
		case "darwin":
			searchDirs = append(searchDirs,
				filepath.Join(home, "Library", "Fonts"),
				"/Library/Fonts",
			)
		case "windows":
			localAppData := os.Getenv("LOCALAPPDATA")
			if localAppData != "" {
				searchDirs = append(searchDirs, filepath.Join(localAppData, "Microsoft", "Windows", "Fonts"))
			}
			winDir := os.Getenv("WINDIR")
			if winDir != "" {
				searchDirs = append(searchDirs, filepath.Join(winDir, "Fonts"))
			}
		}
	}

	for _, dir := range searchDirs {
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				lower := strings.ToLower(e.Name())
				if strings.Contains(lower, "nerd") {
					return strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
				}
			}
		}
	}

	return ""
}

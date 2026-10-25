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
	"github.com/MarkOne-dev/pkt-git/internal/i18n"
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
				ui.Error(out, "%s", i18n.T("DoctorGitMissing"))
				fmt.Fprintln(out, "  "+i18n.T("DoctorGitRemediation"))
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
				ui.Warning(out, "%s", i18n.T("DoctorGitIdentityMissing"))
				fmt.Fprintln(out, "  "+i18n.T("DoctorGitIdentityRemediation"))
				fmt.Fprintln(out, "    git config --global user.name \"Your Name\"")
				fmt.Fprintln(out, "    git config --global user.email \"you@example.com\"")
			}

			// 3. Binary & PATH
			exePath, err := os.Executable()
			if err == nil {
				exeDir := filepath.Dir(exePath)
				inPath := isDirInPATH(exeDir)
				if inPath {
					ui.Success(out, "%s", i18n.T("DoctorBinaryInPath", map[string]interface{}{"Path": exePath}))
				} else {
					hasWarnings = true
					ui.Warning(out, "%s", i18n.T("DoctorBinaryNotInPath", map[string]interface{}{"Path": exePath}))
					if runtime.GOOS == "windows" {
						fmt.Fprintf(out, "  %s\n", i18n.T("DoctorBinaryRemediationWindows", map[string]interface{}{"Dir": exeDir}))
					} else {
						fmt.Fprintf(out, "  %s\n", i18n.T("DoctorBinaryRemediationUnix", map[string]interface{}{"Dir": exeDir}))
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
					ui.Success(out, "%s", i18n.T("DoctorSkillsInstalled", map[string]interface{}{"Agents": strings.Join(installed, ", ")}))
				} else {
					ui.Info(out, "%s", i18n.T("DoctorSkillsNone"))
				}
			}

			// 5. Cisco Packet Tracer detection
			ptInst := ptdetect.Find()
			if ptInst != nil {
				details := ""
				if ptInst.Version != "" {
					details = fmt.Sprintf("v%s (%s via %s)", ptInst.Version, ptInst.Path, ptInst.Method)
				} else {
					details = fmt.Sprintf("%s (via %s)", ptInst.Path, ptInst.Method)
				}
				ui.Success(out, "%s", i18n.T("DoctorPacketTracerDetected", map[string]interface{}{"Details": details}))
			} else {
				ui.Info(out, "%s", i18n.T("DoctorPacketTracerMissing"))
			}

			// 6. Go compiler availability
			goPath, err := exec.LookPath("go")
			if err == nil {
				goVerBytes, _ := exec.Command("go", "version").Output()
				goVer := strings.TrimSpace(string(goVerBytes))
				ui.Info(out, "%s", i18n.T("DoctorGoInstalled", map[string]interface{}{"Details": fmt.Sprintf("%s (%s)", goVer, goPath)}))
			} else {
				ui.Info(out, "%s", i18n.T("DoctorGoMissing"))
			}

			// 7. Nerd Font Iconography
			fontName := detectNerdFont()
			if fontName != "" {
				ui.Success(out, "%s", i18n.T("DoctorNerdFontDetected", map[string]interface{}{"Name": fontName}))
			} else {
				ui.Info(out, "%s", i18n.T("DoctorNerdFontMissing"))
			}

			fmt.Fprintln(out)
			if hasWarnings {
				ui.Warning(out, "%s", i18n.T("DoctorAdvisoryWarnings"))
			} else {
				ui.Success(out, "%s", i18n.T("DoctorAllPassed"))
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

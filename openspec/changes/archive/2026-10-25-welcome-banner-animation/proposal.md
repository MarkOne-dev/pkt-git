# Proposal: Welcome Banner, Animated Spinner, and Nerd Font Integration

## Why

Currently, installing and running `pkt-git` presents only static terminal text without visual branding or animated feedback. Furthermore, modern CLI tools increasingly benefit from sharp developer iconography (Nerd Fonts) rather than generic, inconsistent color emojis that render unpredictably across terminal emulators. Introducing a styled ASCII art banner, an animated terminal spinner, a dedicated `pkt welcome` onboarding command, and automated cross-platform installation and configuration of a recommended Nerd Font provides an engaging, professional developer experience while preserving script-safety and strict `NO_COLOR` compliance.

## What Changes

- **ASCII Art Branding Banner**: Implement a styled Lipgloss block ASCII art banner for `pkt-git` with brand coloring, centered subtitle, runtime metadata, and automatic plain-text fallback for non-TTY or `NO_COLOR` environments.
- **Animated Terminal Spinner**: Add an interactive spinner component in `internal/cli/ui` to display animated progress during multi-step tasks, degrading cleanly to static symbols in non-interactive pipes.
- **Nerd Font Iconography in CLI**: Replace emojis with dedicated Nerd Font glyphs (e.g. rocket, git, folder, check, cross, terminal) in UI output and welcome screens, with clean plain-text fallback.
- **Automated Nerd Font Installation Dependency**:
  - In `scripts/install.sh`: Detect and automatically install `JetBrainsMono Nerd Font` into `${HOME}/.local/share/fonts/NerdFonts` on Linux (running `fc-cache`) and `${HOME}/Library/Fonts` on macOS; configure GNOME Terminal font profile where available.
  - In `scripts/install.ps1`: Detect and install `JetBrainsMono Nerd Font` into `$env:LOCALAPPDATA\Microsoft\Windows\Fonts`, register in the Windows Registry, broadcast font changes, and configure Windows Terminal `settings.json` defaults where present.
- **`pkt doctor` Nerd Font Verification**: Add a diagnostic check to `pkt doctor` verifying whether a Nerd Font is detected in the system.
- **`pkt welcome` Command**: Add a new user-facing `pkt welcome` command that displays the branded banner, a brief animated environment check, and an actionable 3-step quickstart onboarding guide.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli`: Add requirements for the `pkt welcome` onboarding command, ASCII art branding banner, animated progress spinners, and Nerd Font icon integration with environment diagnostics.

## Impact

- **Affected Packages**: `internal/cli`, `internal/cli/ui`, `scripts/install.sh`, `scripts/install.ps1`.
- **User Experience**: Users get an integrated, beautiful terminal interface with automatically configured fonts and consistent icons across Linux, macOS, and Windows.
- **Compatibility**: All animations and ANSI escapes automatically disable when output is piped, redirected, or running under `NO_COLOR=1` / `TERM=dumb`.

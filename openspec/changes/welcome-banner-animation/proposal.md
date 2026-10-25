# Proposal: Welcome Banner, Animated Spinner, and Onboarding Experience

## Why

Currently, installing and running `pkt-git` presents only static terminal text without visual branding or animated feedback. Users installing the tool for the first time receive no welcoming visual onboarding, and long-running operations (such as environment diagnostics or installation steps) lack lively visual progress indicators. Introducing a styled ASCII art banner, an animated terminal spinner, and a dedicated `pkt welcome` onboarding command provides an engaging, professional developer experience while preserving script-safety and strict `NO_COLOR` compliance.

## What Changes

- **ASCII Art Branding Banner**: Implement a styled Lipgloss ASCII art banner for `pkt-git` with brand coloring, subtitle, and an automatic plain-text fallback for non-TTY or `NO_COLOR` environments.
- **Animated Terminal Spinner**: Add an interactive spinner component in `internal/cli/ui` to display animated progress during multi-step tasks (e.g. running diagnostics or discovering topologies), degrading cleanly to static symbols in non-interactive pipes.
- **`pkt welcome` Command**: Add a new user-facing `pkt welcome` command that displays the branded banner, a brief animated environment check, and an actionable 3-step quickstart onboarding guide.
- **Installer Integration**: Update `scripts/install.sh` and `scripts/install.ps1` to showcase the welcome banner and onboarding screen upon successful installation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli`: Add requirements for the `pkt welcome` onboarding command, ASCII art branding banner, and animated progress spinners adhering to TTY and `NO_COLOR` rules.

## Impact

- **Affected Packages**: `internal/cli`, `internal/cli/ui`, `scripts/install.sh`, `scripts/install.ps1`.
- **User Experience**: First-time users see an engaging welcome presentation with immediate next steps.
- **Compatibility**: All animations and ANSI escapes automatically disable when output is piped, redirected, or running under `NO_COLOR=1` / `TERM=dumb`.

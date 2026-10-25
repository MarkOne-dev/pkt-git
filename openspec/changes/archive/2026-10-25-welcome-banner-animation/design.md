# Design: Welcome Banner, Animated Spinner, and Nerd Font Integration

## Context

See `proposal.md` for background. The `pkt-git` CLI already includes `internal/cli/ui`, which utilizes Charmbracelet Lipgloss for styling, `mattn/go-isatty` for interactive terminal detection, and handles `NO_COLOR` and `TERM=dumb` conventions.

While the block banner and inline spinner provide structure and animation, using generic emojis (like `🚀`) can lead to inconsistent rendering or missing glyphs depending on terminal font configurations. By establishing a recommended developer font (`JetBrainsMono Nerd Font`), automating its installation as a setup prerequisite across Linux, macOS, and Windows, and configuring terminal defaults where feasible, the CLI can guarantee sharp, professional glyphs across all platforms.

## Goals / Non-Goals

**Goals:**
- Provide a branded block ASCII art banner (`PKT-GIT`) in `internal/cli/ui` styled with Lipgloss, centered with its subtitle and runtime metadata.
- Implement an inline, non-blocking terminal spinner (`ui.Spinner`) in `internal/cli/ui` with automatic static fallback for non-TTY / `NO_COLOR` environments.
- Use sharp Nerd Font glyphs (e.g. `\uf135` 󱄵, `\uf120` , `\ue702` ) instead of color emojis in UI output and welcome guidance.
- Automate user-level installation of `JetBrainsMono Nerd Font` in `scripts/install.sh` (Linux & macOS) and `scripts/install.ps1` (Windows) without requiring administrator/root privileges.
- Configure terminal font settings automatically where possible (e.g. Windows Terminal `settings.json` and GNOME Terminal `gsettings`).
- Enhance `pkt doctor` with a diagnostic check verifying whether a Nerd Font is installed on the host machine.

**Non-Goals:**
- Do not make missing fonts a fatal failure: if offline or font installation is skipped, the CLI falls back gracefully to standard Unicode symbols and plain text labels.
- Do not introduce full-screen TUI frameworks (e.g. alternate screen buffers); the UI remains inline and streaming-friendly.

## Decisions

### 1. Cross-Platform Font Installation Strategy

- **Linux**:
  - Target directory: `${HOME}/.local/share/fonts/NerdFonts` (XDG user font directory, requires no `sudo`).
  - Method: Download `JetBrainsMono.tar.xz` from GitHub releases (`ryanoasis/nerd-fonts`), extract `.ttf` files, and trigger `fc-cache -f`.
  - Terminal auto-config: If `gsettings` is present (GNOME Terminal default profile), set `use-system-font false` and `font 'JetBrainsMono Nerd Font 11'`.
- **macOS**:
  - Target directory: `${HOME}/Library/Fonts` (native user font folder, automatically detected by macOS font subsystem).
  - Method: Prefer `brew install --cask font-jetbrains-mono-nerd-font` if Homebrew is available; otherwise download and extract directly into `~/Library/Fonts`.
- **Windows**:
  - Target directory: `$env:LOCALAPPDATA\Microsoft\Windows\Fonts`.
  - Method: Download `JetBrainsMono.zip`, extract `.ttf` files, copy to user fonts directory, register entries under `HKCU:\Software\Microsoft\Windows NT\CurrentVersion\Fonts`, and broadcast `WM_FONTCHANGE`.
  - Terminal auto-config: If Windows Terminal `settings.json` exists, configure `"face": "JetBrainsMono Nerd Font"` under `profiles.defaults.font`.

### 2. Nerd Font Glyph Integration in `internal/cli/ui`

- **Choice**: Replace emojis with Nerd Font icons:
  - Welcome title: `\uf135` (󱄵 / cohete Nerd Font) or `\uf120` ( terminal).
  - Git integration: `\ue702` ( git).
  - Check / Pass: `\uf00c` () or `✔`.
  - Warnings: `\uf071` () or `⚠`.
  - Error: `\uf00d` () or `✖`.
  - Info: `\uf05a` () or `ℹ`.
- **Fallback**: When `ColorEnabled(w)` is false, all icons map to clean text badges (`[ok]`, `[warn]`, `[error]`, `[info]`, `->`).

### 3. Diagnostic Verification in `pkt doctor`

- **Choice**: In `internal/cli/doctor.go`, execute `fc-list : family` on Unix-like systems and inspect registry/fonts on Windows to detect if any font family matching `*Nerd Font*` is present.
- **Reporting**: Report status with green check if found, or actionable advice if not found.

## Risks / Trade-offs

- **[Network latency or offline installation]**:
  - *Risk*: Downloading fonts adds network traffic during installer execution.
  - *Mitigation*: The installer skips font installation if a Nerd Font is already detected, or if `--skip-fonts` / `PKTG_SKIP_FONTS=1` is provided. If download fails (e.g. offline), it warns and continues without aborting the CLI installation.

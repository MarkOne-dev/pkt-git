# Design: Welcome Banner, Animated Spinner, and Onboarding Experience

## Context

See `proposal.md` for background. The `pkt-git` CLI already includes `internal/cli/ui`, which utilizes Charmbracelet Lipgloss for styling, `mattn/go-isatty` for interactive terminal detection, and handles `NO_COLOR` and `TERM=dumb` conventions. 

Currently, all CLI feedback is static. When users run `install.sh` or `install.ps1`, the output is sequential plain text. Introducing visual branding and animated progress indicators will elevate developer experience without adding external heavy dependencies.

## Goals / Non-Goals

**Goals:**
- Provide a branded ASCII art banner (`PKT-GIT`) in `internal/cli/ui` styled with Lipgloss, fitting comfortably within standard 80-column terminals.
- Implement an inline, non-blocking terminal spinner (`ui.Spinner`) in `internal/cli/ui` using goroutines and ANSI carriage returns, with automatic static fallback for non-TTY / `NO_COLOR` environments.
- Create a dedicated user-facing command `pkt welcome` that displays the banner, runs animated environment validation, and shows a 3-step quickstart card.
- Integrate the welcome experience into `scripts/install.sh` and `scripts/install.ps1` upon completion of installation.

**Non-Goals:**
- Do not introduce full-screen TUI frameworks (e.g. alternate screen buffers); the UI remains inline and streaming-friendly.
- Do not disrupt non-interactive script automation or CI runners.

## Decisions

### 1. Compact ASCII Art Banner in `internal/cli/ui/banner.go`

- **Choice**: Design a compact 5-line ASCII banner for `PKT-GIT` styled with Lipgloss (cyan/blue gradient and subtitle).
- **Rationale**: Fits standard terminals without wrapping, immediately communicates visual brand identity, and strips ANSI codes cleanly when `NO_COLOR` or non-TTY is detected.

### 2. Stream-Friendly Inline Spinner in `internal/cli/ui/spinner.go`

- **Choice**: Implement `ui.Spinner` using a background goroutine ticking at 80ms intervals over Unicode braille frames (`⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`).
- **Control**:
  - `s := ui.NewSpinner(out, "Running diagnostics...")`
  - `s.Start()` / `s.Success("All checks passed")` / `s.Fail("Check failed")` / `s.Stop()`
- **Fallback**: If `!ui.IsTTY(out)` or `NO_COLOR` is active, animation frames are skipped entirely. The start message and completion status are written as standard single-line messages (`[ok]` or `✔`).

### 3. Dedicated `pkt welcome` Command in `internal/cli/welcome.go`

- **Choice**: Expose `pkt welcome` in the CLI command hierarchy under a dedicated "Getting Started" or "Core Workflow" section.
- **Components displayed**:
  1. Branded ASCII Banner & Version badge.
  2. Animated spinner performing a quick system health check (`git` and `pkt` readiness).
  3. Actionable 3-step quickstart guide:
     - `1. pkt init` (configure repository & auto-export)
     - `2. git commit` (automatic snapshot synchronization)
     - `3. pkt doctor` (inspect environment & agent skills)

### 4. Installer Integration

- **Choice**: Update `scripts/install.sh` and `scripts/install.ps1` to execute `"${TARGET}" welcome` as the final celebratory step of installation.
- **Rationale**: Gives every new user an immediate, delightful onboarding presentation without adding complex shell-specific animation scripts.

## Risks / Trade-offs

- **[Carriage return pollution in CI / logs]**:
  - *Risk*: Piped outputs or CI logs could become filled with `\r` control characters.
  - *Mitigation*: The spinner checks `ui.IsTTY(w)` before emitting any frame updates or carriage returns. If false, it acts as a passive logger.

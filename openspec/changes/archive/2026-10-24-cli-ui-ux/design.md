# Design

## Context

The current `internal/cli` package relies on raw `fmt.Fprintf` calls directly to `cmd.OutOrStdout()` or `cmd.ErrOrStderr()`, using Go standard library's `text/tabwriter` for tabular output. See [proposal.md](proposal.md) for background and motivation.

The project requires static, cross-platform binaries without C runtime dependencies (`CGO_ENABLED=0`). Any terminal presentation upgrade must strictly preserve this invariant while respecting non-interactive script safety.

## Goals / Non-Goals

**Goals:**
- Provide a centralized UI package (`internal/cli/ui`) that handles terminal detection, colors, icons, tables, and `NO_COLOR` compliance.
- Keep the binary completely static and free of cgo.
- Ensure that non-TTY executions (pipes, redirects, CI) produce clean, escape-free plain text.
- Reorganize `pkt --help` into functional command groups via native Cobra APIs.
- Upgrade `pkt inspect` with structured tables, category rollups, and `--type` filtering.
- Format file and memory sizes into human-readable strings (e.g., `5.62 MB`).

**Non-Goals:**
- Full-screen TUI apps (e.g., Bubbletea alternative screen): commands remain standard CLI utilities that stream to stdout/stderr and exit.
- Interactive user prompts or confirmations: all actions remain scriptable and flag-driven.
- Modifying underlying codec, XML editing, or hook algorithms.

## Decisions

### Decision 1: Use `charmbracelet/lipgloss` for styling and tables
- **Rationale**: `lipgloss` is the modern pure-Go standard for terminal layout. It automatically detects terminal capabilities, handles color profiles (ANSI, 256, TrueColor), supports Windows VT mode out-of-the-box, and honors `NO_COLOR`.
- **Alternatives considered**:
  - *Raw ANSI escape strings*: Fragile, error-prone on Windows, and cumbersome to toggle cleanly based on TTY status.
  - *`fatih/color`*: Good for basic text coloring, but lacks modern table formatting and layout primitives.

### Decision 2: Use `dustin/go-humanize` for byte and count formatting
- **Rationale**: Tiny, zero-dependency Go library that converts raw byte counters (e.g., `5894851`) into clean, human-readable strings (`5.6 MB`).
- **Alternatives considered**: Writing custom byte-formatting helpers; using `go-humanize` avoids reinventing standard number formatting.

### Decision 3: Encapsulate presentation under `internal/cli/ui`
- **Rationale**: Keeps Cobra command handlers clean. The UI package exposes high-level helpers:
  - `ui.Success(w, format, args...)`
  - `ui.Warning(w, format, args...)`
  - `ui.Error(w, format, args...)`
  - `ui.RenderTable(w, headers, rows)`
  - `ui.FormatBytes(n)`
- **Alternatives considered**: Inlining lipgloss styles inside each `internal/cli/*.go` file; rejected because it scatters style definitions and makes testing harder.

### Decision 4: Use Cobra Command Groups for top-level help
- **Rationale**: Cobra natively supports `cobra.Group` and `cmd.AddGroup(...)`. This allows categorizing commands into:
  - `core`: Core Workflow (`export`, `build`, `sync`, `init`)
  - `inspection`: Inspection & Editing (`inspect`, `edit`, `verify`)
  - `collab`: Collaboration & Tools (`resolve`, `skills`, `version`)
- **Alternatives considered**: Custom template override in Cobra; using native command groups is cleaner, standard, and maintains Cobra's built-in flag rendering.

## Risks / Trade-offs

- **[Risk] CI / Script breakage due to ANSI sequences** → *Mitigation*: The `internal/cli/ui` package inspects standard output/error file descriptors using `mattn/go-isatty` (or standard `os.Stdout` TTY checks) and environment variables (`NO_COLOR`, `TERM=dumb`). If either check fails, all styles and colors are stripped.
- **[Risk] Binary size growth** → *Mitigation*: `lipgloss` and `go-humanize` add less than 350 KB to the stripped binary (`dist/pkt*`), well within acceptable limits.
- **[Risk] Test assertions on command outputs** → *Mitigation*: Tests inject `bytes.Buffer` for command stdout/stderr. Since `bytes.Buffer` is not a TTY, tests naturally verify the clean plain-text output by default.

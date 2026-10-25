# Proposal

## Why

The `pkt-git` CLI currently produces unstyled, monochrome output with raw byte counts, basic `[ok]` prefixes, and flat tables that flood the terminal on real-world topologies (such as 107-device networks). Elevating the terminal presentation with semantic colors, Unicode tables, device role badges, grouped help commands, and humanized metrics will significantly enhance user experience and readability without breaking script safety or CI/CD pipelines.

## What Changes

- Add automatic TTY detection and `NO_COLOR` / `TERM=dumb` compliance to guarantee plain, escape-free text in non-interactive environments, pipes, and CI.
- Modernize command outputs (`export`, `build`, `init`, `sync`, `verify`, `skills`, `version`) using semantic status indicators (e.g., `✔`, `⚠`, `✖`), human-readable byte sizes (e.g., `5.62 MB`), and structured metadata summaries.
- Redesign `pkt inspect` to render structured Unicode tables with colored device role badges (`[Router]`, `[Switch]`, `[Server]`, `[Host]`), category totals, and an optional `--type` filter.
- Group commands in `pkt --help` into logical functional sections (Core Workflow, Inspection & Editing, Git Integration & Agents) using Cobra command groups.
- Format runtime and usage errors with structured context (operation, target path, cause, and remediation hint) while strictly preserving existing exit codes (0, 1, 2).
- Add colorized configuration diff previews for `pkt edit` mutations (`setiface`, `chgline`, `insertafter`).
- Introduce pure-Go presentation dependencies (`github.com/charmbracelet/lipgloss` and `github.com/dustin/go-humanize`), preserving the strict `CGO_ENABLED=0` static binary build policy.

## Capabilities

### New Capabilities

*(None)*

### Modified Capabilities

- `cli`: Add requirements for TTY-aware visual formatting, `NO_COLOR` compliance, grouped command help, enhanced tabular inspection, and structured diagnostics while preserving script safety and stable exit codes.

## Impact

- **Dependencies**: Add `github.com/charmbracelet/lipgloss` and `github.com/dustin/go-humanize` to `go.mod`. Both are pure Go with zero C/cgo dependencies.
- **Codebase**: Updates to `internal/cli/` commands and addition of a helper package or module (e.g. `internal/cli/ui`) for shared styling, badges, and TTY checks.
- **Compatibility**: Non-breaking for automated workflows; scripts and pipes reading stdout/stderr will receive clean, escape-free plain text when stdout is not a TTY or when `NO_COLOR` is set.

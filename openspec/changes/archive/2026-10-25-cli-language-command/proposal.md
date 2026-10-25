# Proposal

## Why

While `pkt-git` supports Spanish and English localization through the `--lang` flag and environment variables, users need to pass `--lang` repeatedly or alter system environment variables to maintain their preference. Introducing a dedicated `pkt language` (and `pkt lang`) command provides a polished, interactive experience for viewing and switching interface languages with persistent storage in Git configuration.

## What Changes

- Add a new Cobra command `pkt language [code]` with alias `pkt lang` registered under Collaboration & Tools.
- Implement polished Lipgloss terminal UI with two operating modes:
  - **Interactive TTY Mode (no arguments)**: Displays a styled language selector showing the currently active language, numbered choices, and prompt-based selection.
  - **Direct Argument Mode (`pkt lang es`, `pkt lang en`)**: Immediately validates and persists the language preference.
  - **Non-Interactive Mode**: Prints clean plain-text information about the active language and available locales without blocking on stdin.
- Support `--global` (default, stored in user's global `~/.gitconfig`) and `--local` (stored in `.git/config` for repository-specific overrides).
- Update `internal/i18n.DetectLanguage()` to consult `git config pkt.language` before falling back to system environment variables (`LANG`, `LC_ALL`).

## Capabilities

### Modified Capabilities

- `cli`: Extend the CLI contract to include the `language` (alias `lang`) command for interactive selection and persistent configuration of interface language.

## Impact

- `internal/config/config.go`: Add `GitConfigLanguage = "pkt.language"`.
- `internal/cli/language.go`: New command implementation with Lipgloss UI and Cobra bindings.
- `internal/cli/root.go`: Register `newLanguageCmd()` in root command tree.
- `internal/i18n/i18n.go`: Check `git config pkt.language` in `DetectLanguage()`.
- Backward compatibility: Completely non-breaking; purely additive command.

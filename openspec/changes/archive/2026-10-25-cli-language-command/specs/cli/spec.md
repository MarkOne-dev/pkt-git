# Spec Delta

## ADDED Requirements

### Requirement: Manage interface language with language command

The CLI SHALL expose a `language` command (with alias `lang`) to view, switch, and persist the preferred interface language. The command MUST accept an optional language code argument (`en`, `es`), support `--global` (default) and `--local` scope flags, and persist the preference in Git configuration (`pkt.language`). When run interactively without arguments in a TTY, it MUST present an interactive selection menu.

#### Scenario: View current language in non-interactive environment
- **WHEN** a user or script runs `pkt language` in a non-interactive stream
- **THEN** the command prints the active language code and available languages without waiting for standard input
- **AND** it exits with status 0

#### Scenario: Set language with direct code argument
- **WHEN** a user runs `pkt language es` or `pkt lang es`
- **THEN** the command persists `es` in `git config --global pkt.language`
- **AND** it displays a confirmation indicating the language was updated
- **AND** it exits with status 0

#### Scenario: Reject unsupported language code
- **WHEN** a user runs `pkt language de`
- **THEN** the command writes an error indicating `de` is not supported
- **AND** it displays the list of supported language codes (`en`, `es`)
- **AND** it exits with status 2

#### Scenario: Interactive language selection
- **WHEN** a user runs `pkt language` in an interactive terminal (TTY)
- **THEN** the CLI renders an interactive menu displaying available languages and marks the currently active language
- **AND** it prompts the user to select an option

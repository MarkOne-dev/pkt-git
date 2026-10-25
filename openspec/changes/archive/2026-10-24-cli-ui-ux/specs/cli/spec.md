# Spec Delta

## ADDED Requirements

### Requirement: Format terminal output according to environment and TTY

The CLI SHALL render semantic color highlights, status badges, and human-readable numbers when output is directed to an interactive terminal. When output is redirected, piped, or when `NO_COLOR` or `TERM=dumb` is set, the CLI MUST emit plain text without ANSI escape sequences.

#### Scenario: Interactive terminal rendering
- **WHEN** a command executes in an interactive terminal without `NO_COLOR`
- **THEN** output includes status symbols, styled labels, and human-formatted sizes (e.g., KB/MB)

#### Scenario: Non-interactive pipe or redirection
- **WHEN** a command output is piped or redirected to a file
- **THEN** output contains only plain ASCII/UTF-8 text without ANSI escape sequences

#### Scenario: NO_COLOR environment variable set
- **WHEN** the `NO_COLOR` environment variable is non-empty
- **THEN** output contains only unstyled text without ANSI escape sequences

### Requirement: Group command help by functional category

Top-level help (`pkt --help`) SHALL organize public commands into distinct functional groups including Core Workflow, Inspection & Editing, and Git Integration & Agents.

#### Scenario: View grouped top-level help
- **WHEN** a user runs `pkt --help`
- **THEN** the commands are organized and displayed under labeled functional sections
- **AND** hidden hook entry points remain omitted

### Requirement: Provide structured tabular inspection

`pkt inspect` SHALL display devices in an aligned tabular layout with device role indicators and summary counts. When a type filter flag is supplied, it MUST display only devices matching the specified type.

#### Scenario: Inspect full topology summary
- **WHEN** a user runs `pkt inspect <file.pkt>`
- **THEN** the output displays an aligned table with device names, models, types, and config lines
- **AND** it appends a summary breakdown of totals by device category

#### Scenario: Inspect with type filter
- **WHEN** a user runs `pkt inspect <file.pkt> --type Router`
- **THEN** the table displays only devices whose type matches Router

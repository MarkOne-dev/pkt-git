# cli Specification

## Purpose

Defines the command-line contract for pkt-git: discoverable commands,
predictable argument handling, stable exit codes, version reporting, and
management of the bundled agent skill.

## Requirements

### Requirement: Provide a discoverable command tree

The `pkt` command SHALL expose the commands `init`, `export`, `build`,
`inspect`, `edit`, `sync`, `resolve`, `verify`, `skills`, and `version`.
Running `pkt --help` or `pkt help` MUST list the public commands with short
descriptions and MUST omit hidden hook entry points.

#### Scenario: Show top-level help

- **WHEN** a user runs `pkt --help`
- **THEN** the output lists every public command and its short description
- **AND** the hidden hook commands are not listed
- **AND** the command exits successfully

#### Scenario: Show subcommand help

- **WHEN** a user runs `pkt edit --help`
- **THEN** the output explains the edit subcommands and their arguments
- **AND** the command exits successfully

### Requirement: Validate arguments and return stable exit codes

Every command SHALL validate its required arguments and supported flags before
performing work. The CLI MUST return exit code 0 on success, exit code 2 for a
usage error, and exit code 1 for a runtime failure.

#### Scenario: Missing required argument

- **WHEN** a user runs `pkt export` without the required paths
- **THEN** the command prints usage guidance for `export`
- **AND** it exits with status 2
- **AND** it does not write any file

#### Scenario: Unknown command

- **WHEN** a user runs `pkt does-not-exist`
- **THEN** the command reports the unknown command
- **AND** it exits with status 2

#### Scenario: Runtime failure

- **WHEN** a valid command cannot complete because of a decode, filesystem,
  or Git error
- **THEN** the command reports the failure
- **AND** it exits with status 1

### Requirement: Report version and build metadata

`pkt version` and `pkt --version` SHALL print the version and available build
metadata without requiring a repository or any other command context.

#### Scenario: Version output

- **WHEN** a user runs `pkt version`
- **THEN** the output contains the release version
- **AND** the command exits successfully
- **AND** it works outside a Git repository

### Requirement: Expose hidden hook entry points

The `pre-commit`, `post-merge`, and `post-checkout` commands SHALL remain
invokable for generated hooks but MUST be hidden from normal help output.
They MUST reject unsupported extra arguments according to their documented
hook contract.

#### Scenario: Invoke a hook entry point

- **WHEN** a generated hook invokes `pkt pre-commit`
- **THEN** the command performs the pre-commit automation
- **AND** it is not advertised in `pkt --help`

#### Scenario: Invalid hook arguments

- **WHEN** a hook entry point receives arguments outside its contract
- **THEN** the command exits non-zero with an actionable error
- **AND** it does not modify the repository

### Requirement: Manage the bundled agent skill

`pkt skills` SHALL install the bundled pkt-git skill into the global skill
directories for OpenCode, Claude Code, and Gemini CLI. `pkt skills --remove`
SHALL remove only the pkt-git skill from those locations. Both operations MUST
be idempotent and MUST report each agent's outcome.

#### Scenario: Install the skill

- **WHEN** a user runs `pkt skills`
- **THEN** the pkt-git skill file is present in each supported agent's global
  skill directory
- **AND** the command reports success for each agent

#### Scenario: Remove the skill

- **WHEN** a user runs `pkt skills --remove`
- **THEN** only the pkt-git skill directories are removed
- **AND** other skills in those locations are untouched

#### Scenario: Missing bundled skill

- **WHEN** the bundled skill asset is unavailable
- **THEN** the command exits non-zero with a reinstall instruction
- **AND** it does not delete an existing installed skill

### Requirement: Remain non-interactive and script-safe

The CLI MUST NOT prompt for input during normal command execution. All
decisions that affect data, such as conflict resolution policy, MUST be
supplied explicitly through arguments or flags.

#### Scenario: Conflict policy required

- **WHEN** a user runs `pkt resolve` without `--ours` or `--theirs` during a
  conflict
- **THEN** the command reports the conflict and exits without choosing a side
- **AND** it does not wait for interactive input

### Requirement: Produce actionable errors

Command failures SHALL identify the operation and the relevant path or
revision, and MUST NOT print a stack trace or expose credentials. Errors
intended for users MUST be written to standard error.

#### Scenario: Actionable failure

- **WHEN** a command fails because a file does not exist
- **THEN** the error names the command and the missing path
- **AND** the error is written to standard error
- **AND** no stack trace is printed

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

### Requirement: Diagnose environment prerequisites with doctor

The CLI SHALL expose a `doctor` command that inspects the host environment and reports the status of Git installation, configured Git user identity (`user.name` and `user.email`), binary availability on `PATH`, terminal color support, agent skills, and a comprehensive discovery of installed Cisco Packet Tracer instances across Linux (AppImage, XDG desktop entries, Flatpak, Snap, standard paths), Windows (Registry and versioned directories), and macOS (Spotlight and Applications).

#### Scenario: Run doctor in a healthy environment
- **WHEN** a user runs `pkt doctor` on a fully configured machine
- **THEN** the command outputs status indicators for Git, Git identity, and PATH availability
- **AND** it exits with status 0

#### Scenario: Run doctor with unconfigured Git identity
- **WHEN** a user runs `pkt doctor` and `user.name` or `user.email` is not configured
- **THEN** the command outputs an advisory warning indicating the missing configuration
- **AND** it displays remediation guidance for `git config`
- **AND** it exits with status 0

#### Scenario: Detect Cisco Packet Tracer from AppImage and desktop entries on Linux
- **WHEN** a user runs `pkt doctor` on a Linux system where Cisco Packet Tracer is installed as an AppImage or registered via XDG `.desktop` file
- **THEN** the command parses the executable path and detected version
- **AND** it displays a success indicator with the path and version details without false negative warnings

#### Scenario: Detect Cisco Packet Tracer from versioned install paths or Windows registry
- **WHEN** a user runs `pkt doctor` on a Windows system where Cisco Packet Tracer is installed in a versioned folder (e.g. `Cisco Packet Tracer 8.2.2`) or registered in Windows App Paths/Uninstall registry keys
- **THEN** the command discovers the executable path through dynamic globbing and registry lookups
- **AND** it displays a success indicator with the detected executable location

#### Scenario: Verify and advise on binary PATH inclusion
- **WHEN** a user runs `pkt doctor` and the current `pkt` binary directory is not present in the user's `$PATH`
- **THEN** the command displays an advisory warning and shell-specific remediation commands

### Requirement: Display welcome onboarding screen

`pkt welcome` SHALL display an onboarding overview comprising an ASCII art branding banner, current binary version, agent skills status, and a 3-step quickstart guide. The command MUST exit with code 0 on completion and MUST be advertised in the top-level help menu.

#### Scenario: Interactive welcome execution
- **WHEN** a user runs `pkt welcome` in an interactive terminal
- **THEN** the command renders the styled ASCII banner with brand color
- **AND** it displays the version and 3-step quickstart guidance
- **AND** it exits with status 0

#### Scenario: Non-interactive or NO_COLOR welcome execution
- **WHEN** `pkt welcome` is piped, redirected, or executed with `NO_COLOR=1`
- **THEN** it outputs plain text and an unstyled banner without ANSI escape codes
- **AND** it exits with status 0

### Requirement: Render animated terminal progress indicators

The CLI SHALL provide an animated terminal spinner for long-running operations when executing in an interactive TTY without `NO_COLOR`. In non-interactive environments, pipelines, or when `NO_COLOR` is active, the CLI MUST omit spinner animation frames and print only the static operation message and completion status.

#### Scenario: Interactive spinner execution
- **WHEN** an operation with a spinner executes in an interactive terminal
- **THEN** animated spinner frames update in place until the operation completes
- **AND** a success or failure completion symbol is rendered upon completion

#### Scenario: Non-interactive execution suppresses animation
- **WHEN** an operation with a spinner executes with output redirected or under `NO_COLOR=1`
- **THEN** no animation escape sequences or carriage returns are emitted
- **AND** only the initial message and final status are written to output

### Requirement: Use Nerd Font iconography with diagnostic reporting

The CLI SHALL use Nerd Font glyphs rather than color emojis for UI icons when executing in an interactive TTY without `NO_COLOR`. In non-interactive or uncolored environments, the CLI MUST fall back to plain text labels (e.g. `[ok]`, `[warn]`, `[error]`). The `doctor` command MUST verify and report whether a Nerd Font is installed in the host system.

#### Scenario: Interactive execution displays Nerd Font glyphs
- **WHEN** commands execute in an interactive terminal with color enabled
- **THEN** output icons use Nerd Font glyphs
- **AND** standard emojis are omitted

#### Scenario: Doctor reports Nerd Font status
- **WHEN** a user runs `pkt doctor`
- **THEN** the output includes an entry indicating whether a Nerd Font is detected on the host

### Requirement: Support multi-language internationalization (i18n)

The CLI SHALL localize user-facing messages, command summaries, banners, and diagnostic statuses according to the user's preferred language. The system MUST support English (`en`) and Spanish (`es`), detecting the initial locale from standard environment variables (`LC_ALL`, `LC_MESSAGES`, `LANG`), and MUST provide a persistent global flag (`--lang`) to override the language explicitly.

#### Scenario: Default language resolution from system environment
- **WHEN** a user with Spanish locale (`LANG=es_ES.UTF-8`) runs `pkt welcome` without explicit language flags
- **THEN** the CLI renders the welcome guide and status messages in Spanish

#### Scenario: Explicit language override flag
- **WHEN** a user runs `pkt --lang es doctor` or `pkt --lang en doctor`
- **THEN** the CLI displays diagnostic output localized in the requested language regardless of environment variables

#### Scenario: Fallback to English for unknown locales
- **WHEN** a user runs the CLI with an unsupported or unconfigured locale
- **THEN** the CLI falls back gracefully to English without errors or missing translation keys

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

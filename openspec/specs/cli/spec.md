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

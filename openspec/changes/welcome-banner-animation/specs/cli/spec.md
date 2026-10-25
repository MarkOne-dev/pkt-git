# Spec Delta

## ADDED Requirements

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

# Spec Delta

## MODIFIED Requirements

### Requirement: Produce actionable errors

Command failures SHALL identify the operation and the relevant path or
revision, and MUST NOT print a stack trace or expose credentials. User-facing
errors MUST be written to standard error, formatted with semantic styling in TTY
environments, and localized in the active interface language (supporting `en`
and `es`). Structured errors SHALL classify failures into distinct categories
with actionable guidance.

#### Scenario: Actionable failure

- **WHEN** a command fails because a file does not exist
- **THEN** the error names the command and the missing path
- **AND** the error is written to standard error
- **AND** no stack trace is printed

#### Scenario: Localized error output in Spanish
- **WHEN** a user with Spanish locale or `--lang es` runs a command that fails due to a missing Git repository
- **THEN** the error output is rendered in Spanish indicating Git initialization is required
- **AND** the command exits with non-zero status

#### Scenario: Categorized usage error with status code
- **WHEN** a command invocation violates syntax or misses required arguments
- **THEN** the error is classified as a usage error with actionable syntax guidance
- **AND** the CLI exits with status code 2

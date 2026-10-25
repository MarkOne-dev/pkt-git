# Spec Delta

## ADDED Requirements

### Requirement: Support conventional and auto-detected export destinations

`pkt export` SHALL support omitting the destination directory argument, defaulting by convention to `snapshot/<basename>/` where basename is the stem of the input file. When invoked with zero arguments, `pkt export` SHALL auto-detect an unambiguous topology binary in the repository or report usage options when multiple or none exist.

#### Scenario: Export with default destination convention
- **WHEN** a user runs `pkt export topologia/lab.pkt` without a second argument
- **THEN** it exports the snapshot to `snapshot/lab/`
- **AND** it exits successfully with status 0

#### Scenario: Export with auto-detected repository binary
- **WHEN** a user runs `pkt export` in a repository containing exactly one `.pkt` file
- **THEN** it detects the binary and exports its snapshot to `snapshot/<basename>/`
- **AND** it exits successfully with status 0

#### Scenario: Export with multiple binaries and no arguments
- **WHEN** a user runs `pkt export` in a repository with multiple `.pkt` files and no arguments
- **THEN** it reports the ambiguity and exits with status 2

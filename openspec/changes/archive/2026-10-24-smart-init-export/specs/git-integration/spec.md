# Spec Delta

## ADDED Requirements

### Requirement: Auto-export existing topologies during initialization

`pkt init [dir]` SHALL discover existing `.pkt` and `.pka` files within the target repository. For each discovered topology, it MUST automatically export the initial snapshot into `snapshot/<project>/` and report the export in its output.

#### Scenario: Initialize repository with existing .pkt file
- **WHEN** a user runs `pkt init` in a repository containing `network.pkt`
- **THEN** git hooks and attributes are configured
- **AND** `network.pkt` is automatically exported to `snapshot/network/`
- **AND** the output reports both the hook initialization and the exported snapshot

#### Scenario: Initialize repository without topology files
- **WHEN** a user runs `pkt init` in a repository with no `.pkt` or `.pka` files
- **THEN** git hooks and attributes are configured
- **AND** no snapshot directories are created
- **AND** the command exits successfully

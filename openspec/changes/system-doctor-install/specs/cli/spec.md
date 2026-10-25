# Spec Delta

## ADDED Requirements

### Requirement: Diagnose environment prerequisites with doctor

The CLI SHALL expose a `doctor` command that inspects the host environment and reports the status of Git installation, configured Git user identity (`user.name` and `user.email`), binary availability on `PATH`, terminal color support, and agent skills.

#### Scenario: Run doctor in a healthy environment
- **WHEN** a user runs `pkt doctor` on a fully configured machine
- **THEN** the command outputs status indicators for Git, Git identity, and PATH availability
- **AND** it exits with status 0

#### Scenario: Run doctor with unconfigured Git identity
- **WHEN** a user runs `pkt doctor` and `user.name` or `user.email` is not configured
- **THEN** the command outputs an advisory warning indicating the missing configuration
- **AND** it displays remediation guidance for `git config`
- **AND** it exits with status 0

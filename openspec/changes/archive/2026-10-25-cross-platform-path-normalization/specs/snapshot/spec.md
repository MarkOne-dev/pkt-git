# Spec Delta

## ADDED Requirements

### Requirement: Normalize repository paths to canonical forward slashes

All repository-relative paths returned by topology discovery and default snapshot conventions SHALL be normalized to canonical forward-slash separated paths (`/`), regardless of operating system path separators. CLI commands that report discovered topologies or conventional snapshot export paths MUST format paths with canonical forward slashes on all platforms.

#### Scenario: Default snapshot directory normalization on Windows
- **WHEN** the default snapshot directory is computed for a path containing Windows backslashes or running on Windows
- **THEN** the resulting relative path is `snapshot/<stem>` with forward slashes

#### Scenario: Discovered topology paths use canonical forward slashes
- **WHEN** topologies are discovered across repository subdirectories on Windows
- **THEN** all returned relative paths use forward slashes (e.g. `topologia/network.pkt`)
- **AND** CLI export and initialization output displays forward-slash paths

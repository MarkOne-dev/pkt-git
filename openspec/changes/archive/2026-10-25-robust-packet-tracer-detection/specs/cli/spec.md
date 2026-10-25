# Spec Delta

## MODIFIED Requirements

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

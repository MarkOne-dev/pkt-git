# Proposal

## Why

Setting up `pkt-git` requires external environment prerequisites — primarily a functioning Git installation, configured Git user identity (`user.name` and `user.email`), and the binary on the user's `PATH`. Users currently receive no unified environment diagnostic, and missing dependencies or unconfigured Git credentials only surface later as cryptic hook or commit errors. Providing a `pkt doctor` diagnostic command and enhancing installers with automated dependency remediation eliminates onboarding hurdles across Linux, macOS, and Windows.

## What Changes

- Add a new `pkt doctor` command to the CLI that checks and visually displays:
  - Git installation and version
  - Git user identity (`user.name` and `user.email`)
  - Binary installation and `PATH` availability
  - Terminal color / TTY capability
  - Bundled agent skill installation status
  - Detection of Cisco Packet Tracer executable paths (purely diagnostic/informational)
  - Go SDK availability (informational, highlighting that Go is optional for runtime)
- Update `scripts/install.sh`:
  - Auto-detect missing Git and attempt package installation via system package managers (`apt-get`, `dnf`, `pacman`, or Homebrew on macOS).
  - Verify Git identity (`user.name` / `user.email`) and prompt to configure globally if unset.
  - Offer or automatically append `~/.local/bin` to `~/.bashrc` / `~/.zshrc` when missing from `PATH`.
  - Run `pkt doctor` upon successful install for immediate verification.
- Update `scripts/install.ps1`:
  - Auto-detect missing Git and attempt installation via `winget` (Windows Package Manager).
  - Check Git user identity and prompt to configure if missing.
  - Run `pkt doctor` upon completion.

## Capabilities

### New Capabilities

*(None)*

### Modified Capabilities

- `cli`: Add `doctor` command to inspect environment prerequisites, Git configuration, and tool availability with styled diagnostic reports.

## Impact

- **CLI Layer**: New command `internal/cli/doctor.go`, updated `internal/cli/root.go` and `internal/cli/root_test.go`.
- **Installers**: Updated `scripts/install.sh` and `scripts/install.ps1`.
- **Documentation**: Updated `docs/cli-reference.md`, `docs/installation.md`, and `README.md`.

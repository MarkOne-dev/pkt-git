# Design

## Context

`pkt-git` relies on Git for version tracking and hooks. While the Go binary itself is static and self-contained, user environments often suffer from missing Git installations, unconfigured Git identities (`user.name` and `user.email`), or missing PATH entries. See [proposal.md](proposal.md) for background and motivation.

## Goals / Non-Goals

**Goals:**
- Provide a `pkt doctor` CLI command using `internal/cli/ui` to produce a styled health report covering Git, identity, PATH, skills, and optional tools.
- Enhance `scripts/install.sh` to auto-detect missing Git and invoke system package managers (`apt`, `dnf`, `pacman`, `brew`).
- Auto-configure `$PATH` in `~/.bashrc` / `~/.zshrc` when missing.
- Prompt for missing Git identity interactively in installers when stdin is a terminal.
- Enhance `scripts/install.ps1` with equivalent Windows `winget` and Git identity checks.
- Run `pkt doctor` at the conclusion of both installers.

**Non-Goals:**
- Automating Cisco Packet Tracer installation: Cisco requires NetAcad authentication and license acceptance. Detection remains informational only.
- Forcing interactive prompts in non-interactive CI/CD runs.

## Decisions

### Decision 1: Implement `pkt doctor` within Cobra
Add `doctorCmd` to `internal/cli/doctor.go` registered under the `collab` ("Git Collaboration & Tools:") command group. It inspects:
- Git binary path and version via `exec.Command("git", "--version")`.
- Git identity via `exec.Command("git", "config", "user.name")` and `user.email`.
- Current executable path and whether its directory matches any entry in `os.Getenv("PATH")`.
- Status of agent skills via `internal/skills.Targets`.
- Detection of common Packet Tracer installation paths (`/opt/pt/bin/PacketTracer`, `/Applications/Cisco Packet Tracer.app`, `C:\Program Files\Cisco Packet Tracer`).
- Availability of `go` compiler (flagged as optional for runtime).

### Decision 2: Standardized Diagnostic Exit Codes
`pkt doctor` returns:
- `0`: Environment is healthy or carries only non-critical advisory notices (e.g. uninstalled Packet Tracer or missing Go SDK).
- `1`: Critical prerequisite missing (Git is not installed).

### Decision 3: Installer Auto-Remediation
- In `scripts/install.sh`:
  - If `git` is not found, detect package manager (`apt-get`, `dnf`, `pacman`, `brew`) and install it.
  - If `git config --global user.name` is empty and running interactively, prompt for user name and email.
  - If `BIN_DIR` is not on `PATH`, append export line to `~/.bashrc` or `~/.zshrc`.
  - Invoke `"${TARGET}" doctor` before exiting.
- In `scripts/install.ps1`:
  - If `git` is not found, invoke `winget install --id Git.Git -e --source winget`.
  - If Git identity is empty, prompt and configure via `git config --global`.
  - Invoke `pkt.exe doctor` before exiting.

## Risks / Trade-offs

- **[Risk] Package manager requires sudo privileges** → *Mitigation*: The installer checks if running as root or invokes `sudo` with clear user feedback. If `sudo` is unavailable or denied, it falls back to printing manual installation instructions.
- **[Risk] Non-interactive / unattended installation hangs on prompts** → *Mitigation*: All interactive prompts check `[ -t 0 ]` (terminal attached) and honor `CI=true` / `PKTG_NONINTERACTIVE=1` to proceed without blocking.

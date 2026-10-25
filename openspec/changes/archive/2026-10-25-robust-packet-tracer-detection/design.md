# Design

## Context

`pkt doctor` currently checks for Cisco Packet Tracer using hardcoded binary names (`PacketTracer`, `packettracer`) and a few fixed directory paths (`/opt/pt/bin/PacketTracer`, `/Applications/Cisco Packet Tracer.app`, `C:\Program Files\Cisco Packet Tracer\bin\PacketTracer.exe`). In practice, installations vary widely across operating systems:
- Linux users frequently run Packet Tracer via AppImages (e.g. `/usr/lib/packettracer/packettracer.AppImage`), custom Arch Linux packages, Flatpak, or Snap, with desktop entries in `~/.local/share/applications` or `/usr/share/applications`.
- Windows installers place binaries in versioned directories (e.g. `C:\Program Files\Cisco Packet Tracer 8.2.2\bin\PacketTracer.exe`) and register them in the Windows Registry rather than standard unversioned paths.
- macOS users may have versioned app bundles or user-level `~/Applications` installations.

For motivation, see `proposal.md - Why`.

## Goals / Non-Goals

**Goals:**
- Provide a dedicated `internal/ptdetect` package that implements cross-platform discovery without external runtime dependencies.
- Support XDG `.desktop` entry parsing on Linux to resolve arbitrary executable wrappers and AppImage locations.
- Support versioned directory globbing and registry queries on Windows.
- Support Spotlight (`mdfind`) and standard `.app` bundle inspection on macOS.
- Extract version numbers when available from desktop files, directory names, or bundle metadata.
- Provide clear diagnostics and PATH remediation instructions in `pkt doctor`.

**Non-Goals:**
- Automatically installing Cisco Packet Tracer (it is proprietary software requiring Cisco NetAcad licensing).
- Launching or controlling Cisco Packet Tracer GUI processes from the CLI.

## Decisions

### 1. Dedicated `internal/ptdetect` Package vs. Inline Logic in `doctor.go`
- **Decision**: Encapsulate all discovery algorithms in `internal/ptdetect`.
- **Rationale**: Isolates OS-specific filesystem and registry queries, allowing clean modular unit testing with mock directories and mock desktop entries without bloating `internal/cli`.
- **Alternatives Considered**: Keeping helper functions directly in `doctor.go`. Rejected because `doctor.go` would become overly lengthy and harder to test across platforms.

### 2. Multi-tier Detection Pipeline
- **Decision**: Execute detection in a priority waterfall:
  1. System `$PATH` (`exec.LookPath`)
  2. Platform-specific application registries (XDG desktop files on Linux; Windows Registry `App Paths` / `Uninstall`; macOS Spotlight `mdfind`)
  3. Versioned directory globbing (`/usr/lib/packettracer/*`, `/opt/pt*`, `Program Files\Cisco Packet Tracer *`, `/Applications/Cisco Packet Tracer*.app`)
  4. Containerized wrappers (Flatpak, Snap)
- **Rationale**: Captures standard PATH invocations first, then respects user-configured system desktop/registry entries, and finally falls back to brute filesystem globbing.

### 3. Version Extraction Heuristics
- **Decision**: Infer version from:
  - Directory name patterns (e.g., `Cisco Packet Tracer 8.2.2` -> `v8.2.2`).
  - Desktop entry comments / names (e.g., `Comment=Cisco Packet Tracer 8.2` -> `v8.2`).
  - macOS `Info.plist` (`CFBundleShortVersionString`).
- **Rationale**: Running `PacketTracer --version` is unreliable because Packet Tracer is a GUI-only Qt application that attempts to initialize an X11/Wayland display and hangs in headless environments.

### 4. PATH Verification and Remediation
- **Decision**: In `scripts/install.sh` and `scripts/install.ps1`, verify and persist `${HOME}/.local/bin` (or custom `$PKTG_BIN_DIR`) across all common user shell profiles (`.bashrc`, `.zshrc`, User Environment Registry). In `pkt doctor`, verify that the running binary's directory is present in the current process PATH, printing the exact shell command to add it if missing.
- **Rationale**: Prevents friction when users open a new shell after running the installation script.

## Risks / Trade-offs

- **[Risk]** Headless execution hanging if `PacketTracer` binary is executed directly.
  - **Mitigation**: Never invoke `PacketTracer` with `exec.Command` without a GUI or timeout; extract version via metadata files (desktop entries, directory names, plists).
- **[Risk]** Slow filesystem globbing on network or large disks.
  - **Mitigation**: Limit directory traversal depth to immediate parent containers (e.g. `/opt`, `/usr/lib/packettracer`, `Program Files`) without recursive whole-disk searches.

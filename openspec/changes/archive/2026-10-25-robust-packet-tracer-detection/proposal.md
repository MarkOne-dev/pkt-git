# Proposal

## Why

Cisco Packet Tracer is the primary authoring tool for `.pkt` network topologies, yet `pkt doctor` currently uses an overly simplistic detection mechanism limited to a few hardcoded paths and exact binary names. On modern systems, Packet Tracer is frequently deployed through diverse packaging models—such as AppImages, XDG desktop application entries, Flatpak containers, Snap packages, versioned Windows program directories (e.g. `Cisco Packet Tracer 8.2.2`), and macOS application bundles. Consequently, `pkt doctor` frequently reports false negatives even when Packet Tracer is fully functional on the user's host.

Furthermore, ensuring `pkt` is reliably added to the system PATH across shell environments and properly diagnosed by `pkt doctor` prevents command-not-found issues after installation. A robust, multi-layer discovery engine is needed to accurately locate Packet Tracer, detect its version, and verify terminal PATH readiness.

## What Changes

- Implement a deep, cross-platform Cisco Packet Tracer detection engine:
  - **Linux**: Discover installations via XDG Desktop Entry parsing (`~/.local/share/applications/*.desktop`, `/usr/share/applications/*.desktop`), AppImage executables (`/usr/lib/packettracer/*.AppImage`), directory globs (`/opt/pt*/bin/*`, `/opt/packettracer*/bin/*`), Flatpak identifiers (`com.cisco.PacketTracer`), and Snap binaries (`/snap/bin/packet-tracer`).
  - **Windows**: Discover installations via Registry inspection (`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\PacketTracer.exe` and 32/64-bit Uninstall keys) and versioned globbing (`ProgramFiles\Cisco Packet Tracer *\bin\PacketTracer.exe`).
  - **macOS**: Discover installations via Spotlight (`mdfind "kMDItemCFBundleIdentifier == 'com.cisco.PacketTracer*'"`), `/Applications/Cisco Packet Tracer*.app`, and `~/Applications/`.
  - **Version Extraction**: Parse and report the installed version (e.g., `v8.2.2` or AppImage metadata) alongside the executable path.
- Enhance `pkt doctor` to report detailed Packet Tracer diagnostics, distinguish between full graphical environments and headless servers, and provide actionable PATH remediation.
- Verify and harden PATH configuration across `scripts/install.sh` (Linux/macOS) and `scripts/install.ps1` (Windows) to guarantee persistence across user shell profiles (`~/.bashrc`, `~/.zshrc`, User Registry).

## Capabilities

### Modified Capabilities

- `cli`: Extend the `Diagnose environment prerequisites with doctor` requirement to specify comprehensive, multi-platform Cisco Packet Tracer discovery (supporting AppImages, desktop entries, containers, registry paths, and version globs) and actionable PATH verification.

## Impact

- `internal/cli/doctor.go`: Refactor Packet Tracer detection into a modular discovery subsystem (`internal/cli/ptdetect` or helper functions in `doctor.go`).
- `internal/cli/doctor_test.go`: Add comprehensive unit and mock tests validating desktop entry parsing, path globbing, and registry/platform fallbacks.
- `scripts/install.sh` & `scripts/install.ps1`: Verify PATH injection and ensure seamless onboarding.
- Backward compatibility: Completely non-breaking; improves diagnostic accuracy without altering core CLI arguments or data formats.

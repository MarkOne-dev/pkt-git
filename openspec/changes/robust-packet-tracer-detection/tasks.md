# Tasks

## 1. Packet Tracer Detection Subsystem (`internal/ptdetect`)

- [x] 1.1 Create `internal/ptdetect` package defining the `Installation` data model (`Path`, `Version`, `Method`) and the primary `Find()` entrypoint with system PATH inspection; verify with unit tests in `internal/ptdetect/ptdetect_test.go`.
- [x] 1.2 Implement Linux detection in `internal/ptdetect` supporting XDG `.desktop` file parsing (`Exec=`, `Name=`, `Comment=`), AppImage executables (`/usr/lib/packettracer/*`), directory globs (`/opt/pt*`), Flatpak, and Snap; verify with unit tests in `internal/ptdetect/linux_test.go`.
- [x] 1.3 Implement Windows detection in `internal/ptdetect` with Registry lookup (`App Paths` and `Uninstall` keys) and versioned directory globbing (`Cisco Packet Tracer *\bin\PacketTracer.exe`); verify with unit tests in `internal/ptdetect/windows_test.go`.
- [x] 1.4 Implement macOS detection in `internal/ptdetect` with Spotlight (`mdfind`) and application bundle inspection (`/Applications/Cisco Packet Tracer*.app`); verify with unit tests in `internal/ptdetect/darwin_test.go`.

## 2. Doctor CLI Integration & PATH Verification

- [ ] 2.1 Refactor `findPacketTracer()` in `internal/cli/doctor.go` to use `ptdetect.Find()`, displaying detailed path, version, and discovery method in `pkt doctor`; verify with updated tests in `internal/cli/doctor_test.go`.
- [ ] 2.2 Verify PATH remediation guidance in `internal/cli/doctor.go` and ensure `scripts/install.sh` and `scripts/install.ps1` reliably persist the binary location in user shell profiles; verify shell syntax with `bash -n scripts/install.sh`.

## 3. Verification & Quality Checks

- [ ] 3.1 Run `go test ./...`, `gofmt -l .`, and `make lint` across all packages to verify complete test suite and linters pass cleanly.

## Workflow follow-up

- Merge into `develop` (DEFERRED per user instruction: keep working on current branch).
- Archive the change `robust-packet-tracer-detection`.

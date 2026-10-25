# Tasks

## 1. CLI Doctor Command

- [x] 1.1 Implement `internal/cli/doctor.go` to inspect Git version, identity, PATH, agent skills, and optional tools using `internal/cli/ui`, and register under the `collab` command group.
- [x] 1.2 Add test coverage in `internal/cli/doctor_test.go` and update `internal/cli/root_test.go` to verify `pkt doctor` execution and output formatting.

## 2. Linux and macOS Installer Enhancements

- [x] 2.1 Enhance `scripts/install.sh` to auto-detect and install missing Git via package managers (`apt`, `dnf`, `pacman`, `brew`), prompt for Git identity when unset, configure PATH in shell profiles, and run `pkt doctor`.

## 3. Windows Installer Enhancements

- [x] 3.1 Enhance `scripts/install.ps1` to detect and install Git via `winget`, prompt for Git identity if missing, and run `pkt.exe doctor`.

## 4. Verification and Documentation

- [x] 4.1 Run the full test suite with `go test -count=1 ./...` and verify `make lint` passes.
- [x] 4.2 Update `docs/cli-reference.md`, `docs/installation.md`, and `README.md` documenting `pkt doctor` and installer prerequisite automation.

## Workflow follow-up

- Archive the change with `openspec archive system-doctor-install` after review and verification.

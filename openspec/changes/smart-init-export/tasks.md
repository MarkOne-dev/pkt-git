# Tasks

## 1. Topology Discovery and Path Helpers

- [x] 1.1 Implement `FindTopologies` and `DefaultSnapshotDir` helpers in `internal/snapshot/` and verify with unit tests in `internal/snapshot/export_test.go`.

## 2. Conventional and Zero-Argument Export

- [x] 2.1 Update `internal/cli/export.go` to make the destination directory optional and support zero-argument auto-detection.
- [x] 2.2 Add tests in `internal/cli/commands_test.go` verifying `pkt export` with 0, 1, and 2 arguments.

## 3. Auto-Export on Repository Initialization

- [x] 3.1 Update `internal/cli/init.go` to discover and export detected topologies during repository initialization.
- [x] 3.2 Add test cases in `internal/cli/init_hooks_version_test.go` verifying that `pkt init` auto-exports existing `.pkt` files.

## 4. Verification and Documentation

- [x] 4.1 Run the full test suite with `go test -count=1 ./...` and verify `make lint` passes.
- [x] 4.2 Update `docs/cli-reference.md`, `docs/workflow.md`, and `README.md` reflecting the simplified single-command onboarding flow.

## Workflow follow-up

- Archive the change with `openspec archive smart-init-export` after review and verification.

# Tasks

## 1. Core Path Normalization

- [ ] 1.1 Update `snapshot.DefaultSnapshotDir` and `snapshot.FindTopologies` in `internal/snapshot/snapshot.go` to return canonical forward-slash separated paths, and expand `TestDefaultSnapshotDir` in `internal/snapshot/export_test.go` to verify normalization on Windows backslash input.
- [ ] 1.2 Ensure `internal/cli/init.go` and `internal/cli/export.go` display canonical forward-slash paths in CLI output, and update `internal/cli/init_hooks_version_test.go` to assert canonical path formatting.

## 2. Verification and Quality Checks

- [ ] 2.1 Run `go test ./...`, `gofmt -l .`, and `make lint` to verify all unit tests and quality checks pass.

## Workflow follow-up

- Merge `feature/cross-platform-path-normalization` into `develop`.
- Archive the change `cross-platform-path-normalization`.
- Push `develop` to `origin`.

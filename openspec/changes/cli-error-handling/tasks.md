# Tasks

## 1. Error Taxonomy & Translation Catalogs

- [x] 1.1 Define structured error types in `internal/cli/errors.go` (`RepoError`, `CodecError`, `SnapshotError`, `GitOpError`) supporting `Unwrap()` and categorization.
- [x] 1.2 Add localized error message keys in `internal/i18n/locales/en.json` and `internal/i18n/locales/es.json` with parameter templates (`{{.Path}}`, `{{.Device}}`, `{{.Details}}`).

## 2. CLI Boundary Integration & Formatting

- [x] 2.1 Integrate error mapping in `internal/cli/root.go` (`ExecuteArgs`) to detect structured and sentinel errors, format localized messages, and provide actionable guidance.
- [x] 2.2 Add comprehensive unit tests in `internal/cli/errors_test.go` verifying English and Spanish error translations, exit code mappings (0, 1, 2), and plain-text fallback under `NO_COLOR`.

## 3. Verification & Quality Checks

- [x] 3.1 Run `go test ./...`, `gofmt -l .`, and `go vet ./...` across all packages to verify test suite and code quality pass cleanly.

## Workflow follow-up

- Merge `feature/cli-error-handling` into `develop`.
- Archive change `cli-error-handling`.

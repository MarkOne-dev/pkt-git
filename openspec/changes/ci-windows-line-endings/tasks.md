# Tasks

## 1. Configure Git Line Endings

- [x] 1.1 Update `.gitattributes` with normalization (`* text=auto`) and enforce LF on Go (`*.go text eol=lf`), shell scripts (`*.sh text eol=lf`), Markdown, and YAML files, while preserving Packet Tracer binary and snapshot rules; verify using `git check-attr text eol internal/cli/root.go`.
- [x] 1.2 Ensure checkout configuration in `.github/workflows/ci.yml` cleanly supports cross-platform line endings and verify workflow validity.

## 2. Verification and Formatting

- [ ] 2.1 Run `gofmt -l .`, `go test ./...`, and `make lint` to verify that all Go files comply with LF formatting checks.

## Workflow follow-up

- Merge `feature/ci-windows-line-endings` into `develop`.
- Archive the change `ci-windows-line-endings`.
- Push `develop` to `origin`.

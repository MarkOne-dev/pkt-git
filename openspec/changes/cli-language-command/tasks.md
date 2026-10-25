# Tasks

## 1. Core Persistence & Detection Integration

- [x] 1.1 Add `GitConfigLanguage` constant in `internal/config/config.go` and update `internal/i18n.DetectLanguage()` to inspect `git config pkt.language`; verify with unit tests in `internal/i18n/i18n_test.go`.

## 2. Language Command Implementation

- [x] 2.1 Implement `newLanguageCmd()` in `internal/cli/language.go` with Lipgloss styling, interactive prompt, direct argument handling, `--global`/`--local` flags, and register in `internal/cli/root.go`.
- [x] 2.2 Add unit tests in `internal/cli/language_test.go` verifying direct argument switching, validation errors on invalid codes, non-interactive output, and help alias registration.

## 3. Verification & Quality Checks

- [ ] 3.1 Run `go test ./...`, `gofmt -l .`, and `make lint` across all packages to verify test suite and code quality pass cleanly.

## Workflow follow-up

- Merge `feature/cli-language-command` into `develop`.
- Archive change `cli-language-command`.

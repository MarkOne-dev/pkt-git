# Tasks

## 1. Localization Infrastructure (`internal/i18n`)

- [x] 1.1 Add `github.com/nicksnyder/go-i18n/v2` to `go.mod` and scaffold `internal/i18n` package with `embed.FS` catalog loading and JSON unmarshaler.
- [x] 1.2 Author base English (`locales/en.json`) and Spanish (`locales/es.json`) message catalogs covering banner, welcome guide, doctor diagnostics, and CLI common strings.
- [x] 1.3 Implement locale detection, `Init(lang string)`, and translation helper `T(id string, data ...interface{}) string` in `internal/i18n/i18n.go`; verify with unit tests in `internal/i18n/i18n_test.go`.

## 2. CLI Integration (`pkt --lang`)

- [x] 2.1 Register global persistent flag `--lang` in `internal/cli/root.go` and initialize the active locale in root command pre-run.
- [x] 2.2 Localize `pkt welcome` and `ui.Banner` in `internal/cli/welcome.go` and `internal/cli/ui/banner.go`, supporting both Spanish and English; verify with unit tests in `internal/cli/welcome_test.go`.
- [x] 2.3 Localize `pkt doctor` diagnostic outputs in `internal/cli/doctor.go`; verify with unit tests in `internal/cli/doctor_test.go`.

## 3. Verification & Quality Checks

- [ ] 3.1 Run `go test ./...`, `gofmt -l .`, and `make lint` across all packages to verify test suite and code quality pass cleanly.

## Workflow follow-up

- Merge `feature/cli-i18n` into `develop`.
- Archive change `cli-i18n`.

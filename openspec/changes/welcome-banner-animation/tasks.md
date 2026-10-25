# Tasks

## 1. UI Components (Banner and Animated Spinner)

- [x] 1.1 Implement styled ASCII art banner and subtitle in `internal/cli/ui/banner.go` with Lipgloss styling and plain text fallback; verify with unit tests in `internal/cli/ui/banner_test.go`.
- [x] 1.2 Implement `ui.Spinner` in `internal/cli/ui/spinner.go` with goroutine animation, braille frames, clean completion symbols, and TTY/NO_COLOR fallback; verify with unit tests in `internal/cli/ui/spinner_test.go`.

## 2. Welcome Command and Installer Integration

- [ ] 2.1 Implement `pkt welcome` command in `internal/cli/welcome.go` displaying banner, animated check, and 3-step quickstart guide; register in root command tree and verify with `internal/cli/welcome_test.go`.
- [ ] 2.2 Update `scripts/install.sh` and `scripts/install.ps1` to call `pkt welcome` upon successful installation; verify shell syntax with `bash -n scripts/install.sh` and PowerShell parser.

## 3. Verification and Quality Checks

- [ ] 3.1 Run `go test ./...`, `gofmt -l .`, and `make lint` across all packages to verify test suite and formatting pass cleanly.

## Workflow follow-up

- Merge into `develop` (DEFERRED per user instruction: do not merge yet).
- Archive the change `welcome-banner-animation`.

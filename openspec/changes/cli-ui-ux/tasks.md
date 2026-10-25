# Tasks

## 1. Dependencies and UI Foundation

- [x] 1.1 Add `github.com/charmbracelet/lipgloss` and `github.com/dustin/go-humanize` to `go.mod` and verify with `go mod tidy`.
- [ ] 1.2 Create `internal/cli/ui` package with TTY detection, `NO_COLOR`/`TERM=dumb` handling, semantic status markers (`✔`, `⚠`, `✖`), and verify with unit tests in `internal/cli/ui/ui_test.go`.

## 2. Command Help Organization

- [ ] 2.1 Configure Cobra command groups in `internal/cli/root.go` categorizing commands into Core Workflow, Inspection & Editing, and Git Collaboration & Agents.
- [ ] 2.2 Update `internal/cli/root_test.go` to verify `pkt --help` outputs grouped sections while maintaining exit code 0.

## 3. Tabular Topology Inspection

- [ ] 3.1 Refactor `internal/cli/inspect.go` to use Unicode-bordered tables, device role badges, category summary rollups, and add the `--type` filter flag.
- [ ] 3.2 Add test cases in `internal/cli/commands_test.go` verifying table output formatting and `--type` filtering.

## 4. Modernize Outputs and Error Reporting

- [ ] 4.1 Update `export.go`, `build.go`, `init.go`, `sync.go`, and `verify.go` to use structured status messages and humanized file sizes.
- [ ] 4.2 Format CLI errors with actionable context and visual indicators while strictly preserving exit codes 0, 1, and 2.
- [ ] 4.3 Run the full test suite with `go test -count=1 ./...` and run `make lint` to verify zero regressions.

## 5. Documentation

- [ ] 5.1 Update `docs/cli-reference.md` and `README.md` reflecting the new help groups, `inspect --type` flag, and `NO_COLOR` compliance.

## Workflow follow-up

- Archive the change with `openspec archive cli-ui-ux` after review and verification.

# Proposal: Structured CLI Error Handling and Localization

## Why

While `pkt-git` maintains strict internal error invariants (atomic writes, memory transactions, zero unhandled panics), user-facing error messages are currently emitted as static English strings directly from internal packages (for example, `"not a git repository; run git init first"`, `"resulting XML is invalid"`). Furthermore, error presentation lacks structured classification across distinct failure modes (usage, repository environment, container codec, filesystem, Git operations).

Localizing user-facing errors into English and Spanish (via the existing `internal/i18n` subsystem) and introducing structured error classification will provide clear, actionable guidance in the user's selected language, improve terminal accessibility for international network engineering teams, and streamline error verification in tests.

## What Changes

- **Structured Error Types**: Define clean error taxonomy under `internal/cli` (`UserError`, `RepositoryError`, `CodecError`, `GitOpError`) supporting error unwrapping (`errors.Unwrap` / `errors.Is` / `errors.As`).
- **Localized Error Messages**: Add translation keys in `internal/i18n/locales/en.json` and `internal/i18n/locales/es.json` for common user-facing failure scenarios (missing repository, file not found, invalid Packet Tracer container, XML syntax violation, conflict markers).
- **Localized CLI Error Formatting**: Update top-level error dispatching in `internal/cli/root.go` (`ExecuteArgs`) to detect categorized errors, translate their user messages using the active localizer, and present clean, formatted error notices with actionable hints without leaking stack traces or credentials.
- **Spec Updates**: Update `specs/cli/spec.md` to define localized error reporting requirements.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `cli`: Update Requirement `Produce actionable errors` to require multi-language localization (`en`, `es`) and structured error categories for user-facing CLI failures.

## Impact

- **Affected Code**: `internal/cli/root.go`, `internal/cli/errors.go` (new), command implementations in `internal/cli/`, and `internal/i18n/locales/`.
- **Dependencies**: Uses existing `nicksnyder/go-i18n/v2` and `charmbracelet/lipgloss`; no new external dependencies.
- **Compatibility**: 100% backward-compatible with exit codes (`0` for success, `1` for runtime error, `2` for usage error) and non-interactive `NO_COLOR` plain-text fallbacks.

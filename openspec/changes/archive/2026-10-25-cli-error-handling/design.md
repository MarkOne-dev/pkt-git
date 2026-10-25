# Design: Structured CLI Error Handling and Localization

## Context

Currently, `internal/cli/root.go` catches errors in `ExecuteArgs` and prints them via `ui.Error(errOut, "%s", err)`. Subcommands return errors from internal packages (`internal/hooks`, `internal/codec`, `internal/snapshot`, `internal/editor`), which contain hardcoded English strings.

See `proposal.md` for motivation and `specs/cli/spec.md` for requirements.

## Goals / Non-Goals

**Goals:**
- Provide clear, localized error messages in English and Spanish for all common user-facing failure scenarios.
- Introduce a structured error taxonomy in `internal/cli/errors.go` implementing `error` and `Unwrap()` to facilitate programmatic error classification.
- Preserve consistent exit codes: `0` (success), `1` (runtime failure), `2` (usage error).
- Keep core internal packages (`internal/codec`, `internal/crypto`, `internal/fsx`) decoupled from `internal/i18n`.

**Non-Goals:**
- Do not inject internationalization into low-level cryptographic or codec primitives.
- Do not alter the semantic behavior of Git hooks or internal rollback policies.

## Decisions

### Decision 1: Boundary Translation Pattern

- **Choice**: Translate errors at the CLI layer boundary in `internal/cli/root.go` and `internal/cli/errors.go`.
- **Rationale**: Keeps feature packages pure and testable without requiring an active `i18n.Localizer`. The CLI inspects sentinel errors (e.g., `hooks.ErrNotRepository`, `codec.ErrTruncated`, `codec.ErrInvalidXML`) and structured errors, mapping them to localized messages.
- **Alternative considered**: Injecting `i18n.Localizer` into all internal functions. Rejected because it pollutes clean APIs and couples core algorithms to translation files.

### Decision 2: Structured Error Taxonomy

Define lightweight error wrappers under `internal/cli/errors.go`:
- `usageError`: Missing or invalid flags/arguments (maps to exit code 2).
- `RepoError`: Missing Git repository or uninitialized pkt-git hooks (maps to exit code 1).
- `CodecError`: Packet Tracer binary decryption, tag mismatch, or decompression failure (maps to exit code 1).
- `SnapshotError`: Corrupt XML, invalid edit operation, or missing device configuration (maps to exit code 1).
- `GitOpError`: Git execution failures during commit, merge, or config operations (maps to exit code 1).

### Decision 3: Translation Keys & Contextual Hints

Translation keys in `internal/i18n/locales/` will include actionable remediation hints:
- `ErrNotGitRepository`: "not a git repository; run git init first" / "no es un repositorio git; ejecuta git init primero".
- `ErrPktNotInitialized`: "repository not initialized with pkt-git; run pkt init" / "repositorio no inicializado con pkt-git; ejecuta pkt init".
- `ErrInvalidPacketTracerFile`: "invalid or unsupported Packet Tracer file {{.Path}}" / "archivo de Packet Tracer inválido o no soportado {{.Path}}".
- `ErrDeviceNotFound`: "device \"{{.Device}}\" not found in topology" / "dispositivo \"{{.Device}}\" no encontrado en la topología".
- `ErrXMLInvalid`: "resulting XML document is invalid" / "el documento XML resultante no es válido".

## Risks / Trade-offs

- **[Risk] Parameterized error detail loss** → Mitigation: Use Go template variables (`{{.Path}}`, `{{.Details}}`) in `i18n.T` maps to preserve exact paths, device names, and system details.
- **[Risk] Unhandled/unmapped errors** → Mitigation: Top-level error dispatcher falls back gracefully to `err.Error()` if no specific translation key is mapped, ensuring no failure is hidden or silenced.

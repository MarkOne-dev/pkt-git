# Design

## Context

`pkt-git` currently hardcodes all terminal strings in English across command definitions, status helpers, and interactive components. As the project reaches a global academic audience, particularly throughout Latin America and Spain, providing full Spanish and English support is critical.

For motivation, see `proposal.md - Why`.

## Goals / Non-Goals

**Goals:**
- Implement `internal/i18n` backed by `github.com/nicksnyder/go-i18n/v2`.
- Embed catalog files (`locales/en.json`, `locales/es.json`) at compile time using Go's `embed.FS` with zero runtime file dependencies.
- Use standard Go `encoding/json` for unmarshaling translation catalogs, avoiding unnecessary third-party parser dependencies.
- Resolve user language in priority order:
  1. CLI flag `--lang`
  2. System environment variables (`LC_ALL`, `LC_MESSAGES`, `LANG`)
  3. Default fallback: `en`
- Provide localized strings for `pkt welcome`, `pkt doctor`, `pkt init`, banners, subtitles, and command help.

**Non-Goals:**
- Machine translation at runtime (all strings are pre-translated and verified).
- Translating raw XML Packet Tracer contents, topology device names, or user commit messages.

## Decisions

### 1. `go-i18n/v2` with Embedded JSON Catalogs
- **Decision**: Store message catalogs in `internal/i18n/locales/*.json` and register `json.Unmarshal`.
- **Rationale**: `encoding/json` is built into the Go standard library, keeping binary size minimal and avoiding extra YAML/TOML parser dependencies while supporting CLDR pluralization and templated interpolation.
- **Alternatives Considered**: YAML or TOML. Rejected because standard library JSON requires zero extra external dependencies and embeds cleanly.

### 2. Global Language Flag and Root Persistent Hook
- **Decision**: Define a `--lang` persistent flag on the root Cobra command (`pkt --lang <lang> ...`) and initialize the active `i18n.Localizer` in `PersistentPreRunE`.
- **Rationale**: Ensures every subcommand executes with the user's explicit language preference before printing any messages.

### 3. Locale Resolution Heuristics
- **Decision**: Parse language tags using `golang.org/x/text/language` (already an indirect dependency) to normalize strings like `es_ES.UTF-8`, `es_PE`, or `es` into standard language tags.
- **Rationale**: Robustly handles diverse POSIX and Windows locale environment variable formats.

## Risks / Trade-offs

- **[Risk]** Missing translation key in a secondary language.
  - **Mitigation**: `go-i18n` automatically falls back to the default bundle language (`en`). Unit tests will assert that all keys defined in `en.json` exist in `es.json`.
- **[Risk]** Flag parsing order when help flags (`--help`) are evaluated early.
  - **Mitigation**: Initialize default environment locale before command execution, and re-localize if `--lang` is passed.

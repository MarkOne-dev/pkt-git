# Proposal

## Why

Cisco Packet Tracer is widely taught in networking curricula worldwide (e.g. Cisco Networking Academy across Latin America, Spain, and global institutions). Currently, `pkt-git` prints all command messages, welcome guides, banners, and diagnostic checks strictly in English. Introducing internationalization (i18n) allows users to interact with `pkt-git` natively in Spanish (`es`) and English (`en`), improving accessibility for students, educators, and engineers.

## What Changes

- Introduce a dedicated, embeddable localization subsystem in `internal/i18n` using `github.com/nicksnyder/go-i18n/v2`.
- Embed translation catalogs directly into the Go binary using `//go:embed locales/*.json` so no external language files are required at runtime.
- Provide full Spanish (`es`) and English (`en`) translation catalogs covering:
  - Welcome banner, quickstart guide, and runtime status messages.
  - Command descriptions, usage examples, and flag help in Cobra.
  - Environment diagnostic checks in `pkt doctor`.
  - Operational feedback for `pkt init`, `pkt export`, `pkt build`, `pkt inspect`, `pkt edit`, and `pkt sync`.
- Support automatic locale detection from environment variables (`LC_ALL`, `LC_MESSAGES`, `LANG`) on Linux/macOS and OS locale on Windows, with fallback to English (`en`).
- Provide an explicit `--lang` global flag (e.g. `pkt --lang es welcome` or `pkt --lang en doctor`) to override detection at runtime.

## Capabilities

### Modified Capabilities

- `cli`: Add requirement `Support multi-language internationalization (i18n)` to govern message localization, locale discovery, fallback behaviors, and language override flags.

## Impact

- Dependencies: Add `github.com/nicksnyder/go-i18n/v2` to `go.mod`.
- Architecture: Create `internal/i18n` providing localized formatters (`i18n.T("Key", data...)`).
- CLI: Wire language resolution into root command persistent pre-run hook.
- Non-breaking: Defaults to system locale or English; all existing arguments, flags, and machine-readable outputs remain compatible.

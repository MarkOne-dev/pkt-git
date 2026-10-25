# Spec Delta

## ADDED Requirements

### Requirement: Support multi-language internationalization (i18n)

The CLI SHALL localize user-facing messages, command summaries, banners, and diagnostic statuses according to the user's preferred language. The system MUST support English (`en`) and Spanish (`es`), detecting the initial locale from standard environment variables (`LC_ALL`, `LC_MESSAGES`, `LANG`), and MUST provide a persistent global flag (`--lang`) to override the language explicitly.

#### Scenario: Default language resolution from system environment
- **WHEN** a user with Spanish locale (`LANG=es_ES.UTF-8`) runs `pkt welcome` without explicit language flags
- **THEN** the CLI renders the welcome guide and status messages in Spanish

#### Scenario: Explicit language override flag
- **WHEN** a user runs `pkt --lang es doctor` or `pkt --lang en doctor`
- **THEN** the CLI displays diagnostic output localized in the requested language regardless of environment variables

#### Scenario: Fallback to English for unknown locales
- **WHEN** a user runs the CLI with an unsupported or unconfigured locale
- **THEN** the CLI falls back gracefully to English without errors or missing translation keys

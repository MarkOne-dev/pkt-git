# Design

## Context

While `pkt-git` supports i18n localization via `internal/i18n`, the active locale is currently ephemeral unless environment variables like `LANG` are modified system-wide. A persistent CLI command allows developers to toggle between languages globally or within specific repositories.

For motivation, see `proposal.md - Why`.

## Goals / Non-Goals

**Goals:**
- Provide a first-class Cobra command `pkt language` with alias `pkt lang`.
- Deliver a clean, stylized Lipgloss UI for terminal users with active indicator markers and numbered selections.
- Support direct non-interactive argument usage (`pkt lang es`, `pkt lang en`).
- Persist choices natively using Git configuration (`git config pkt.language`), supporting both `--global` and `--local` scopes.
- Update `internal/i18n.DetectLanguage()` to automatically read `git config pkt.language`.

**Non-Goals:**
- Supporting arbitrary third-party language packs at runtime (catalogs are statically verified).
- Translating git commit logs or Packet Tracer device names.

## Decisions

### 1. Git Config as the Persistence Store
- **Decision**: Persist language preferences using Git's configuration system (`pkt.language`), defaulted to `--global`.
- **Rationale**: Git is already a mandatory prerequisite for `pkt-git`. Leveraging Git configuration provides cross-platform persistence across Linux, macOS, and Windows without inventing new dotfiles or config directories, and naturally supports per-repository overrides.
- **Alternatives Considered**: Custom JSON in `~/.config/pkt-git/config.json`. Rejected because it duplicates configuration mechanisms when Git already handles user and local scopes reliably.

### 2. Interactive vs. Stream Detection
- **Decision**: Use `isatty.IsTerminal(os.Stdin.Fd())` combined with `ui.ColorEnabled(out)` to determine whether to prompt the user.
- **Rationale**: When invoked in automated scripts, CI, or pipes, the command must never block or wait for stdin input; it must print informational status and exit immediately.

### 3. Integrated Resolution Hierarchy
- **Decision**: In `internal/i18n.DetectLanguage()`, check `git config pkt.language` after CLI flags but before system environment variables.
- **Rationale**: Explicit user preference configured via `pkt language` should take precedence over the generic OS-wide desktop locale, while a temporary `--lang` flag retains highest priority for one-off commands.

## Risks / Trade-offs

- **[Risk]** Running `pkt language --local` outside a Git repository.
  - **Mitigation**: Detect non-repository status using `gitx.FindRepositoryRoot()` or check `git config --local` exit code, displaying a clear usage error.

# Design: CI Windows Line Endings Normalization

## Context

See `proposal.md` for problem background. The repository's current `.gitattributes` only covers `.pkt`/`.pka` binary artifacts and `snapshot/**` files. Without explicit `eol` rules for Go source files, Git on Windows runners converts text files to CRLF (`\r\n`) during checkout (`actions/checkout@v4`). Because `gofmt` requires LF, `gofmt -l .` reports every `.go` file and exits with error code 1.

## Goals / Non-Goals

**Goals:**
- Ensure all `.go` files are checked out with LF line endings across all platforms (Windows, Linux, macOS) in CI and local clones.
- Ensure `gofmt -l .` succeeds cleanly in `.github/workflows/ci.yml` on `windows-latest`.
- Ensure shell scripts (`.sh`) retain LF line endings and PowerShell scripts (`.ps1`) retain CRLF line endings.
- Maintain existing pkt-git ownership rules for binaries (`*.pkt`, `*.pka`) and snapshots.

**Non-Goals:**
- Modify `internal/hooks/init.go` or alter the `pkt init` attribute merging logic for end-user network repositories.
- Alter Go code or remove `gofmt` checking from CI.

## Decisions

### 1. Enforce `eol=lf` via `.gitattributes`

- **Choice**: Add `* text=auto` and explicit `*.go text eol=lf` to `.gitattributes`.
- **Rationale**: `.gitattributes` is checked into the repository and is respected by Git during checkout on all platforms and runners, overriding any host-level `core.autocrlf` setting. This guarantees that both CI runners and local Windows contributors have identical LF line endings in Go source code.
- **Alternatives considered**:
  - *Adding `git config --global core.autocrlf false` in CI workflow*: Fixes CI runners but does not solve line ending issues for local Windows developers.
  - *Normalizing in CI step via dos2unix or tr*: Adds external dependency and masks checkout inconsistency rather than fixing it at the source.

### 2. Comprehensive rule set for scripts and configuration files

- **Choice**:
  - `* text=auto` (auto-detect text files)
  - `*.go text eol=lf`
  - `*.sh text eol=lf`
  - `*.ps1 text eol=crlf`
  - `*.md text eol=lf`
  - `*.yml text eol=lf`
  - `*.yaml text eol=lf`
  - Existing `*.pkt binary`, `*.pka binary`, and `snapshot/**` text rules.
- **Rationale**: Prevents Unix shell scripts from breaking with `\r: command not found` on Windows WSL/runners, while preserving CRLF for PowerShell scripts and LF for structured configuration files.

## Risks / Trade-offs

- **[Working tree line endings]**: Existing clones on Windows might retain CRLF until files are re-checked out.
  - *Mitigation*: Fresh clones and CI checkouts immediately apply `.gitattributes`. For existing clones, `git checkout -- .` or `git add --renormalize .` will normalize files.

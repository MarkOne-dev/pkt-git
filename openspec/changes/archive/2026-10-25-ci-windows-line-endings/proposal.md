# Proposal: CI Windows Line Endings Normalization

## Why

On Windows CI runners (`windows-latest`), Git default checkout behavior converts Unix LF line endings to Windows CRLF (`\r\n`). The CI workflow step `Check formatting` runs `gofmt -l .`, which strictly expects LF line endings and flags every Go source file in the repository when CRLF is present, causing the Windows CI job to fail. Normalizing line endings across platforms ensures consistent builds and clean formatting checks on Windows, Linux, and macOS.

## What Changes

- Update the repository `.gitattributes` to explicitly configure Git line-ending normalization:
  - Configure `* text=auto` for automatic line-ending detection.
  - Enforce LF line endings for Go files (`*.go text eol=lf`), shell scripts (`*.sh text eol=lf`), Markdown (`*.md text eol=lf`), and YAML files (`*.yml text eol=lf`, `*.yaml text eol=lf`).
  - Enforce CRLF for PowerShell scripts (`*.ps1 text eol=crlf`).
  - Preserve all existing pkt-git rules (`*.pkt binary`, `*.pka binary`, and snapshot text rules).
- Verify that `gofmt -l .` and CI formatting checks pass without flagging line-ending differences on Windows.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None (pure tooling and repository configuration; `skip_specs: true` declared).

## Impact

- **Affected files**: `.gitattributes`, `.github/workflows/ci.yml`.
- **APIs and Runtime**: No impact on `pkt` runtime behavior, CLI syntax, or exported binaries/snapshots.
- **Developer Experience & CI**: Developers and CI runners on Windows will check out `.go` files with LF line endings, preventing false formatting failures in `gofmt`.

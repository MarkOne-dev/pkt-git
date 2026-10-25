# Design: Cross-Platform Canonical Path Normalization

## Context

See `proposal.md` for background. Within `pkt-git`, the `internal/hooks` package already establishes the architectural invariant of converting Git paths with `filepath.ToSlash()` (e.g. in `postmerge.go`, `precommit.go`, and `resolve.go`), because Git internally treats repository paths as forward-slash-separated (`/`).

However, the topology discovery helper `snapshot.FindTopologies` and conventional path helper `snapshot.DefaultSnapshotDir` were using `filepath.Join` and `filepath.Rel`, which produce Windows backslashes (`\`) on Windows runners. This caused CLI commands (`pkt init`, `pkt export`) to emit Windows-specific backslashes and caused assertions in `init_hooks_version_test.go` to fail on Windows CI.

## Goals / Non-Goals

**Goals:**
- Enforce the architectural rule that all repository-relative paths are canonically slash-separated (`/`) across Windows, Linux, and macOS.
- Ensure `snapshot.DefaultSnapshotDir` and `snapshot.FindTopologies` always return canonical slash-separated paths.
- Ensure CLI outputs (`pkt init`, `pkt export`) display uniform slash-separated relative paths on all operating systems.
- Verify through automated tests that path normalization functions identically on both Unix and Windows path formats.

**Non-Goals:**
- Modify absolute system paths (such as the paths displayed in `pkt doctor` or `pktgit.bin` configuration).

## Decisions

### 1. Canonical normalization in `snapshot.DefaultSnapshotDir`

- **Choice**: Convert incoming `pktPath` with `filepath.ToSlash`, extract the base file stem, and construct the directory using `path.Join("snapshot", stem)`.
- **Rationale**: The `path` standard library package is purpose-built for slash-separated logical paths. Using `path.Join` guarantees `snapshot/<stem>` across all operating systems while correctly handling both `/` and `\` in input strings.
- **Alternatives considered**:
  - *Keep `filepath.Join` and patch test assertions with `filepath.Join`*: Rejected because it leaves CLI output inconsistent between operating systems and contradicts Git's canonical path representation.

### 2. Canonical normalization in `snapshot.FindTopologies`

- **Choice**: Wrap discovered relative paths with `filepath.ToSlash(filepath.Clean(rel))` before appending to the results slice.
- **Rationale**: Normalizing at the point of discovery ensures all downstream callers (CLI commands, menus, logs, hooks) receive uniform paths without requiring defensive normalization downstream.

### 3. Verification across platforms

- **Choice**: Expand unit tests in `internal/snapshot/export_test.go` and `internal/cli/init_hooks_version_test.go` to explicitly verify slash normalization when given Windows-style path inputs (e.g. `topologia\network.pkt`).

## Risks / Trade-offs

- **[Windows filesystem compatibility]**:
  - *Risk*: Passing forward-slash paths to Windows filesystem calls.
  - *Mitigation*: The Windows NT kernel and Go's runtime natively support forward slashes in file system paths (`os.Open`, `os.Stat`, `os.Create`). Furthermore, when combined with absolute roots via `filepath.Join(root, rel)`, Go's `filepath.Join` handles any OS-specific adaptation automatically.

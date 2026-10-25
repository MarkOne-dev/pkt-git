# Proposal: Cross-Platform Canonical Path Normalization

## Why

In Git repositories, file and directory paths are canonically forward-slash-separated (`/`), regardless of the underlying operating system. On Windows, relying on OS-native `filepath.Join` produces backslashes (`\`) for relative repository paths in `snapshot.DefaultSnapshotDir` and `snapshot.FindTopologies` (e.g. `snapshot\network`). This leads to inconsistent CLI output across operating systems, breaks cross-platform contract uniformity with Git, and causes CI test failures on Windows runners. A scalable architectural solution enforces canonical forward-slash normalization for all repository-relative paths across all platforms.

## What Changes

- Normalize all repository-relative paths produced by `snapshot.DefaultSnapshotDir` and `snapshot.FindTopologies` to canonical forward-slash (`/`) format.
- Ensure CLI commands (`pkt init`, `pkt export`) display canonical forward-slash relative paths uniformly on Windows, Linux, and macOS.
- Update `internal/cli` and `internal/snapshot` tests to verify that canonical relative paths are strictly enforced across operating systems.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `snapshot`: Require that relative snapshot directory paths and discovered topology paths always use canonical forward slashes (`/`), ensuring seamless interoperability with Git and uniform cross-platform behavior.

## Impact

- **Affected Packages**: `internal/snapshot`, `internal/cli`.
- **APIs and Invariants**: `snapshot.DefaultSnapshotDir` and `snapshot.FindTopologies` guarantee forward-slash-separated relative paths on all operating systems. Filesystem disk I/O remains fully functional on Windows as Go and Windows OS APIs natively accept slash-separated paths.
- **CI & Testing**: Eliminates path separator mismatches in tests on Windows CI runners.

# Proposal

## Why

The current Python implementation requires a per-machine virtualenv and a
C-compiled Twofish binding, has no automated test suite or CI, and carries
correctness and security defects in XML editing, snapshot synchronization,
Git hook handling, and secret exposure. Rewriting the tool in Go with Cobra
removes the Python and native-toolchain burden, produces a single
cross-platform binary, and gives the same Git-native workflow a typed,
testable foundation.

## What Changes

- Replace the Python implementation (`tools/*.py`, `setup.sh`, `setup.ps1`,
  and the Python-dependent `.githooks/`) with a Go 1.27 CLI built on Cobra.
  **BREAKING**: the `ptg.py`, `edit_config.py`, and `spike_roundtrip.py`
  entry points and per-repository virtualenv are removed.
- Implement the `.pkt`/`.pka` codec in pure Go, using
  `golang.org/x/crypto/twofish` for the block cipher and an in-house
  CMAC/EAX implementation. No cgo or C toolchain is required.
- Preserve the user-facing command surface: `init`, `export`, `build`,
  `inspect`, `edit`, `sync`, `resolve`, `verify`, `skills`, `version`, plus the
  hidden `pre-commit`, `post-merge`, and `post-checkout` hook entry points.
- Replace string-index XML surgery with byte-offset, device-scoped edits that
  preserve the original document formatting and reject ambiguous matches.
  **BREAKING**: `chgline` matches a complete config line, not an arbitrary
  substring.
- Validate the XML document before `build` encodes it and before any edit is
  committed to disk. **BREAKING**: invalid XML now fails instead of producing
  a corrupt binary.
- Make `export` and `sync` authoritative for derived files: regenerate
  `configs/*.cfg` and `metadata/topology.json`, and remove stale generated
  files instead of leaving them behind.
- Treat a failed EAX authentication tag or declared-length mismatch as a
  decode error by default. **BREAKING**: corrupted or unsupported binaries no
  longer decode with a warning.
- Harden Git integration: merge missing `.gitattributes` rules instead of
  skipping the file, handle squash/rebase/cherry-pick ranges, avoid binary
  drift after merges, and resolve text and binary conflicts deterministically.
- Redact credentials and other sensitive CLI lines from `show` and any
  human-readable output using a shared redaction policy.
- Add table-driven unit tests, codec golden tests against the bundled
  specimen, hook integration tests in temporary repositories, and CI for
  Linux, macOS, and Windows.
- Rewrite the installers and documentation for a compiled binary. Keep all
  implementation packages under `internal/`; no public `pkg/` API is exposed
  in this change.

## Capabilities

### New Capabilities

- `codec`: decode, encode, and verify Packet Tracer binaries, including the
  four-layer container, Twofish-EAX authentication, compression, and
  corruption handling.
- `snapshot`: export binaries into the textual snapshot layout, parse topology
  metadata, rebuild binaries, synchronize derived files, and inspect projects.
- `snapshot-editing`: device-scoped, byte-preserving XML editing and
  read-only config display with secret redaction.
- `git-integration`: repository initialization, thin hooks, pre-commit and
  post-merge/checkout automation, conflict resolution, and Git attributes.
- `cli`: the Cobra command tree, argument validation, exit codes, help,
  version reporting, and bundled agent-skill management.

### Modified Capabilities

None. This is the first OpenSpec specification for the project; no existing
capability specs are present.

## Impact

- **Code**: the entire `tools/` tree, `setup.sh`, `setup.ps1`, and
  `.githooks/` are replaced by `cmd/pkt/` and `internal/` packages.
- **Dependencies**: Go 1.27, Cobra, and `golang.org/x/crypto/twofish` are
  added; Python, `twofish`, `pycryptodome`, and virtualenv are removed.
- **Behavior**: stricter codec validation, exact-line editing, authoritative
  snapshot synchronization, merged Git attributes, safer hook ranges, and
  secret redaction. The `.pkt` container format itself is unchanged.
- **Compatibility**: the codec remains verified against Packet Tracer
  9.0.1.0858 and the bundled 107-device specimen. Older PT versions and
  `.pka` activity files remain explicitly untested.
- **Distribution**: installers move from Python/venv setup to installing a
  compiled binary; Linux, macOS, and Windows remain supported targets.
- **Public API**: no importable `pkg/` API is introduced; external consumers
  continue to use the `pkt` command.

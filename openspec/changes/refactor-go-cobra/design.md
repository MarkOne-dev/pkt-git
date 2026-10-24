# Design

## Context

The repository currently contains a Python implementation, a per-machine
virtualenv, a C-compiled Twofish binding, and Python-dependent Git hooks.
The refactor replaces all of it with a single Go binary while preserving the
Git-native workflow and the `.pkt` container format. See `proposal.md` for
motivation and `specs/` for the behavioral contract.

Constraints:

- No Python, virtualenv, cgo, or C toolchain in the shipped tool.
- Keep the command surface and the snapshot layout (`snapshot/<project>/`
  with `base.snapshot.xml`, `configs/`, `metadata/`).
- Keep all implementation packages under `internal/`; no public `pkg/` API.
- Preserve the verified Packet Tracer 9.0.1.0858 behavior and the bundled
  107-device regression specimen.
- Keep OpenSpec artifacts in English.

## Goals / Non-Goals

**Goals:**

- One statically linked, cross-platform `pkt` binary built with Go and Cobra.
- Pure-Go codec with authenticated decoding and deterministic encoding.
- Byte-preserving, device-scoped XML editing with validation before write.
- Authoritative snapshot export/sync, including stale-file pruning.
- Safer Git hooks that handle squash merges, branch checkouts, and conflict
  resolution without unrelated staging.
- Test coverage for the codec, topology parser, editor, hooks, and CLI, with
  CI on Linux, macOS, and Windows.

**Non-Goals:**

- Supporting Packet Tracer versions or `.pka` activity documents that are not
  covered by a specimen.
- Changing the snapshot directory layout or the `.pkt` container format.
- Exposing a stable importable Go library API.
- Automatically choosing a side in a semantic merge conflict.
- Building a GUI or replacing Packet Tracer itself.

## Decisions

### 1. Go module and package layout

Module path: `github.com/MarkOne-dev/pkt-git`. Go 1.27. All code lives under
`internal/`; only `cmd/pkt` is a binary entry point.

```
pkt-git/
├── cmd/pkt/main.go              # process entry point, exit-code mapping
├── internal/
│   ├── cli/                     # Cobra commands, flags, help, exit codes
│   ├── config/                  # build metadata, repository config keys
│   ├── codec/                   # .pkt container: decode, encode, verify
│   ├── crypto/
│   │   ├── cmac/                # CMAC/OMAC1 over a block cipher
│   │   └── eax/                 # EAX mode and tag verification
│   ├── topology/                # XML -> devices, links, config lines
│   ├── snapshot/                # export, build, sync, inspect orchestration
│   ├── xmledit/                 # byte-offset XML index and span edits
│   ├── editor/                  # show/setiface/chgline/insertafter policies
│   ├── security/                # credential redaction policy
│   ├── gitx/                    # git command wrapper and path handling
│   ├── hooks/                   # pre-commit, post-merge, post-checkout logic
│   ├── skills/                  # embedded agent skill asset and installer
│   └── fsx/                     # atomic file writes and safe file helpers
├── testdata/specimens/          # real .pkt regression inputs
├── scripts/                     # install.sh, install.ps1
├── docs/                        # updated user and architecture docs
├── openspec/                    # specs and change history
├── Makefile
├── go.mod
└── go.sum
```

Rationale: the boundaries match the existing four layers (crypto, topology,
editing, automation) but make dependencies one-directional: `cli` depends on
feature packages; feature packages depend on `codec`, `topology`, `gitx`, and
`fsx`; `codec` and `topology` do not depend on `cli`. This keeps the codec
testable in isolation and leaves room to promote packages later without
restructuring.

Alternative considered: a single `internal/app` package. Rejected because it
would mix crypto, XML editing, and Git process handling, making tests and
future extraction harder.

### 2. Codec: x/crypto Twofish plus in-house CMAC/EAX

Use `golang.org/x/crypto/twofish` for the block cipher and implement CMAC
(NIST SP 800-38B) and EAX in `internal/crypto`. The four container layers stay
identical to the verified Python implementation:

1. qCompress framing: 4-byte big-endian uncompressed size + zlib stream.
2. Inner positional XOR: `b[i] ^= (L - i) mod 256`.
3. Twofish-EAX with the fixed key `0x89` x16, nonce `0x10` x16, empty header,
   and a 16-byte tag appended to the ciphertext.
4. Outer positional XOR with reversal:
   `out[L-1-i] = in[i] ^ ((L - i*L) mod 256)`.

Decode reverses 4 -> 3 -> 2 -> 1. Authentication and length checks are hard
failures. `verify` reports byte identity separately because Packet Tracer's
zlib emitter differs from Go's level-9 output.

Alternatives considered:

- **Implement Twofish in-house.** Rejected: high risk for no benefit;
  `x/crypto/twofish` is a maintained, pure-Go implementation with the same
  block-cipher contract.
- **Keep the C binding.** Rejected: violates the no-C requirement and blocks
  Windows installs without MSVC.
- **Use a third-party EAX package.** Rejected: the PT variant is
  Crypto++-style (`ciphertext || tag`) and the tag/OMAC details are already
  documented; a small, tested implementation keeps the behavior explicit.

### 3. XML editing: parse offsets, edit spans, validate, write atomically

`internal/xmledit` builds an index over the raw snapshot bytes using
`encoding/xml.Decoder` with `InputOffset()`. It records:

- the byte span of each device's `NAME` and `RUNNINGCONFIG`;
- the span of each `LINE` start tag, text content, and end tag;
- the enclosing device for each config line.

Edits operate only on those spans. `setiface` replaces the content between the
matching `interface <IFACE>` line and the next `!` line; `chgline` replaces
one exact line's text content; `insertafter` inserts after the first matching
line. Generated lines use the indentation and line ending detected in the
surrounding block. The original bytes outside the edited span are never
re-serialized.

After the edit, the whole document is parsed again. If parsing fails, the
original bytes are returned untouched. Writes go through
`fsx.WriteFileAtomic`: create a temporary file in the target directory, sync,
close, and rename over the target.

Alternatives considered:

- **`encoding/xml` marshal/unmarshal.** Rejected: it normalizes attribute
  quoting, namespaces, whitespace, and self-closing tags, producing huge
  spurious diffs and risking semantic changes.
- **String search and replace.** Rejected: the previous implementation's
  `chgline` matched substrings and could edit the wrong occurrence. Exact
  line spans plus ambiguity errors are safer.

### 4. Topology and snapshot orchestration

`internal/topology` parses the decoded XML into ordered `Device`, `Link`, and
`Endpoint` values using `encoding/xml` structs. It does not write files.
`internal/snapshot` owns the filesystem contract:

- `Export` decodes and parses first, computes all output bytes in memory, and
  only then writes `base.snapshot.xml`, `configs/*.cfg`, and
  `metadata/topology.json`. Decode/parse failures therefore leave the old
  snapshot intact.
- `Sync` regenerates derived files from the XML, creates missing directories,
  and removes `configs/*.cfg` files that are not in the expected set.
- `Build` validates the XML, encodes it, and writes the output atomically.
- `Inspect` decodes and prints a summary without writing.

Config file names use the existing sanitization rule (`[^A-Za-z0-9_.-]+` ->
`_`, trim `._`, fallback `device`). Two devices that sanitize to the same name
are an error, not a silent overwrite.

`metadata/topology.json` is emitted from typed structs, not maps, to keep field
order and formatting deterministic. Unresolved cable endpoints keep the
existing deterministic fallback (last eight characters of the reference id).

Alternative considered: exporting directly to the target directory while
parsing. Rejected because a parse failure midway would leave a half-updated
snapshot.

### 5. Git integration: repository-local hooks over a `git` wrapper

`internal/gitx` wraps `git` with `exec.Command`, preserving stdout/stderr and
NUL-separated path output (`-z`) so paths with spaces are handled correctly.
`internal/hooks` implements the three hook entry points.

`pkt init` writes thin POSIX shell hooks:

```sh
#!/bin/sh
# Generated by pkt init. Do not edit.
exec "$(git config --get pktgit.bin || command -v pkt)" pre-commit "$@"
```

The post-merge and post-checkout variants substitute their command name. The
hook records the absolute path of the running executable in
`pktgit.bin`, sets `core.hooksPath .githooks`, and merges required rules into
`.gitattributes`:

```
*.pkt binary
*.pka binary
snapshot/**/base.snapshot.xml text
snapshot/**/*.json text
snapshot/**/*.cfg text
```

Existing unrelated rules are preserved; required rules are appended only when
missing. `pkt init` is idempotent and refuses to run outside a Git repository.

Hook behavior:

- **pre-commit:** list staged `.pkt`/`.pka` files with
  `git diff --cached --name-only --diff-filter=ACMR -z`. Export each into
  `snapshot/<stem>/`, then stage that directory with `git add -A -- <dir>`.
  A failure aborts the commit. If no binary is staged but a snapshot XML is
  staged, print a stale-binary warning with the rebuild command.
- **post-merge:** Git passes one argument, the squash flag (`0` or `1`).
  Treat it as a flag, never as a revision. Use `ORIG_HEAD..HEAD` for the
  changed-path range; if `ORIG_HEAD` is unavailable, fall back to
  `HEAD^..HEAD`.
- **post-checkout:** Git passes old revision, new revision, and a branch
  flag. Only run when the flag is `1`. Use `git diff --name-only -z <old>
  <new>`; when the old revision is all zeros, use `git diff-tree --root -r`
  for the new commit.
- Both hooks rebuild only changed `snapshot/*/base.snapshot.xml` projects,
  writing `topologia/<project>.pkt` atomically. A failed rebuild is reported
  and does not prevent the checkout/merge; the hook exits non-zero to signal
  the failure.

`pkt resolve`:

- Report mode reads `git diff --name-only --diff-filter=U -z`, classifies
  `.pkt`/`.pka` as binary and the rest as text, counts conflict blocks, and
  prints the first marker line. It writes nothing.
- With `--ours` or `--theirs`, apply `git checkout --ours|--theirs -- <text
  paths>`, verify no `<<<<<<<`, `=======`, or `>>>>>>>` markers remain,
  refresh each affected snapshot, rebuild each conflicted binary from its
  resolved snapshot, and stage only the resolved text paths, snapshot
  directories, and rebuilt binaries.
- Binary-only conflicts rebuild from the existing snapshot. A missing
  snapshot is an error.
- No policy means no writes and a non-zero exit; the user must choose.

Alternatives considered:

- **Keep the Python hooks.** Rejected: they invoke Python and inherit the
  interpreter/venv problem.
- **Use a Go Git library.** Rejected: shelling out to Git preserves the
  user's configuration, credential helpers, and behavior with less surface
  area.
- **`git add -A` after resolve.** Rejected: it stages unrelated work. The
  resolver stages only paths it touched.

### 6. CLI: Cobra with explicit exit-code mapping

`internal/cli` defines one file per command and a root command with
`SilenceUsage` and `SilenceErrors`. Commands use `cobra.ExactArgs` or a
custom validator and return a typed `usageError` for invalid invocations.
`Execute` maps:

- `nil` -> 0
- `usageError` -> 2
- any other error -> 1

The root command prints help when invoked with no arguments and exits 2.
`version` and the root `--version` flag print `config.Version` plus commit and
date when available. Build metadata is injected with `-ldflags`.

Hook commands are registered with `Hidden: true` so generated hooks can call
them but `pkt --help` stays clean.

`pkt skills` embeds the skill with `//go:embed` and writes it to the OpenCode,
Claude Code, and Gemini CLI global skill directories. Removal targets only the
`pkt-git` skill directory. This replaces the old `tools/skills/pkt-git`
copy that no longer exists.

### 7. Security: redaction and non-leaking errors

`internal/security` owns a redaction policy for displayed config lines. It
matches case-insensitive credential keywords (for example `password`,
`secret`, `key`, `community`, `pre-shared-key`, `wpa-psk`) and replaces the
value portion with `<redacted>`. The policy is applied only to human-readable
output (`show` and summaries), never to the snapshot itself.

Errors include the command, path, and relevant revision but never line
contents. The codec never logs key or nonce material.

Alternative considered: hiding any line containing `PASSWORD` only (the old
behavior). Rejected: it misses `secret`, SNMP communities, and key material.

### 8. Testing and CI

- Unit tests: CMAC/EAX known-answer vectors, each container layer, topology
  parsing, config sanitization/collisions, JSON output, XML span edits,
  redaction, CLI argument validation, and Git path parsing.
- Golden tests: decode the bundled specimen, assert XML well-formedness and
  round-trip equality, and compare config/metadata hashes against values
  captured from the Python implementation (`v1-python` tag).
- Integration tests: create temporary Git repositories, run `pkt init`,
  commit a binary, simulate merges/checkouts/squash merges, and verify
  snapshots, staged paths, and rebuilt binaries.
- CI: GitHub Actions matrix for Ubuntu, macOS, and Windows running
  `gofmt -l`, `go vet ./...`, `go test ./...`, and `go build ./...`.

## Risks / Trade-offs

- **[Twofish dependency changes behavior]** -> Pin `golang.org/x/crypto` and
  run known-answer vectors plus the real specimen round trip in CI.
- **[Byte-offset editing mis-indexes unusual XML]** -> Use the standard
  decoder for offsets, reject ambiguous matches, re-parse after every edit,
  and cover indentation/CRLF/self-closing cases in tests.
- **[Atomic rename differs on Windows]** -> Use a same-directory temp file
  plus `os.Rename` (which replaces on Windows), and test on the Windows CI
  runner.
- **[Hook ranges are wrong for squash/rebase/cherry-pick]** -> Treat the
  post-merge argument as a flag, test the documented ranges in temporary
  repositories, and fall back to `HEAD^..HEAD` only when `ORIG_HEAD` is
  unavailable.
- **[Removing Python breaks existing installations]** -> Keep the
  `v1-python` tag, document the one-time `pkt init` migration, and rewrite
  the installers to install the Go binary.
- **[Large snapshots consume memory]** -> The verified specimen is about
  5.9 MB; in-memory export and editing are acceptable. If larger specimens
  appear, the index can stream while preserving the same API.
- **[Redaction can miss a vendor-specific secret form]** -> Keep the policy
  centralized, test common IOS forms, and prefer over-redaction in display
  output.

## Migration Plan

1. Land the Go scaffold and codec/topology/editor packages behind tests.
2. Replace the CLI commands, then hooks and installers.
3. Update docs and `.gitattributes`/`.gitignore` for Go.
4. Existing repositories run `pkt init` once; it rewrites the hooks and
   records `pktgit.bin`. Old `pktgit.py`/`pktgit.script` keys are ignored and
   can be removed.
5. Rollback: check out the `v1-python` tag and reinstall the Python tool; the
   snapshot layout and `.pkt` format are unchanged.

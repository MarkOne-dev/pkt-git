# Tasks

## 1. Foundation and build tooling

- [x] 1.1 Confirm the Go module `github.com/MarkOne-dev/pkt-git`, Cobra root, and `cmd/pkt` scaffold build with `go build ./...`
- [x] 1.2 Add and pin `golang.org/x/crypto` for Twofish, run `go mod tidy`, and verify `go build ./...` and `go test ./...` succeed
- [x] 1.3 Add `internal/config` build metadata and repository configuration keys, then verify `go run ./cmd/pkt version` prints the injected version
- [x] 1.4 Add `internal/fsx` atomic same-directory write helper with tests for overwrite, permissions, and failure cleanup; verify with `go test ./internal/fsx`
- [x] 1.5 Add `Makefile` targets for `build`, `test`, `fmt`, `vet`, `lint`, `dist`, and `clean`; verify `make test` and `make build` succeed
- [x] 1.6 Update `.gitignore` for Go build artifacts and `.gitattributes` for `.pkt`, `.pka`, and snapshot text rules; verify with `git check-attr binary text -- testdata/specimens/Caso1_Grupo09-2.pkt snapshot/demo/base.snapshot.xml`

## 2. Crypto and codec

- [x] 2.1 Implement CMAC/OMAC1 in `internal/crypto/cmac` and verify NIST SP 800-38B known-answer tests with `go test ./internal/crypto/cmac`
- [x] 2.2 Implement EAX encryption, decryption, and tag verification in `internal/crypto/eax`; verify round-trip and tamper tests with `go test ./internal/crypto/eax`
- [x] 2.3 Implement the qCompress framing and both positional XOR layers, including the reversed outer index; verify table-driven tests with `go test ./internal/codec`
- [x] 2.4 Implement `Decode`, `Encode`, and `Verify` in `internal/codec` using `golang.org/x/crypto/twofish`; verify against `testdata/specimens/Caso1_Grupo09-2.pkt` with `go test ./internal/codec`
- [ ] 2.5 Add malformed, truncated, tag-mismatch, and length-mismatch tests; verify each fails without producing usable output using `go test ./internal/codec`
- [ ] 2.6 Update `docs/packet-tracer-format.md` for the Go implementation and authenticated-failure behavior; verify the documented `pkt verify` example against the specimen

## 3. Topology and snapshot

- [ ] 3.1 Implement ordered device, link, and endpoint parsing in `internal/topology`; verify devices, links, and unresolved references with `go test ./internal/topology`
- [ ] 3.2 Implement device-name sanitization and collision detection; verify safe names and duplicate-name failures with `go test ./internal/topology`
- [ ] 3.3 Implement deterministic `metadata/topology.json` generation from typed structs; verify field order, totals, and fallback endpoints with `go test ./internal/topology`
- [ ] 3.4 Implement `Export` with decode/parse-before-write and atomic file writes; verify a first export and a failed refresh that preserves the previous snapshot with `go test ./internal/snapshot`
- [ ] 3.5 Implement `Build` with XML validation and atomic output; verify invalid XML leaves an existing binary unchanged with `go test ./internal/snapshot`
- [ ] 3.6 Implement `Sync` with missing-directory creation, stale config pruning, and no XML modification; verify with `go test ./internal/snapshot`
- [ ] 3.7 Implement read-only `Inspect`; verify it writes no files with `go test ./internal/snapshot`
- [ ] 3.8 Add golden tests for the specimen's XML hash, config set, and metadata hash; verify with `go test ./internal/snapshot -run Golden`
- [ ] 3.9 Update `docs/snapshot-format.md` and `docs/architecture.md` for the Go layout and snapshot invariants; verify the documented export/sync commands

## 4. XML editing and redaction

- [ ] 4.1 Implement the byte-offset XML index in `internal/xmledit` using `encoding/xml` offsets; verify device spans, line spans, self-closing tags, and CRLF input with `go test ./internal/xmledit`
- [ ] 4.2 Implement read-only `show` with credential redaction; verify secret lines are masked and the file is unchanged with `go test ./internal/editor`
- [ ] 4.3 Implement `setiface` block replacement; verify the block is replaced and surrounding bytes are preserved with `go test ./internal/editor`
- [ ] 4.4 Implement `chgline` full-line matching with zero-match and ambiguous-match failures; verify substring rejection and multiple-match rejection with `go test ./internal/editor`
- [ ] 4.5 Implement `insertafter` insertion after the first matching line; verify the insertion point and unchanged remainder with `go test ./internal/editor`
- [ ] 4.6 Add post-edit XML validation, atomic write, and automatic derived-file sync; verify invalid edits leave the original bytes and sync failures are reported with `go test ./internal/editor`
- [ ] 4.7 Implement the centralized redaction policy in `internal/security`; verify password, secret, key, and SNMP community cases with `go test ./internal/security`
- [ ] 4.8 Update `docs/cli-reference.md` and `docs/troubleshooting.md` for the new editing behavior; verify the documented examples against a temporary snapshot

## 5. Git integration

- [ ] 5.1 Implement the `internal/gitx` wrapper with NUL-separated path parsing, config get/set, diff, checkout, and add; verify paths with spaces and revisions with `go test ./internal/gitx`
- [ ] 5.2 Implement `pkt init` hook generation, `pktgit.bin` configuration, `.gitattributes` merging, idempotency, and non-repository rejection; verify with `go test ./internal/hooks`
- [ ] 5.3 Implement `pre-commit` staged-binary export, automatic staging, failure abort, and text-only warning; verify with `go test ./internal/hooks`
- [ ] 5.4 Implement `post-merge` and `post-checkout` changed-snapshot detection, squash-flag handling, zero-old-revision handling, and atomic rebuilds; verify with `go test ./internal/hooks`
- [ ] 5.5 Implement `resolve` report mode, `--ours`/`--theirs`, marker verification, binary-only conflicts, and scoped staging; verify with `go test ./internal/hooks`
- [ ] 5.6 Add temporary-repository integration tests covering init, commit, merge, checkout, squash, and conflict resolution; verify with `go test ./internal/hooks -run Integration`
- [ ] 5.7 Update `docs/workflow.md`, `docs/installation.md`, and `docs/cli-reference.md` for the Go hooks and `pkt init`; verify the documented migration and conflict commands

## 6. CLI

- [ ] 6.1 Wire the Cobra root with help, no-argument behavior, typed usage errors, and exit codes 0/1/2; verify with `go test ./internal/cli`
- [ ] 6.2 Implement `export`, `build`, and `inspect` commands with exact argument validation; verify success and usage failures with `go test ./internal/cli`
- [ ] 6.3 Implement the `edit` command tree and its subcommands; verify argument validation and dispatch with `go test ./internal/cli`
- [ ] 6.4 Implement `sync`, `resolve`, and `verify` commands; verify flags, policy validation, and exit codes with `go test ./internal/cli`
- [ ] 6.5 Implement `init`, `version`, and hidden `pre-commit`/`post-merge`/`post-checkout` commands; verify help visibility and version output with `go test ./internal/cli`
- [ ] 6.6 Implement `skills` with an embedded skill asset and install/remove behavior for OpenCode, Claude Code, and Gemini CLI; verify with `go test ./internal/skills`
- [ ] 6.7 Add CLI integration tests for `pkt --help`, missing arguments, unknown commands, and version; verify with `go test ./internal/cli -run Integration`
- [ ] 6.8 Update `README.md` and `docs/cli-reference.md` for the Go command surface; verify the documented commands with `go run ./cmd/pkt --help`

## 7. Installers, release, and documentation

- [ ] 7.1 Rewrite `scripts/install.sh` to install the compiled binary without Python or a virtualenv, including PATH handling and idempotency; verify with `bash -n` and a temporary install prefix
- [ ] 7.2 Rewrite `scripts/install.ps1` for the compiled Windows binary with PATH handling and idempotency; verify with PowerShell parsing and the Windows CI job
- [ ] 7.3 Add `make dist` cross-compilation with version, commit, and date linker flags; verify Linux, macOS, and Windows artifacts are produced
- [ ] 7.4 Update `docs/installation.md` for binary installation and migration from Python; verify the documented install and `pkt init` steps
- [ ] 7.5 Update the `README.md` badges, quickstart, and CLI reference for Go; verify all internal links and commands

## 8. CI and end-to-end verification

- [ ] 8.1 Add a GitHub Actions matrix for Ubuntu, macOS, and Windows running formatting, vet, tests, and build; verify the workflow syntax and a local equivalent
- [ ] 8.2 Run `gofmt -l .`, `go vet ./...`, `go test ./...`, and `go build ./...`; verify all commands are clean
- [ ] 8.3 Run an end-to-end specimen flow: `export`, `sync`, `build`, `verify`, and `inspect`; verify round-trip equality and unchanged source snapshots
- [ ] 8.4 Run an end-to-end hook flow in a temporary repository: `init`, commit a `.pkt`, merge, checkout, and resolve a conflict; verify snapshots, staged paths, and rebuilt binaries
- [ ] 8.5 Remove remaining Python, virtualenv, and old tool references outside historical documentation; verify with `rg -n 'python|\\.venv|ptg\\.py|edit_config|spike_roundtrip' --glob '!openspec/**'`
- [ ] 8.6 Review `git status`, confirm no secrets or generated artifacts are staged, and verify the final diff against the specifications

## Workflow follow-up

- Archive the change after review and verification are complete.
- Verify the archived result and the main capability specs.

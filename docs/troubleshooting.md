# Troubleshooting & Pitfalls

Every entry here is a verified behavior of the current code or a real incident
from development. Untested areas are flagged as such.

## Verification Commands (start here)

```bash
# Verify host environment prerequisites, tool installation, and PATH:
pkt doctor

# Container round trip against the bundled specimen.
# Expect XML is a document: true and XML round trip: true.
pkt verify testdata/specimens/Caso1_Grupo09-2.pkt

# Codec-level check while working on the Go packages:
go test ./internal/codec -run TestVerifySpecimen -v

# Is the installed binary healthy?
pkt --version

# Is a snapshot still valid XML after hand editing?
pkt sync snapshot/<project>    # fails on malformed XML and does not write it
```

## Symptoms and Fixes

### `pkt verify` reports an authentication failure

An EAX tag mismatch or a declared-length mismatch is a **hard failure**. The
codec returns no XML from a failed authentication check.

- The tag covers the ciphertext in Packet Tracer's EAX variant.
- A mismatch means a container layer was decoded or re-encoded incorrectly.
- Do not bypass tag verification. Recheck the stage-4 inversion described in
  `packet-tracer-format.md` (the XOR index is the *output* position).

### `XML round trip : false`

The round trip must always be true; this is the project invariant. Usual causes
are invalid XML before encoding or line-ending corruption. Re-export from the
original `.pkt` and retry before changing code.

### `byte-identical : false`

**Expected, not a bug.** Packet Tracer's zlib emitter uses different compression
parameters. A rebuilt file is valid when Packet Tracer opens it and the XML
round trip is exact. Byte identity would require matching Packet Tracer's exact
compressor.

### `pre-commit` fails with an export error

Read the hook output. Common causes:

- Packet Tracer is still saving, or the `.pkt` came from an untested version.
- The installed `pkt` binary is missing or stale. Reinstall the Go binary.
- The staged file is truncated or not a valid container.

The hook aborts the commit and leaves the previous snapshot unchanged.

### Hook printed `textual snapshot changed without its .pkt`

This is informational. A text-only edit or conflict resolution is valid, but
the binary is stale. Rebuild it before opening Packet Tracer:

```bash
pkt build snapshot/<project>/base.snapshot.xml \
    topologia/<project>.pkt
git add topologia/<project>.pkt
```

### The rebuilt `.pkt` does not reflect my textual edits

Check that the edit targeted `RUNNINGCONFIG`, not `STARTUPCONFIG`. They are
separate blocks with separate contents; `pkt edit` only operates on the running
config.

```bash
pkt edit snapshot/<project>/base.snapshot.xml show <DEVICE>
```

If the line is present in `show` but missing in Packet Tracer, rebuild the
binary with `pkt build`.

### `device not found` from `pkt edit`

Device names are exact matches on `<NAME translate="true">X</NAME>`. Check the
real name in `metadata/topology.json`; case and spelling matter
(`Sw1_LIMA` ≠ `sw1_lima`).

### `device name is ambiguous` from `pkt edit`

More than one device has the same exact name. Fix the duplicate in the snapshot
before editing; the command refuses to guess which device was intended.

### `device has no RUNNINGCONFIG`

End devices such as PCs and servers usually have no CLI config. There is nothing
for `pkt edit` to change; the snapshot was left unchanged.

### `interface not found`

The exact `interface <IFACE>` line is not in that device's running config. Run
`pkt edit ... show <DEVICE>` and copy the interface name exactly, including case.

### `interface block has no closing ! line`

The running config is malformed: an `interface` line has no following `!` line.
Add the closing `!` to the snapshot or restore the device config from a known
good revision. No edit is written.

### `line not found` from `chgline`

`chgline` matches the **complete** line value, including leading spaces. A
substring does not match. Use `show` to copy the exact line, or use `setiface`
to replace the whole block.

### `line is ambiguous` from `chgline`

More than one line in that interface block equals the old value. No line is
changed. Make the old value unique, or use `setiface` to replace the block
deliberately.

### `phrase not found` from `insertafter`

The phrase appears in no running-config line for that device. Use `show` to find
the exact text; matching is case-sensitive and searches complete line values.

### `resulting XML is invalid`

The proposed edit would produce malformed XML. The original snapshot was not
written. This usually means a hand-edited snapshot was already invalid; restore
a valid snapshot and retry.

### `derived files may be stale`

The XML edit was written, but refreshing `configs/*.cfg` or
`metadata/topology.json` failed. Fix the filesystem problem (for example, a path
that should be a directory but is a file) and run:

```bash
pkt sync snapshot/<project>
```

### A config line containing a line break was rejected

`pkt edit` accepts one IOS line per argument. Split the input into separate
arguments instead of embedding `\n` or `\r`.

### `show` displays `<redacted>`

This is intentional. `show` masks values after credential keywords such as
`password`, `secret`, `community`, `key`, `key-string`, `pre-shared-key`, and
`wpa-psk`. The snapshot itself is not modified. Use a Packet Tracer session if
you need the real value.

### Merge conflict markers left the snapshot invalid

A conflicted `base.snapshot.xml` contains `<<<<<<<` blocks and is not valid XML
until resolved. Never stage a snapshot with unresolved markers. Use
`pkt resolve --ours` or `pkt resolve --theirs`, or resolve the text manually,
then run `pkt sync` and rebuild the binary.

## Dangerous Patterns (do not do)

- **Global find-replace or `sed` over `base.snapshot.xml`.** A real incident
  duplicated a VLAN assignment across a 5.9 MB document. Edits must be scoped
  per device; `pkt edit` exists for this reason.
- **Using `setiface` for a one-line change.** It replaces the whole interface
  block. Use `chgline` for surgical edits.
- **Editing Packet Tracer files while the tool writes them.** Finish saving (or
  close Packet Tracer) before committing; Packet Tracer holds the file mid-write.
- **Assuming hooks exist before running `pkt init`.** A repository without
  `pkt init` silently has no automation.
- **Assuming hooks activate automatically after `git clone`.** For security reasons,
  Git never activates hooks on clone; every teammate must run `pkt init` once
  locally.
- **Editing `STARTUPCONFIG` when you meant `RUNNINGCONFIG`.** They are separate
  blocks; `pkt edit` only changes the running config.

## Compatibility Limits (tested vs. unknown)

| Area | Status |
|---|---|
| Packet Tracer **9.0.1.0858** `.pkt` | Verified: full round trip on a real 107-device project. |
| PT 8.x / 7.x / 6.x | Untested. The container scheme is expected to match; verify with `pkt verify` on a specimen before relying on it. |
| `.pka` activity files | Untested. Same container expected; inner document differs. |
| Windows / macOS | Supported by design with the pure-Go binary and hooks. Physical validation is pending the first Windows/macOS teammate. |
| PT newer than the pinned 9.0.1.0858 | The PT version is pinned team-wide; resaving with a newer PT may produce an unverified format variant. |

## Reporting a Broken Specimen

1. Run `pkt verify <file>` and keep the full report.
2. Note the exact Packet Tracer version that wrote the file (Help → About).
3. Place the specimen under `testdata/specimens/` only if it contains no
   confidential topology; otherwise attach a minimal reproduction.
4. Open an issue or PR describing both the report and the producing version.

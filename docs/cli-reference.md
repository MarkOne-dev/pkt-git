# CLI Reference

`pkt` is a single Go binary. It has no Python, virtualenv, cgo, or C runtime
dependency. The command tree is built with Cobra and implemented under
`internal/cli`; the feature packages live under `internal/`.

## Command surface

```bash
pkt init [dir]
pkt export <project.pkt> <output_dir>
pkt build  <snapshot.xml> <output.pkt>
pkt inspect <project.pkt>

pkt edit <xml> show <DEVICE>
pkt edit <xml> setiface <DEVICE> <IFACE> "line 1" "line 2" ...
pkt edit <xml> chgline <DEVICE> <IFACE> <old_line> <new_line>
pkt edit <xml> insertafter <DEVICE> <phrase> "line 1" ...

pkt verify <file.pkt>
pkt sync   <snapshot-dir|snapshot.xml>
pkt resolve [--ours|--theirs]
pkt skills [--remove]
pkt version

# Hidden hook entry points, invoked by generated hooks:
pkt pre-commit
pkt post-merge [squash-flag]
pkt post-checkout <old> <new> <branch-flag>
```

`pkt --help` lists the public commands and omits the hook entry points. The CLI
never prompts for input: choices such as `resolve --ours` must be explicit.

### Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Runtime failure (decode, filesystem, Git, validation, synchronization) |
| `2` | Usage error (missing/invalid argument or unknown command) |

Errors are written to standard error and identify the failing operation and
path. No stack traces or credentials are printed.

## `init`

Wires pkt-git into an existing Git repository:

```bash
pkt init
pkt init path/to/repository
```

- Writes executable `.githooks/pre-commit`, `.githooks/post-merge`, and
  `.githooks/post-checkout`.
- Records the absolute path of the running `pkt` binary in repository-local
  Git configuration (`pktgit.bin`).
- Sets `core.hooksPath` to `.githooks`.
- Creates or merges `.gitattributes`: `.pkt`/`.pka` are binary; snapshot XML,
  JSON, and config files are text.
- Refuses to modify a directory that is not a Git repository.
- Is idempotent: rerunning it leaves the same usable configuration and does not
  duplicate attribute rules.

The hooks are repository-local. `pkt init` does not change global Git
configuration and does not affect other repositories.

## `export`

```bash
pkt export testdata/specimens/Caso1_Grupo09-2.pkt snapshot/Caso1_Grupo09-2
```

Creates or refreshes:

```
snapshot/<Project>/
├── base.snapshot.xml
├── configs/<Device>.cfg
└── metadata/topology.json
```

- The decoded XML is written unchanged.
- One `.cfg` is written for each device whose `RUNNINGCONFIG` has at least one
  line. Configs preserve line order and text and end with a newline.
- Device names are sanitized to safe file names. Two names that sanitize to the
  same file name are an error naming both devices.
- `metadata/topology.json` records the source file, document root, totals, and
  ordered device/link catalogs. Unresolved cable endpoints get a deterministic
  fallback name instead of failing the export.
- The command reports the XML size, config count, device count, and link count.
- Decoding and parsing finish before any output is written. A failed refresh
  leaves the previous snapshot readable and unchanged.

## `build`

```bash
pkt build snapshot/Caso1_Grupo09-2/base.snapshot.xml \
    topologia/Caso1_Grupo09-2.pkt
```

Validates the XML, encodes a Packet Tracer binary, and writes it atomically.
The source snapshot is not modified. If validation or encoding fails, an
existing binary at the output path remains unchanged.

A successful build decodes back to exactly the input XML. Byte-for-byte
identity with Packet Tracer's own encoder is not required.

## `inspect`

```bash
pkt inspect testdata/specimens/Caso1_Grupo09-2.pkt
```

Decodes the binary and prints a read-only summary: device and link totals plus
one row per device with type, model, and CLI line count. It creates or modifies
no file.

## `edit`

Every mutating edit is scoped to exactly one device, identified by the exact
`<NAME>` value, and only touches that device's `RUNNINGCONFIG`. There is no
document-wide replacement.

### `show`

```bash
pkt edit snapshot/Caso1_Grupo09-2/base.snapshot.xml \
    show SW_ACC_ADMIN_PUNO
```

Prints the running config in order without writing any file. Credential values
are redacted before printing. The redacted text is never written back to the
snapshot.

Redacted keywords include `password`, `secret`, `community`, `key`,
`key-string`, `pre-shared-key`, and `wpa-psk`. The value after the keyword is
replaced with `<redacted>`:

```
 enable secret <redacted>
 username admin password <redacted>
 snmp-server community <redacted>
```

`no service password-encryption` is not treated as a secret.

### `setiface`

```bash
pkt edit snapshot/Caso1_Grupo09-2/base.snapshot.xml \
    setiface SW_ACC_ADMIN_PUNO FastEthernet0/3 \
    " switchport access vlan 60" \
    " switchport mode access"
```

Replaces everything between the exact `interface FastEthernet0/3` line and the
next `!` line. The block becomes exactly the supplied lines. All content outside
the block is preserved byte-for-byte.

Use `setiface` when replacing a whole block. For a one-line change, prefer
`chgline`.

### `chgline`

```bash
pkt edit snapshot/Caso1_Grupo09-2/base.snapshot.xml \
    chgline SW_ACC_ADMIN_PUNO FastEthernet0/3 \
    " switchport access vlan 50" \
    " switchport access vlan 60"
```

Replaces one complete `<LINE>` value inside the specified interface block.

- The old value must match the full line text, including leading spaces.
- A substring match is rejected.
- Zero matches fail with `line not found`.
- More than one match fails with `line is ambiguous`; no line is changed.
- Only the matching line's text changes. Indentation, the `<LINE>` tags, and the
  line ending are preserved.

### `insertafter`

```bash
pkt edit snapshot/Caso1_Grupo09-2/base.snapshot.xml \
    insertafter SW_ACC_ADMIN_PUNO "hostname SW_ACC_ADMIN_PUNO" \
    " ntp server 1.1.1.1"
```

Inserts the supplied lines immediately after the first config line containing
the phrase. The original matching line and every other line remain in place.
Inserted lines use the matched line's indentation and line-ending convention.
No match fails with `phrase not found`.

### Editing guarantees

Every mutating command (`setiface`, `chgline`, `insertafter`) follows the same
pipeline:

1. Build a byte-offset index of the XML and find the exact device and interface.
2. Compute the new document in memory.
3. Verify the result is well-formed XML before writing.
4. Write the snapshot atomically, preserving the file mode.
5. Refresh `configs/*.cfg` and `metadata/topology.json` from the edited XML.

If the transform fails, validation fails, or the write fails, the original
snapshot is left byte-for-byte unchanged. If the XML write succeeds but derived
file synchronization fails, the command exits non-zero and reports that the
snapshot was changed but derived files may be stale.

Input lines are XML-escaped (`&`, `<`, `>`). A line containing a raw line break
is rejected. Untouched spans keep their original encoding and line endings,
including CRLF snapshots.

### Editing errors

| Error | Meaning |
|---|---|
| `device not found` | No exact `<NAME>` match. Check `metadata/topology.json`; names are case-sensitive. |
| `device name is ambiguous` | More than one device has the same exact name. Fix the snapshot first. |
| `device has no RUNNINGCONFIG` | End devices such as PCs/servers have no CLI config to edit. |
| `interface not found` | The exact `interface <IFACE>` line is absent from that device's running config. |
| `interface block has no closing ! line` | The config is malformed; add the closing `!` or fix the snapshot. |
| `line not found` | `chgline` old value did not match a full line in that interface block. |
| `line is ambiguous` | More than one line in that block equals the old value. |
| `phrase not found` | `insertafter` phrase appears in no config line. |
| `resulting XML is invalid` | The proposed edit would produce malformed XML; the original file was not written. |
| `derived files may be stale` | The XML was written but `pkt sync` failed. Fix the filesystem issue and run `pkt sync`. |

After a successful textual edit, the `.pkt` binary is stale. Rebuild it with
`pkt build` before opening Packet Tracer.

## `sync`

```bash
pkt sync snapshot/Caso1_Grupo09-2
pkt sync snapshot/Caso1_Grupo09-2/base.snapshot.xml
```

Regenerates `configs/*.cfg` and `metadata/topology.json` from the snapshot XML.

- Creates missing `configs/` and `metadata/` directories.
- Writes one config per device with a non-empty running config.
- Removes stale generated `.cfg` files that no longer correspond to a device.
- Preserves an existing `source_file` value in `metadata/topology.json`.
- Never writes `base.snapshot.xml`.
- Reports refreshed counts and removed config file names.

`pkt edit` runs the same synchronization automatically after a successful
mutation.

## `verify`

```bash
pkt verify testdata/specimens/Caso1_Grupo09-2.pkt
```

Decodes the specimen, re-encodes the decoded XML, decodes the rebuilt binary,
and reports:

```
binary size        : 598810
extracted XML      : 5894851
rebuilt binary     : 585372
XML is a document  : true
XML round trip     : true
byte-identical     : false
```

- `XML is a document` and `XML round trip` are required; the command exits `0`
  only when both are true.
- `byte-identical` is informational. Packet Tracer's zlib emitter uses
  different compression parameters; a valid rebuilt file need not match its
  bytes.
- Authentication-tag or declared-length mismatches are hard failures. No XML is
  returned from a failed decode.

## `resolve`

```bash
pkt resolve            # report mode
pkt resolve --ours     # accept the current branch
pkt resolve --theirs   # accept the incoming branch
```

Report mode lists conflicted text and binary files, including the first conflict
marker line for text files, without modifying or staging anything. It exits
non-zero after printing the report; scripts must choose a policy.

With a policy, `pkt resolve` applies the chosen side to conflicted text files,
verifies that no conflict markers remain, refreshes derived snapshot files,
rebuilds conflicted binaries from the resolved snapshot, and stages only the
paths involved in the resolution. It never uses a repository-wide `git add`.

Running `pkt resolve` without an active merge exits non-zero with a
"nothing to resolve" message.

## `skills`

```bash
pkt skills
pkt skills --remove
```

Installs or removes the bundled pkt-git skill for OpenCode, Claude Code, and
Gemini CLI. Both operations are idempotent and report each agent's outcome.
Removal targets only the pkt-git skill directories and leaves other skills
untouched. If the embedded skill asset is unavailable, the command fails with a
reinstall instruction and does not delete an existing installed skill.

## `version`

```bash
pkt version
pkt --version
```

Prints the release version plus commit and build date when injected by the
build. It works outside a Git repository.

## Hook entry points

Generated hooks call these commands; they are hidden from normal help:

```bash
pkt pre-commit
pkt post-merge [squash-flag]
pkt post-checkout <old> <new> <branch-flag>
```

- `pre-commit` exports staged `.pkt`/`.pka` files, stages the regenerated
  snapshot, and aborts the commit on export failure. A text-only snapshot change
  is allowed with a warning that the binary is stale.
- `post-merge` and `post-checkout` detect snapshots changed by incoming history
  and rebuild the corresponding `topologia/<project>.pkt` atomically. They
  handle the squash flag and checkout revision pairs and leave unchanged
  projects alone.

## Why device-scoped editing

Snapshots are multi-megabyte XML documents. Global find-replace tools such as
`sed` or editor-wide replace have already duplicated a VLAN assignment across a
real project. Every mutation therefore goes through `pkt edit`, which:

1. Limits the blast radius to one device's running config.
2. Validates the whole document before writing.
3. Produces a reviewable text diff.
4. Refreshes derived files so the snapshot cannot contradict itself.

The editing examples above are exercised against a temporary snapshot by
`go test ./internal/editor`.

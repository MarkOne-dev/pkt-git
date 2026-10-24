# Snapshot Format

The snapshot is the textual, human-reviewable representation of a Packet Tracer
project, written by `pkt export` and stored inside the repository under
`snapshot/<ProjectName>/`. It is the **authoritative** state of the network;
the `.pkt` binary is derived from it.

```
snapshot/Caso1_Grupo09-2/
├── base.snapshot.xml          # Complete decrypted XML document (~5.9 MB)
├── configs/                   # One .cfg per device that has a CLI config
│   ├── Sw1_LIMA.cfg
│   ├── SW_ACC_ADMIN_PUNO.cfg
│   └── ...
└── metadata/
    └── topology.json          # Device and link catalog for quick inspection
```

The Go implementation owns this contract in `internal/snapshot`, using
`internal/codec` for the container and `internal/topology` for the XML model.

## base.snapshot.xml

The full decrypted XML document, byte-for-byte what Packet Tracer holds in the
file. Key structural facts (learned from the PT 9.0.1.0858 schema):

- Root element: `<PACKETTRACER5>`.
- The network lives under `NETWORK` with two central sections:
  - `DEVICES/DEVICE` — one element per device.
  - `LINKS/LINK` — one element per cable.
- Inside a device, everything hangs off the `<ENGINE>` element:
  - `<TYPE model="...">` — device type (router/switch/…) and model.
  - `<NAME translate="true">Sw1_LIMA</NAME>` — device name.
  - `<SAVE_REF_ID>` — the ID that cables use to reference the device.
  - `<RUNNINGCONFIG>` vs `<STARTUPCONFIG>` — **two distinct blocks**, each
    holding IOS CLI lines as repeated `<LINE>` elements. A common early mistake
    was editing the wrong one; check which one you need (running = what PT
    shows in `show running-config`).
- Cables reference devices via `save-ref-id:<N>` inside `<FROM>` / `<TO>`, and
  carry the port names in `<PORT>` elements.

`pkt export` writes the decoded XML unchanged. Future snapshot edits in
`internal/xmledit` and `internal/editor` operate on byte spans and validate the
whole document before replacing the file, so an invalid edit leaves the
original bytes intact.

## configs/<Device>.cfg

Flat text of every device's `RUNNINGCONFIG`, one IOS line per text line. This
is the file whose diffs humans actually review; on GitHub a config change looks
like an ordinary code diff:

```diff
 interface FastEthernet0/3
  description Connection to Dept Admin
- switchport access vlan 50
+ switchport access vlan 60
  switchport mode access
  spanning-tree portfast
```

Rules of the extract (`internal/snapshot` and `internal/topology`):

- Only devices with a non-empty `RUNNINGCONFIG` get a `.cfg` (end devices like
  PCs have none).
- Device names are sanitized to safe file names: anything outside
  `[A-Za-z0-9_.-]` becomes `_`, leading/trailing `._` is trimmed, and an empty
  result becomes `device`.
- Two devices that sanitize to the same file name are an error naming both
  devices; the export never overwrites one config with another.
- Line order and content are preserved exactly. The regression invariant is
  that `configs/*.cfg` and `metadata/topology.json` stay byte-identical across
  refactors of the parser; `internal/snapshot/golden_test.go` pins their hashes
  for the bundled specimen.

## metadata/topology.json

A compact catalog generated on every export and refresh:

```json
{
  "source_file": "Caso1_Grupo09-2.pkt",
  "document_root": "PACKETTRACER5",
  "total_devices": 107,
  "total_links": 106,
  "devices": [ { "name": "Sw1_LIMA", "type": "SWITCH", "model": "...", "ref": "..." } ],
  "links":   [ { "link_type": "copper",
                 "from": { "name": "Sw1_LIMA", "port": "FastEthernet0/1" },
                 "to":   { "name": "PC_ADMIN",  "port": "FastEthernet0" } } ]
}
```

The JSON is emitted from typed structs, not maps, so field order and indentation
are deterministic. `pkt sync` preserves an existing `source_file` value and
otherwise derives it from the project directory.

Cable endpoints whose device reference cannot be resolved keep a truncated
fallback name (the last 8 characters of the ref id) rather than failing the
export.

## Representation Rules in Git

From `.gitattributes` / `.gitignore`:

- `*.pkt` and `*.pka` are marked **binary** — Git will not attempt text diffs or
  merges on the binary artifact.
- Snapshot files (`*.xml`, `*.cfg`, `*.json`) are normal text — line diffs and
  three-way merges are enabled (this is the property that makes PR reviews and
  conflict resolution possible).
- Go build output (`bin/`, `dist/`), Go workspace files, and local scratch
  output (`out/`) are gitignored. Sources, tests, scripts, docs, and snapshots
  are tracked.

## Synchronization Invariants

1. Committing a changed `.pkt` without its snapshot is impossible in the
   normal flow: the pre-commit hook exports and stages it first.
2. Committing a changed snapshot without its `.pkt` is allowed (textual edit
   or merge), with an explicit warning that the binary is stale.
3. After any merge/pull/checkout that changes a snapshot, the post-merge /
   post-checkout hooks rebuild `topologia/<Project>.pkt` locally.
4. `pkt sync <snapshot-dir|snapshot.xml>` regenerates `configs/*.cfg` and
   `metadata/topology.json`, creates missing derived directories, and removes
   stale `configs/*.cfg` files. It never writes `base.snapshot.xml`.
5. `pkt build <snapshot.xml> <output.pkt>` validates the XML, encodes it, and
   writes the output atomically. Invalid XML leaves an existing binary
   unchanged.
6. `pkt inspect <project.pkt>` is read-only: it decodes and summarizes the
   topology without creating or modifying any file.

## Verification

The snapshot contract is exercised against the bundled specimen:

```bash
go test ./internal/snapshot -run 'Golden|Export|Sync|Build|Inspect'
```

The package-level equivalents of the user commands are:

```bash
pkt export testdata/specimens/Caso1_Grupo09-2.pkt snapshot/Caso1_Grupo09-2
pkt sync snapshot/Caso1_Grupo09-2
pkt build snapshot/Caso1_Grupo09-2/base.snapshot.xml topologia/Caso1_Grupo09-2.pkt
```

The CLI wiring for these commands is completed in the CLI phase; the snapshot
tests above verify the same operations end to end at the package boundary.

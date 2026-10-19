# CLI Reference

All commands run inside the project virtual environment. POSIX path is
`.venv/bin/python`; on Windows it is `.venv\Scripts\python.exe` (hooks resolve
this automatically via `.githooks/common.sh`).

```
tools/
├── spike_roundtrip.py   # Codec layer: decode / encode / verify
├── ptg.py               # Topology layer: export / build / inspect
└── edit_config.py       # Editing layer: show / setiface / chgline / insertafter
```

---

## `tools/spike_roundtrip.py`

Reversible codec for `.pkt`/`.pka` files. See `packet-tracer-format.md` for the
format behind it.

```bash
.venv/bin/python tools/spike_roundtrip.py decode <file.pkt> <out.xml>
.venv/bin/python tools/spike_roundtrip.py encode <file.xml> <out.pkt>
.venv/bin/python tools/spike_roundtrip.py verify <file.pkt>
```

| Command | Behavior |
|---|---|
| `decode SRC OUT` | Writes the decrypted XML document to `OUT`. Warnings (non-fatal) on EAX tag mismatch or declared-length mismatch. |
| `encode SRC OUT` | Rebuilds a valid binary from an XML document (deflate level 9, Crypto++-style EAX). |
| `verify SRC` | Round-trip test: `decode → encode → decode`. Prints a report; exit code 0 only if the XML round trip is exact and the output is a document. |

`verify` report fields:

```
specimen / binary size / extracted XML / rebuilt binary
XML is a document : True        # required
XML ROUND TRIP    : True        # required — project invariant
byte-identical    : False       # bonus only; needs PT's own zlib emitter
```

API entry points (importable): `decode(raw: bytes) -> str` and
`encode(xml_text: str) -> bytes`, plus the in-house `cmac`, `omac`,
`eax_encrypt`, `eax_decrypt` primitives.

---

## `tools/ptg.py`

CLI and topology parser. Works from real binaries; knows nothing about crypto
beyond calling `decode`/`encode`.

```bash
.venv/bin/python tools/ptg.py export <project.pkt> <output_dir>
.venv/bin/python tools/ptg.py build  <snapshot.xml> <output.pkt>
.venv/bin/python tools/ptg.py inspect <project.pkt>
```

### `export <project.pkt> <dir>`

Creates (or refreshes) the textual snapshot, printing a per-artifact summary:

```
[ok] base.snapshot.xml     : 5,891,234 B
[ok] per-device .cfg files : 34 devices with CLI config
[ok] metadata/topology.json: 107 devices, 106 links
```

- `base.snapshot.xml` — full decrypted document.
- `configs/<Device>.cfg` — one per device with a `RUNNINGCONFIG`.
- `metadata/topology.json` — device/link catalog.

This is exactly what the pre-commit hook invokes; you only call it directly for
a first import or a manual re-export.

### `build <snapshot.xml> <output.pkt>`

Re-encrypts the XML into a binary Packet Tracer file ("ready to open in PT").
Manually required only after textual edits/conflict resolution.

### `inspect <project.pkt>`

Console summary without writing any file:

```
NAME                        TYPE            MODEL         CLI
------------------------------------------------------------------
Sw1_LIMA                    SWITCH          2960          ✔ 342 lines
```

### Parsing notes

- Devices are read from `.//NETWORK/DEVICES/DEVICE/ENGINE` (`TYPE`, `NAME`,
  `SAVE_REF_ID`, `RUNNINGCONFIG`, `STARTUPCONFIG`).
- Links from `LINKS/LINK/CABLE`: endpoints resolved through
  `save-ref-id:<N>` against the device table; unresolvable refs fall back to a
  truncated ref name instead of failing.
- Internal data types are dataclasses: `Device`, `Link`, `Endpoint`.

---

## `tools/edit_config.py`

Safe, device-scoped textual editor for snapshots. Every mutating command is
scoped to a single device's `RUNNINGCONFIG`, and the resulting document is XML
validated (`ET.fromstring`) **before** the file is written.

```bash
# show the running config of one device (password lines hidden)
edit_config.py <xml> show <DEVICE>

# replace the ENTIRE interface block with the given lines
edit_config.py <xml> setiface <DEVICE> <IFACE> "line 1" "line 2" ...

# change one line inside an interface block (first occurrence)
edit_config.py <xml> chgline <DEVICE> <IFACE> <old_text> <new_text>

# insert lines right after the first config line containing <phrase>
edit_config.py <xml> insertafter <DEVICE> <phrase> "line 1" ...
```

| Command | Scope | Semantics |
|---|---|---|
| `show` | device | Prints the device's `RUNNINGCONFIG`; lines containing `PASSWORD` are suppressed. Read-only. |
| `setiface` | device + interface | Replaces everything between `interface <IFACE>` and the next `!` line. Blunt but predictable — the whole block becomes exactly the lines you pass. |
| `chgline` | device + interface | Finds `old_text` **inside that interface block only** and replaces its first occurrence. Fails loudly if the text is not in that block (never edits another interface). |
| `insertafter` | device | Inserts new lines after the first line containing `phrase`, wherever it appears in the device config. |

All failures are explicit and abort before any write (`device not found`,
`interface not in device RUNNINGCONFIG`, `line not found in <iface>`,
`has no RUNNINGCONFIG (PC/server?)`).

### Why device-scoped editing

Snapshots are large (multi-MB XML). Global replace tools (`sed`, editor-wide
find-replace) have already caused a real incident (duplicated VLAN assignment)
— the workflow forbids them. Every change goes through this tool so that:

1. The blast radius is one device block.
2. The result is validated XML.
3. The edit is reviewable as a normal text diff.

---

## `setup.sh` / `setup.ps1`

Idempotent one-command installer — full details in `installation.md`.

## Hook entry points

Not invoked by users directly; documented in `workflow.md`:

```bash
.githooks/pre-commit        # .pkt staged  -> export snapshot, stage it
.githooks/post-merge        # snapshot changed in merge -> build .pkt
.githooks/post-checkout     # snapshot changed in checkout -> build .pkt
.githooks/common.sh         # resolve_py(): venv-python path per platform
```

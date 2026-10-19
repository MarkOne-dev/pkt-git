# Troubleshooting & Pitfalls

Every entry here is either a verified behavior of the current code or a real
incident from development. Nothing speculative is listed as a fact; untested
areas are flagged as such.

## Verification Commands (start here)

```bash
# Crypto pipeline sanity against the bundled specimen:
# expect "XML ROUND TRIP : True" (exit code 0)
pkt verify proof-pkt/Caso1_Grupo09-2.pkt

# Is the global install healthy?
pkt --version   # global CLI alive

# Is a snapshot still valid XML after hand editing?
python3 -c "import xml.etree.ElementTree as ET, pathlib; \
    ET.fromstring(pathlib.Path('snapshot/<project>/base.snapshot.xml').read_text())" && echo "XML valid"

## Symptoms and Fixes

### `verify` prints an EAX MAC mismatch warning

`[warn] EAX MAC mismatch (... vs ...)`

- The tag covers the ciphertext in PT's EAX variant. A mismatch means the
  decryption produced garbage → you decoded or re-encoded a byte layer
  incorrectly. Do **not** "fix" it by skipping tag verification; instead recheck
  the stage-4 inversion (see `packet-tracer-format.md`, the reversal subtlety:
  the XOR index is the *output* position).

### `XML ROUND TRIP : False`

Round trip must always be `True` — this is the project invariant. Usual
causes: an edit left the XML invalid before encoding, or line endings were
mangled (editors converting LF/UTF-8). Export fresh and retry before touching
code.

### `byte-identical : False`

**Expected, not a bug.** PT's own zlib emitter produces different bytes;
validity is what matters (PT opens the rebuilt file identically). Only
byte-identity would require matching PT's exact compression parameters.

### `pre-commit` fails with an export error

Read the tool output the hook prints. Common causes:

- PT is mid-save or the `.pkt` was written by a different PT version (see
  Compatibility below).
- `.venv` missing or twofish broken → re-run `./install.sh`.

### Hook printed `⚠ Warning: textual snapshot changed without its .pkt`

Informational, not an error — typical after conflict resolution or
`edit_config.py`. If your intent was to change the network, finish with:

```bash
pkt build snapshot/<project>/base.snapshot.xml \
    topologia/<project>.pkt
git add topologia/<project>.pkt
```

### The rebuilt `.pkt` does not reflect my textual edits

You probably edited `STARTUPCONFIG` instead of `RUNNINGCONFIG` — they are
separate blocks with separate contents (`edit_config.py` operates on the
running one). Verify which block actually holds the line you edited:

```bash
grep -n "interface <IFACE>" snapshot/<project>/base.snapshot.xml
# then inspect both <RUNNINGCONFIG> and <STARTUPCONFIG> sections around it
```

### `device '<X>' not found in snapshot` (edit_config.py)

Device names are exact matches on `<NAME translate="true">X</NAME>`. Check the
real name in `metadata/topology.json` — case and spelling matter
(`Sw1_LIMA` ≠ `sw1_lima`).

### Merge conflict markers left the snapshot invalid

`base.snapshot.xml` inside a conflict contains `<<<<<<<` blocks and is not
valid XML until resolved. Never `git add` a snapshot with unresolved markers
(the next `pkt build` will fail). Resolve with device-scoped edits
(`view the conflict, choose one side, strip the markers around the chosen
side`), then `git add`, then rebuild.

## Dangerous Patterns (do not do)

- **Global find-replace / `sed` over `base.snapshot.xml`.** A real incident: a
  global replace duplicated a VLAN assignment across the 5.9 MB document.
  Edits must be scoped per device — `edit_config.py` exists for this reason.
- **`setiface` as a substitute for `chgline`.** `setiface` replaces the whole
  interface block; the safe default for a one-line change is `chgline`.
- **Editing PT files while the tool writes them.** Save and close (or at least
  finish saving) before committing; PT holds the file mid-write.
- **Assuming hooks exist before running `pkt init`.** The per-repository
  wiring is `pkt init`; a repository without it silently has no automation.

## Compatibility Limits (tested vs. unknown)

| Area | Status |
|---|---|
| Packet Tracer **9.0.1.0858** `.pkt` | Verified: full round trip, real 107-device project. |
| PT 8.x / 7.x / 6.x | Untested. The container scheme is expected to match; verify with `spike_roundtrip.py verify` on a specimen before relying on it. |
| `.pka` activity files | Untested. Same container expected; inner document differs. |
| Windows / macOS | Supported by design; hooks' Windows layout and `setup.ps1` implemented, but physical validation pending (first Windows teammate). `setup.ps1` needs no Git Bash. If the Twofish C binding fails to build without MSVC Build Tools, the planned pure-Python Twofish removes the issue. |
| PT newer than the pinned 9.0.1.0858 | The PT version is pinned team-wide; resaving with a newer PT may produce an unverified format variant. |

## Reporting a Broken Specimen

1. Run `spike_roundtrip.py verify <file>` and keep the full report.
2. Note the exact PT version that wrote the file (Help → About).
3. Place the specimen under `proof-pkt/` (do not commit confidential
   topologies) and open the PR describing both.

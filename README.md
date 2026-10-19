# pkt-git

Git-style versioning for **Cisco Packet Tracer** projects.

## The problem
`.pkt` files are opaque binaries: Git cannot diff or merge them, and on a
multi-person team work gets lost (everyone saving over the same file, or
multiplying `project_final_v3(2).pkt` copies).

## The solution
Every `.pkt` is, internally, a **compressed and reversibly-encrypted XML**.
If every save of the project also produces its **textual snapshot**
(XML + configs + metadata), then Git can:

- show **line-by-line diffs** of the network (IP changes, VLANs, routing…),
- perform **human-readable merges / PRs**,
- keep **full history** and a single source of truth on `main`.

The binary becomes a *regenerable artifact*; the versioned text is the
repository's source of truth.

## How it works

```
        Packet Tracer (.pkt)                          Git / GitHub
   ┌───────────────────────────┐                ┌─────────────────────┐
   │   encrypted binary        │   ptg export   │ base.snapshot.xml   │
   │   599 KB                  │ ─────────────► │ 5.9 MB of XML       │
   │   (Twofish-EAX + XOR +    │                │ configs/*.cfg       │
   │    zlib + XOR)            │   ptg build    │ (textual IOS CLI)   │
   │   ←─────────────────────  │                │ metadata/*.json     │
   └───────────────────────────┘                └─────────────────────┘
```

Nothing is installed inside Packet Tracer: the codec works on the binary
from outside, even without PT running.

## Project status

| Step | Scope | Status |
|---|---|---|
| 0 | Fixed decisions + environment | ✅ done (this file) |
| 1 | SPIKE: reversible `.pkt → XML → .pkt` | ✅ tested on a real file |
| 2 | CLI v0.1: `ptg export / build / inspect` | ✅ done |
| 3 | Git integration (`pre-commit` / `post-merge` hooks) | ✅ done |
| 4 | Collaborative workflow (branches + conflict resolution) | ✅ tested live |
| 5 | GitHub publication (PRs, `main` protection) | ⬜ |

## Fixed decisions (Step 0)

- **Required Packet Tracer version:** `9.0.1.0858` (tested). Nobody upgrades
  PT until the new version is validated against the specimens in this repo.
- **Python:** 3.12+ in a project-local venv (`.venv/`). Modern API only.
- **Crypto / file format:** in-house implementation (pure Python + the
  reference Twofish C binding). No third-party dependency for the core.
- **Snapshot layout** (what gets committed):
  ```
  base.snapshot.xml     ← full document (textual source of truth)
  configs/<NAME>.cfg    ← per-device CLI config (34 in the specimen)
  metadata/topology.json← device + link map (names + ports)
  ```

## `.pkt` format details (from our own work)

Decryption pipeline (4 reversible layers, verified against PT 9.0.1.0858
with a real 599 KB file):

1. **outer deobfuscation:** `processed[i] = raw[L-1-i] ^ (L - i*L mod 256)` —
   the XOR is indexed by the *output* position, not the input.
2. **decryption:** Twofish in **EAX** mode (encrypt-then-MAC: the 16-byte
   tag goes at the end of the blob and covers the ciphertext).
   Key `0x89*16`, IV `0x10*16`. Standard NIST CMAC/OMAC with the tweak
   prefix as a full block `00×15||t`.
3. **inner deobfuscation:** `b[i] = a[i] ^ (L2 - i mod 256)` where L2 is
   the ciphertext length (tag excluded).
4. **qCompress:** 4 big-endian bytes of the uncompressed size + zlib stream.

PT's own emitter produces a slightly different zlib stream (different zlib
build): the rebuilt file is **not byte-identical** but is valid and opens
the same.

## Usage

```bash
git clone <repo> pkt-git && cd pkt-git
./setup.sh          # venv + dependencies + patch + hooks — one command
```

Machine requirements: **Python 3.12+** and **Git**. Packet Tracer
**9.0.1.0858** only for whoever edits the network in PT (someone just
reviewing diffs on GitHub does not need it).

### Platform support

| Platform | Install | Notes |
|---|---|---|
| Linux | `./setup.sh` | fully tested |
| macOS | `./setup.sh` | needs Xcode Command Line Tools (C compiler for the Twofish binding) |
| Windows | `powershell -ExecutionPolicy Bypass -File setup.ps1` (native) or `./setup.sh` from Git Bash | hooks run under Git Bash, which ships with Git for Windows |

```bash
.venv/bin/python tools/ptg.py export "my_project.pkt" snapshot/my_project/
.venv/bin/python tools/ptg.py build  snapshot/my_project/base.snapshot.xml rebuilt.pkt
.venv/bin/python tools/ptg.py inspect "my_project.pkt"              # network summary
.venv/bin/python tools/spike_roundtrip.py verify "my_project.pkt"   # pipeline test
.venv/bin/python tools/edit_config.py snapshot/my_project/base.snapshot.xml show SW1_LIMA
```

The team's daily flow is just: **edit in PT → save → `git add` → `git commit`** —
the pre-commit hook generates the textual snapshot automatically; after a
`git pull`/merge the post-merge hook rebuilds the local `.pkt` files.
# platform support v0.8

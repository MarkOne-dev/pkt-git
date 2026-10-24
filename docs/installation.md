# Installation & Setup

`pkt` is a single Go binary. It has no Python, virtualenv, cgo, or C runtime
dependency. Install the binary once per machine, then wire each repository with
`pkt init`.

## Prerequisites

- Git 2.x (Git for Windows supplies the shell used by the hooks).
- A prebuilt `pkt` binary, or Go 1.27+ to build from source.

No Packet Tracer installation is needed to decode or encode files, and nothing
is installed inside Packet Tracer.

## Install the Binary

### Script installer (recommended)

Linux and macOS, from a clone of this repository:

```bash
git clone https://github.com/MarkOne-dev/pkt-git.git
cd pkt-git
./scripts/install.sh
```

Windows, from a PowerShell prompt in the clone:

```powershell
git clone https://github.com/MarkOne-dev/pkt-git.git
cd pkt-git
powershell -ExecutionPolicy Bypass -File scripts/install.ps1
```

The installers:

1. Use `bin/pkt` (or `bin/pkt.exe`) when it already exists, otherwise build a
   static binary with `CGO_ENABLED=0 go build ./cmd/pkt`.
2. Install `pkt` (or `pkt.exe`) into `~/.local/bin` by default.
3. Report when that directory is not on `PATH`; the Windows installer adds it
   to the user `PATH` when needed.
4. Install the bundled agent skill unless `PKTG_SKIP_SKILLS=1`.
5. Are idempotent: rerunning them replaces the binary only when it changed.

Useful overrides:

| Variable | Effect |
|---|---|
| `PKTG_BIN_DIR` | Install into a different directory. |
| `PKTG_SOURCE` | Install a specific binary instead of `bin/pkt` or a build. |
| `PKTG_SKIP_SKILLS=1` | Do not run `pkt skills`. |
| `PKTG_SKIP_PATH=1` | Windows only: do not modify or inspect `PATH`. |

### From source manually

```bash
git clone https://github.com/MarkOne-dev/pkt-git.git
cd pkt-git
make build
install -m 0755 bin/pkt ~/.local/bin/pkt
```

Ensure `~/.local/bin` is on `PATH`. On Windows:

```powershell
$env:CGO_ENABLED = "0"
go build -o pkt.exe ./cmd/pkt
# Copy pkt.exe to a directory on PATH, for example:
# $env:USERPROFILE\.local\bin
```

`make build` and `make dist` force `CGO_ENABLED=0`, so the resulting binary is
static and does not require a C toolchain.

### From release artifacts

`make dist` produces self-contained binaries:

```text
dist/pkt_linux_amd64
dist/pkt_linux_arm64
dist/pkt_darwin_amd64
dist/pkt_darwin_arm64
dist/pkt_windows_amd64.exe
```

Place the artifact for your platform on `PATH` as `pkt` (or `pkt.exe`).

### Verify the install

From a clone, so the specimen is available:

```bash
pkt --version
pkt verify testdata/specimens/Caso1_Grupo09-2.pkt
```

The specimen check must report `XML is a document  : true` and
`XML round trip     : true`. A `byte-identical     : false` result is expected
for many files because Packet Tracer's own zlib emitter differs.

## Wire a Repository

```bash
cd my-existing-pt-project
pkt init
```

`pkt init` performs, without copying a template:

1. Writes executable thin hooks into `.githooks/`:
   ```sh
   #!/bin/sh
   exec "$(git config --get pktgit.bin || command -v pkt)" pre-commit "$@"
   ```
   The `post-merge` and `post-checkout` variants substitute their command name.
2. Records the absolute path of the running `pkt` binary in repository-local
   `pktgit.bin`.
3. Activates the hooks with `core.hooksPath .githooks`.
4. Creates or merges `.gitattributes`:
   ```
   *.pkt binary
   *.pka binary
   snapshot/**/base.snapshot.xml text
   snapshot/**/*.json text
   snapshot/**/*.cfg text
   ```
   Unrelated rules are preserved, and required rules are appended only once.

`pkt init` refuses to modify a directory that is not a Git repository and is
idempotent: rerunning it leaves the same usable configuration.

### Import an existing project

```bash
pkt export my_project.pkt snapshot/my_project/
mkdir -p topologia && cp my_project.pkt topologia/
git add .
git commit -m "chore: import my_project into version control"
```

From that commit onward, the hooks maintain both representations.

## Migration from the Python Implementation

Existing repositories may still have `.githooks` that call `pktgit.py` and
`pktgit.script`, plus a `.venv` used only by the old tool. The snapshot layout
is unchanged; migrate in place:

```bash
# 1. Install and verify the Go binary (script installer or release artifact).
pkt --version

# 2. Rewrite the hooks, attributes, and pktgit.bin key.
pkt init

# 3. Optional: remove the old repository-local keys.
git config --unset pktgit.py 2>/dev/null || true
git config --unset pktgit.script 2>/dev/null || true

# 4. Delete the old tool environment if present (never commit it).
rm -rf .venv

# 5. Refresh derived files and rebuild the binary once.
pkt sync snapshot/<project>
pkt build snapshot/<project>/base.snapshot.xml topologia/<project>.pkt
```

After migration, `pkt init` can be rerun safely on any machine. Hooks are
repository-local, so other repositories keep their existing behavior until
they are migrated individually.

## Upgrading

Replace the binary on `PATH` (or rerun the installer). Hooks resolve
`pktgit.bin` on every run, so an upgrade takes effect immediately in every
wired repository. If the binary moves to a different absolute path, rerun
`pkt init` in each repository.

## Why the Hooks Are Safe Elsewhere

Activation is repository-local (`core.hooksPath` is per repository), and the
generated hooks act only on pkt-git paths:

- `pre-commit` acts only on staged `.pkt`/`.pka` files and snapshot XML.
- `post-merge` and `post-checkout` act only on changed
  `snapshot/*/base.snapshot.xml` paths.
- `pkt resolve` stages only paths involved in the conflict resolution.

No global Git configuration, aliases, or system files are touched. Other
repositories on the machine behave conventionally.

## Platform Notes

| Platform | Status | Notes |
|---|---|---|
| Linux | Fully tested (project of record) | The reference environment. |
| macOS | Supported by design | Build the binary or use a release artifact; no C toolchain is required. |
| Windows | Supported by design | Use `pkt.exe`; hooks run through the POSIX shell shipped with Git for Windows. |

The PT format itself is pinned to **9.0.1.0858** for the team. Newer Packet
Tracer versions may produce an unverified format variant; run `pkt verify` on a
specimen before relying on it.

## Rebuilding from Source

```bash
make test
make build
pkt --version
pkt verify testdata/specimens/Caso1_Grupo09-2.pkt
```

`make dist` produces cross-platform release binaries. The source tree is the
only build input; no generated snapshot or binary artifact is required.

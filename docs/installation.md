# Installation & Setup

Two installers implement the same steps once per machine and take under a
minute: **`install.sh`** for POSIX (Linux, macOS, and Windows via Git Bash) and
**`install.ps1`** for native PowerShell on Windows. Both are idempotent — safe
to re-run.

## The One-Command Model

pkt-git installs **globally, once per machine**, and the per-repository wiring
is a single command:

```bash
# 1) install globally (per machine; POSIX / Windows respectively)
./install.sh
# Windows: powershell -ExecutionPolicy Bypass -File install.ps1

# 2) in ANY repository (existing or brand new):
pkt init
```

`pkt init` performs, without any template copying:

1. Writes **thin hooks** into `.githooks/` — each is a single line that resolves
   two repository configs (`pktgit.py`, `pktgit.script`) and executes them:
   ```bash
   exec "$(git config --get pktgit.py)" "$(git config --get pktgit.script)" pre-commit
   ```
2. Writes `.gitattributes` if missing — `*.pkt binary` (the artifact never
   text-merges) plus text attributes for snapshot files.
3. Saves the absolute interpreter/CLI paths in `git config`.
4. Activates `git config core.hooksPath .githooks`.

Because the hook logic lives in the **global** CLI (`~/.local/share/pkt-git/`
on POSIX, `%LOCALAPPDATA%\pkt-git\` on Windows), repositories contain no tool
copies: one install, upgrades take effect everywhere. The `pkt` command is on
PATH for human use (`pkt export / build / inspect / edit / verify / sync`).

### Per-repository import (integrating an existing project)

```bash
cd my-existing-pt-project
pkt init
pkt export my_project.pkt snapshot/my_project/
mkdir -p topologia && cp my_project.pkt topologia/
git add . && git commit -m "chore: import existing Packet Tracer project"
```

## What Setup/installer Actually Do

The installers (`install.sh` / `install.ps1`) run the following once per
machine, on install and re-run:

1. **Interpreter detection.** Looks for a working Python 3.12+ in this order:
   `python3` → `python` → `py -3` (the `py -3` launcher covers Windows installs
   where no `python3` alias exists).
2. **Global install root.** Files live in `~/.local/share/pkt-git/` (POSIX) or
   `%LOCALAPPDATA%\pkt-git\` (Windows): the tool source (`tools/`) plus **one
   shared venv** — no per-repository venvs are ever created.
3. **Dependencies.** Installs `twofish` (block-cipher binding, validation
   reference) and `pycryptodome` into the global venv.
4. **Twofish patch.** The `twofish` package imports the removed-in-3.12 `imp`
   module to locate its C extension. The installer rewrites that loader to use
   `importlib.util.find_spec('_twofish')` instead — keeping whatever binary
   extension `pip` built (`.so` on Linux/macOS, `.pyd` on Windows), with a
   file-name fallback. The patch is idempotent (applied only if `import imp`
   is still present).
5. **CLI on PATH.** A `pkt` wrapper is placed in `~/.local/bin/` (POSIX) or
   `%USERPROFILE%\.local\bin\pkt.cmd` (`ptg.cmd` kept as legacy alias) added to the user PATH (Windows).
6. **Crypto self-test.** One Twofish encrypt/decrypt cycle with the project
   key. Repositories are then wired per-project with `pkt init` (see above).

> **Legacy repository-built mode:** the original `setup.sh` (creates
> `.venv/` inside the repository and self-hosts the hooks) still works and is
> kept for the development clone of pkt-git itself; the global install is the
> recommended path for every other repository.

## Per-Platform Table

| Platform | Command | Notes |
|---|---|---|
| Linux | `pkt` (global CLI) | Fully tested (project of record). |
| macOS | `pkt` (global CLI) | Needs Xcode Command Line Tools (`xcode-select --install`) so pip can build the Twofish C binding. |
| Windows | `pkt.cmd` (global CLI) | `install.ps1` adds it to the user PATH; thin hooks still run under Git Bash (ships with Git for Windows). |

Prerequisites per OS:

```bash
# Debian / Ubuntu
sudo apt update && sudo apt install -y python3 python3-venv git curl

# Arch
sudo pacman -S --needed python git curl

# Fedora
sudo dnf install -y python3 git curl

# macOS (Homebrew)
brew install python git
```

```powershell
# Windows (winget)
winget install -e --id Python.Python.3.13
winget install -e --id Git.Git
```

### First import of an existing project

```bash
pkt export "my_project.pkt" snapshot/my_project/
mkdir -p topologia && cp my_project.pkt topologia/    # if you keep it as an artifact
git add . && git commit -m "chore: import my_project into version control"
```

From this commit onward, the hooks maintain both representations
automatically.

## The Hooks and Why They Are Safe Elsewhere

Activation is **repository-local** (`core.hooksPath` is per-repo) and hooks are
thin files generated into each repository's `.githooks/`:

- `pre-commit` — acts only on staged `*.pkt`; any other commit (a README fix,
  a normal code project if you were to copy the pattern) passes untouched.
- `post-merge` / `post-checkout` — act only on changed
  `snapshot/*/base.snapshot.xml` paths.
- All actual logic runs in the global CLI; the hook just resolves
  `pktgit.py` / `pktgit.script` from the repository configuration.

No global Git configuration, aliases, or system files are touched — other
repositories on the machine behave conventionally.

## Known Windows Caveat (declared debt)

The tests executed during development ran on Linux with the Windows layouts
simulated. The first physical Windows validation will be done by the first
teammate on a Windows box. If `pip install twofish` fails there for lack of
MSVC Build Tools, the planned fix is already on the roadmap: replace the C
binding with the **in-house pure-Python Twofish** implementation, which removes
compilation from all three platforms.

## Re-installing from Scratch

Everything is reproducible from the repository:

```bash
git clone https://github.com/MarkOne-dev/pkt-git.git && cd pkt-git
./install.sh             # or install.ps1 on Windows
pkt --version
pkt verify proof-pkt/Caso1_Grupo09-2.pkt
```

The last command re-establishes confidence in the crypto pipeline against the
bundled reference specimen (expect `XML ROUND TRIP : True`).

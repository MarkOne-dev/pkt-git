# Installation & Setup

Two installers implement the same steps once per machine and take under a
minute: **`setup.sh`** for POSIX (Linux, macOS, and Windows via Git Bash) and
**`setup.ps1`** for native PowerShell on Windows. Both are idempotent — safe to
re-run.

## What Setup Does (step by step)

1. **Interpreter detection.** Looks for a working Python 3.12+ in this order:
   `python3` → `python` → `py -3` (the `py -3` launcher covers Windows installs
   where no `python3` alias exists).
2. **Virtual environment.** Creates `.venv/` once (skipped if it already
   exists). The venv layout is resolved per platform:
   - POSIX: `.venv/bin/` (`bin/python`, `bin/pip`)
   - Windows: `.venv/Scripts/` (`Scripts/python.exe`, `Scripts/pip.exe`)
3. **Dependencies.** Installs `twofish` (block-cipher binding, validation
   reference) and `pycryptodome` into the venv.
4. **Twofish patch.** The `twofish` package imports the removed-in-3.12 `imp`
   module to locate its C extension. Setup rewrites that loader to use
   `importlib.util.find_spec('_twofish')` instead — keeping whatever binary
   extension `pip` built (`.so` on Linux/macOS, `.pyd` on Windows), with a
   file-name fallback. The patch is idempotent (applied only if `import imp`
   is still present).
5. **Hooks activation.** `git config core.hooksPath .githooks` — the
   repository-versioned hooks become live. (Skipped with a warning if the
   directory is not a Git repo yet.)
6. **Crypto self-test.** Runs one Twofish encrypt/decrypt cycle with the
   project's key and asserts the round trip, printing `crypto core verified ✓`.

## Platform Matrix

| Platform | Install command | Notes |
|---|---|---|
| Linux | `./setup.sh` | Fully tested (project of record). |
| macOS | `./setup.sh` | Needs Xcode Command Line Tools (`xcode-select --install`) so pip can build the Twofish C binding. |
| Windows (native) | `powershell -ExecutionPolicy Bypass -File setup.ps1` | Primary Windows path. |
| Windows (Git Bash) | `bash setup.sh` | Works too; hooks run under Git Bash either way. |

### Prerequisites per OS

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
.venv/bin/python tools/ptg.py export "my_project.pkt" snapshot/my_project/
mkdir -p topologia && cp my_project.pkt topologia/    # if you keep it as an artifact
git add . && git commit -m "chore: import my_project into version control"
```

From this commit onward, the hooks maintain both representations
automatically.

## The Hooks and Why They Are Safe Elsewhere

Activation is **repository-local** (`core.hooksPath` is per-repo) and hooks are
tracked in `.githooks/`:

- `pre-commit` — acts only on staged `*.pkt`; any other commit (a README fix,
  a normal code project if you were to copy the pattern) passes untouched.
- `post-merge` / `post-checkout` — act only on changed
  `snapshot/*/base.snapshot.xml` paths.
- `.githooks/common.sh` — solves the one genuine cross-platform trap:
  `pre-commit`/`post-merge`/`post-checkout` call `resolve_py` to pick
  `.venv/bin/python` (POSIX) or `.venv/Scripts/python.exe` (Windows).

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
git clone https://github.com/<your-user>/pkt-git.git && cd pkt-git
./setup.sh              # or setup.ps1 on Windows
.venv/bin/python tools/spike_roundtrip.py verify proof-pkt/Caso1_Grupo09-2.pkt
```

The last command re-establishes confidence in the crypto pipeline against the
bundled reference specimen (expect `XML ROUND TRIP : True`).

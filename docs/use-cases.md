# Use Cases & Tested Flows

Concrete ways to use pkt-git. Every flow marked **[tested]** was executed
end-to-end against a real Packet Tracer project during development (including
the public demo repository `github.com/MarkOne-dev/PRUEBA-PKT`).

## Flow 1 — Machine Bootstrap & Environment Verification (once per computer)  [tested]

Install the global `pkt` binary and verify your environment:

```bash
git clone https://github.com/MarkOne-dev/pkt-git.git && cd pkt-git
./install.sh            # Windows: powershell -ExecutionPolicy Bypass -File install.ps1
pkt --version           # global CLI alive
pkt doctor              # inspect Git identity, Packet Tracer detection, and PATH
```

Nothing else per machine — the tool lives in `~/.local/share/pkt-git`
(POSIX) / `%LOCALAPPDATA%\pkt-git` (Windows), with `pkt` on PATH.

## Flow 2 — Import an Existing Project (Project Initializer)  [tested]

You already have a `.pkt` worked locally and want it versioned from now on:

```bash
cd my-project && git init -b main
# Place your topology file in the project (e.g., topologia/proyecto.pkt or proyecto.pkt)
pkt init                                            # auto-detects .pkt, wires hooks, exports snapshot
git add . && git commit -m "chore: import PT project"
git remote add origin git@github.com:<you>/my-project.git
git push -u origin main
```

`pkt init` automatically discovers any `.pkt` or `.pka` files in the repository and generates the initial `snapshot/<project>/` (XML, device `.cfg` files, and `topology.json`). Each device config becomes reviewable on GitHub from the very first commit.

## Flow 3 — Clone an Existing Project (Teammate / Collaborator Onboarding)  [tested]

When a teammate joins an existing repository initialized by another collaborator:

```bash
# 1. Clone the repository
git clone git@github.com:<org>/my-project.git && cd my-project

# 2. Wire the hooks locally (MANDATORY ONCE PER CLONE)
pkt init
```

> [!IMPORTANT]
> **Why `pkt init` is required after cloning:**  
> For security reasons, Git **never** activates repository hooks automatically upon `git clone`. Running `pkt init` once connects `.githooks/` (`core.hooksPath = .githooks`) and registers the absolute path to your local `pkt` binary (`pktgit.bin`), ensuring hooks execute reliably across differing operating systems and user paths.

### End-to-End Collaboration Cycle:

1. **Pulling latest updates:**
   ```bash
   git pull
   ```
   Git downloads the incoming `snapshot/` text changes. The **`post-merge` hook** runs automatically, rebuilds the local `.pkt` binary, and reports the updated topology. When you open Cisco Packet Tracer, the network is already updated.

2. **Making edits in Cisco Packet Tracer:**
   - Open the `.pkt` in Cisco Packet Tracer, make topology and device changes, and save with `Ctrl+S`.

3. **Committing and pushing changes:**
   ```bash
   git add mi-red.pkt
   git commit -m "feat: add distribution switch and vlan 30"
   git push
   ```
   Before the commit finishes, the **`pre-commit` hook** automatically exports the modified binary into `snapshot/mi-red/` (refreshing `base.snapshot.xml`, per-device configs in `configs/`, and `metadata/topology.json`) and stages them alongside the commit. When pushed, teammates simply run `git pull` to receive the changes and have their local `.pkt` rebuilt automatically.

## Flow 4 — Daily Save (the 3-command loop)  [tested]

The whole "user experience" for a person with Packet Tracer open:

```bash
# in Packet Tracer: edit the network → Ctrl+S
git add .
git commit -m "feat: add VLAN 60 to SW_ACC_ADMIN_PUNO"   # hook exports the snapshot
git push
```

**[tested]** live: an interface change (`clock rate 2000000` on ISP2) plus new
devices appeared in the commit both as binary diff *and* as plain-text config
diffs humans can read (`+ clock rate 2000000`).

## Flow 5 — Receive Work (pull / merge)  [tested]

```bash
git pull          # or git merge / git switch <branch>
# → post-merge / post-checkout hook: snapshot changed → binary rebuilt
open Packet Tracer → the network is already updated
```

**[tested]** both hooks live (verified by decoding the rebuilt binary and
confirming the new VLAN inside it).

## Flow 6 — Pull Request Review (GitHub)  [tested]

The PR shows the *network* instead of opaque binaries:

```diff
 --- a/snapshot/.../config/SW_ACC_ADMIN_PUNO.cfg
+++ b/snapshot/.../config/SW_ACC_ADMIN_PUNO.cfg
 interface FastEthernet0/3
- switchport access vlan 50
+ switchport access vlan 60
```

Reviewers read line-by-line changes (IPs, VLANs, ACLs, clock rates) without
opening Packet Tracer. Merging the PR is a normal GitHub merge; teammates get
the merged binary after their next `git pull`.

## Flow 7 — Conflict Resolution (one thought + one command)  [tested]

Two branches changed the same config line (VLAN 55 vs 77):

```bash
git merge feature/x                    # conflict in snapshot texts (expected)
pkt resolve                            # report: exactly what conflicts, where
pkt resolve --theirs                   # choose the incoming version
# (or fix the <<<<<<< markers manually, or pick --ours)
git commit -m "merge: resolve vlan conflict"    # hook re-exports; done
```

**[tested]** live with two divergent branches editing the same `SW1_LIMA`
interface. `pkt resolve` takes the chosen policy, regenerates `configs/*.cfg`
and `metadata/topology.json`, rebuilds the conflicted binary and stages
everything — the human only picked the winner.

Honestly, fully automatic resolution is impossible when both branches edit the
same line with different values: something meaningful is lost no matter what
an algorithm picks. `pkt resolve` automates everything *around* that decision.

## Flow 8 — Gitflow (main / develop / features)  [tested]

The hooks are branch-agnostic; gitflow works unchanged:

```bash
git switch develop
git switch -c feature/vlan-60
# ...edit PT, save, commit (hook exports)... push
# PR → develop, review on GitHub, merge, release main ← develop
```

The local binary always represents **the checked-out branch**: switching
branch rebuilds the binary to that branch's snapshot (post-checkout), pulling
on develop rebuilds to develop (post-merge). Recommended GitHub setup:
protect `main` (and optionally `develop`) to require PRs.

## Flow 9 — Reviewers & CI Without Packet Tracer

A teammate whose contribution is review/audit only does not need Packet Tracer installed:

```bash
git clone ... && pkt init      # hooks wire fine; PT never needed
git pull                        # text arrives; binary rebuild optional
# inspect any device from the terminal:
pkt inspect topologia/proyecto.pkt
pkt edit snapshot/proyecto/base.snapshot.xml show SW1_LIMA
```

## Flow 10 — Multi-Project Repositories & Smart Batch Export  [tested]

Multiple independent PT projects in one repo; the hooks and export command handle **each** one
separately (`snapshot/<name>/` ↔ `topologia/<name>.pkt`, name-matched):

```bash
# Export all topologies across the workspace at once:
pkt export --all

# Or export a single topology by convention:
pkt export lab_router.pkt          # destination defaults to snapshot/lab_router/
```

**[tested]** — `pkt` commands and hooks are per-file; no project limit.

## Flow 11 — Interface Language & Localization  [tested]

Configure the CLI language globally across your machine or per repository:

```bash
# Set preferred language globally to Spanish (~/.gitconfig)
pkt lang es

# Set preferred language globally to English (~/.gitconfig)
pkt lang en

# Interactive selection via Lipgloss terminal menu (TTY)
pkt language

# Set language strictly for the current repository (.git/config)
pkt lang es --local

# Temporary one-off execution in a specific language
pkt doctor --lang es
```

The preferred language is stored in Git configuration (`pkt.language`), guaranteeing cross-platform persistence across Linux, macOS, and Windows without external state files.

## Flow 12 — Terminal-Only Network Surgery

Change the network *without opening PT* (then rebuild and commit):

```bash
pkt edit snapshot/proyecto/base.snapshot.xml chgline SW1_LIMA FastEthernet0/1 \
    "switchport access vlan 10" "switchport access vlan 20"
pkt build snapshot/proyecto/base.snapshot.xml topologia/proyecto.pkt
git add . && git commit -m "feat: move vlan 20 to SW1_LIMA"   # hook re-exports
```

Useful for scripted/batch changes (rename a VLAN across devices, standardize
descriptions). Every edit is device-scoped and XML-validated; derived
`configs/*.cfg` are re-synced automatically.

## Anti-Patterns (do not combine with pkt-git)

- Two people editing the *same device* at the same time in PT and pushing
  simultaneously — one line will conflict (unavoidable; resolve with Flow 7).
- Resolving conflicts with global `sed`/find-replace over the 5.9 MB XML —
  past real incident (duplicated VLAN assignment); always device-scoped.
- Keeping a `.pkt` outside the repository and copying it by hand — the
  repository's binary is the single canonical artifact; `pkt init` flows cover
  importing external ones exactly once.
- Forgetting to run `pkt init` after cloning a repository on a new machine.

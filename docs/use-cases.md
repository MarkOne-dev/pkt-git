# Use Cases & Tested Flows

Concrete ways to use pkt-git. Every flow marked **[tested]** was executed
end-to-end against a real Packet Tracer project during development (including
the public demo repository `github.com/MarkOne-dev/PRUEBA-PKT`).

## Flow 1 — Machine Bootstrap (once per computer)  [tested]

```bash
git clone https://github.com/MarkOne-dev/pkt-git.git && cd pkt-git
./install.sh            # Windows: powershell -ExecutionPolicy Bypass -File install.ps1
pkt --version           # global CLI alive
```

Nothing else per machine — the tool lives in `~/.local/share/pkt-git`
(POSIX) / `%LOCALAPPDATA%\pkt-git` (Windows), with `pkt` on PATH.

## Flow 2 — Import an Existing Project (once per repository)  [tested]

You already have a `.pkt` worked locally; you want it versioned from now on:

```bash
cd my-project && git init -b main
pkt init                                            # thin hooks + ownership rules
mkdir -p topologia && mv proyecto.pkt topologia/
pkt export topologia/proyecto.pkt snapshot/proyecto/
git add . && git commit -m "chore: import PT project"
git remote add origin git@github.com:<you>/my-project.git
git push -u origin main
```

**[tested]** with a 599 KB / 107-device project, and live with the 368 KB /
59-device `PRUEBA-PKT` demo (each device config appears as a reviewable file
on GitHub from the very first commit).

## Flow 3 — Daily Save (the 3-command loop)  [tested]

The whole "user experience" for a person with PT open:

```bash
# in Packet Tracer: edit the network → Ctrl+S
git add .
git commit -m "feat: add VLAN 60 to SW_ACC_ADMIN_PUNO"   # hook exports the snapshot
git push
```

**[tested]** live: an interface change (`clock rate 2000000` on ISP2) plus new
devices appeared in the commit both as binary diff *and* as plain-text config
diffs humans can read (`+ clock rate 2000000`).

## Flow 4 — Receive Work (pull / merge)  [tested]

```bash
git pull          # or git merge / git switch <branch>
# → post-merge / post-checkout hook: snapshot changed → binary rebuilt
open Packet Tracer → the network is already updated
```

**[tested]** both hooks live (verified by decoding the rebuilt binary and
confirming the new VLAN inside it).

## Flow 5 — Pull Request Review (GitHub)  [tested]

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

## Flow 6 — Conflict Resolution (one thought + one command)  [tested]

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

## Flow 7 — Gitflow (main / develop / features)  [tested]

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

## Flow 8 — Reviewers Without Packet Tracer

A teammate whose contribution is review/audit only does not need PT installed:

```bash
git clone ... && pkt init      # hooks wire fine; PT never needed
git pull                        # text arrives; binary rebuild optional
# inspect any device from the terminal:
pkt inspect topologia/proyecto.pkt
pkt edit snapshot/proyecto/base.snapshot.xml show SW1_LIMA
```

## Flow 9 — Multi-Project Repository  [tested]

Multiple independent PT projects in one repo; the hooks handle **each** one
separately (`snapshot/<name>/` ↔ `topologia/<name>.pkt`, name-matched):

```bash
pkt export lab_router.pkt snapshot/lab_router/
pkt export lab_switch.pkt snapshot/lab_switch/
cp lab_router.pkt topologia/ && cp lab_switch.pkt topologia/
git add . && git commit -m "feat: import two lab projects"
```

**[tested]** — `pkt` commands and hooks are per-file; no project limit.

## Flow 10 — Terminal-Only Network Surgery

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
  simultaneously — one line will conflict (unavoidable; resolve with Flow 6).
- Resolving conflicts with global `sed`/find-replace over the 5.9 MB XML —
  past real incident (duplicated VLAN assignment); always device-scoped.
- Keeping a `.pkt` outside the repository and copying it by hand — the
  repository's binary is the single canonical artifact; `pkt init` flows cover
  importing external ones exactly once.

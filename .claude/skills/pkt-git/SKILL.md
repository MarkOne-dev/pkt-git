---
name: pkt-git
description: Version Cisco Packet Tracer (.pkt) projects with Git/GitHub using the pkt-git tool. Use this skill whenever the user wants to import a .pkt project into a repository, commit/pull PRs around Packet Tracer topologies, resolve merge conflicts in network snapshots, or explain/review network changes from textual diffs. Covers installation, `pkt init`, daily flows, conflicts (`pkt resolve`), and the invariants (binary = artifact, snapshot = source of truth).
---

# pkt-git

pkt-git converts Cisco Packet Tracer binaries (`.pkt`) into Git-friendly text
and keeps both in sync via thin Git hooks. Rules first, then how to act.

## Invariants (never violate)

1. **The text is the source of truth; the binary is an artifact.**
   `snapshot/<Project>/` (XML + per-device `.cfg` + `metadata/topology.json`)
   is authoritative; `topologia/<Project>.pkt` is regenerated from it and is
   never text-merged (`.gitattributes` marks it binary).
2. **Never global find-replace or `sed` over snapshot XML** (multi-MB; caused a
   real duplicated-VLAN incident). Edits are device-scoped via
   `pkt edit <xml> <device> ...` only.
3. Every commit containing a `.pkt` must contain its fresh snapshot — the
   pre-commit hook normally does this automatically; never bypass hooks with
   `--no-verify` to "fix" binary/text mismatches.
4. Fix conflicts one of two ways only: `pkt resolve --theirs|--ours`, or
   device-scoped manual edits, then `pkt build` + `git add`.

## Detecting the environment

Check availability before promising anything:

```bash
pkt --version                # global CLI (installed by pkt-git's install.sh)
git rev-parse --git-dir      # is this a repository?
pktgit=$(git config --get pktgit.script)   # hooks wired via `pkt init`?
```

- `pkt` not found → the machine lacks the global install; instruct the user to
  run `./install.sh` (or `install.ps1` on Windows) from a pkt-git clone.
  Do not fabricate a local install.
- Repository has no `pktgit.py` config → run `pkt init` in its root (requires
  a git repository; a fresh repo needs `git init` first).

## Command surface

```bash
pkt init [dir]                                  # wire hooks + .gitattributes into a repo
pkt export <project.pkt> <output_dir>           # binary → textual snapshot (bootstrapping)
pkt build  <snapshot.xml> <output.pkt>          # snapshot → valid binary artifact
pkt inspect <project.pkt>                       # devices/links summary
pkt edit  <xml> show <DEVICE>                   # read a device config (safe, read-only)
pkt edit  <xml> chgline <DEV> <IFACE> "<old>" "<new>"   # surgical single-line change
pkt edit  <xml> setiface <DEV> <IFACE> "line"...        # replaces the WHOLE interface block (blunt)
pkt edit  <xml> insertafter <DEV> <phrase> "line"...   # insert after a matching line
pkt sync  <snapshot-dir>                        # refresh configs/*.cfg + metadata from the XML
pkt resolve [--theirs|--ours]                   # merge-conflict resolution
pkt verify <file.pkt>                           # crypto round-trip verification
```

## Recipes (what the user probably wants)

### "I have a .pkt project, put it on GitHub"

```bash
cd <project-dir> && git init -b main
pkt init
mkdir -p topologia && mv proyecto.pkt topologia/
pkt export topologia/proyecto.pkt snapshot/proyecto/
git add . && git commit -m "chore: import PT project"
git remote add origin <user-provided-url> && git push -u origin main
```

### "I edited the network in PT" (after their Ctrl+S)

```bash
git add . && git commit -m "feat: <what changed>"
git push
```

Expect the pre-commit hook output to show the auto-export. Then to describe
what changed, read the text diffs:
`git diff HEAD~1 -- 'snapshot/**/configs/*.cfg' 'snapshot/**/metadata/topology.json'`.

### "Pull my teammate's changes"

```bash
git pull
```

The post-merge hook rebuilds binaries. Report rebuilt projects from its
output; do not run `pkt build` manually unless the hook is missing.

### "There's a merge conflict in the snapshot"

```bash
pkt resolve                     # report mode: see exactly what conflicts
pkt resolve --theirs | --ours   # after deciding (ask the user WHO wins if unclear!)
```

Never auto-pick a side in a real conflict without asking the user — two
branches editing the same line means someone's work will be dropped.

### "Change the network without opening PT"

```bash
pkt edit snapshot/<p>/base.snapshot.xml chgline <DEV> <IFACE> "<old>" "<new>"
pkt build snapshot/<p>/base.snapshot.xml topologia/<p>.pkt
git add . && git commit -m "feat: <what changed>"
```

## Verification defaults

After any binary-affecting operation, prefer `pkt verify <specimen.pkt>`
(expects `XML ROUND TRIP : True`). `byte-identical: False` in that report is
expected and fine (PT's own zlib emitter differs).

## Where the details live

Machine-readable docs ship with the repository under `docs/`:
`use-cases.md` (10 validated flows), `workflow.md` (SOP), `cli-reference.md`,
`packet-tracer-format.md` (4-layer crypto pipeline), `troubleshooting.md`.
Read them before unusual operations (older PT versions, `.pka` files — both
untested; verify with `pkt verify` on a specimen first and say so).

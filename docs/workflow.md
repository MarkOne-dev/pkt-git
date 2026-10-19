# Collaborative Workflow

This document is the team SOP. It covers the day-to-day loop, what the hooks do
automatically, branching and Pull Requests, and the explicit rules for conflict
resolution.

## The One-Command Model (v1.0)

Install the tool **globally, once per machine**, then wire any repository with
a single command — no templates to copy:

```bash
# any repo, new or existing:
ptg init
```

`ptg init` generates thin hooks that call the global CLI (they store the exact
interpreter and script paths in the repository configuration), writes the
`*.pkt binary` ownership rules into `.gitattributes` (if missing) and activates
everything. See `installation.md` for the full model.

To re-import a project that already exists on disk:

```bash
ptg export my_project.pkt snapshot/my_project/     # textual snapshot
mkdir -p topologia && cp my_project.pkt topologia/ # binary artifact
git add . && git commit -m "chore: import PT project"
```

## What `ptg init` Actually Does

Instead of copying a template, the hooks are **thin**: three lines each. They
resolve two repository-local configurations — `pktgit.py` (the interpreter) and
`pktgit.script` (the installed CLI) — and hand control to the global
installation. This means:

- Repositories carry no copies of the tool — one install per machine,
  upgrades propagate to every project.
- The daily work is identical to the workflow described below.


## The Daily Loop (zero manual conversion)

### Saving your work

```bash
# 1. Edit the network in Packet Tracer, then Ctrl+S
# 2. From the repository:
git add .
git commit -m "feat: add VLAN 60 to SW_ACC_ADMIN_PUNO"
git push
```

The **pre-commit** hook guarantees that every commit containing a `.pkt` also
contains its freshly generated textual snapshot:

1. Detects staged `.pkt` files (added/copied/modified/renamed).
2. Runs `ptg export <file> snapshot/<name>/` for each one.
3. Stages the regenerated XML, `.cfg` files, and `topology.json`.
4. As a bonus guard, warns if the opposite case appears (snapshot changed
   without its `.pkt`) — valid after textual edits, with a reminder to rebuild.

If the export ever fails, the hook aborts the commit with the tool's error
output — a broken commit cannot be created.

### Receiving work

```bash
git pull        # or git merge, or git checkout of a branch
```

The **post-merge** (and its twin **post-checkout**) hook:

1. Computes the incoming commit range (`ORIG_HEAD..HEAD` for merges, the old→new
   range for checkouts).
2. Lists `snapshot/*/base.snapshot.xml` files that changed.
3. Runs `ptg build` for each, regenerating `topologia/<name>.pkt`.
4. Reports success — open the `.pkt` in Packet Tracer and the network is there.

The two directions are mirror images, fully automatic:

| Event | Hook | Action |
|---|---|---|
| `.pkt` staged | pre-commit | export → snapshot staged with it |
| snapshot arrives via merge/pull | post-merge | build → local `.pkt` rebuilt |
| branch switch changes snapshot | post-checkout | build → local `.pkt` rebuilt |

**If you only ever use this conventional workflow, you never run `export` or
`build` yourself. They are hook automation.**

## Branching and Pull Requests

Example from the project's own history:

```bash
git switch -c feature/vlans-mario     # Mario adds VLANs
git switch -c feature/routing-vanessa # Vanessa adds OSPF
```

Both branches can develop in parallel because the textual snapshots merge
semantically: OSPF config lines and VLAN assignment lines touch different
devices/files, so Git merges them cleanly. The binary `.pkt` files are treated
as binary and never auto-merged — text decides.

On GitHub:

1. Push the feature branch and open a Pull Request.
2. Reviewers see **the network as text diffs** — every IP, VLAN, and ACL line
   with `+/-` markers, reviewable without opening Packet Tracer.
3. After merge, teammates just `git pull`: the hook rebuilds their local
   binaries.

## Conflict Resolution (the one manual case)

When two branches edit the same config line (real example: VLAN 50 vs VLAN 60
on `SW_ACC_ADMIN_PUNO` `FastEthernet0/3`), Git reports a conflict on the
`.cfg`/XML text — exactly like code. The resolver:

1. Edits the conflicted snapshot **by hand or with device-scoped tools**:
   ```bash
   ptg edit snapshot/<project>/base.snapshot.xml \
       chgline SW_ACC_ADMIN_PUNO FastEthernet0/3 "switchport access vlan 50" \
       "switchport access vlan 60"
   ```

2. `git add` the snapshot.
3. Rebuild the binary — **the only manual conversion command in the whole
   workflow**:
   ```bash
   ptg build snapshot/<project>/base.snapshot.xml \
       topologia/<project>.pkt
   git add topologia/<project>.pkt
   ```
4. `git commit` to finish the merge commit.

The pre-commit hook prints a warning whenever a snapshot was changed without
its `.pkt`, so this step is hard to forget even if you don't read this doc.

### Rules for editing snapshots

- **Never** fix conflicts with global tools (`sed` over the whole XML, any
  find-replace across the 5.9 MB document). Editions must be scoped to one
  device — use `edit_config.py` (its `show/setiface/chgline/insertafter`
  commands all take the device name first). A past incident where a global
  replace duplicated a VLAN assignment motivated this rule.
- Be explicit about which config block you mean: `RUNNINGCONFIG` and
  `STARTUPCONFIG` are separate; `edit_config.py` operates on the running one.
- `setiface` replaces the **entire interface block** — convenient but blunt;
  use `chgline` for surgical single-line changes and keep the surrounding
  lines (`description`, `portfast`, …) that you did not intend to touch.

## Small-Team Etiquette

- One logical change per commit (one device config, one VLAN, one subset of
  interfaces) — the diff is the review, keep it readable.
- Pull/merge before starting PT work; PT holds no merge intelligence of its own.
- Never edit `snapshot/` and `topologia/*.pkt` in PT at the same time as making
  textual edits — pick one source of change per commit session.
- The PT version is pinned team-wide (**9.0.1.0858**): opening and resaving with
  a newer PT may produce a format variant the codec has not been verified
  against.

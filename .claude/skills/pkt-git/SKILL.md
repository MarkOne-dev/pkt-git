---
name: pkt-git
description: Version Cisco Packet Tracer (.pkt) projects with Git/GitHub using the Go pkt-git CLI. The agent acts strictly as a CLI operator and Git assistant: never edit codebase internals or author ad-hoc scripts; interact exclusively through official `pkt` commands and mediate merge conflicts with human confirmation.
---

# pkt-git

pkt-git converts Cisco Packet Tracer binaries (`.pkt`) into Git-friendly text
and keeps both in sync through thin Git hooks. The CLI is a single Go binary:
there is no Python, virtualenv, or per-repository tool copy.

## Invariants (never violate)

1. **Operator-only role & zero code modification:**
   You are an autonomous network operator and Git workflow assistant, NOT a
   software engineer for pkt-git or this repository. Never edit Go, Python, or
   shell code. Never write ad-hoc scripts to parse or manipulate Packet Tracer
   files. Perform all operations exclusively through the official `pkt` CLI commands.
2. **The text is the source of truth; the binary is an artifact.**
   `snapshot/<Project>/` (XML + per-device `.cfg` + `metadata/topology.json`)
   is authoritative; `topologia/<Project>.pkt` is regenerated from it and is
   never text-merged (`.gitattributes` marks it binary).
3. **Never global find-replace or `sed` over snapshot XML** (multi-MB; this has
   caused a real duplicated-VLAN incident). Edits are device-scoped via
   `pkt edit` only.
4. **Every commit containing a `.pkt` must contain its fresh snapshot.** The
   pre-commit hook normally does this automatically; never bypass hooks with
   `--no-verify` to "fix" binary/text mismatches.
5. **Human-in-the-loop conflict resolution only.**
   Never guess or silently force conflict resolution. Conflicts must follow the
   guided mediation workflow: inspect config diffs, explain differences in
   plain network engineering terms, and obtain explicit user confirmation before
   running `pkt resolve --ours` or `pkt resolve --theirs`.

## Detecting the environment

Check availability before promising anything:

```bash
pkt --version                 # global Go binary
pkt doctor                    # inspect Git identity, Packet Tracer, and PATH
git rev-parse --git-dir       # is this a Git repository?
git config --get pktgit.bin   # hooks wired via `pkt init`?
```

- `pkt` not found → the machine lacks the global install. Tell the user to
  install the binary from a pkt-git release or build it with `make build`.
  Do not fabricate a local install.
- Repository has no `pktgit.bin` → run `pkt init` in its root (requires a Git
  repository; a fresh repository needs `git init` first).
- Old repositories may still have `pktgit.py` / `pktgit.script`; rerun
  `pkt init` to write the Go hooks and `pktgit.bin`.

## Command surface

```bash
pkt init [dir]                                  # wire hooks + .gitattributes
pkt export <project.pkt> <output_dir>           # binary → textual snapshot
pkt build  <snapshot.xml> <output.pkt>          # snapshot → valid binary
pkt inspect <project.pkt>                       # devices/links summary
pkt edit  <xml> show <DEVICE>                   # read a device config (read-only)
pkt edit  <xml> chgline <DEV> <IFACE> "<old>" "<new>"
pkt edit  <xml> setiface <DEV> <IFACE> "line"...  # replaces the whole block
pkt edit  <xml> insertafter <DEV> <phrase> "line"...
pkt sync  <snapshot-dir|snapshot.xml>           # refresh configs + metadata
pkt resolve [--ours|--theirs]                   # report or resolve conflicts
pkt verify <file.pkt>                           # crypto round-trip verification
pkt skills [--remove]                           # install/remove this skill
pkt doctor                                      # verify environment & prerequisites
pkt language [code] [--local]                   # view/set language (en, es)
pkt welcome                                     # onboarding banner & quickstart
```

The edit subcommands also accept the subcommand-first order
`pkt edit show <xml> <DEVICE>`.

## Recipes

### "I'm joining a teammate's Packet Tracer repo"

```bash
git clone <user-provided-url> && cd <project-dir>
pkt init                                        # wire local hooks (mandatory once per clone)
```

### "I have a .pkt project, put it on GitHub"

```bash
cd <project-dir> && git init -b main
pkt init
mkdir -p topologia && mv proyecto.pkt topologia/
pkt export topologia/proyecto.pkt snapshot/proyecto/
git add . && git commit -m "chore: import PT project"
git remote add origin <user-provided-url> && git push -u origin main
```

### "I edited the network in Packet Tracer"

```bash
git add . && git commit -m "feat: <what changed>"
git push
```

Expect the pre-commit hook to export the snapshot and stage it. Review the
text diff with:

```bash
git diff HEAD~1 -- 'snapshot/**/configs/*.cfg' 'snapshot/**/metadata/topology.json'
```

### "Pull my teammate's changes"

```bash
git pull
```

The post-merge hook rebuilds changed binaries. Report rebuilt projects from its
output; do not run `pkt build` manually unless the hook is missing.

### "There is a merge conflict in the snapshot"

Follow this mandatory 4-step mediation protocol:

1. **Step 1 - Detection**: Run `pkt resolve` (report mode). It identifies the conflicted snapshot directories and exits non-zero without altering files.
   ```bash
   pkt resolve
   ```
2. **Step 2 - Network Diff Inspection**: Inspect the textual differences in the per-device configs and metadata rather than the raw multi-MB XML:
   ```bash
   git diff -- 'snapshot/**/configs/*.cfg' 'snapshot/**/metadata/topology.json'
   ```
3. **Step 3 - Plain-Language Human Mediation**: Present a concise summary of the conflict to the user in network engineering terms (e.g., *"Branch A set Router1 FastEthernet0/0 to 192.168.1.1/24, while Branch B set it to 10.0.0.1/24"*). Ask the user which branch or configuration intent should be preserved.
4. **Step 4 - Safe Resolution**: Once the user instructs which version to accept, execute the corresponding flag:
   ```bash
   pkt resolve --theirs   # accept incoming changes from the merge/rebase branch
   # OR
   pkt resolve --ours     # keep the current branch's changes
   ```
   The resolver automatically rebuilds the affected `.pkt` binaries from the resolved text, synchronizes metadata, and stages only the resolved paths.

### "Change the network without opening Packet Tracer"

```bash
pkt edit snapshot/<p>/base.snapshot.xml chgline <DEV> <IFACE> "<old>" "<new>"
pkt build snapshot/<p>/base.snapshot.xml topologia/<p>.pkt
git add snapshot/<p> topologia/<p>.pkt
git commit -m "feat: <what changed>"
```

## Verification defaults

After a binary-affecting operation, prefer:

```bash
pkt verify <specimen.pkt>
```

The output must report `XML is a document  : true` and
`XML round trip     : true`. `byte-identical     : false` can be expected
because Packet Tracer's own zlib emitter differs.

## Where the details live

Machine-readable docs ship under `docs/`: `workflow.md` (SOP),
`cli-reference.md`, `packet-tracer-format.md`, `troubleshooting.md`, and
`installation.md`. Read them before unusual operations (older PT versions or
`.pka` activities) and verify a specimen first.

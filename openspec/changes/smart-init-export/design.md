# Design

## Context

`pkt init` currently focuses purely on installing repository-local Git hooks and attributes. If an engineer already has an existing `.pkt` topology, they must discover and execute `pkt export <file.pkt> snapshot/<name>/` separately. Furthermore, `pkt export` requires two positional arguments (`exactArgs(2)`), adding unnecessary friction for the default convention (`snapshot/<name>`).

## Goals / Non-Goals

**Goals:**
- Enable `pkt init [dir]` to scan for existing `.pkt` / `.pka` files and automatically export them to `snapshot/<name>/` on initialization.
- Make the destination argument in `pkt export <file.pkt> [destination]` optional, defaulting to `snapshot/<file_stem>/`.
- Allow `pkt export` with zero arguments to auto-detect a single repository topology.
- Keep full backward compatibility for invocations that pass explicit destination directories.

**Non-Goals:**
- Auto-staging or committing in Git during `pkt init`: the user retains explicit control over `git add` and `git commit`.
- Relocating or altering original `.pkt` files on disk.

## Decisions

### Decision 1: Shared repository topology discovery helper
Implement a helper `FindTopologies(dir string) ([]string, error)` that searches for `.pkt` and `.pka` files while skipping `.git/`, `.githooks/`, `snapshot/`, `bin/`, and `dist/`. This helper will be shared by `init` and `export`.

### Decision 2: Convention-based default snapshot path
Provide a helper `DefaultSnapshotDir(pktPath string) string` that extracts the file stem (e.g. `Caso1.pkt` -> `Caso1`) and resolves to `filepath.Join("snapshot", stem)`.

### Decision 3: Flexible argument handling in `pkt export`
Update `newExportCmd` to accept 0, 1, or 2 arguments:
- **0 arguments**: auto-detect topology via `FindTopologies`. If exactly 1 is found, export to default snapshot directory. If none or multiple are found, return actionable guidance.
- **1 argument**: export specified binary to `DefaultSnapshotDir(arg)`.
- **2 arguments**: export specified binary to explicitly given destination (unchanged behavior).

## Risks / Trade-offs

- **[Risk] Multiple topologies in a repository during `pkt init`** → *Mitigation*: `pkt init` iterates over and exports all discovered topologies, reporting each one clearly in the UI output.
- **[Risk] Accidental re-export overwriting snapshot on re-running `pkt init`** → *Mitigation*: Snapshot export decodes the `.pkt` and refreshes derived files cleanly without corrupting the working tree.

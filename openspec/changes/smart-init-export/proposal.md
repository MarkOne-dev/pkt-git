# Proposal

## Why

Onboarding an existing Packet Tracer project into `pkt-git` currently requires engineers to run `pkt init` and then manually remember and execute `pkt export <path.pkt> snapshot/<name>/`. Enhancing `pkt init` to automatically discover and export existing `.pkt` files, while making destination paths in `pkt export` optional by convention, reduces the onboarding flow to a single frictionless command.

## What Changes

- Enhance `pkt init [dir]` to scan for existing `.pkt` or `.pka` binaries within the target repository (including root and `topologia/`). When found, `pkt init` automatically exports initial snapshots to `snapshot/<project>/` and reports the exported topologies.
- Make the `<output_dir>` argument optional in `pkt export <project.pkt> [output_dir]`. When omitted, it defaults by convention to `snapshot/<project_basename>/`.
- Allow `pkt export` with zero arguments to auto-detect a single `.pkt` / `.pka` file in the repository or support an `--all` flag when multiple topologies are present.
- Maintain full backward compatibility for existing scripts passing explicit arguments (e.g., `pkt export input.pkt custom/path/`).

## Capabilities

### New Capabilities

*(None)*

### Modified Capabilities

- `git-integration`: `pkt init` SHALL auto-detect existing topology binaries in the repository and automatically export initial snapshots into `snapshot/<project>/`.
- `snapshot`: `pkt export` SHALL support default conventional destination directories (`snapshot/<project>/`) when the destination argument is omitted.

## Impact

- **CLI Layer**: `internal/cli/init.go` and `internal/cli/export.go`.
- **Hooks & Snapshot**: `internal/hooks/init.go` (discovery and initial export orchestration) and `internal/snapshot/` (default destination naming helper).
- **Documentation**: Update `docs/workflow.md`, `docs/cli-reference.md`, and `README.md` to highlight the one-command onboarding flow.

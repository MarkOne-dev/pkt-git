# Proposal: Agent Skill Guardrails and Guided Conflict Mediation

## Why

When AI coding assistants (such as Claude Code, Gemini CLI, and OpenCode) are invoked in Packet Tracer repositories, their lack of explicit boundary enforcement can lead them to hallucinate source code changes, write ad-hoc Python/Bash scripts to parse `.pkt` or `.xml` files, or make uncoordinated edits that violate internal invariants and corrupt topologies. Furthermore, during merge conflicts, agents should not blindly apply resolution flags (`--ours` or `--theirs`) without human oversight.

Establishing strict "operator-only" guardrails in the bundled skill ensures that AI agents interact with Packet Tracer repositories strictly through the official `pkt` CLI. In addition, providing a guided, human-in-the-loop conflict mediation protocol empowers the agent to inspect configuration diffs, explain changes in network engineering terms (devices, interfaces, IPs), and safely assist the user in resolving merge conflicts without risking network topology corruption.

## What Changes

- **Zero-Code Guardrail Invariant**: Explicitly define in the bundled skill that the agent is an operator/network assistant, not a repository software engineer. Forbid writing ad-hoc scripts, modifying tool internals, or running raw XML edits outside of `pkt edit`.
- **Guided Conflict Mediation Workflow**: Document a standard 4-step workflow for agents handling merge conflicts:
  1. Detect conflicted files using `pkt resolve` (report mode).
  2. Inspect textual diffs in `snapshot/**/configs/*.cfg` and `snapshot/**/metadata/topology.json`.
  3. Formulate a plain-language network explanation summarizing differences between branches (device names, IPs, VLANs, interfaces).
  4. Request user confirmation before applying `pkt resolve --ours` or `pkt resolve --theirs`.
- **Comprehensive Command Knowledge Base**: Consolidate and clarify the agent's knowledge of the full `pkt` CLI command surface (`init`, `export`, `build`, `inspect`, `edit`, `sync`, `resolve`, `verify`, `doctor`, `language`, `skills`).
- **Synchronized Skill Deployment**: Update the embedded asset in `internal/skills/assets/pkt-git/SKILL.md` and keep project repository agent directories (`.claude/skills/pkt-git/`, `.gemini/skills/pkt-git/`, `.opencode/skills/pkt-git/`) and installed user skills in sync.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `cli`: Update Requirement `Manage bundled agent skill` to specify that the bundled skill mandates operator-only guardrails (prohibiting code edits and ad-hoc scripts) and defines a guided conflict mediation workflow.

## Impact

- **Affected Code**: `internal/skills/assets/pkt-git/SKILL.md`, `.claude/skills/pkt-git/SKILL.md`, `.gemini/skills/pkt-git/SKILL.md`, `.opencode/skills/pkt-git/SKILL.md`, and test assertions in `internal/skills/skills_test.go`.
- **Dependencies**: No external dependencies.
- **Compatibility**: 100% backward-compatible. Enhances agent guidance and prompt security boundaries across all supported AI assistants.

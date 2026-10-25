# Design: Agent Skill Guardrails and Guided Conflict Mediation

## Context

`pkt-git` bundles an agent skill (`SKILL.md`) that is embedded into the Go binary at compile time via `go:embed` and deployed to Claude Code, Gemini CLI, and OpenCode via `pkt skills`. In addition, repository-level skill directories exist in `.claude/skills/pkt-git/`, `.gemini/skills/pkt-git/`, and `.opencode/skills/pkt-git/`.

While previous iterations documented the CLI command surface and low-level XML rules (e.g., avoiding `sed`), they lacked explicit role constraints to prevent agents from attempting to modify Go source code or writing ad-hoc scripts to parse `.pkt` or `.xml` files.

See `proposal.md` for background motivation and `specs/cli/spec.md` for requirement contracts.

## Goals / Non-Goals

**Goals:**
- Define explicit, high-priority negative constraints establishing that the agent is an operator/network assistant, not a software engineer.
- Forbid code edits, ad-hoc scripts, and direct XML manipulation outside the `pkt` CLI.
- Establish a structured, 4-step human-in-the-loop conflict mediation protocol for network topology diffs.
- Maintain complete content synchronization across embedded assets, repo-level skill files, and installed user skills.
- Validate guardrail presence via automated unit testing in `internal/skills/skills_test.go`.

**Non-Goals:**
- Modifying the underlying Go implementation of `pkt resolve`, `pkt edit`, or other subcommands.
- Creating an interactive TTY prompt inside the `pkt` Go binary (keeping CLI script-safe).

## Decisions

### Decision 1: Explicit Role & Operational Boundaries in Invariants

- **Choice**: Place an explicit "Role & Operational Boundary" invariant directly in the top section of `SKILL.md`.
- **Rationale**: LLM agents prioritize rules near the top of skill instructions. Explicitly defining:
  1. *"You are a network operator and Git workflow assistant, NOT a software developer for pkt-git."*
  2. *"Zero Code Modification: Do NOT edit Go/Python/shell code in this repository. Do NOT create ad-hoc scripts to parse or manipulate Packet Tracer files. Execute actions exclusively through the official `pkt` CLI."*
  prevents catastrophic hallucinations where the agent tries to invent Python scripts to unpack `.pkt` files or modify `internal/` packages.
- **Alternatives considered**: Relying on generic system prompts. Rejected because system prompts vary across agents, whereas `SKILL.md` is universal and embedded.

### Decision 2: 4-Step Guided Conflict Mediation Workflow

- **Choice**: Structure the "There is a merge conflict" recipe into four distinct phases:
  1. **Detection**: Run `pkt resolve` (report mode) to list conflicted snapshot paths.
  2. **Inspection**: Inspect configuration diffs (`snapshot/**/configs/*.cfg` and `snapshot/**/metadata/topology.json`) rather than opaque binary or XML.
  3. **Mediation**: Explain the network differences to the user in plain language (e.g., router names, interface IP changes, VLANs) and ask which branch's intent to apply.
  4. **Resolution**: Execute `pkt resolve --theirs` or `pkt resolve --ours` based strictly on user instruction, then verify with `pkt verify`.
- **Rationale**: Network topology conflicts cannot be resolved purely by mechanical text merging. Human domain knowledge is required to decide whether Branch A or Branch B carries the intended network state.

### Decision 3: Automated Invariant Testing

- **Choice**: Add assertions to `internal/skills/skills_test.go` verifying that `Asset()` contains the exact guardrail markers (e.g., `"Zero code modification"`, `"CLI operator"`, `"pkt resolve"`).
- **Rationale**: Prevents future regressions or accidental edits from dropping these safety guardrails.

## Risks / Trade-offs

- **[Risk] Prompt length vs context window** → Mitigation: Keep the guardrail additions concise and punchy without inflating token count.
- **[Risk] Desynchronization between embedded asset and repository skills** → Mitigation: Mirror changes across `internal/skills/assets/pkt-git/SKILL.md` and repo `.claude/`, `.gemini/`, `.opencode/` directories in a single atomic commit.

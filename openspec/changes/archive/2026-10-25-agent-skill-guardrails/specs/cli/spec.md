# Spec Delta

## MODIFIED Requirements

### Requirement: Manage the bundled agent skill

`pkt skills` SHALL install the bundled pkt-git skill into the global skill
directories for OpenCode, Claude Code, and Gemini CLI. `pkt skills --remove`
SHALL remove only the pkt-git skill from those locations. Both operations MUST
be idempotent and MUST report each agent's outcome. The bundled skill MUST
explicitly define operator-only guardrails forbidding source code modifications
and ad-hoc scripts, and MUST provide a structured, human-in-the-loop conflict
mediation workflow.

#### Scenario: Install the skill

- **WHEN** a user runs `pkt skills`
- **THEN** the pkt-git skill file is present in each supported agent's global
  skill directory
- **AND** the command reports success for each agent

#### Scenario: Remove the skill

- **WHEN** a user runs `pkt skills --remove`
- **THEN** only the pkt-git skill directories are removed
- **AND** other skills in those locations are untouched

#### Scenario: Missing bundled skill

- **WHEN** the bundled skill asset is unavailable
- **THEN** the command exits non-zero with a reinstall instruction
- **AND** it does not delete an existing installed skill

#### Scenario: Verify operator-only guardrail in bundled skill

- **WHEN** an AI assistant or test inspects the installed or bundled skill
- **THEN** the skill contains explicit invariants prohibiting source code modifications and ad-hoc scripts
- **AND** it instructs the agent to operate exclusively via official `pkt` CLI commands

#### Scenario: Verify guided conflict mediation in bundled skill

- **WHEN** a merge conflict is detected in a Packet Tracer repository
- **THEN** the skill directs the agent to inspect textual diffs in device configurations
- **AND** it requires the agent to explain differences in network terms and request user approval before applying `--ours` or `--theirs`

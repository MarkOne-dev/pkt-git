# Tasks

## 1. Skill Definition & Operator Guardrails

- [ ] 1.1 Update `internal/skills/assets/pkt-git/SKILL.md` to establish the Operator-Only Role and Zero Code Modification invariant (forbidding repository code edits and ad-hoc scripts), and verify content structure.
- [ ] 1.2 Expand the merge conflict section in `internal/skills/assets/pkt-git/SKILL.md` with the 4-step guided network mediation protocol (detection, config diff inspection, human-in-the-loop explanation, and resolution via `pkt resolve`).
- [ ] 1.3 Synchronize updated `SKILL.md` across repository agent skill directories (`.claude/skills/pkt-git/`, `.gemini/skills/pkt-git/`, `.opencode/skills/pkt-git/`) and user skill directory (`~/.gemini/skills/pkt-git/`).

## 2. Unit Testing & Invariant Assertions

- [ ] 2.1 Add invariant assertions in `internal/skills/skills_test.go` verifying that the embedded skill asset contains operator-only boundaries and conflict mediation directives.
- [ ] 2.2 Run `go test ./...` and `gofmt` to verify that all unit tests pass with 100% test suite success.

## Workflow follow-up

- Merge `feature/agent-skill-guardrails` into `develop`.
- Archive change `agent-skill-guardrails`.

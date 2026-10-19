# pkt-git Documentation

Complete technical documentation for **pkt-git**, the tool that turns Cisco Packet
Tracer projects into Git-native repositories.

## Document Index

| Document | What it covers |
|---|---|
| [Architecture Overview](architecture.md) | System components, the artifact-vs-source-of-truth model, and the full data flow from Packet Tracer to GitHub and back. |
| [The .pkt Binary Format](packet-tracer-format.md) | Deep dive into the reverse-engineered 4-layer crypto pipeline: qCompress, positional XOR, Twofish-EAX, and the outer XOR/reversal — with the exact constants. |
| [Snapshot Format](snapshot-format.md) | The textual source of truth written into the repo: `base.snapshot.xml`, per-device `.cfg` files, and `metadata/topology.json`. |
| [Collaborative Workflow](workflow.md) | The team SOP: daily loop, Git hooks automation, branching, Pull Requests, conflict resolution rules, and the one manual case (`pkt build`). |
| [CLI Reference](cli-reference.md) | Full reference for `tools/ptg.py`, `tools/edit_config.py`, and `tools/spike_roundtrip.py` — every command, argument, and output. |
| [Installation & Setup](installation.md) | Global one-machine install + `pkt init` per repository, platform notes. |
| [Troubleshooting & Pitfalls](troubleshooting.md) | Known failure modes, dangerous patterns (global `sed` on the XML), verification commands, and compatibility limits. |

## Suggested Reading Order

1. New team member → `workflow.md` (how to use it day to day).
2. Curious engineer → `architecture.md` → `packet-tracer-format.md` (how it works).
3. Maintainer resolving merges → `workflow.md` + `troubleshooting.md`.
4. Contributor modifying the code → everything, plus the source in `tools/`.

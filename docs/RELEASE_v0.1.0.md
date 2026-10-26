# pkt-git v0.1.0 - Git-Native Version Control for Cisco Packet Tracer

pkt-git bridges Cisco Packet Tracer network topologies with modern Git and GitHub workflows. It unpacks proprietary, encrypted Packet Tracer binaries (`.pkt` and `.pka`) into structured XML, individual Cisco IOS device configurations (`.cfg`), and topology metadata.

This initial release (`v0.1.0`) delivers a lightweight, statically compiled Go binary that operates externally on network files without requiring runtime dependencies or modifications to Packet Tracer itself.

### Key Capabilities

- **Self-Contained Architecture**: Statically compiled binary requiring no Python runtime, virtual environments, or external C toolchains.
- **Transparent Git Automation**: Native hooks that automatically export textual snapshots during commits and rebuild binary topologies during pulls and branch switches.
- **Granular Textual State**: Extracts complete topology XML alongside individual per-device Cisco IOS configurations for line-by-line diffing and pull request reviews.
- **Headless Terminal Editing**: Inspect and modify device interfaces or configuration lines directly from the command line without opening the graphical interface.
- **Conflict Resolution Engine**: Deterministic merge resolution that safely applies branch states and immediately regenerates consistent binary topologies.
- **Native Internationalization**: Built-in bilingual support in English and Spanish with structured error handling and contextual remediation guidance.
- **Environment Diagnostics**: Automated auditing of Git identity, system PATH, agent integrations, and cross-platform discovery of installed Cisco Packet Tracer instances.
- **Agent Skill Integration**: Bundled safety protocols and operational guardrails for AI coding assistants (Claude Code, Gemini CLI, OpenCode).

### Compatibility

Compatible with Cisco Packet Tracer 8.x and 9.x topologies across Linux, macOS, and Windows.

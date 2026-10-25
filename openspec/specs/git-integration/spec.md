# git-integration Specification

## Purpose

Defines how pkt-git wires itself into a Git repository and keeps binary
artifacts, textual snapshots, and merge resolutions synchronized across
commits, pulls, branch switches, and conflicts.

## Requirements

### Requirement: Initialize a repository

`pkt init [dir]` SHALL configure an existing Git repository by writing
executable thin hooks, recording the absolute path of the running `pkt`
executable in repository-local Git configuration, and activating the hooks
directory. The command MUST be idempotent and MUST refuse to modify a
directory that is not a Git repository.

#### Scenario: Initialize a new repository

- **WHEN** a user runs `pkt init` in a Git repository
- **THEN** `.githooks/pre-commit`, `.githooks/post-merge`, and
  `.githooks/post-checkout` exist and are executable
- **AND** the repository configuration points `core.hooksPath` at the hooks
  directory
- **AND** rerunning the command leaves the same usable configuration

#### Scenario: Reject a non-repository

- **WHEN** a user runs `pkt init` outside a Git repository
- **THEN** the command exits non-zero with an actionable message
- **AND** it does not create hooks or configuration

### Requirement: Merge required Git attributes

`pkt init` SHALL ensure that `.gitattributes` marks `.pkt` and `.pka` files
as binary and marks snapshot XML, JSON, and config files as text. It MUST
preserve unrelated existing rules, MUST NOT duplicate an existing required
rule, and MUST create the file when it is absent.

#### Scenario: Create missing attributes

- **WHEN** the repository has no `.gitattributes`
- **THEN** `pkt init` creates it with the binary and snapshot text rules
- **AND** reports that the rules were written

#### Scenario: Preserve unrelated attributes

- **WHEN** `.gitattributes` already contains unrelated rules
- **THEN** those rules remain unchanged
- **AND** only the missing pkt-git rules are appended once

### Requirement: Export staged binaries before commit

The `pre-commit` hook SHALL detect staged `.pkt` and `.pka` files, export each
into `snapshot/<project>/`, and stage the regenerated snapshot files. If any
export fails, the hook MUST abort the commit and MUST NOT stage a partial
snapshot.

#### Scenario: Commit a changed binary

- **WHEN** a user stages a changed `.pkt` and commits
- **THEN** the hook exports its snapshot before the commit completes
- **AND** the XML, configs, and metadata are included in the commit
- **AND** the command reports the exported project

#### Scenario: Export failure aborts the commit

- **WHEN** a staged binary cannot be decoded or exported
- **THEN** the hook exits non-zero
- **AND** Git does not create the commit
- **AND** the previous snapshot remains unchanged

#### Scenario: Text-only snapshot change

- **WHEN** a commit stages a changed `snapshot/*/base.snapshot.xml` without a
  corresponding binary
- **THEN** the hook allows the commit
- **AND** it prints a warning that the binary is stale and shows the rebuild
  command

### Requirement: Rebuild binaries after history changes

The `post-merge` and `post-checkout` hooks SHALL identify snapshot XML files
changed by the incoming history and rebuild the corresponding
`topologia/<project>.pkt` artifacts. The hooks MUST distinguish the
post-merge squash flag from a commit range, MUST use the checkout's old and
new revisions, and MUST NOT rebuild a binary whose snapshot did not change.

#### Scenario: Pull changes a snapshot

- **WHEN** a merge or pull changes `snapshot/<project>/base.snapshot.xml`
- **THEN** the hook rebuilds `topologia/<project>.pkt`
- **AND** it reports the rebuilt path

#### Scenario: Squash merge

- **WHEN** `post-merge` receives the squash flag
- **THEN** the hook treats the argument as a flag, not as a revision
- **AND** it rebuilds only snapshots changed by the merge

#### Scenario: Branch checkout

- **WHEN** a branch checkout changes a snapshot
- **THEN** the hook rebuilds that project's binary using the old and new
  revisions supplied by Git

#### Scenario: Unchanged snapshot

- **WHEN** a merge or checkout does not change a project's snapshot
- **THEN** the hook leaves that project's binary untouched

### Requirement: Resolve merge conflicts deterministically

`pkt resolve` SHALL report conflicted files and conflict-marker locations
without changing them. With `--ours` or `--theirs`, it SHALL apply the chosen
policy to the conflicted text files, verify that no conflict markers remain,
refresh derived snapshot files, rebuild conflicted binaries from the resolved
snapshot, and stage only the files involved in the resolution.

#### Scenario: Report mode

- **WHEN** a merge has conflicts and the user runs `pkt resolve` without a
  policy
- **THEN** the command lists conflicted text and binary files
- **AND** it reports the first conflict marker line for each text file
- **AND** it does not modify or stage any file

#### Scenario: Accept one side

- **WHEN** the user runs `pkt resolve --ours` or `pkt resolve --theirs`
- **THEN** the selected side is applied to the conflicted text files
- **AND** the command fails if any conflict marker remains
- **AND** derived files and conflicted binaries are regenerated
- **AND** only the resolved and regenerated paths are staged

#### Scenario: Binary-only conflict

- **WHEN** only a `.pkt` file is conflicted and its snapshot exists
- **THEN** the command rebuilds the binary from the resolved snapshot
- **AND** it stages the rebuilt binary

#### Scenario: No merge in progress

- **WHEN** a user runs `pkt resolve` without an active merge
- **THEN** the command exits non-zero
- **AND** it reports that there is nothing to resolve

### Requirement: Keep hook automation repository-local

The hooks and initialization MUST operate only through repository-local Git
configuration and MUST affect only pkt-git paths. They MUST NOT modify global
Git configuration or act on unrelated files in the repository.

#### Scenario: Unrelated commit

- **WHEN** a user commits a change that does not touch a `.pkt` or snapshot
  path
- **THEN** the hook performs no export or rebuild
- **AND** the commit proceeds normally

#### Scenario: No global side effects

- **WHEN** `pkt init` is run in one repository
- **THEN** other repositories on the machine keep their existing hook and
  attribute behavior

### Requirement: Auto-export existing topologies during initialization

`pkt init [dir]` SHALL discover existing `.pkt` and `.pka` files within the target repository. For each discovered topology, it MUST automatically export the initial snapshot into `snapshot/<project>/` and report the export in its output.

#### Scenario: Initialize repository with existing .pkt file
- **WHEN** a user runs `pkt init` in a repository containing `network.pkt`
- **THEN** git hooks and attributes are configured
- **AND** `network.pkt` is automatically exported to `snapshot/network/`
- **AND** the output reports both the hook initialization and the exported snapshot

#### Scenario: Initialize repository without topology files
- **WHEN** a user runs `pkt init` in a repository with no `.pkt` or `.pka` files
- **THEN** git hooks and attributes are configured
- **AND** no snapshot directories are created
- **AND** the command exits successfully

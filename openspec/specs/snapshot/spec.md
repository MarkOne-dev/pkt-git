# snapshot Specification

## Purpose

Defines the textual snapshot that acts as the source of truth for a Packet
Tracer project, including export, topology metadata, binary rebuilding,
derived-file synchronization, and read-only inspection.

## Requirements

### Requirement: Export a binary into the snapshot layout

`pkt export <project.pkt> <output_dir>` SHALL create or refresh a snapshot
directory containing `base.snapshot.xml`, `configs/`, and `metadata/`. The
command MUST write the decoded XML unchanged and MUST leave the previous
snapshot usable if the export fails.

#### Scenario: First export

- **WHEN** a user exports a valid Packet Tracer binary into an empty directory
- **THEN** the directory contains the full XML document, per-device configs,
  and `metadata/topology.json`
- **AND** the command reports the XML size, config count, device count, and
  link count

#### Scenario: Failed export preserves the previous snapshot

- **WHEN** decoding or parsing fails during a refresh of an existing snapshot
- **THEN** the command exits non-zero
- **AND** the previously exported files remain readable and unchanged

### Requirement: Derive per-device configuration files

The export and synchronization commands SHALL write one `.cfg` file for each
device whose `RUNNINGCONFIG` contains at least one line. The file content MUST
preserve the original line order and text, with one IOS line per text line.
Device names MUST be converted to safe file names without silently
overwriting another device.

#### Scenario: Device with a running config

- **WHEN** a device named `SW_ACC_ADMIN_PUNO` has a non-empty running config
- **THEN** `configs/SW_ACC_ADMIN_PUNO.cfg` contains those lines in order
- **AND** the file ends with a newline

#### Scenario: Device without a running config

- **WHEN** a device has no running-config lines
- **THEN** no `.cfg` file is created for that device

#### Scenario: Sanitized names collide

- **WHEN** two device names sanitize to the same file name
- **THEN** the command fails with an error naming both devices
- **AND** it does not overwrite either config file

### Requirement: Generate topology metadata

The export and synchronization commands SHALL generate
`metadata/topology.json` with the source file name, document root, total
device and link counts, a device catalog, and a link catalog. Device and link
order MUST follow document order, and unresolved cable endpoints MUST use a
deterministic fallback name rather than failing the export.

#### Scenario: Complete topology catalog

- **WHEN** a snapshot contains devices and links
- **THEN** `topology.json` lists each device's name, type, model, and reference
- **AND** each link records its type, source, destination, and ports
- **AND** the totals match the catalog lengths

#### Scenario: Unresolved cable endpoint

- **WHEN** a link references a device ID that is not present in the device
  catalog
- **THEN** the link is still exported with a deterministic fallback endpoint
  name
- **AND** the export completes successfully

### Requirement: Rebuild a binary from a snapshot

`pkt build <snapshot.xml> <output.pkt>` SHALL validate the XML document and
encode it into a Packet Tracer binary. The output MUST be written atomically,
and an existing binary MUST remain unchanged when validation or encoding
fails.

#### Scenario: Rebuild an existing artifact

- **WHEN** a user rebuilds a binary from a valid snapshot
- **THEN** the output path contains the newly encoded binary
- **AND** the snapshot XML is not modified

#### Scenario: Invalid snapshot leaves the binary untouched

- **WHEN** the snapshot XML is malformed
- **THEN** the command exits non-zero
- **AND** the previous binary at the output path remains unchanged

### Requirement: Synchronize derived snapshot files

`pkt sync <snapshot-dir|snapshot.xml>` SHALL regenerate `configs/*.cfg` and
`metadata/topology.json` from the snapshot XML. It MUST create missing derived
directories, remove stale generated files that no longer correspond to a
device, and MUST NOT modify `base.snapshot.xml`.

#### Scenario: Refresh derived files after a textual edit

- **WHEN** a user edits `base.snapshot.xml` and runs `pkt sync`
- **THEN** the configs and metadata reflect the edited XML
- **AND** the XML file is byte-for-byte unchanged

#### Scenario: Remove a stale config

- **WHEN** a device is removed from the snapshot XML and `pkt sync` runs
- **THEN** its old `.cfg` file is removed
- **AND** the command reports the refreshed counts

### Requirement: Inspect a project without writing files

`pkt inspect <project.pkt>` SHALL decode the binary and print a read-only
summary of devices, types, models, and whether each device has a CLI config.
It MUST NOT create or modify any file.

#### Scenario: Inspect a topology

- **WHEN** a user inspects a valid binary
- **THEN** the output lists the device and link totals and one row per device
- **AND** the working tree is unchanged

### Requirement: Preserve snapshot consistency invariants

The snapshot XML SHALL be the authoritative state. Commands that only read or
rebuild a binary MUST NOT rewrite the snapshot, and commands that mutate the
snapshot MUST leave it well-formed XML or leave the original file intact.

#### Scenario: Build does not rewrite the source

- **WHEN** a binary is rebuilt from a snapshot
- **THEN** the snapshot's bytes and modification time are unchanged

#### Scenario: Invalid mutation is rejected

- **WHEN** a snapshot mutation would produce malformed XML
- **THEN** the command exits non-zero
- **AND** the original snapshot remains valid and unchanged

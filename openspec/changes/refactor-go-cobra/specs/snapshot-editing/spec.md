# Spec Delta

## Purpose

Defines safe, device-scoped editing of snapshot XML and read-only display of
device configurations, so changes stay reviewable and cannot accidentally
alter unrelated network devices.

## ADDED Requirements

### Requirement: Scope every mutation to one device

Every mutating `pkt edit` command SHALL identify exactly one device by its
exact `<NAME>` value and SHALL restrict the edit to that device's
`RUNNINGCONFIG`. The command MUST fail when the device is missing or when the
name is ambiguous, and MUST NOT perform a document-wide replacement.

#### Scenario: Device not found

- **WHEN** a user edits a device name that is not present in the snapshot
- **THEN** the command exits non-zero with a "device not found" error
- **AND** the snapshot is unchanged

#### Scenario: Device without a running config

- **WHEN** a user edits an end device that has no `RUNNINGCONFIG`
- **THEN** the command exits non-zero with an explanatory error
- **AND** the snapshot is unchanged

### Requirement: Display a device configuration read-only

`pkt edit <xml> show <DEVICE>` SHALL print the device's running config without
writing any file. Lines that contain credentials or other sensitive values
MUST be redacted before printing.

#### Scenario: Show a normal configuration

- **WHEN** a user shows a device with a running config
- **THEN** the output lists its IOS lines in order
- **AND** the snapshot's bytes are unchanged

#### Scenario: Redact secrets

- **WHEN** a config line contains a password, secret, key, or community value
- **THEN** the displayed line masks the sensitive value
- **AND** the redacted text is not written back to the snapshot

### Requirement: Replace an interface block

`pkt edit <xml> setiface <DEVICE> <IFACE> <line>...` SHALL replace the content
between the matching `interface <IFACE>` line and the next `!` line with the
provided lines. It MUST preserve all content outside that interface block and
MUST fail if the interface is not present.

#### Scenario: Replace an interface

- **WHEN** a user sets an interface block on an existing interface
- **THEN** that block contains exactly the supplied lines
- **AND** the rest of the device config and document is unchanged

#### Scenario: Missing interface

- **WHEN** the requested interface does not exist in the device config
- **THEN** the command exits non-zero
- **AND** no file is written

### Requirement: Change one exact configuration line

`pkt edit <xml> chgline <DEVICE> <IFACE> <old> <new>` SHALL replace one
complete `<LINE>` value inside the specified interface block. The old value
MUST match the full line text, not a substring, and the command MUST fail when
zero or multiple lines match.

#### Scenario: Replace a complete line

- **WHEN** the interface contains exactly one line equal to the old value
- **THEN** that line is replaced with the new value
- **AND** no other line is changed

#### Scenario: Substring does not match

- **WHEN** the old value appears only as part of a longer line
- **THEN** the command exits non-zero with a "line not found" error
- **AND** the snapshot is unchanged

#### Scenario: Ambiguous match

- **WHEN** more than one line in the interface equals the old value
- **THEN** the command exits non-zero with an ambiguity error
- **AND** the snapshot is unchanged

### Requirement: Insert lines after a matching config line

`pkt edit <xml> insertafter <DEVICE> <phrase> <line>...` SHALL insert the new
lines immediately after the first config line containing the phrase. It MUST
fail when no line contains the phrase and MUST preserve all other lines.

#### Scenario: Insert after a match

- **WHEN** a device config contains the requested phrase
- **THEN** the new lines appear immediately after the first matching line
- **AND** the original matching line and all other lines remain in place

#### Scenario: Phrase not found

- **WHEN** no config line contains the phrase
- **THEN** the command exits non-zero
- **AND** the snapshot is unchanged

### Requirement: Validate and preserve the document

Before writing, every mutation SHALL verify that the resulting document is
well-formed XML. The write MUST be atomic, MUST preserve the original
encoding and line-ending convention outside the edited span, and MUST leave
the original file intact on failure.

#### Scenario: Invalid result is rejected

- **WHEN** a proposed edit would produce malformed XML
- **THEN** the command exits non-zero
- **AND** the original snapshot remains byte-for-byte unchanged

#### Scenario: Preserve CRLF formatting

- **WHEN** the snapshot uses CRLF line endings and a single line is changed
- **THEN** all untouched lines keep their CRLF endings
- **AND** only the edited span reflects the requested change

### Requirement: Refresh derived files after an edit

After a successful mutation, the command SHALL refresh the snapshot's derived
`configs/*.cfg` and `metadata/topology.json` files so the snapshot does not
contradict itself. If synchronization fails, the command MUST report the
failure and exit non-zero.

#### Scenario: Successful edit refreshes derived files

- **WHEN** a user changes an interface line
- **THEN** the corresponding `.cfg` file reflects the new line
- **AND** `metadata/topology.json` remains consistent with the snapshot
- **AND** the command exits successfully

#### Scenario: Synchronization failure is reported

- **WHEN** the edit succeeds but derived-file synchronization fails
- **THEN** the command exits non-zero
- **AND** it reports that the snapshot was changed but derived files may be
  stale

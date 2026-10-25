# codec Specification

## Purpose

Defines the reversible, authenticated conversion between Cisco Packet Tracer
binary projects and their UTF-8 XML document, so Git tooling can inspect and
rebuild them without losing data or accepting corrupt input.

## Requirements

### Requirement: Decode Packet Tracer binaries

The system SHALL decode `.pkt` and `.pka` files by reversing the four
container layers and SHALL produce the original UTF-8 XML document. Decoding
MUST fail without producing usable output when the container cannot be
authenticated or decompressed.

#### Scenario: Decode a supported specimen

- **WHEN** a user exports or inspects a binary produced by Packet Tracer
  9.0.1.0858
- **THEN** the decoded document is well-formed XML with a `PACKETTRACER5` root
- **AND** the command completes successfully

#### Scenario: Reject a truncated binary

- **WHEN** the input ends before the authentication tag or compressed payload
  is complete
- **THEN** the command exits non-zero
- **AND** it reports that the container is truncated or unauthenticated
- **AND** it does not leave a partial snapshot or output file behind

### Requirement: Encode XML into a Packet Tracer binary

The system SHALL encode a UTF-8 XML document into a binary that Packet Tracer
can open, using the same compression, masking, and authenticated-encryption
layers as the decoder. Encoding the same XML document twice MUST produce the
same bytes.

#### Scenario: Rebuild a binary from a snapshot

- **WHEN** a user runs `pkt build` with a well-formed snapshot XML
- **THEN** the command writes a binary file
- **AND** decoding that binary reproduces the input XML exactly
- **AND** the command reports the output path and byte size

#### Scenario: Refuse invalid XML

- **WHEN** the input document is not well-formed XML
- **THEN** the command exits non-zero before writing the output binary
- **AND** it reports the XML validation failure with the source path

### Requirement: Authenticate and validate the container

Every decode operation MUST verify the EAX authentication tag and the declared
uncompressed length before returning XML. A mismatch MUST be a hard failure by
default; the system MUST NOT silently return data from a failed
authentication check.

#### Scenario: Authentication tag mismatch

- **WHEN** the decoded ciphertext does not match the appended tag
- **THEN** the operation fails with an authentication error
- **AND** the caller receives no XML document

#### Scenario: Declared length mismatch

- **WHEN** the decompressed XML length differs from the length stored in the
  container header
- **THEN** the operation fails with a length-mismatch error
- **AND** the caller receives no XML document

### Requirement: Verify round-trip integrity

The `pkt verify` command SHALL decode a specimen, re-encode the decoded XML,
decode the rebuilt binary, and report the result. It MUST exit successfully
only when the decoded XML is well-formed and the XML round trip is exact.
Byte-for-byte equality with the original binary MUST be reported as
informational and MUST NOT be required for success.

#### Scenario: Successful verification

- **WHEN** a user runs `pkt verify` on a supported specimen
- **THEN** the report states that the XML is a document and that the XML round
  trip is exact
- **AND** the command exits with status 0

#### Scenario: Failed verification

- **WHEN** the rebuilt binary decodes to different XML or to a non-document
- **THEN** the report identifies the failed invariant
- **AND** the command exits non-zero

### Requirement: Preserve format compatibility boundaries

The system SHALL treat the container format as verified only for the Packet
Tracer versions covered by the bundled regression specimen. When decoding
fails, the error MUST identify the input and suggest verifying the producing
Packet Tracer version rather than claiming support for an untested format.

#### Scenario: Unsupported or damaged specimen

- **WHEN** a binary from an untested Packet Tracer version cannot be decoded
- **THEN** the command reports the file path and the failed layer
- **AND** it advises collecting the producing Packet Tracer version before
  treating the format as supported

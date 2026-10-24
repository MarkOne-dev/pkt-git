# The .pkt Binary Format

Reverse-engineered and verified against a real Cisco Packet Tracer file
(`testdata/specimens/Caso1_Grupo09-2.pkt`, Packet Tracer **9.0.1.0858**).
Encoding and decoding are implemented in `internal/codec`, with CMAC/EAX in
`internal/crypto/{cmac,eax}` and Twofish from `golang.org/x/crypto/twofish`.
A full round trip (`.pkt → XML → .pkt → same XML`) is the standing regression
test.

## Summary of the Pipeline

A `.pkt` file is a UTF-8 XML document (`<PACKETTRACER5>` root, ~5.9 MB for a
107-device topology) protected by four reversible layers. **Encoding** applies
them 1 → 4; **decoding** inverts them 4 → 1.

| # | Layer | Direction (encode) |
|---|---|---|
| 1 | qCompress framing | Prefix the zlib stream with the 4-byte big-endian uncompressed size |
| 2 | Inner positional XOR | `b[i] ^= (L - i) mod 256` |
| 3 | Twofish-EAX encryption | Fixed key/IV, 16-byte OMAC tag appended to the ciphertext |
| 4 | Outer positional XOR + reversal | `out[L-1-i] = in[i] ^ ((L - i*L) mod 256)` |

## Layer Details

### 1. qCompress framing (Qt's qCompress)

Packet Tracer (a Qt application) wraps the XML with `qCompress`:

```
payload = uncompressed_size (4 bytes, big-endian) || zlib_stream
```

The leading length lets a decoder know the exact decompressed size in advance.
The Go implementation uses `compress/zlib` at level 9 and writes the size header
manually; the output is compatible with Qt's `qUncompress` and with
`zlib.decompress`.

### 2. Inner positional XOR

```
b[i] = payload[i] XOR ((L - i) mod 256)
```

A per-byte mask that slides with position: the first byte is XORed with
`L mod 256`, the last one with `1 mod 256`. `L` is the payload length. The
operation is its own inverse shape (`maskStage2` == `unmaskStage2`).

### 3. Twofish-EAX encryption

AEAD built over the Twofish block cipher (128-bit blocks, 16-byte key):

- **Key:** `0x89` repeated 16 times (`b"\x89" * 16`)
- **Nonce/IV:** `0x10` repeated 16 times (`b"\x10" * 16`)
- **Header:** empty (`header = b""`)
- **Layout:** the tag (16 bytes) is **appended at the end** of the ciphertext —
  Crypto++-style `data || tag` — and the OMAC tag covers the ciphertext (not the
  plaintext). This is the exact EAX variant PT uses.

Two fixed constants are enough because there is no per-file keying: the
encryption strength is the same for every PT file of this version, which is
precisely what makes external tooling feasible.

#### The EAX internals (CMAC/EAX in-house, Twofish from x/crypto)

- **CMAC / OMAC1** exactly as NIST SP 800-38B: derive subkeys `k1`, `k2` by
  doubling in GF(2^128) with reduction polynomial `0x87`; the last block is
  either complete (XOR `k1`) or a padded fragment (`0x80` terminator, XOR `k2`).
- **OMAC^t(x)** — the EAX tweak: CMAC over `x` *prefixed* by a full block
  `0x00 × 15 || t` (tweak 0 = nonce, 1 = header, 2 = ciphertext).
  Getting this prefix wrong (using a single byte instead of a full block) is the
  classic EAX pitfall — documented in the source so nobody "simplifies" it.
- **CTR mode:** keystream of Twofish-ECB encryptions of the 128-bit counter
  starting at `n = OMAC^0(nonce)`, big-endian increment.
- **Tag:** `T = N ⊕ H ⊕ C` where `N = OMAC^0(nonce)`, `H = OMAC^1(header)`,
  `C = OMAC^2(ciphertext)`.

On decode, the tag is verified before any XML is returned. A mismatch is a hard
failure (`eax.ErrAuthentication`): the codec returns no plaintext and no XML, so
callers cannot accidentally inspect or export unauthenticated data.

### 4. Outer positional XOR + byte reversal

The outermost layer combines a mirror and a position-indexed mask:

```
out[L-1-i] = in[i] XOR ((L - i*L) mod 256)
```

Crucial subtlety (learned the hard way): the XOR operand `L - i*L` indexes the
**output** position `i`, while the byte is read from the **reversed** input
position (`L-1-i`). Inverting it writes `out[i]` from `raw[L-1-i]`. Both
`maskStage4` / `unmaskStage4` implement and document this.

## What Decoding Looks Like

For the reference specimen (real project, PT 9.0.1.0858):

```
binary size        : 598810 bytes
extracted XML      : 5894851 bytes   (107 devices, 106 links, 34 CLI configs)
rebuilt binary     : 585372 bytes
```

Verification command:

```bash
pkt verify testdata/specimens/Caso1_Grupo09-2.pkt
```

Until the CLI command is wired, the same check runs directly against the codec:

```bash
go test ./internal/codec -run TestVerifySpecimen -v
```

Expected report from `codec.Verify`:

- `XML is a document : true` — decrypts into well-formed XML.
- `XML round trip : true` — required; the invariant of the whole project.
- `byte-identical : false` — **expected and acceptable**: Packet Tracer's own
  zlib emitter uses different compression parameters than the Go level-9
  encoder, so the rebuilt bytes differ while remaining a fully valid PT file
  (PT opens it identically).

## Known Limits

- Verified only against PT **9.0.1.0858**. Older branches (8.x, 7.x, 6.x) are
  expected to share the scheme but are untested — collect specimens before
  claiming support (see `troubleshooting.md`).
- `.pka` activity files are expected to use the same container, but the inner
  document structure differs (activities), so they are untested.
- `pkt build` recompresses deterministically at zlib level 9; it does not
  attempt to byte-match PT's own output (unnecessary for validity).

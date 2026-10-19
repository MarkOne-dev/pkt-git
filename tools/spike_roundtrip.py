"""Reversible codec for Cisco Packet Tracer files (.pkt/.pka).

Pipeline documented and verified against PT 9.0.1.0858:

    1. qCompress: 4-byte BE uncompressed size + zlib stream
    2. Positional XOR: b[i] ^= (L - i) mod 256
    3. Twofish-EAX: key 0x89*16, IV 0x10*16, tag appended to the blob
    4. Positional XOR with reversal: out[L-1-i] = in[i] ^ (L - i*L) mod 256

Decoding inverts 4 -> 3 -> 2 -> 1.

The Twofish block comes from the `twofish` package as a validation
reference only; CMAC/OMAC and EAX are an in-house implementation.
"""

import sys
import zlib
from pathlib import Path

from twofish import Twofish

KEY = b"\x89" * 16
NONCE = b"\x10" * 16
BS = 16


def _dbl(s: bytes) -> bytes:
    """Double in GF(2^128) with reduction polynomial 0x87."""
    n = int.from_bytes(s, "big")
    carry = (n >> 127) & 1
    n = (n << 1) & ((1 << 128) - 1)
    if carry:
        n ^= 0x87
    return n.to_bytes(BS, "big")


def _xor(a: bytes, b: bytes) -> bytes:
    return bytes(x ^ y for x, y in zip(a, b))


def cmac(cipher: Twofish, msg: bytes) -> bytes:
    """Standard CMAC/OMAC1 (NIST SP 800-38B) over the given block cipher."""
    L = cipher.encrypt(bytes(BS))
    k1, k2 = _dbl(L), _dbl(_dbl(L))
    acc = bytes(BS)

    if len(msg) == 0:
        last = _xor(b"\x80" + bytes(BS - 1), k2)
    else:
        r = len(msg) % BS
        if r == 0:
            body, last = msg[:-BS], _xor(msg[-BS:], k1)
        else:
            body, last = msg[:-r], _xor(msg[-r:] + b"\x80" + bytes(BS - r - 1), k2)
        for i in range(0, len(body), BS):
            acc = cipher.encrypt(_xor(acc, body[i:i + BS]))
    return cipher.encrypt(_xor(acc, last))


def omac(cipher: Twofish, tweak: int, data: bytes) -> bytes:
    """EAX's OMAC^t(x): full-block prefix 00*15 || t."""
    return cmac(cipher, b"\x00" * (BS - 1) + bytes([tweak]) + data)


def _ctr_keystream(cipher: Twofish, n: bytes, length: int) -> bytes:
    """CTR keystream with initial counter n, big-endian 128-bit increment."""
    out = bytearray()
    counter = int.from_bytes(n, "big")
    while len(out) < length:
        out += cipher.encrypt((counter & ((1 << 128) - 1)).to_bytes(BS, "big"))
        counter += 1
    return bytes(out[:length])


def eax_encrypt(cipher: Twofish, plain: bytes, nonce: bytes, header: bytes = b"") -> bytes:
    """EAX-encrypt Crypto++ style: returns ciphertext + tag (16 B) concatenated."""
    n = omac(cipher, 0, nonce)
    ks = _ctr_keystream(cipher, n, len(plain))
    data = _xor(plain, ks)
    tag = _xor(_xor(n, omac(cipher, 1, header)), omac(cipher, 2, data))
    return data + tag


def eax_decrypt(cipher: Twofish, data: bytes, nonce: bytes, header: bytes = b""):
    """EAX-decrypt with trailing tag. Returns (plaintext, expected_tag_hex, actual_tag_hex)."""
    blob, tag = data[:-16], data[-16:]
    n = omac(cipher, 0, nonce)
    ks = _ctr_keystream(cipher, n, len(blob))
    plain = _xor(blob, ks)
    expected = _xor(_xor(n, omac(cipher, 1, header)), omac(cipher, 2, blob))
    return plain, expected.hex(), tag.hex()


def _unmask_stage4(raw: bytes) -> bytes:
    """Invert out[L-1-i] = enc[i] ^ (L - i*L): XOR indexed by output position."""
    L = len(raw)
    out = bytearray(L)
    for i in range(L):
        out[i] = raw[L - 1 - i] ^ ((L - i * L) & 0xFF)
    return bytes(out)


def _mask_stage4(data: bytes) -> bytes:
    L = len(data)
    out = bytearray(L)
    for i in range(L):
        out[L - 1 - i] = data[i] ^ ((L - i * L) & 0xFF)
    return bytes(out)


def _unmask_stage2(data: bytes) -> bytes:
    return bytes(b ^ ((len(data) - i) & 0xFF) for i, b in enumerate(data))


def _mask_stage2(data: bytes) -> bytes:
    return bytes(b ^ ((len(data) - i) & 0xFF) for i, b in enumerate(data))


def decode(raw: bytes) -> str:
    """Convert a binary .pkt file into its originating XML (str)."""
    data = _unmask_stage4(raw)
    cipher = Twofish(KEY)
    decrypted, expected_tag, actual_tag = eax_decrypt(cipher, data, NONCE)
    if expected_tag != actual_tag:
        print(f"[warn] EAX MAC mismatch ({expected_tag[:16]}... vs {actual_tag[:16]}...)",
              file=sys.stderr)
    payload = _unmask_stage2(decrypted)
    size = int.from_bytes(payload[:4], "big")
    xml = zlib.decompress(payload[4:])
    if len(xml) != size:
        print(f"[warn] declared length {size} != actual {len(xml)}", file=sys.stderr)
    return xml.decode("utf-8")


def encode(xml_text: str) -> bytes:
    """Rebuild the binary .pkt from the XML (str)."""
    xml = xml_text.encode("utf-8")
    payload = len(xml).to_bytes(4, "big") + zlib.compress(xml, 9)
    payload = _mask_stage2(payload)
    cipher = Twofish(KEY)
    data = eax_encrypt(cipher, payload, NONCE)
    return _mask_stage4(data)


# backward-compatibility aliases
_inferno_decode = decode
_inferno_encode = encode


def cmd_decode(src: Path, out: Path) -> int:
    xml = decode(src.read_bytes())
    out.write_text(xml, encoding="utf-8")
    print(f"[ok] decode {src} -> {out}  ({src.stat().st_size} B -> {len(xml)} B)")
    return 0


def cmd_encode(src: Path, out: Path) -> int:
    data = encode(src.read_text(encoding="utf-8"))
    out.write_bytes(data)
    print(f"[ok] encode {src} -> {out}  ({len(data)} B)")
    return 0


def cmd_verify(src: Path, tmpdir: Path) -> int:
    """Round-trip test: decode -> encode -> decode must yield the same XML."""
    raw = src.read_bytes()

    xml1 = decode(raw)
    rebuilt = encode(xml1)

    xml_path = tmpdir / f"{src.stem}.snapshot.xml"
    xml_path.write_text(xml1, encoding="utf-8")
    pkt_path = tmpdir / f"{src.stem}.rebuilt.pkt"
    pkt_path.write_bytes(rebuilt)

    xml2 = decode(rebuilt)
    ok_xml = xml1 == xml2
    ok_start = xml1.lstrip().startswith("<")
    ok_bytes = rebuilt == raw

    print("=" * 60)
    print(f"specimen           : {src.name}")
    print(f"binary size        : {len(raw)} B")
    print(f"extracted XML      : {len(xml1)} B  ({xml_path})")
    print(f"rebuilt binary     : {len(rebuilt)} B  ({pkt_path})")
    print(f"XML is a document  : {ok_start}")
    print(f"XML ROUND TRIP     : {ok_xml}")
    print(f"byte-identical     : {ok_bytes} (bonus: requires PT's zlib emitter)")
    print("=" * 60)
    return 0 if (ok_xml and ok_start) else 1


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(2)
    cmd, *args = sys.argv[1:]
    if cmd == "decode" and len(args) == 2:
        sys.exit(cmd_decode(Path(args[0]), Path(args[1])))
    if cmd == "encode" and len(args) == 2:
        sys.exit(cmd_encode(Path(args[0]), Path(args[1])))
    if cmd == "verify" and len(args) == 1:
        sys.exit(cmd_verify(Path(args[0]), Path("out")))
    print(__doc__)
    sys.exit(2)

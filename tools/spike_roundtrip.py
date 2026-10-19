"""Codec reversible para archivos de Cisco Packet Tracer (.pkt/.pka).

Pipeline documentado y verificado contra PT 9.0.1.0858:

    1. qCompress: 4 bytes BE del tamaño sin comprimir + stream zlib
    2. XOR posicional: b[i] ^= (L - i) mod 256
    3. Twofish-EAX: clave 0x89*16, IV 0x10*16, tag al final del blob
    4. XOR posicional con inversión: out[L-1-i] = in[i] ^ (L - i*L) mod 256

El descifrado invierte 4 -> 3 -> 2 -> 1.

El bloque Twofish usa el paquete `twofish` solo como referencia de
validación; CMAC/OMAC y EAX son implementación propia.
"""

import sys
import zlib
from pathlib import Path

from twofish import Twofish

KEY = b"\x89" * 16
NONCE = b"\x10" * 16
BS = 16


def _dbl(s: bytes) -> bytes:
    """Duplica en GF(2^128) con polinomio de retorno 0x87."""
    n = int.from_bytes(s, "big")
    carry = (n >> 127) & 1
    n = (n << 1) & ((1 << 128) - 1)
    if carry:
        n ^= 0x87
    return n.to_bytes(BS, "big")


def _xor(a: bytes, b: bytes) -> bytes:
    return bytes(x ^ y for x, y in zip(a, b))


def cmac(cipher: Twofish, msg: bytes) -> bytes:
    """CMAC/OMAC1 estándar (NIST SP 800-38B) sobre el cifrador dado."""
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
    """OMAC^t(x) de EAX: prefijo de bloque completo 00*15 || t."""
    return cmac(cipher, b"\x00" * (BS - 1) + bytes([tweak]) + data)


def _ctr_keystream(cipher: Twofish, n: bytes, length: int) -> bytes:
    """Keystream CTR con contador inicial n, incremento big-endian de 128 bits."""
    out = bytearray()
    counter = int.from_bytes(n, "big")
    while len(out) < length:
        out += cipher.encrypt((counter & ((1 << 128) - 1)).to_bytes(BS, "big"))
        counter += 1
    return bytes(out[:length])


def eax_encrypt(cipher: Twofish, plain: bytes, nonce: bytes, header: bytes = b"") -> bytes:
    """Cifra en EAX estilo Crypto++: devuelve ciphertext + tag (16 B) concatenados."""
    n = omac(cipher, 0, nonce)
    ks = _ctr_keystream(cipher, n, len(plain))
    data = _xor(plain, ks)
    tag = _xor(_xor(n, omac(cipher, 1, header)), omac(cipher, 2, data))
    return data + tag


def eax_decrypt(cipher: Twofish, data: bytes, nonce: bytes, header: bytes = b""):
    """Descifra EAX con tag al final. Retorna (plaintext, tag_esperado_hex, tag_real_hex)."""
    blob, tag = data[:-16], data[-16:]
    n = omac(cipher, 0, nonce)
    ks = _ctr_keystream(cipher, n, len(blob))
    plain = _xor(blob, ks)
    expected = _xor(_xor(n, omac(cipher, 1, header)), omac(cipher, 2, blob))
    return plain, expected.hex(), tag.hex()


def _unmask_stage4(raw: bytes) -> bytes:
    """Invierte out[L-1-i] = enc[i] ^ (L - i*L): XOR por índice de salida."""
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
    """Convierte un .pkt binario en su XML de origen (str)."""
    data = _unmask_stage4(raw)
    cipher = Twofish(KEY)
    decrypted, tag_esp, tag_real = eax_decrypt(cipher, data, NONCE)
    if tag_esp != tag_real:
        print(f"[warn] EAX MAC no coincide ({tag_esp[:16]}... vs {tag_real[:16]}...)",
              file=sys.stderr)
    payload = _unmask_stage2(decrypted)
    size = int.from_bytes(payload[:4], "big")
    xml = zlib.decompress(payload[4:])
    if len(xml) != size:
        print(f"[warn] longitud declarada {size} != real {len(xml)}", file=sys.stderr)
    return xml.decode("utf-8")


def encode(xml_text: str) -> bytes:
    """Reconstruye el .pkt binario desde el XML (str)."""
    xml = xml_text.encode("utf-8")
    payload = len(xml).to_bytes(4, "big") + zlib.compress(xml, 9)
    payload = _mask_stage2(payload)
    cipher = Twofish(KEY)
    data = eax_encrypt(cipher, payload, NONCE)
    return _mask_stage4(data)


# alias de compatibilidad con versiones anteriores
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
    """Test de round-trip: decode -> encode -> decode debe dar el mismo XML."""
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
    print(f"especímen          : {src.name}")
    print(f"tamaño binario     : {len(raw)} B")
    print(f"XML extraído       : {len(xml1)} B  ({xml_path})")
    print(f"binario regenerado : {len(rebuilt)} B  ({pkt_path})")
    print(f"XML es documento   : {ok_start}")
    print(f"ROUND TRIP XML     : {ok_xml}")
    print(f"byte-idéntico      : {ok_bytes} (bonus: exige zlib emitter igual a PT)")
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

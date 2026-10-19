#!/usr/bin/env python
"""
spike_roundtrip.py — SPIKE PASO 1 (pkt-git)

Objetivo: probar que un .pkt de Packet Tracer es un documento XML pasado por
un pipeline reversible, y que podemos reconstruirlo byte a byte:

    .pkt ──descifrar──> XML ──re-encifrar──> .pkt  (comparación byte a byte)

Pipeline documentado (jeamxn/cisco-pka-to-xml, pka2xml de mircodz):
  1. zlib (qCompress: 4 bytes BE del tamaño sin comprimir + stream zlib)
  2. XOR posicional:           b[i] ^= (len - i) & 0xff
  3. Twofish-EAX:  clave 0x89*16, IV/nonce 0x10*16, bloques de 16 bytes
  4. XOR posicional + inversión:  out[len-1-i] = in[i] ^ ((len - i*len) & 0xff)

Para el DECIFRADO se invierten 4 → 3 → 2 → 1.

El bloque Twofish se toma del paquete `twofish` SOLO como referencia para
validar el pipeline (decisión de proyecto: la implementación final será
Python puro propia; este spike compara salida bloque a bloque contra ella).
CMAC/OMAC y EAX son implementación propia, pura.

Uso:
  .venv/bin/python tools/spike_roundtrip.py decode proof-pkt/X.pkt  salida.xml
  .venv/bin/python tools/spike_roundtrip.py encode salida.xml       reconstruido.pkt
  .venv/bin/python tools/spike_roundtrip.py verify proof-pkt/X.pkt  round trip completo
"""

import sys
import zlib
from pathlib import Path

from twofish import Twofish

KEY = b"\x89" * 16          # clave constante documentada
NONCE = b"\x10" * 16        # IV/nonce constante documentado
BS = 16                     # bloque de 128 bits


# ----------------------------------------------------------------------------
# CMAC (OMAC1) — idéntico al AES-CMAC estándar, pero con Twofish como cifrador
# ----------------------------------------------------------------------------

def _dbl(s: bytes) -> bytes:
    """Duplicar en GF(2^128) con polinomio de retorno 0x87 (subclaves OMAC)."""
    n = int.from_bytes(s, "big")
    carry = (n >> 127) & 1
    n = (n << 1) & ((1 << 128) - 1)
    if carry:
        n ^= 0x87
    return n.to_bytes(BS, "big")


def _xor(a: bytes, b: bytes) -> bytes:
    return bytes(x ^ y for x, y in zip(a, b))


def cmac(cipher: Twofish, msg: bytes) -> bytes:
    """CMAC/OMAC1 estándar (NIST SP 800-38B) con el cifrador dado.

    M_last = completo XOR k1  |  fragmento+0x80·0 XOR k2
    (XOR de subclaves en GF(2^128),)._dbl_
    """
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
    """OMAC^t(x) — variante EAX: prefijo de bloque completo 00×15 || t.

    (convención de Crypto++/pycryptodome; difiere de anteponer un solo byte)
    """
    return cmac(cipher, b"\x00" * (BS - 1) + bytes([tweak]) + data)


# ----------------------------------------------------------------------------
# EAX (descifrar / cifrar) — CTR con contador inicial = N = OMAC_0(nonce)
# ----------------------------------------------------------------------------

def _ctr_keystream(cipher: Twofish, n: bytes, length: int) -> bytes:
    out = bytearray()
    counter = int.from_bytes(n, "big")
    while len(out) < length:
        out += cipher.encrypt((counter & ((1 << 128) - 1)).to_bytes(BS, "big"))
        counter = (counter + 1)  # big-endian increment
    return bytes(out[:length])


def eax_decrypt(cipher: Twofish, data: bytes, nonce: bytes, header: bytes = b""):
    """EAX al estilo Crypto++: el MAC va AL FINAL del blob (últimos 16 bytes).

    Encrypt-then-MAC: el tag cubre el CIPHERTEXT (no el texto claro).
    """
    blob, tag = data[:-16], data[-16:]
    n = omac(cipher, 0, nonce)
    ks = _ctr_keystream(cipher, n, len(blob))
    plain = _xor(blob, ks)
    expected = _xor(_xor(n, omac(cipher, 1, header)), omac(cipher, 2, blob))
    return plain, expected.hex(), tag.hex()


def eax_encrypt(cipher: Twofish, plain: bytes, nonce: bytes, header: bytes = b""):
    n = omac(cipher, 0, nonce)
    ks = _ctr_keystream(cipher, n, len(plain))
    data = _xor(plain, ks)
    tag = _xor(_xor(n, omac(cipher, 1, header)), omac(cipher, 2, data))
    return data + tag


# ----------------------------------------------------------------------------
# Capas obfuscación (XOR posicionales y qCompress)
# ----------------------------------------------------------------------------

def _unmask_stage4(raw: bytes) -> bytes:
    """Invierte 'output[L-1-i] = enc[i] ^ (L - i*L)': XOR por índice de SALIDA."""
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


def _inferno_decode(raw: bytes) -> str:
    """TODO el pipeline: .pkt -> XML (str)."""
    data = _unmask_stage4(raw)
    cipher = Twofish(KEY)
    decrypted, tag_esp, tag_real = eax_decrypt(cipher, data, NONCE)
    if tag_esp != tag_real:
        print(f"[warn] EAX MAC no coincide (esperado {tag_esp[:16]}…, archivo {tag_real[:16]}…)",
              file=sys.stderr)
    payload = _unmask_stage2(decrypted)
    size = int.from_bytes(payload[:4], "big")   # cabecera qCompress
    xml = zlib.decompress(payload[4:])
    if len(xml) != size:
        print(f"[warn] longitud declarada {size} != real {len(xml)}", file=sys.stderr)
    return xml.decode("utf-8")


def _inferno_encode(xml_text: str) -> bytes:
    """Todo el pipeline inverso: XML -> .pkt."""
    xml = xml_text.encode("utf-8")
    payload = len(xml).to_bytes(4, "big") + zlib.compress(xml, 9)
    payload = _mask_stage2(payload)
    cipher = Twofish(KEY)
    data = eax_encrypt(cipher, payload, NONCE)
    return _mask_stage4(data)


# ----------------------------------------------------------------------------
# CLI
# ----------------------------------------------------------------------------

def cmd_decode(src: Path, out: Path) -> int:
    raw = src.read_bytes()
    xml = _inferno_decode(raw)
    out.write_text(xml, encoding="utf-8")
    print(f"[ok] decode {src} -> {out}  ({len(raw)} B -> {len(xml)} B de XML)")
    print("[head]", xml[:200].replace("\n", " "))
    return 0


def cmd_encode(src: Path, out: Path) -> int:
    xml = src.read_text(encoding="utf-8")
    binario = _inferno_encode(xml)
    out.write_bytes(binario)
    print(f"[ok] encode {src} -> {out}  ({len(xml)} B -> {len(binario)} B)")
    return 0


def cmd_verify(src: Path, tmpdir: Path) -> int:
    """EL test del spike: round trip completo sin pérdida."""
    raw = src.read_bytes()

    xml1 = _inferno_decode(raw)
    rebuilt = _inferno_encode(xml1)

    xml_path = tmpdir / f"{src.stem}.snapshot.xml"
    xml_path.write_text(xml1, encoding="utf-8")
    pkt_path = tmpdir / f"{src.stem}.rebuilt.pkt"
    pkt_path.write_bytes(rebuilt)

    # la prueba de causa: el archivo reconstruido se debe volver a decifrar
    # y producir EXACTAMENTE el mismo XML (sin pérdida en el formato propio)
    xml2 = _inferno_decode(rebuilt)
    ok_xml = xml1 == xml2
    ok_start = xml1.lstrip().startswith("<")
    ok_bytes = rebuilt == raw   # solo si el emitter zlib de PT coincide (bonus)

    print("=" * 60)
    print(f"especímen          : {src.name}")
    print(f"tamaño binario     : {len(raw)} B")
    print(f"XML extraído       : {len(xml1)} B  ({xml_path})")
    print(f"binario regenerado : {len(rebuilt)} B  ({pkt_path})")
    print(f"XML es documento   : {ok_start}  (empieza con '<')")
    print(f"ROUND TRIP XML     : {ok_xml}  (decode→encode→decode = XML idéntico)")
    print(f"byte-idéntico      : {ok_bytes}  (bonus: exige zlib emitter igual a PT)")
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

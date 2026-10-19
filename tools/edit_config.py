"""Edición textual segura del snapshot de pkt-git.

Toda edición queda delimitada por device y valida el XML antes de escribir.

Uso:
    edit_config.py <xml> show <DEVICE>
    edit_config.py <xml> setiface <DEVICE> <IFACE> "línea 1" "línea 2" ...
    edit_config.py <xml> chgline <DEVICE> <IFACE> <texto_original> <texto_nuevo>
    edit_config.py <xml> insertafter <DEVICE> <frase> "línea 1" ...
"""

import sys
import xml.etree.ElementTree as ET
from pathlib import Path


def scope_device(xml: str, device: str) -> tuple[int, int]:
    """Segmento XML del device: [posición de su NAME, siguiente NAME)."""
    i = xml.find(f'<NAME translate="true">{device}</NAME>')
    if i < 0:
        raise SystemExit(f"device '{device}' no existe en el snapshot")
    iN = xml.find('<NAME translate="true">', i + 40)
    return i, (iN if iN > 0 else len(xml))


def config_bounds(xml: str, start: int, end: int) -> tuple[int | None, int | None]:
    """Límites del bloque <RUNNINGCONFIG> dentro del segmento dado."""
    ra = xml.find("<RUNNINGCONFIG>", start, end)
    if ra < 0:
        return None, None
    rb = xml.find("</RUNNINGCONFIG>", ra)
    return ra, rb


def block_bounds(seg: str, iface: str) -> tuple[int, int] | None:
    """(inicio, fin) del contenido del bloque interface <iface>."""
    open_pat = f"<LINE>interface {iface}</LINE>"
    idx = seg.find(open_pat)
    if idx < 0:
        return None
    a = idx + len(open_pat) + 1
    b = seg.find("<LINE>!</LINE>", idx)
    return a, b


def block_replace(seg: str, iface: str, lines: list[str]) -> str:
    """Reemplaza todo el contenido del bloque interface <iface>."""
    open_pat = f"<LINE>interface {iface}</LINE>"
    idx = seg.find(open_pat)
    if idx < 0:
        raise SystemExit(f"interface '{iface}' no está en RUNNINGCONFIG del device")
    end_idx = seg.find("<LINE>!</LINE>", idx)
    content = "\n".join(f"      <LINE>{l}</LINE>" for l in lines)
    return seg[:idx] + open_pat + "\n" + content + "\n      " + seg[end_idx:]


def insert_after(seg: str, phrase: str, lines: list[str]) -> str:
    """Inserta líneas tras la primera línea de config que contenga <phrase>."""
    idx = seg.find(phrase)
    if idx < 0:
        raise SystemExit(f"frase '{phrase}' no está en RUNNINGCONFIG del device")
    line_end = seg.find("</LINE>", idx) + len("</LINE>")
    content = "\n" + "\n".join(f"      <LINE>{l}</LINE>" for l in lines)
    return seg[:line_end] + content + seg[line_end:]


def verify(xml: str) -> None:
    """Lanza excepción si el XML quedó inválido."""
    ET.fromstring(xml)


def cmd_show(xml: str, device: str) -> None:
    """Imprime el RUNNINGCONFIG del device (sin líneas sensibles)."""
    s, e = scope_device(xml, device)
    ra, rb = config_bounds(xml, s, e)
    if ra is None:
        print(f"(device '{device}' sin RUNNINGCONFIG)")
        return
    for line in xml[ra:rb].split("\n"):
        if line.strip().startswith("<LINE>") and "PASSWORD" not in line.upper():
            print(line.replace("<LINE>", "").replace("</LINE>", "").strip())


def main() -> int:
    if len(sys.argv) < 3:
        print(__doc__)
        return 2
    xml_path, cmd = Path(sys.argv[1]), sys.argv[2]
    xml = xml_path.read_text(encoding="utf-8")

    if cmd == "show" and len(sys.argv) == 4:
        cmd_show(xml, sys.argv[3])
        return 0

    device = sys.argv[3]
    s, e = scope_device(xml, device)
    ra, rb = config_bounds(xml, s, e)
    if ra is None:
        raise SystemExit(f"'{device}' no tiene RUNNINGCONFIG (¿PC/server?)")

    seg = xml[ra:rb]
    if cmd == "setiface" and len(sys.argv) >= 6:
        seg2 = block_replace(seg, sys.argv[4], sys.argv[5:])
    elif cmd == "chgline" and len(sys.argv) == 7:
        iface, old, new = sys.argv[4], sys.argv[5], sys.argv[6]
        bb = block_bounds(seg, iface)
        if bb is None:
            raise SystemExit(f"interface '{iface}' sin contenido en RUNNINGCONFIG")
        a, b = bb
        old_ct = seg[a:b]
        if old not in old_ct:
            raise SystemExit(f"la línea '{old}' no está en {iface} del device")
        seg2 = seg[:a] + old_ct.replace(old, new, 1) + seg[b:]
    elif cmd == "insertafter" and len(sys.argv) >= 6:
        seg2 = insert_after(seg, sys.argv[4], sys.argv[5:])
    else:
        print(__doc__)
        return 2

    xml2 = xml[:ra] + seg2 + xml[rb:]
    verify(xml2)
    xml_path.write_text(xml2, encoding="utf-8")
    print(f"[ok] {device}: {cmd} aplicado (XML verificado)")
    return 0


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python
"""
edit_config.py — edición textual segura del snapshot de pkt-git

Toda edición está delimitada por DEVICE y valida el XML después de escribir.

  .venv/bin/python tools/edit_config.py <snapshot.xml> show <DEVICE>
  .venv/bin/python tools/edit_config.py <snapshot.xml> setiface <DEVICE> <IFACE> "línea 1" "línea 2" ...
  .venv/bin/python tools/edit_config.py <snapshot.xml> insertafter <DEVICE> <frase> "línea 1" ...

Ejemplos:
  ... setiface SW1_LIMA FastEthernet0/3 " switchport access vlan 40" " switchport mode access"
  ... insertafter CORE_LIBERTAD "ip routing" "router ospf 1" " network 10.30.0.0 0.0.0.255 area 0"
"""

import sys
import xml.etree.ElementTree as ET
from pathlib import Path


def scope_device(xml: str, device: str) -> tuple[int, int]:
    """Devuelve (inicio, fin) del segmento XML del device: [NAME_x, NAME_x+1)."""
    i = xml.find(f'<NAME translate="true">{device}</NAME>')
    if i < 0:
        raise SystemExit(f"❌ device '{device}' no existe en el snapshot")
    iN = xml.find('<NAME translate="true">', i + 40)
    return i, (iN if iN > 0 else len(xml))


def config_bounds(xml: str, start: int, end: int) -> tuple[int | None, int | None]:
    ra = xml.find("<RUNNINGCONFIG>", start, end)
    if ra < 0:
        return None, None
    rb = xml.find("</RUNNINGCONFIG>", ra)
    return ra, rb


def verify(xml: str) -> None:
    ET.fromstring(xml)  # un XML roto aquí aborta ANTES de escribir


def cmd_show(xml: str, device: str) -> None:
    s, e = scope_device(xml, device)
    ra, rb = config_bounds(xml, s, e)
    if ra is None:
        print(f"(device '{device}' sin RUNNINGCONFIG)")
        return
    seg = xml[ra:rb]
    for line in seg.split("\n"):
        if line.strip().startswith("<LINE>") and "PASSWORD" not in line.upper():
            print(line.replace("<LINE>", "").replace("</LINE>", "").strip())


def block_replace(seg: str, iface: str, lines: list[str]) -> str:
    """Reemplaza el contenido del bloque interface <iface> en RUNNINGCONFIG."""
    open_pat = f"<LINE>interface {iface}</LINE>"
    idx = seg.find(open_pat)
    if idx < 0:
        raise SystemExit(f"❌ interface '{iface}' no está en RUNNINGCONFIG del device")
    end_idx = seg.find("<LINE>!</LINE>", idx)
    content = "\n".join(f"      <LINE>{l}</LINE>" for l in lines)
    return (seg[:idx] + open_pat + "\n" + content + "\n      " + seg[end_idx:])


def insert_after(seg: str, phrase: str, lines: list[str]) -> str:
    """Inserta <LINE>s tras la 1ª línea de config que contenga <phrase>."""
    idx = seg.find(phrase)
    if idx < 0:
        raise SystemExit(f"❌ frase '{phrase}' no está en RUNNINGCONFIG del device")
    line_end = seg.find("</LINE>", idx) + len("</LINE>")
    content = "\n" + "\n".join(f"      <LINE>{l}</LINE>" for l in lines)
    return seg[:line_end] + content + seg[line_end:]


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
        raise SystemExit(f"❌ '{device}' no tiene RUNNINGCONFIG (¿PC/server?)")

    seg = xml[ra:rb]
    if cmd == "setiface" and len(sys.argv) >= 6:
        seg2 = block_replace(seg, sys.argv[4], sys.argv[5:])
    elif cmd == "insertafter" and len(sys.argv) >= 6:
        seg2 = insert_after(seg, sys.argv[4], sys.argv[5:])
    else:
        print(__doc__)
        return 2

    xml2 = xml[:ra] + seg2 + xml[rb:]
    verify(xml2)                       # XML válido antes de tocar disco
    xml_path.write_text(xml2, encoding="utf-8")
    print(f"[ok] {device}: {cmd} aplicado ✓ (XML verificado)")
    return 0


if __name__ == "__main__":
    sys.exit(main())

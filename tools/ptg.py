#!/usr/bin/env python
"""
ptg.py — CLI v0.1 del proyecto pkt-git (PASO 2)

La herramienta que convierte un proyecto Packet Tracer en versionable con Git.

    ptg export <proyecto.pkt> <dir_salida>
        .pkt -> base.snapshot.xml (texto versionable)
              + configs por dispositivo (<dir>/configs/<NAME>.cfg)
              + metadata/topology.json (mapa de topología: devices + links)

    ptg build <dir>/base.snapshot.xml <salida.pkt>
        XML -> binario reversible (reconstruye el archivo para PT)

    ptg inspect <proyecto.pkt> [partido]
        resumen rápido en consola (devices/links/modelos)

Uso de referencia:
    .venv/bin/python tools/ptg.py export "proof-pkt/Caso1_Grupo09-2.pkt" snapshot/
"""

import json
import re
import sys
import xml.etree.ElementTree as ET
from pathlib import Path

from spike_roundtrip import _inferno_decode, _inferno_encode

RE_REF = re.compile(r"save-ref-id:(\d+)")


def sanitize(name: str) -> str:
    return re.sub(r"[^\w.-]+", "_", name).strip("._") or "device"


# ----------------------------------------------------------------------------
# parse del XMML de Packet Tracer
# ----------------------------------------------------------------------------

def parse_topology(xml_text: str):
    """Extrae devices (con configs) y links del XML de Packet Tracer."""
    root = ET.fromstring(xml_text)
    net = root.find(".//NETWORK")

    devices = []
    by_ref = {}
    for d in net.findall("DEVICES/DEVICE"):
        eng = d.find("ENGINE")
        if eng is None:
            continue
        tnode = eng.find("TYPE")
        dtype = (tnode.text or "").strip() if tnode is not None else ""
        model = tnode.attrib.get("model", "") if tnode is not None else ""
        nnode = eng.find("NAME")
        name = (nnode.text or "").strip() if nnode is not None else "device"
        refn = eng.find("SAVE_REF_ID")
        ref = refn.text if refn is not None else ""

        cfg_node = eng.find("RUNNINGCONFIG")
        lines = [l.text or "" for l in cfg_node.findall("LINE")] if cfg_node is not None else None
        st_node = eng.find("STARTUPCONFIG")
        slines = [l.text or "" for l in st_node.findall("LINE")] if st_node is not None else None

        dev = {
            "name": name,
            "type": dtype,
            "model": model,
            "ref": ref,
            "runningconfig": lines,
            "startupconfig": slines,
        }
        devices.append(dev)
        if ref:
            by_ref[ref.split(":")[-1]] = dev

    links = []
    for lk in net.findall("LINKS/LINK"):
        cable = lk.find("CABLE")
        if cable is None:
            continue
        def dst(key, pkey):
            node = cable.find(key)
            if node is None:
                return None
            m = RE_REF.match(node.text or "")
            if m is None:
                return None
            from_dev = by_ref.get(m.group(1))
            return {
                "ref": m.group(1),
                "name": from_dev["name"] if from_dev else f"ref:{m.group(1)[-8:]}",
                "port": (cable.find(pkey).text if cable.find(pkey) is not None else ""),
            }

        a, b = dst("FROM", "PORT"), None
        # los LINK traen DOS <PORT>; el primero corresponde a FROM
        ports = [p.text or "" for p in cable.findall("PORT")]
        mfrom = RE_REF.match(cable.find("FROM").text or "") if cable.find("FROM") is not None else None
        mto   = RE_REF.match(cable.find("TO").text or "")   if cable.find("TO")   is not None else None
        da = by_ref.get(mfrom.group(1)) if mfrom else None
        db = by_ref.get(mto.group(1))   if mto   else None
        links.append({
            "link_type": (lk.find("TYPE").text or "") if lk.find("TYPE") is not None else "",
            "from": {"name": da["name"] if da else mfrom.group(1)[-8:] if mfrom else None,
                     "port": ports[0] if ports else ""},
            "to":   {"name": db["name"] if db else mto.group(1)[-8:] if mto else None,
                     "port": ports[1] if len(ports) > 1 else ""},
        })

    return devices, links


# ----------------------------------------------------------------------------
# comandos
# ----------------------------------------------------------------------------

def cmd_export(pkt: Path, outdir: Path) -> int:
    outdir.mkdir(parents=True, exist_ok=True)
    cfgdir = outdir / "configs";  cfgdir.mkdir(exist_ok=True)
    metadir = outdir / "metadata"; metadir.mkdir(exist_ok=True)

    xml_text = _inferno_decode(pkt.read_bytes())
    (outdir / "base.snapshot.xml").write_text(xml_text, encoding="utf-8")
    devices, links = parse_topology(xml_text)

    n_cfg = 0
    for dev in devices:
        if dev["runningconfig"]:
            fname = cfgdir / f"{sanitize(dev['name'])}.cfg"
            fname.write_text("\n".join(dev["runningconfig"]) + "\n", encoding="utf-8")
            n_cfg += 1

    catalogo = {
        "source_file": pkt.name,
        "document_root": "PACKETTRACER5",
        "total_devices": len(devices),
        "total_links": len(links),
        "devices": [{k: v for k, v in d.items() if k.startswith(("name", "type", "model", "ref"))} for d in devices],
        "links": links,
    }
    (metadir / "topology.json").write_text(
        json.dumps(catalogo, indent=2, ensure_ascii=False), encoding="utf-8")

    print(f"[ok] base.snapshot.xml  : {len(xml_text):,} B")
    print(f"[ok] configs .cfg       : {n_cfg} dispositivos con CLI")
    print(f"[ok] metadata/topology.json: {len(devices)} devices, {len(links)} links")
    return 0


def cmd_build(xmlfile: Path, outpkt: Path) -> int:
    xml_text = xmlfile.read_text(encoding="utf-8")
    data = _inferno_encode(xml_text)
    outpkt.write_bytes(data)
    print(f"[ok] {xmlfile} -> {outpkt}  ({len(data):,} B) — listo para abrir en PT")
    return 0


def cmd_inspect(pkt: Path) -> int:
    xml_text = _inferno_decode(pkt.read_bytes())
    devices, links = parse_topology(xml_text)
    print(f"{pkt.name}: {len(devices)} devices, {len(links)} links\n")
    print(f"{'NOMBRE':<28}{'TIPO':<16}{'MODELO':<14}CLI")
    print("-" * 66)
    for d in devices:
        print(f"{d['name'][:27]:<28}{d['type'][:15]:<16}{d['model'][:13]:<14}"
              f"{'✔ ' + str(len(d['runningconfig'])) + ' líneas' if d['runningconfig'] else '—'}")
    return 0


if __name__ == "__main__":
    cmds = {"export": cmd_export, "build": cmd_build, "inspect": cmd_inspect}
    if len(sys.argv) < 2 or sys.argv[1] not in cmds:
        print(__doc__)
        sys.exit(2)
    cmd, *args = sys.argv[1:]
    if cmd == "export" and len(args) == 2:
        sys.exit(cmd_export(Path(args[0]), Path(args[1])))
    if cmd == "build" and len(args) == 2:
        sys.exit(cmd_build(Path(args[0]), Path(args[1])))
    if cmd == "inspect" and len(args) == 1:
        sys.exit(cmd_inspect(Path(args[0])))
    print(__doc__)
    sys.exit(2)

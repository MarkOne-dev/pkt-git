"""pkt-git CLI v0.1: turns Packet Tracer projects into versionable repositories.

Commands:
    export <project.pkt> <dir>   .pkt -> base.snapshot.xml + configs/*.cfg + metadata
    build  <xml> <output.pkt>    XML -> rebuilt binary, ready to open in PT
    inspect <project.pkt>        console summary of devices/links
"""

import json
import re
import sys
import xml.etree.ElementTree as ET
from dataclasses import dataclass, field
from pathlib import Path

from spike_roundtrip import decode as _decode, encode as _encode

RE_REF = re.compile(r"save-ref-id:(\d+)")


@dataclass
class Endpoint:
    """One end of a link: device and physical port."""
    name: str | None
    port: str


@dataclass
class Link:
    """A cable between two topology devices."""
    link_type: str
    src: Endpoint
    dst: Endpoint

    def to_dict(self) -> dict:
        """JSON representation keeping the original 'from'/'to' schema keys."""
        return {
            "link_type": self.link_type,
            "from": {"name": self.src.name, "port": self.src.port},
            "to": {"name": self.dst.name, "port": self.dst.port},
        }


@dataclass
class Device:
    """Topology device with its CLI configurations."""
    name: str
    type: str
    model: str
    ref: str
    runningconfig: list[str] | None = field(default=None)
    startupconfig: list[str] | None = field(default=None)


def sanitize(name: str) -> str:
    """Turn a device name into a safe file name."""
    return re.sub(r"[^\w.-]+", "_", name).strip("._") or "device"


def parse_topology(xml_text: str) -> tuple[list[Device], list[Link]]:
    """Extract devices (with configs) and links from the Packet Tracer XML."""
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

        dev = Device(name=name, type=dtype, model=model, ref=ref,
                     runningconfig=lines, startupconfig=slines)
        devices.append(dev)
        if ref:
            by_ref[ref.split(":")[-1]] = dev

    links = []
    for lk in net.findall("LINKS/LINK"):
        cable = lk.find("CABLE")
        if cable is None:
            continue
        ports = [p.text or "" for p in cable.findall("PORT")]
        mfrom = RE_REF.match(cable.find("FROM").text or "") if cable.find("FROM") is not None else None
        mto = RE_REF.match(cable.find("TO").text or "") if cable.find("TO") is not None else None
        da = by_ref.get(mfrom.group(1)) if mfrom else None
        db = by_ref.get(mto.group(1)) if mto else None
        links.append(Link(
            link_type=(lk.find("TYPE").text or "") if lk.find("TYPE") is not None else "",
            src=Endpoint(name=da.name if da else (mfrom.group(1)[-8:] if mfrom else None),
                         port=ports[0] if ports else ""),
            dst=Endpoint(name=db.name if db else (mto.group(1)[-8:] if mto else None),
                         port=ports[1] if len(ports) > 1 else ""),
        ))

    return devices, links


def cmd_export(pkt: Path, outdir: Path) -> int:
    """Export a .pkt into a textual snapshot: XML + per-device configs + metadata."""
    outdir.mkdir(parents=True, exist_ok=True)
    cfgdir = outdir / "configs"
    cfgdir.mkdir(exist_ok=True)
    metadir = outdir / "metadata"
    metadir.mkdir(exist_ok=True)

    xml_text = _decode(pkt.read_bytes())
    (outdir / "base.snapshot.xml").write_text(xml_text, encoding="utf-8")
    devices, links = parse_topology(xml_text)

    n_cfg = 0
    for dev in devices:
        if dev.runningconfig:
            (cfgdir / f"{sanitize(dev.name)}.cfg").write_text(
                "\n".join(dev.runningconfig) + "\n", encoding="utf-8")
            n_cfg += 1

    catalog = {
        "source_file": pkt.name,
        "document_root": "PACKETTRACER5",
        "total_devices": len(devices),
        "total_links": len(links),
        "devices": [{"name": d.name, "type": d.type, "model": d.model, "ref": d.ref}
                    for d in devices],
        "links": [lk.to_dict() for lk in links],
    }
    (metadir / "topology.json").write_text(
        json.dumps(catalog, indent=2, ensure_ascii=False), encoding="utf-8")

    print(f"[ok] base.snapshot.xml     : {len(xml_text):,} B")
    print(f"[ok] per-device .cfg files : {n_cfg} devices with CLI config")
    print(f"[ok] metadata/topology.json: {len(devices)} devices, {len(links)} links")
    return 0


def cmd_build(xmlfile: Path, outpkt: Path) -> int:
    """Rebuild the binary .pkt from the textual snapshot."""
    data = _encode(xmlfile.read_text(encoding="utf-8"))
    outpkt.write_bytes(data)
    print(f"[ok] {xmlfile} -> {outpkt}  ({len(data):,} B) — ready to open in PT")
    return 0


def cmd_inspect(pkt: Path) -> int:
    """Print a topology summary: devices, types, models and configs."""
    xml_text = _decode(pkt.read_bytes())
    devices, links = parse_topology(xml_text)
    print(f"{pkt.name}: {len(devices)} devices, {len(links)} links\n")
    print(f"{'NAME':<28}{'TYPE':<16}{'MODEL':<14}CLI")
    print("-" * 66)
    for d in devices:
        cli = f"✔ {len(d.runningconfig)} lines" if d.runningconfig else "—"
        print(f"{d.name[:27]:<28}{d.type[:15]:<16}{d.model[:13]:<14}{cli}")
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

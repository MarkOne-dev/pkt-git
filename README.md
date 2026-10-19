# pkt-git

Versionado estilo Git para proyectos de **Cisco Packet Tracer**.

## El problema
Los archivos `.pkt` son binarios opacos: Git no puede hacer diff, no puede hacer
merge, y en un equipo de varias personas se pierde trabajo (todos guardando
encima del mismo archivo o multiplicando `proyecto_final_v3(2).pkt`).

## La solución
Cada `.pkt` es, internamente, un **XML comprimido y cifrado reversible**.
Si cada vez que se guarda el proyecto también existe su **instantánea textual**
(XML + configs + metadata), entonces Git puede:

- mostrar **diffs línea a línea** de la red (cambios de IP, VLANs, routing…),
- hacer **merge / PRs legibles por humanos**,
- conservar **historia completa** y una única fuente de verdad en `main`.

El binario pasa a ser un *artefacto regenerable*; el texto versionable es la
fuente de verdad del repositorio.

## Cómo funciona (lo que ya está hecho)

```
        Packet Tracer (.pkt)                          Git / GitHub
   ┌───────────────────────────┐                ┌─────────────────────┐
   │   binario cifrado         │   ptg export   │ base.snapshot.xml   │
   │   599 KB                  │ ─────────────► │ 5.9 MB de XML       │
   │   (Twofish-EAX + XOR +    │                │ configs/*.cfg       │
   │    zlib + XOR)            │   ptg build    │ (CLI IOS textual)   │
   │   ←─────────────────────  │                │ metadata/*.json     │
   └───────────────────────────┘                └─────────────────────┘
```

## Estado del proyecto

| Paso | Alcance | Estado |
|---|---|---|
| 0 | Decisiones fijas + entorno | ✅ hecho (esta sección) |
| 1 | SPIKE: `.pkt → XML → .pkt` reversible | ✅ PROBADO con archivo real |
| 2 | CLI v0.1: `ptg export / build / inspect` | ✅ hecho |
| 3 | Integración Git (hooks `pre-commit` / `post-merge`) | ⬜ |
| 4 | Workflow colaborativo 3 personas (ramas + merge) | ⬜ |
| 5 | Publicación GitHub (PRs, protección `main`, Action) | ⬜ |

## Decisiones fijas (Paso 0)

- **Versión de Packet Tracer obligatoria:** `9.0.1.0858` (probada). Nadie
  actualiza PT hasta que se valide la nueva versión contra los espírit
  especímenes de este repo.
- **Python:** 3.14.8 en venv propio (`.venv/`). API moderna, sin `imp`/deprecated.
- **Criptografía / formato:** implementación propia (Python puro + ref C de
  Twofish). Nada de dependencia de proyectos ajenos para el core.
- **Estructura del snapshot** (lo que se commitea):
  ```
  base.snapshot.xml     ← documento completo (fuente de verdad textual)
  configs/<NAME>.cfg    ← config CLI por dispositivo (34 en el espécimen)
  metadata/topology.json← mapa devices + links (nombres + puertos)
  ```

## Detalles del formato `.pkt` (documentado por el trabajo propio)

Pipeline de descifrado (4 capas, todas invertibles, verificadas contra PT
9.0.1.0858 con un archivo real de 599 KB):

1. **deobfuscación externa:** `processed[i] = raw[L-1-i] ^ (L - i*L mod 256)` —
   el XOR va por el índice de *salida*, no de entrada.
2. **descifrado:** Twofish modo **EAX** (encrypt-then-MAC: el tag de 16 bytes
   va al final del blob, cubre el ciphertext). Clave `0x89*16`, IV `0x10*16`.
   OMAC/CMAC estándar NIST con prefijo de tweak = bloque completo `00×15||t`.
3. **deobfuscación interna:** `b[i] = a[i] ^ (L2 - i mod 256)` donde L2 es la
   longitud del ciphertext (sin el tag).
4. **qCompress:** 4 bytes big-endian del tamaño sin comprimir + stream zlib.

El emisor de PT genera un stream zlib ligeramente distinto (build diferente de
zlib): el archivo regenerado **no es byte-igual** pero es válido y abre igual.

## Cómo usar (en desarrollo)

```bash
git clone <repo> pkt-git && cd pkt-git
./setup.sh          # venv + dependencias + parche + hooks — UN solo comando
```

Requisitos de la máquina: **Python 3.12+** y **Git**. Packet Tracer **9.0.1.0858**
solo para quien edite la red en PT (quien solo revisa diffs en GitHub no lo necesita).

```bash
.venv/bin/python tools/ptg.py export "mi_proyecto.pkt" snapshot/mi_proyecto/
.venv/bin/python tools/ptg.py build  snapshot/mi_proyecto/base.snapshot.xml regenerado.pkt
.venv/bin/python tools/ptg.py inspect "mi_proyecto.pkt"          # resumen de la red
.venv/bin/python tools/spike_roundtrip.py verify "mi_proyecto.pkt"  # test del pipeline
.venv/bin/python tools/edit_config.py snapshot/mi_proyecto/base.snapshot.xml show SW1_LIMA
```

El flujo diario del equipo es solo: **editar en PT → guardar → `git add` → `git commit`** —
el hook pre-commit genera el snapshot textual automáticamente; tras un `git pull`/merge
el hook post-merge regenera los `.pkt` locales.

## Pendiente de validación humana (Rol del equipo)
- [ ] Abrir `out/reconstruido.pkt` (o el del archivo de prueba) en PT y
      comprobar que la topología es idéntica línea a línea.

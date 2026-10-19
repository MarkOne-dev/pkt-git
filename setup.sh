#!/usr/bin/env bash
# ============================================================================
# setup.sh — instalación única de pkt-git para cualquier miembro del equipo
#
#   git clone <repo> && cd pkt-git && ./setup.sh
#
# Hace: venv + dependencias + parche de compatibilidad Python 3.12+ +
#       hooks de Git activados + verificación del cripto-core.
# Es idempotente: puede correrse las veces que haga falta sin romper nada.
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")"

# ── 0. Requisito: Python 3.12+ ──────────────────────────────────────────────
PY=python3
ver=$($PY -c 'import sys; print(f"{sys.version_info[0]}.{sys.version_info[1]}")')
ok=$($PY -c 'import sys; print(1 if sys.version_info >= (3,12) else 0)')
[[ "$ok" == "1" ]] || { echo "❌ Se requiere Python 3.12+ (tienes $ver)"; exit 1; }
echo "✓ Python $ver"

# ── 1. venv ─────────────────────────────────────────────────────────────────
if [[ ! -d .venv ]]; then
    echo "→ creando .venv ..."
    $PY -m venv .venv
else
    echo "✓ .venv ya existe"
fi

# ── 2. dependencias ─────────────────────────────────────────────────────────
if ! .venv/bin/python -c "import twofish" 2>/dev/null; then
    echo "→ instalando twofish + pycryptodome ..."
    .venv/bin/pip --quiet install twofish pycryptodome
else
    echo "✓ twofish instalado"
fi

# ── 3. parche de compatibilidad (el paquete twofish usa 'imp', muerto en 3.12+) ──
TF=$(find .venv -name twofish.py -path '*/site-packages/*' | head -1)
if [[ -n "$TF" ]] && grep -q "^import imp$" "$TF"; then
    echo "→ parcheando $TF (imp → importlib) ..."
    $PY - "$TF" <<'PYEOF'
import sys, pathlib
p = pathlib.Path(sys.argv[1])
src = p.read_text()
src = src.replace(
    "import imp\nimport sys",
    "# PATCH pkt-git (Python 3.12+): 'imp' fue eliminado; usar importlib\n"
    "import importlib.util, os\nimport sys\n\n"
    "_spec = importlib.util.find_spec('_twofish')",
    1,
)
src = src.replace(
    "_twofish = cdll.LoadLibrary(imp.find_module('_twofish')[1])",
    "_twofish = cdll.LoadLibrary(\n"
    "    _spec.origin if _spec and _spec.origin\n"
    "    else os.path.join(os.path.dirname(__file__), '_twofish.so')\n"
    ")",
    1,
)
p.write_text(src)
print("  parche aplicado ✓")
PYEOF
else
    echo "✓ parche ya aplicado (o no necesario)"
fi

# ── 4. hooks de Git (solo si estamos dentro de un repo) ─────────────────────
if git rev-parse --git-dir >/dev/null 2>&1; then
    git config core.hooksPath .githooks
    echo "✓ hooks activados (core.hooksPath=.githooks)"
else
    echo "⚠ no es un repo git todavía — corre 'git init' y luego ./setup.sh de nuevo"
fi

# ── 5. verificación del cripto-core (vector de test Twofish ECB) ────────────
.venv/bin/python - <<'PYEOF' || { echo "❌ cripto-core NO verificado"; exit 1; }
from twofish import Twofish
t = Twofish(b"\x89" * 16)
assert t.decrypt(t.encrypt(bytes(16))) == bytes(16)
print("✓ cripto-core verificado (Twofish ECB round-trip)")
PYEOF

echo
echo "════════════════════════════════════════════════════"
echo " pkt-git listo. Uso:"
echo "   .venv/bin/python tools/ptg.py export <proyecto.pkt> snapshot/<nombre>/"
echo "   .venv/bin/python tools/ptg.py build  snapshot/<nombre>/base.snapshot.xml <salida.pkt>"
echo "   git commit  ← el hook adjunta el snapshot textual automáticamente"
echo "════════════════════════════════════════════════════"

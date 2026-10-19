#!/usr/bin/env bash
# setup.sh — instalación única de pkt-git (idempotente)
set -euo pipefail
cd "$(dirname "$0")"

PY=python3
ver=$($PY -c 'import sys; print(f"{sys.version_info[0]}.{sys.version_info[1]}")')
ok=$($PY -c 'import sys; print(1 if sys.version_info >= (3,12) else 0)')
[[ "$ok" == "1" ]] || { echo "Se requiere Python 3.12+ (tienes $ver)"; exit 1; }
echo "Python $ver ✓"

[[ -d .venv ]] || $PY -m venv .venv
.venv/bin/python -c "import twofish" 2>/dev/null \
  || .venv/bin/pip --quiet install twofish pycryptodome

# parche: el paquete twofish usa 'imp', eliminado en Python 3.12+
TF=$(find .venv -name twofish.py -path '*/site-packages/*' | head -1)
if [[ -n "$TF" ]] && grep -q "^import imp$" "$TF"; then
    $PY - "$TF" <<'PYEOF'
import sys, pathlib
p = pathlib.Path(sys.argv[1])
src = p.read_text()
src = src.replace(
    "import imp\nimport sys",
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
PYEOF
    echo "parche twofish aplicado ✓"
fi

if git rev-parse --git-dir >/dev/null 2>&1; then
    git config core.hooksPath .githooks
    echo "hooks activados ✓"
else
    echo "aviso: no es un repo git — corre 'git init' y repite ./setup.sh"
fi

.venv/bin/python -c "
from twofish import Twofish
t = Twofish(b'\x89' * 16)
assert t.decrypt(t.encrypt(bytes(16))) == bytes(16)
" && echo "cripto-core verificado ✓"

#!/usr/bin/env bash
# setup.sh — one-command pkt-git installation (idempotent)
set -euo pipefail
cd "$(dirname "$0")"

# interpreter detection: python3 (POSIX), python / py -3 (Windows)
PY3=""
for cand in "python3" "python" "py -3"; do
    if $cand -c 'import sys; print(1 if sys.version_info >= (3,12) else 0)' 2>/dev/null | grep -q 1; then
        PY3="$cand"
        break
    fi
done
[[ -n "$PY3" ]] || { echo "Python 3.12+ required (python3, python or py -3)"; exit 1; }
$PY3 -c 'import sys; print(f"Python {sys.version_info[0]}.{sys.version_info[1]} ✓")'

[[ -d .venv ]] || $PY3 -m venv .venv

# venv python path differs between platforms
if [[ -f .venv/bin/python ]]; then VPY=.venv/bin/python
elif [[ -f .venv/Scripts/python.exe ]]; then VPY=.venv/Scripts/python.exe
else echo "venv creation failed"; exit 1; fi

$VPY -c "import twofish" 2>/dev/null || $VPY -m pip --quiet install twofish pycryptodome

# patch: the twofish package uses 'imp', removed in Python 3.12+
TF=$(find .venv -name twofish.py -path '*/site-packages/*' | head -1)
if [[ -n "$TF" ]] && grep -q "^import imp$" "$TF"; then
    $VPY - "$TF" <<'PYEOF'
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
    echo "twofish patch applied ✓"
fi

if git rev-parse --git-dir >/dev/null 2>&1; then
    git config core.hooksPath .githooks
    echo "hooks activated ✓"
else
    echo "warning: not a git repo yet — run 'git init' and re-run ./setup.sh"
fi

$VPY -c "
from twofish import Twofish
t = Twofish(b'\x89' * 16)
assert t.decrypt(t.encrypt(bytes(16))) == bytes(16)
" && echo "crypto core verified ✓"

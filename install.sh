#!/usr/bin/env bash
# install.sh — install pkt-git globally on this machine (Linux / macOS / Git Bash).
# Repositories are then wired with a single `ptg init` — no templates to copy.
set -euo pipefail
cd "$(dirname "$0")"

ROOT="${HOME}/.local/share/pkt-git"
BIN_DIR="${HOME}/.local/bin"

# interpreter detection: python3 (POSIX) or python / py -3 (Windows)
PY3=""
for cand in "python3" "python" "py -3"; do
    if $cand -c 'import sys; print(1 if sys.version_info >= (3,12) else 0)' 2>/dev/null | grep -q 1; then
        PY3="$cand"; break
    fi
done
[[ -n "$PY3" ]] || { echo "Python 3.12+ required (python3, python or py -3)"; exit 1; }

mkdir -p "$ROOT/tools" "$BIN_DIR"

# tool source -> global location
cp -a tools/. "$ROOT/tools/"

# global venv + dependencies
[[ -d "$ROOT/venv" ]] || $PY3 -m venv "$ROOT/venv"
VPY="$ROOT/venv/bin/python"
[[ -f "$VPY" ]] || VPY="$ROOT/venv/Scripts/python.exe"

"$VPY" -c "import twofish" 2>/dev/null || "$VPY" -m pip --quiet install twofish pycryptodome

# patch: the twofish package uses 'imp', removed in Python 3.12+
TF=$(find "$ROOT/venv" -name twofish.py -path '*/site-packages/*' | head -1)
if [[ -n "$TF" ]] && grep -q "^import imp$" "$TF"; then
    "$VPY" - "$TF" <<'PYEOF'
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
    echo "twofish patch applied"
fi

# CLI wrapper on PATH (bash) — hardcodes the resolved locations
cat > "$BIN_DIR/ptg" <<EOF
#!/usr/bin/env bash
export PKTG_CLI="$BIN_DIR/ptg"
export PKTG_PY="$VPY"
export PKTG_SCRIPT="$ROOT/tools/cli.py"
exec "$VPY" "$ROOT/tools/cli.py" "\$@"
EOF
chmod +x "$BIN_DIR/ptg"

case ":$PATH:" in
    *":$BIN_DIR:"*) ;;
    *) echo "note: $BIN_DIR is not on PATH — add 'export PATH=\"\$PATH:$BIN_DIR\"' to your shell profile" ;;
esac

# crypto self-test
"$VPY" -c "
from twofish import Twofish
t = Twofish(b'\x89' * 16)
assert t.decrypt(t.encrypt(bytes(16))) == bytes(16)
" && echo "crypto core verified"

echo "pkt-git installed globally. In any repository run:"
echo "    ptg init          # wire hooks and configuration"
echo "    ptg --help        # full command list"

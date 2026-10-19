# setup.ps1 — one-command pkt-git installation for Windows (idempotent)
# Run from the repo root:  powershell -ExecutionPolicy Bypass -File setup.ps1

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

# interpreter detection: py -3 (official launcher) or python
$py = $null
foreach ($cand in @(@("py", "-3"), @("python"))) {
    try {
        $v = & $cand[0] $cand[1] -c "import sys; print(1 if sys.version_info >= (3,12) else 0)" 2>$null
        if ("$v".Trim() -eq "1") { $py = $cand; break }
    } catch { }
}
if (-not $py) { Write-Host "Python 3.12+ required (py -3 or python)"; exit 1 }
& $py[0] $py[1] -c "import sys; print(f'Python {sys.version_info[0]}.{sys.version_info[1]} OK')"

if (-not (Test-Path .venv)) { & $py[0] $py[1] -m venv .venv }
$VPY = ".venv\Scripts\python.exe"
if (-not (Test-Path $VPY)) { Write-Host "venv creation failed"; exit 1 }

& $VPY -c "import twofish" 2>$null
if ($LASTEXITCODE -ne 0) { & $VPY -m pip --quiet install twofish pycryptodome }

# patch: the twofish package uses 'imp', removed in Python 3.12+
$tf = Get-ChildItem -Recurse .venv -Filter twofish.py |
      Where-Object { $_.FullName -like "*site-packages*" } | Select-Object -First 1
if ($tf -and (Select-String -Path $tf.FullName -Pattern "^import imp$" -Quiet)) {
    $patch = @'
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
print("twofish patch applied OK")
'@
    $patch | & $VPY - $tf.FullName
}

git config core.hooksPath .githooks
Write-Host "hooks activated OK"

& $VPY -c "
from twofish import Twofish
t = Twofish(b'\x89' * 16)
assert t.decrypt(t.encrypt(bytes(16))) == bytes(16)
"
if ($LASTEXITCODE -eq 0) { Write-Host "crypto core verified OK" }

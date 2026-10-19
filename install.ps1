# install.ps1 — install pkt-git globally on this machine (Windows, PowerShell).
# Repositories are then wired with a single `ptg init` — no templates to copy.
# Run:  powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"
$root = Join-Path $env:LOCALAPPDATA "pkt-git"
$binDir = Join-Path $HOME ".local\bin"
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

New-Item -ItemType Directory -Force -Path (Join-Path $root "tools") | Out-Null
New-Item -ItemType Directory -Force -Path $binDir | Out-Null

# tool source -> global location
Copy-Item -Path "tools\*" -Destination (Join-Path $root "tools") -Recurse -Force

# global venv + dependencies
$vpy = Join-Path $root "venv\Scripts\python.exe"
if (-not (Test-Path $vpy)) {
    & $py[0] $py[1] -m venv (Join-Path $root "venv")
}
if (-not (Test-Path $vpy)) { Write-Host "venv creation failed"; exit 1 }

& $vpy -c "import twofish" 2>$null
if ($LASTEXITCODE -ne 0) { & $vpy -m pip --quiet install twofish pycryptodome }

# patch: the twofish package uses 'imp', removed in Python 3.12+
$tf = Get-ChildItem -Recurse $root -Filter twofish.py |
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
print("twofish patch applied")
'@
    $patch | & $vpy - $tf.FullName
}

# cmd shims on PATH — primary name `pkt`, `ptg` stays as a legacy alias.
# The thin hooks call the repository config pktgit.py directly;
# these shims exist for human use in cmd/PowerShell.
foreach ($name in @("pkt", "ptg")) {
$shim = Join-Path $binDir "$name.cmd"
@"
@echo off
set "PKTG_CLI=%~f0"
set "PKTG_PY=$vpy"
set "PKTG_SCRIPT=$root\tools\cli.py"
"$vpy" "$root\tools\cli.py" %*
"@ | Set-Content -Path $shim -Encoding Ascii
}

# add to user PATH once
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (($userPath -split ";") -notcontains $binDir) {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$binDir", "User")
    Write-Host "added $binDir to user PATH (open a new terminal)"
}

# crypto self-test
& $vpy -c (@'
from twofish import Twofish
t = Twofish(b'\x89' * 16)
assert t.decrypt(t.encrypt(bytes(16))) == bytes(16)
'@)
if ($LASTEXITCODE -eq 0) { Write-Host "crypto core verified" }

# agent skills (OpenCode / Claude Code / Gemini CLI) into their global dirs
& $vpy (Join-Path $root "tools\cli.py") skills

Write-Host "pkt-git installed globally. In any repository run:"
Write-Host "    ptg init          # wire hooks and configuration"
Write-Host "    ptg edit ...      # available in the tool command list"

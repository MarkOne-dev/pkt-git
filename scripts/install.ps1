#Requires -Version 5.1
<#
.SYNOPSIS
Installs the pkt-git Go binary on Windows.

.DESCRIPTION
Copies a prebuilt pkt.exe (or builds one with Go) into a per-user binary
directory, adds that directory to the user PATH when needed, and installs the
bundled agent skill. Safe to run repeatedly.
#>
[CmdletBinding()]
param(
    [string]$BinDir
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent $scriptDir
$buildDir = $null
$tmpTarget = $null

function Test-PathEntry {
    param(
        [string]$PathValue,
        [string]$Entry
    )

    if ([string]::IsNullOrWhiteSpace($PathValue)) {
        return $false
    }
    $normalizedEntry = $Entry.TrimEnd([char[]]@('\', '/'))
    foreach ($part in ($PathValue -split ';')) {
        if ([string]::IsNullOrWhiteSpace($part)) {
            continue
        }
        if ($part.TrimEnd([char[]]@('\', '/')) -ieq $normalizedEntry) {
            return $true
        }
    }
    return $false
}

function Install-GitIfMissing {
    if (Get-Command git -ErrorAction SilentlyContinue) {
        return
    }
    Write-Host "pkt-git: Git is not installed. Attempting automatic installation via winget..."
    if (Get-Command winget -ErrorAction SilentlyContinue) {
        try {
            winget install --id Git.Git -e --source winget --accept-source-agreements --accept-package-agreements
        }
        catch {
            Write-Warning "winget failed to install Git. Please install Git manually from https://git-scm.com/."
        }
    }
    else {
        Write-Warning "winget is not available. Please install Git manually from https://git-scm.com/."
    }
}

function Configure-GitIdentityIfMissing {
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
        return
    }
    $userName = & git config user.name 2>$null
    $userEmail = & git config user.email 2>$null

    if ([string]::IsNullOrWhiteSpace($userName) -or [string]::IsNullOrWhiteSpace($userEmail)) {
        if ([Environment]::UserInteractive -and $env:CI -ne "1" -and $env:PKTG_NONINTERACTIVE -ne "1") {
            Write-Host "`nGit identity (user.name / user.email) is not configured."
            Write-Host "pkt-git requires a configured Git identity to prevent commit errors."
            if ([string]::IsNullOrWhiteSpace($userName)) {
                $userName = Read-Host "Enter your full name for Git commits"
                if (-not [string]::IsNullOrWhiteSpace($userName)) {
                    & git config --global user.name $userName
                }
            }
            if ([string]::IsNullOrWhiteSpace($userEmail)) {
                $userEmail = Read-Host "Enter your email for Git commits"
                if (-not [string]::IsNullOrWhiteSpace($userEmail)) {
                    & git config --global user.email $userEmail
                }
            }
        }
    }
}

function Install-NerdFontIfMissing {
    if ($env:PKTG_SKIP_FONTS -eq "1") {
        return
    }

    # Check if a Nerd Font is already installed
    $fontInstalled = $false
    $userFontKey = "HKCU:\Software\Microsoft\Windows NT\CurrentVersion\Fonts"
    $sysFontKey = "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts"

    foreach ($keyPath in @($userFontKey, $sysFontKey)) {
        if (Test-Path $keyPath) {
            $props = (Get-ItemProperty -Path $keyPath -ErrorAction SilentlyContinue).PSObject.Properties
            foreach ($p in $props) {
                if ($p.Name -like "*Nerd Font*" -or $p.Value -like "*Nerd Font*") {
                    $fontInstalled = $true
                    break
                }
            }
        }
        if ($fontInstalled) { break }
    }

    $userFontsDir = Join-Path $env:LOCALAPPDATA "Microsoft\Windows\Fonts"
    if (-not $fontInstalled -and (Test-Path $userFontsDir)) {
        if (Get-ChildItem -Path $userFontsDir -Filter "*Nerd*" -ErrorAction SilentlyContinue) {
            $fontInstalled = $true
        }
    }

    if ($fontInstalled) {
        return
    }

    Write-Host "pkt-git: Installing JetBrainsMono Nerd Font..."

    $installedViaWinget = $false
    if (Get-Command winget -ErrorAction SilentlyContinue) {
        try {
            & winget install --id "DEVCOM.JetBrainsMonoNerdFont" -e --source winget --accept-source-agreements --accept-package-agreements
            if ($LASTEXITCODE -eq 0) {
                $installedViaWinget = $true
            }
        }
        catch {
            # fallback to direct download
        }
    }

    if (-not $installedViaWinget) {
        $fontZipUrl = "https://github.com/ryanoasis/nerd-fonts/releases/latest/download/JetBrainsMono.zip"
        $tmpFontZip = Join-Path ([System.IO.Path]::GetTempPath()) ("JetBrainsMono-" + [guid]::NewGuid().ToString("N") + ".zip")
        $tmpFontDir = Join-Path ([System.IO.Path]::GetTempPath()) ("JetBrainsMono-" + [guid]::NewGuid().ToString("N"))

        try {
            [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
            Invoke-WebRequest -Uri $fontZipUrl -OutFile $tmpFontZip -UseBasicParsing -TimeoutSec 60
            Expand-Archive -LiteralPath $tmpFontZip -DestinationPath $tmpFontDir -Force

            New-Item -ItemType Directory -Path $userFontsDir -Force | Out-Null
            $ttfFiles = Get-ChildItem -Path $tmpFontDir -Filter "*.ttf" -Recurse

            foreach ($file in $ttfFiles) {
                $dest = Join-Path $userFontsDir $file.Name
                Copy-Item -LiteralPath $file.FullName -Destination $dest -Force
                $regName = [System.IO.Path]::GetFileNameWithoutExtension($file.Name) + " (TrueType)"
                New-ItemProperty -Path $userFontKey -Name $regName -Value $dest -PropertyType String -Force | Out-Null
            }

            try {
                Add-Type -TypeDefinition @"
                using System;
                using System.Runtime.InteropServices;
                public class FontBroadcaster {
                    [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
                    public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
                }
"@ -ErrorAction SilentlyContinue
                $HWND_BROADCAST = [IntPtr]0xffff
                $WM_FONTCHANGE = 0x001D
                $SMTO_ABORTIFHUNG = 0x0002
                $res = [UIntPtr]::Zero
                [FontBroadcaster]::SendMessageTimeout($HWND_BROADCAST, $WM_FONTCHANGE, [UIntPtr]::Zero, $null, $SMTO_ABORTIFHUNG, 1000, [ref]$res) | Out-Null
            }
            catch {}

            Write-Host "pkt-git: JetBrainsMono Nerd Font installed successfully."
        }
        catch {
            Write-Warning "Could not download JetBrainsMono Nerd Font automatically. You can install it manually from https://www.nerdfonts.com."
        }
        finally {
            if (Test-Path $tmpFontZip) { Remove-Item -LiteralPath $tmpFontZip -Force }
            if (Test-Path $tmpFontDir) { Remove-Item -LiteralPath $tmpFontDir -Recurse -Force }
        }
    }

    # Configure Windows Terminal default font if present
    $wtSettings = Join-Path $env:LOCALAPPDATA "Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\settings.json"
    if (Test-Path -LiteralPath $wtSettings) {
        try {
            $jsonContent = Get-Content -LiteralPath $wtSettings -Raw | ConvertFrom-Json
            if ($jsonContent.profiles -and $jsonContent.profiles.defaults) {
                if (-not $jsonContent.profiles.defaults.font) {
                    $jsonContent.profiles.defaults | Add-Member -MemberType NoteProperty -Name "font" -Value ([PSCustomObject]@{})
                }
                $jsonContent.profiles.defaults.font.face = "JetBrainsMono Nerd Font"
                $jsonContent | ConvertTo-Json -Depth 32 | Set-Content -LiteralPath $wtSettings -Encoding UTF8
                Write-Host "Configured Windows Terminal default font to JetBrainsMono Nerd Font."
            }
        }
        catch {}
    }
}

try {
    Install-GitIfMissing
    Configure-GitIdentityIfMissing
    Install-NerdFontIfMissing
    if (-not $BinDir) {
        if ($env:PKTG_BIN_DIR) {
            $BinDir = $env:PKTG_BIN_DIR
        }
        else {
            $BinDir = Join-Path $HOME ".local\bin"
        }
    }
    $target = Join-Path $BinDir "pkt.exe"

    $source = $env:PKTG_SOURCE
    if (-not $source) {
        foreach ($candidate in @((Join-Path $rootDir "bin\pkt.exe"), (Join-Path $rootDir "pkt.exe"))) {
            if (Test-Path -LiteralPath $candidate -PathType Leaf) {
                $source = $candidate
                break
            }
        }
    }

    if (-not $source) {
        if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
            throw "No prebuilt pkt.exe found and Go is not installed. Build with 'make build' or install a release archive."
        }
        $buildDir = Join-Path ([System.IO.Path]::GetTempPath()) ("pkt-git-build-" + [guid]::NewGuid().ToString("N"))
        New-Item -ItemType Directory -Path $buildDir -Force | Out-Null
        $buildTarget = Join-Path $buildDir "pkt.exe"
        $previousCgo = $env:CGO_ENABLED
        Push-Location $rootDir
        try {
            $env:CGO_ENABLED = "0"
            & go build -o $buildTarget ./cmd/pkt
            if ($LASTEXITCODE -ne 0) {
                throw "go build failed with exit code $LASTEXITCODE."
            }
        }
        finally {
            if ($null -eq $previousCgo) {
                Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
            }
            else {
                $env:CGO_ENABLED = $previousCgo
            }
            Pop-Location
        }
        $source = $buildTarget
    }

    if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {
        throw "pkt-git binary not found: $source"
    }

    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
    $changed = $true
    if (Test-Path -LiteralPath $target -PathType Leaf) {
        $sourceHash = (Get-FileHash -LiteralPath $source -Algorithm SHA256).Hash
        $targetHash = (Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash
        if ($sourceHash -eq $targetHash) {
            $changed = $false
        }
    }

    if ($changed) {
        $tmpTarget = Join-Path $BinDir (".pkt-" + [guid]::NewGuid().ToString("N") + ".tmp")
        Copy-Item -LiteralPath $source -Destination $tmpTarget
        if (Test-Path -LiteralPath $target) {
            Remove-Item -LiteralPath $target -Force
        }
        Move-Item -LiteralPath $tmpTarget -Destination $target
        $tmpTarget = $null
    }

    if ($env:PKTG_SKIP_PATH -ne "1") {
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if (-not (Test-PathEntry $userPath $BinDir)) {
            $newPath = if ([string]::IsNullOrWhiteSpace($userPath)) {
                $BinDir
            }
            else {
                $userPath.TrimEnd(';') + ';' + $BinDir
            }
            [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
            Write-Host "Added $BinDir to the user PATH. Open a new terminal to use it."
        }
        if (-not (Test-PathEntry $env:PATH $BinDir)) {
            $env:PATH = $env:PATH.TrimEnd(';') + ';' + $BinDir
        }
    }

    & $target version
    if ($LASTEXITCODE -ne 0) {
        throw "pkt version failed with exit code $LASTEXITCODE."
    }

    if ($env:PKTG_SKIP_SKILLS -ne "1") {
        & $target skills
        if ($LASTEXITCODE -ne 0) {
            throw "pkt skills failed with exit code $LASTEXITCODE."
        }
    }

    $state = if ($changed) { "installed" } else { "unchanged" }
    Write-Host ""
    & $target welcome

    Write-Host "`npkt-git $state at $target."
}
finally {
    if ($tmpTarget -and (Test-Path -LiteralPath $tmpTarget)) {
        Remove-Item -LiteralPath $tmpTarget -Force
    }
    if ($buildDir -and (Test-Path -LiteralPath $buildDir)) {
        Remove-Item -LiteralPath $buildDir -Recurse -Force
    }
}

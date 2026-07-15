param(
    [string]$OutputPath = "dist/itsmyto-windows-amd64.exe"
)

$ErrorActionPreference = "Stop"

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$outputFile = if ([System.IO.Path]::IsPathRooted($OutputPath)) {
    [System.IO.Path]::GetFullPath($OutputPath)
} else {
    [System.IO.Path]::GetFullPath((Join-Path $repoRoot $OutputPath))
}
$outputDirectory = Split-Path -Parent $outputFile
New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

Push-Location $repoRoot
try {
    & go build -mod=vendor -trimpath -ldflags="-s -w -H=windowsgui" -o $outputFile .
    if ($LASTEXITCODE -ne 0) {
        throw "Windows build failed with exit code $LASTEXITCODE"
    }
} finally {
    Pop-Location
}

$bytes = [System.IO.File]::ReadAllBytes($outputFile)
if ($bytes.Length -lt 256 -or $bytes[0] -ne 0x4d -or $bytes[1] -ne 0x5a) {
    throw "Built file is not a valid PE executable: $outputFile"
}

$peHeaderOffset = [System.BitConverter]::ToInt32($bytes, 0x3c)
$optionalHeaderOffset = $peHeaderOffset + 24
$subsystem = [System.BitConverter]::ToUInt16($bytes, $optionalHeaderOffset + 68)
if ($subsystem -ne 2) {
    throw "Expected IMAGE_SUBSYSTEM_WINDOWS_GUI (2), got $subsystem"
}

$hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $outputFile).Hash.ToLowerInvariant()
Write-Host "Built $outputFile"
Write-Host "PE subsystem: IMAGE_SUBSYSTEM_WINDOWS_GUI (2)"
Write-Host "SHA256: $hash"

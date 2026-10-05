# Builds and publishes a GitHub release. Every release always ships both
# widget-stats.exe and widget_control.lua (the updater refreshes both).
# Usage: .\scripts\release.ps1 -Version v2.6.0 -Title "v2.6.0 — ..." [-NotesFile notes.md]
param(
  [Parameter(Mandatory = $true)][string]$Version,
  [Parameter(Mandatory = $true)][string]$Title,
  [string]$NotesFile = ""
)
$ErrorActionPreference = "Stop"
Set-Location (Split-Path $PSScriptRoot -Parent)
if ($Version -notmatch '^v') { $Version = "v$Version" }

$lua = ".\obs\widget_control.lua"
if (-not (Test-Path $lua)) { throw "obs/widget_control.lua is missing — every release must ship it (git restore obs/widget_control.lua)" }

$env:WIDGET_STATS_VERSION = $Version
.\scripts\build.ps1
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

go test ./...
if ($LASTEXITCODE -ne 0) { throw "tests failed — not releasing" }

$assets = @(".\widget-stats.exe", $lua)
$notesArgs = if ($NotesFile) { @("--notes-file", $NotesFile) } else { @("--generate-notes") }
gh release create $Version @assets --title $Title @notesArgs
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "Released $Version with:" ($assets -join ", ")

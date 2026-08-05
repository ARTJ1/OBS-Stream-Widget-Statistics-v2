$ErrorActionPreference = "Stop"
Set-Location (Split-Path $PSScriptRoot -Parent)

Write-Host "Embedding Windows icon..."
$env:PATH = "$(go env GOPATH)\bin;$env:PATH"
if (Get-Command rsrc -ErrorAction SilentlyContinue) {
  rsrc -ico .\assets\app.ico -o .\cmd\widget-stats\rsrc_windows.syso
} else {
  Write-Host "rsrc not found (optional). Install: go install github.com/akavel/rsrc@latest"
}

Write-Host "Building widget-stats.exe (windowsgui, no console)..."
# -H windowsgui: no console flash if something launches the exe from a hotkey path.
go build -ldflags "-H windowsgui" -o widget-stats.exe ./cmd/widget-stats
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "OK:" (Resolve-Path .\widget-stats.exe)
Write-Host "Run the exe, then open http://127.0.0.1:19123/admin/"

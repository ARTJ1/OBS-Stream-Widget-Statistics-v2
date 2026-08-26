# Fetches Tesseract OCR into internal/owtracker/assets/tesseract for go:embed (build-time only).
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
$Dest = Join-Path $Root "internal\owtracker\assets\tesseract"
$Exe = Join-Path $Dest "tesseract.exe"

function Has-ValidTesseract($path) {
  return (Test-Path $path) -and ((Get-Item $path).Length -gt 500000)
}

if (Has-ValidTesseract $Exe) {
  Write-Host "Tesseract already present at $Dest"
  exit 0
}

if (Test-Path $Dest) {
  Remove-Item $Dest -Recurse -Force
}

$installer = Join-Path $env:TEMP "tesseract-ocr-setup.exe"
$staging = Join-Path $env:TEMP "tesseract-embed-staging"
$downloaded = $false

Write-Host "Downloading Tesseract for embed..."
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/UB-Mannheim/tesseract/releases/latest" -Headers @{ "User-Agent" = "widget-stats-build" }
$asset = $release.assets | Where-Object { $_.name -like "tesseract-ocr-w64-setup-*.exe" } | Select-Object -First 1
if ($asset) {
  Write-Host "  $($asset.browser_download_url)"
  Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $installer -Headers @{ "User-Agent" = "widget-stats-build" }
  if ((Get-Item $installer).Length -gt 1MB) { $downloaded = $true }
}

if (-not $downloaded) {
  $fallback = "https://github.com/UB-Mannheim/tesseract/releases/download/v5.4.0.20240606/tesseract-ocr-w64-setup-5.4.0.20240606.exe"
  Write-Host "  fallback $fallback"
  Invoke-WebRequest -Uri $fallback -OutFile $installer -Headers @{ "User-Agent" = "widget-stats-build" }
  if ((Get-Item $installer).Length -gt 1MB) { $downloaded = $true }
}

if (-not $downloaded) {
  $pf = "${env:ProgramFiles}\Tesseract-OCR"
  if (Has-ValidTesseract (Join-Path $pf "tesseract.exe")) {
    Write-Host "Using installed Tesseract from $pf"
    New-Item -ItemType Directory -Force -Path $Dest | Out-Null
    Copy-Item (Join-Path $pf "tesseract.exe") $Dest -Force
    Copy-Item (Join-Path $pf "*.dll") $Dest -Force
    $tessdata = Join-Path $Dest "tessdata"
    New-Item -ItemType Directory -Force -Path $tessdata | Out-Null
    Copy-Item (Join-Path $pf "tessdata\eng.traineddata") $tessdata -Force -ErrorAction SilentlyContinue
    Copy-Item (Join-Path $pf "tessdata\rus.traineddata") $tessdata -Force -ErrorAction SilentlyContinue
    exit 0
  }
  Write-Error "Could not download or find Tesseract for embed"
  exit 1
}

if (Test-Path $staging) { Remove-Item $staging -Recurse -Force }
New-Item -ItemType Directory -Force -Path $staging | Out-Null

Write-Host "Extracting Tesseract to staging..."
Start-Process -FilePath $installer -ArgumentList "/S", "/D=$staging" -Wait

if (-not (Has-ValidTesseract (Join-Path $staging "tesseract.exe"))) {
  $pf = Join-Path ${env:ProgramFiles} "Tesseract-OCR"
  if (Has-ValidTesseract (Join-Path $pf "tesseract.exe")) {
    Write-Host "Silent install used default path, copying from $pf"
    $staging = $pf
  }
}

if (-not (Has-ValidTesseract (Join-Path $staging "tesseract.exe"))) {
  Write-Error "Tesseract install failed - tesseract.exe not found in $staging"
  exit 1
}

New-Item -ItemType Directory -Force -Path $Dest | Out-Null
Write-Host "Copying into $Dest ..."
Copy-Item (Join-Path $staging "tesseract.exe") $Dest -Force
Copy-Item (Join-Path $staging "*.dll") $Dest -Force
$tessdata = Join-Path $Dest "tessdata"
New-Item -ItemType Directory -Force -Path $tessdata | Out-Null
$srcTess = Join-Path $staging "tessdata"
if (Test-Path $srcTess) {
  Copy-Item (Join-Path $srcTess "eng.traineddata") $tessdata -Force -ErrorAction SilentlyContinue
  Copy-Item (Join-Path $srcTess "rus.traineddata") $tessdata -Force -ErrorAction SilentlyContinue
}

function Ensure-TessData($name) {
  $path = Join-Path $tessdata $name
  if (-not (Test-Path $path)) {
    $url = "https://github.com/tesseract-ocr/tessdata/raw/main/$name"
    Write-Host "Downloading $name ..."
    Invoke-WebRequest -Uri $url -OutFile $path -Headers @{ "User-Agent" = "widget-stats-build" }
  }
}

Ensure-TessData "eng.traineddata"
Ensure-TessData "rus.traineddata"

Get-ChildItem $tessdata -Filter "*.traineddata" | Where-Object { $_.Name -notin @("eng.traineddata", "rus.traineddata") } | Remove-Item -Force

Write-Host "OK: embedded OCR assets ready ($Dest)"
Get-ChildItem $Dest -Recurse | Select-Object FullName, Length

# Windows OCR server for widget-stats (built into Windows 10/11, no downloads).
# Protocol on stdin/stdout, one line each:
#   <- LANGS <tag,tag,...>          once at startup
#   -> REQ <id> <absolute png path>
#   <- RES <id> <lang> <base64 utf8 text>   one per OCR language
#   <- ERR <id> <base64 utf8 message>
#   <- END <id>
$ErrorActionPreference = 'Stop'
[Console]::InputEncoding = New-Object System.Text.UTF8Encoding $false
[Console]::OutputEncoding = New-Object System.Text.UTF8Encoding $false

Add-Type -AssemblyName System.Runtime.WindowsRuntime
$null = [Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType = WindowsRuntime]
$null = [Windows.Graphics.Imaging.BitmapDecoder, Windows.Graphics, ContentType = WindowsRuntime]
$null = [Windows.Storage.StorageFile, Windows.Storage, ContentType = WindowsRuntime]

$asTask = ([System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object {
    $_.Name -eq 'AsTask' -and $_.GetParameters().Count -eq 1 -and
    $_.GetParameters()[0].ParameterType.Name -eq 'IAsyncOperation`1' })[0]
function Await($op, [Type]$t) {
  $task = $asTask.MakeGenericMethod($t).Invoke($null, @($op))
  $task.Wait(-1) | Out-Null
  $task.Result
}
function B64([string]$s) { [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($s)) }

# Only languages whose banner words we know (see banner_words.go).
$wanted = @('ru', 'en', 'de', 'fr', 'es', 'pt', 'it', 'pl')
$engines = [ordered]@{}
foreach ($l in [Windows.Media.Ocr.OcrEngine]::AvailableRecognizerLanguages) {
  $base = $l.LanguageTag.Split('-')[0].ToLower()
  if ($wanted -notcontains $base) { continue }
  if ($engines.Contains($base)) { continue }
  $e = [Windows.Media.Ocr.OcrEngine]::TryCreateFromLanguage($l)
  if ($e) { $engines[$base] = $e }
}
# None of them installed (e.g. a Ukrainian or Turkish Windows): the profile language
# still reads the Latin rank and role words of an English client.
if ($engines.Count -eq 0) {
  $e = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
  if (-not $e) {
    foreach ($l in [Windows.Media.Ocr.OcrEngine]::AvailableRecognizerLanguages) {
      $e = [Windows.Media.Ocr.OcrEngine]::TryCreateFromLanguage($l)
      if ($e) { break }
    }
  }
  if ($e) { $engines[$e.RecognizerLanguage.LanguageTag.Split('-')[0].ToLower()] = $e }
}
[Console]::Out.WriteLine('LANGS ' + ($engines.Keys -join ','))
[Console]::Out.Flush()

while ($true) {
  $line = [Console]::In.ReadLine()
  if ($null -eq $line) { break }
  # REQL <id> <lang> <path>: one language only (faster once the client language is known).
  $only = $null
  if ($line.StartsWith('REQL ')) {
    $p4 = $line.Split(' ', 4)
    if ($p4.Count -lt 4) { continue }
    $only = $p4[2]
    $parts = @('REQ', $p4[1], $p4[3])
  } else {
    $parts = $line.Split(' ', 3)
  }
  if ($parts.Count -lt 3 -or $parts[0] -ne 'REQ') { continue }
  $id = $parts[1]
  $stream = $null
  try {
    $file = Await ([Windows.Storage.StorageFile]::GetFileFromPathAsync($parts[2])) ([Windows.Storage.StorageFile])
    $stream = Await ($file.OpenAsync([Windows.Storage.FileAccessMode]::Read)) ([Windows.Storage.Streams.IRandomAccessStream])
    $dec = Await ([Windows.Graphics.Imaging.BitmapDecoder]::CreateAsync($stream)) ([Windows.Graphics.Imaging.BitmapDecoder])
    $bmp = Await ($dec.GetSoftwareBitmapAsync()) ([Windows.Graphics.Imaging.SoftwareBitmap])
    foreach ($k in $engines.Keys) {
      if ($only -and $engines.Contains($only) -and $k -ne $only) { continue }
      $r =Await ($engines[$k].RecognizeAsync($bmp)) ([Windows.Media.Ocr.OcrResult])
      [Console]::Out.WriteLine("RES $id $k " + (B64 $r.Text))
    }
  } catch {
    [Console]::Out.WriteLine("ERR $id " + (B64 $_.Exception.Message))
  } finally {
    if ($stream) { $stream.Dispose() }
  }
  [Console]::Out.WriteLine("END $id")
  [Console]::Out.Flush()
}

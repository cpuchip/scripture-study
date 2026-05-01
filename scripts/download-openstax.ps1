$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$base = 'C:\Users\cpuch\Documents\code\stuffleberry\scripture-study\books\openstax'
$downloads = @(
  @{ url='https://assets.openstax.org/oscms-prodcms/media/documents/university-physics-volume-1_-_WEB.pdf'; dest="$base\university-physics-vol-1\university-physics-vol-1.pdf" }
  @{ url='https://assets.openstax.org/oscms-prodcms/media/documents/university-physics-volume-2_-_WEB.pdf'; dest="$base\university-physics-vol-2\university-physics-vol-2.pdf" }
  @{ url='https://assets.openstax.org/oscms-prodcms/media/documents/university-physics-volume-3_-_WEB.pdf'; dest="$base\university-physics-vol-3\university-physics-vol-3.pdf" }
  @{ url='https://assets.openstax.org/oscms-prodcms/media/documents/chemistry-2e_-_WEB.pdf'; dest="$base\chemistry-2e\chemistry-2e.pdf" }
  @{ url='https://assets.openstax.org/oscms-prodcms/media/documents/chemistry-atoms-first-2e_-_WEB.pdf'; dest="$base\chemistry-atoms-first-2e\chemistry-atoms-first-2e.pdf" }
  @{ url='https://assets.openstax.org/oscms-prodcms/media/documents/astronomy-2e_-_WEB.pdf'; dest="$base\astronomy-2e\astronomy-2e.pdf" }
)
foreach ($d in $downloads) {
  if (Test-Path $d.dest) {
    $existing = [math]::Round((Get-Item $d.dest).Length / 1MB, 1)
    Write-Host "SKIP (exists): $($d.dest) ($existing MB)"
    continue
  }
  Write-Host "Downloading $($d.url)"
  $sw = [System.Diagnostics.Stopwatch]::StartNew()
  try {
    Invoke-WebRequest -Uri $d.url -OutFile $d.dest -UserAgent 'Mozilla/5.0'
    $sw.Stop()
    $size = [math]::Round((Get-Item $d.dest).Length / 1MB, 1)
    Write-Host "  -> $($d.dest) ($size MB in $([math]::Round($sw.Elapsed.TotalSeconds,1))s)"
  } catch {
    Write-Host "  FAILED: $_"
  }
}
Write-Host "All downloads complete."

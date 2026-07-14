$root = $PSScriptRoot
$utf8bom = New-Object System.Text.UTF8Encoding $true

$files = @(
    "encoding.ps1",
    "release.ps1",
    "build.ps1",
    "release.config.example.ps1",
    "release.config.ps1",
    "fix-encoding.ps1"
)

foreach ($name in $files) {
    $path = Join-Path $root $name
    if (-not (Test-Path $path)) {
        continue
    }
    $content = [System.IO.File]::ReadAllText($path)
    [System.IO.File]::WriteAllText($path, $content, $utf8bom)
    Write-Host "UTF-8 BOM applied: $name"
}

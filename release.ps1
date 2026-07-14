param(
    [Parameter(Mandatory = $true)]
    [string]$Version,

    [string]$Owner = "",
    [string]$Repo = "kohiCrawling",
    [string]$Notes = "",
    [switch]$Draft,
    [switch]$SkipPublish
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "encoding.ps1")
. (Join-Path $PSScriptRoot "scripts\platform-build.ps1")

$configPath = Join-Path $PSScriptRoot "release.config.ps1"
if (Test-Path $configPath) {
    . $configPath
    if ($Owner -eq "" -and $ReleaseOwner) { $Owner = $ReleaseOwner }
    if ($ReleaseRepo) { $Repo = $ReleaseRepo }
}

if ($Owner -eq "") {
    throw "GitHub Owner가 필요합니다. -Owner 옵션을 쓰거나 release.config.ps1 을 만드세요."
}

if ([string]::IsNullOrWhiteSpace($Repo)) {
    throw "GitHub Repo가 필요합니다. release.config.ps1 의 ReleaseRepo 를 확인하세요."
}

if ($Version -match "^v") {
    $Version = $Version.Substring(1)
}

$tag = "v$Version"
$updateURL = "https://raw.githubusercontent.com/$Owner/$Repo/main/update.json"
if ($updateURL -match "githubusercontent\.com/[^/]+//") {
    throw "Update URL 형식 오류: $updateURL (Owner/Repo 확인)"
}
$distDir = Join-Path $PSScriptRoot "dist"

Write-Host "=== KohiCrawling Release ==="
Write-Host "Version     : $Version"
Write-Host "Repository  : $Owner/$Repo"
Write-Host "Tag         : $tag"
Write-Host "Update URL  : $updateURL"
Write-Host "Platforms   : windows/amd64, darwin/amd64, darwin/arm64"
Write-Host ""

$ldflags = "-X kohiCrawling/kohi.Version=$Version -X kohiCrawling/kohi.UpdateURL=$updateURL"

Write-Host "빌드 중..."
Push-Location $PSScriptRoot
try {
    $built = Build-KohiPlatforms -DistDir $distDir -Ldflags $ldflags
} finally {
    Pop-Location
}

$manifest = New-KohiUpdateManifest -Version $Version -Owner $Owner -Repo $Repo -Tag $tag -Built $built
$distManifestPath = Join-Path $distDir "update.json"
$rootManifestPath = Join-Path $PSScriptRoot "update.json"
Write-JsonNoBom -Path $distManifestPath -Object $manifest
Write-JsonNoBom -Path $rootManifestPath -Object $manifest

Write-Host "빌드 완료:"
foreach ($key in @("windows_amd64", "darwin_amd64", "darwin_arm64")) {
    Write-Host "  $key -> $($built[$key].File)"
    Write-Host "  SHA256: $($built[$key].SHA256)"
}
Write-Host ""

if ($SkipPublish) {
    Write-Host "SkipPublish 옵션으로 GitHub Release 생성을 건너뜁니다."
    Write-Host "dist/update.json 과 update.json 이 생성되었습니다."
    exit 0
}

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    throw "GitHub CLI(gh)가 설치되어 있지 않습니다. https://cli.github.com/"
}

gh auth status | Out-Null

Write-Host "GitHub Release 생성 중..."

$ghArgs = @(
    "release", "create", $tag,
    "--repo", "$Owner/$Repo",
    "--title", "kohiCrawling $tag"
)

if ($Notes -ne "") {
    $ghArgs += @("--notes", $Notes)
} else {
    $ghArgs += "--generate-notes"
}

if ($Draft) {
    $ghArgs += "--draft"
}

foreach ($key in @("windows_amd64", "darwin_amd64", "darwin_arm64")) {
    $ghArgs += $built[$key].Path
}
$ghArgs += $distManifestPath

& gh @ghArgs
if ($LASTEXITCODE -ne 0) {
    throw "GitHub Release 생성 실패 (exit code: $LASTEXITCODE)"
}

Write-Host ""
Write-Host "=== 배포 완료 ==="
Write-Host "Release : https://github.com/$Owner/$Repo/releases/tag/$tag"
Write-Host ""
Write-Host "update.json 을 main 브랜치에 커밋/푸시하면 클라이언트가 자동 업데이트를 확인합니다."
Write-Host "  git add update.json"
Write-Host "  git commit -m ""chore: release $tag"""
Write-Host "  git push origin main"

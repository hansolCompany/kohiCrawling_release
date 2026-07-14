param(
    [ValidateSet("kohi", "longterm", "server", "all")]
    [string]$Target = "kohi",

    [string]$Version = "1.0.0",
    [string]$Owner = "",
    [string]$Repo = "kohiCrawling",
    [string]$UpdateURL = "",

    [switch]$AllPlatforms
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

function Get-DefaultUpdateURL([string]$name) {
    if ($UpdateURL -ne "") { return $UpdateURL }
    if ($Owner -ne "" -and $Repo -ne "") {
        return "https://raw.githubusercontent.com/$Owner/$Repo/main/update.json"
    }
    return "https://example.com/$name/update.json"
}

function Build-CurrentPlatform {
    param(
        [string]$Name,
        [string]$PackagePath,
        [string]$LdflagsVersion,
        [string]$LdflagsUpdateURL
    )

    $ldflags = "-X $LdflagsVersion=$Version -X $LdflagsUpdateURL=$(Get-DefaultUpdateURL $Name)"
    if ($IsWindows -or ($env:OS -match "Windows")) {
        $output = "$Name.exe"
    } else {
        $output = $Name
    }

    Write-Host "빌드: $output"
    go build -ldflags $ldflags -o $output $PackagePath
    Write-Host "  완료: $output"
    Write-Host "  SHA256: $(Get-FileHashLower $output)"
    Write-Host ""
}

Write-Host "빌드 버전: $Version"
if ($AllPlatforms) {
    Write-Host "플랫폼: windows/amd64, darwin/amd64, darwin/arm64 -> dist/"
}
Write-Host ""

$distDir = Join-Path $PSScriptRoot "dist"

Push-Location $PSScriptRoot
try {
    switch ($Target) {
        "kohi" {
            if ($AllPlatforms) {
                $ldflags = "-X kohiCrawling/kohi.Version=$Version -X kohiCrawling/kohi.UpdateURL=$(Get-DefaultUpdateURL 'kohiCrawling')"
                Build-KohiPlatforms -DistDir $distDir -Ldflags $ldflags | Out-Null
            } else {
                Build-CurrentPlatform "kohiCrawling" "./cmd/kohi" "kohiCrawling/kohi.Version" "kohiCrawling/kohi.UpdateURL"
            }
        }
        "longterm" {
            if ($AllPlatforms) {
                $ldflags = "-X kohiCrawling/longterm.Version=$Version -X kohiCrawling/longterm.UpdateURL=$(Get-DefaultUpdateURL 'longtermCrawling')"
                Build-AppPlatforms -Name "longtermCrawling" -PackagePath "./cmd/longterm" -DistDir $distDir -Ldflags $ldflags
            } else {
                Build-CurrentPlatform "longtermCrawling" "./cmd/longterm" "kohiCrawling/longterm.Version" "kohiCrawling/longterm.UpdateURL"
            }
        }
        "server" {
            if ($AllPlatforms) {
                Build-AppPlatforms -Name "kohiCrawlingServer" -PackagePath "./cmd/server" -DistDir $distDir
            } else {
                if ($IsWindows -or ($env:OS -match "Windows")) {
                    $output = "kohiCrawlingServer.exe"
                } else {
                    $output = "kohiCrawlingServer"
                }
                Write-Host "빌드: $output"
                go build -o $output ./cmd/server
                Write-Host "  완료: $output"
                Write-Host ""
            }
        }
        "all" {
            if ($AllPlatforms) {
                $kohiLd = "-X kohiCrawling/kohi.Version=$Version -X kohiCrawling/kohi.UpdateURL=$(Get-DefaultUpdateURL 'kohiCrawling')"
                $longLd = "-X kohiCrawling/longterm.Version=$Version -X kohiCrawling/longterm.UpdateURL=$(Get-DefaultUpdateURL 'longtermCrawling')"
                Build-KohiPlatforms -DistDir $distDir -Ldflags $kohiLd | Out-Null
                Build-AppPlatforms -Name "longtermCrawling" -PackagePath "./cmd/longterm" -DistDir $distDir -Ldflags $longLd
                Build-AppPlatforms -Name "kohiCrawlingServer" -PackagePath "./cmd/server" -DistDir $distDir
            } else {
                Build-CurrentPlatform "kohiCrawling" "./cmd/kohi" "kohiCrawling/kohi.Version" "kohiCrawling/kohi.UpdateURL"
                Build-CurrentPlatform "longtermCrawling" "./cmd/longterm" "kohiCrawling/longterm.Version" "kohiCrawling/longterm.UpdateURL"
                if ($IsWindows -or ($env:OS -match "Windows")) {
                    go build -o kohiCrawlingServer.exe ./cmd/server
                } else {
                    go build -o kohiCrawlingServer ./cmd/server
                }
            }
        }
    }
} finally {
    Pop-Location
}

Write-Host "GitHub Release 배포는 .\release.ps1 -Version $Version 를 사용하세요."

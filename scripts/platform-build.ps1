function Reset-GoPlatformEnv {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
}

function Get-FileHashLower {
    param([string]$Path)
    return (Get-FileHash -Path $Path -Algorithm SHA256).Hash.ToLower()
}

function Build-GoBinary {
    param(
        [string]$GoOS,
        [string]$GoArch,
        [string]$OutputPath,
        [string]$PackagePath,
        [string]$Ldflags = ""
    )

    $env:GOOS = $GoOS
    $env:GOARCH = $GoArch
    $env:CGO_ENABLED = "0"

    Write-Host "  $GoOS/$GoArch -> $OutputPath"
    if ($Ldflags -ne "") {
        go build -ldflags $Ldflags -o $OutputPath $PackagePath
    } else {
        go build -o $OutputPath $PackagePath
    }

    if ($LASTEXITCODE -ne 0) {
        Reset-GoPlatformEnv
        throw "build failed: $GoOS/$GoArch ($OutputPath)"
    }
}

function Build-KohiPlatforms {
    param(
        [string]$DistDir,
        [string]$Ldflags
    )

    New-Item -ItemType Directory -Force -Path $DistDir | Out-Null

    $artifacts = @(
        @{ Key = "windows_amd64"; GoOS = "windows"; GoArch = "amd64"; File = "kohiCrawling.exe" },
        @{ Key = "darwin_amd64";  GoOS = "darwin";  GoArch = "amd64"; File = "kohiCrawling-darwin-amd64" },
        @{ Key = "darwin_arm64";  GoOS = "darwin";  GoArch = "arm64"; File = "kohiCrawling-darwin-arm64" }
    )

    $built = @{}
    foreach ($item in $artifacts) {
        $outPath = Join-Path $DistDir $item.File
        Build-GoBinary -GoOS $item.GoOS -GoArch $item.GoArch -OutputPath $outPath -PackagePath "./cmd/kohi" -Ldflags $Ldflags
        $built[$item.Key] = @{
            Path = $outPath
            File = $item.File
            SHA256 = Get-FileHashLower $outPath
        }
    }

    Reset-GoPlatformEnv
    return $built
}

function Build-AppPlatforms {
    param(
        [string]$Name,
        [string]$PackagePath,
        [string]$DistDir,
        [string]$Ldflags = ""
    )

    New-Item -ItemType Directory -Force -Path $DistDir | Out-Null

    $artifacts = @(
        @{ GoOS = "windows"; GoArch = "amd64"; File = "$Name.exe" },
        @{ GoOS = "darwin";  GoArch = "amd64"; File = "$Name-darwin-amd64" },
        @{ GoOS = "darwin";  GoArch = "arm64"; File = "$Name-darwin-arm64" }
    )

    foreach ($item in $artifacts) {
        $outPath = Join-Path $DistDir $item.File
        Build-GoBinary -GoOS $item.GoOS -GoArch $item.GoArch -OutputPath $outPath -PackagePath $PackagePath -Ldflags $Ldflags
        Write-Host "  SHA256: $(Get-FileHashLower $outPath)"
    }

    Reset-GoPlatformEnv
}

function New-KohiUpdateManifest {
    param(
        [string]$Version,
        [string]$Owner,
        [string]$Repo,
        [string]$Tag,
        [hashtable]$Built
    )

    $base = "https://github.com/$Owner/$Repo/releases/download/$Tag"
    $assets = [ordered]@{}
    foreach ($key in @("windows_amd64", "darwin_amd64", "darwin_arm64")) {
        $item = $Built[$key]
        $assets[$key] = [ordered]@{
            download_url = "$base/$($item.File)"
            sha256       = $item.SHA256
        }
    }

    return [ordered]@{
        version      = $Version
        download_url = $assets.windows_amd64.download_url
        sha256       = $assets.windows_amd64.sha256
        assets       = $assets
    }
}

function Write-JsonNoBom {
    param(
        [string]$Path,
        [object]$Object
    )

    $json = $Object | ConvertTo-Json -Depth 5
    $utf8NoBom = [System.Text.UTF8Encoding]::new($false)
    [System.IO.File]::WriteAllText($Path, $json, $utf8NoBom)
}

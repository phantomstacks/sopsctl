# Heavily inspired by: fluxcd/flux2 install script
# PowerShell installation script for sopsctl on Windows

param(
    [string]$BinDir = "$env:LOCALAPPDATA\sopsctl\bin",
    [string]$Version = ""
)

$ErrorActionPreference = 'Stop'

$GITHUB_REPO = "phantomstacks/sopsctl"

# Helper functions for logs
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Warning "[WARN] $Message"
}

function Write-Fatal {
    param([string]$Message)
    Write-Error "[ERROR] $Message"
    exit 1
}

# Set OS - should always be windows for this script
function Get-OperatingSystem {
    return "windows"
}

# Set architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        "x86" { return "386" }
        default { Write-Fatal "Unsupported architecture: $arch" }
    }
}

# Create temporary directory
function New-TempDirectory {
    $tmpDir = Join-Path $env:TEMP "sopsctl-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
    return $tmpDir
}

# Get release version from GitHub
function Get-ReleaseVersion {
    param(
        [string]$TmpMetadata,
        [string]$SpecifiedVersion
    )

    if ($SpecifiedVersion) {
        $suffixUrl = "tags/v$SpecifiedVersion"
    } else {
        $suffixUrl = "latest"
    }

    $metadataUrl = "https://api.github.com/repos/$GITHUB_REPO/releases/$suffixUrl"

    Write-Info "Downloading metadata from $metadataUrl"

    try {
        $headers = @{}
        if ($env:GITHUB_TOKEN) {
            $headers["Authorization"] = "token $env:GITHUB_TOKEN"
        }

        Invoke-WebRequest -Uri $metadataUrl -OutFile $TmpMetadata -Headers $headers -UseBasicParsing
        $metadata = Get-Content $TmpMetadata | ConvertFrom-Json
        $version = $metadata.tag_name -replace '^v', ''

        if ($version) {
            Write-Info "Using $version as release"
            return $version
        } else {
            Write-Fatal "Unable to determine release version"
        }
    } catch {
        Write-Fatal "Failed to download metadata: $_"
    }
}

# Download file from URL
function Get-FileFromUrl {
    param(
        [string]$Url,
        [string]$OutFile
    )

    try {
        $headers = @{}
        if ($env:GITHUB_TOKEN) {
            $headers["Authorization"] = "token $env:GITHUB_TOKEN"
        }

        Invoke-WebRequest -Uri $Url -OutFile $OutFile -Headers $headers -UseBasicParsing
    } catch {
        Write-Fatal "Download failed: $_"
    }
}

# Download hash from GitHub
function Get-Hash {
    param(
        [string]$Version,
        [string]$OS,
        [string]$Arch,
        [string]$TmpHash
    )

    $hashUrl = "https://github.com/$GITHUB_REPO/releases/download/v$Version/sopsctl_${Version}_checksums.txt"
    Write-Info "Downloading hash from $hashUrl"
    Get-FileFromUrl -Url $hashUrl -OutFile $TmpHash

    $hashContent = Get-Content $TmpHash
    $expectedHash = $hashContent | Where-Object { $_ -match "sopsctl_${Version}_${OS}_${Arch}.zip" }

    if ($expectedHash) {
        $hash = ($expectedHash -split '\s+')[0]
        Write-Info "Expected hash: $hash"
        return $hash
    } else {
        Write-Fatal "Could not find hash for sopsctl_${Version}_${OS}_${Arch}.zip"
    }
}

# Download binary from GitHub
function Get-Binary {
    param(
        [string]$Version,
        [string]$OS,
        [string]$Arch,
        [string]$TmpBin
    )

    $binUrl = "https://github.com/$GITHUB_REPO/releases/download/v$Version/sopsctl_${Version}_${OS}_${Arch}.zip"
    Write-Info "Downloading binary from $binUrl"
    Get-FileFromUrl -Url $binUrl -OutFile $TmpBin
}

# Compute SHA256 hash
function Get-Sha256Hash {
    param([string]$FilePath)

    $hash = Get-FileHash -Path $FilePath -Algorithm SHA256
    return $hash.Hash.ToLower()
}

# Verify binary hash
function Test-BinaryHash {
    param(
        [string]$ExpectedHash,
        [string]$TmpBin
    )

    Write-Info "Verifying binary download"
    $actualHash = Get-Sha256Hash -FilePath $TmpBin

    if ($ExpectedHash -ne $actualHash) {
        Write-Fatal "Download sha256 does not match. Expected: $ExpectedHash, Got: $actualHash"
    }

    Write-Info "Hash verification successful"
}

# Extract and install binary
function Install-Binary {
    param(
        [string]$TmpDir,
        [string]$TmpBin,
        [string]$BinDir
    )

    Write-Info "Extracting binary"

    # Extract zip - requires tar.exe (available in Windows 10 1803+ and Windows Server 2019+)
    if (Get-Command tar -ErrorAction SilentlyContinue) {
        $extractDir = Join-Path $TmpDir "extract"
        New-Item -ItemType Directory -Path $extractDir -Force | Out-Null
        tar -xzf $TmpBin -C $extractDir
        $exePath = Join-Path $extractDir "sopsctl.exe"
    } else {
        Write-Fatal "tar.exe not found. Please install tar or use Windows 10 1803+ / Windows Server 2019+"
    }

    # Create bin directory if it doesn't exist
    if (-not (Test-Path $BinDir)) {
        Write-Info "Creating directory $BinDir"
        New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
    }

    # Move binary to bin directory
    $destination = Join-Path $BinDir "sopsctl.exe"
    Write-Info "Installing sopsctl to $destination"

    try {
        Copy-Item -Path $exePath -Destination $destination -Force
    } catch {
        Write-Fatal "Failed to install binary: $_"
    }

    # Check if bin directory is in PATH
    $pathParts = $env:PATH -split ';'
    if ($pathParts -notcontains $BinDir) {
        Write-Warn "The installation directory is not in your PATH."
        Write-Info "Add it to your PATH by running:"
        Write-Host "`n  `$env:PATH += ';$BinDir'`n" -ForegroundColor Yellow
        Write-Info "To make it permanent, add the directory to your system PATH environment variable."
    }

    Write-Info "Installation complete!"
}

# Main installation process
try {
    Write-Info "Starting sopsctl installation"

    $os = Get-OperatingSystem
    $arch = Get-Architecture

    Write-Info "Detected OS: $os"
    Write-Info "Detected Architecture: $arch"

    # Create temporary directory
    $tmpDir = New-TempDirectory
    $tmpMetadata = Join-Path $tmpDir "sopsctl.json"
    $tmpHash = Join-Path $tmpDir "sopsctl.hash"
    $tmpBin = Join-Path $tmpDir "sopsctl.zip"

    try {
        # Get version
        $releaseVersion = Get-ReleaseVersion -TmpMetadata $tmpMetadata -SpecifiedVersion $Version

        # Download hash
        $expectedHash = Get-Hash -Version $releaseVersion -OS $os -Arch $arch -TmpHash $tmpHash

        # Download binary
        Get-Binary -Version $releaseVersion -OS $os -Arch $arch -TmpBin $tmpBin

        # Verify hash
        Test-BinaryHash -ExpectedHash $expectedHash -TmpBin $tmpBin

        # Install binary
        Install-Binary -TmpDir $tmpDir -TmpBin $tmpBin -BinDir $BinDir

    } finally {
        # Cleanup temporary directory
        if (Test-Path $tmpDir) {
            Write-Info "Cleaning up temporary files"
            Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }

} catch {
    Write-Fatal "Installation failed: $_"
}


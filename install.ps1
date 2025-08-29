# Requires -RunAsAdministrator

# Function to get the latest release version
function Get-LatestRelease {
    $releaseUrl = "https://api.github.com/repos/base-go/cmd/releases/latest"
    try {
        $release = Invoke-RestMethod -Uri $releaseUrl -ErrorAction Stop
        return $release.tag_name
    }
    catch {
        Write-Error "Failed to get latest release: $_"
        exit 1
    }
}

# Function to create directory if it doesn't exist
function Ensure-Directory {
    param([string]$Path)
    if (-not (Test-Path $Path)) {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }
}

Write-Host "Installing Base CLI..." -ForegroundColor Green

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
Write-Host "Architecture: windows_$arch"

# Set installation paths
$installDir = Join-Path $env:USERPROFILE ".base"
$binDir = Join-Path $env:USERPROFILE "bin"

# Create directories
Ensure-Directory $installDir
Ensure-Directory $binDir

# Get latest release
$version = Get-LatestRelease
Write-Host "Latest version: $version"

# Download URL
$downloadUrl = "https://github.com/base-go/cmd/releases/download/$version/base_windows_$arch.zip"
$zipPath = Join-Path $env:TEMP "base.zip"
$exePath = Join-Path $installDir "base.exe"
$binPath = Join-Path $binDir "base.exe"

Write-Host "Downloading from: $downloadUrl"

try {
    # Download the zip file
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -ErrorAction Stop

    # Extract the zip
    Expand-Archive -Path $zipPath -DestinationPath $installDir -Force

    # Copy to bin directory
    Copy-Item -Path $exePath -Destination $binPath -Force

    # Add to PATH if not already there
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$binDir*") {
        $newPath = "$userPath;$binDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "Added $binDir to PATH"
    }

    Write-Host "`nBase CLI has been installed successfully!" -ForegroundColor Green
    
    # Install Go dependencies
    Write-Host "`nInstalling Base CLI dependencies..." -ForegroundColor Yellow
    
    # Check if Go is installed
    if (Get-Command go -ErrorAction SilentlyContinue) {
        Write-Host "Installing go-swagger (API documentation generator)..."
        try {
            $null = & go install github.com/swaggo/swag/cmd/swag@latest 2>$null
            Write-Host "✓ go-swagger installed successfully" -ForegroundColor Green
        }
        catch {
            Write-Host "Warning: Failed to install go-swagger. You can install it manually later with:" -ForegroundColor Yellow
            Write-Host "  go install github.com/swaggo/swag/cmd/swag@latest"
        }
    }
    else {
        Write-Host "Warning: Go is not installed or not in PATH." -ForegroundColor Yellow
        Write-Host "Base CLI dependencies (go-swagger) will be installed automatically when needed."
        Write-Host "To install Go, visit: https://golang.org/dl/"
    }
    
    Write-Host "`nPlease restart your terminal to use the 'base' command"
}
catch {
    Write-Error "Installation failed: $_"
    exit 1
}
finally {
    # Cleanup
    if (Test-Path $zipPath) {
        Remove-Item $zipPath -Force
    }
}

Write-Host "`nTo get started, run: base --help"

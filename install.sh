#!/bin/bash

set -e

# Detect OS and architecture
detect_os() {
    case "$(uname -s)" in
        Darwin*)   echo "darwin" ;;
        Linux*)    echo "linux" ;;
        MINGW64*|MSYS*|CYGWIN*) echo "windows" ;;
        *)         echo "unknown" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)            echo "unknown" ;;
    esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
    echo "Unsupported operating system or architecture"
    exit 1
fi

# Set installation directories based on OS
if [ "$OS" = "windows" ]; then
    INSTALL_DIR="$USERPROFILE/.base"
    BIN_DIR="$USERPROFILE/bin"
    BINARY_NAME="base.exe"
else
    INSTALL_DIR="$HOME/.base"
    BIN_DIR="/usr/local/bin"
    BINARY_NAME="base"
fi

# Create installation directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$BIN_DIR" 2>/dev/null || {
    echo "Error: Unable to create $BIN_DIR directory. Please run with sudo:"
    echo "curl -sSL https://raw.githubusercontent.com/BaseTechStack/basecmd/main/install.sh | sudo bash"
    exit 1
}

echo "Installing Base CLI..."
echo "OS: $OS"
echo "Architecture: $ARCH"

# Get the latest release version
echo "Fetching latest release information..."
API_RESPONSE=$(curl -s https://api.github.com/repos/BaseTechStack/basecmd/releases/latest)
if [ $? -ne 0 ]; then
    echo "Error: Failed to fetch release information"
    exit 1
fi

LATEST_RELEASE=$(echo "$API_RESPONSE" | grep '"tag_name"' | head -n1 | cut -d '"' -f 4)
if [ -z "$LATEST_RELEASE" ]; then
    echo "Error: Could not determine latest version"
    echo "API Response debug info:"
    echo "$API_RESPONSE" | head -n 10
    echo "Please check if the repository exists and has releases"
    echo "Repository: https://github.com/BaseTechStack/basecmd"
    exit 1
fi

echo "Latest version: $LATEST_RELEASE"

# Download the appropriate binary
DOWNLOAD_URL="https://github.com/BaseTechStack/basecmd/releases/download/$LATEST_RELEASE/base_${OS}_${ARCH}.tar.gz"
if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="https://github.com/BaseTechStack/basecmd/releases/download/$LATEST_RELEASE/base_${OS}_${ARCH}.zip"
fi

echo "Downloading from: $DOWNLOAD_URL"
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if [ "$OS" = "windows" ]; then
    curl -sL "$DOWNLOAD_URL" -o base.zip
    unzip base.zip
else
    curl -sL "$DOWNLOAD_URL" | tar xz
fi

# Install the binary
mv "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Create symlink
if [ "$OS" = "windows" ]; then
    # On Windows, copy to bin directory
    cp "$INSTALL_DIR/$BINARY_NAME" "$BIN_DIR/"
else
    # On Unix systems, create symlink with sudo
    echo "Creating symlink in $BIN_DIR (requires sudo)..."
    if ! sudo ln -sf "$INSTALL_DIR/$BINARY_NAME" "$BIN_DIR/$BINARY_NAME"; then
        echo "Error: Failed to create symlink. Please run the install script with sudo:"
        echo "curl -sSL https://raw.githubusercontent.com/BaseTechStack/basecmd/main/install.sh | sudo bash"
        exit 1
    fi
fi

# Cleanup
cd - > /dev/null
rm -rf "$TMP_DIR"

echo "Base CLI has been installed successfully!"

# Install Go dependencies
echo ""
echo "Installing Base CLI dependencies..."

# Check if Go is installed
if command -v go >/dev/null 2>&1; then
    echo "Installing Air (hot reloading tool)..."
    if ! go install github.com/air-verse/air@latest 2>/dev/null; then
        echo "Warning: Failed to install Air. You can install it manually later with:"
        echo "  go install github.com/air-verse/air@latest"
    else
        echo "✓ Air installed successfully"
    fi
    
    echo "Installing Swag (Swagger documentation generator)..."
    if ! go install github.com/swaggo/swag/cmd/swag@latest 2>/dev/null; then
        echo "Warning: Failed to install Swag. You can install it manually later with:"
        echo "  go install github.com/swaggo/swag/cmd/swag@latest"
    else
        echo "✓ Swag installed successfully"
    fi
else
    echo "Warning: Go is not installed or not in PATH."
    echo "Base CLI dependencies (Air and Swag) will be installed automatically when needed."
    echo "To install Go, visit: https://golang.org/dl/"
fi

echo ""
echo "Run 'base --help' to get started"

# Add to PATH for Windows if needed
if [ "$OS" = "windows" ]; then
    if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
        echo "Please add $BIN_DIR to your PATH to use the 'base' command"
        echo "You can do this by running:"
        echo "    setx PATH \"%PATH%;$BIN_DIR\""
    fi
fi
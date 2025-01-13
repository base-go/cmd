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
    echo "curl -sSL https://raw.githubusercontent.com/base-go/cmd/main/install.sh | sudo bash"
    exit 1
}

echo "Installing Base CLI..."
echo "OS: $OS"
echo "Architecture: $ARCH"

# Get the latest release version
LATEST_RELEASE=$(curl -s https://api.github.com/repos/base-go/cmd/releases/latest | grep "tag_name" | cut -d '"' -f 4)
if [ -z "$LATEST_RELEASE" ]; then
    echo "Error: Could not determine latest version"
    exit 1
fi

echo "Latest version: $LATEST_RELEASE"

# Download the appropriate binary
DOWNLOAD_URL="https://github.com/base-go/cmd/releases/download/$LATEST_RELEASE/base_${OS}_${ARCH}.tar.gz"
if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="https://github.com/base-go/cmd/releases/download/$LATEST_RELEASE/base_${OS}_${ARCH}.zip"
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
    # On Unix systems, create symlink
    ln -sf "$INSTALL_DIR/$BINARY_NAME" "$BIN_DIR/$BINARY_NAME"
fi

# Cleanup
cd - > /dev/null
rm -rf "$TMP_DIR"

echo "Base CLI has been installed successfully!"
echo "Run 'base --help' to get started"

# Add to PATH for Windows if needed
if [ "$OS" = "windows" ]; then
    if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
        echo "Please add $BIN_DIR to your PATH to use the 'base' command"
        echo "You can do this by running:"
        echo "    setx PATH \"%PATH%;$BIN_DIR\""
    fi
fi
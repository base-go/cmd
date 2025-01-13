#!/bin/bash

# Step 1: Check if Go is installed
if ! [ -x "$(command -v go)" ]; then
  echo "Error: Go is not installed." >&2
  echo "Please install Go and rerun this script." >&2
  exit 1
fi

# Step 2: Define the repository URL and installation directory
REPO_URL="https://github.com/base-go/cmd/archive/main.zip"
INSTALL_DIR="$HOME/.base"
BIN_PATH="/usr/local/bin"

# Set version information
VERSION=${VERSION:-"1.0.0"}
COMMIT_HASH=${COMMIT_HASH:-$(curl -s https://api.github.com/repos/base-go/cmd/commits/main | grep '"sha"' | head -n 1 | cut -d'"' -f4)}
BUILD_DATE=${BUILD_DATE:-$(date -u '+%Y-%m-%dT%H:%M:%SZ')}
GO_VERSION=${GO_VERSION:-$(go version | cut -d' ' -f3)}

echo "Version information:"
echo "Version: $VERSION"
echo "Commit: $COMMIT_HASH"
echo "Build Date: $BUILD_DATE"
echo "Go Version: $GO_VERSION"

# Step 3: Download and extract the repository
if [ -d "$INSTALL_DIR" ]; then
  echo "Directory $INSTALL_DIR already exists. Removing existing directory."
  rm -rf "$INSTALL_DIR" || { echo "Failed to remove existing directory."; exit 1; }
fi

echo "Downloading the repository..."
mkdir -p "$INSTALL_DIR" || { echo "Failed to create installation directory."; exit 1; }
curl -L "$REPO_URL" -o "$INSTALL_DIR/base.zip" || { echo "Failed to download repository."; exit 1; }

echo "Extracting the repository..."
unzip "$INSTALL_DIR/base.zip" -d "$INSTALL_DIR" || { echo "Failed to extract repository."; exit 1; }
mv "$INSTALL_DIR"/cmd-main/* "$INSTALL_DIR" || { echo "Failed to move files."; exit 1; }
rm -rf "$INSTALL_DIR/cmd-main" "$INSTALL_DIR/base.zip"

# Step 4: Initialize and tidy up the module
echo "Initializing and tidying module..."
cd "$INSTALL_DIR" || { echo "Failed to change directory."; exit 1; }
echo "Current directory: $(pwd)"

go mod tidy || { echo "Failed to tidy module."; exit 1; }

# Step 5: Build the tool with version information
echo "Building the tool..."
go build -v \
  -ldflags "-X 'base/version.Version=$VERSION' \
            -X 'base/version.CommitHash=$COMMIT_HASH' \
            -X 'base/version.BuildDate=$BUILD_DATE' \
            -X 'base/version.GoVersion=$GO_VERSION'" \
  -o base || { echo "Failed to build the tool."; exit 1; }

# Step 6: Install the binary
echo "Installing the tool..."
if [ -f "$BIN_PATH/base" ]; then
  echo "Existing binary found. Removing..."
  sudo rm -f "$BIN_PATH/base" || { echo "Failed to remove existing binary."; exit 1; }
fi

sudo mv base "$BIN_PATH/base" || { echo "Failed to install the tool."; exit 1; }

echo "Installation completed successfully."
echo "You can now use 'base' from anywhere in your terminal."
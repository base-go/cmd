#!/bin/bash

# Step 1: Check if Go and Git are installed
if ! [ -x "$(command -v go)" ]; then
  echo "Error: Go is not installed." >&2
  echo "Please install Go and rerun this script." >&2
  exit 1
fi

if ! [ -x "$(command -v git)" ]; then
  echo "Error: Git is not installed." >&2
  echo "Please install Git and rerun this script." >&2
  exit 1
fi

# Step 2: Define the repository URL and installation directory
REPO_URL="https://github.com/base-go/cmd.git"
INSTALL_DIR="$HOME/.base"
BIN_PATH="/usr/local/bin"

# Step 3: Clone the repository
if [ -d "$INSTALL_DIR" ]; then
  echo "Directory $INSTALL_DIR already exists. Removing existing directory."
  rm -rf "$INSTALL_DIR" || { echo "Failed to remove existing directory."; exit 1; }
fi

echo "Cloning the repository..."
git clone "$REPO_URL" "$INSTALL_DIR" || { echo "Failed to clone repository."; exit 1; }

# Step 4: Initialize and tidy up the module
echo "Initializing and tidying module..."
cd "$INSTALL_DIR" || { echo "Failed to change directory."; exit 1; }
echo "Current directory: $(pwd)"

go mod tidy || { echo "Failed to tidy module."; exit 1; }

# Step 5: Build the tool
echo "Building the tool..."
go build -v -o base || { echo "Failed to build the tool."; exit 1; }

# Step 6: Install the binary
echo "Installing the tool..."
if [ -f "$BIN_PATH/base" ]; then
  echo "Existing binary found. Removing..."
  sudo rm -f "$BIN_PATH/base" || { echo "Failed to remove existing binary."; exit 1; }
fi

sudo mv base "$BIN_PATH/base" || { echo "Failed to install the tool."; exit 1; }

echo "Installation completed successfully."
echo "You can now use 'base' from anywhere in your terminal."
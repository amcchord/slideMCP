#!/bin/bash

# Slide MCP Server
# This script installs the OS X version the MCP server to the /usr/local/bin directory
# Only use this script for testing purposes


# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Run build-and-sign first
if ! ./build-and-sign.sh; then
    echo "Error: build-and-sign.sh failed"
    exit 1
fi



# Get the darwin-arm64 version from build directory
if [ ! -f "../build/slide-mcp-server-darwin-arm64" ]; then
    echo "Error: ../build/slide-mcp-server-darwin-arm64 not found"
    exit 1
fi

# Rename and set permissions
cp ../build/slide-mcp-server-darwin-arm64 ../build/slide-mcp-server
chmod +x ../build/slide-mcp-server

# Move to /usr/local/bin
mv ../build/slide-mcp-server /usr/local/bin/

echo "Installation complete"

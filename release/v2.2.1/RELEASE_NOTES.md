# Slide MCP Server v2.2.1

## Changes

- c7c34cf Cleanup
- Updated version to v2.2.1

## Installation

Download the appropriate binary for your platform:

- **Linux x64**: slide-mcp-server-v2.2.1-linux-x64.tar.gz
- **Linux ARM64**: slide-mcp-server-v2.2.1-linux-arm64.tar.gz  
- **macOS x64**: slide-mcp-server-v2.2.1-macos-x64.tar.gz (or darwin-amd64.tar.gz)
- **macOS ARM64**: slide-mcp-server-v2.2.1-macos-arm64.tar.gz (or darwin-arm64.tar.gz)
- **Windows x64**: slide-mcp-server-v2.2.1-windows-x64.zip

Note: Both macos and darwin named packages are provided for macOS. They contain identical binaries - use whichever naming convention you prefer.

## Installer

GUI installers are also available for easier setup:

- **macOS x64**: slide-mcp-installer-v2.2.1-darwin-amd64-signed.tar.gz
- **macOS ARM64**: slide-mcp-installer-v2.2.1-darwin-arm64-signed.tar.gz
- **Windows x64**: slide-mcp-installer-v2.2.1-windows-amd64.zip
- **Windows 32-bit**: slide-mcp-installer-v2.2.1-windows-386.zip
- **Linux x64**: slide-mcp-installer-v2.2.1-linux-amd64.tar.gz
- **Linux ARM64**: slide-mcp-installer-v2.2.1-linux-arm64.tar.gz

## Verification

Verify the integrity of your download using the checksums.sha256 file:

```bash
shasum -a 256 -c checksums.sha256
```

## macOS Security

The macOS binaries are signed and notarized by Apple. They should run without security warnings on macOS 10.15+ systems.

For older macOS versions or if you encounter security warnings, you may need to run:

```bash
xattr -d com.apple.quarantine slide-mcp-server
``` 
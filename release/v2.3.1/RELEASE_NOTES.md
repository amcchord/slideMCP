# Slide MCP Server v2.3.1

## Changes

- 7d5e054 Installer updates current installed version after updating.
- 2d7b890 Rebuilt Documentation to request data from the actual documentation website
- d8513a3 Add RDP caching for easy download links

## Installation

Download the appropriate binary for your platform:

- **Linux x64**: slide-mcp-server-v2.3.1-linux-x64.tar.gz
- **Linux ARM64**: slide-mcp-server-v2.3.1-linux-arm64.tar.gz  
- **macOS x64**: slide-mcp-server-v2.3.1-macos-x64.tar.gz (or darwin-amd64.tar.gz)
- **macOS ARM64**: slide-mcp-server-v2.3.1-macos-arm64.tar.gz (or darwin-arm64.tar.gz)
- **Windows x64**: slide-mcp-server-v2.3.1-windows-x64.zip

Note: Both macos and darwin named packages are provided for macOS. They contain identical binaries - use whichever naming convention you prefer.

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

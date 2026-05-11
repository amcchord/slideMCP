# Slide MCP Server v2.6.0

## Changes

- eafe1c7 Release v2.6.0: Add push restore support and update snapshot location default

## Installation

Download the appropriate binary for your platform:

- **Linux x64**: slide-mcp-server-$NEW_VERSION-macos-x64.tar.gz
- **Linux ARM64**: slide-mcp-server-$NEW_VERSION-macos-arm64.tar.gz  
- **macOS x64**: slide-mcp-server-$NEW_VERSION-macos-x64.tar.gz (or darwin-amd64.tar.gz)
- **macOS ARM64**: slide-mcp-server-$NEW_VERSION-macos-arm64.tar.gz (or darwin-arm64.tar.gz)
- **Windows x64**: slide-mcp-server-$NEW_VERSION-windows-x64.zip

## Verification

Verify the integrity of your download using the checksums.sha256 file:

```bash
shasum -a 256 -c checksums.sha256
```

## macOS Security

The macOS binaries are signed and notarized by Apple. They should run without security warnings on macOS 10.15+ systems.


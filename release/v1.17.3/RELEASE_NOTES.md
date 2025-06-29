# Slide MCP Server v1.17.3

## Changes

- b59cc50 Make the default safer
- 486f954 Dedicated list_deleted
- a2f0dd4 Permissions
- bec30b8 Release v1.17 with notarized macOS binaries
- 6b4c3df Fix infinite recursion in enrichWithClientName function
- f0cb4d8 Update v1.16 with notarized macOS binaries

## Installation

Download the appropriate binary for your platform:

- **Linux x64**: slide-mcp-server-v1.17.3-linux-x64.tar.gz
- **Linux ARM64**: slide-mcp-server-v1.17.3-linux-arm64.tar.gz  
- **macOS x64**: slide-mcp-server-v1.17.3-macos-x64.tar.gz
- **macOS ARM64**: slide-mcp-server-v1.17.3-macos-arm64.tar.gz
- **Windows x64**: slide-mcp-server-v1.17.3-windows-x64.zip

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

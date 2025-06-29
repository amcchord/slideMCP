# Slide MCP Server v1.17.4

## Changes

- fb3bcd3 Add --exit-after-first flag for single request processing
- 3cec572 Permissions tweaks
- 1417907 Readme Clarity
- fa96059 Security: Remove hardcoded Apple credentials from release script

## Installation

Download the appropriate binary for your platform:

- **Linux x64**: slide-mcp-server-v1.17.4-linux-x64.tar.gz
- **Linux ARM64**: slide-mcp-server-v1.17.4-linux-arm64.tar.gz  
- **macOS x64**: slide-mcp-server-v1.17.4-macos-x64.tar.gz
- **macOS ARM64**: slide-mcp-server-v1.17.4-macos-arm64.tar.gz
- **Windows x64**: slide-mcp-server-v1.17.4-windows-x64.zip

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

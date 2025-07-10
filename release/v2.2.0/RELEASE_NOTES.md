# Slide MCP Server v2.2.0

## Changes

- cef9b76 Intermediate Commit
- 86cb1f6 Add safety measures to prevent reports tool from hanging
- 0fb5fe3 Fix API batch size limit back to 50
- 7045b9a Optimize reports generation with parallel processing and caching
- f4b6919 Meta Tools
- d9fdd7f Update README: Add Mac installer to v2.0.2 release and update version references

## Installation

Download the appropriate binary for your platform:

- **Linux x64**: slide-mcp-server-v2.2.0-linux-x64.tar.gz
- **Linux ARM64**: slide-mcp-server-v2.2.0-linux-arm64.tar.gz  
- **macOS x64**: slide-mcp-server-v2.2.0-macos-x64.tar.gz (or darwin-amd64.tar.gz)
- **macOS ARM64**: slide-mcp-server-v2.2.0-macos-arm64.tar.gz (or darwin-arm64.tar.gz)
- **Windows x64**: slide-mcp-server-v2.2.0-windows-x64.zip

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

# Slide MCP Server & Installer v2.2.0

This release includes both the Slide MCP Server and the new GUI Installer application.

## What's New

### MCP Server v2.2.0
- Version bump to v2.2.0
- Continued improvements to stability and performance
- All macOS binaries are signed and notarized

### MCP Installer v2.2.0 (NEW!)
- **New GUI Installer**: Easy-to-use graphical installer for Windows and macOS
- **Improved Windows Support**: Uses native Go libraries for ZIP extraction
- **Full macOS Notarization**: Both Intel and Apple Silicon builds are signed and notarized
- **Automatic Configuration**: Detects Claude Desktop and configures it automatically
- **API Key Management**: Securely manages your Slide API key

## Downloads

### MCP Server Binaries
- **Linux x64**: `slide-mcp-server-v2.2.0-linux-x64.tar.gz`
- **Linux ARM64**: `slide-mcp-server-v2.2.0-linux-arm64.tar.gz`  
- **macOS x64**: `slide-mcp-server-v2.2.0-macos-x64.tar.gz`
- **macOS ARM64**: `slide-mcp-server-v2.2.0-macos-arm64.tar.gz`
- **Windows x64**: `slide-mcp-server-v2.2.0-windows-x64.zip`

### GUI Installer (Recommended for Easy Setup)
- **macOS ARM64 (M1/M2/M3/M4)**: `slide-mcp-installer-v2.2.0-darwin-arm64-signed.tar.gz`
- **macOS Intel**: `slide-mcp-installer-v2.2.0-darwin-amd64-signed.tar.gz`
- **Windows 64-bit**: `slide-mcp-installer-v2.2.0-windows-amd64.zip`
- **Windows 32-bit**: `slide-mcp-installer-v2.2.0-windows-386.zip`

## Installation

### Option 1: Using the GUI Installer (Recommended)

#### macOS
1. Download the installer for your Mac (arm64 for Apple Silicon, amd64 for Intel)
2. Extract: `tar -xzf slide-mcp-installer-v2.2.0-darwin-[arch]-signed.tar.gz`
3. Double-click `slide-mcp-installer.app` or run `open slide-mcp-installer.app`
4. Enter your Slide API key and click Install

#### Windows
1. Download the installer (amd64 for 64-bit, 386 for 32-bit)
2. Extract the ZIP file
3. Run `slide-mcp-installer.exe`
4. Enter your Slide API key and click Install

The installer will:
- Download the correct MCP server binary
- Install it to the appropriate location
- Configure Claude Desktop automatically
- Manage your API key securely

### Option 2: Manual Installation

Follow the instructions in the main README for manual installation of the server binaries.

## Verification

All macOS binaries (both server and installer) are signed and notarized by Apple. They should run without security warnings on macOS 10.15+.

To verify checksums:
```bash
shasum -a 256 -c checksums.sha256
```

## Requirements

- Claude Desktop must be installed
- A valid Slide API key from your [Slide account](https://console.slide.tech/)

## Support

For issues or questions, please visit: https://github.com/amcchord/slideMCP 
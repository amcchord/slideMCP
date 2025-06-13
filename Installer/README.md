# Slide MCP Installer

A cross-platform GUI installer for the Slide MCP Server that makes it easy to install and manage the Slide MCP Server integration with Claude Desktop.

## Features

- **Cross-Platform GUI**: Native desktop application for Windows, macOS, and Linux
- **Native macOS Integration**: Proper `.app` bundle with icon support, no terminal window
- **Automatic Detection**: Checks for Claude Desktop installation automatically
- **Smart Status Detection**: Detects existing installations and configurations
- **Easy Installation**: Download and install the slide-mcp-server binary with one click
- **Configuration Management**: Automatically configures Claude Desktop with your API key
- **API Key Management**: Shows current API key and allows easy updates
- **Uninstall Support**: Clean removal of the server and configuration
- **Progress Tracking**: Visual progress bar and status updates during installation

## Installation

### Download Pre-built Binary

Download the appropriate installer for your platform from the releases page:

- **macOS ARM64 (Apple Silicon)**: `slide-mcp-installer-darwin-arm64.tar.gz` (contains `.app` bundle)
- **macOS AMD64**: `slide-mcp-installer-darwin-amd64.tar.gz` (contains `.app` bundle)
- **Linux AMD64**: `slide-mcp-installer-linux-amd64.tar.gz`
- **Windows AMD64**: `slide-mcp-installer-windows-amd64.zip` (contains `.exe`)

### macOS Installation Notes

On macOS, the installer is packaged as a native `.app` bundle that:
- Displays the proper icon in Finder and the Dock
- Launches without opening a terminal window
- Integrates seamlessly with macOS

After downloading and extracting the `.tar.gz` file, simply drag the `slide-mcp-installer.app` to your Applications folder or run it directly.

### Build from Source

```bash
git clone https://github.com/austinmcchord/slide-mcp-server.git
cd slide-mcp-server/Installer
make build
```

## Usage

1. **Launch the installer**: Double-click the downloaded binary or run from terminal
2. **Check status**: The installer automatically detects:
   - Claude Desktop installation status
   - Whether slide-mcp-server is already installed
   - Current configuration and API key (if any)
3. **API Key management**:
   - If already configured, your current API key will be shown
   - You can update/change your API key and click "Update Configuration"
   - For new installations, enter your Slide API key in the password field
4. **Install/Update**: Click the appropriate button:
   - "Install Slide MCP Server" (new installation)
   - "Configure Slide MCP Server" (binary exists, needs configuration)
   - "Update Configuration" (update existing setup)
5. **Uninstall**: Click "Uninstall" to remove the server and configuration
6. **Restart Claude Desktop**: After any changes, restart Claude Desktop for changes to take effect

## What the Installer Does

### Installation Process

1. **Downloads the slide-mcp-server binary** from the latest GitHub release for your platform
2. **Installs the binary** to an appropriate location:
   - **macOS/Linux**: `~/.local/bin/slide-mcp-server`
   - **Windows**: `%LOCALAPPDATA%\slide-mcp\slide-mcp-server.exe`
3. **Updates Claude Desktop configuration** at:
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
   - **Linux**: `~/.config/Claude/claude_desktop_config.json`

### Configuration Added

The installer adds this configuration to your Claude Desktop:

```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

### Uninstallation

The uninstaller:
1. Removes the slide-mcp-server binary
2. Removes the "slide" entry from your Claude Desktop configuration
3. Preserves other MCP servers you may have configured

## Development

### Requirements

- Go 1.21+
- Fyne GUI framework dependencies

### Building for Development

```bash
# Install dependencies
make deps

# Run locally (for development/testing)
make run

# Build native app package for current platform
make build

# Build and open the app (macOS only)
make open

# Generate icon resource (if MCP-Installer.png changes)
make generate-icon

# Create release packages
make release
```

The `make build` command creates platform-specific packages:
- **macOS**: `.app` bundle with embedded icon
- **Windows**: `.exe` with embedded resources
- **Linux**: Standard binary

**Note**: GUI applications with Fyne require CGO and native compilation on each target platform. Cross-compilation is not straightforward due to OpenGL dependencies. For multi-platform releases, build natively on each target OS.

#### Icon Management

The installer includes a custom icon (`MCP-Installer.png`) that gets embedded into the binary. If you update the icon:

1. Replace `MCP-Installer.png` with your new icon
2. Run `make generate-icon` to regenerate the `icon.go` resource file
3. Rebuild the installer with `make build`

### Dependencies

- [Fyne](https://fyne.io/) - Cross-platform GUI framework
- [go-github](https://github.com/google/go-github) - GitHub API client

## API Key

You'll need a Slide API key to use the MCP server:

1. Log in to your [Slide account](https://console.slide.tech/)
2. Navigate to your account settings
3. Generate your API key from the API section

## Troubleshooting

### "Claude Desktop not found"

The installer looks for Claude Desktop in standard installation locations. If you have Claude Desktop installed in a custom location, the installer may not detect it, but you can still proceed with the installation.

### Installation Fails

- Ensure you have an internet connection to download the binary
- Check that you have write permissions to the installation directory
- Verify your API key is correct

### Claude Desktop Doesn't See the Server

- Restart Claude Desktop after installation
- Check that the configuration file was updated correctly
- Verify the binary was installed to the correct location

### Installer Shows Wrong Status

- Close and reopen the installer to refresh the detection
- Check file permissions on the Claude Desktop configuration directory
- Verify you're running the installer with the same user account that uses Claude Desktop

### API Key Not Updating

- Make sure to restart Claude Desktop after updating the API key
- Verify the new API key is valid by checking your Slide account
- If issues persist, use "Uninstall" then "Install" for a clean setup

## License

This installer is part of the Slide MCP Server project and is licensed under the MIT License. 
# Slide MCP Server

An MCP server implementation that integrates with the Slide API, providing comprehensive device and infrastructure management capabilities.

## 🚀 Go Binary Implementation ⚡
- **Single binary**: No dependencies, just download and run
- **Fast startup**: ~50ms startup time
- **Low memory usage**: 10-20MB memory footprint
- **Cross-platform**: Linux, macOS, Windows binaries
- **Zero Installation Hassle**: Simple download and configure

---

For quick setup instructions with Claude Desktop, see the installation section below.

## Features

- **Device Management**: List, update, and control devices with power operations
- **Agent Management**: Create, pair, and manage backup agents
- **Backup & Snapshot Management**: Initiate backups and manage snapshots
- **File Restore**: Browse and restore files from snapshots
- **Image Export**: Export snapshots as disk images (VHD, VHDX, Raw)
- **Virtual Machines**: Create and manage VMs from snapshots with browser-based console access
- **Network Management**: Create and manage isolated networks with VPN connections
- **User & Alert Management**: Monitor system alerts and manage users
- **Account & Client Management**: Organize resources by client accounts
- **Flexible Filtering & Pagination**: Control results with advanced filtering options

## Available Tools

### Core Management
- **Devices**: `list`, `get`, `update`, `poweroff`, `reboot`
- **Agents**: `list`, `get`, `create`, `pair`, `update`
- **Backups**: `list`, `get`, `start`
- **Snapshots**: `list`, `get`

### Data Recovery & Export
- **File Restores**: `list`, `get`, `create`, `delete`, `browse`
- **Image Exports**: `list`, `get`, `create`, `delete`, `browse`
  - Supports VHD, VHDX (dynamic/fixed), and Raw formats
  - Optional boot modifications (e.g., passwordless admin user)

### Virtual Machines
- **Virtual Machines**: `list`, `get`, `create`, `update`, `delete`
  - Browser-based VNC console access
  - Configurable CPU (1-16 cores) and RAM (1-12GB)
  - Multiple network modes and disk bus types

### Network Infrastructure
- **Networks**: `list`, `get`, `create`, `update`, `delete`
- **IPSec Connections**: `create`, `update`, `delete`
- **Port Forwarding**: `create`, `update`, `delete`
- **WireGuard Peers**: `create`, `update`, `delete`

### Administration
- **Users**: `list`, `get`
- **Alerts**: `list`, `get`, `update` (resolve)
- **Accounts**: `list`, `get`, `update` (alert emails)
- **Clients**: `list`, `get`, `create`, `update`, `delete`

All tools support pagination (`limit`, `offset`) and sorting options where applicable.

## 📦 Installation & Configuration

### Getting an API Key

1. Log in to your [Slide account](https://console.slide.tech/)
2. Navigate to your account settings
3. Generate your API key from the API section

## 🎯 Quick Setup with Claude Desktop

### 🖥️ GUI Installer (Recommended)

For the easiest installation experience, use our cross-platform GUI installer:

1. **Download the installer** for your platform:
   - **macOS ARM64 (Apple Silicon)**: `slide-mcp-installer-darwin-arm64`
   - **macOS AMD64**: `slide-mcp-installer-darwin-amd64` 
   - **Linux AMD64**: `slide-mcp-installer-linux-amd64`
   - **Windows AMD64**: `slide-mcp-installer-windows-amd64.exe`

2. **Run the installer**: Double-click or run from terminal
3. **Enter your API key**: Input your Slide API key when prompted
4. **Install**: Click "Install Slide MCP Server"
5. **Restart Claude Desktop**: The installer will configure everything automatically

The GUI installer will:
- ✅ Check if Claude Desktop is installed
- ✅ Download the latest slide-mcp-server binary
- ✅ Install it to the correct location
- ✅ Update your Claude Desktop configuration
- ✅ Provide easy uninstall option

### Manual Installation

#### Download Pre-built Binary (v1.14)
```bash
# For macOS ARM64 (Apple Silicon)
curl -L -o slide-mcp-server-v1.14-darwin-arm64.tar.gz https://github.com/austinmcchord/slide-mcp-server/releases/latest/download/slide-mcp-server-v1.14-darwin-arm64.tar.gz
tar -xzf slide-mcp-server-v1.14-darwin-arm64.tar.gz
mv slide-mcp-server-darwin-arm64 slide-mcp-server
chmod +x slide-mcp-server

# For macOS AMD64 
curl -L -o slide-mcp-server-v1.14-darwin-amd64.tar.gz https://github.com/austinmcchord/slide-mcp-server/releases/latest/download/slide-mcp-server-v1.14-darwin-amd64.tar.gz
tar -xzf slide-mcp-server-v1.14-darwin-amd64.tar.gz
mv slide-mcp-server-darwin-amd64 slide-mcp-server
chmod +x slide-mcp-server

# For Linux AMD64
curl -L -o slide-mcp-server-v1.14-linux-amd64.tar.gz https://github.com/austinmcchord/slide-mcp-server/releases/latest/download/slide-mcp-server-v1.14-linux-amd64.tar.gz
tar -xzf slide-mcp-server-v1.14-linux-amd64.tar.gz
mv slide-mcp-server-linux-amd64 slide-mcp-server
chmod +x slide-mcp-server

# For Windows AMD64
curl -L -o slide-mcp-server-v1.14-windows-amd64.zip https://github.com/austinmcchord/slide-mcp-server/releases/latest/download/slide-mcp-server-v1.14-windows-amd64.zip
unzip slide-mcp-server-v1.14-windows-amd64.zip
mv slide-mcp-server-windows-amd64.exe slide-mcp-server.exe
```

#### Build from Source
```bash
git clone https://github.com/austinmcchord/slide-mcp-server.git
cd slide-mcp-server
make build
# Binary will be in build/slide-mcp-server
```

#### Claude Desktop Configuration
Add this to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "YOUR_API_KEY_HERE"
      }
    }
  }
}
```

If installed system-wide:
```json
{
  "mcpServers": {
    "slide": {
      "command": "slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "YOUR_API_KEY_HERE"
      }
    }
  }
}
```

#### Test Your Installation
```bash
# Set your API key
export SLIDE_API_KEY="your-api-key-here"

# Test the server
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./slide-mcp-server

# Should respond with server info and capabilities
```

### Usage with VS Code

For VS Code integration, add the following JSON block to your User Settings (JSON) file. You can do this by pressing `Ctrl + Shift + P` and typing `Preferences: Open User Settings (JSON)`.

Optionally, you can add it to a file called `.vscode/mcp.json` in your workspace. This will allow you to share the configuration with others.

> Note that the `mcp` key is not needed in the `.vscode/mcp.json` file.

```json
{
  "mcp": {
    "inputs": [
      {
        "type": "promptString",
        "id": "slide_api_key",
        "description": "Slide API Key",
        "password": true
      }
    ],
    "servers": {
      "slide": {
        "command": "/path/to/slide-mcp-server",
        "env": {
          "SLIDE_API_KEY": "${input:slide_api_key}"
        }
      }
    }
  }
}
```

## Build

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create release packages
make release

# View available commands
make help
```

## License

This MCP server is licensed under the MIT License. This means you are free to use, modify, and distribute the software, subject to the terms and conditions of the MIT License. For more details, please see the LICENSE file in the project repository.


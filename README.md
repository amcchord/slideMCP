# Slide MCP Server

An MCP server implementation that integrates with the Slide API, providing comprehensive device and infrastructure management capabilities through a streamlined meta-tools architecture.

## 🚀 Go Binary Implementation ⚡
- **Single binary**: No dependencies, just download and run
- **Fast startup**: ~50ms startup time
- **Low memory usage**: 10-20MB memory footprint
- **Cross-platform**: Linux, macOS, Windows binaries
- **Zero Installation Hassle**: Simple download and configure
- **Streamlined Interface**: 10 meta-tools instead of 52+ individual tools for better LLM interaction

---

For quick setup instructions with Claude Desktop, see the installation section below.

## 🎯 Major Architecture Improvement

**Meta-Tools Design**: This MCP server uses an innovative meta-tools architecture that consolidates 52+ individual API operations into just **10 focused meta-tools**. This design significantly reduces complexity for LLMs while maintaining full functionality.

Each meta-tool accepts an `operation` parameter that specifies the action to perform, along with the relevant parameters for that operation.

### Example Usage Pattern
```json
{
  "name": "slide_devices",
  "arguments": {
    "operation": "list",
    "limit": 10,
    "client_id": "client-123"
  }
}
```

## Available Meta-Tools

### 🔧 Core Infrastructure
1. **`slide_devices`** - Physical device management
   - Operations: `list`, `get`, `update`, `poweroff`, `reboot`
   - Power control, hostname/display name updates, client assignment

2. **`slide_agents`** - Backup agent management  
   - Operations: `list`, `get`, `create`, `pair`, `update`
   - Agent creation, pairing with devices, display name management

3. **`slide_networks`** - Network infrastructure
   - Operations: `list`, `get`, `create`, `update`, `delete`
   - Network creation with DHCP, VPN support, client isolation
   - **IPSec**: `create_ipsec`, `update_ipsec`, `delete_ipsec`
   - **Port Forwarding**: `create_port_forward`, `update_port_forward`, `delete_port_forward`  
   - **WireGuard VPN**: `create_wg_peer`, `update_wg_peer`, `delete_wg_peer`

### 💾 Data Management
4. **`slide_backups`** - Backup operations
   - Operations: `list`, `get`, `start`
   - Initiate and monitor backup jobs

5. **`slide_snapshots`** - Snapshot management
   - Operations: `list`, `get`
   - Browse and access point-in-time snapshots with advanced filtering

6. **`slide_restores`** - File & image restoration
   - **File Restores**: `list_files`, `get_file`, `create_file`, `delete_file`, `browse_file`
   - **Image Exports**: `list_images`, `get_image`, `create_image`, `delete_image`, `browse_image`
   - Support for VHD, VHDX (dynamic/fixed), and Raw disk formats
   - Optional boot modifications (passwordless admin user)

### ☁️ Virtual Infrastructure  
7. **`slide_vms`** - Virtual machine management
   - Operations: `list`, `get`, `create`, `update`, `delete`
   - Browser-based VNC console access
   - Configurable CPU (1-16 cores) and RAM (1-12GB)
   - Multiple network modes and disk bus types

### 👥 Administration
8. **`slide_users`** - User management
   - Operations: `list`, `get`
   - User account information and permissions

9. **`slide_alerts`** - Alert monitoring
   - Operations: `list`, `get`, `update` (resolve)
   - System alert management and resolution

10. **`slide_accounts`** - Account & client organization
    - **Accounts**: `list_accounts`, `get_account`, `update_account` (alert emails)
    - **Clients**: `list_clients`, `get_client`, `create_client`, `update_client`, `delete_client`
    - Organize resources by client, manage alert notifications

### 🔍 Special Tools
- **`list_all_clients_devices_and_agents`** - Hierarchical overview
  - Get complete view of all clients, their devices, and agents in one call
  - Perfect for answering questions about infrastructure scale and organization

## Key Features

### 🔐 Infrastructure Management
- **Device Control**: Remote power operations, hostname management, client assignment
- **Agent Deployment**: Automated pairing, display name management  
- **Network Isolation**: Client-specific networks with VPN access
- **Advanced Networking**: IPSec tunnels, port forwarding, WireGuard peers

### 💽 Data Protection & Recovery
- **Automated Backups**: Agent-based backup initiation and monitoring
- **Point-in-Time Recovery**: Snapshot browsing with location filtering
- **Flexible Restores**: File-level and full disk image exports
- **Multiple Formats**: VHD, VHDX (dynamic/fixed), Raw disk images
- **Boot Modifications**: Optional passwordless admin account creation

### ☁️ Virtualization
- **VM Creation**: Create VMs from any snapshot
- **Resource Control**: Configurable CPU/RAM allocation  
- **Network Integration**: Connect VMs to isolated networks
- **Console Access**: Browser-based VNC for direct VM interaction

### 📊 Monitoring & Organization
- **Alert Management**: Centralized alert monitoring and resolution
- **Client Organization**: Group resources by client for better management
- **User Management**: Account access and permissions
- **Comprehensive Filtering**: Advanced pagination and sorting across all resources

All meta-tools support pagination (`limit`, `offset`) and sorting options where applicable.

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

#### Download Pre-built Binary (v1.17.3)
```bash
# For macOS ARM64 (Apple Silicon)
curl -L -o slide-mcp-server-v1.17.3-macos-arm64.tar.gz https://github.com/austinmcchord/slide-mcp-server/releases/latest/download/slide-mcp-server-v1.17.3-macos-arm64.tar.gz
tar -xzf slide-mcp-server-v1.17.3-macos-arm64.tar.gz
chmod +x slide-mcp-server-v1.17.3-macos-arm64
mv slide-mcp-server-v1.17.3-macos-arm64 slide-mcp-server

# For macOS AMD64 
curl -L -o slide-mcp-server-v1.17.3-macos-x64.tar.gz https://github.com/austinmcchord/slide-mcp-server/releases/latest/download/slide-mcp-server-v1.17.3-macos-x64.tar.gz
tar -xzf slide-mcp-server-v1.17.3-macos-x64.tar.gz
chmod +x slide-mcp-server-v1.17.3-macos-x64
mv slide-mcp-server-v1.17.3-macos-x64 slide-mcp-server

# For Linux AMD64
curl -L -o slide-mcp-server-v1.17.3-linux-x64.tar.gz https://github.com/austinmcchord/slide-mcp-server/releases/latest/download/slide-mcp-server-v1.17.3-linux-x64.tar.gz
tar -xzf slide-mcp-server-v1.17.3-linux-x64.tar.gz
chmod +x slide-mcp-server-v1.17.3-linux-x64
mv slide-mcp-server-v1.17.3-linux-x64 slide-mcp-server

# For Linux ARM64
curl -L -o slide-mcp-server-v1.17.3-linux-arm64.tar.gz https://github.com/austinmcchord/slide-mcp-server/releases/latest/download/slide-mcp-server-v1.17.3-linux-arm64.tar.gz
tar -xzf slide-mcp-server-v1.17.3-linux-arm64.tar.gz
chmod +x slide-mcp-server-v1.17.3-linux-arm64
mv slide-mcp-server-v1.17.3-linux-arm64 slide-mcp-server

# For Windows AMD64
curl -L -o slide-mcp-server-v1.17.3-windows-x64.zip https://github.com/austinmcchord/slide-mcp-server/releases/latest/download/slide-mcp-server-v1.17.3-windows-x64.zip
unzip slide-mcp-server-v1.17.3-windows-x64.zip
mv slide-mcp-server-v1.17.3-windows-x64.exe slide-mcp-server.exe
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

## 🔧 CLI Arguments & Configuration

The Slide MCP Server supports several command-line arguments for flexible configuration:

### Command Line Arguments

```bash
# Basic usage with API key
./slide-mcp-server --api-key YOUR_API_KEY

# All available flags
./slide-mcp-server [OPTIONS]
```

| Flag | Description | Environment Variable | Default |
|------|-------------|---------------------|---------|
| `--api-key` | Slide API key for authentication | `SLIDE_API_KEY` | Required |
| `--base-url` | Base URL for Slide API endpoint | `SLIDE_BASE_URL` | `https://api.slide.tech` |
| `--tools` | Permission mode for tool access | `SLIDE_TOOLS` | `full-safe` |
| `--version` | Show version information and exit | - | - |

**Priority**: CLI flags take precedence over environment variables.

### Examples

```bash
# Using CLI flags
./slide-mcp-server --api-key sk_test_123 --base-url https://custom.api.endpoint --tools reporting

# Using environment variables
export SLIDE_API_KEY="sk_test_123"
export SLIDE_BASE_URL="https://custom.api.endpoint" 
export SLIDE_TOOLS="reporting"
./slide-mcp-server

# Mixed usage (CLI overrides environment)
export SLIDE_TOOLS="full"
./slide-mcp-server --api-key sk_test_123 --tools reporting  # Uses reporting mode

# Show version
./slide-mcp-server --version
# Output: slide-mcp-server version 1.2.5
```

## 🔒 Permission Modes

The server includes a sophisticated permission system with four distinct access levels:

### Permission Levels

#### `reporting` - Read-Only Access
**Use Case**: Monitoring, reporting, and dashboard integrations
- ✅ **Allowed**: All read operations (`list`, `get`, `browse`)
- ❌ **Blocked**: All create, update, delete operations
- ❌ **Blocked**: Power control operations

```bash
./slide-mcp-server --api-key YOUR_KEY --tools reporting
```

#### `restores` - Data Recovery & VM Management  
**Use Case**: IT support teams performing data recovery and VM management
- ✅ **Allowed**: All reporting operations
- ✅ **Allowed**: VM management (create, update, delete)
- ✅ **Allowed**: File restore operations
- ✅ **Allowed**: Image export operations  
- ✅ **Allowed**: Network management
- ✅ **Allowed**: Device management and power control
- ✅ **Allowed**: Account/client management
- ✅ **Allowed**: Backup management and alert resolution
- ❌ **Blocked**: Agent deletion
- ❌ **Blocked**: Snapshot deletion

```bash
./slide-mcp-server --api-key YOUR_KEY --tools restores
```

#### `full-safe` - Comprehensive Access (Default)
**Use Case**: General administration with safety guardrails
- ✅ **Allowed**: All operations except dangerous ones
- ❌ **Blocked**: Agent deletion (prevents accidental backup disruption)
- ❌ **Blocked**: Snapshot deletion (prevents data loss)

```bash
./slide-mcp-server --api-key YOUR_KEY --tools full-safe
# OR simply (default mode)
./slide-mcp-server --api-key YOUR_KEY
```

#### `full` - Complete Access
**Use Case**: Advanced administrators who need unrestricted access
- ✅ **Allowed**: All operations including dangerous ones
- ⚠️ **Warning**: Includes agent and snapshot deletion

```bash
./slide-mcp-server --api-key YOUR_KEY --tools full
```

### Permission Matrix

| Operation Category | `reporting` | `restores` | `full-safe` | `full` |
|--------------------|-------------|------------|-------------|--------|
| List/Get/Browse | ✅ | ✅ | ✅ | ✅ |
| Device Power Control | ❌ | ✅ | ✅ | ✅ |
| VM Management | ❌ | ✅ | ✅ | ✅ |
| Network Management | ❌ | ✅ | ✅ | ✅ |
| File Restores | ❌ | ✅ | ✅ | ✅ |
| Image Exports | ❌ | ✅ | ✅ | ✅ |
| Backup Jobs | ❌ | ✅ | ✅ | ✅ |
| Account Management | ❌ | ✅ | ✅ | ✅ |
| Alert Resolution | ❌ | ✅ | ✅ | ✅ |
| Agent Creation/Updates | ❌ | ✅ | ✅ | ✅ |
| Agent Deletion | ❌ | ❌ | ❌ | ✅ |
| Snapshot Deletion | ❌ | ❌ | ❌ | ✅ |

### Security Recommendations

- **Production Monitoring**: Use `reporting` mode for read-only dashboards and monitoring systems
- **Support Teams**: Use `restores` mode for IT support staff performing data recovery
- **General Administration**: Use `full-safe` mode (default) for most administrative tasks
- **Advanced Users Only**: Use `full` mode only when agent or snapshot deletion is specifically required

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

## 💡 Usage Examples

### List All Devices
```json
{
  "name": "slide_devices",
  "arguments": {
    "operation": "list",
    "limit": 20,
    "client_id": "client-123"
  }
}
```

### Create a Network with VPN
```json
{
  "name": "slide_networks",
  "arguments": {
    "operation": "create",
    "name": "Development Network",
    "type": "standard",
    "router_prefix": "192.168.100.1/24",
    "dhcp": true,
    "dhcp_range_start": "192.168.100.10",
    "dhcp_range_end": "192.168.100.200",
    "wg": true,
    "wg_prefix": "10.100.0.0/24",
    "client_id": "client-123"
  }
}
```

### Create VM from Snapshot
```json
{
  "name": "slide_vms",
  "arguments": {
    "operation": "create",
    "snapshot_id": "snapshot-456",
    "device_id": "device-789",
    "cpu_count": 4,
    "memory_in_mb": 8192,
    "network_type": "network-id",
    "network_source": "network-123"
  }
}
```

### Start Backup Job
```json
{
  "name": "slide_backups",
  "arguments": {
    "operation": "start",
    "agent_id": "agent-456"
  }
}
```

### Export Snapshot as VHD Image
```json
{
  "name": "slide_restores",
  "arguments": {
    "operation": "create_image",
    "snapshot_id": "snapshot-789",
    "device_id": "device-123",
    "image_type": "vhd-dynamic",
    "boot_remove_passwords": true
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

## Architecture Benefits

### For LLMs
- **Reduced Complexity**: 10 meta-tools vs 52+ individual tools
- **Logical Grouping**: Related operations organized together
- **Consistent Interface**: All meta-tools follow the same operation pattern
- **Better Context**: Less tool switching, more focused conversations

### For Developers  
- **Maintainable**: Each meta-tool in its own file
- **Extensible**: Easy to add new operations to existing categories
- **Backward Compatible**: All original functionality preserved
- **Schema Validation**: Conditional parameter validation per operation

## License

This MCP server is licensed under the MIT License. This means you are free to use, modify, and distribute the software, subject to the terms and conditions of the MIT License. For more details, please see the LICENSE file in the project repository.


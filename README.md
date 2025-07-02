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

### 📊 Data Presentation & Reporting
11. **`slide_presentation`** - Professional data formatting and documentation
    - **Operations**: `get_card`, `get_runbook_template`, `get_daily_report_template`, `get_monthly_report_template`
    - **Card Types**: Individual item cards (agent, client, device, snapshot) and table cards (agents_table, clients_table, etc.)
    - **Report Templates**: Runbook procedures, daily activity summaries, monthly analysis reports
    - **Formats**: HTML, Markdown, HAML support for multiple output needs
    - Perfect for status displays, dashboards, documentation, and professional reporting
**⚠️ IMPORTANT**: If you are building your own presentation logic or custom formatting, you may want to disable the `slide_presentation` tool to avoid conflicts with your custom implementation. To disable just the `slide_presentation` tool, add it to the `DISABLED_TOOLS` environment variable or in the --disabled-tools part of the CLI

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

### 📋 Professional Data Presentation
- **Smart Cards**: Individual and table-based cards for agents, clients, devices, and snapshots
- **Report Templates**: Runbook procedures, daily summaries, and monthly analysis reports
- **Multiple Formats**: HTML, Markdown, and HAML output for different use cases
- **Dashboard Ready**: Pre-formatted cards perfect for status displays and monitoring
- **Documentation Support**: Professional templates for operational procedures and troubleshooting

All meta-tools support pagination (`limit`, `offset`) and sorting options where applicable.

## 📊 Data Presentation Tool Guide

The **`slide_presentation`** tool is your primary resource for professional data formatting and documentation. It provides pre-built templates and smart cards that transform raw data into polished, readable formats.

### 🎯 When to Use the Presentation Tool

**Always consider this tool first** when you need to:
- Display system status or monitoring data to users
- Show lists of items (agents, clients, devices, snapshots)
- Present individual item details in a structured format
- Create reports or summaries
- Generate documentation or procedures
- Format any data that could benefit from professional presentation

### 📋 Report Templates

Perfect for comprehensive documentation and analysis:

#### **Runbook Templates** (`get_runbook_template`)
- **Purpose**: Operational procedures, troubleshooting guides, step-by-step instructions
- **Use Cases**: Incident response, maintenance procedures, troubleshooting guides
- **Formats**: HTML, Markdown, HAML

#### **Daily Report Templates** (`get_daily_report_template`)
- **Purpose**: Activity summaries, status updates, end-of-day reports
- **Use Cases**: Daily operational summaries, status briefings, activity tracking
- **Formats**: HTML (default), Markdown, HAML

#### **Monthly Report Templates** (`get_monthly_report_template`)
- **Purpose**: Comprehensive analysis, trends, monthly summaries
- **Use Cases**: Executive summaries, trend analysis, performance reviews
- **Formats**: HTML (default), Markdown, HAML

### 📊 Smart Cards

Perfect for status displays, dashboards, and data visualization:

#### **Single Item Cards** - Detailed Views
- **`agent`**: Individual backup agent with hostname, OS, status, recent backups
- **`client`**: Individual client with name, agent count, device assignments, stats
- **`device`**: Individual backup device with capacity, assignments, storage info
- **`snapshot`**: Individual backup snapshot with date, size, status, retention

#### **Table Cards** - Overview Dashboards
- **`agents_table`**: Multiple agents comparison with status overview and assignments
- **`clients_table`**: Multiple clients summary with agent counts and status
- **`devices_table`**: Multiple devices overview with capacity and utilization
- **`snapshots_table`**: Chronological backup history with sizes and status

### 💡 Decision Guide

Choose the right presentation format based on your needs:

| Need | Recommendation | Example |
|------|----------------|---------|
| Show ONE item in detail | Single item cards | `agent`, `client`, `device`, `snapshot` |
| Show MULTIPLE items overview | Table cards | `agents_table`, `clients_table`, `devices_table` |
| Create documentation | Report templates | `get_runbook_template` |
| Generate status reports | Daily/Monthly templates | `get_daily_report_template` |
| Dashboard display | Table cards | `agents_table`, `devices_table` |
| Troubleshooting guide | Runbook template | `get_runbook_template` |

### 🚀 Best Practices

1. **Start with Presentation**: Always consider using the presentation tool before displaying raw data
2. **Choose the Right Card**: Use single cards for details, table cards for overviews
3. **Format for Purpose**: Use HTML for web displays, Markdown for documentation
4. **Professional Output**: Let the tool handle formatting instead of manual formatting
5. **Consistent Experience**: Use cards for a consistent look and feel across all data displays

## 📦 Installation & Configuration

### Getting an API Key

1. Log in to your [Slide account](https://console.slide.tech/)
2. Navigate to your account settings
3. Generate your API key from the API section

## 🎯 Quick Setup with Claude Desktop

### 🖥️ GUI Installer (macOS Only)

**Note: Currently, there is only a GUI installer available for macOS. Other platforms must use manual installation.**

For the easiest installation experience on macOS, use our GUI installer:

1. **Download the Mac installer**: `SlideMCP-MacInstaller.zip` from the [latest release](https://github.com/amcchord/slideMCP/releases/latest)

2. **Extract and run**: Unzip the installer and double-click to run
3. **Enter your API key**: Input your Slide API key when prompted
4. **Install**: Click "Install Slide MCP Server"
5. **Restart Claude Desktop**: The installer will configure everything automatically

The Mac GUI installer will:
- ✅ Check if Claude Desktop is installed
- ✅ Download the latest slide-mcp-server binary
- ✅ Install it to the correct location
- ✅ Update your Claude Desktop configuration
- ✅ Provide easy uninstall option

### Manual Installation

#### Download Pre-built Binary (v2.0.2)
```bash
# For macOS ARM64 (Apple Silicon)
curl -L -o slide-mcp-server-v2.0.2-macos-arm64.tar.gz https://github.com/amcchord/slideMCP/releases/latest/download/slide-mcp-server-v2.0.2-macos-arm64.tar.gz
tar -xzf slide-mcp-server-v2.0.2-macos-arm64.tar.gz
chmod +x slide-mcp-server-v2.0.2-macos-arm64
mv slide-mcp-server-v2.0.2-macos-arm64 slide-mcp-server

# For macOS AMD64 
curl -L -o slide-mcp-server-v2.0.2-macos-x64.tar.gz https://github.com/amcchord/slideMCP/releases/latest/download/slide-mcp-server-v2.0.2-macos-x64.tar.gz
tar -xzf slide-mcp-server-v2.0.2-macos-x64.tar.gz
chmod +x slide-mcp-server-v2.0.2-macos-x64
mv slide-mcp-server-v2.0.2-macos-x64 slide-mcp-server

# For Linux AMD64
curl -L -o slide-mcp-server-v2.0.2-linux-x64.tar.gz https://github.com/amcchord/slideMCP/releases/latest/download/slide-mcp-server-v2.0.2-linux-x64.tar.gz
tar -xzf slide-mcp-server-v2.0.2-linux-x64.tar.gz
chmod +x slide-mcp-server-v2.0.2-linux-x64
mv slide-mcp-server-v2.0.2-linux-x64 slide-mcp-server

# For Linux ARM64
curl -L -o slide-mcp-server-v2.0.2-linux-arm64.tar.gz https://github.com/amcchord/slideMCP/releases/latest/download/slide-mcp-server-v2.0.2-linux-arm64.tar.gz
tar -xzf slide-mcp-server-v2.0.2-linux-arm64.tar.gz
chmod +x slide-mcp-server-v2.0.2-linux-arm64
mv slide-mcp-server-v2.0.2-linux-arm64 slide-mcp-server

# For Windows AMD64
curl -L -o slide-mcp-server-v2.0.2-windows-x64.zip https://github.com/amcchord/slideMCP/releases/latest/download/slide-mcp-server-v2.0.2-windows-x64.zip
unzip slide-mcp-server-v2.0.2-windows-x64.zip
mv slide-mcp-server-v2.0.2-windows-x64.exe slide-mcp-server.exe
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

With custom permission mode and disabled tools:
```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "YOUR_API_KEY_HERE",
        "SLIDE_TOOLS": "reporting",
        "SLIDE_DISABLED_TOOLS": "slide_accounts,slide_users"
      }
    }
  }
}
```

Or using CLI arguments:
```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "args": [
        "--api-key", "YOUR_API_KEY_HERE",
        "--tools", "full-safe", 
        "--disabled-tools", "slide_agents,slide_backups"
      ]
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
| `--disabled-tools` | Comma-separated list of tools to disable | `SLIDE_DISABLED_TOOLS` | None |
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

# Disable specific tools
./slide-mcp-server --api-key sk_test_123 --disabled-tools "slide_agents,slide_backups"

# Show version
./slide-mcp-server --version
# Output: slide-mcp-server version 2.0.2
```

### 🚫 Disabling Specific Tools

In addition to permission modes, you can disable specific tools entirely using the `--disabled-tools` flag or `SLIDE_DISABLED_TOOLS` environment variable. This provides fine-grained control over which tools are available.

#### Usage Examples

```bash
# Disable specific tools via CLI flag
./slide-mcp-server --api-key YOUR_KEY --disabled-tools "slide_agents,slide_backups"

# Disable tools via environment variable
export SLIDE_DISABLED_TOOLS="slide_devices,slide_users"
./slide-mcp-server --api-key YOUR_KEY

# Combined with permission modes
./slide-mcp-server --api-key YOUR_KEY --tools reporting --disabled-tools "slide_snapshots"

# Whitespace is handled gracefully
./slide-mcp-server --api-key YOUR_KEY --disabled-tools " slide_agents , slide_backups , slide_devices "
```

#### Available Tool Names
- `slide_agents` - Agent management
- `slide_backups` - Backup operations  
- `slide_snapshots` - Snapshot management
- `slide_restores` - File and image restoration
- `slide_networks` - Network management
- `slide_users` - User management
- `slide_alerts` - Alert monitoring
- `slide_accounts` - Account and client management
- `slide_devices` - Device management
- `slide_vms` - Virtual machine management
- `slide_presentation` - Data presentation and reporting
- `list_all_clients_devices_and_agents` - Hierarchical overview

#### Key Features

- **Precedence**: CLI flags take precedence over environment variables
- **Whitespace Handling**: Extra spaces around tool names are automatically trimmed
- **Error Messages**: Clear error messages when attempting to use disabled tools
- **Combined Filtering**: Works alongside permission modes for layered access control
- **Transparency**: Logs which tools are disabled on server startup

#### Use Cases

```bash
# Create a read-only server that can't access sensitive data
./slide-mcp-server --tools reporting --disabled-tools "slide_accounts,slide_users"

# Allow restores but prevent network changes
./slide-mcp-server --tools restores --disabled-tools "slide_networks"

# Monitoring setup that excludes VM management
./slide-mcp-server --tools reporting --disabled-tools "slide_vms,slide_networks"
```

When a disabled tool is called, the server returns:
```json
{
  "error": {
    "code": -32601,
    "message": "Tool 'slide_agents' is disabled"
  }
}
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
- ✅ **Allowed**: Device management (updates only)
- ✅ **Allowed**: Agent management (create, pair, update)
- ✅ **Allowed**: Backup management
- ❌ **Blocked**: Device power control (poweroff, reboot)
- ❌ **Blocked**: Account/client management
- ❌ **Blocked**: Alert resolution
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
- ❌ **Blocked**: Device power control (prevents accidental shutdowns)

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
| Device Power Control | ❌ | ❌ | ❌ | ✅ |
| VM Management | ❌ | ✅ | ✅ | ✅ |
| Network Management | ❌ | ✅ | ✅ | ✅ |
| File Restores | ❌ | ✅ | ✅ | ✅ |
| Image Exports | ❌ | ✅ | ✅ | ✅ |
| Backup Jobs | ❌ | ✅ | ✅ | ✅ |
| Account Management | ❌ | ❌ | ✅ | ✅ |
| Alert Resolution | ❌ | ❌ | ✅ | ✅ |
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

With custom configuration and disabled tools:
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
          "SLIDE_API_KEY": "${input:slide_api_key}",
          "SLIDE_TOOLS": "reporting",
          "SLIDE_DISABLED_TOOLS": "slide_accounts,slide_users"
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

### Display Agent Status Card
```json
{
  "name": "slide_presentation",
  "arguments": {
    "operation": "get_card",
    "card_type": "agent"
  }
}
```

### Generate Multiple Devices Overview
```json
{
  "name": "slide_presentation",
  "arguments": {
    "operation": "get_card",
    "card_type": "devices_table"
  }
}
```

### Create Runbook Template
```json
{
  "name": "slide_presentation",
  "arguments": {
    "operation": "get_runbook_template",
    "format": "markdown"
  }
}
```

### Generate Daily Report Template
```json
{
  "name": "slide_presentation",
  "arguments": {
    "operation": "get_daily_report_template",
    "format": "html"
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


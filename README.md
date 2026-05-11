# Slide MCP Server

An MCP server implementation that integrates with the Slide API, providing
comprehensive device and infrastructure management capabilities through a
streamlined meta-tools architecture.

## What's new in v3.1.0

- **Drag-and-drop install for Claude Desktop**. One signed, notarized
  `slide-mcp-server.mcpb` works on every platform - the right binary
  (macOS universal / Linux amd64 + arm64 / Windows amd64) is selected at
  runtime via the manifest's `platform_overrides`.
- **Retired the legacy Fyne GUI installer.** The `.mcpb` bundle plus a
  one-line `claude mcp add` does everything the GUI did, with less to
  download and no CGO/OpenGL build hell.
- **Slide API token is owned by Claude Desktop's `user_config` UI**, not
  written into `claude_desktop_config.json` by us anymore. Power users
  / CI / Cursor still get `SLIDE_API_KEY` env + `--api-key` flag.
- See v3.0.0 for the previous big release (SDK migration + Slide API v1.27.0
  coverage). Full history in [CHANGELOG.md](CHANGELOG.md).

## 🚀 Go Binary Implementation ⚡
- **Single binary**: No dependencies, just download and run
- **Fast startup**: ~50ms startup time
- **Low memory usage**: 10-20MB memory footprint
- **Cross-platform**: Linux, macOS (Apple Silicon + Intel), Windows
- **Streamlined Interface**: 13 meta-tools instead of 70+ individual tools for better LLM interaction
- **Initial Context Resource**: Cached `slide://context/clients-devices-agents` resource
  for fast first-turn answers

---

## ⚡ Install

Pick the path for your host. Each is one step.

### 🟢 Claude Desktop - drag and drop (recommended)

1. Download **[slide-mcp-server.mcpb](https://github.com/austinmcchord/slideMCP/releases/latest/download/slide-mcp-server.mcpb)**.
2. Drag it onto Claude Desktop's **Settings -> Extensions** screen.
3. Paste your Slide API token when prompted (generate one at
   [console.slide.tech](https://console.slide.tech/) under
   **My Settings -> API Tokens**).

That's it. Claude Desktop owns the token storage, picks the right binary
for your OS, and starts the server. The bundle is signed and notarized,
so Gatekeeper won't get in the way on macOS.

### 🟦 Claude Code

```bash
claude mcp add slide --env SLIDE_API_KEY=tk_... -- /path/to/slide-mcp-server
```

Or, from a clone of this repo:

```bash
make build
SLIDE_API_KEY=tk_... make install-claude-code
```

The wrapper is idempotent - re-running it overwrites the prior `slide`
registration cleanly.

### 🟪 Cursor / other MCP hosts

Drop the binary somewhere on `$PATH` and add this to `~/.cursor/mcp.json`
(or your host's equivalent):

```json
{
  "mcpServers": {
    "slide": {
      "command": "slide-mcp-server",
      "env": { "SLIDE_API_KEY": "${SLIDE_API_KEY}" }
    }
  }
}
```

Claude Code, Cursor, and most other modern hosts support `${VAR}`
expansion so the literal token never has to live in the config file.

### 🛠️ From source / on a new dev box

```bash
git clone https://github.com/austinmcchord/slideMCP.git && cd slideMCP
make setup-dev   # idempotently installs go, jq, gh via Homebrew
make build       # produces build/slide-mcp-server
make pack-dxt    # produces build/slide-mcp-server.mcpb (unsigned, dev)
```

### 🔏 Code signing setup (release builds only)

`make pack-dxt-signed` (and the release scripts) need a valid Apple
Developer ID Application certificate plus a notarytool keychain profile.
A first-time setup helper handles the whole dance - generates a CSR,
opens the Apple Developer portal in your browser, imports the resulting
certificate, stores notarytool credentials, and writes
`scripts/release-env.sh` (gitignored) so subsequent builds Just Work:

```bash
./scripts/setup-signing.sh    # interactive, idempotent, ~5 minutes
make doctor-signing            # read-only check; run before each release
```

Once setup is done, on every new shell:

```bash
source scripts/release-env.sh   # exports DEVELOPER_ID, KEYCHAIN_PROFILE, etc.
make pack-dxt-signed            # signed + notarized .mcpb in build/
make verify-dxt                 # confirms Gatekeeper accepts the inner binary
```

The setup script is **not tracked in git** (covered by the existing
`**/setup-signing.sh` rule) and never writes plaintext secrets outside
of the local `scripts/release-env.sh` (also gitignored) and your macOS
keychain.

**Read [`docs/SIGNING.md`](docs/SIGNING.md) before touching anything in
the signing pipeline.** It documents 10+ macOS-specific quirks we
already debugged the hard way (DER vs PEM, PKCS#1 vs PKCS#8, why .p12
imports fail on modern macOS, why `add-trusted-cert -r trustAsRoot`
breaks chain validation, why `spctl --type execute` always rejects raw
Mach-O even when notarized, and more). LLMs working on this repo: also
see [`AGENTS.md`](AGENTS.md).

---

For more setup detail (CLI flags, environment variables, manual JSON snippets),
see the installation section further down.

## 🎯 Major Architecture Improvement

**Meta-Tools Design**: This MCP server uses an innovative meta-tools architecture that consolidates 52+ individual API operations into just **13 focused meta-tools**. This design significantly reduces complexity for LLMs while maintaining full functionality.

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
   - **Push Restores**: `list_pushes`, `create_push`, `update_push` - Restore files directly to protected systems (destination must be `X:\SlideRestore`)
   - **Image Exports**: `list_images`, `get_image`, `create_image`, `delete_image`, `browse_image`
   - Support for VHD, VHDX (dynamic/fixed), and Raw disk formats
   - Optional boot modifications (passwordless admin user)

### ☁️ Virtual Infrastructure  
7. **`slide_vms`** - Virtual machine management
   - Operations: `list`, `get`, `create`, `update`, `delete`, `get_rdp_bookmark`
   - Browser-based VNC console access and downloadable RDP bookmarks
   - Configurable CPU (1-16 cores) and RAM (1-12GB)
   - Multiple network modes and disk bus types

### 👥 Administration
8. **`slide_user_management`** - User and account management
   - **Users**: `list_users`, `get_user` - User account information and permissions
   - **Accounts**: `list_accounts`, `get_account`, `update_account` - Account settings and alert email configuration
   - **Clients**: `list_clients`, `get_client`, `create_client`, `update_client`, `delete_client` - Client organization and resource management

9. **`slide_alerts`** - Alert monitoring
   - Operations: `list`, `get`, `update` (resolve)
   - System alert management and resolution

### 📊 Data Presentation & Reporting
10. **`slide_presentation`** - Professional data formatting and documentation
    - **Operations**: `get_card`, `get_runbook_template`, `get_daily_report_template`, `get_monthly_report_template`
    - **Card Types**: Individual item cards (agent, client, device, snapshot) and table cards (agents_table, clients_table, etc.)
    - **Report Templates**: Runbook procedures, daily activity summaries, monthly analysis reports
    - **Formats**: HTML, Markdown, HAML support for multiple output needs
    - Perfect for status displays, dashboards, documentation, and professional reporting
    - **⚠️ DISABLED BY DEFAULT**: Must be explicitly enabled with `--enable-presentation` flag or `SLIDE_ENABLE_PRESENTATION=true`
**⚠️ IMPORTANT**: If you are building your own presentation logic or custom formatting, you may want to disable the `slide_presentation` tool to avoid conflicts with your custom implementation. To disable just the `slide_presentation` tool, add it to the `DISABLED_TOOLS` environment variable or in the --disabled-tools part of the CLI

11. **`slide_meta`** - Meta tools for reporting and aggregated data views
    - **Operations**: `list_all_clients_devices_and_agents`, `get_snapshot_changes`, `get_reporting_data`
    - **list_all_clients_devices_and_agents**: Complete hierarchical view of infrastructure
    - **get_snapshot_changes**: Track new and deleted snapshots over time periods (day, week, month)
    - **get_reporting_data**: Pre-formatted data for populating report templates
    - Perfect for generating reports with accurate, pre-calculated metrics

12. **`slide_reports`** - Pre-calculated statistics and reports for backup/snapshot analysis
    - **⚠️ DISABLED BY DEFAULT**: Must be explicitly enabled with `--enable-reports` flag or `SLIDE_ENABLE_REPORTS=true`
    - **Operations**: `daily_backup_snapshot`, `weekly_backup_snapshot`, `monthly_backup_snapshot`
    - **Daily Reports**: Single day statistics with backup success rates and failure reasons
    - **Weekly Reports**: 7-day breakdown with daily agent counts and success metrics
    - **Monthly Reports**: Full month analysis with visual calendar view (in markdown format)
    - **Filtering**: By agent_id, device_id, or client_id for targeted reporting
    - **Formats**: JSON (structured data) or Markdown (human-readable)
    - **Performance**: Use verbose mode to track progress on large reports

13. **`slide_docs`** - Access to official Slide documentation
    - **Operations**: `list_sections`, `get_topics`, `search_docs`, `get_content`, `get_api_reference`
    - **Documentation Access**: Browse and search docs.slide.tech content directly
    - **Contextual Help**: Get best practices, troubleshooting guidance, and feature explanations
    - **API Reference**: Quick access to API endpoint documentation
    - **Integration**: Complements other tools by providing context and guidance

### 🔍 Special Tools
- **`list_all_clients_devices_and_agents`** - Hierarchical overview (now part of `slide_meta`)
  - Get complete view of all clients, their devices, and agents in one call
  - Perfect for answering questions about infrastructure scale and organization
  - Can be called directly or via `slide_meta` with operation `list_all_clients_devices_and_agents`

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
- **RDP Bookmarks**: Generate downloadable .rdp files for easy Windows Remote Desktop access

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

## 📦 Detailed installation reference

The "⚡ Install" section near the top of this README covers the
recommended paths. This section documents the underlying configuration so
you can hand-roll setups for unusual hosts or CI.

### Getting an API Key

1. Log in to your [Slide account](https://console.slide.tech/)
2. Navigate to **My Settings -> API Tokens**
3. Generate your API token. Treat it like a password - it grants full
   access to your Slide account.

### Pre-built downloads

The headline release asset is the `.mcpb` (Claude Desktop Extension):

- **slide-mcp-server.mcpb** (one file, every platform):
  `https://github.com/austinmcchord/slideMCP/releases/latest/download/slide-mcp-server.mcpb`

Standalone binaries (for `claude mcp add` / Cursor / CI):

- **macOS** (universal Apple Silicon + Intel): `slide-mcp-server-vX.Y.Z-darwin-universal.tar.gz`
- **Linux x64**: `slide-mcp-server-vX.Y.Z-linux-amd64.tar.gz`
- **Linux ARM64**: `slide-mcp-server-vX.Y.Z-linux-arm64.tar.gz`
- **Windows x64**: `slide-mcp-server-vX.Y.Z-windows-x64.zip`

All macOS artifacts are signed with a Developer ID and notarized.

### Build from source

```bash
git clone https://github.com/austinmcchord/slideMCP.git
cd slideMCP
make build      # build/slide-mcp-server (current platform)
make pack-dxt   # build/slide-mcp-server.mcpb (universal bundle, unsigned)
```

### Manual Claude Desktop config

If you can't use the `.mcpb` (older Claude Desktop, locked-down corp
machine, etc.), edit `claude_desktop_config.json` directly:

```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "env": { "SLIDE_API_KEY": "YOUR_API_KEY_HERE" }
    }
  }
}
```

With non-default tool mode and disabled tools:

```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "YOUR_API_KEY_HERE",
        "SLIDE_TOOLS": "reporting",
        "SLIDE_DISABLED_TOOLS": "slide_presentation"
      }
    }
  }
}
```

Or pass everything as CLI args:

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

### Test Your Installation

Use the official MCP Inspector to drive the server interactively:

```bash
export SLIDE_API_KEY=tk_...
npx @modelcontextprotocol/inspector ./slide-mcp-server
```

Or run the bundled smoke test (covers initialize, tools/list, resources/list, the
v1.27.0 validation paths, and a sampling of live tool calls):

```bash
SLIDE_API_KEY=tk_... ./smoke_test.sh
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
| `--enable-presentation` | Enable the `slide_presentation` tool | `SLIDE_ENABLE_PRESENTATION` | `false` |
| `--enable-reports` | Enable the `slide_reports` tool | `SLIDE_ENABLE_REPORTS` | `false` |
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

# Enable presentation and reports tools (disabled by default)
./slide-mcp-server --api-key sk_test_123 --enable-presentation --enable-reports

# Enable only the presentation tool
./slide-mcp-server --api-key sk_test_123 --enable-presentation

# Show version
./slide-mcp-server --version
# Output: slide-mcp-server version 2.3.2
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
- `slide_user_management` - User and account management
- `slide_alerts` - Alert monitoring
- `slide_devices` - Device management
- `slide_vms` - Virtual machine management
- `slide_presentation` - Data presentation and reporting
- `slide_meta` - Meta tools for reporting and aggregated data views
- `slide_reports` - Pre-calculated backup/snapshot statistics and reports
- `slide_docs` - Access to official Slide documentation
- `list_all_clients_devices_and_agents` - Hierarchical overview (legacy, use slide_meta instead)

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

### 🎯 Enabling Presentation & Reports Tools

The `slide_presentation` and `slide_reports` tools are **disabled by default** and must be explicitly enabled using CLI flags or environment variables. This design prevents accidental exposure of potentially sensitive reporting capabilities.

#### Why These Tools Are Disabled by Default

- **`slide_presentation`**: Provides advanced formatting and templating capabilities that could potentially be misused for data extraction or system information gathering
- **`slide_reports`**: Generates comprehensive system reports that may contain sensitive operational data

#### Enabling These Tools

```bash
# Enable both tools via CLI flags
./slide-mcp-server --api-key YOUR_KEY --enable-presentation --enable-reports

# Enable only presentation tool
./slide-mcp-server --api-key YOUR_KEY --enable-presentation  

# Enable only reports tool
./slide-mcp-server --api-key YOUR_KEY --enable-reports

# Enable via environment variables
export SLIDE_ENABLE_PRESENTATION=true
export SLIDE_ENABLE_REPORTS=true
./slide-mcp-server --api-key YOUR_KEY

# CLI flags take precedence over environment variables
export SLIDE_ENABLE_PRESENTATION=false
./slide-mcp-server --api-key YOUR_KEY --enable-presentation  # Presentation tool will be enabled
```

#### Combined with Other Options

```bash
# Enable with specific tools mode
./slide-mcp-server --api-key YOUR_KEY --tools reporting --enable-presentation --enable-reports

# Enable while disabling other tools
./slide-mcp-server --api-key YOUR_KEY --enable-reports --disabled-tools "slide_agents,slide_backups"
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

### Generate RDP Bookmark for VM
```json
{
  "name": "slide_vms",
  "arguments": {
    "operation": "get_rdp_bookmark",
    "virt_id": "vm-123"
  }
}
```

### Search Documentation
```json
{
  "name": "slide_docs",
  "arguments": {
    "operation": "search_docs",
    "query": "backup retention policies"
  }
}
```

## 🆕 What's New in v2.4.0

### 🎯 LLM Optimization
- **Streamlined Tool Descriptions**: All tool descriptions optimized for modern LLMs with concise, action-oriented language
- **Consistent Format**: Standardized description format across all 13 meta-tools for better LLM comprehension
- **Reduced Verbosity**: Tool descriptions reduced by 60-80% while maintaining clarity
- **Faster Tool Selection**: LLMs can now select the right tool more quickly with clearer, more focused descriptions

### 🔧 MCP Protocol Enhancements
- **Enhanced Capabilities**: Added MCP protocol capabilities metadata for better client integration
- **Richer Initial Context**: Startup context now includes tools mode, API base URL, and cache metadata
- **Better Performance**: Initial context loading provides immediate infrastructure overview at startup

### 🚀 Unified Release Automation
- **One-Command Release**: New `release-all-in-one.sh` script handles complete release workflow
- **Automated Version Management**: Auto-increment patch version or specify custom version
- **Integrated Workflow**: Commits, builds, signs, packages, and publishes to GitHub in one step
- **Dry-Run Mode**: Test release process without making changes
- **Flexible Options**: Skip tests, local-only builds, and more

### 📚 Previous Features (v2.3.x)
- Initial Context Loading for faster startup
- RDP Bookmark Generation for VM access
- Built-in Documentation Access via `slide_docs` tool
- Cross-Platform GUI Installer

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

### Generate Backup Reports
```json
{
  "name": "slide_reports",
  "arguments": {
    "operation": "daily_backup_snapshot",
    "agent_id": "agent-123",
    "format": "markdown"
  }
}
```

### Search Documentation
```json
{
  "name": "slide_docs",
  "arguments": {
    "operation": "search_docs",
    "query": "backup retention policies"
  }
}
```

## Documentation System

The MCP server includes a comprehensive documentation access system through the `slide_docs` tool. The documentation system has been enhanced with contextual descriptions to help LLMs make better choices when navigating between similar-sounding sections.

### Enhanced Context Features

1. **Section Descriptions**: Each documentation section now includes a detailed description explaining its purpose
   - Example: "Slide Console > Networks" is clarified as "managing virtual networks on Slide devices/cloud"
   - Example: "Product > Networking" is clarified as "network infrastructure requirements and prerequisites"

2. **Topic Descriptions**: Ambiguous topic names include contextual descriptions
   - Topics like "Networks (Managing Networks)" vs "Networking (Requirements)" are clearly differentiated

3. **Context-Aware Search**: Search results include section and topic descriptions to help identify the correct documentation

4. **Improved Navigation**: The LLM can now better distinguish between:
   - Configuration vs Requirements documentation
   - Console UI features vs System prerequisites
   - User management vs Client organization management

### Testing Documentation Context

Run the test script to verify the context improvements:

```bash
./test_scripts/test_docs_context.sh
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

## Release Process

### Unified Release Script (Recommended)

The `release-all-in-one.sh` script automates the entire release workflow:

```bash
# Auto-increment patch version and create release
./scripts/release-all-in-one.sh

# Release specific version
./scripts/release-all-in-one.sh v2.5.0

# Test release process without making changes
./scripts/release-all-in-one.sh --dry-run v2.5.0

# Build locally without pushing to GitHub
./scripts/release-all-in-one.sh --no-push

# Re-release existing version (overwrite tag and GitHub release)
./scripts/release-all-in-one.sh --force v2.5.0
```

**What it does:**
1. Checks prerequisites (Go, GitHub CLI, jq)
2. Validates git status (clean working directory, on main branch)
3. Updates version numbers in Makefile and config.go
4. Runs tests (optional with `--skip-tests`)
5. Commits version changes
6. Builds binaries for all platforms
7. Signs and notarizes macOS binaries (if credentials provided)
8. Creates release packages (.tar.gz for Unix, .zip for Windows)
9. Generates SHA256 checksums
10. Creates release notes from git history
11. Tags the release
12. Pushes to GitHub
13. Creates GitHub release with all assets

**Environment Variables (REQUIRED for releases):**
- `APPLE_ID`: Apple ID for macOS notarization
- `APP_SPECIFIC_PASSWORD`: App-specific password for notarization

⚠️ **Important**: macOS binaries MUST be signed and notarized to work on user systems. The script enforces this requirement and will fail if credentials are not provided. For local testing only, use `--no-push --skip-macos-signing`.

**Setting Up Credentials (One-Time Setup):**

```bash
# Option 1: Use credentials file (Recommended - saves credentials locally)
# File is already created at scripts/release-env.sh and gitignored for security
source scripts/release-env.sh

# Option 2: Set manually each time
export APPLE_ID='your-apple-id@example.com'
export APP_SPECIFIC_PASSWORD='xxxx-xxxx-xxxx-xxxx'
```

**Getting Apple Credentials:**
1. **Apple ID**: Your Apple Developer account email
2. **App-Specific Password**:
   - Go to [appleid.apple.com](https://appleid.apple.com)
   - Sign in and navigate to Security section
   - Click "Generate Password" under App-Specific Passwords
   - Copy the generated password (format: xxxx-xxxx-xxxx-xxxx)
3. **Developer ID Certificate**: Install from Apple Developer account
   - Required for code signing
   - Must be "Developer ID Application" certificate

**Security Note**: The `scripts/release-env.sh` file is gitignored and will never be committed to the repository.

**Options:**
- `--dry-run`: Simulate without making changes
- `--skip-tests`: Skip running tests
- `--no-push`: Build locally without GitHub operations
- `--skip-macos-signing`: Skip signing (TESTING ONLY, must use with --no-push)
- `--force` or `-f`: Force overwrite existing release (deletes tag and GitHub release)
- `--help`: Show detailed help

### Manual Release (Legacy)

For manual control, use the existing `automated-release.sh` script (see `scripts/RELEASE_README.md`).

## Architecture Benefits

### For LLMs
- **Reduced Complexity**: 14 meta-tools vs 52+ individual tools
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

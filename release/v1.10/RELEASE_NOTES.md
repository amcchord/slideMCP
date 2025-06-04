# Slide MCP Server v1.10 - Go Implementation

## 🎉 Complete Feature Parity Release

This release brings the Go implementation of Slide MCP Server to **complete feature parity** with the TypeScript version, implementing the full Slide API v1.10 specification.

## 📊 What's New

### ✅ **Complete API Coverage (29 Tools)**
- **Expanded from 2 tools to 29 tools** - Full Slide API coverage
- **Perfect API Compliance** - Matches Slide API v1.10 specification exactly
- **Cross-Platform Binaries** - Native binaries for all major operating systems

### 🛠️ **New Tool Categories**

**Agent Management (4 tools):**
- `slide_get_agent` - Get detailed agent information
- `slide_create_agent` - Create agent for auto-pair installation  
- `slide_pair_agent` - Pair agent with device using pair code
- `slide_update_agent` - Update agent properties

**Backup Operations (3 tools):**
- `slide_list_backups` - List backup jobs with filtering
- `slide_get_backup` - Get detailed backup information
- `slide_start_backup` - Start backup process for agent

**Snapshot Management (2 tools):**
- `slide_list_snapshots` - List snapshots with location filtering
- `slide_get_snapshot` - Get detailed snapshot information

**File Restore Operations (5 tools):**
- `slide_list_file_restores` - List file restore jobs
- `slide_get_file_restore` - Get file restore details
- `slide_create_file_restore` - Create file restore from snapshot
- `slide_delete_file_restore` - Delete file restore
- `slide_browse_file_restore` - Browse and download files

**Image Export Operations (5 tools):**
- `slide_list_image_exports` - List image export jobs
- `slide_get_image_export` - Get image export details
- `slide_create_image_export` - Create VHDX/VHD/RAW exports
- `slide_delete_image_export` - Delete image export
- `slide_browse_image_export` - Browse exportable disk images

**Virtual Machine Operations (5 tools):**
- `slide_list_virtual_machines` - List disaster recovery VMs
- `slide_get_virtual_machine` - Get VM details with VNC access
- `slide_create_virtual_machine` - Create VM from snapshot
- `slide_update_virtual_machine` - Start/stop/pause VMs
- `slide_delete_virtual_machine` - Delete virtual machine

**System Administration (8 tools):**
- `slide_list_users` / `slide_get_user` - User management
- `slide_list_alerts` / `slide_get_alert` / `slide_update_alert` - Alert management
- `slide_list_accounts` / `slide_get_account` / `slide_update_account` - Account management

**Client Management (5 tools):**
- `slide_list_clients` / `slide_get_client` - List and view clients
- `slide_create_client` / `slide_update_client` / `slide_delete_client` - Full CRUD operations

## 🚀 **Key Features**

### **Enhanced Functionality**
- **VNC Integration**: Automatic VNC viewer URL generation for virtual machines
- **Rich Metadata**: Enhanced responses with LLM-friendly guidance and context
- **Complete Pagination**: Full offset/limit support with proper next_offset handling
- **Advanced Filtering**: Comprehensive query parameter support for all list operations
- **Error Handling**: Detailed error messages with proper HTTP status codes

### **Developer Experience**
- **Type Safety**: Strong typing with comprehensive Go structs
- **API Compliance**: Perfect adherence to Slide API v1.10 specification
- **Documentation**: Detailed tool descriptions and parameter definitions
- **Validation**: Proper input validation and type checking

## 🔧 **Bug Fixes**
- ✅ Fixed API endpoints: `/v1/devices` → `/v1/device`, `/v1/agents` → `/v1/agent`
- ✅ Added missing `sort_by` parameter for device listing with proper defaults
- ✅ Added missing `encryption_algorithm` field to Agent struct
- ✅ Corrected parameter handling and validation across all endpoints

## 📦 **Available Binaries**

This release includes native binaries for all major operating systems:

### **Windows**
- `slide-mcp-server-v1.10-windows-x64.exe` - Windows x64 (Intel/AMD)
- `slide-mcp-server-v1.10-windows-arm64.exe` - Windows ARM64

### **macOS** 
- `slide-mcp-server-v1.10-macos-x64` - macOS Intel (x64)
- `slide-mcp-server-v1.10-macos-arm64` - macOS Apple Silicon (M1/M2/M3)

### **Linux**
- `slide-mcp-server-v1.10-linux-x64` - Linux x64 (Intel/AMD)
- `slide-mcp-server-v1.10-linux-arm64` - Linux ARM64

## 🎯 **Use Cases**

This release enables comprehensive backup and disaster recovery management through MCP:

- **Backup Management**: Schedule, monitor, and manage backup jobs
- **Disaster Recovery**: Create and manage virtual machines from snapshots
- **File Recovery**: Browse and restore individual files from backups
- **Image Export**: Export disk images for migration or archival
- **System Monitoring**: Manage alerts, users, and system health
- **MSP Operations**: Full client and account management capabilities

## 📋 **Requirements**

- **Environment Variable**: `SLIDE_API_KEY` must be set with your Slide API token
- **Network Access**: Connection to `https://api.slide.tech`
- **MCP Client**: Compatible MCP client (e.g., Claude Desktop, Continue, etc.)

## 🔗 **API Documentation**

For complete API documentation, visit: [https://docs.slide.tech/api](https://docs.slide.tech/api)

## 📞 **Support**

- **Documentation**: [https://docs.slide.tech](https://docs.slide.tech)
- **API Reference**: [https://docs.slide.tech/api](https://docs.slide.tech/api)
- **Support**: [hey@slide.tech](mailto:hey@slide.tech)

---

**Full Changelog**: Complete rewrite and expansion from 2 tools to 29 tools with full Slide API v1.10 compatibility. 
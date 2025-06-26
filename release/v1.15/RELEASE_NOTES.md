# Slide MCP Server v1.15 - Meta-Tools Revolution 🚀

## 🎯 Major Release: Tool Architecture Transformation

This is a **major release** that revolutionizes the MCP tool architecture by consolidating **52+ individual tools** into **10 intuitive meta-tools**. This change dramatically simplifies LLM interactions while preserving all existing functionality.

## 📊 What's New in v1.15

### ✨ **Meta-Tools Architecture**
- **Unified Interface**: 10 logical meta-tools replace 52+ individual tools
- **79% Reduction**: From 52+ tools down to 10 meta-tools + 1 special tool
- **Better Organization**: Tools grouped by functional area (agents, networks, backups, etc.)
- **Consistent Pattern**: All meta-tools follow the same `operation` parameter pattern

### 🛠️ **The 10 Meta-Tools**

1. **`slide_agents`** - All agent operations (list, get, create, pair, update)
2. **`slide_backups`** - All backup operations (list, get, start)
3. **`slide_snapshots`** - All snapshot operations (list, get)
4. **`slide_restores`** - File restore and image export operations (10 operations)
5. **`slide_networks`** - Network operations including IPSec, port forwards, WireGuard (14 operations)
6. **`slide_users`** - User operations (list, get)
7. **`slide_alerts`** - Alert operations (list, get, update)
8. **`slide_accounts`** - Account and client operations (8 operations)
9. **`slide_devices`** - Physical device operations (list, get, update, poweroff, reboot)
10. **`slide_vms`** - Virtual machine operations (list, get, create, update, delete)

### 🔧 **Enhanced Security & Distribution**
- **Signed macOS Binaries**: Both ARM64 and x64 macOS binaries are code-signed
- **Developer ID**: Signed with "Developer ID Application: Austin McChord (7PTN7E8EDS)"
- **Runtime Hardening**: Binaries built with runtime protection enabled
- **Reduced Warnings**: Fewer macOS Gatekeeper security warnings

## 🚀 **Key Benefits**

### **For LLMs**
- **Simpler Decision Making**: 10 tools instead of 52+ to choose from
- **Logical Grouping**: Related operations grouped together
- **Consistent Interface**: All meta-tools follow the same pattern
- **Reduced Complexity**: Fewer tools to understand and manage

### **For Developers**
- **Better Maintainability**: Each meta-tool has its own file
- **Consistent Architecture**: Unified handler pattern across all tools
- **Backward Compatibility**: All original functionality preserved
- **Enhanced Documentation**: Clear operation mapping and examples

## 🔄 **Migration Guide**

The meta-tools use an `operation` parameter to specify which action to perform:

### Before (v1.14 and earlier):
```json
{"name": "slide_list_agents", "arguments": {"limit": 10}}
```

### After (v1.15):
```json
{"name": "slide_agents", "arguments": {"operation": "list", "limit": 10}}
```

### More Examples:

```json
// Create Network
{
  "name": "slide_networks", 
  "arguments": {
    "operation": "create",
    "name": "Test Network",
    "type": "standard",
    "router_prefix": "192.168.1.1/24",
    "dhcp": true
  }
}

// Start Backup
{
  "name": "slide_backups",
  "arguments": {
    "operation": "start",
    "agent_id": "agent-456"
  }
}

// Reboot Device
{
  "name": "slide_devices",
  "arguments": {
    "operation": "reboot",
    "device_id": "device-789"
  }
}
```

## 📦 **Available Binaries**

All binaries rebuilt with latest toolchain and enhanced security:

### **Windows**
- `slide-mcp-server-v1.15-windows-x64.exe` - Windows x64 (Intel/AMD)
- `slide-mcp-server-v1.15-windows-arm64.exe` - Windows ARM64

### **macOS** 🔐
- `slide-mcp-server-v1.15-macos-x64` - macOS Intel (x64) **[SIGNED]**
- `slide-mcp-server-v1.15-macos-arm64` - macOS Apple Silicon (M1/M2/M3/M4) **[SIGNED]**

### **Linux**
- `slide-mcp-server-v1.15-linux-x64` - Linux x64 (Intel/AMD)
- `slide-mcp-server-v1.15-linux-arm64` - Linux ARM64

## 🔧 **Technical Implementation**

### Meta-Tool Pattern
Each meta-tool follows a consistent architecture:
- **Handler Function**: Routes operations to appropriate API functions
- **Tool Info Function**: Returns JSON schema with conditional requirements
- **Operation Parameter**: Specifies which action to perform
- **Conditional Schemas**: Dynamic validation based on operation type

### File Structure
```
slideMCP/
├── tools_agents.go      # Agent meta-tool (5 operations)
├── tools_backups.go     # Backup meta-tool (3 operations)
├── tools_snapshots.go   # Snapshot meta-tool (2 operations)
├── tools_restores.go    # Restore meta-tool (10 operations)
├── tools_networks.go    # Network meta-tool (14 operations)
├── tools_users.go       # User meta-tool (2 operations)
├── tools_alerts.go      # Alert meta-tool (3 operations)
├── tools_accounts.go    # Account meta-tool (8 operations)
├── tools_devices.go     # Device meta-tool (5 operations)
├── tools_vms.go         # VM meta-tool (5 operations)
└── api.go               # API functions (unchanged)
```

## 🛡️ **Backwards Compatibility**

- ✅ **Full Compatibility**: All existing functionality preserved
- ✅ **API Unchanged**: No changes to underlying API endpoints
- ✅ **Same Parameters**: All parameters work exactly as before
- ✅ **Same Responses**: API responses unchanged

## 📋 **Requirements**

- **Environment Variable**: `SLIDE_API_KEY` must be set with your Slide API token
- **Network Access**: Connection to `https://api.slide.tech`
- **MCP Client**: Compatible MCP client (e.g., Claude Desktop, Continue, etc.)
- **macOS**: macOS 10.15+ for signed binary compatibility

## 🔍 **Verification**

### Verify macOS Code Signing:
```bash
# Verify signature
codesign --verify --verbose=2 slide-mcp-server-v1.15-macos-arm64

# Check signature details
codesign --display --verbose=4 slide-mcp-server-v1.15-macos-arm64
```

### Verify Checksums:
```bash
# Verify download integrity
shasum -a 256 -c checksums.sha256
```

## 🌟 **What's Next**

This meta-tools architecture provides a solid foundation for:
- **Future Enhancements**: Easier to add new operations to existing categories
- **Better UX**: Simplified tool selection for LLMs
- **Maintenance**: Cleaner, more organized codebase
- **Documentation**: Better API documentation and examples

## 🔗 **Resources**

- **API Documentation**: [https://docs.slide.tech/api](https://docs.slide.tech/api)
- **Meta-Tools Guide**: See `changes.md` for detailed operation mapping
- **Support**: [hey@slide.tech](mailto:hey@slide.tech)

---

**Changelog**: Major architecture refactor - consolidated 52+ individual tools into 10 logical meta-tools. Added code signing for macOS binaries. Enhanced build system and documentation. 
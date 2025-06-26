# Slide MCP Server v1.16 Release Notes

## 🔧 Bug Fixes & Improvements

This release addresses critical issues from v1.15 and provides enhanced stability and reliability for the Slide MCP Server.

### Fixed Issues
- **Release Artifact Integrity**: Resolved build configuration inconsistencies that affected v1.15 release integrity
- **Version Management**: Fixed version synchronization across build scripts and configuration files
- **Build Process**: Enhanced build script reliability and error handling

### Code Signing & Security
- **macOS Code Signing**: All macOS binaries are signed with Developer ID for improved security and compatibility
- **Cross-Platform Support**: Binaries built for all major platforms with optimized performance
- **SHA256 Checksums**: Complete checksum verification for all release binaries

## 🚀 Core Features (Maintained)

### Meta-Tools Architecture
The revolutionary meta-tools system introduced in earlier versions continues to provide:
- **10 Consolidated Tools**: Reduced from 52+ individual tools to 10 logical groupings
- **Simplified Integration**: Easier for LLMs and developers to work with fewer, more focused tools
- **Enhanced Organization**: Related operations grouped by functionality (agents, networks, backups, etc.)

### Supported Operations
- **Agent Management**: List, create, pair, and update Slide agents
- **Backup Operations**: Start and monitor backup processes
- **Network Configuration**: Create and manage networks, IPSec connections, port forwards, WireGuard peers
- **Device Control**: Monitor, update, and control physical devices
- **Virtual Machines**: Create, manage, and control virtual machine instances
- **User & Account Management**: Comprehensive user and account administration
- **Alert Handling**: Monitor and manage system alerts
- **Snapshot Management**: Access and manage system snapshots
- **File & Image Restoration**: Comprehensive restore capabilities

## 📋 Technical Specifications

### Supported Platforms
- **Linux**: x64 and ARM64 architectures
- **macOS**: Intel (x64) and Apple Silicon (ARM64) - **Code Signed**
- **Windows**: x64 and ARM64 architectures

### Security Features
- macOS binaries are code-signed with Developer ID
- All binaries include SHA256 checksums for integrity verification
- Runtime hardening enabled for macOS builds

### Binary Sizes
- Linux x64: ~6.2MB
- Linux ARM64: ~5.9MB  
- macOS x64: ~6.4MB (signed)
- macOS ARM64: ~6.0MB (signed)
- Windows x64: ~6.4MB
- Windows ARM64: ~5.9MB

## 🔄 Migration & Compatibility

### From v1.15
- No breaking changes - direct replacement supported
- Enhanced reliability and stability
- Maintained API compatibility

### From Earlier Versions
- Full backward compatibility with meta-tools architecture
- All original functionality preserved through delegation pattern
- Consistent JSON schema and parameter structure

## 📦 Installation

### Download & Verification
1. Download the appropriate binary for your platform
2. Verify integrity using provided SHA256 checksums:
   ```bash
   shasum -a 256 -c checksums.sha256
   ```

### Platform-Specific Notes
- **macOS**: Signed binaries should run without additional security prompts
- **Linux**: May require execute permissions: `chmod +x slide-mcp-server-*`
- **Windows**: Run from command prompt or PowerShell

## 🔗 Integration Examples

### Basic Agent Listing
```json
{
  "name": "slide_agents",
  "arguments": {
    "operation": "list",
    "limit": 10
  }
}
```

### Network Creation
```json
{
  "name": "slide_networks",
  "arguments": {
    "operation": "create",
    "name": "Production Network",
    "type": "standard",
    "router_prefix": "192.168.1.1/24",
    "dhcp": true
  }
}
```

### Device Management
```json
{
  "name": "slide_devices", 
  "arguments": {
    "operation": "reboot",
    "device_id": "device-123"
  }
}
```

## 📚 Documentation

- **API Reference**: Complete meta-tools documentation in codebase
- **Integration Guide**: Examples for all supported operations
- **Migration Guide**: Detailed transition instructions from individual tools

## 🛡️ Security Notice

macOS users should verify code signature after download:
```bash
codesign --verify --verbose=2 slide-mcp-server-v1.16-macos-*
```

---

**Version**: v1.16  
**Release Date**: June 2024  
**Compatibility**: All previous versions  
**Supported Platforms**: Linux, macOS, Windows (x64/ARM64)

For technical support and documentation, visit the project repository. 
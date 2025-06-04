# Slide MCP Server v1.11 - Enhanced VNC Integration

## 🌐 Enhanced Virtual Machine Access

This release significantly improves the virtual machine experience by adding **browser-accessible VNC viewer URLs** to all virtual machine operations, making it easier than ever to access and manage disaster recovery VMs.

## 📊 What's New in v1.11

### ✨ **Enhanced VNC Integration**
- **Browser-Based VNC Access**: Automatic generation of browser-accessible VNC viewer URLs
- **Zero Client Installation**: Access VMs through any modern web browser
- **Secure URL Generation**: Base64-encoded passwords with proper URL encoding
- **Instant VM Access**: Direct links to VM consoles in all VM tool responses

### 🛠️ **Updated Virtual Machine Tools**

**Enhanced VM Operations:**
- `slide_list_virtual_machines` - Now includes `_vnc_viewer_url` for each VM when available
- `slide_get_virtual_machine` - Enhanced response with VNC viewer URL and metadata  
- `slide_create_virtual_machine` - Immediate VNC access URL in creation response
- `slide_update_virtual_machine` - VNC URL preserved through state changes

**New Response Fields:**
- `_vnc_viewer_url` - Direct browser access to VM console
- Enhanced `_metadata` with VNC guidance and best practices
- Consistent response formatting across all VM operations

## 🚀 **Key Improvements**

### **Better User Experience**
- **One-Click VM Access**: Click the VNC viewer URL to instantly access VM console
- **No VNC Client Required**: Works with any modern web browser
- **Consistent URLs**: All VM tools now provide the same enhanced VNC access
- **Enhanced Guidance**: Better metadata and usage instructions for LLMs

### **Technical Enhancements**
- **URL Format**: `https://slide.recipes/mcpTools/vncViewer.php?id={vm_id}&ws={websocket_uri}&password={base64_password}&encoding=base64`
- **Security**: Proper URL encoding and base64 password encoding
- **Reliability**: Consistent VNC URL generation across all VM operations
- **Performance**: Efficient URL generation without additional API calls

## 🔧 **Backwards Compatibility**

- ✅ **Full Compatibility**: All existing functionality preserved
- ✅ **Response Format**: Enhanced responses are backwards compatible
- ✅ **API Endpoints**: No changes to existing API endpoints
- ✅ **Tool Names**: All tool names remain unchanged

## 🎯 **Use Cases Enhanced**

This release significantly improves these workflows:

- **Disaster Recovery Testing**: Instantly access VMs to verify backup integrity
- **Emergency Access**: Quick VM console access during outages or incidents
- **System Administration**: Easy VM management without VNC client setup
- **Remote Support**: Share VNC URLs with technicians for collaborative troubleshooting
- **Training & Demos**: Easy VM access for training sessions and demonstrations

## 📦 **Available Binaries**

Updated native binaries for all major operating systems:

### **Windows**
- `slide-mcp-server-v1.11-windows-x64.exe` - Windows x64 (Intel/AMD)
- `slide-mcp-server-v1.11-windows-arm64.exe` - Windows ARM64

### **macOS** 
- `slide-mcp-server-v1.11-macos-x64` - macOS Intel (x64)
- `slide-mcp-server-v1.11-macos-arm64` - macOS Apple Silicon (M1/M2/M3)

### **Linux**
- `slide-mcp-server-v1.11-linux-x64` - Linux x64 (Intel/AMD)
- `slide-mcp-server-v1.11-linux-arm64` - Linux ARM64

## 🔄 **Migration Guide**

### From v1.10 to v1.11
1. **Download** the new v1.11 binary for your platform
2. **Replace** your existing binary with the v1.11 version
3. **Restart** your MCP server
4. **Enjoy** enhanced VNC integration - no configuration changes needed!

### New VNC Features
- VM responses now include `_vnc_viewer_url` field when websocket URI is available
- Enhanced metadata provides VNC guidance and usage instructions
- Same API, better experience - no breaking changes

## 📋 **Requirements**

- **Environment Variable**: `SLIDE_API_KEY` must be set with your Slide API token
- **Network Access**: Connection to `https://api.slide.tech`
- **MCP Client**: Compatible MCP client (e.g., Claude Desktop, Continue, etc.)
- **VNC Access**: Modern web browser for VNC viewer URLs

## 🌟 **Example Usage**

When you create or access a virtual machine, you'll now get responses like:

```json
{
  "virt_id": "vm_abc123",
  "state": "running",
  "_vnc_viewer_url": "https://slide.recipes/mcpTools/vncViewer.php?id=vm_abc123&ws=wss%3A//vm.slide.tech%3A6080&password=dGVzdA%3D%3D&encoding=base64",
  "_metadata": {
    "vnc_guidance": "Use the _vnc_viewer_url to access the virtual machine's console via a browser-based VNC client."
  }
}
```

Simply click the VNC viewer URL to access your VM console instantly!

## 🔗 **Resources**

- **API Documentation**: [https://docs.slide.tech/api](https://docs.slide.tech/api)
- **VNC Viewer**: Browser-based VNC access through slide.recipes
- **Support**: [hey@slide.tech](mailto:hey@slide.tech)

---

**Changelog**: Enhanced virtual machine operations with browser-accessible VNC viewer URLs for improved user experience and instant VM console access. 
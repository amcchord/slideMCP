# Slide MCP Server v1.12 - Code Signing & Maintenance Release

## 🔐 Enhanced Security & Distribution

This release focuses on **improved security** and **distribution reliability** by introducing **code signing for macOS binaries** and general maintenance improvements.

## 📊 What's New in v1.12

### ✨ **Code Signing for macOS**
- **Signed macOS Binaries**: All macOS binaries are now properly code-signed with Developer ID
- **Improved Security**: Reduced Gatekeeper warnings and enhanced system trust
- **Apple Silicon & Intel**: Both ARM64 and x64 macOS binaries are signed
- **Runtime Hardening**: Binaries built with runtime protection enabled

### 🛠️ **Build System Improvements**

**Enhanced Build Process:**
- **Automated Signing**: Makefile targets for signed binary builds
- **Cross-Platform Support**: Consistent builds across all supported platforms
- **Release Automation**: Streamlined release packaging and checksums
- **Developer Experience**: Improved build scripts with better error handling

### 🔧 **Security Enhancements**
- **macOS Compatibility**: Signed binaries work seamlessly with macOS security features
- **Timestamp Signing**: All signatures include secure timestamps
- **Verification Process**: Built-in signature verification in build process
- **Distribution Ready**: Binaries ready for secure distribution

## 🚀 **Key Improvements**

### **Better Security**
- **Code Signing**: macOS binaries signed with Developer ID Application certificate
- **Runtime Protection**: Enhanced runtime security options enabled
- **System Trust**: Reduced security warnings on macOS systems
- **Secure Distribution**: Signed binaries for improved trust and security

### **Build System**
- **Makefile Targets**: New targets for signed builds (`build-all-signed`)
- **Version Management**: Automated version handling in build process
- **Checksum Generation**: Automatic SHA256 checksum generation for releases
- **Cross-Platform**: Consistent builds for all supported architectures

## 🔧 **Backwards Compatibility**

- ✅ **Full Compatibility**: All existing functionality preserved
- ✅ **API Unchanged**: No changes to API endpoints or responses
- ✅ **Tool Names**: All tool names remain unchanged
- ✅ **Configuration**: Same environment variables and setup process

## 🎯 **Security Benefits**

This release improves security in these areas:

- **macOS Integration**: Signed binaries work better with macOS security features
- **Enterprise Deployment**: Easier deployment in enterprise environments
- **Developer Trust**: Code signing provides developer identity verification
- **System Compatibility**: Better compatibility with macOS security policies

## 📦 **Available Binaries**

All binaries rebuilt with latest toolchain and security improvements:

### **Windows**
- `slide-mcp-server-v1.12-windows-x64.exe` - Windows x64 (Intel/AMD)
- `slide-mcp-server-v1.12-windows-arm64.exe` - Windows ARM64

### **macOS** 🔐
- `slide-mcp-server-v1.12-macos-x64` - macOS Intel (x64) **[SIGNED]**
- `slide-mcp-server-v1.12-macos-arm64` - macOS Apple Silicon (M1/M2/M3/M4) **[SIGNED]**

### **Linux**
- `slide-mcp-server-v1.12-linux-x64` - Linux x64 (Intel/AMD)
- `slide-mcp-server-v1.12-linux-arm64` - Linux ARM64

## 🔄 **Migration Guide**

### From v1.11 to v1.12
1. **Download** the new v1.12 binary for your platform
2. **Replace** your existing binary with the v1.12 version
3. **Restart** your MCP server
4. **Enjoy** improved security and signed macOS binaries!

### macOS Users
- **Reduced Warnings**: Signed binaries should produce fewer security warnings
- **Better Integration**: Improved compatibility with macOS security features
- **Same Functionality**: All features work exactly the same as before

## 📋 **Requirements**

- **Environment Variable**: `SLIDE_API_KEY` must be set with your Slide API token
- **Network Access**: Connection to `https://api.slide.tech`
- **MCP Client**: Compatible MCP client (e.g., Claude Desktop, Continue, etc.)
- **macOS**: macOS 10.15+ for signed binary compatibility

## 🔍 **Verification**

You can verify the macOS binaries are properly signed:

```bash
# Verify signature
codesign --verify --verbose=2 slide-mcp-server-v1.12-macos-arm64

# Check signature details
codesign --display --verbose=4 slide-mcp-server-v1.12-macos-arm64
```

## 🌟 **Technical Details**

### Code Signing Information
- **Certificate**: Developer ID Application: Austin McChord (7PTN7E8EDS)
- **Runtime**: Hardened runtime enabled
- **Timestamp**: Secure timestamping included
- **Verification**: Built-in signature verification

### Build Information
- **Go Version**: Latest stable Go toolchain
- **Build Flags**: Optimized builds with `-ldflags "-s -w"`
- **Checksums**: SHA256 checksums provided for all binaries

## 🔗 **Resources**

- **API Documentation**: [https://docs.slide.tech/api](https://docs.slide.tech/api)
- **Code Signing Info**: See `CODE_SIGNING.md` for developer details
- **Support**: [hey@slide.tech](mailto:hey@slide.tech)

---

**Changelog**: Added code signing for macOS binaries, improved build system, and enhanced security for distribution. 
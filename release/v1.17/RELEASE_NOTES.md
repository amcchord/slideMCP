# Slide MCP Server v1.17 Release Notes

**Release Date:** June 26, 2025  
**Application Version:** 1.2.2

## 🐛 Bug Fixes

### Fixed Infinite Recursion in enrichWithClientName Function

- **Issue:** The `enrichWithClientName` function would enter infinite recursion when processing primitive data types (strings, integers, booleans, etc.)
- **Root Cause:** Primitive types would continuously hit the `default` case in the function, get converted through JSON marshal/unmarshal cycles without changing, and then recursively call the function again
- **Fix:** 
  - Added explicit handling for primitive types (`string`, `int`, `int64`, `float64`, `bool`, `nil`) that returns them as-is without processing
  - Added recursion prevention check in the default case to detect when JSON conversion doesn't change the data
  - This ensures the function only processes complex data structures that actually need client name enrichment

### Code Signing & Notarization Improvements

- **Enhanced:** macOS binaries are now properly code-signed with hardened runtime enabled
- **Enhanced:** macOS binaries have been submitted to Apple for notarization to improve security and user experience
- **Enhanced:** All macOS binaries now include secure timestamps as required for notarization

## 📦 What's Included

### Supported Platforms

- **Linux x64** (`slide-mcp-server-v1.17-linux-x64`)
- **Linux ARM64** (`slide-mcp-server-v1.17-linux-arm64`)  
- **macOS x64** (`slide-mcp-server-v1.17-macos-x64`) - ✅ **Notarized**
- **macOS ARM64** (`slide-mcp-server-v1.17-macos-arm64`) - ✅ **Notarized**
- **Windows x64** (`slide-mcp-server-v1.17-windows-x64.exe`)

### Archive Formats

- **Linux/macOS:** `.tar.gz` compressed archives
- **Windows:** `.zip` compressed archive
- **Individual binaries:** Available for direct download

## 🔐 Security & Verification

- All files include SHA256 checksums in `checksums.sha256`
- macOS binaries are code-signed with Developer ID and notarized by Apple
- Windows and Linux binaries are built with security best practices

## 📈 Impact

This release fixes a critical bug that could cause the MCP server to crash or become unresponsive when processing certain API responses that contain primitive data types. The infinite recursion issue primarily affected:

- Device listings with mixed data types
- Agent information processing  
- Network configuration responses
- Any API response containing primitive values in nested structures

**Recommended:** All users should upgrade to v1.17 to ensure stable operation.

## 🚀 Installation

Replace your existing `slide-mcp-server` binary with the appropriate version for your platform. No configuration changes are required.

### macOS Users

The notarized macOS binaries will no longer trigger security warnings when first run. The system will recognize them as properly signed and notarized applications.

---

**Full Changelog:** [Compare v1.16...v1.17](https://github.com/your-repo/slide-mcp-server/compare/v1.16...v1.17) 
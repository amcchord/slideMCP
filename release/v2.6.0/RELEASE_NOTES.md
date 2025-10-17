# Slide MCP Server v2.6.0 Release Notes

## 🚀 New Features

### Push Restore Support
This release introduces **Push Restore** functionality, allowing you to restore files directly to protected systems through the MCP interface.

**New Operations:**
- `list_pushes` - List push operations for a file restore
- `create_push` - Create a push operation to restore files to a protected system
- `update_push` - Update push operation state (primarily for cancellation)

**Key Features:**
- Restore files directly to protected systems without manual download/upload
- Real-time state tracking (created → in_progress → completed/failed)
- Ability to cancel in-progress push operations
- Secure destination folder validation (`X:\SlideRestore` only)

**Security:**
For security reasons, files can only be restored to the `SlideRestore` folder at the root of any drive (e.g., `C:\SlideRestore`, `D:\SlideRestore`, etc.). This restriction prevents accidental overwrites of system or user files.

**Example Usage:**
```json
{
  "name": "slide_restores",
  "arguments": {
    "operation": "create_push",
    "file_restore_id": "fr_0123456789ab",
    "source_file_path": "C/Users/Administrator/Documents/ImportantFile.docx",
    "destination_folder": "C:\\SlideRestore"
  }
}
```

## 🔄 Changes

### Snapshot Location Default
The default behavior for snapshot listing has been updated to be more inclusive:

- **Previous behavior**: Snapshots had to be explicitly filtered by location
- **New behavior**: By default, snapshots from all locations (local, cloud, and deleted) are shown using the `location_any` filter

This change provides a more complete view of snapshot availability without requiring explicit filters. You can still use specific location filters (`exists_local`, `exists_cloud`, etc.) when needed.

## 🛠️ Technical Details

### API Changes
- Added `FileRestorePush` data structure with complete state management
- Implemented three new API functions with comprehensive validation
- Enhanced error messages for better debugging and user guidance

### Tool Updates
- Updated `slide_restores` tool with push restore operations
- Added new parameters: `file_restore_push_id`, `source_file_path`, `destination_folder`, `state`
- Updated `slide_snapshots` tool to include `location_any` in the location enum
- Enhanced tool descriptions with push restore guidance

### Permission Updates
- Added push restore operations to the restore management permission set
- Maintains existing security model and access controls

## 📋 Installation

### Download Pre-built Binaries
Choose the appropriate binary for your platform:

- **macOS (Intel)**: `slide-mcp-server-v2.6.0-darwin-amd64.tar.gz`
- **macOS (Apple Silicon)**: `slide-mcp-server-v2.6.0-darwin-arm64.tar.gz`
- **Linux (x64)**: `slide-mcp-server-v2.6.0-linux-amd64.tar.gz`
- **Linux (ARM64)**: `slide-mcp-server-v2.6.0-linux-arm64.tar.gz`
- **Windows (x64)**: `slide-mcp-server-v2.6.0-windows-x64.zip`

### Configuration
Add your Slide API key to your MCP settings:

```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

## 🔐 Code Signing
macOS binaries are signed and notarized with Apple Developer ID for seamless installation without security warnings.

## 📚 Documentation
For complete documentation, see:
- [README.md](../../README.md) - Full server documentation
- [CHANGELOG.md](../../CHANGELOG.md) - Complete change history

## 🐛 Bug Reports
Report issues at: https://github.com/amcchord/slideMCP/issues

---

**Full Changelog**: v2.5.0...v2.6.0


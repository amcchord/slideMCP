# Slide MCP Changes

## 2025-01-17 - v2.6.0 - Push Restore Support

### Added
- **Push Restore Functionality**: New operations to restore files directly to protected systems
  - `list_pushes` - List push operations for a file restore
  - `create_push` - Create a push operation to restore files to protected system
  - `update_push` - Update push operation state (cancel operations)
- Push restore destination folder validation (must be `X:\SlideRestore` where X is drive letter)
- Enhanced metadata and guidance for push restore operations
- State tracking for push operations (created, in_progress, completed, failed, canceled)

### Changed
- **Snapshot Location Default**: Updated default snapshot listing to use `location_any` filter
  - Now shows snapshots from all locations (local, cloud, and deleted) by default
  - Added `location_any` to the `snapshot_location` enum in tool definitions
- Updated `slide_restores` tool description to include push restore functionality
- Updated permissions in config.go to include push restore operations

### Technical Details
- Added `FileRestorePush` struct with fields: file_restore_push_id, file_restore_id, start_time, end_time, state, source_file_path, destination_folder, size
- Implemented API functions: `listFileRestorePushes()`, `createFileRestorePush()`, `updateFileRestorePush()`
- Added strict validation for destination folder format with helpful error messages
- Updated tools_restores.go with push operations and parameter definitions
- Updated tools_snapshots.go to include location_any in enum

## 2025-01-10 - Documentation Access Tool

### Added
- New `slide_docs` tool for accessing Slide documentation
  - `list_sections` - Browse available documentation categories
  - `get_topics` - View topics within a specific section
  - `search_docs` - Search across all documentation
  - `get_content` - Get detailed content for a specific topic
  - `get_api_reference` - Quick access to API reference information
- Added `tools_docs.go` implementing documentation access functionality
- Created `DOCS_TOOL_README.md` explaining the new tool usage

### Purpose
The slide_docs tool allows LLMs to reference official Slide documentation from docs.slide.tech when answering questions about Slide backup and recovery features. This provides contextual help, best practices, troubleshooting guidance, and API reference information directly within the MCP agent.

## 2025-01-09 - Meta Tool Performance Improvements 
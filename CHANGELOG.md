# Slide MCP Changes

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
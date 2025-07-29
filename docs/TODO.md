# MCP Server Refactor - Meta-Tools Implementation

## Overview
Consolidating 52+ individual tools into 8 meta-tools organized by API segments to reduce complexity for LLMs.

## Meta-Tool Organization

### 1. **Agents Meta-Tool** (`tools_agents.go`)
- [ ] `slide_list_agents`
- [ ] `slide_get_agent`
- [ ] `slide_create_agent`
- [ ] `slide_pair_agent`
- [ ] `slide_update_agent`

### 2. **Backups Meta-Tool** (`tools_backups.go`)
- [ ] `slide_list_backups`
- [ ] `slide_get_backup`
- [ ] `slide_start_backup`

### 3. **Snapshots Meta-Tool** (`tools_snapshots.go`)
- [ ] `slide_list_snapshots`
- [ ] `slide_get_snapshot`

### 4. **Restores Meta-Tool** (`tools_restores.go`)
- [ ] `slide_list_file_restores`
- [ ] `slide_get_file_restore`
- [ ] `slide_create_file_restore`
- [ ] `slide_delete_file_restore`
- [ ] `slide_browse_file_restore`
- [ ] `slide_list_image_exports`
- [ ] `slide_get_image_export`
- [ ] `slide_create_image_export`
- [ ] `slide_delete_image_export`
- [ ] `slide_browse_image_export`

### 5. **Networks Meta-Tool** (`tools_networks.go`)
- [ ] `slide_list_networks`
- [ ] `slide_get_network`
- [ ] `slide_create_network`
- [ ] `slide_update_network`
- [ ] `slide_delete_network`
- [ ] `slide_create_network_ipsec_conn`
- [ ] `slide_update_network_ipsec_conn`
- [ ] `slide_delete_network_ipsec_conn`
- [ ] `slide_create_network_port_forward`
- [ ] `slide_update_network_port_forward`
- [ ] `slide_delete_network_port_forward`
- [ ] `slide_create_network_wg_peer`
- [ ] `slide_update_network_wg_peer`
- [ ] `slide_delete_network_wg_peer`

### 6. **Users Meta-Tool** (`tools_users.go`)
- [ ] `slide_list_users`
- [ ] `slide_get_user`

### 7. **Alerts Meta-Tool** (`tools_alerts.go`)
- [ ] `slide_list_alerts`
- [ ] `slide_get_alert`
- [ ] `slide_update_alert`

### 8. **Accounts Meta-Tool** (`tools_accounts.go`)
- [ ] `slide_list_accounts`
- [ ] `slide_get_account`
- [ ] `slide_update_account`
- [ ] `slide_list_clients`
- [ ] `slide_get_client`
- [ ] `slide_create_client`
- [ ] `slide_update_client`
- [ ] `slide_delete_client`

### 9. **Devices Meta-Tool** (`tools_devices.go`)
- [x] `slide_list_devices`
- [x] `slide_get_device`
- [x] `slide_update_device`
- [x] `slide_poweroff_device`
- [x] `slide_reboot_device`

### 10. **VMs Meta-Tool** (`tools_vms.go`)
- [x] `slide_list_virtual_machines`
- [x] `slide_get_virtual_machine`
- [x] `slide_create_virtual_machine`
- [x] `slide_update_virtual_machine`
- [x] `slide_delete_virtual_machine`

### 11. **Special Tool** (keep as-is)
- [x] `list_all_clients_devices_and_agents`

## Implementation Tasks

### Phase 1: Create Meta-Tool Files
- [x] Create `tools_agents.go`
- [x] Create `tools_backups.go`
- [x] Create `tools_snapshots.go`
- [x] Create `tools_restores.go`
- [x] Create `tools_networks.go`  
- [x] Create `tools_users.go`
- [x] Create `tools_alerts.go`
- [x] Create `tools_accounts.go`
- [x] Create `tools_devices.go`
- [x] Create `tools_vms.go`

### Phase 2: Implement Meta-Tool Logic
- [x] Design unified input schema for each meta-tool
- [x] Implement operation routing within each meta-tool
- [x] Migrate API function calls from `api.go`
- [x] Add comprehensive error handling
- [x] Add operation validation

### Phase 3: Update Main Handler
- [x] Update `handleToolCall()` to route to meta-tools
- [x] Update `getAllTools()` to return meta-tool definitions
- [x] Remove individual tool handlers from switch statement

### Phase 4: Testing & Validation
- [x] Test basic compilation
- [ ] Test each meta-tool individually
- [ ] Test end-to-end functionality
- [ ] Validate all original operations still work
- [ ] Test error scenarios

### Phase 5: Documentation & Cleanup
- [x] Update `changes.md` with implementation details
- [x] Add usage examples for each meta-tool
- [x] Clean up unused code in `main.go`
- [x] Separate VMs into their own meta-tool
- [x] Bump version to 1.2.0
- [x] Bump version to 1.2.1
- [ ] Update README if needed

## CLI Enhancement Tasks

### Phase 6: Command Line Interface Improvements
- [x] **Task 1**: Add CLI argument parsing for API key
  - [x] Import `flag` package
  - [x] Add `--api-key` flag support
  - [x] Update API key initialization logic to check both CLI and environment
  - [x] Prioritize CLI flag over environment variable if both provided
  
- [x] **Task 2**: Add version flag support
  - [x] Add `--version` flag
  - [x] Print version information and exit when flag is used
  - [x] Update version constant to 1.2.2 for these CLI improvements

- [x] **Task 3**: Add configurable base URL support
  - [x] Change APIBaseURL from constant to variable in api.go
  - [x] Add `--base-url` CLI flag support
  - [x] Add `SLIDE_BASE_URL` environment variable support
  - [x] Set default to "https://api.slide.tech" for backward compatibility
  - [x] Prioritize CLI flag over environment variable
  - [x] Update version constant to 1.2.3 for this enhancement

- [x] **Task 4**: Add tools filtering/permission system
  - [x] Define tool permission categories (reporting, restores, full-safe, full)
    - **reporting**: Only read operations (list, get) - no modifications allowed
    - **restores**: Reporting + VM/restore/network management - no agent/snapshot deletion
    - **full-safe**: Everything except deleting agents and snapshots (default)
    - **full**: All operations available
  - [x] Add `--tools` CLI flag support  
  - [x] Add `SLIDE_TOOLS` environment variable support
  - [x] Implement tool filtering in getAllTools()
  - [x] Implement execution filtering in handleToolCall()
  - [x] Set default to "full" for backward compatibility
  - [x] Update version constant to 1.2.4 for this enhancement

- [x] **Task 5**: Change default tools mode to full-safe for improved security
  - [x] Update default from ToolsFull to ToolsFullSafe
  - [x] Update version to 2.0.1
  - [x] Update all documentation to reflect new default
  - [x] Add support for list_deleted operation in permission system

### Phase 7: CLI Testing & Documentation
- [x] Test CLI flag functionality
- [x] Test version flag
- [x] Update help text/usage information
- [x] Document new CLI options
- [x] Test base URL configuration functionality
- [x] Test tools filtering functionality

## API v1.15.1 Update Tasks

### Phase 8: Slide API v1.15.1 Integration
- [x] **Agent Passphrase Management** (Tickets 1979, 2255, 2256)
  - [x] Add AgentPassphrase and AgentVSSWriterConfig data structures
  - [x] Update Agent struct with passphrases, sealed, and vss_writer_configs fields  
  - [x] Add add_passphrase and delete_passphrase operations to slide_agents tool
  - [x] Update agent update operation to support new parameters
  
- [x] **QCOW2 Image Type Support** (Ticket 682)
  - [x] Add "qcow2" to image_type enum in slide_restores tool
  
- [x] **Enhanced VM Information** (Ticket 874)
  - [x] Add ip_address, mac_address, and rdp_endpoint fields to VirtualMachine struct
  
- [x] **Testing & Validation**
  - [x] Basic compilation testing completed
  - [x] Version verification completed  
  - [x] Update version to 2.2.0 for API v1.15.1 support
  - [ ] Integration testing with actual API calls (requires API access)
  
### Additional Notes for API v1.15.1
- **Note**: The WireGuard server public key (`wg_public_key`) and total protected data (`total_agent_included_volume_used_bytes`) fields mentioned in the changelog were already present in the existing Network and Device structs respectively, so no changes were needed for those features.

## Phase 9: RDP Bookmark Enhancement

### User Experience Improvement for VM Access
- [x] Add `get_rdp_bookmark` operation to `slide_vms` tool
- [x] Generate standard Windows RDP (.rdp) files for easy VM access
- [x] Include comprehensive RDP configuration with optimal defaults
- [x] Provide clear usage instructions and metadata
- [x] Validate RDP endpoint availability before generating bookmark

**RDP Bookmark Features:**
- **One-click Access**: Generate downloadable .rdp files that users can double-click to connect
- **Standard Format**: Compatible with Windows Remote Desktop, macOS Remote Desktop, and other RDP clients
- **Optimized Settings**: Includes sensible defaults for compression, audio, clipboard, and display settings
- **Security**: Prompts for credentials on connection, supports modern RDP security features
- **User-friendly**: Clear instructions and suggested filename for easy use

## Phase 10: Initial Context Enhancement

### Performance Optimization for MCP Initialization
- [x] Add `fetchInitialContext()` function to load client/device/agent data at startup
- [x] Include initial context data in MCP `initialize` response
- [x] Add comprehensive metadata for context usage guidance
- [x] Implement graceful error handling for context loading failures
- [x] Add timestamp for context freshness tracking

**Initial Context Features:**
- **Immediate Availability**: Client/device/agent hierarchy loaded at startup
- **Performance Boost**: Eliminates the typical first API call delay
- **Fallback Support**: Gracefully handles failures with helpful error messages
- **Fresh Data**: Includes timestamp for cache management
- **Metadata Rich**: Clear guidance on usage and refresh patterns
- **Non-blocking**: Initialization succeeds even if context loading fails

## Version History
- **2.3.0**: Added initial context loading and RDP bookmark generation features for improved performance and user experience
- **2.2.0**: Added support for Slide API v1.15.1 features including agent passphrase management, QCOW2 image type, VSS writer configuration, and enhanced VM information
- **2.0.1**: Added granular tool disabling feature - disable specific tools via --disabled-tools flag or SLIDE_DISABLED_TOOLS environment variable for fine-grained access control
- **1.2.5**: Changed default tools mode from "full" to "full-safe" for improved security
- **1.2.4**: Added tools filtering system for granular permission control (reporting, restores, full-safe, full)
- **1.2.3**: Added configurable base URL support via --base-url flag and SLIDE_BASE_URL environment variable
- **1.2.2**: Added CLI enhancements - API key via command line and --version flag
- **1.2.1**: Added client name enrichment system - automatically includes `client_name` alongside `client_id` in all API responses for improved LLM reporting readability
- **1.2.0**: Consolidated 52+ individual tools into 10 organized meta-tools

## Notes
- Each meta-tool will accept an `operation` parameter to determine which specific API function to call
- Maintain backward compatibility during transition
- Preserve all existing functionality and error handling
- Use consistent patterns across all meta-tools

## CLI Usage

### New Command Line Options (v1.2.2+)

**API Key Configuration:**
```bash
# Using CLI flag (takes precedence)
./slide-mcp-server --api-key YOUR_API_KEY

# Using environment variable (backward compatible)
export SLIDE_API_KEY=YOUR_API_KEY
./slide-mcp-server

# CLI flag overrides environment variable
export SLIDE_API_KEY=env_key
./slide-mcp-server --api-key cli_key  # Uses cli_key
```

**Base URL Configuration (v1.2.3+):**
```bash
# Using CLI flag (takes precedence)
./slide-mcp-server --api-key YOUR_API_KEY --base-url https://custom.api.example.com

# Using environment variable
export SLIDE_BASE_URL=https://custom.api.example.com
./slide-mcp-server --api-key YOUR_API_KEY

# CLI flag overrides environment variable
export SLIDE_BASE_URL=https://env.api.example.com
./slide-mcp-server --api-key YOUR_API_KEY --base-url https://cli.api.example.com  # Uses cli.api.example.com

# Default if neither is provided: https://api.slide.tech
```

**Version Information:**
```bash
# Display version and exit
./slide-mcp-server --version
# Output: slide-mcp-server version 2.0.1
```

**Tools Filtering Configuration (v1.2.4+):**
```bash
# Full-safe access mode (default - everything except deleting agents and snapshots)
./slide-mcp-server --api-key YOUR_API_KEY
./slide-mcp-server --api-key YOUR_API_KEY --tools full-safe

# Reporting mode - only read operations (list, get, browse)
./slide-mcp-server --api-key YOUR_API_KEY --tools reporting

# Restores mode - reporting + VM/restore/network management
./slide-mcp-server --api-key YOUR_API_KEY --tools restores

# Full access mode - complete access including dangerous operations
./slide-mcp-server --api-key YOUR_API_KEY --tools full

# Using environment variable
export SLIDE_TOOLS=reporting
./slide-mcp-server --api-key YOUR_API_KEY

# CLI flag overrides environment variable
export SLIDE_TOOLS=reporting
./slide-mcp-server --api-key YOUR_API_KEY --tools restores  # Uses restores mode
```

**Tools Filtering Permission Levels:**

- **reporting**: Read-only access
  - ✅ All list/get/browse operations
  - ❌ No create/update/delete operations
  - ❌ No power control operations

- **restores**: Data recovery and VM management
  - ✅ All reporting operations
  - ✅ VM operations (create, update, delete)
  - ✅ File restore operations (create, delete, browse)
  - ✅ Image export operations (create, delete, browse)
  - ✅ Network management (create, update, delete)
  - ✅ Account/client management
  - ✅ Device management and power control
  - ✅ Backup management and alerts
  - ❌ Agent deletion
  - ❌ Snapshot deletion

- **full-safe**: Everything except dangerous operations (default)
  - ✅ All operations from restores mode
  - ✅ Agent creation, pairing, and updates
  - ❌ Agent deletion
  - ❌ Snapshot deletion

- **full**: Complete access
  - ✅ All operations available
  - ✅ Agent deletion
  - ✅ Snapshot deletion

**Error Handling:**
- If no API key is provided via CLI or environment, the server will exit with an error message
- The error message explains both methods for providing the API key
- Base URL is optional - defaults to https://api.slide.tech if not specified
- Tools mode is optional - defaults to "full-safe" if not specified
- Invalid tools mode will cause the server to exit with an error listing valid options 
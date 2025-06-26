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
- [ ] Update README if needed

## Notes
- Each meta-tool will accept an `operation` parameter to determine which specific API function to call
- Maintain backward compatibility during transition
- Preserve all existing functionality and error handling
- Use consistent patterns across all meta-tools 
# MCP Server Refactor: Meta-Tools Implementation

## Overview

This refactor consolidates 52+ individual MCP tools into 9 meta-tools organized by API functionality. This change reduces complexity for LLMs by providing fewer, more focused tools that group related operations together.

## What Changed

### Before: Individual Tools
Previously, the MCP server exposed individual tools for each API operation:
- `slide_list_devices`, `slide_get_device`, `slide_update_device`, etc.
- `slide_list_agents`, `slide_get_agent`, `slide_create_agent`, etc.
- Total: 52+ individual tools

### After: Meta-Tools
Now the server exposes 10 meta-tools that group related functionality:
1. **slide_agents** - All agent operations
2. **slide_backups** - All backup operations  
3. **slide_snapshots** - All snapshot operations
4. **slide_restores** - File restore and image export operations
5. **slide_networks** - Network operations (including IPSec, port forwards, WireGuard)
6. **slide_users** - User operations
7. **slide_alerts** - Alert operations
8. **slide_accounts** - Account and client operations
9. **slide_devices** - Physical device operations
10. **slide_vms** - Virtual machine operations

## How Meta-Tools Work

Each meta-tool accepts an `operation` parameter that specifies which underlying API function to call, along with the specific parameters for that operation.

### General Pattern
```json
{
  "name": "slide_agents",
  "arguments": {
    "operation": "list",
    "limit": 10,
    "offset": 0,
    "device_id": "device-123"
  }
}
```

### Operation Mapping

#### slide_agents
- `list` → `slide_list_agents`
- `get` → `slide_get_agent` (requires `agent_id`)
- `create` → `slide_create_agent` (requires `display_name`, `device_id`)
- `pair` → `slide_pair_agent` (requires `pair_code`, `device_id`)
- `update` → `slide_update_agent` (requires `agent_id`, `display_name`)

#### slide_backups
- `list` → `slide_list_backups`
- `get` → `slide_get_backup` (requires `backup_id`)
- `start` → `slide_start_backup` (requires `agent_id`)

#### slide_snapshots
- `list` → `slide_list_snapshots`
- `get` → `slide_get_snapshot` (requires `snapshot_id`)

#### slide_restores
- `list_files` → `slide_list_file_restores`
- `get_file` → `slide_get_file_restore` (requires `file_restore_id`)
- `create_file` → `slide_create_file_restore` (requires `snapshot_id`, `device_id`)
- `delete_file` → `slide_delete_file_restore` (requires `file_restore_id`)
- `browse_file` → `slide_browse_file_restore` (requires `file_restore_id`, `path`)
- `list_images` → `slide_list_image_exports`
- `get_image` → `slide_get_image_export` (requires `image_export_id`)
- `create_image` → `slide_create_image_export` (requires `snapshot_id`, `device_id`, `image_type`)
- `delete_image` → `slide_delete_image_export` (requires `image_export_id`)
- `browse_image` → `slide_browse_image_export` (requires `image_export_id`)

#### slide_networks
- `list` → `slide_list_networks`
- `get` → `slide_get_network` (requires `network_id`)
- `create` → `slide_create_network` (requires `name`, `type`)
- `update` → `slide_update_network` (requires `network_id`)
- `delete` → `slide_delete_network` (requires `network_id`)
- `create_ipsec` → `slide_create_network_ipsec_conn`
- `update_ipsec` → `slide_update_network_ipsec_conn`
- `delete_ipsec` → `slide_delete_network_ipsec_conn`
- `create_port_forward` → `slide_create_network_port_forward`
- `update_port_forward` → `slide_update_network_port_forward`
- `delete_port_forward` → `slide_delete_network_port_forward`
- `create_wg_peer` → `slide_create_network_wg_peer`
- `update_wg_peer` → `slide_update_network_wg_peer`
- `delete_wg_peer` → `slide_delete_network_wg_peer`

#### slide_users
- `list` → `slide_list_users`
- `get` → `slide_get_user` (requires `user_id`)

#### slide_alerts
- `list` → `slide_list_alerts`
- `get` → `slide_get_alert` (requires `alert_id`)
- `update` → `slide_update_alert` (requires `alert_id`, `resolved`)

#### slide_accounts
- `list_accounts` → `slide_list_accounts`
- `get_account` → `slide_get_account` (requires `account_id`)
- `update_account` → `slide_update_account` (requires `account_id`, `alert_emails`)
- `list_clients` → `slide_list_clients`
- `get_client` → `slide_get_client` (requires `client_id`)
- `create_client` → `slide_create_client` (requires `name`)
- `update_client` → `slide_update_client` (requires `client_id`)
- `delete_client` → `slide_delete_client` (requires `client_id`)

#### slide_devices
- `list` → `slide_list_devices`
- `get` → `slide_get_device` (requires `device_id`)
- `update` → `slide_update_device` (requires `device_id`)
- `poweroff` → `slide_poweroff_device` (requires `device_id`)
- `reboot` → `slide_reboot_device` (requires `device_id`)

#### slide_vms
- `list` → `slide_list_virtual_machines`
- `get` → `slide_get_virtual_machine` (requires `virt_id`)
- `create` → `slide_create_virtual_machine` (requires `snapshot_id`, `device_id`)
- `update` → `slide_update_virtual_machine` (requires `virt_id`)
- `delete` → `slide_delete_virtual_machine` (requires `virt_id`)

## Usage Examples

### List Agents
```json
{
  "name": "slide_agents",
  "arguments": {
    "operation": "list",
    "limit": 10,
    "device_id": "device-123"
  }
}
```

### Create Network
```json
{
  "name": "slide_networks", 
  "arguments": {
    "operation": "create",
    "name": "Test Network",
    "type": "standard",
    "router_prefix": "192.168.1.1/24",
    "dhcp": true,
    "dhcp_range_start": "192.168.1.100",
    "dhcp_range_end": "192.168.1.200"
  }
}
```

### Start Backup
```json
{
  "name": "slide_backups",
  "arguments": {
    "operation": "start",
    "agent_id": "agent-456"
  }
}
```

### Reboot Device
```json
{
  "name": "slide_devices",
  "arguments": {
    "operation": "reboot",
    "device_id": "device-789"
  }
}
```

### Create Virtual Machine
```json
{
  "name": "slide_vms",
  "arguments": {
    "operation": "create",
    "snapshot_id": "snapshot-123",
    "device_id": "device-456",
    "cpu_count": 2,
    "memory_in_mb": 4096,
    "network_type": "network-id",
    "network_source": "network-789"
  }
}
```

## Benefits

1. **Reduced Complexity**: LLMs now work with 9 tools instead of 52+
2. **Logical Grouping**: Related operations are grouped together
3. **Better Organization**: Each meta-tool has its own file for maintainability
4. **Consistent Interface**: All meta-tools follow the same pattern
5. **Backward Compatibility**: All original functionality is preserved

## Implementation Details

### Meta-Tool Architecture
Each meta-tool follows a consistent pattern:
- **Handler Function**: `handle{Category}Tool(args)` - Routes operations to appropriate API functions
- **Tool Info Function**: `get{Category}ToolInfo()` - Returns JSON schema for the meta-tool
- **Operation Parameter**: All meta-tools require an `operation` parameter that specifies the action
- **Conditional Schemas**: JSON schemas use `allOf` with conditional requirements based on operation

### Error Handling
- Meta-tools validate the `operation` parameter and return descriptive errors for unknown operations
- All original API error handling is preserved through the delegation pattern
- Invalid operation parameters return clear error messages

### Schema Design
- Each meta-tool accepts ALL possible parameters for its category
- JSON schema uses conditional validation (`if`/`then`) to enforce required parameters per operation
- Parameter descriptions include operation context (e.g., "used with 'list' operation")

### Backward Compatibility
- All original API functions in `api.go` remain unchanged
- Original functionality is preserved through function delegation
- Tool naming follows the pattern `slide_{category}` for consistency

## File Structure

```
slideMCP/
├── main.go              # Updated to use meta-tools
├── api.go               # API functions (unchanged)
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
├── TODO.md              # Refactor tracking
└── changes.md           # This file
```

### Reduction in Tool Count
- **Before**: 52+ individual tools
- **After**: 10 meta-tools + 1 special tool = 11 tools total
- **Reduction**: ~79% fewer tools for LLMs to manage

## Migration Guide

For existing integrations, the meta-tools provide the same functionality with a slightly different call pattern. Replace individual tool calls with the corresponding meta-tool operation.

Example migrations:
```json
// Before
{"name": "slide_list_agents", "arguments": {"limit": 10}}
// After  
{"name": "slide_agents", "arguments": {"operation": "list", "limit": 10}}

// Before
{"name": "slide_create_virtual_machine", "arguments": {"snapshot_id": "snap-123", "device_id": "dev-456"}}
// After
{"name": "slide_vms", "arguments": {"operation": "create", "snapshot_id": "snap-123", "device_id": "dev-456"}}

// Before  
{"name": "slide_reboot_device", "arguments": {"device_id": "dev-789"}}
// After
{"name": "slide_devices", "arguments": {"operation": "reboot", "device_id": "dev-789"}}
``` 
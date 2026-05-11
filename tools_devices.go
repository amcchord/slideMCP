package main

// handleDevicesTool handles all device-related operations through a single meta-tool
func handleDevicesTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_devices", ToolOperations{
		"list":     listDevices,
		"get":      getDevice,
		"update":   updateDevice,
		"poweroff": powerOffDevice,
		"reboot":   rebootDevice,

		// Slide API v1.27.0 additions
		"get_network":    handleDeviceGetNetwork,
		"update_network": handleDeviceUpdateNetwork,
		"list_vlans":     handleDeviceListVLANs,
		"get_vlan":       handleDeviceGetVLAN,
		"create_vlan":    handleDeviceCreateVLAN,
		"update_vlan":    handleDeviceUpdateVLAN,
		"delete_vlan":    handleDeviceDeleteVLAN,
	}), args)
}

var deviceOperationEnums = []string{
	"list", "get", "update", "poweroff", "reboot",
	"get_network", "update_network",
	"list_vlans", "get_vlan", "create_vlan", "update_vlan", "delete_vlan",
}

// getDevicesToolInfo returns the tool definition for the devices meta-tool
func getDevicesToolInfo() ToolInfo {
	return ToolInfo{
		Name: "slide_devices",
		Description: "Manage physical Slide backup appliances. Operations: list, get, update, poweroff, reboot, " +
			"plus v1.27.0 additions: get_network / update_network (device-level network config) and " +
			"list_vlans / get_vlan / create_vlan / update_vlan / delete_vlan (per-device virtual interfaces). " +
			"Get/list responses now include device_warranty_expiration_date and network_update_pending.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        deviceOperationEnums,
				},
				// list parameters
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Number of results per page (max 50) - used with list operations",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Pagination offset - used with list operations",
				},
				"client_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by client ID - used with list_devices operation",
				},
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with list operations",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with list operations",
					"enum":        []string{"hostname", "created"},
				},
				// device parameters
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the device - required for device operations and VM creation",
				},
				"display_name": map[string]interface{}{
					"type":        "string",
					"description": "Display name for the device - used with 'update' operation",
				},
				"hostname": map[string]interface{}{
					"type":        "string",
					"description": "Hostname for the device - used with 'update' operation",
				},

				// v1.27.0: device network
				"network_mode": map[string]interface{}{
					"type":        "string",
					"description": "Network mode - used with update_network and VLAN operations.",
				},
				"network_address": map[string]interface{}{
					"type":        "string",
					"description": "Static IP/CIDR - used with update_network when network_mode=static.",
				},
				"network_gateway": map[string]interface{}{
					"type":        "string",
					"description": "Default gateway IP - used with update_network when network_mode=static.",
				},
				"dns_server_primary": map[string]interface{}{
					"type":        "string",
					"description": "Primary DNS server - used with update_network.",
				},
				"dns_server_secondary": map[string]interface{}{
					"type":        "string",
					"description": "Secondary DNS server - used with update_network.",
				},

				// v1.27.0: VLAN
				"vlan_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the VLAN - required for get_vlan / update_vlan / delete_vlan operations.",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Friendly name for the VLAN - required for create_vlan, optional for update_vlan.",
				},
				"vlan_tag": map[string]interface{}{
					"type":        "number",
					"description": "VLAN tag (802.1Q) - required for create_vlan, optional for update_vlan.",
				},
				"ip_address": map[string]interface{}{
					"type":        "string",
					"description": "VLAN IP/CIDR - used with create_vlan/update_vlan when network_mode=static.",
				},
				"gateway": map[string]interface{}{
					"type":        "string",
					"description": "VLAN gateway IP - used with create_vlan/update_vlan when network_mode=static.",
				},
			},
			"required": []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get"), "then": req("device_id")},
				{"if": ifOp("update"), "then": req("device_id")},
				{"if": ifOp("poweroff"), "then": req("device_id")},
				{"if": ifOp("reboot"), "then": req("device_id")},

				// v1.27.0
				{"if": ifOp("get_network"), "then": req("device_id")},
				{"if": ifOp("update_network"), "then": req("device_id")},
				{"if": ifOp("list_vlans"), "then": req("device_id")},
				{"if": ifOp("get_vlan"), "then": req("device_id", "vlan_id")},
				{"if": ifOp("create_vlan"), "then": req("device_id", "name", "vlan_tag", "network_mode")},
				{"if": ifOp("update_vlan"), "then": req("device_id", "vlan_id")},
				{"if": ifOp("delete_vlan"), "then": req("device_id", "vlan_id")},
			},
		},
	}
}

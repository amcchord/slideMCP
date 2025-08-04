package main

// handleVMsTool handles all virtual machine-related operations through a single meta-tool
func handleVMsTool(args map[string]interface{}) (string, error) {

	return HandleToolWithOperations(CreateToolConfig("slide_vms", ToolOperations{
		"list":             listVirtualMachines,
		"get":              getVirtualMachine,
		"create":           createVirtualMachine,
		"update":           updateVirtualMachine,
		"delete":           deleteVirtualMachine,
		"get_rdp_bookmark": generateRDPBookmark,
	}), args)
}

// getVMsToolInfo returns the tool definition for the VMs meta-tool
func getVMsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_vms",
		Description: "Manage virtual machines created from snapshots. Virtual machines allow you to boot and interact with backed-up systems for testing, recovery, or migration purposes. Includes RDP bookmark generation for easy desktop access.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "get", "create", "update", "delete", "get_rdp_bookmark"},
				},
				// Common parameters for list operations
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Number of results per page (max 50) - used with 'list' operation",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Pagination offset - used with 'list' operation",
				},
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with 'list' operation",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with 'list' operation",
					"enum":        []string{"created"},
				},
				// VM identification
				"virt_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the virtual machine - required for 'get', 'update', and 'delete' operations",
				},
				// VM creation parameters
				"snapshot_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the snapshot to create VM from - required for 'create' operation",
				},
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the device to create VM on - required for 'create' operation",
				},
				"cpu_count": map[string]interface{}{
					"type":        "number",
					"description": "Number of CPUs for the VM - used with 'create' and 'update' operations",
				},
				"memory_in_mb": map[string]interface{}{
					"type":        "number",
					"description": "Memory in MB for the VM - used with 'create' and 'update' operations",
				},
				"disk_bus": map[string]interface{}{
					"type":        "string",
					"description": "Disk bus type - used with 'create' operation",
					"enum":        []string{"sata", "virtio"},
				},
				"network_model": map[string]interface{}{
					"type":        "string",
					"description": "Network model - used with 'create' operation",
					"enum":        []string{"hypervisor_default", "e1000", "rtl8139"},
				},
				"network_type": map[string]interface{}{
					"type":        "string",
					"description": "Network type - used with 'create' operation",
					"enum":        []string{"network", "network-isolated", "bridge", "network-id"},
				},
				"network_source": map[string]interface{}{
					"type":        "string",
					"description": "Network source ID - used with 'create' operation when network_type is 'network-id'",
				},
				"boot_mods": map[string]interface{}{
					"type":        "array",
					"description": "Optional boot modifications - used with 'create' operation",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				// VM update parameters
				"state": map[string]interface{}{
					"type":        "string",
					"description": "VM state - used with 'update' operation",
					"enum":        []string{"running", "stopped", "paused"},
				},
				"expires_at": map[string]interface{}{
					"type":        "string",
					"description": "Expiration timestamp - used with 'update' operation",
				},
			},
			"required": []string{"operation"},
			"allOf": []map[string]interface{}{
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"virt_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "create"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"snapshot_id", "device_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"virt_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "delete"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"virt_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get_rdp_bookmark"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"virt_id"},
					},
				},
			},
		},
	}
}

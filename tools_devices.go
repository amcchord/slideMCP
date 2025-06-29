package main

import (
	"fmt"
)

// handleDevicesTool handles all device-related operations through a single meta-tool
func handleDevicesTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !isOperationAllowed("slide_devices", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_devices in '%s' mode", operation, toolsMode)
	}

	switch operation {
	// Device operations
	case "list":
		return listDevices(args)
	case "get":
		return getDevice(args)
	case "update":
		return updateDevice(args)
	case "poweroff":
		return powerOffDevice(args)
	case "reboot":
		return rebootDevice(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getDevicesToolInfo returns the tool definition for the devices meta-tool
func getDevicesToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_devices",
		Description: "Manage physical devices (Slide appliances). Devices are the physical hardware that run the Slide backup software and store backups.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "get", "update", "poweroff", "reboot"},
				},
				// Common parameters for list operations
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
				// Device parameters
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the device - required for device operations and VM creation",
				},
				"display_name": map[string]interface{}{
					"type":        "string",
					"description": "Display name for the device - used with update_device operation",
				},
				"hostname": map[string]interface{}{
					"type":        "string",
					"description": "Hostname for the device - used with 'update' operation",
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
						"required": []string{"device_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"device_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "poweroff"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"device_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "reboot"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"device_id"},
					},
				},
			},
		},
	}
}

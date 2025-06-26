package main

import (
	"fmt"
)

// handleAlertsTool handles all alert-related operations through a single meta-tool
func handleAlertsTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	switch operation {
	case "list":
		return listAlerts(args)
	case "get":
		return getAlert(args)
	case "update":
		return updateAlert(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getAlertsToolInfo returns the tool definition for the alerts meta-tool
func getAlertsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_alerts",
		Description: "Manage system alerts and notifications. Supports list, get, and update (resolve/unresolve) operations.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "get", "update"},
				},
				// Parameters for list operation
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Number of results per page (max 50) - used with 'list' operation",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Pagination offset - used with 'list' operation",
				},
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by device ID - used with 'list' operation",
				},
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by agent ID - used with 'list' operation",
				},
				"resolved": map[string]interface{}{
					"type":        "boolean",
					"description": "Filter by resolved status - used with 'list' operation, or required for 'update' operation",
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
				// Parameters for get and update operations
				"alert_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the alert - required for 'get' and 'update' operations",
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
						"required": []string{"alert_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"alert_id", "resolved"},
					},
				},
			},
		},
	}
}

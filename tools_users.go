package main

import (
	"fmt"
)

// handleUsersTool handles all user-related operations through a single meta-tool
func handleUsersTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !isOperationAllowed("slide_users", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_users in '%s' mode", operation, toolsMode)
	}

	switch operation {
	case "list":
		return listUsers(args)
	case "get":
		return getUser(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getUsersToolInfo returns the tool definition for the users meta-tool
func getUsersToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_users",
		Description: "Manage user accounts and view user information. Supports list and get operations.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "get"},
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
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with 'list' operation",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with 'list' operation",
					"enum":        []string{"id"},
				},
				// Parameters for get operation
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the user - required for 'get' operation",
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
						"required": []string{"user_id"},
					},
				},
			},
		},
	}
}

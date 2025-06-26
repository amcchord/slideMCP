package main

import (
	"fmt"
)

// handleAgentsTool handles all agent-related operations through a single meta-tool
func handleAgentsTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	switch operation {
	case "list":
		return listAgents(args)
	case "get":
		return getAgent(args)
	case "create":
		return createAgent(args)
	case "pair":
		return pairAgent(args)
	case "update":
		return updateAgent(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getAgentsToolInfo returns the tool definition for the agents meta-tool
func getAgentsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_agents",
		Description: "Manage agents - software installed on computers that get backed up to Slide devices. Supports list, get, create, pair, and update operations.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "get", "create", "pair", "update"},
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
					"description": "Filter by device ID - used with 'list' operation, or required for 'create' and 'pair' operations",
				},
				"client_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by client ID - used with 'list' operation",
				},
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with 'list' operation",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with 'list' operation",
					"enum":        []string{"id", "hostname", "name"},
				},
				// Parameters for get operation
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the agent - required for 'get' and 'update' operations",
				},
				// Parameters for create operation
				"display_name": map[string]interface{}{
					"type":        "string",
					"description": "Display name for the agent - required for 'create' and 'update' operations",
				},
				// Parameters for pair operation
				"pair_code": map[string]interface{}{
					"type":        "string",
					"description": "Pair code generated during agent creation - required for 'pair' operation",
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
						"required": []string{"agent_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "create"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"display_name", "device_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "pair"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"pair_code", "device_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"agent_id", "display_name"},
					},
				},
			},
		},
	}
}

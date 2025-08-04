package main

// handleAgentsTool handles all agent-related operations through a single meta-tool
func handleAgentsTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_agents", ToolOperations{
		"list":              listAgents,
		"get":               getAgent,
		"create":            createAgent,
		"pair":              pairAgent,
		"update":            updateAgent,
		"add_passphrase":    addAgentPassphrase,
		"delete_passphrase": deleteAgentPassphrase,
	}), args)
}

// getAgentsToolInfo returns the tool definition for the agents meta-tool
func getAgentsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_agents",
		Description: "Manage agents - software installed on computers that get backed up to Slide devices. Supports list, get, create, pair, update, add_passphrase, and delete_passphrase operations. Includes support for VSS writer configuration and passphrase management.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "get", "create", "pair", "update", "add_passphrase", "delete_passphrase"},
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
				// Parameters for passphrase operations
				"passphrase_name": map[string]interface{}{
					"type":        "string",
					"description": "Friendly name for the passphrase - required for 'add_passphrase' operation",
				},
				"passphrase": map[string]interface{}{
					"type":        "string",
					"description": "The passphrase to add - required for 'add_passphrase' operation, or current passphrase for 'delete_passphrase' operation",
				},
				"agent_passphrase_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the passphrase to delete - required for 'delete_passphrase' operation",
				},
				// Parameters for VSS writer configuration (update operation)
				"vss_writer_configs": map[string]interface{}{
					"type":        "array",
					"description": "VSS writer configurations - used with 'update' operation",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"writer_id": map[string]interface{}{
								"type": "string",
							},
							"excluded": map[string]interface{}{
								"type": "boolean",
							},
						},
						"required": []string{"writer_id", "excluded"},
					},
				},
				// Parameters for sealed status (update operation)
				"sealed": map[string]interface{}{
					"type":        "boolean",
					"description": "Set to false to unseal an agent with a user-managed passphrase - used with 'update' operation",
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
						"required": []string{"agent_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "add_passphrase"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"agent_id", "passphrase_name", "passphrase"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "delete_passphrase"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"agent_id", "agent_passphrase_id", "passphrase"},
					},
				},
			},
		},
	}
}

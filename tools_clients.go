package main

// slide_clients: client (organisational unit) CRUD. Split out of the
// v3 slide_user_management mega-tool because in MSP workflows clients
// are touched far more often than users/accounts.

func handleClientsTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_clients", ToolOperations{
		"list":   listClients,
		"get":    getClient,
		"create": createClient,
		"update": updateClient,
		"delete": deleteClient,
	}), args)
}

var clientsOperationEnums = []string{"list", "get", "create", "update", "delete"}

func getClientsToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        clientsOperationEnums,
		},
		"client_id": map[string]interface{}{
			"type":        "string",
			"description": "Client ID. Required for `get`, `update`, `delete`.",
		},
		"name": map[string]interface{}{
			"type":        "string",
			"description": "Client name. Required for `create`, optional for `update`.",
		},
		"comments": map[string]interface{}{
			"type":        "string",
			"description": "Free-form notes.",
		},
	}
	for k, v := range commonListProperties() {
		if _, exists := props[k]; !exists {
			props[k] = v
		}
	}

	return ToolInfo{
		Name: "slide_clients",
		Description: "Manage clients (organisational units, typically end-customers in an MSP scenario). " +
			"Operations: `list`, `get`, `create`, `update`, `delete`. Use `slide_overview for_client` for a richer client + devices + agents view.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get"), "then": req("client_id")},
				{"if": ifOp("create"), "then": req("name")},
				{"if": ifOp("update"), "then": req("client_id")},
				{"if": ifOp("delete"), "then": req("client_id")},
			},
		},
	}
}

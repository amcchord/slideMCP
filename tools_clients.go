package main

// slide_clients: client (organisational unit) CRUD. Split out of the
// v3 slide_user_management mega-tool because in MSP workflows clients
// are touched far more often than users/accounts.

func handleClientsTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfigWithResolutions("slide_clients", ToolOperations{
		"list":   listClients,
		"get":    getClient,
		"create": createClient,
		"update": updateClient,
		"delete": deleteClient,
	}, map[string]ResolutionSpec{
		"get":    {IDKey: "client_id", Kind: "client"},
		"update": {IDKey: "client_id", Kind: "client"},
		"delete": {IDKey: "client_id", Kind: "client"},
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
			"description": "Client ID. Required for `get`, `update`, `delete` (alternative: `name_hint`).",
		},
		"name_hint": map[string]interface{}{
			"type":        "string",
			"description": "Alternative to client_id on `get` / `update` / `delete`: a client name (case-insensitive substring match).",
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
		Description: "Slide MCP - manage MSP clients (organisational units). " +
			"REACH FOR THIS whenever the user mentions an MSP client by name, 'add a new client', 'list my clients', " +
			"'who are my customers in Slide', or wants to update a client's details. " +
			"Operations: `list`, `get`, `create`, `update`, `delete`. Accepts client_id OR name_hint on single-client operations. " +
			"Use `slide_overview for_client` for a richer client + devices + agents view.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get"), "then": reqEither("client_id", "name_hint")},
				{"if": ifOp("create"), "then": req("name")},
				{"if": ifOp("update"), "then": reqEither("client_id", "name_hint")},
				{"if": ifOp("delete"), "then": reqEither("client_id", "name_hint")},
			},
		},
	}
}

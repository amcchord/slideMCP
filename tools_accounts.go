package main

import (
	"fmt"
)

// handleAccountsTool handles all account and client-related operations through a single meta-tool
func handleAccountsTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !isOperationAllowed("slide_accounts", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_accounts in '%s' mode", operation, toolsMode)
	}

	switch operation {
	// Account operations
	case "list_accounts":
		return listAccounts(args)
	case "get_account":
		return getAccount(args)
	case "update_account":
		return updateAccount(args)
	// Client operations
	case "list_clients":
		return listClients(args)
	case "get_client":
		return getClient(args)
	case "create_client":
		return createClient(args)
	case "update_client":
		return updateClient(args)
	case "delete_client":
		return deleteClient(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getAccountsToolInfo returns the tool definition for the accounts meta-tool
func getAccountsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_accounts",
		Description: "Manage accounts and clients - organizational structures for managing devices and services. Supports operations for both accounts and clients.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list_accounts", "get_account", "update_account", "list_clients", "get_client", "create_client", "update_client", "delete_client"},
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
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with list operations",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with list operations",
					"enum":        []string{"name", "id"},
				},
				// Account parameters
				"account_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the account - required for get_account and update_account operations",
				},
				"alert_emails": map[string]interface{}{
					"type":        "array",
					"description": "List of email addresses for alerts - required for update_account operation",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				// Client parameters
				"client_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the client - required for get_client, update_client, and delete_client operations",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the client - required for create_client operation",
				},
				"comments": map[string]interface{}{
					"type":        "string",
					"description": "Comments about the client - used with create_client and update_client operations",
				},
			},
			"required": []string{"operation"},
			"allOf": []map[string]interface{}{
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get_account"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"account_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update_account"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"account_id", "alert_emails"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get_client"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"client_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "create_client"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"name"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update_client"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"client_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "delete_client"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"client_id"},
					},
				},
			},
		},
	}
}

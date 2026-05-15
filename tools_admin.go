package main

// slide_admin: users + accounts + user avatar. Renamed from
// slide_user_management; clients moved to their own slide_clients tool.

import "fmt"

func handleAdminTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_admin", ToolOperations{
		"list_users":      listUsers,
		"get_user":        getUser,
		"get_user_avatar": handleGetUserAvatar,
		"list_accounts":   listAccounts,
		"get_account":     getAccount,
		"update_account":  updateAccount,
	}), args)
}

var adminOperationEnums = []string{"list_users", "get_user", "get_user_avatar", "list_accounts", "get_account", "update_account"}

func getAdminToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        adminOperationEnums,
		},
		"user_id": map[string]interface{}{
			"type":        "string",
			"description": "User ID. Required for `get_user`, `get_user_avatar`.",
		},
		"account_id": map[string]interface{}{
			"type":        "string",
			"description": "Account ID. Required for `get_account`, `update_account`.",
		},
		"alert_emails": map[string]interface{}{
			"type":        "array",
			"description": "Email addresses to receive account-level alerts. Required for `update_account`.",
			"items":       map[string]interface{}{"type": "string"},
		},
	}
	for k, v := range commonListProperties() {
		if _, exists := props[k]; !exists {
			props[k] = v
		}
	}

	return ToolInfo{
		Name: "slide_admin",
		Description: "Slide MCP - Slide account + user administration. " +
			"REACH FOR THIS whenever the user mentions 'Slide users', 'who has access to Slide', 'add a user', " +
			"'change alert email recipients', 'Slide account settings', or any account-level admin task. " +
			"Operations: `list_users`, `get_user`, `get_user_avatar` (returns a data: URL), " +
			"`list_accounts`, `get_account`, `update_account` (currently just `alert_emails`). " +
			"Note: client management moved to `slide_clients`.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get_user"), "then": req("user_id")},
				{"if": ifOp("get_user_avatar"), "then": req("user_id")},
				{"if": ifOp("get_account"), "then": req("account_id")},
				{"if": ifOp("update_account"), "then": req("account_id", "alert_emails")},
			},
		},
	}
}

// (handleGetUserAvatar continues to live in tools_v127.go)
var _ = fmt.Sprintf

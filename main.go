package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// MCP Protocol structures
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type ToolInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Configuration
const (
	ServerName = "slide-mcp-server"
	Version    = "2.3.0"
)

// Tools filtering modes
const (
	ToolsReporting = "reporting"
	ToolsRestores  = "restores"
	ToolsFullSafe  = "full-safe"
	ToolsFull      = "full"
)

var toolsMode string = ToolsFullSafe // Default to full-safe access
var disabledTools []string           // List of disabled tool names
// Tool enablement flags - presentation and reports tools are disabled by default
var enablePresentation bool = false
var enableReports bool = false

// Helper functions for tools filtering
func isToolAllowed(toolName string) bool {
	// First check if tool is explicitly disabled
	if isToolDisabled(toolName) {
		return false
	}

	// Check if presentation or reports tools need explicit enablement
	if toolName == "slide_presentation" && !enablePresentation {
		return false
	}
	if toolName == "slide_reports" && !enableReports {
		return false
	}

	// Then check tools mode permissions
	switch toolsMode {
	case ToolsReporting:
		// Only allow read-only tools
		return isReadOnlyTool(toolName)
	case ToolsRestores:
		// Allow reporting tools + restore/VM/network management
		return isReadOnlyTool(toolName) || isRestoreManagementTool(toolName)
	case ToolsFullSafe:
		// Allow everything except dangerous operations
		return !isDangerousTool(toolName)
	case ToolsFull:
		// Allow all tools
		return true
	default:
		return false
	}
}

func isToolDisabled(toolName string) bool {
	for _, disabled := range disabledTools {
		if disabled == toolName {
			return true
		}
	}
	return false
}

func isOperationAllowed(toolName, operation string) bool {
	switch toolsMode {
	case ToolsReporting:
		// Only allow read operations
		return isReadOperation(operation)
	case ToolsRestores:
		// Allow read operations + specific management operations
		return isReadOperation(operation) || isRestoreManagementOperation(toolName, operation)
	case ToolsFullSafe:
		// Allow everything except dangerous operations
		return !isDangerousOperation(toolName, operation)
	case ToolsFull:
		// Allow all operations
		return true
	default:
		return false
	}
}

func isReadOnlyTool(toolName string) bool {
	readOnlyTools := []string{
		"slide_agents", "slide_backups", "slide_snapshots", "slide_user_management",
		"slide_alerts", "slide_devices", "slide_networks",
		"slide_vms", "slide_restores", "slide_presentation", "slide_meta", "slide_docs", "list_all_clients_devices_and_agents",
	}
	for _, tool := range readOnlyTools {
		if tool == toolName {
			return true
		}
	}
	return false
}

func isRestoreManagementTool(toolName string) bool {
	restoreTools := []string{
		"slide_vms", "slide_restores", "slide_networks",
	}
	for _, tool := range restoreTools {
		if tool == toolName {
			return true
		}
	}
	return false
}

func isDangerousTool(toolName string) bool {
	// No tools are completely dangerous in full-safe mode
	// Danger is at the operation level
	return false
}

func isReadOperation(operation string) bool {
	readOps := []string{
		"list", "get", "browse", "list_deleted",
		// Restores tool read operations
		"list_files", "get_file", "browse_file", "list_images", "get_image", "browse_image",
		// User management tool read operations
		"list_users", "get_user", "list_accounts", "get_account", "list_clients", "get_client",
		// Reports tool read operations
		"get_runbook_template",
	}
	for _, op := range readOps {
		if op == operation {
			return true
		}
	}
	return false
}

func isRestoreManagementOperation(toolName, operation string) bool {
	switch toolName {
	case "slide_vms":
		return operation == "create" || operation == "update" || operation == "delete"
	case "slide_restores":
		return operation == "create_file" || operation == "delete_file" || operation == "browse_file" ||
			operation == "create_image" || operation == "delete_image" || operation == "browse_image"
	case "slide_networks":
		return operation == "create" || operation == "update" || operation == "delete" ||
			operation == "create_ipsec" || operation == "update_ipsec" || operation == "delete_ipsec" ||
			operation == "create_port_forward" || operation == "update_port_forward" || operation == "delete_port_forward" ||
			operation == "create_wg_peer" || operation == "update_wg_peer" || operation == "delete_wg_peer"
	case "slide_devices":
		// Removed device power control (poweroff, reboot) from restores mode
		return operation == "update"
	case "slide_agents":
		return operation == "create" || operation == "pair" || operation == "update"
	case "slide_backups":
		return operation == "start"
	case "slide_user_management":
		return operation == "update_account" || operation == "create_client" || operation == "update_client" || operation == "delete_client"
		// Removed alert resolution from restores mode
	}
	return false
}

func isDangerousOperation(toolName, operation string) bool {
	// Define dangerous operations that are blocked in full-safe mode
	if toolName == "slide_agents" && operation == "delete" {
		return true
	}
	if toolName == "slide_snapshots" && operation == "delete" {
		return true
	}
	// Block device power control operations in full-safe mode
	if toolName == "slide_devices" && (operation == "poweroff" || operation == "reboot") {
		return true
	}
	return false
}

func main() {
	// Parse command line flags
	var cliAPIKey = flag.String("api-key", "", "API key for Slide service (overrides SLIDE_API_KEY environment variable)")
	var cliBaseURL = flag.String("base-url", "", "Base URL for Slide API (overrides SLIDE_BASE_URL environment variable)")
	var cliTools = flag.String("tools", "", "Tools mode: reporting, restores, full-safe, full (overrides SLIDE_TOOLS environment variable)")
	var cliDisabledTools = flag.String("disabled-tools", "", "Comma-separated list of tool names to disable (overrides SLIDE_DISABLED_TOOLS environment variable)")
	var cliEnablePresentation = flag.Bool("enable-presentation", false, "Enable the slide_presentation tool (overrides SLIDE_ENABLE_PRESENTATION environment variable)")
	var cliEnableReports = flag.Bool("enable-reports", false, "Enable the slide_reports tool (overrides SLIDE_ENABLE_REPORTS environment variable)")
	var showVersion = flag.Bool("version", false, "Show version information and exit")
	var exitAfterFirst = flag.Bool("exit-after-first", false, "Exit after processing the first request instead of running continuously")
	// One-shot tool execution flags
	var cliOneShotTool = flag.String("tool", "", "Run a single tool then exit (e.g. --tool slide_reports)")
	var cliToolArgs = flag.String("args", "", "JSON string with arguments for --tool (e.g. '{\"operation\":\"daily_backup_snapshot\"}')")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("%s version %s\n", ServerName, Version)
		os.Exit(0)
	}

	// Get tools mode from CLI flag or environment variable
	// CLI flag takes precedence over environment variable
	if *cliTools != "" {
		toolsMode = *cliTools
	} else if envTools := os.Getenv("SLIDE_TOOLS"); envTools != "" {
		toolsMode = envTools
	}
	// If neither is provided, toolsMode keeps its default value (full-safe)

	// Get disabled tools from CLI flag or environment variable
	// CLI flag takes precedence over environment variable
	var disabledToolsStr string
	if *cliDisabledTools != "" {
		disabledToolsStr = *cliDisabledTools
	} else if envDisabledTools := os.Getenv("SLIDE_DISABLED_TOOLS"); envDisabledTools != "" {
		disabledToolsStr = envDisabledTools
	}

	// Parse disabled tools list
	if disabledToolsStr != "" {
		// Split by comma and trim whitespace
		toolsList := strings.Split(disabledToolsStr, ",")
		for _, tool := range toolsList {
			tool = strings.TrimSpace(tool)
			if tool != "" {
				disabledTools = append(disabledTools, tool)
			}
		}
		log.Printf("Disabled tools: %v", disabledTools)
	}

	// Get presentation tool enablement from CLI flag or environment variable
	// CLI flag takes precedence over environment variable
	if *cliEnablePresentation {
		enablePresentation = true
	} else if envEnablePresentation := os.Getenv("SLIDE_ENABLE_PRESENTATION"); envEnablePresentation != "" {
		enablePresentation = envEnablePresentation == "true" || envEnablePresentation == "1"
	}

	// Get reports tool enablement from CLI flag or environment variable
	// CLI flag takes precedence over environment variable
	if *cliEnableReports {
		enableReports = true
	} else if envEnableReports := os.Getenv("SLIDE_ENABLE_REPORTS"); envEnableReports != "" {
		enableReports = envEnableReports == "true" || envEnableReports == "1"
	}

	if enablePresentation {
		log.Printf("Presentation tool enabled")
	}
	if enableReports {
		log.Printf("Reports tool enabled")
	}

	// Validate tools mode
	switch toolsMode {
	case ToolsReporting, ToolsRestores, ToolsFullSafe, ToolsFull:
		// Valid mode
	default:
		log.Fatalf("Error: Invalid tools mode '%s'. Valid options: reporting, restores, full-safe, full", toolsMode)
	}

	// Get base URL from CLI flag or environment variable
	// CLI flag takes precedence over environment variable
	if *cliBaseURL != "" {
		APIBaseURL = *cliBaseURL
	} else if envBaseURL := os.Getenv("SLIDE_BASE_URL"); envBaseURL != "" {
		APIBaseURL = envBaseURL
	}
	// If neither is provided, APIBaseURL keeps its default value

	// Get API key from CLI flag or environment variable
	// CLI flag takes precedence over environment variable
	if *cliAPIKey != "" {
		apiKey = *cliAPIKey
	} else {
		apiKey = os.Getenv("SLIDE_API_KEY")
	}

	if apiKey == "" {
		log.Fatalf("Error: API key not provided. Use --api-key flag or set SLIDE_API_KEY environment variable.")
	}

	// -------------------------------------------------
	// One-shot tool execution (if --tool flag provided)
	// -------------------------------------------------
	if *cliOneShotTool != "" {
		// Parse JSON args (if provided)
		var argsMap map[string]interface{}
		if *cliToolArgs != "" {
			if err := json.Unmarshal([]byte(*cliToolArgs), &argsMap); err != nil {
				log.Fatalf("Invalid JSON for --args: %v", err)
			}
		} else {
			argsMap = make(map[string]interface{})
		}

		// Construct a mock MCPRequest to reuse existing handler logic
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      *cliOneShotTool,
				"arguments": argsMap,
			},
		}

		resp := handleToolCall(req)

		// Marshal and print the response so callers can parse
		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			log.Fatalf("Failed to encode response: %v", err)
		}
		fmt.Println(string(out))
		return // Exit after one-shot execution
	}

	log.Println("Slide MCP Server starting...")

	// Start MCP server
	startMCPServer(*exitAfterFirst)
}

func startMCPServer(exitAfterFirst bool) {
	scanner := bufio.NewScanner(os.Stdin)
	requestCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var request MCPRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			// Only send error response if we can determine there was an ID
			var rawMsg map[string]interface{}
			if json.Unmarshal([]byte(line), &rawMsg) == nil {
				if id, exists := rawMsg["id"]; exists {
					response := sendError(id, -32700, "Parse error", nil)
					if responseJSON, err := json.Marshal(response); err == nil {
						fmt.Println(string(responseJSON))
					}
				}
			}
			continue
		}

		// Check if this is a notification (no ID field)
		var rawMsg map[string]interface{}
		json.Unmarshal([]byte(line), &rawMsg)
		_, hasID := rawMsg["id"]

		if !hasID {
			// This is a notification - handle it but don't send a response
			handleNotification(request)
			continue
		}

		// This is a request - handle it and send a response
		response := handleRequest(request)

		responseJSON, err := json.Marshal(response)
		if err != nil {
			errorResponse := sendError(request.ID, -32603, "Internal error", nil)
			if errorJSON, err := json.Marshal(errorResponse); err == nil {
				fmt.Println(string(errorJSON))
			}
			continue
		}

		fmt.Println(string(responseJSON))

		// Increment request count and check if we should exit
		requestCount++
		if exitAfterFirst && requestCount >= 1 {
			log.Println("Exiting after processing first request as requested")
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}

func fetchInitialContext() map[string]interface{} {
	// Fetch the initial context data that's typically requested first
	contextData, err := listAllClientsDevicesAndAgents(map[string]interface{}{})
	if err != nil {
		log.Printf("Warning: Failed to fetch initial context data: %v", err)
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to fetch initial context: %v", err),
			"note":  "Initial context will be available via the list_all_clients_devices_and_agents tool",
		}
	}

	// Parse the JSON string back to a map for inclusion
	var contextMap map[string]interface{}
	if err := json.Unmarshal([]byte(contextData), &contextMap); err != nil {
		log.Printf("Warning: Failed to parse initial context data: %v", err)
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to parse initial context: %v", err),
			"note":  "Initial context will be available via the list_all_clients_devices_and_agents tool",
		}
	}

	return contextMap
}

func handleNotification(request MCPRequest) {
	switch request.Method {
	case "notifications/initialized":
		// Client has initialized - no response needed
		log.Println("Client initialized")
	case "notifications/cancelled":
		// Request was cancelled - no response needed
		log.Println("Request cancelled")
	default:
		// Unknown notification - just log it
		log.Printf("Unknown notification: %s", request.Method)
	}
}

func handleRequest(request MCPRequest) MCPResponse {
	switch request.Method {
	case "initialize":
		// Fetch initial context data
		initialContext := fetchInitialContext()

		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    ServerName,
					"version": Version,
				},
				"initialContext": map[string]interface{}{
					"clients_devices_agents": initialContext,
					"_metadata": map[string]interface{}{
						"description": "Initial overview of all clients, devices, and agents loaded at startup for improved performance",
						"source_tool": "list_all_clients_devices_and_agents",
						"usage_note":  "This data is also available via the list_all_clients_devices_and_agents tool and should be refreshed if needed",
						"timestamp":   fmt.Sprintf("%d", time.Now().Unix()),
					},
				},
			},
		}

	case "tools/list":
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result: map[string]interface{}{
				"tools": getAllTools(),
			},
		}

	case "tools/call":
		return handleToolCall(request)

	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
			},
		}
	}
}

func handleToolCall(request MCPRequest) MCPResponse {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return sendError(request.ID, -32602, "Invalid params", nil)
	}

	name, ok := params["name"].(string)
	if !ok {
		return sendError(request.ID, -32602, "Tool name required", nil)
	}

	// Check if tool is explicitly disabled
	if isToolDisabled(name) {
		return sendError(request.ID, -32601, fmt.Sprintf("Tool '%s' is disabled", name), nil)
	}

	// Check if tool is allowed in current tools mode
	if !isToolAllowed(name) {
		return sendError(request.ID, -32601, fmt.Sprintf("Tool '%s' not available in '%s' mode", name, toolsMode), nil)
	}

	args, ok := params["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	var result ToolResult

	switch name {
	// Meta-tools
	case "slide_agents":
		data, err := handleAgentsTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_backups":
		data, err := handleBackupsTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_snapshots":
		data, err := handleSnapshotsTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_restores":
		data, err := handleRestoresTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_networks":
		data, err := handleNetworksTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_user_management":
		data, err := handleUserManagementTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_alerts":
		data, err := handleAlertsTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_devices":
		data, err := handleDevicesTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_vms":
		data, err := handleVMsTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_presentation":
		data, err := handlePresentationTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_reports":
		data, err := handleReportsTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_meta":
		data, err := handleMetaTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_docs":
		data, err := handleDocsTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	// Special tools (keep as-is) - now handled by slide_meta
	case "list_all_clients_devices_and_agents":
		// Redirect to slide_meta for backward compatibility
		args["operation"] = "list_all_clients_devices_and_agents"
		data, err := handleMetaTool(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	default:
		result = ToolResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", name)}},
			IsError: true,
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

func sendError(id interface{}, code int, message string, data interface{}) MCPResponse {
	errorObj := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	if data != nil {
		errorObj["data"] = data
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   errorObj,
	}
}

func getAllTools() []ToolInfo {
	allTools := []ToolInfo{
		// Meta-tools
		getAgentsToolInfo(),
		getBackupsToolInfo(),
		getSnapshotsToolInfo(),
		getRestoresToolInfo(),
		getNetworksToolInfo(),
		getUserManagementToolInfo(),
		getAlertsToolInfo(),
		getDevicesToolInfo(),
		getVMsToolInfo(),
		getPresentationToolInfo(),
		getReportsToolInfo(),
		getMetaToolInfo(),
		getDocsToolInfo(), // Documentation access tool
		// Special tools (kept for backward compatibility)
		{
			Name:        "list_all_clients_devices_and_agents",
			Description: "Get a complete hierarchical view of all clients, their devices, and the agents on those devices. Use this tool when answers questions about how many agents, devices, or clients",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	// Filter tools based on current tools mode
	var filteredTools []ToolInfo
	for _, tool := range allTools {
		if isToolAllowed(tool.Name) {
			// For reporting mode, update tool descriptions to indicate read-only access
			if toolsMode == ToolsReporting {
				tool.Description = tool.Description + " (Read-only mode: only list/get operations available)"
			}
			filteredTools = append(filteredTools, tool)
		}
	}

	return filteredTools
}

func getOldAllTools() []ToolInfo {
	return []ToolInfo{
		{
			Name:        "slide_list_devices",
			Description: "List all devices with pagination and filtering options. Hostname is the primary identifier for devices.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by client ID",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (hostname)",
						"enum":        []string{"hostname"},
					},
				},
			},
		},
		{
			Name:        "slide_get_device",
			Description: "Get detailed information about a specific device by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the device to retrieve",
					},
				},
				"required": []string{"device_id"},
			},
		},
		{
			Name:        "slide_update_device",
			Description: "Update a device's properties",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the device to update",
					},
					"display_name": map[string]interface{}{
						"type":        "string",
						"description": "New display name for the device",
					},
					"hostname": map[string]interface{}{
						"type":        "string",
						"description": "New hostname for the device",
					},
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "New client ID for the device",
					},
				},
				"required": []string{"device_id"},
			},
		},
		{
			Name:        "slide_poweroff_device",
			Description: "Power off a device remotely",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the device to power off",
					},
				},
				"required": []string{"device_id"},
			},
		},
		{
			Name:        "slide_reboot_device",
			Description: "Reboot a device remotely",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the device to reboot",
					},
				},
				"required": []string{"device_id"},
			},
		},
		{
			Name:        "slide_list_agents",
			Description: "List all agents with pagination and filtering options. Display Name is the primary identifier for agents.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by device ID",
					},
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by client ID",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (id, hostname, name)",
						"enum":        []string{"id", "hostname", "name"},
					},
				},
			},
		},
		{
			Name:        "slide_get_agent",
			Description: "Get detailed information about a specific agent by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the agent to retrieve",
					},
				},
				"required": []string{"agent_id"},
			},
		},
		{
			Name:        "slide_create_agent",
			Description: "Create an agent for auto-pair installation",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"display_name": map[string]interface{}{
						"type":        "string",
						"description": "Display name for the agent",
					},
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the device to associate with the agent",
					},
				},
				"required": []string{"display_name", "device_id"},
			},
		},
		{
			Name:        "slide_pair_agent",
			Description: "Pair an agent with a device using a pair code",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pair_code": map[string]interface{}{
						"type":        "string",
						"description": "Pair code generated during agent creation",
					},
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the device to pair with",
					},
				},
				"required": []string{"pair_code", "device_id"},
			},
		},
		{
			Name:        "slide_update_agent",
			Description: "Update an agent's properties",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the agent to update",
					},
					"display_name": map[string]interface{}{
						"type":        "string",
						"description": "New display name for the agent",
					},
				},
				"required": []string{"agent_id", "display_name"},
			},
		},
		{
			Name:        "slide_list_backups",
			Description: "List all backups with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by agent ID",
					},
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by device ID",
					},
					"snapshot_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by snapshot ID",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (id, start_time)",
						"enum":        []string{"id", "start_time"},
					},
				},
			},
		},
		{
			Name:        "slide_get_backup",
			Description: "Get detailed information about a specific backup",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"backup_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the backup to retrieve",
					},
				},
				"required": []string{"backup_id"},
			},
		},
		{
			Name:        "slide_start_backup",
			Description: "Start a backup for a specific agent",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the agent to backup",
					},
				},
				"required": []string{"agent_id"},
			},
		},
		{
			Name:        "slide_list_snapshots",
			Description: "List all snapshots with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by agent ID",
					},
					"snapshot_location": map[string]interface{}{
						"type":        "string",
						"description": "Filter by snapshot location. 'exists_local' means the snapshot is on the local device. 'exists_cloud' means the snapshot is on the cloud. 'exists_deleted' means the snapshot was deleted. 'exists_deleted_retention' means the snapshot was deleted by the retention policy. 'exists_deleted_manual' means the snapshot was deleted by a user. 'exists_deleted_other' means the snapshot is deleted for an unknown reason.",
						"enum":        []string{"exists_local", "exists_cloud", "exists_deleted", "exists_deleted_retention", "exists_deleted_manual", "exists_deleted_other"},
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field",
						"enum":        []string{"backup_start_time", "backup_end_time", "created"},
					},
				},
			},
		},
		{
			Name:        "slide_get_snapshot",
			Description: "Get detailed information about a specific snapshot",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"snapshot_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the snapshot to retrieve",
					},
				},
				"required": []string{"snapshot_id"},
			},
		},
		{
			Name:        "slide_list_file_restores",
			Description: "List all file restores with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (id)",
						"enum":        []string{"id"},
					},
				},
			},
		},
		{
			Name:        "slide_get_file_restore",
			Description: "Get detailed information about a specific file restore",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_restore_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the file restore to retrieve",
					},
				},
				"required": []string{"file_restore_id"},
			},
		},
		{
			Name:        "slide_create_file_restore",
			Description: "Create a file restore from a snapshot",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"snapshot_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the snapshot to restore from",
					},
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the device to restore to",
					},
				},
				"required": []string{"snapshot_id", "device_id"},
			},
		},
		{
			Name:        "slide_delete_file_restore",
			Description: "Delete a file restore",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_restore_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the file restore to delete",
					},
				},
				"required": []string{"file_restore_id"},
			},
		},
		{
			Name:        "slide_browse_file_restore",
			Description: "Browse files in a file restore",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_restore_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the file restore to browse",
					},
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to browse within the restore",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
				},
				"required": []string{"file_restore_id", "path"},
			},
		},
		{
			Name:        "slide_list_image_exports",
			Description: "List all image export restores with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (id)",
						"enum":        []string{"id"},
					},
				},
			},
		},
		{
			Name:        "slide_get_image_export",
			Description: "Get detailed information about a specific image export",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image_export_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the image export to retrieve",
					},
				},
				"required": []string{"image_export_id"},
			},
		},
		{
			Name:        "slide_create_image_export",
			Description: "Create an image export from a snapshot",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"snapshot_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the snapshot to export from",
					},
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the device to export to",
					},
					"image_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of image to create",
						"enum":        []string{"vhdx", "vhdx-dynamic", "vhd", "raw"},
					},
					"boot_mods": map[string]interface{}{
						"type":        "array",
						"description": "Optional boot modifications",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"snapshot_id", "device_id", "image_type"},
			},
		},
		{
			Name:        "slide_delete_image_export",
			Description: "Delete an image export",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image_export_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the image export to delete",
					},
				},
				"required": []string{"image_export_id"},
			},
		},
		{
			Name:        "slide_browse_image_export",
			Description: "Browse images in an image export",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image_export_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the image export to browse",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
				},
				"required": []string{"image_export_id"},
			},
		},
		{
			Name:        "slide_list_virtual_machines",
			Description: "List all virtual machines with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (created)",
						"enum":        []string{"created"},
					},
				},
			},
		},
		{
			Name:        "slide_get_virtual_machine",
			Description: "Get detailed information about a specific virtual machine",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"virt_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the virtual machine to retrieve",
					},
				},
				"required": []string{"virt_id"},
			},
		},
		{
			Name:        "slide_create_virtual_machine",
			Description: "Create a virtual machine from a snapshot",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"snapshot_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the snapshot to create VM from",
					},
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the device to create VM on",
					},
					"cpu_count": map[string]interface{}{
						"type":        "number",
						"description": "Number of CPUs for the VM",
					},
					"memory_in_mb": map[string]interface{}{
						"type":        "number",
						"description": "Memory in MB for the VM",
					},
					"disk_bus": map[string]interface{}{
						"type":        "string",
						"description": "Disk bus type",
						"enum":        []string{"sata", "virtio"},
					},
					"network_model": map[string]interface{}{
						"type":        "string",
						"description": "Network model",
						"enum":        []string{"hypervisor_default", "e1000", "rtl8139"},
					},
					"network_type": map[string]interface{}{
						"type":        "string",
						"description": "Network type",
						"enum":        []string{"network", "network-isolated", "bridge", "network-id"},
					},
					"network_source": map[string]interface{}{
						"type":        "string",
						"description": "Network source ID",
					},
					"boot_mods": map[string]interface{}{
						"type":        "array",
						"description": "Optional boot modifications",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"snapshot_id", "device_id"},
			},
		},
		{
			Name:        "slide_update_virtual_machine",
			Description: "Update a virtual machine's properties",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"virt_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the virtual machine to update",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "New state for the VM",
						"enum":        []string{"running", "stopped", "paused"},
					},
					"expires_at": map[string]interface{}{
						"type":        "string",
						"description": "Expiration timestamp",
					},
					"memory_in_mb": map[string]interface{}{
						"type":        "number",
						"description": "Memory in MB",
					},
					"cpu_count": map[string]interface{}{
						"type":        "number",
						"description": "Number of CPUs",
					},
				},
				"required": []string{"virt_id"},
			},
		},
		{
			Name:        "slide_delete_virtual_machine",
			Description: "Delete a virtual machine",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"virt_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the virtual machine to delete",
					},
				},
				"required": []string{"virt_id"},
			},
		},
		{
			Name:        "slide_list_users",
			Description: "List all users with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (id)",
						"enum":        []string{"id"},
					},
				},
			},
		},
		{
			Name:        "slide_get_user",
			Description: "Get detailed information about a specific user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the user to retrieve",
					},
				},
				"required": []string{"user_id"},
			},
		},
		{
			Name:        "slide_list_alerts",
			Description: "List all alerts with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by device ID",
					},
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by agent ID",
					},
					"resolved": map[string]interface{}{
						"type":        "boolean",
						"description": "Filter by resolved status",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (created)",
						"enum":        []string{"created"},
					},
				},
			},
		},
		{
			Name:        "slide_get_alert",
			Description: "Get detailed information about a specific alert",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"alert_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the alert to retrieve",
					},
				},
				"required": []string{"alert_id"},
			},
		},
		{
			Name:        "slide_update_alert",
			Description: "Update an alert's status (resolve/unresolve)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"alert_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the alert to update",
					},
					"resolved": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the alert is resolved",
					},
				},
				"required": []string{"alert_id", "resolved"},
			},
		},
		{
			Name:        "slide_list_accounts",
			Description: "List all accounts with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (name)",
						"enum":        []string{"name"},
					},
				},
			},
		},
		{
			Name:        "slide_get_account",
			Description: "Get detailed information about a specific account",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the account to retrieve",
					},
				},
				"required": []string{"account_id"},
			},
		},
		{
			Name:        "slide_update_account",
			Description: "Update an account's properties",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the account to update",
					},
					"alert_emails": map[string]interface{}{
						"type":        "array",
						"description": "List of email addresses for alerts",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"account_id", "alert_emails"},
			},
		},
		{
			Name:        "slide_list_clients",
			Description: "List all clients with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (id)",
						"enum":        []string{"id"},
					},
				},
			},
		},
		{
			Name:        "slide_get_client",
			Description: "Get detailed information about a specific client",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the client to retrieve",
					},
				},
				"required": []string{"client_id"},
			},
		},
		{
			Name:        "slide_create_client",
			Description: "Create a new client",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the client",
					},
					"comments": map[string]interface{}{
						"type":        "string",
						"description": "Comments about the client",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "slide_update_client",
			Description: "Update a client's properties",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the client to update",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "New name for the client",
					},
					"comments": map[string]interface{}{
						"type":        "string",
						"description": "New comments for the client",
					},
				},
				"required": []string{"client_id"},
			},
		},
		{
			Name:        "slide_delete_client",
			Description: "Delete a client",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the client to delete",
					},
				},
				"required": []string{"client_id"},
			},
		},
		{
			Name:        "slide_list_networks",
			Description: "List all networks with pagination and filtering options",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (id)",
						"enum":        []string{"id"},
					},
				},
			},
		},
		{
			Name:        "slide_get_network",
			Description: "Get detailed information about a specific network",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network to retrieve",
					},
				},
				"required": []string{"network_id"},
			},
		},
		{
			Name:        "slide_create_network",
			Description: "Create a new network for virtual machines. Important: The network's client_id must match the client_id of the VMs that will be placed on this network. Otherwise it will not work.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the network",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Network type",
						"enum":        []string{"standard", "bridge-lan"},
					},
					"bridge_device_id": map[string]interface{}{
						"type":        "string",
						"description": "Device ID for bridge networks",
					},
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "Client ID for the network - should match the client_id of VMs that will use this network",
					},
					"comments": map[string]interface{}{
						"type":        "string",
						"description": "Comments about the network",
					},
					"dhcp": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable DHCP server",
					},
					"dhcp_range_start": map[string]interface{}{
						"type":        "string",
						"description": "DHCP range start address",
					},
					"dhcp_range_end": map[string]interface{}{
						"type":        "string",
						"description": "DHCP range end address",
					},
					"internet": map[string]interface{}{
						"type":        "boolean",
						"description": "Allow internet access",
					},
					"nameservers": map[string]interface{}{
						"type":        "array",
						"description": "DNS servers",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"router_prefix": map[string]interface{}{
						"type":        "string",
						"description": "The router_prefix is the IP address of the router that will be used to connect to the network. It should NOT be the same as the network address (the first IP in the subnet). For example, use '192.168.1.1/24' not '192.168.1.0/24'. When creating standard networks the router_prefix must be in private IP space (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16).",
					},
					"wg": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable WireGuard VPN",
					},
					"wg_prefix": map[string]interface{}{
						"type":        "string",
						"description": "WireGuard network prefix which must not overlap with any other network's prefix",
					},
				},
				"required": []string{"name", "type"},
			},
		},
		{
			Name:        "slide_update_network",
			Description: "Update a network's properties",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network to update",
					},
					"bridge_device_id": map[string]interface{}{
						"type":        "string",
						"description": "Device ID for bridge networks",
					},
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "Client ID for the network",
					},
					"comments": map[string]interface{}{
						"type":        "string",
						"description": "Comments about the network",
					},
					"dhcp": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable DHCP server",
					},
					"dhcp_range_start": map[string]interface{}{
						"type":        "string",
						"description": "DHCP range start address",
					},
					"dhcp_range_end": map[string]interface{}{
						"type":        "string",
						"description": "DHCP range end address",
					},
					"internet": map[string]interface{}{
						"type":        "boolean",
						"description": "Allow internet access",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the network",
					},
					"nameservers": map[string]interface{}{
						"type":        "array",
						"description": "DNS servers",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"router_prefix": map[string]interface{}{
						"type":        "string",
						"description": "Network prefix for router",
					},
					"wg": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable WireGuard VPN",
					},
					"wg_prefix": map[string]interface{}{
						"type":        "string",
						"description": "WireGuard network prefix",
					},
				},
				"required": []string{"network_id"},
			},
		},
		{
			Name:        "slide_delete_network",
			Description: "Delete a network",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network to delete",
					},
				},
				"required": []string{"network_id"},
			},
		},
		{
			Name:        "slide_create_network_ipsec_conn",
			Description: "Create an IPSec connection for a network",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the IPSec connection",
					},
					"remote_addrs": map[string]interface{}{
						"type":        "array",
						"description": "Remote addresses",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"remote_networks": map[string]interface{}{
						"type":        "array",
						"description": "Remote networks",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"network_id", "name", "remote_addrs", "remote_networks"},
			},
		},
		{
			Name:        "slide_update_network_ipsec_conn",
			Description: "Update an IPSec connection for a network",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network",
					},
					"ipsec_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the IPSec connection",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the IPSec connection",
					},
					"remote_addrs": map[string]interface{}{
						"type":        "array",
						"description": "Remote addresses",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"remote_networks": map[string]interface{}{
						"type":        "array",
						"description": "Remote networks",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"network_id", "ipsec_id"},
			},
		},
		{
			Name:        "slide_delete_network_ipsec_conn",
			Description: "Delete an IPSec connection from a network",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network",
					},
					"ipsec_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the IPSec connection",
					},
				},
				"required": []string{"network_id", "ipsec_id"},
			},
		},
		{
			Name:        "slide_create_network_port_forward",
			Description: "Create a port forward for a network",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network",
					},
					"proto": map[string]interface{}{
						"type":        "string",
						"description": "Protocol (tcp/udp)",
						"enum":        []string{"tcp", "udp"},
					},
					"dest": map[string]interface{}{
						"type":        "string",
						"description": "Destination address:port",
					},
				},
				"required": []string{"network_id", "proto", "dest"},
			},
		},
		{
			Name:        "slide_update_network_port_forward",
			Description: "Update a port forward for a network",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network",
					},
					"port_forward_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the port forward",
					},
					"proto": map[string]interface{}{
						"type":        "string",
						"description": "Protocol (tcp/udp)",
						"enum":        []string{"tcp", "udp"},
					},
					"dest": map[string]interface{}{
						"type":        "string",
						"description": "Destination address:port",
					},
				},
				"required": []string{"network_id", "port_forward_id"},
			},
		},
		{
			Name:        "slide_delete_network_port_forward",
			Description: "Delete a port forward from a network",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network",
					},
					"port_forward_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the port forward",
					},
				},
				"required": []string{"network_id", "port_forward_id"},
			},
		},
		{
			Name:        "slide_create_network_wg_peer",
			Description: "Create a WireGuard peer for a network VPN",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network",
					},
					"peer_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the WireGuard peer",
					},
					"remote_networks": map[string]interface{}{
						"type":        "array",
						"description": "Remote networks accessible through this peer",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"network_id", "peer_name"},
			},
		},
		{
			Name:        "slide_update_network_wg_peer",
			Description: "Update a WireGuard peer for a network VPN",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network",
					},
					"wg_peer_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the WireGuard peer",
					},
					"peer_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the WireGuard peer",
					},
					"remote_networks": map[string]interface{}{
						"type":        "array",
						"description": "Remote networks accessible through this peer",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"network_id", "wg_peer_id"},
			},
		},
		{
			Name:        "slide_delete_network_wg_peer",
			Description: "Delete a WireGuard peer from a network VPN",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"network_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the network",
					},
					"wg_peer_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the WireGuard peer",
					},
				},
				"required": []string{"network_id", "wg_peer_id"},
			},
		},
		{
			Name:        "list_all_clients_devices_and_agents",
			Description: "Get a complete hierarchical view of all clients, their devices, and the agents on those devices. This combines multiple API calls into a single comprehensive response that's easier for LLMs to understand and work with.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

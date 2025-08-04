package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// startMCPServer starts the MCP server and handles incoming requests
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

// fetchInitialContext gets initial context data for faster startup
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

// handleNotification processes MCP notifications (no response needed)
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

// handleRequest processes MCP requests and returns responses
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

// handleToolCall processes tool execution requests
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
	if config.IsToolDisabled(name) {
		return sendError(request.ID, -32601, fmt.Sprintf("Tool '%s' is disabled", name), nil)
	}

	// Check if tool is allowed in current tools mode
	if !config.IsToolAllowed(name) {
		return sendError(request.ID, -32601, fmt.Sprintf("Tool '%s' not available in '%s' mode", name, config.ToolsMode), nil)
	}

	args, ok := params["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	// Use tool registry for clean dispatch
	handler, exists := toolRegistry[name]
	if !exists {
		return sendError(request.ID, -32601, fmt.Sprintf("Unknown tool: %s", name), nil)
	}

	// Execute tool handler and create consistent result
	result := createToolResult(handler(args))

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

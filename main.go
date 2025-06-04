package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	Version    = "0.1.0"
)

func main() {
	// Get API key from environment
	apiKey = os.Getenv("SLIDE_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: SLIDE_API_KEY environment variable not set")
	}

	log.Println("Slide MCP Server starting...")
	
	// Start MCP server
	startMCPServer()
}

func startMCPServer() {
	scanner := bufio.NewScanner(os.Stdin)
	
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
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
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
	
	args, ok := params["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	
	var result ToolResult
	
	switch name {
	case "slide_list_devices":
		data, err := listDevices(args)
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
	
	case "slide_list_agents":
		data, err := listAgents(args)
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
				},
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
					},
				},
			},
		},
	}
} 
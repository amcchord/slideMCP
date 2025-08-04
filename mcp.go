package main

import (
	"fmt"
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

// Helper function to create consistent tool results and eliminate repetitive error handling
func createToolResult(data string, err error) ToolResult {
	if err != nil {
		return ToolResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}
	}
	return ToolResult{
		Content: []ToolContent{{Type: "text", Text: data}},
		IsError: false,
	}
}

// sendError creates a standardized error response (legacy compatibility)
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

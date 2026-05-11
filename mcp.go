package main

import (
	"fmt"
)

// ToolInfo is the legacy tool descriptor used by the per-tool getXxxToolInfo()
// helpers. The SDK migration adapts these into mcp.Tool values via
// toolInfoToSDKTool() in registry.go, so each tool file can keep its existing
// rich JSON Schema (with allOf / if / then / required) untouched.
type ToolInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// ToolResult / ToolContent are kept for legacy callers (one-shot --tool mode
// and any in-process callers) that want a JSON-RPC-shaped result rather than
// the SDK's *mcp.CallToolResult.
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// createToolResult preserves the historical helper used by tests / one-shot
// mode. New SDK-backed tool handlers go through adaptToolHandler() instead.
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

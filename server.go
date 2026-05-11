package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// initialContextResourceURI is the canonical URI clients can read for the
// startup overview that older protocol versions used to receive on initialize.
const initialContextResourceURI = "slide://context/clients-devices-agents"

// buildMCPServer constructs the SDK-backed MCP server, registers every tool,
// installs the mode-aware tool filter, and exposes the initial context as a
// readable MCP resource.
func buildMCPServer() (*server.MCPServer, error) {
	srv := server.NewMCPServer(
		ServerName,
		Version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
		server.WithRecovery(),
		server.WithInstructions(serverInstructions()),
		server.WithToolFilter(toolFilterForMode()),
	)

	if err := registerTools(srv); err != nil {
		return nil, fmt.Errorf("register tools: %w", err)
	}

	registerInitialContextResource(srv)

	return srv, nil
}

// runStdioServer wires the MCP server to stdio. Blocks until stdin closes.
func runStdioServer() error {
	srv, err := buildMCPServer()
	if err != nil {
		return err
	}
	log.Printf("%s %s ready on stdio (mode=%s)", ServerName, Version, config.ToolsMode)
	return server.ServeStdio(srv)
}

// registerInitialContextResource exposes a snapshot of clients/devices/agents
// at slide://context/clients-devices-agents. This replaces the non-standard
// initialContext field we used to stuff into the initialize response.
func registerInitialContextResource(srv *server.MCPServer) {
	resource := mcp.NewResource(
		initialContextResourceURI,
		"Slide initial context",
		mcp.WithResourceDescription(
			"Hierarchical overview of every client, device, and agent in the account. "+
				"Equivalent to calling list_all_clients_devices_and_agents; cached at "+
				"startup for faster first-turn answers.",
		),
		mcp.WithMIMEType("application/json"),
	)

	srv.AddResource(resource, func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		payload := buildInitialContextPayload()
		body, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal initial context: %w", err)
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      initialContextResourceURI,
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}

// buildInitialContextPayload returns the structured payload used both by the
// initial-context resource and by any caller that wants the same overview.
func buildInitialContextPayload() map[string]interface{} {
	return map[string]interface{}{
		"clients_devices_agents": fetchInitialContext(),
		"_metadata": map[string]interface{}{
			"description":   "Initial overview of all clients, devices, and agents loaded on demand for improved first-turn performance.",
			"source_tool":   "list_all_clients_devices_and_agents",
			"usage_note":    "Re-read this resource (or call list_all_clients_devices_and_agents) for live data.",
			"timestamp":     fmt.Sprintf("%d", time.Now().Unix()),
			"cache_age_sec": 0,
			"tools_mode":    config.ToolsMode,
			"api_base_url":  config.BaseURL,
		},
	}
}

// fetchInitialContext gets the same structured overview the
// list_all_clients_devices_and_agents tool returns.
func fetchInitialContext() map[string]interface{} {
	contextData, err := listAllClientsDevicesAndAgents(map[string]interface{}{})
	if err != nil {
		log.Printf("Warning: Failed to fetch initial context data: %v", err)
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to fetch initial context: %v", err),
			"note":  "Initial context will be available via the list_all_clients_devices_and_agents tool",
		}
	}

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

// serverInstructions is shown to clients in the initialize response and is a
// good spot for usage tips that benefit any LLM driving the server.
func serverInstructions() string {
	return "Slide MCP server (" + Version + "). " +
		"Each meta-tool takes an `operation` parameter; see the tool description for the supported set. " +
		"Read the resource " + initialContextResourceURI + " for a hierarchical overview of every client, device, and agent. " +
		"Mutations (create/update/delete/poweroff/reboot) are gated by the active tools mode (" + ToolsReporting + "/" + ToolsRestores + "/" + ToolsFullSafe + "/" + ToolsFull + ")."
}

// runOneShotTool runs a single tool by name with the supplied JSON arguments
// and writes the result to stdout. Used by the --tool / --args CLI mode and by
// scripts/show_mcp_context.sh-style integrations.
func runOneShotTool(name string, args map[string]interface{}) error {
	if args == nil {
		args = map[string]interface{}{}
	}
	if config.IsToolDisabled(name) {
		return fmt.Errorf("tool '%s' is disabled", name)
	}
	if !config.IsToolAllowed(name) {
		return fmt.Errorf("tool '%s' not available in '%s' mode", name, config.ToolsMode)
	}
	handler, ok := toolRegistry[name]
	if !ok {
		return fmt.Errorf("unknown tool: %s", name)
	}
	result := createToolResult(handler(args))
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	fmt.Println(string(out))
	return nil
}

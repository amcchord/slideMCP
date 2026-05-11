package main

// MCP Resources - cheap, read-only context Claude Desktop can pull at the
// start of any conversation without burning a tool call. v4.0.0 expands
// from one resource (slide://context/...) to nine, including URI-templated
// per-client / per-device / per-agent views.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Resource URIs the server advertises. The legacy slide://context URI
// keeps working as an alias for slide://overview/inventory so existing
// Claude Desktop installs that cached the old context resource don't break.
const (
	resourceURIInventory       = "slide://overview/inventory"
	resourceURIHealth          = "slide://overview/health"
	resourceURIAlertsOpen      = "slide://alerts/unresolved"
	resourceURIAuditRecent     = "slide://audit/recent"
	resourceURIDocsOpenAPI     = "slide://docs/openapi"
	resourceURITplClient       = "slide://client/{client_id}"
	resourceURITplDevice       = "slide://device/{device_id}"
	resourceURITplAgent        = "slide://agent/{agent_id}"
	resourceURITplAgentRecents = "slide://agent/{agent_id}/snapshots/recent"
	resourceURILegacyContext   = "slide://context/clients-devices-agents"
)

// registerResources wires every v4 resource onto the SDK server.
func registerResources(s *server.MCPServer) {
	addStaticResource(s, resourceURIInventory, "Slide overview - full inventory",
		"Hierarchical view of every client, device, and agent (clients -> devices -> agents).",
		handleResourceInventory)

	addStaticResource(s, resourceURIHealth, "Slide overview - health summary",
		"One-line-per-device-and-agent health flags using the default 30-minute stale cutoff.",
		handleResourceHealth)

	addStaticResource(s, resourceURIAlertsOpen, "Slide alerts - unresolved",
		"All currently unresolved alerts, prioritised by severity hint.",
		handleResourceAlertsOpen)

	addStaticResource(s, resourceURIAuditRecent, "Slide audit log - last 24h",
		"Most recent audit-log entries (last 24 hours, summary view).",
		handleResourceAuditRecent)

	addStaticResource(s, resourceURIDocsOpenAPI, "Slide API - OpenAPI spec",
		"Live OpenAPI spec for the Slide API (cached for 1 hour).",
		handleResourceOpenAPI)

	// Backward-compat alias for v3 installs.
	addStaticResource(s, resourceURILegacyContext, "Slide initial context (legacy)",
		"Backward-compat alias for slide://overview/inventory.",
		handleResourceInventory)

	addTemplate(s, resourceURITplClient, "Slide client (by ID)",
		"Single client + its devices + open alerts. URI: slide://client/{client_id}.",
		handleResourceClient)
	addTemplate(s, resourceURITplDevice, "Slide device (by ID)",
		"Single device + its agents + open alerts. URI: slide://device/{device_id}.",
		handleResourceDevice)
	addTemplate(s, resourceURITplAgent, "Slide agent (by ID)",
		"Single agent + recent snapshots + open alerts. URI: slide://agent/{agent_id}.",
		handleResourceAgent)
	addTemplate(s, resourceURITplAgentRecents, "Slide agent recent snapshots",
		"Last 14 days of snapshots for one agent. URI: slide://agent/{agent_id}/snapshots/recent.",
		handleResourceAgentRecentSnapshots)
}

// addStaticResource is a tiny wrapper for non-templated URIs.
func addStaticResource(s *server.MCPServer, uri, name, desc string, fn func(string) ([]byte, error)) {
	r := mcp.NewResource(uri, name,
		mcp.WithResourceDescription(desc),
		mcp.WithMIMEType("application/json"),
	)
	s.AddResource(r, func(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		body, err := fn(req.Params.URI)
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}

// addTemplate is a wrapper for URI-templated resources.
func addTemplate(s *server.MCPServer, tpl, name, desc string, fn func(string) ([]byte, error)) {
	rt := mcp.NewResourceTemplate(tpl, name,
		mcp.WithTemplateDescription(desc),
		mcp.WithTemplateMIMEType("application/json"),
	)
	s.AddResourceTemplate(rt, func(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		body, err := fn(req.Params.URI)
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}

// --- Handlers ---------------------------------------------------------

func handleResourceInventory(_ string) ([]byte, error) {
	body, err := listAllClientsDevicesAndAgents(map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

func handleResourceHealth(_ string) ([]byte, error) {
	body, err := handleOverviewHealth(map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

func handleResourceAlertsOpen(_ string) ([]byte, error) {
	body, err := handleAlertsTriage(map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

func handleResourceAuditRecent(_ string) ([]byte, error) {
	body, err := handleAuditRecent(map[string]interface{}{"hours": float64(24)})
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

// OpenAPI cached fetch ---------------------------------------------------

var openapiCache = struct {
	sync.Mutex
	body    []byte
	fetched time.Time
}{}

const openapiCacheTTL = time.Hour

func handleResourceOpenAPI(_ string) ([]byte, error) {
	openapiCache.Lock()
	defer openapiCache.Unlock()
	if openapiCache.body != nil && time.Since(openapiCache.fetched) < openapiCacheTTL {
		return openapiCache.body, nil
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get("https://api.slide.tech/openapi.json")
	if err != nil {
		if openapiCache.body != nil {
			return openapiCache.body, nil
		}
		return nil, fmt.Errorf("failed to fetch openapi: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openapi fetch returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	openapiCache.body = body
	openapiCache.fetched = time.Now()
	return body, nil
}

// Templated handlers ----------------------------------------------------

// extractIDFromURI pulls the trailing ID out of a slide://X/{id}[/...] URI.
func extractIDFromURI(uri, prefix string) (string, error) {
	if !strings.HasPrefix(uri, prefix) {
		return "", fmt.Errorf("URI %q does not match prefix %q", uri, prefix)
	}
	rest := strings.TrimPrefix(uri, prefix)
	if rest == "" {
		return "", fmt.Errorf("missing ID in URI %q", uri)
	}
	// Strip trailing path segments (e.g. /snapshots/recent).
	if idx := strings.Index(rest, "/"); idx > 0 {
		rest = rest[:idx]
	}
	return rest, nil
}

func handleResourceClient(uri string) ([]byte, error) {
	id, err := extractIDFromURI(uri, "slide://client/")
	if err != nil {
		return nil, err
	}
	body, err := handleOverviewForClient(map[string]interface{}{"client_id": id})
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

func handleResourceDevice(uri string) ([]byte, error) {
	id, err := extractIDFromURI(uri, "slide://device/")
	if err != nil {
		return nil, err
	}
	body, err := handleOverviewForDevice(map[string]interface{}{"device_id": id})
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

func handleResourceAgent(uri string) ([]byte, error) {
	id, err := extractIDFromURI(uri, "slide://agent/")
	if err != nil {
		return nil, err
	}
	// Compose: agent detail + recent snapshots + open alerts.
	agentData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/agent/%s", id), nil)
	if err != nil {
		return nil, err
	}
	var agent Agent
	if err := json.Unmarshal(agentData, &agent); err != nil {
		return nil, err
	}
	snapshotsData, _ := makeAPIRequest("GET", fmt.Sprintf("/v1/snapshot?agent_id=%s&limit=20&sort_by=backup_start_time&sort_asc=false", id), nil)
	var snaps PaginatedResponse[Snapshot]
	_ = json.Unmarshal(snapshotsData, &snaps)
	alertsData, _ := makeAPIRequest("GET", fmt.Sprintf("/v1/alert?agent_id=%s&resolved=false&limit=50", id), nil)
	var alerts PaginatedResponse[Alert]
	_ = json.Unmarshal(alertsData, &alerts)
	out := map[string]interface{}{
		"agent":            agent,
		"recent_snapshots": snaps.Data,
		"open_alerts":      alerts.Data,
	}
	return json.Marshal(out)
}

func handleResourceAgentRecentSnapshots(uri string) ([]byte, error) {
	id, err := extractIDFromURI(uri, "slide://agent/")
	if err != nil {
		return nil, err
	}
	body, err := handleSnapshotsRecentForAgent(map[string]interface{}{"agent_id": id})
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

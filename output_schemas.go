package main

// outputSchemaForTool returns a JSON-Schema describing the most common
// envelope shape for each tool. We deliberately keep these permissive
// (additionalProperties:true) because each meta-tool's `operation`
// dispatches to many different response shapes. The schema is
// informational - it signals "structured object response" to clients
// that support outputSchema rendering, without rejecting valid payloads.

import "encoding/json"

const sharedListEnvelopeSchema = `{
  "type": "object",
  "additionalProperties": true,
  "properties": {
    "data":       {"description": "Items in the current page (may be summary or detailed depending on the requested format).", "type": "array"},
    "pagination": {"type": "object", "properties": {"next_offset": {"type": "integer"}, "total": {"type": "integer"}}},
    "count":      {"type": "integer"}
  }
}`

const sharedSingleObjectSchema = `{
  "type": "object",
  "additionalProperties": true
}`

const sharedRollupSchema = `{
  "type": "object",
  "additionalProperties": true,
  "properties": {
    "summary":      {"type": "object"},
    "entries":      {"type": "array"},
    "agents":       {"type": "array"},
    "agents_status":{"type": "array"},
    "snapshots":    {"type": "array"},
    "alerts":       {"type": "array"}
  }
}`

func outputSchemaForTool(name string) json.RawMessage {
	switch name {
	case "slide_overview":
		return json.RawMessage(sharedRollupSchema)
	case "slide_files", "slide_audit", "slide_clients", "slide_admin",
		"slide_devices", "slide_agents", "slide_snapshots":
		// Mostly list+single hybrid surfaces; envelope shape is closest.
		return json.RawMessage(sharedListEnvelopeSchema)
	case "slide_recovery", "slide_alerts", "slide_backups":
		// Mix of lists and rollups.
		return json.RawMessage(sharedRollupSchema)
	case "list_all_clients_devices_and_agents":
		return json.RawMessage(sharedSingleObjectSchema)
	}
	return nil
}

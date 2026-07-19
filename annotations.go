package main

// Tool-level MCP annotations. The MCP spec (2025-06-18+) lets each tool
// advertise hints about its behaviour so well-behaved hosts can decide
// whether to ask the user for confirmation:
//
//	readOnlyHint:    Pure read; never mutates state. Safe to call freely.
//	destructiveHint: Mutation that's irreversible (deletes, power-off, etc).
//	idempotentHint:  Repeat calls with the same args produce the same result.
//	openWorldHint:   The result depends on external state (true for nearly
//	                 every tool here since they hit the live Slide API).
//
// We classify per-tool because all our meta-tools mix read and write
// operations behind a single `operation` enum. The classification leans
// conservative: if the tool *can* mutate, we drop the readOnly hint.

import "github.com/mark3labs/mcp-go/mcp"

func annotationsForTool(name string) mcp.ToolAnnotation {
	openWorld := true
	switch name {
	case "slide_help":
		// Pure read, fully local content (markdown baked into the binary)
		// for most operations. Never touches external state, so openWorld=false.
		ro := true
		idempotent := true
		notOpenWorld := false
		return mcp.ToolAnnotation{
			Title:          humanTitle(name),
			ReadOnlyHint:   &ro,
			IdempotentHint: &idempotent,
			OpenWorldHint:  &notOpenWorld,
		}
	case "slide_overview", "slide_audit", "list_all_clients_devices_and_agents":
		// Pure read tools.
		ro := true
		idempotent := true
		return mcp.ToolAnnotation{
			Title:          humanTitle(name),
			ReadOnlyHint:   &ro,
			IdempotentHint: &idempotent,
			OpenWorldHint:  &openWorld,
		}
	case "slide_snapshots", "slide_alerts", "slide_clients", "slide_admin":
		// Mostly read, with light mutation.
		ro := false
		destructive := false
		idempotent := true
		return mcp.ToolAnnotation{
			Title:           humanTitle(name),
			ReadOnlyHint:    &ro,
			DestructiveHint: &destructive,
			IdempotentHint:  &idempotent,
			OpenWorldHint:   &openWorld,
		}
	case "slide_files":
		// Reads + restore-creation + push-back. Push-back is non-trivial
		// (writes to a protected system), so destructiveHint=true.
		ro := false
		destructive := true
		return mcp.ToolAnnotation{
			Title:           humanTitle(name),
			ReadOnlyHint:    &ro,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		}
	case "slide_recovery":
		// VM boot + image export + DR network changes. Pretty much
		// every operation mutates external state.
		ro := false
		destructive := true
		return mcp.ToolAnnotation{
			Title:           humanTitle(name),
			ReadOnlyHint:    &ro,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		}
	case "slide_devices":
		// Includes poweroff/reboot in `full` mode.
		ro := false
		destructive := true
		return mcp.ToolAnnotation{
			Title:           humanTitle(name),
			ReadOnlyHint:    &ro,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		}
	case "slide_agents":
		// Includes delete_passphrase. Because annotations are tool-level and
		// this is an operation-oriented meta-tool, advertise the conservative
		// destructive hint even though most agent operations are reversible.
		ro := false
		destructive := true
		return mcp.ToolAnnotation{
			Title:           humanTitle(name),
			ReadOnlyHint:    &ro,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		}
	case "slide_backups":
		// Settings updates + backup-start. Reversible-ish.
		ro := false
		destructive := false
		return mcp.ToolAnnotation{
			Title:           humanTitle(name),
			ReadOnlyHint:    &ro,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		}
	}
	// Default conservative.
	ro := false
	destructive := false
	return mcp.ToolAnnotation{
		Title:           humanTitle(name),
		ReadOnlyHint:    &ro,
		DestructiveHint: &destructive,
		OpenWorldHint:   &openWorld,
	}
}

// humanTitle returns a friendly display title for a tool name.
func humanTitle(name string) string {
	switch name {
	case "slide_help":
		return "Slide Help"
	case "slide_overview":
		return "Slide Overview"
	case "slide_files":
		return "Slide Files"
	case "slide_recovery":
		return "Slide Recovery"
	case "slide_audit":
		return "Slide Audit Log"
	case "slide_clients":
		return "Slide Clients"
	case "slide_admin":
		return "Slide Admin"
	case "slide_devices":
		return "Slide Devices"
	case "slide_agents":
		return "Slide Agents"
	case "slide_snapshots":
		return "Slide Snapshots"
	case "slide_backups":
		return "Slide Backups"
	case "slide_alerts":
		return "Slide Alerts"
	case "list_all_clients_devices_and_agents":
		return "List Clients/Devices/Agents (legacy alias)"
	}
	return name
}

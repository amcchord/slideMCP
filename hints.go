package main

// Response-level affordances - next_steps hints + _resolved block.
//
// Both fields are appended to a successful tool response as part of
// formatList / formatSingle (see format.go). They piggyback on the
// existing args map so individual operation handlers never need to be
// touched.
//
// next_steps:
//   A short, opinionated list of follow-up tool calls the LLM should
//   consider after a given (tool, operation). Curated, NOT auto-generated.
//   Suppressed when args["hints"]="off" (or the operator passes hints=off
//   in the tool args).
//
// _resolved:
//   Populated by name_resolver.resolveNameHint when a name_hint matched
//   exactly one candidate. Surfaces the resolution so the LLM can
//   confirm to the user which agent/device/client it actually used.

import (
	"encoding/json"
	"strings"
)

// extractHintsOff returns true when the caller asked us NOT to attach
// next_steps to this response. Default is hints-on.
func extractHintsOff(args map[string]interface{}) bool {
	raw, ok := args["hints"].(string)
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "off", "no", "false", "0":
		return true
	}
	return false
}

// nextStepsFor returns a curated list of suggested follow-up tool
// calls for (tool, operation), drawing on the current args so the
// suggestions reference real IDs the caller already resolved.
//
// Keep this list short (1-3 entries) and concrete. Empty result means
// "no useful suggestion" - the caller will skip the next_steps field.
func nextStepsFor(tool, op string, args map[string]interface{}) []string {
	get := func(k string) string {
		v, _ := args[k].(string)
		return v
	}
	switch tool {
	case "slide_overview":
		switch op {
		case "health":
			return []string{
				"Anything 'stale' or 'unknown'? Call slide_overview operation=for_device name_hint=<that device> for a deeper view.",
				"Call slide_alerts operation=triage to see if any unresolved alerts correlate with the stale entries.",
			}
		case "inventory":
			return []string{
				"Call slide_overview operation=health for a one-line-per-device-and-agent health summary.",
				"Or slide_alerts operation=triage if the user is asking 'is anything broken?'.",
			}
		case "for_client":
			return []string{
				"Call slide_backups operation=status_for_client client_id=" + get("client_id") + " for backup status across this client.",
				"Or slide_alerts operation=triage for prioritised unresolved alerts.",
			}
		case "for_device":
			return []string{
				"Call slide_backups operation=status_for_device device_id=" + get("device_id") + " for per-agent backup status on this device.",
				"slide_snapshots operation=list device_id=" + get("device_id") + " for recent snapshots.",
			}
		}
	case "slide_files":
		switch op {
		case "search":
			return []string{
				"Pick a path from the results and call slide_files operation=versions agent_id=" + get("agent_id") + " path=<chosen path> to list every snapshot that contains it.",
			}
		case "versions":
			return []string{
				"Pick a snapshot_id and call slide_files operation=create_restore snapshot_id=<id> device_id=<id> to start a restore session.",
			}
		case "create_restore":
			return []string{
				"Call slide_files operation=browse file_restore_id=<id> browse_path=/ to walk the restored filesystem.",
				"Or slide_files operation=create_push file_restore_id=<id> source_file_path=<path> destination_folder=C:\\SlideRestore to push a file back to the protected system.",
			}
		case "create_push":
			return []string{
				"Call slide_files operation=get_push_status file_restore_id=<id> file_restore_push_id=<id> to monitor the push.",
			}
		case "browse":
			return []string{
				"Once you pick a file, call slide_files operation=create_push to push it back to the protected system.",
			}
		}
	case "slide_recovery":
		switch op {
		case "boot_vm":
			return []string{
				"Call slide_recovery operation=get_vm virt_id=<from response> to confirm the VM is running.",
				"slide_recovery operation=get_rdp_bookmark virt_id=<id> for an RDP file to access the VM.",
			}
		case "export_image":
			return []string{
				"Call slide_recovery operation=get_image image_export_id=<from response> to monitor progress.",
				"slide_recovery operation=browse_image image_export_id=<id> once the export completes.",
			}
		case "create_network":
			return []string{
				"Add a port forward with slide_recovery operation=create_port_forward network_id=<from response>.",
				"Or add a WireGuard peer with slide_recovery operation=create_wg_peer network_id=<id>.",
			}
		}
	case "slide_alerts":
		switch op {
		case "triage":
			return []string{
				"For each critical alert, call slide_overview operation=for_device device_id=<the device> for context.",
				"Mark an alert resolved with slide_alerts operation=update alert_id=<id> resolved=true (requires safe or full mode).",
			}
		}
	case "slide_backups":
		switch op {
		case "status_for_client", "status_for_device":
			return []string{
				"For any agent with failures, call slide_backups operation=recent_for_agent agent_id=<id> hours=24 to inspect specific runs.",
			}
		case "start":
			return []string{
				"Backup launched. Call slide_backups operation=recent_for_agent agent_id=" + get("agent_id") + " hours=2 to monitor.",
			}
		}
	case "slide_snapshots":
		switch op {
		case "recent_for_agent":
			return []string{
				"Pick a snapshot_id and call slide_recovery operation=boot_vm to spin up a recovery VM.",
				"Or slide_files operation=versions agent_id=" + get("agent_id") + " path=<path> if the user wants a specific file.",
			}
		}
	case "slide_audit":
		switch op {
		case "recent":
			return []string{
				"Filter to a specific action with slide_audit operation=list audit_action_name=<action> (see slide_audit operation=actions for the valid set).",
			}
		}
	case "slide_help":
		switch op {
		case "getting_started":
			return []string{
				"Try slide_help operation=examples for copy-pasteable example questions.",
				"Or slide_help operation=what_can_you_do for a capability summary scoped to the active permission tier.",
			}
		}
	}
	return nil
}

// augmentJSONResponse parses the JSON response, splices in `_resolved`
// and `next_steps` when present in args, and re-marshals. If the body
// isn't a JSON object (rare), wraps it in `{"result": ...}` so we can
// still attach the affordances. Returns the original body on parse
// failure - hints must never break a real response.
func augmentJSONResponse(body string, args map[string]interface{}) string {
	if args == nil {
		return body
	}

	var resolution interface{}
	if v, ok := args["_resolution"]; ok {
		resolution = v
	}

	var hints []string
	if !extractHintsOff(args) {
		tool, _ := args["_tool"].(string)
		op, _ := args["operation"].(string)
		hints = nextStepsFor(tool, op, args)
	}

	if resolution == nil && len(hints) == 0 {
		return body
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return body
	}
	m, ok := parsed.(map[string]interface{})
	if !ok {
		m = map[string]interface{}{"result": parsed}
	}
	if resolution != nil {
		m["_resolved"] = resolution
	}
	if len(hints) > 0 {
		m["next_steps"] = hints
	}
	out, err := json.Marshal(m)
	if err != nil {
		return body
	}
	return string(out)
}

package main

// handleAgentsTool handles all agent-related operations through a single meta-tool
func handleAgentsTool(args map[string]interface{}) (string, error) {
	agentRes := ResolutionSpec{IDKey: "agent_id", Kind: "agent"}
	deviceRes := ResolutionSpec{IDKey: "device_id", Kind: "device"}
	return HandleToolWithOperations(CreateToolConfigWithResolutions("slide_agents", ToolOperations{
		// Core CRUD (kept stable from v2.x)
		"list":              listAgents,
		"get":               getAgent,
		"create":            createAgent,
		"pair":              pairAgent,
		"update":            updateAgent,
		"add_passphrase":    addAgentPassphrase,
		"delete_passphrase": deleteAgentPassphrase,

		// Slide API v1.27.0 additions
		"list_services":          handleAgentListServices,
		"update_services":        handleAgentUpdateServices,
		"set_schedule":           handleAgentSetSchedule,
		"clear_schedule":         handleAgentClearSchedule,
		"pause_backups":          handleAgentPauseBackups,
		"resume_backups":         handleAgentResumeBackups,
		"set_retention":          handleAgentSetRetention,
		"set_restore_defaults":   handleAgentSetRestoreDefaults,
		"set_volumes":            handleAgentSetVolumes,
		"set_file_index_enabled": handleAgentSetFileIndex,
		"set_timezone":           handleAgentSetTimezone,
		"set_comments":           handleAgentSetComments,
		"update_alert_config":    handleAgentUpdateAlertConfig,
	}, map[string]ResolutionSpec{
		// Operations that take agent_id resolve an agent.
		"get":                    agentRes,
		"update":                 agentRes,
		"add_passphrase":         agentRes,
		"delete_passphrase":      agentRes,
		"list_services":          agentRes,
		"update_services":        agentRes,
		"set_schedule":           agentRes,
		"clear_schedule":         agentRes,
		"pause_backups":          agentRes,
		"resume_backups":         agentRes,
		"set_retention":          agentRes,
		"set_restore_defaults":   agentRes,
		"set_volumes":            agentRes,
		"set_file_index_enabled": agentRes,
		"set_timezone":           agentRes,
		"set_comments":           agentRes,
		"update_alert_config":    agentRes,
		// create / pair take device_id (the agent doesn't exist yet).
		"create": deviceRes,
		"pair":   deviceRes,
	}), args)
}

// agentOperationEnums is the canonical list of operations exposed by
// slide_agents. Keep this in lockstep with the dispatch map above and the
// allOf conditional requirements in the schema below.
var agentOperationEnums = []string{
	"list", "get", "create", "pair", "update", "add_passphrase", "delete_passphrase",
	"list_services", "update_services",
	"set_schedule", "clear_schedule",
	"pause_backups", "resume_backups",
	"set_retention", "set_restore_defaults",
	"set_volumes", "set_file_index_enabled",
	"set_timezone", "set_comments",
	"update_alert_config",
}

// getAgentsToolInfo returns the tool definition for the agents meta-tool
func getAgentsToolInfo() ToolInfo {
	return ToolInfo{
		Name: "slide_agents",
		Description: "Slide MCP - manage backup agents installed on protected systems (servers, endpoints, laptops). " +
			"REACH FOR THIS whenever the user mentions a Slide agent, 'the agent on <server>', 'pause backups on X', " +
			"'resume backups', 'backup schedule', 'retention policy', 'pair a new agent', 'Bob's laptop', " +
			"'protected system' settings, agent-level alert configuration, VSS writers, or Windows service verification. " +
			"Operations: list, get, create, pair, update, add_passphrase, delete_passphrase, " +
			"plus v1.27.0 additions: list_services, update_services (Windows service verification), " +
			"set_schedule / clear_schedule, pause_backups / resume_backups, set_retention, set_restore_defaults, " +
			"set_volumes, set_file_index_enabled, set_timezone, set_comments, update_alert_config (per-agent alert pause/resume). " +
			"Single-agent operations accept agent_id OR name_hint (resolves a hostname or display name to the agent_id server-side).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        agentOperationEnums,
				},
				// list
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Number of results per page (max 50) - used with 'list' operation",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Pagination offset - used with 'list' operation",
				},
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by device ID - used with 'list' operation, or required for 'create' and 'pair' operations (alternative for create/pair: pass `name_hint`)",
				},
				"client_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by client ID - used with 'list' operation",
				},
				"name_hint": map[string]interface{}{
					"type":        "string",
					"description": "Alternative to agent_id (or device_id for create/pair): a hostname or display name (case-insensitive substring match). Resolves to an agent for single-agent operations, to a device for `create` and `pair`.",
				},
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with 'list' operation",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with 'list' operation",
					"enum":        []string{"id", "hostname", "name"},
				},
				// get / update / pair / passphrase / etc.
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the agent - required for 'get', 'update', and every v1.27.0 operation that targets a single agent (alternative: pass `name_hint`)",
				},
				"display_name": map[string]interface{}{
					"type":        "string",
					"description": "Display name for the agent - required for 'create' and 'update' operations",
				},
				"pair_code": map[string]interface{}{
					"type":        "string",
					"description": "Pair code generated during agent creation - required for 'pair' operation",
				},
				"passphrase_name": map[string]interface{}{
					"type":        "string",
					"description": "Friendly name for the passphrase - required for 'add_passphrase' operation",
				},
				"passphrase": map[string]interface{}{
					"type":        "string",
					"description": "The passphrase to add - required for 'add_passphrase' operation, or current passphrase for 'delete_passphrase' operation",
				},
				"agent_passphrase_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the passphrase to delete - required for 'delete_passphrase' operation",
				},
				"vss_writer_configs": map[string]interface{}{
					"type":        "array",
					"description": "VSS writer configurations - used with 'update' operation",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"writer_id": map[string]interface{}{"type": "string"},
							"excluded":  map[string]interface{}{"type": "boolean"},
						},
						"required": []string{"writer_id", "excluded"},
					},
				},
				"sealed": map[string]interface{}{
					"type":        "boolean",
					"description": "Set to false to unseal an agent with a user-managed passphrase - used with 'update' operation",
				},

				// v1.27.0: services
				"services": map[string]interface{}{
					"type":        "array",
					"description": "Services to update - required for 'update_services'. Each entry is {service_id, verify_on_boot}.",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"service_id":     map[string]interface{}{"type": "string"},
							"verify_on_boot": map[string]interface{}{"type": "boolean"},
						},
						"required": []string{"service_id", "verify_on_boot"},
					},
				},

				// v1.27.0: backup schedule
				"interval_in_minutes": map[string]interface{}{
					"type":        "number",
					"description": "Backup interval - required for 'set_schedule'. One of 30, 60, 120, 180, 360.",
					"enum":        []int{30, 60, 120, 180, 360},
				},
				"start_hour": map[string]interface{}{
					"type":        "number",
					"description": "Schedule start hour 0-23 - required for 'set_schedule'.",
					"minimum":     0,
					"maximum":     23,
				},
				"end_hour": map[string]interface{}{
					"type":        "number",
					"description": "Schedule end hour 0-23 - required for 'set_schedule'.",
					"minimum":     0,
					"maximum":     23,
				},
				"days": map[string]interface{}{
					"type":        "array",
					"description": "Days of week (0=Sun..6=Sat) - required for 'set_schedule'.",
					"items": map[string]interface{}{
						"type":    "number",
						"minimum": 0,
						"maximum": 6,
					},
					"minItems": 1,
				},

				// v1.27.0: pause / resume
				"indefinite": map[string]interface{}{
					"type":        "boolean",
					"description": "Pause indefinitely - used with 'pause_backups' or 'update_alert_config'.",
				},
				"paused_until": map[string]interface{}{
					"type":        "string",
					"description": "RFC3339 timestamp - used with 'pause_backups' to pause backups until that time.",
				},

				// v1.27.0: retention
				"retention_policy_name": map[string]interface{}{
					"type":        "string",
					"description": "Retention policy preset - required for 'set_retention'.",
					"enum":        []string{"lean", "balanced", "comprehensive"},
				},
				"retention_policy_max_age_months": map[string]interface{}{
					"type":        "number",
					"description": "Max snapshot age in months - required for 'set_retention'.",
					"enum":        []int{3, 6, 12, 24, 36, 84},
				},

				// v1.27.0: default restore settings
				"cpu_count": map[string]interface{}{
					"type":        "number",
					"description": "Default CPU count for restores - used with 'set_restore_defaults'.",
					"enum":        []int{1, 2, 4, 8, 16},
				},
				"memory_mb": map[string]interface{}{
					"type":        "number",
					"description": "Default memory in MB for restores - used with 'set_restore_defaults'.",
				},
				"disk_bus": map[string]interface{}{
					"type":        "string",
					"description": "Default disk bus for restores - used with 'set_restore_defaults'.",
					"enum":        []string{"sata", "virtio"},
				},
				"network_model": map[string]interface{}{
					"type":        "string",
					"description": "Default network model for restores - used with 'set_restore_defaults'.",
					"enum":        []string{"virtio", "hypervisor_default", "e1000", "rtl8139"},
				},

				// v1.27.0: volumes
				"volumes": map[string]interface{}{
					"type":        "array",
					"description": "Volumes to update - used with 'set_volumes'. Each entry is {volume_id, include, mount_points?}.",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"volume_id":    map[string]interface{}{"type": "string"},
							"include":      map[string]interface{}{"type": "boolean"},
							"mount_points": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						},
						"required": []string{"volume_id", "include"},
					},
				},
				"volumes_include_default": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether new volumes should be included by default - used with 'set_volumes'.",
				},

				// v1.27.0: misc
				"file_index_enabled": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to enable file indexing - required for 'set_file_index_enabled'.",
				},
				"timezone": map[string]interface{}{
					"type":        "string",
					"description": "IANA timezone string (e.g. 'America/New_York') - required for 'set_timezone'.",
				},
				"comments": map[string]interface{}{
					"type":        "string",
					"description": "Free-form comments - required for 'set_comments'.",
				},

				// v1.27.0: alert configs
				"alert_type": map[string]interface{}{
					"type":        "string",
					"description": "Alert type to update - required for 'update_alert_config'.",
					"enum": []string{
						"device_not_checking_in", "device_out_of_date",
						"device_storage_not_healthy", "device_storage_space_low", "device_storage_space_critical",
						"agent_not_checking_in", "agent_not_backing_up", "agent_backup_failed",
					},
				},
				"resume": map[string]interface{}{
					"type":        "boolean",
					"description": "Set true to resume a paused alert - used with 'update_alert_config'.",
				},
				"pause_for_minutes": map[string]interface{}{
					"type":        "number",
					"description": "Pause this alert for N minutes - used with 'update_alert_config'.",
					"enum":        []int{0, 30, 60, 120, 240, 480, 1440, 10080},
				},
			},
			"required": []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get"), "then": reqEither("agent_id", "name_hint")},
				{"if": ifOp("create"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						req("display_name"),
						reqEither("device_id", "name_hint"),
					},
				}},
				{"if": ifOp("pair"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						req("pair_code"),
						reqEither("device_id", "name_hint"),
					},
				}},
				{"if": ifOp("update"), "then": reqEither("agent_id", "name_hint")},
				{"if": ifOp("add_passphrase"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						reqEither("agent_id", "name_hint"),
						req("passphrase_name", "passphrase"),
					},
				}},
				{"if": ifOp("delete_passphrase"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						reqEither("agent_id", "name_hint"),
						req("agent_passphrase_id", "passphrase"),
					},
				}},

				// v1.27.0
				{"if": ifOp("list_services"), "then": reqEither("agent_id", "name_hint")},
				{"if": ifOp("update_services"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						reqEither("agent_id", "name_hint"),
						req("services"),
					},
				}},
				{"if": ifOp("set_schedule"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						reqEither("agent_id", "name_hint"),
						req("interval_in_minutes", "start_hour", "end_hour", "days"),
					},
				}},
				{"if": ifOp("clear_schedule"), "then": reqEither("agent_id", "name_hint")},
				{"if": ifOp("pause_backups"), "then": reqEither("agent_id", "name_hint")},
				{"if": ifOp("resume_backups"), "then": reqEither("agent_id", "name_hint")},
				{"if": ifOp("set_retention"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						reqEither("agent_id", "name_hint"),
						req("retention_policy_name", "retention_policy_max_age_months"),
					},
				}},
				{"if": ifOp("set_restore_defaults"), "then": reqEither("agent_id", "name_hint")},
				{"if": ifOp("set_volumes"), "then": reqEither("agent_id", "name_hint")},
				{"if": ifOp("set_file_index_enabled"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						reqEither("agent_id", "name_hint"),
						req("file_index_enabled"),
					},
				}},
				{"if": ifOp("set_timezone"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						reqEither("agent_id", "name_hint"),
						req("timezone"),
					},
				}},
				{"if": ifOp("set_comments"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						reqEither("agent_id", "name_hint"),
						req("comments"),
					},
				}},
				{"if": ifOp("update_alert_config"), "then": map[string]interface{}{
					"allOf": []map[string]interface{}{
						reqEither("agent_id", "name_hint"),
						req("alert_type"),
					},
				}},
			},
		},
	}
}

// ifOp / req are tiny helpers that keep the conditional schema readable.
func ifOp(op string) map[string]interface{} {
	return map[string]interface{}{
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{"const": op},
		},
	}
}

func req(fields ...string) map[string]interface{} {
	return map[string]interface{}{"required": fields}
}

// reqEither expresses "this operation needs at least one of the listed
// fields" via JSON Schema `anyOf`. Used to allow either `*_id` OR
// `name_hint` to satisfy a parameter requirement.
func reqEither(fields ...string) map[string]interface{} {
	options := make([]map[string]interface{}, 0, len(fields))
	for _, f := range fields {
		options = append(options, map[string]interface{}{"required": []string{f}})
	}
	return map[string]interface{}{"anyOf": options}
}

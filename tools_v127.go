package main

// Operation handlers for surfaces added in Slide API v1.27.0.
// These are referenced from tools_agents.go, tools_devices.go,
// tools_snapshots.go, and tools_user_management.go.

import (
	"encoding/json"
	"fmt"
)

// helpers ---------------------------------------------------------------------

func requireString(args map[string]interface{}, key string) (string, error) {
	v, ok := args[key].(string)
	if !ok || v == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return v, nil
}

func optionalBool(args map[string]interface{}, key string) (bool, bool) {
	if v, ok := args[key].(bool); ok {
		return v, true
	}
	return false, false
}

func optionalString(args map[string]interface{}, key string) (string, bool) {
	if v, ok := args[key].(string); ok {
		return v, true
	}
	return "", false
}

func optionalInt(args map[string]interface{}, key string) (int, bool) {
	switch v := args[key].(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	case json.Number:
		n, err := v.Int64()
		if err == nil {
			return int(n), true
		}
	}
	return 0, false
}

func toJSONString(v interface{}) (string, error) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}

// agent: services -------------------------------------------------------------

func handleAgentListServices(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	services, err := listAgentServices(agentID)
	if err != nil {
		return "", err
	}
	return toJSONString(map[string]interface{}{
		"agent_id": agentID,
		"services": services,
	})
}

func handleAgentUpdateServices(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	raw, ok := args["services"].([]interface{})
	if !ok || len(raw) == 0 {
		return "", fmt.Errorf("services is required and must be a non-empty array of {service_id, verify_on_boot}")
	}
	items := make([]AgentServiceUpdateItem, 0, len(raw))
	for i, entry := range raw {
		m, ok := entry.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("services[%d] must be an object with service_id and verify_on_boot", i)
		}
		sid, ok := m["service_id"].(string)
		if !ok || sid == "" {
			return "", fmt.Errorf("services[%d].service_id is required", i)
		}
		vob, ok := m["verify_on_boot"].(bool)
		if !ok {
			return "", fmt.Errorf("services[%d].verify_on_boot is required (boolean)", i)
		}
		items = append(items, AgentServiceUpdateItem{ServiceID: sid, VerifyOnBoot: vob})
	}
	updated, err := updateAgentServices(agentID, items)
	if err != nil {
		return "", err
	}
	return toJSONString(map[string]interface{}{
		"agent_id": agentID,
		"updated":  updated,
	})
}

// agent: backup schedule ------------------------------------------------------

func handleAgentSetSchedule(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	sched := BackupSchedule{}
	if v, ok := optionalInt(args, "interval_in_minutes"); ok {
		sched.IntervalInMinutes = v
	}
	if v, ok := optionalInt(args, "start_hour"); ok {
		sched.StartHour = v
	}
	if v, ok := optionalInt(args, "end_hour"); ok {
		sched.EndHour = v
	}
	if raw, ok := args["days"].([]interface{}); ok {
		for _, d := range raw {
			if n, ok := d.(float64); ok {
				sched.Days = append(sched.Days, int(n))
			}
		}
	}
	if err := validateBackupSchedule(sched); err != nil {
		return "", err
	}

	payload := map[string]interface{}{
		"backup_schedule": sched,
	}
	return patchAgentJSON(agentID, payload)
}

func handleAgentClearSchedule(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	// Sending null clears the schedule on the API side.
	payload := map[string]interface{}{
		"backup_schedule": nil,
	}
	return patchAgentJSON(agentID, payload)
}

// agent: pause / resume backups ----------------------------------------------

func handleAgentPauseBackups(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{}
	if indef, ok := optionalBool(args, "indefinite"); ok && indef {
		payload["backup_paused_indefinite"] = true
	} else if until, ok := optionalString(args, "paused_until"); ok && until != "" {
		payload["backup_paused_until"] = until
	} else {
		return "", fmt.Errorf("provide either indefinite=true or paused_until=<RFC3339 timestamp>")
	}
	return patchAgentJSON(agentID, payload)
}

func handleAgentResumeBackups(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{
		"backup_resume": true,
	}
	return patchAgentJSON(agentID, payload)
}

// agent: retention / restore defaults / volumes / misc ------------------------

func handleAgentSetRetention(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	name, err := requireString(args, "retention_policy_name")
	if err != nil {
		return "", err
	}
	if err := validateRetentionPolicyName(name); err != nil {
		return "", err
	}
	maxAge, ok := optionalInt(args, "retention_policy_max_age_months")
	if !ok {
		return "", fmt.Errorf("retention_policy_max_age_months is required (one of 3, 6, 12, 24, 36, 84)")
	}
	switch maxAge {
	case 3, 6, 12, 24, 36, 84:
	default:
		return "", fmt.Errorf("retention_policy_max_age_months must be one of 3, 6, 12, 24, 36, 84 (got %d)", maxAge)
	}
	payload := map[string]interface{}{
		"local_retention_policy": LocalRetentionPolicy{
			RetentionPolicyName:         name,
			RetentionPolicyMaxAgeMonths: maxAge,
		},
	}
	return patchAgentJSON(agentID, payload)
}

func handleAgentSetRestoreDefaults(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	settings := DefaultRestoreSettings{}
	any := false
	if v, ok := optionalInt(args, "cpu_count"); ok {
		settings.CPUCount = &v
		any = true
	}
	if v, ok := optionalInt(args, "memory_mb"); ok {
		settings.MemoryMB = &v
		any = true
	}
	if v, ok := optionalString(args, "disk_bus"); ok {
		settings.DiskBus = &v
		any = true
	}
	if v, ok := optionalString(args, "network_model"); ok {
		settings.NetworkModel = &v
		any = true
	}
	if !any {
		return "", fmt.Errorf("provide at least one of: cpu_count, memory_mb, disk_bus, network_model")
	}
	payload := map[string]interface{}{
		"default_restore_settings": settings,
	}
	return patchAgentJSON(agentID, payload)
}

func handleAgentSetVolumes(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{}
	if raw, ok := args["volumes"].([]interface{}); ok {
		volumes := make([]VolumeSetting, 0, len(raw))
		for i, entry := range raw {
			m, ok := entry.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("volumes[%d] must be an object with volume_id and include", i)
			}
			vid, ok := m["volume_id"].(string)
			if !ok || vid == "" {
				return "", fmt.Errorf("volumes[%d].volume_id is required", i)
			}
			incl, ok := m["include"].(bool)
			if !ok {
				return "", fmt.Errorf("volumes[%d].include is required (boolean)", i)
			}
			vs := VolumeSetting{VolumeID: vid, Include: incl}
			if mp, ok := m["mount_points"].([]interface{}); ok {
				for _, p := range mp {
					if s, ok := p.(string); ok {
						vs.MountPoints = append(vs.MountPoints, s)
					}
				}
			}
			volumes = append(volumes, vs)
		}
		payload["volumes"] = volumes
	}
	if v, ok := optionalBool(args, "volumes_include_default"); ok {
		payload["volumes_include_default"] = v
	}
	if len(payload) == 0 {
		return "", fmt.Errorf("provide at least one of: volumes, volumes_include_default")
	}
	return patchAgentJSON(agentID, payload)
}

func handleAgentSetFileIndex(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	enabled, ok := optionalBool(args, "file_index_enabled")
	if !ok {
		return "", fmt.Errorf("file_index_enabled is required (boolean)")
	}
	return patchAgentJSON(agentID, map[string]interface{}{
		"file_index_enabled": enabled,
	})
}

func handleAgentSetTimezone(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	tz, err := requireString(args, "timezone")
	if err != nil {
		return "", err
	}
	return patchAgentJSON(agentID, map[string]interface{}{
		"timezone": tz,
	})
}

func handleAgentSetComments(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	comments, ok := optionalString(args, "comments")
	if !ok {
		return "", fmt.Errorf("comments is required")
	}
	return patchAgentJSON(agentID, map[string]interface{}{
		"comments": comments,
	})
}

func handleAgentUpdateAlertConfig(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	alertType, err := requireString(args, "alert_type")
	if err != nil {
		return "", err
	}
	cfg := AlertConfig{AlertType: alertType}
	used := false
	if v, ok := optionalBool(args, "indefinite"); ok {
		cfg.AlertPausedIndefinite = &v
		used = true
	}
	if v, ok := optionalBool(args, "resume"); ok && v {
		t := true
		cfg.AlertResume = &t
		used = true
	}
	if v, ok := optionalInt(args, "pause_for_minutes"); ok {
		switch v {
		case 0, 30, 60, 120, 240, 480, 1440, 10080:
		default:
			return "", fmt.Errorf("pause_for_minutes must be one of 0, 30, 60, 120, 240, 480, 1440, 10080 (got %d)", v)
		}
		cfg.PauseForMinutes = &v
		used = true
	}
	if !used {
		return "", fmt.Errorf("provide one of: indefinite, resume, pause_for_minutes")
	}
	return patchAgentJSON(agentID, map[string]interface{}{
		"alert_configs": []AlertConfig{cfg},
	})
}

// device: network -------------------------------------------------------------

func handleDeviceGetNetwork(args map[string]interface{}) (string, error) {
	deviceID, err := requireString(args, "device_id")
	if err != nil {
		return "", err
	}
	dn, err := getDeviceNetwork(deviceID)
	if err != nil {
		return "", err
	}
	return toJSONString(dn)
}

func handleDeviceUpdateNetwork(args map[string]interface{}) (string, error) {
	deviceID, err := requireString(args, "device_id")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{}
	if v, ok := optionalString(args, "network_mode"); ok {
		payload["network_mode"] = v
	}
	if v, ok := optionalString(args, "network_address"); ok {
		payload["network_address"] = v
	}
	if v, ok := optionalString(args, "network_gateway"); ok {
		payload["network_gateway"] = v
	}
	if v, ok := optionalString(args, "dns_server_primary"); ok {
		payload["dns_server_primary"] = v
	}
	if v, ok := optionalString(args, "dns_server_secondary"); ok {
		payload["dns_server_secondary"] = v
	}
	if len(payload) == 0 {
		return "", fmt.Errorf("provide at least one of: network_mode, network_address, network_gateway, dns_server_primary, dns_server_secondary")
	}
	dn, err := updateDeviceNetwork(deviceID, payload)
	if err != nil {
		return "", err
	}
	return toJSONString(dn)
}

// device: VLAN ----------------------------------------------------------------

func handleDeviceListVLANs(args map[string]interface{}) (string, error) {
	deviceID, err := requireString(args, "device_id")
	if err != nil {
		return "", err
	}
	vlans, err := listDeviceVLANs(deviceID)
	if err != nil {
		return "", err
	}
	return toJSONString(map[string]interface{}{
		"device_id": deviceID,
		"vlans":     vlans,
	})
}

func handleDeviceGetVLAN(args map[string]interface{}) (string, error) {
	deviceID, err := requireString(args, "device_id")
	if err != nil {
		return "", err
	}
	vlanID, err := requireString(args, "vlan_id")
	if err != nil {
		return "", err
	}
	vlan, err := getDeviceVLAN(deviceID, vlanID)
	if err != nil {
		return "", err
	}
	return toJSONString(vlan)
}

func handleDeviceCreateVLAN(args map[string]interface{}) (string, error) {
	deviceID, err := requireString(args, "device_id")
	if err != nil {
		return "", err
	}
	name, err := requireString(args, "name")
	if err != nil {
		return "", err
	}
	tag, ok := optionalInt(args, "vlan_tag")
	if !ok {
		return "", fmt.Errorf("vlan_tag is required (integer)")
	}
	mode, err := requireString(args, "network_mode")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{
		"name":         name,
		"vlan_tag":     tag,
		"network_mode": mode,
	}
	if v, ok := optionalString(args, "ip_address"); ok {
		payload["ip_address"] = v
	}
	if v, ok := optionalString(args, "gateway"); ok {
		payload["gateway"] = v
	}
	vlan, err := createDeviceVLAN(deviceID, payload)
	if err != nil {
		return "", err
	}
	return toJSONString(vlan)
}

func handleDeviceUpdateVLAN(args map[string]interface{}) (string, error) {
	deviceID, err := requireString(args, "device_id")
	if err != nil {
		return "", err
	}
	vlanID, err := requireString(args, "vlan_id")
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{}
	if v, ok := optionalString(args, "name"); ok {
		payload["name"] = v
	}
	if v, ok := optionalInt(args, "vlan_tag"); ok {
		payload["vlan_tag"] = v
	}
	if v, ok := optionalString(args, "network_mode"); ok {
		payload["network_mode"] = v
	}
	if v, ok := optionalString(args, "ip_address"); ok {
		payload["ip_address"] = v
	}
	if v, ok := optionalString(args, "gateway"); ok {
		payload["gateway"] = v
	}
	if len(payload) == 0 {
		return "", fmt.Errorf("provide at least one of: name, vlan_tag, network_mode, ip_address, gateway")
	}
	vlan, err := updateDeviceVLAN(deviceID, vlanID, payload)
	if err != nil {
		return "", err
	}
	return toJSONString(vlan)
}

func handleDeviceDeleteVLAN(args map[string]interface{}) (string, error) {
	deviceID, err := requireString(args, "device_id")
	if err != nil {
		return "", err
	}
	vlanID, err := requireString(args, "vlan_id")
	if err != nil {
		return "", err
	}
	if err := deleteDeviceVLAN(deviceID, vlanID); err != nil {
		return "", err
	}
	return toJSONString(map[string]interface{}{
		"device_id": deviceID,
		"vlan_id":   vlanID,
		"deleted":   true,
	})
}

// snapshot: service verification ---------------------------------------------

func handleSnapshotGetServiceVerification(args map[string]interface{}) (string, error) {
	snapshotID, err := requireString(args, "snapshot_id")
	if err != nil {
		return "", err
	}
	res, err := getSnapshotServiceVerification(snapshotID)
	if err != nil {
		return "", err
	}
	return toJSONString(res)
}

// user_management: avatar -----------------------------------------------------

func handleGetUserAvatar(args map[string]interface{}) (string, error) {
	userID, err := requireString(args, "user_id")
	if err != nil {
		return "", err
	}
	av, err := getUserAvatar(userID)
	if err != nil {
		return "", err
	}
	return toJSONString(av)
}

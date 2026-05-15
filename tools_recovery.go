package main

// slide_recovery: the "I need to actually recover something" toolkit.
// Consolidates virtual machines (boot snapshots), image exports
// (download disk images), and DR networks (so booted VMs can talk to
// the world). Replaces the v3 trio of slide_vms + slide_restores image
// surface + slide_networks under one task-oriented umbrella.

func handleRecoveryTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_recovery", ToolOperations{
		// Virtual machines
		"list_vms":         listVirtualMachines,
		"get_vm":           getVirtualMachine,
		"boot_vm":          createVirtualMachine,
		"update_vm":        updateVirtualMachine,
		"delete_vm":        deleteVirtualMachine,
		"get_rdp_bookmark": generateRDPBookmark,

		// Image exports
		"list_images":  listImageExports,
		"get_image":    getImageExport,
		"export_image": createImageExport,
		"delete_image": deleteImageExport,
		"browse_image": browseImageExport,

		// DR networks
		"list_networks":  listNetworks,
		"get_network":    getNetwork,
		"create_network": createNetwork,
		"update_network": updateNetwork,
		"delete_network": deleteNetwork,

		// IPSec / port-forward / WireGuard
		"create_ipsec":        createNetworkIPSecConn,
		"update_ipsec":        updateNetworkIPSecConn,
		"delete_ipsec":        deleteNetworkIPSecConn,
		"create_port_forward": createNetworkPortForward,
		"update_port_forward": updateNetworkPortForward,
		"delete_port_forward": deleteNetworkPortForward,
		"create_wg_peer":      createNetworkWGPeer,
		"update_wg_peer":      updateNetworkWGPeer,
		"delete_wg_peer":      deleteNetworkWGPeer,
	}), args)
}

var recoveryOperationEnums = []string{
	"list_vms", "get_vm", "boot_vm", "update_vm", "delete_vm", "get_rdp_bookmark",
	"list_images", "get_image", "export_image", "delete_image", "browse_image",
	"list_networks", "get_network", "create_network", "update_network", "delete_network",
	"create_ipsec", "update_ipsec", "delete_ipsec",
	"create_port_forward", "update_port_forward", "delete_port_forward",
	"create_wg_peer", "update_wg_peer", "delete_wg_peer",
}

func getRecoveryToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        recoveryOperationEnums,
		},

		// VM identification
		"virt_id":        map[string]interface{}{"type": "string", "description": "VM ID. Required for `get_vm`, `update_vm`, `delete_vm`, `get_rdp_bookmark`."},
		"snapshot_id":    map[string]interface{}{"type": "string", "description": "Snapshot to boot/export. Required for `boot_vm` and `export_image`."},
		"device_id":      map[string]interface{}{"type": "string", "description": "Device that will host the VM/image. Required for `boot_vm` and `export_image`."},
		"cpu_count":      map[string]interface{}{"type": "number", "description": "VM vCPUs.", "enum": []int{1, 2, 4, 8, 16}},
		"memory_in_mb":   map[string]interface{}{"type": "number", "description": "VM memory in MB."},
		"disk_bus":       map[string]interface{}{"type": "string", "description": "Disk bus.", "enum": []string{"sata", "virtio"}},
		"network_model":  map[string]interface{}{"type": "string", "description": "Network model.", "enum": []string{"hypervisor_default", "e1000", "rtl8139", "virtio"}},
		"network_type":   map[string]interface{}{"type": "string", "description": "Network attachment.", "enum": []string{"network", "network-isolated", "bridge", "network-id"}},
		"network_source": map[string]interface{}{"type": "string", "description": "Network ID when `network_type` is `network-id`."},
		"boot_mods":      map[string]interface{}{"type": "array", "description": "Boot modifications.", "items": map[string]interface{}{"type": "string"}},
		"state":          map[string]interface{}{"type": "string", "description": "VM lifecycle state for `update_vm`.", "enum": []string{"running", "stopped", "paused"}},
		"expires_at":     map[string]interface{}{"type": "string", "description": "RFC3339 expiry timestamp for `update_vm`."},

		// Image export
		"image_export_id": map[string]interface{}{"type": "string", "description": "Image export ID. Required for `get_image`, `delete_image`, `browse_image`."},
		"image_type":      map[string]interface{}{"type": "string", "description": "Disk image format. Required for `export_image`.", "enum": []string{"vhd", "vhdx", "vmdk", "qcow2", "raw"}},

		// Network identification
		"network_id":       map[string]interface{}{"type": "string", "description": "DR network ID. Required for `get_network`, `update_network`, `delete_network`, and all peer/port-forward/IPSec ops."},
		"name":             map[string]interface{}{"type": "string", "description": "Friendly name. Required for `create_network`."},
		"type":             map[string]interface{}{"type": "string", "description": "DR network type. Required for `create_network`. (Note: use `network_type` for VM `boot_vm` operations.)", "enum": []string{"standard", "bridge-lan"}},
		"bridge_device_id": map[string]interface{}{"type": "string", "description": "Device whose LAN to bridge into. Used with `create_network` when type=bridge-lan."},
		"router_prefix":    map[string]interface{}{"type": "string", "description": "Router IP in CIDR notation (e.g. `192.168.1.1/24`)."},
		"comments":         map[string]interface{}{"type": "string", "description": "Free-form comments."},
		"client_id":        map[string]interface{}{"type": "string", "description": "Client ID to assign or filter by."},
		"dhcp":             map[string]interface{}{"type": "boolean", "description": "Enable DHCP on the DR network."},
		"dhcp_range_start": map[string]interface{}{"type": "string", "description": "DHCP range start IP."},
		"dhcp_range_end":   map[string]interface{}{"type": "string", "description": "DHCP range end IP."},
		"internet":         map[string]interface{}{"type": "boolean", "description": "Allow internet egress."},
		"nameservers":      map[string]interface{}{"type": "array", "description": "DNS servers.", "items": map[string]interface{}{"type": "string"}},

		// Port forward
		"port_forward_id": map[string]interface{}{"type": "string", "description": "Port-forward ID. Required for update/delete."},
		"proto":           map[string]interface{}{"type": "string", "description": "Port-forward protocol.", "enum": []string{"tcp", "udp"}},
		"dest":            map[string]interface{}{"type": "string", "description": "Port-forward destination host:port."},

		// IPSec
		"ipsec_id": map[string]interface{}{"type": "string", "description": "IPSec connection ID. Required for update/delete."},

		// WireGuard
		"wg_peer_id": map[string]interface{}{"type": "string", "description": "WireGuard peer ID. Required for update/delete."},
	}
	for k, v := range commonListProperties() {
		if _, exists := props[k]; !exists {
			props[k] = v
		}
	}

	return ToolInfo{
		Name: "slide_recovery",
		Description: "Slide MCP - actually recover something from a Slide backup. " +
			"REACH FOR THIS whenever the user mentions disaster recovery, DR, BCDR, RTO, failover, " +
			"'boot a VM from a snapshot', 'recovery VM', 'spin up a recovered server', 'image export' (VHD/VHDX/VMDK/QCOW2/RAW), " +
			"RDP into a recovered server, DR network, VPN/WireGuard/IPSec to a recovered VM, or 'I need to fail over to a Slide snapshot'. " +
			"Three families: " +
			"VMs (`list_vms`, `get_vm`, `boot_vm` <- creates a running VM from a snapshot, `update_vm` to start/stop/pause, `delete_vm`, `get_rdp_bookmark`), " +
			"image exports (`list_images`, `get_image`, `export_image` <- VHD/VHDX/VMDK/QCOW2/RAW for external virtualization, `delete_image`, `browse_image`), " +
			"and DR networks for booted VMs (`list_networks`/`get_network`/`create_network`/`update_network`/`delete_network` plus `create_ipsec`/`create_port_forward`/`create_wg_peer` and matching update/delete). " +
			"Use this when the user wants to actually recover something - boot a server, get a disk image, set up VPN access to recovered VMs.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get_vm"), "then": req("virt_id")},
				{"if": ifOp("boot_vm"), "then": req("snapshot_id", "device_id")},
				{"if": ifOp("update_vm"), "then": req("virt_id")},
				{"if": ifOp("delete_vm"), "then": req("virt_id")},
				{"if": ifOp("get_rdp_bookmark"), "then": req("virt_id")},

				{"if": ifOp("get_image"), "then": req("image_export_id")},
				{"if": ifOp("export_image"), "then": req("snapshot_id", "device_id", "image_type")},
				{"if": ifOp("delete_image"), "then": req("image_export_id")},
				{"if": ifOp("browse_image"), "then": req("image_export_id")},

				{"if": ifOp("get_network"), "then": req("network_id")},
				{"if": ifOp("create_network"), "then": req("name", "type")},
				{"if": ifOp("update_network"), "then": req("network_id")},
				{"if": ifOp("delete_network"), "then": req("network_id")},

				{"if": ifOp("create_ipsec"), "then": req("network_id")},
				{"if": ifOp("update_ipsec"), "then": req("network_id", "ipsec_id")},
				{"if": ifOp("delete_ipsec"), "then": req("network_id", "ipsec_id")},

				{"if": ifOp("create_port_forward"), "then": req("network_id")},
				{"if": ifOp("update_port_forward"), "then": req("network_id", "port_forward_id")},
				{"if": ifOp("delete_port_forward"), "then": req("network_id", "port_forward_id")},

				{"if": ifOp("create_wg_peer"), "then": req("network_id")},
				{"if": ifOp("update_wg_peer"), "then": req("network_id", "wg_peer_id")},
				{"if": ifOp("delete_wg_peer"), "then": req("network_id", "wg_peer_id")},
			},
		},
	}
}

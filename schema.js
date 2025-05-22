/**
 * Slide API Schema
 * This file documents the available endpoints and models in the Slide API
 */

const slideApiSchema = {
  // Entity relationships and conceptual model
  relationships: {
    hierarchy: {
      description: "Hierarchical relationship between Slide entities",
      model: [
        "Device → Agent → Backup → Snapshot",
        "A Device can have multiple Agents",
        "An Agent can have multiple Backups",
        "A successful Backup creates a Snapshot"
      ]
    },
    device: {
      description: "A physical Slide Box hardware appliance",
      hasMany: ["agents"],
      relationship: "A Device is a Slide Box that provides backup storage and disaster recovery capabilities."
    },
    agent: {
      description: "Software installed on a system to be protected",
      belongsTo: ["device"],
      hasMany: ["backups"],
      relationship: "An Agent runs on a client system (like a server or workstation) and connects to a Device (Slide Box) for backup."
    },
    backup: {
      description: "A backup operation that copies data from an Agent to a Device",
      belongsTo: ["agent"],
      hasOne: ["snapshot"],
      relationship: "A Backup is a point-in-time operation that transfers data from an Agent to a Device. Successful backups create Snapshots."
    },
    snapshot: {
      description: "A point-in-time copy of data created by a successful backup",
      belongsTo: ["backup", "agent"],
      relationship: "A Snapshot is created when a Backup completes successfully. It represents a recoverable point-in-time for the Agent's data.",
      recoveryOptions: {
        description: "Snapshots can be recovered in multiple ways",
        options: [
          {
            name: "Virtualization",
            description: "Create a virtual machine from a snapshot to run the system in a virtualized environment",
            entity: "virtualMachines",
            use_case: "Disaster recovery, testing, development environments"
          },
          {
            name: "Image Export",
            description: "Export a snapshot as a disk image (VHD, VHDX, etc.) for use with other virtualization platforms",
            entity: "imageExports",
            use_case: "Migration to other platforms, offline recovery"
          },
          {
            name: "File Restore",
            description: "Mount a snapshot to browse and restore individual files and folders",
            entity: "fileRestores",
            use_case: "Recovering specific files without full system restore"
          }
        ]
      }
    },
    virtualMachine: {
      description: "A virtual machine created from a snapshot",
      belongsTo: ["snapshot", "device"],
      relationship: "A VirtualMachine is created from a Snapshot and runs on a Device. It provides a virtualized environment of the Agent's system at the time of the Snapshot."
    },
    imageExport: {
      description: "An exported disk image created from a snapshot",
      belongsTo: ["snapshot", "device"],
      relationship: "An ImageExport is created from a Snapshot and stored on a Device. It allows snapshots to be used with other virtualization platforms."
    },
    fileRestore: {
      description: "A mounted snapshot for browsing and restoring files",
      belongsTo: ["snapshot", "device"],
      relationship: "A FileRestore mounts a Snapshot on a Device to allow browsing and restoring individual files and folders."
    }
  },
  
  // Sample responses to illustrate data structures and relationships
  sampleResponses: {
    device: {
      list: {
        "pagination": {
          "total": 2
        },
        "data": [
          {
            "device_id": "d_0123456789ab",
            "display_name": "Primary Slide Box",
            "hostname": "slide-primary",
            "client_id": "c_0123456789ab",
            "last_seen_at": "2024-05-22T14:08:12Z",
            "addresses": [
              {
                "mac": "62:bb:d3:0d:db:7d",
                "ips": ["192.168.1.100"]
              }
            ],
            "public_ip_address": "74.83.124.111",
            "image_version": "1.2.3",
            "package_version": "1.2.3",
            "storage_used_bytes": 274877906944,
            "storage_total_bytes": 1099511627776
          },
          {
            "device_id": "d_abcdef123456",
            "display_name": "Secondary Slide Box",
            "hostname": "slide-secondary",
            "last_seen_at": "2024-05-22T14:08:12Z",
            "addresses": [
              {
                "mac": "62:bb:d3:0d:db:7e",
                "ips": ["192.168.1.101"]
              }
            ],
            "public_ip_address": "74.83.124.112",
            "image_version": "1.2.3",
            "package_version": "1.2.3",
            "storage_used_bytes": 137438953472,
            "storage_total_bytes": 1099511627776
          }
        ]
      }
    },
    agent: {
      list: {
        "pagination": {
          "total": 2
        },
        "data": [
          {
            "agent_id": "a_0123456789ab",
            "device_id": "d_0123456789ab",
            "display_name": "Windows Server",
            "hostname": "win-server-2019",
            "last_seen_at": "2024-05-22T14:08:12Z",
            "addresses": [
              {
                "mac": "aa:bb:cc:dd:ee:ff",
                "ips": ["192.168.1.200"]
              }
            ],
            "public_ip_address": "74.83.124.113",
            "agent_version": "1.2.3",
            "platform": "Microsoft Windows Server 2019",
            "os": "windows",
            "os_version": "10.0.17763"
          },
          {
            "agent_id": "a_abcdef123456",
            "device_id": "d_0123456789ab",
            "display_name": "Linux Server",
            "hostname": "ubuntu-server",
            "last_seen_at": "2024-05-22T14:08:12Z",
            "addresses": [
              {
                "mac": "11:22:33:44:55:66",
                "ips": ["192.168.1.201"]
              }
            ],
            "public_ip_address": "74.83.124.114",
            "agent_version": "1.2.3",
            "platform": "Ubuntu 20.04 LTS",
            "os": "linux",
            "os_version": "5.4.0"
          }
        ]
      }
    },
    backup: {
      list: {
        "pagination": {
          "total": 2
        },
        "data": [
          {
            "backup_id": "b_0123456789ab",
            "agent_id": "a_0123456789ab",
            "started_at": "2024-05-22T01:25:08Z",
            "ended_at": "2024-05-22T01:40:08Z",
            "status": "succeeded",
            "snapshot_id": "s_0123456789ab"
          },
          {
            "backup_id": "b_abcdef123456",
            "agent_id": "a_0123456789ab",
            "started_at": "2024-05-21T01:25:08Z",
            "ended_at": "2024-05-21T01:40:08Z",
            "status": "succeeded",
            "snapshot_id": "s_abcdef123456"
          }
        ]
      }
    },
    snapshot: {
      list: {
        "pagination": {
          "total": 2
        },
        "data": [
          {
            "snapshot_id": "s_0123456789ab",
            "agent_id": "a_0123456789ab",
            "locations": [
              {
                "type": "local",
                "device_id": "d_0123456789ab"
              },
              {
                "type": "cloud"
              }
            ],
            "backup_started_at": "2024-05-22T01:25:08Z",
            "backup_ended_at": "2024-05-22T01:40:08Z",
            "verify_boot_status": "success",
            "verify_fs_status": "success"
          },
          {
            "snapshot_id": "s_abcdef123456",
            "agent_id": "a_0123456789ab",
            "locations": [
              {
                "type": "local",
                "device_id": "d_0123456789ab"
              },
              {
                "type": "cloud"
              }
            ],
            "backup_started_at": "2024-05-21T01:25:08Z",
            "backup_ended_at": "2024-05-21T01:40:08Z",
            "verify_boot_status": "success",
            "verify_fs_status": "success"
          }
        ]
      }
    }
  },
  
  // Endpoints
  endpoints: {
    devices: {
      list: {
        method: 'GET',
        path: '/device',
        description: 'List all devices',
        params: ['limit', 'offset', 'sort_by', 'sort_asc', 'client_id']
      },
      get: {
        method: 'GET',
        path: '/device/{device_id}',
        description: 'Get a specific device by ID',
        params: ['device_id']
      },
      update: {
        method: 'PATCH',
        path: '/device/{device_id}',
        description: 'Update a device',
        params: ['device_id'],
        body: ['display_name', 'hostname', 'client_id']
      }
    },
    agents: {
      list: {
        method: 'GET',
        path: '/agent',
        description: 'List all agents',
        params: ['device_id', 'limit', 'offset', 'sort_by', 'sort_asc', 'client_id']
      },
      get: {
        method: 'GET',
        path: '/agent/{agent_id}',
        description: 'Get a specific agent by ID',
        params: ['agent_id']
      },
      update: {
        method: 'PATCH',
        path: '/agent/{agent_id}',
        description: 'Update an agent',
        params: ['agent_id'],
        body: ['display_name']
      },
      create: {
        method: 'POST',
        path: '/agent',
        description: 'Create an agent for auto-pair installation',
        body: ['display_name', 'device_id']
      },
      pair: {
        method: 'POST',
        path: '/agent/pair',
        description: 'Pair an agent',
        body: ['pair_code', 'device_id']
      }
    },
    backups: {
      list: {
        method: 'GET',
        path: '/backup',
        description: 'List all backups',
        params: ['agent_id', 'snapshot_id', 'device_id', 'limit', 'offset', 'sort_by', 'sort_asc']
      },
      get: {
        method: 'GET',
        path: '/backup/{backup_id}',
        description: 'Get a specific backup by ID',
        params: ['backup_id']
      },
      start: {
        method: 'POST',
        path: '/backup',
        description: 'Start a backup',
        body: ['agent_id']
      }
    },
    snapshots: {
      list: {
        method: 'GET',
        path: '/snapshot',
        description: 'List all snapshots',
        params: ['agent_id', 'limit', 'offset', 'snapshot_location', 'sort_by', 'sort_asc']
      },
      get: {
        method: 'GET',
        path: '/snapshot/{snapshot_id}',
        description: 'Get a specific snapshot by ID',
        params: ['snapshot_id']
      }
    },
    fileRestores: {
      list: {
        method: 'GET',
        path: '/restore/file',
        description: 'List all file restores',
        params: ['limit', 'offset', 'sort_by', 'sort_asc']
      },
      get: {
        method: 'GET',
        path: '/restore/file/{file_restore_id}',
        description: 'Get a specific file restore by ID',
        params: ['file_restore_id']
      },
      create: {
        method: 'POST',
        path: '/restore/file',
        description: 'Create a file restore',
        body: ['snapshot_id', 'device_id']
      },
      delete: {
        method: 'DELETE',
        path: '/restore/file/{file_restore_id}',
        description: 'Delete a file restore',
        params: ['file_restore_id']
      },
      browse: {
        method: 'GET',
        path: '/restore/file/{file_restore_id}/browse',
        description: 'Browse files in a file restore',
        params: ['file_restore_id', 'limit', 'offset', 'path']
      }
    },
    imageExports: {
      list: {
        method: 'GET',
        path: '/restore/image',
        description: 'List all image exports',
        params: ['limit', 'offset', 'sort_by', 'sort_asc']
      },
      get: {
        method: 'GET',
        path: '/restore/image/{image_export_id}',
        description: 'Get a specific image export by ID',
        params: ['image_export_id']
      },
      create: {
        method: 'POST',
        path: '/restore/image',
        description: 'Create an image export',
        body: ['snapshot_id', 'device_id', 'image_type', 'boot_mods']
      },
      delete: {
        method: 'DELETE',
        path: '/restore/image/{image_export_id}',
        description: 'Delete an image export',
        params: ['image_export_id']
      },
      browse: {
        method: 'GET',
        path: '/restore/image/{image_export_id}/browse',
        description: 'Browse files in an image export',
        params: ['image_export_id', 'limit', 'offset']
      }
    },
    virtualMachines: {
      list: {
        method: 'GET',
        path: '/restore/virt',
        description: 'List all virtual machines',
        params: ['limit', 'offset', 'sort_by', 'sort_asc']
      },
      get: {
        method: 'GET',
        path: '/restore/virt/{virt_id}',
        description: 'Get a specific virtual machine by ID',
        params: ['virt_id']
      },
      create: {
        method: 'POST',
        path: '/restore/virt',
        description: 'Create a virtual machine',
        body: ['snapshot_id', 'device_id', 'cpu_count', 'memory_in_mb', 'disk_bus', 'network_model', 'network_type', 'network_source', 'boot_mods']
      },
      update: {
        method: 'PATCH',
        path: '/restore/virt/{virt_id}',
        description: 'Update a virtual machine',
        params: ['virt_id'],
        body: ['state', 'expires_at', 'memory_in_mb', 'cpu_count']
      },
      delete: {
        method: 'DELETE',
        path: '/restore/virt/{virt_id}',
        description: 'Delete a virtual machine',
        params: ['virt_id']
      }
    },
    networks: {
      list: {
        method: 'GET',
        path: '/network',
        description: 'List all networks [EXPERIMENTAL]',
        params: ['limit', 'offset', 'sort_by', 'sort_asc']
      },
      get: {
        method: 'GET',
        path: '/network/{network_id}',
        description: 'Get a specific network by ID [EXPERIMENTAL]',
        params: ['network_id']
      },
      create: {
        method: 'POST',
        path: '/network',
        description: 'Create a network [EXPERIMENTAL]',
        body: ['name', 'type', 'client_id', 'comments', 'bridge_device_id', 'router_prefix', 'dhcp', 'dhcp_range_start', 'dhcp_range_end', 'nameservers', 'internet']
      },
      update: {
        method: 'PATCH',
        path: '/network/{network_id}',
        description: 'Update a network [EXPERIMENTAL]',
        params: ['network_id'],
        body: ['name', 'comments', 'router_prefix', 'dhcp', 'dhcp_range_start', 'dhcp_range_end', 'nameservers', 'internet', 'wg']
      },
      delete: {
        method: 'DELETE',
        path: '/network/{network_id}',
        description: 'Delete a network [EXPERIMENTAL]',
        params: ['network_id']
      },
      createPortForward: {
        method: 'POST',
        path: '/network/{network_id}/port-forwards',
        description: 'Create a network port forward [EXPERIMENTAL]',
        params: ['network_id'],
        body: ['proto', 'dest']
      },
      deletePortForward: {
        method: 'DELETE',
        path: '/network/{network_id}/port-forwards',
        description: 'Delete a network port forward [EXPERIMENTAL]',
        params: ['network_id'],
        body: ['proto', 'port']
      },
      createWGPeer: {
        method: 'POST',
        path: '/network/{network_id}/wg-peers',
        description: 'Create a network WireGuard peer [EXPERIMENTAL]',
        params: ['network_id'],
        body: ['peer_name', 'remote_networks']
      },
      updateWGPeer: {
        method: 'PATCH',
        path: '/network/{network_id}/wg-peers',
        description: 'Update a network WireGuard peer [EXPERIMENTAL]',
        params: ['network_id'],
        body: ['wg_address', 'peer_name', 'remote_networks']
      },
      deleteWGPeer: {
        method: 'DELETE',
        path: '/network/{network_id}/wg-peers',
        description: 'Delete a network WireGuard peer [EXPERIMENTAL]',
        params: ['network_id'],
        body: ['wg_address']
      }
    },
    users: {
      list: {
        method: 'GET',
        path: '/user',
        description: 'List all users',
        params: ['limit', 'offset', 'sort_by', 'sort_asc']
      },
      get: {
        method: 'GET',
        path: '/user/{user_id}',
        description: 'Get a specific user by ID',
        params: ['user_id']
      }
    },
    alerts: {
      list: {
        method: 'GET',
        path: '/alert',
        description: 'List all alerts',
        params: ['device_id', 'agent_id', 'resolved', 'limit', 'offset', 'sort_by', 'sort_asc']
      },
      get: {
        method: 'GET',
        path: '/alert/{alert_id}',
        description: 'Get a specific alert by ID',
        params: ['alert_id']
      },
      update: {
        method: 'PATCH',
        path: '/alert/{alert_id}',
        description: 'Update an alert',
        params: ['alert_id'],
        body: ['resolved']
      }
    },
    accounts: {
      list: {
        method: 'GET',
        path: '/account',
        description: 'List all accounts',
        params: ['limit', 'offset', 'sort_by', 'sort_asc']
      },
      get: {
        method: 'GET',
        path: '/account/{account_id}',
        description: 'Get a specific account by ID',
        params: ['account_id']
      },
      update: {
        method: 'PATCH',
        path: '/account/{account_id}',
        description: 'Update an account',
        params: ['account_id'],
        body: ['alert_emails']
      }
    }
  },
  
  // Common models
  models: {
    device: {
      device_id: 'ID of the device (e.g., d_0123456789ab)',
      display_name: 'Customizable display name',
      hostname: 'Hostname of the device',
      client_id: 'ID of the client',
      last_seen_at: 'Last time the device was seen',
      ip_addresses: 'IP addresses of the device (deprecated)',
      addresses: 'Network addresses of the device',
      public_ip_address: 'Public IP address of the device',
      image_version: 'Version of the device image',
      package_version: 'Version of the Slide package',
      storage_used_bytes: 'Used storage in bytes',
      storage_total_bytes: 'Total storage in bytes',
      hardware_model_name: 'Hardware model name',
      service_model_name: 'Service model name',
      service_status: 'Status of the service'
    },
    agent: {
      agent_id: 'ID of the agent (e.g., a_0123456789ab)',
      device_id: 'ID of the device',
      client_id: 'ID of the client',
      display_name: 'Customizable display name',
      hostname: 'Hostname of the agent',
      last_seen_at: 'Last time the agent was seen',
      ip_addresses: 'IP addresses of the agent (deprecated)',
      addresses: 'Network addresses of the agent',
      public_ip_address: 'Public IP address of the agent',
      agent_version: 'Version of the agent',
      platform: 'OS platform of the agent',
      os: 'OS of the agent',
      os_version: 'OS version of the agent'
    },
    backup: {
      backup_id: 'ID of the backup (e.g., b_0123456789ab)',
      agent_id: 'ID of the agent',
      started_at: 'Start time of the backup',
      ended_at: 'End time of the backup',
      status: 'Status of the backup',
      error_code: 'Error code of the backup',
      error_message: 'Error message of the backup',
      snapshot_id: 'ID of the snapshot'
    },
    snapshot: {
      snapshot_id: 'ID of the snapshot (e.g., s_0123456789ab)',
      agent_id: 'ID of the agent',
      locations: 'Locations of the snapshot',
      backup_started_at: 'Start time of the backup',
      backup_ended_at: 'End time of the backup',
      deleted: 'When the snapshot was deleted',
      deletions: 'Deletion details'
    }
  }
};

module.exports = slideApiSchema; 
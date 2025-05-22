/**
 * Example script for Claude Desktop integration with Slide MCP
 * 
 * This is a demonstration of how you would integrate the Slide MCP
 * with Claude using the tool calling functionality.
 */

require('dotenv').config();
const { createClient } = require('./index');

// Get API key from environment variables
const apiKey = process.env.SLIDE_API_KEY;

if (!apiKey) {
  console.error('ERROR: No API key found. Please create a .env file with SLIDE_API_KEY');
  process.exit(1);
}

// Create Slide client
const slideClient = createClient(apiKey);

// Define tools that Claude can use to interact with Slide API
const slideTools = [
  // Devices
  {
    name: "slide_get_devices",
    description: "Get a list of Slide devices",
    input_schema: {
      type: "object",
      properties: {
        limit: {
          type: "integer",
          description: "Maximum number of devices to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_device",
    description: "Get details of a specific device by ID",
    input_schema: {
      type: "object",
      properties: {
        device_id: {
          type: "string",
          description: "ID of the device to retrieve"
        }
      },
      required: ["device_id"]
    }
  },
  {
    name: "slide_update_device",
    description: "Update a device's information",
    input_schema: {
      type: "object",
      properties: {
        device_id: {
          type: "string",
          description: "ID of the device to update"
        },
        data: {
          type: "object",
          description: "Updated device data"
        }
      },
      required: ["device_id", "data"]
    }
  },
  
  // Agents
  {
    name: "slide_get_agents",
    description: "Get a list of Slide agents, optionally filtered by device_id",
    input_schema: {
      type: "object",
      properties: {
        device_id: {
          type: "string",
          description: "Filter agents by device ID"
        },
        limit: {
          type: "integer",
          description: "Maximum number of agents to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_agent",
    description: "Get details of a specific agent by ID",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "ID of the agent to retrieve"
        }
      },
      required: ["agent_id"]
    }
  },
  {
    name: "slide_update_agent",
    description: "Update an agent's information",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "ID of the agent to update"
        },
        data: {
          type: "object",
          description: "Updated agent data"
        }
      },
      required: ["agent_id", "data"]
    }
  },
  {
    name: "slide_create_agent_pair",
    description: "Create a new agent pair",
    input_schema: {
      type: "object",
      properties: {
        data: {
          type: "object",
          description: "Agent pair creation data"
        }
      },
      required: ["data"]
    }
  },
  {
    name: "slide_pair_agent",
    description: "Pair an existing agent",
    input_schema: {
      type: "object",
      properties: {
        data: {
          type: "object",
          description: "Agent pairing data"
        }
      },
      required: ["data"]
    }
  },
  
  // Backups
  {
    name: "slide_get_backups",
    description: "Get a list of backups",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "Filter backups by agent ID"
        },
        limit: {
          type: "integer",
          description: "Maximum number of backups to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_backup",
    description: "Get details of a specific backup by ID",
    input_schema: {
      type: "object",
      properties: {
        backup_id: {
          type: "string",
          description: "ID of the backup to retrieve"
        }
      },
      required: ["backup_id"]
    }
  },
  {
    name: "slide_start_backup",
    description: "Start a new backup for a specific agent",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "ID of the agent to backup"
        }
      },
      required: ["agent_id"]
    }
  },
  
  // Snapshots
  {
    name: "slide_get_snapshots",
    description: "Get a list of snapshots",
    input_schema: {
      type: "object",
      properties: {
        backup_id: {
          type: "string",
          description: "Filter snapshots by backup ID"
        },
        limit: {
          type: "integer",
          description: "Maximum number of snapshots to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_snapshot",
    description: "Get details of a specific snapshot by ID",
    input_schema: {
      type: "object",
      properties: {
        snapshot_id: {
          type: "string",
          description: "ID of the snapshot to retrieve"
        }
      },
      required: ["snapshot_id"]
    }
  },
  
  // File Restores
  {
    name: "slide_get_file_restores",
    description: "Get a list of file restores",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "Filter file restores by agent ID"
        },
        limit: {
          type: "integer",
          description: "Maximum number of file restores to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_file_restore",
    description: "Get details of a specific file restore by ID",
    input_schema: {
      type: "object",
      properties: {
        file_restore_id: {
          type: "string",
          description: "ID of the file restore to retrieve"
        }
      },
      required: ["file_restore_id"]
    }
  },
  {
    name: "slide_create_file_restore",
    description: "Create a new file restore",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "ID of the agent to restore files from"
        },
        snapshot_id: {
          type: "string",
          description: "ID of the snapshot to restore from"
        },
        paths: {
          type: "array",
          description: "Paths to restore",
          items: {
            type: "string"
          }
        }
      },
      required: ["agent_id", "snapshot_id", "paths"]
    }
  },
  {
    name: "slide_delete_file_restore",
    description: "Delete a file restore",
    input_schema: {
      type: "object",
      properties: {
        file_restore_id: {
          type: "string",
          description: "ID of the file restore to delete"
        }
      },
      required: ["file_restore_id"]
    }
  },
  {
    name: "slide_browse_file_restore",
    description: "Browse files in a file restore",
    input_schema: {
      type: "object",
      properties: {
        file_restore_id: {
          type: "string",
          description: "ID of the file restore to browse"
        },
        path: {
          type: "string",
          description: "Path to browse within the restore"
        }
      },
      required: ["file_restore_id"]
    }
  },
  
  // Image Exports
  {
    name: "slide_get_image_exports",
    description: "Get a list of image exports",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "Filter image exports by agent ID"
        },
        limit: {
          type: "integer",
          description: "Maximum number of image exports to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_image_export",
    description: "Get details of a specific image export by ID",
    input_schema: {
      type: "object",
      properties: {
        image_export_id: {
          type: "string",
          description: "ID of the image export to retrieve"
        }
      },
      required: ["image_export_id"]
    }
  },
  {
    name: "slide_create_image_export",
    description: "Create a new image export",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "ID of the agent to export from"
        },
        snapshot_id: {
          type: "string",
          description: "ID of the snapshot to export"
        },
        format: {
          type: "string",
          description: "Format of the image export (e.g., 'vmdk', 'vhd')"
        }
      },
      required: ["agent_id", "snapshot_id", "format"]
    }
  },
  {
    name: "slide_delete_image_export",
    description: "Delete an image export",
    input_schema: {
      type: "object",
      properties: {
        image_export_id: {
          type: "string",
          description: "ID of the image export to delete"
        }
      },
      required: ["image_export_id"]
    }
  },
  {
    name: "slide_browse_image_export",
    description: "Browse an image export",
    input_schema: {
      type: "object",
      properties: {
        image_export_id: {
          type: "string",
          description: "ID of the image export to browse"
        },
        path: {
          type: "string",
          description: "Path to browse within the export"
        }
      },
      required: ["image_export_id"]
    }
  },
  
  // Virtual Machines
  {
    name: "slide_get_virtual_machines",
    description: "Get a list of virtual machines",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "Filter virtual machines by agent ID"
        },
        limit: {
          type: "integer",
          description: "Maximum number of virtual machines to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_virtual_machine",
    description: "Get details of a specific virtual machine by ID",
    input_schema: {
      type: "object",
      properties: {
        virt_id: {
          type: "string",
          description: "ID of the virtual machine to retrieve"
        }
      },
      required: ["virt_id"]
    }
  },
  {
    name: "slide_create_virtual_machine",
    description: "Create a new virtual machine",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "ID of the agent to create VM from"
        },
        snapshot_id: {
          type: "string",
          description: "ID of the snapshot to use"
        },
        name: {
          type: "string",
          description: "Name for the virtual machine"
        },
        network_id: {
          type: "string",
          description: "ID of the network to attach to the VM"
        }
      },
      required: ["agent_id", "snapshot_id", "name"]
    }
  },
  {
    name: "slide_update_virtual_machine",
    description: "Update a virtual machine",
    input_schema: {
      type: "object",
      properties: {
        virt_id: {
          type: "string",
          description: "ID of the virtual machine to update"
        },
        data: {
          type: "object",
          description: "Updated VM data"
        }
      },
      required: ["virt_id", "data"]
    }
  },
  {
    name: "slide_delete_virtual_machine",
    description: "Delete a virtual machine",
    input_schema: {
      type: "object",
      properties: {
        virt_id: {
          type: "string",
          description: "ID of the virtual machine to delete"
        }
      },
      required: ["virt_id"]
    }
  },
  
  // Networks
  {
    name: "slide_get_networks",
    description: "Get a list of networks",
    input_schema: {
      type: "object",
      properties: {
        limit: {
          type: "integer",
          description: "Maximum number of networks to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_network",
    description: "Get details of a specific network by ID",
    input_schema: {
      type: "object",
      properties: {
        network_id: {
          type: "string",
          description: "ID of the network to retrieve"
        }
      },
      required: ["network_id"]
    }
  },
  {
    name: "slide_create_network",
    description: "Create a new network",
    input_schema: {
      type: "object",
      properties: {
        name: {
          type: "string",
          description: "Name for the network"
        },
        subnet: {
          type: "string",
          description: "Subnet CIDR for the network"
        },
        data: {
          type: "object",
          description: "Additional network configuration"
        }
      },
      required: ["name", "subnet"]
    }
  },
  {
    name: "slide_update_network",
    description: "Update a network",
    input_schema: {
      type: "object",
      properties: {
        network_id: {
          type: "string",
          description: "ID of the network to update"
        },
        data: {
          type: "object",
          description: "Updated network data"
        }
      },
      required: ["network_id", "data"]
    }
  },
  {
    name: "slide_delete_network",
    description: "Delete a network",
    input_schema: {
      type: "object",
      properties: {
        network_id: {
          type: "string",
          description: "ID of the network to delete"
        }
      },
      required: ["network_id"]
    }
  },
  
  // Network Port Forwards
  {
    name: "slide_create_network_port_forward",
    description: "Create a new port forward for a network",
    input_schema: {
      type: "object",
      properties: {
        network_id: {
          type: "string",
          description: "ID of the network to add port forward to"
        },
        data: {
          type: "object",
          description: "Port forward configuration data"
        }
      },
      required: ["network_id", "data"]
    }
  },
  {
    name: "slide_delete_network_port_forward",
    description: "Delete a port forward from a network",
    input_schema: {
      type: "object",
      properties: {
        network_id: {
          type: "string",
          description: "ID of the network to remove port forward from"
        },
        data: {
          type: "object",
          description: "Port forward identification data"
        }
      },
      required: ["network_id", "data"]
    }
  },
  
  // Network WireGuard Peers
  {
    name: "slide_create_network_wg_peer",
    description: "Create a new WireGuard peer for a network",
    input_schema: {
      type: "object",
      properties: {
        network_id: {
          type: "string",
          description: "ID of the network to add WireGuard peer to"
        },
        data: {
          type: "object",
          description: "WireGuard peer configuration data"
        }
      },
      required: ["network_id", "data"]
    }
  },
  {
    name: "slide_update_network_wg_peer",
    description: "Update a WireGuard peer for a network",
    input_schema: {
      type: "object",
      properties: {
        network_id: {
          type: "string",
          description: "ID of the network with the WireGuard peer"
        },
        data: {
          type: "object",
          description: "Updated WireGuard peer data"
        }
      },
      required: ["network_id", "data"]
    }
  },
  {
    name: "slide_delete_network_wg_peer",
    description: "Delete a WireGuard peer from a network",
    input_schema: {
      type: "object",
      properties: {
        network_id: {
          type: "string",
          description: "ID of the network to remove WireGuard peer from"
        },
        data: {
          type: "object",
          description: "WireGuard peer identification data"
        }
      },
      required: ["network_id", "data"]
    }
  },
  
  // Users
  {
    name: "slide_get_users",
    description: "Get a list of users",
    input_schema: {
      type: "object",
      properties: {
        limit: {
          type: "integer",
          description: "Maximum number of users to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_user",
    description: "Get details of a specific user by ID",
    input_schema: {
      type: "object",
      properties: {
        user_id: {
          type: "string",
          description: "ID of the user to retrieve"
        }
      },
      required: ["user_id"]
    }
  },
  
  // Alerts
  {
    name: "slide_get_alerts",
    description: "Get a list of alerts",
    input_schema: {
      type: "object",
      properties: {
        agent_id: {
          type: "string",
          description: "Filter alerts by agent ID"
        },
        limit: {
          type: "integer",
          description: "Maximum number of alerts to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_alert",
    description: "Get details of a specific alert by ID",
    input_schema: {
      type: "object",
      properties: {
        alert_id: {
          type: "string",
          description: "ID of the alert to retrieve"
        }
      },
      required: ["alert_id"]
    }
  },
  {
    name: "slide_update_alert",
    description: "Update an alert",
    input_schema: {
      type: "object",
      properties: {
        alert_id: {
          type: "string",
          description: "ID of the alert to update"
        },
        data: {
          type: "object",
          description: "Updated alert data"
        }
      },
      required: ["alert_id", "data"]
    }
  },
  
  // Accounts
  {
    name: "slide_get_accounts",
    description: "Get a list of accounts",
    input_schema: {
      type: "object",
      properties: {
        limit: {
          type: "integer",
          description: "Maximum number of accounts to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "slide_get_account",
    description: "Get details of a specific account by ID",
    input_schema: {
      type: "object",
      properties: {
        account_id: {
          type: "string",
          description: "ID of the account to retrieve"
        }
      },
      required: ["account_id"]
    }
  },
  {
    name: "slide_update_account",
    description: "Update an account",
    input_schema: {
      type: "object",
      properties: {
        account_id: {
          type: "string",
          description: "ID of the account to update"
        },
        data: {
          type: "object",
          description: "Updated account data"
        }
      },
      required: ["account_id", "data"]
    }
  }
];

// Tool execution handler - this would be connected to Claude's tool calling
async function handleToolCall(toolName, args) {
  console.log(`Claude called tool: ${toolName} with args:`, args);
  
  switch(toolName) {
    // Devices
    case "slide_get_devices":
      return await slideClient.getDevices(args);
    case "slide_get_device":
      return await slideClient.getDevice(args.device_id);
    case "slide_update_device":
      return await slideClient.updateDevice(args.device_id, args.data);
    
    // Agents
    case "slide_get_agents":
      return await slideClient.getAgents(args);
    case "slide_get_agent":
      return await slideClient.getAgent(args.agent_id);
    case "slide_update_agent":
      return await slideClient.updateAgent(args.agent_id, args.data);
    case "slide_create_agent_pair":
      return await slideClient.createAgentPair(args.data);
    case "slide_pair_agent":
      return await slideClient.pairAgent(args.data);
    
    // Backups
    case "slide_get_backups":
      return await slideClient.getBackups(args);
    case "slide_get_backup":
      return await slideClient.getBackup(args.backup_id);
    case "slide_start_backup":
      return await slideClient.startBackup(args);
    
    // Snapshots
    case "slide_get_snapshots":
      return await slideClient.getSnapshots(args);
    case "slide_get_snapshot":
      return await slideClient.getSnapshot(args.snapshot_id);
    
    // File Restores
    case "slide_get_file_restores":
      return await slideClient.getFileRestores(args);
    case "slide_get_file_restore":
      return await slideClient.getFileRestore(args.file_restore_id);
    case "slide_create_file_restore":
      return await slideClient.createFileRestore(args);
    case "slide_delete_file_restore":
      return await slideClient.deleteFileRestore(args.file_restore_id);
    case "slide_browse_file_restore":
      return await slideClient.browseFileRestore(args.file_restore_id, args);
    
    // Image Exports
    case "slide_get_image_exports":
      return await slideClient.getImageExports(args);
    case "slide_get_image_export":
      return await slideClient.getImageExport(args.image_export_id);
    case "slide_create_image_export":
      return await slideClient.createImageExport(args);
    case "slide_delete_image_export":
      return await slideClient.deleteImageExport(args.image_export_id);
    case "slide_browse_image_export":
      return await slideClient.browseImageExport(args.image_export_id, args);
    
    // Virtual Machines
    case "slide_get_virtual_machines":
      return await slideClient.getVirtualMachines(args);
    case "slide_get_virtual_machine":
      return await slideClient.getVirtualMachine(args.virt_id);
    case "slide_create_virtual_machine":
      return await slideClient.createVirtualMachine(args);
    case "slide_update_virtual_machine":
      return await slideClient.updateVirtualMachine(args.virt_id, args.data);
    case "slide_delete_virtual_machine":
      return await slideClient.deleteVirtualMachine(args.virt_id);
    
    // Networks
    case "slide_get_networks":
      return await slideClient.getNetworks(args);
    case "slide_get_network":
      return await slideClient.getNetwork(args.network_id);
    case "slide_create_network":
      return await slideClient.createNetwork(args);
    case "slide_update_network":
      return await slideClient.updateNetwork(args.network_id, args.data);
    case "slide_delete_network":
      return await slideClient.deleteNetwork(args.network_id);
    
    // Network Port Forwards
    case "slide_create_network_port_forward":
      return await slideClient.createNetworkPortForward(args.network_id, args.data);
    case "slide_delete_network_port_forward":
      return await slideClient.deleteNetworkPortForward(args.network_id, args.data);
    
    // Network WireGuard Peers
    case "slide_create_network_wg_peer":
      return await slideClient.createNetworkWGPeer(args.network_id, args.data);
    case "slide_update_network_wg_peer":
      return await slideClient.updateNetworkWGPeer(args.network_id, args.data);
    case "slide_delete_network_wg_peer":
      return await slideClient.deleteNetworkWGPeer(args.network_id, args.data);
    
    // Users
    case "slide_get_users":
      return await slideClient.getUsers(args);
    case "slide_get_user":
      return await slideClient.getUser(args.user_id);
    
    // Alerts
    case "slide_get_alerts":
      return await slideClient.getAlerts(args);
    case "slide_get_alert":
      return await slideClient.getAlert(args.alert_id);
    case "slide_update_alert":
      return await slideClient.updateAlert(args.alert_id, args.data);
    
    // Accounts
    case "slide_get_accounts":
      return await slideClient.getAccounts(args);
    case "slide_get_account":
      return await slideClient.getAccount(args.account_id);
    case "slide_update_account":
      return await slideClient.updateAccount(args.account_id, args.data);
    
    default:
      throw new Error(`Unknown tool: ${toolName}`);
  }
}

// Demonstration of how tool execution would work
// In a real implementation, this would be triggered by Claude's tool calls
async function demonstrateToolUsage() {
  console.log("=== Claude Desktop Integration Demo ===\n");
  
  // Simulate Claude calling the slide_get_devices tool
  console.log("1. Simulating Claude calling slide_get_devices tool");
  const devicesResult = await handleToolCall("slide_get_devices", { limit: 5 });
  console.log("Result:", JSON.stringify(devicesResult, null, 2));
  console.log("\n---\n");
  
  // If devices were found, simulate getting agents for the first device
  if (devicesResult.data.length > 0) {
    const deviceId = devicesResult.data[0].device_id;
    
    console.log(`2. Simulating Claude calling slide_get_agents tool for device ${deviceId}`);
    const agentsResult = await handleToolCall("slide_get_agents", { device_id: deviceId, limit: 3 });
    console.log("Result:", JSON.stringify(agentsResult, null, 2));
  }
  
  console.log("\n=== Demo completed ===");
  console.log("\nNOTE: In an actual Claude Desktop integration:");
  console.log("1. You would provide these tools to Claude via its tool calling interface");
  console.log("2. Claude would analyze user requests and call the appropriate tools");
  console.log("3. Your application would handle the tool calls and return results to Claude");
  console.log("4. Claude would incorporate the results into its response to the user");
}

// Run the demonstration
demonstrateToolUsage(); 
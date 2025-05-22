#!/usr/bin/env node

import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema
} from "@modelcontextprotocol/sdk/types.js";
import axios, { AxiosError } from 'axios';
import dotenv from 'dotenv';

dotenv.config();

const SLIDE_API_KEY = process.env.SLIDE_API_KEY;
const API_BASE_URL = 'https://api.slide.tech';

if (!SLIDE_API_KEY) {
  console.error('Error: SLIDE_API_KEY environment variable not set');
  process.exit(1);
}

interface Device {
  device_id: string;
  display_name: string;
  last_seen_at: string;
  hostname: string;
  ip_addresses: string[];
  addresses: any[];
  public_ip_address: string;
  image_version: string;
  package_version: string;
  storage_used_bytes: number;
  storage_total_bytes: number;
  serial_number: string;
  hardware_model_name: string;
  service_model_name: string;
  service_model_name_short: string;
  service_status: string;
  nfr: boolean;
  client_id?: string;
  booted_at?: string;
}

interface PaginatedResponse<T> {
  pagination: {
    total: number;
    next_offset?: number;
  };
  data: T[];
}

// Additional Slide API interfaces
interface Agent {
  agent_id: string;
  device_id: string;
  display_name: string;
  last_seen_at: string;
  hostname: string;
  ip_addresses: string[];
  addresses: any[];
  public_ip_address: string;
  agent_version: string;
  platform: string;
  os: string;
  os_version: string;
  firmware_type: string;
  booted_at?: string;
  client_id?: string;
  encryption_algorithm?: string;
  manufacturer?: string;
}

interface AgentPairCode {
  agent_id: string;
  display_name: string;
  pair_code: string;
}

interface Backup {
  backup_id: string;
  agent_id: string;
  started_at: string;
  status: string;
  ended_at?: string;
  error_code?: number;
  error_message?: string;
  snapshot_id?: string;
}

interface Snapshot {
  snapshot_id: string;
  agent_id: string;
  locations: Location[];
  backup_started_at: string;
  backup_ended_at: string;
  deleted?: string;
  deletions?: Deletion[];
  verify_boot_status?: string;
  verify_fs_status?: string;
  verify_boot_screenshot_url?: string;
}

interface Location {
  type: string;
  device_id: string;
}

interface Deletion {
  type: string;
  deleted: string;
  deleted_by: string;
  first_and_last_name?: string;
}

interface Client {
  client_id: string;
  name: string;
  comments: string;
}

interface Alert {
  alert_id: string;
  alert_type: string;
  alert_fields: string;
  created_at: string;
  resolved: boolean;
  resolved_at?: string;
  resolved_by?: string;
  device_id?: string;
  agent_id?: string;
}

interface FileRestore {
  file_restore_id: string;
  device_id: string;
  agent_id: string;
  snapshot_id: string;
  created_at: string;
  expires_at?: string;
}

interface FileRestoreEntry {
  name: string;
  path: string;
  size: number;
  type: string;
  modified_at: string;
  download_uris: DownloadURI[];
  symlink_target_path?: string;
}

interface DownloadURI {
  type: string;
  uri: string;
}

interface ImageExport {
  image_export_id: string;
  device_id: string;
  agent_id: string;
  snapshot_id: string;
  image_type: string;
  created_at: string;
}

interface ImageExportEntry {
  disk_id: string;
  name: string;
  size: number;
  download_uris: DownloadURI[];
}

interface VirtualMachine {
  virt_id: string;
  device_id: string;
  agent_id: string;
  snapshot_id: string;
  state: string;
  created_at: string;
  cpu_count: number | null;
  memory_in_mb: number | null;
  disk_bus: string;
  network_model: string;
  network_type?: string;
  network_source?: string;
  vnc: VNC[];
  vnc_password: string;
  expires_at?: string;
}

interface VNC {
  type: string;
  host?: string;
  port?: number;
  websocket_uri?: string;
}

interface Network {
  network_id: string;
  type: string;
  name: string;
  comments: string;
  bridge_device_id: string;
  router_prefix: string;
  dhcp: boolean;
  dhcp_range_start: string;
  dhcp_range_end: string;
  nameservers: string;
  internet: boolean;
  connected_virt_ids: string[];
  client_id?: string;
}

interface NetworkPortForward {
  network_id: string;
  proto: string;
  port: number;
  dest: string;
}

interface NetworkWGPeer {
  network_id: string;
  peer_name: string;
  wg_public_key: string;
  wg_private_key: string;
  wg_address: string;
  remote_networks: string[];
}

interface User {
  user_id: string;
  first_name: string;
  last_name: string;
  display_name: string;
  email: string;
  role_id: string;
}

interface Account {
  account_id: string;
  account_name: string;
  primary_contact: string;
  primary_email: string;
  primary_phone: string;
  billing_address: PostalAddress;
  alert_emails: string[];
}

interface PostalAddress {
  Line1: string;
  Line2?: string;
  City: string;
  State: string;
  PostalCode: string;
  Country: string;
}

// Create MCP server
const server = new Server(
  {
    name: "slide-mcp-server",
    version: "0.1.0",
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Define the slide_list_devices tool
const LIST_DEVICES_TOOL = {
  name: "slide_list_devices",
  description: "List all devices with pagination and filtering options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      client_id: {
        type: "string",
        description: "Filter by client ID"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

// Device tools
const GET_DEVICE_TOOL = {
  name: "slide_get_device",
  description: "Get a specific device by ID",
  inputSchema: {
    type: "object",
    properties: {
      device_id: {
        type: "string",
        description: "ID of the device to retrieve"
      }
    },
    required: ["device_id"]
  }
};

const UPDATE_DEVICE_TOOL = {
  name: "slide_update_device",
  description: "Update device properties",
  inputSchema: {
    type: "object",
    properties: {
      device_id: {
        type: "string",
        description: "ID of the device to update"
      },
      display_name: {
        type: "string",
        description: "New display name for the device"
      },
      hostname: {
        type: "string",
        description: "New hostname for the device"
      },
      client_id: {
        type: "string",
        description: "Client ID to associate with the device (or empty string to remove)"
      }
    },
    required: ["device_id"]
  }
};

// Agent tools
const LIST_AGENTS_TOOL = {
  name: "slide_list_agents",
  description: "List all agents with pagination and filtering options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      device_id: {
        type: "string",
        description: "Filter by device ID"
      },
      client_id: {
        type: "string",
        description: "Filter by client ID"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_AGENT_TOOL = {
  name: "slide_get_agent",
  description: "Get a specific agent by ID",
  inputSchema: {
    type: "object",
    properties: {
      agent_id: {
        type: "string",
        description: "ID of the agent to retrieve"
      }
    },
    required: ["agent_id"]
  }
};

const CREATE_AGENT_TOOL = {
  name: "slide_create_agent",
  description: "Create a new agent for auto-pair installation",
  inputSchema: {
    type: "object",
    properties: {
      device_id: {
        type: "string",
        description: "Device ID to associate with the agent"
      },
      display_name: {
        type: "string",
        description: "Display name for the agent"
      }
    },
    required: ["device_id", "display_name"]
  }
};

const PAIR_AGENT_TOOL = {
  name: "slide_pair_agent",
  description: "Pair an agent with a device",
  inputSchema: {
    type: "object",
    properties: {
      device_id: {
        type: "string",
        description: "Device ID to pair with"
      },
      pair_code: {
        type: "string",
        description: "Pair code from the agent"
      }
    },
    required: ["device_id", "pair_code"]
  }
};

const UPDATE_AGENT_TOOL = {
  name: "slide_update_agent",
  description: "Update agent properties",
  inputSchema: {
    type: "object",
    properties: {
      agent_id: {
        type: "string",
        description: "ID of the agent to update"
      },
      display_name: {
        type: "string",
        description: "New display name for the agent"
      }
    },
    required: ["agent_id", "display_name"]
  }
};

// Backup tools
const LIST_BACKUPS_TOOL = {
  name: "slide_list_backups",
  description: "List backups with pagination and filtering options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      agent_id: {
        type: "string",
        description: "Filter by agent ID"
      },
      device_id: {
        type: "string",
        description: "Filter by device ID"
      },
      snapshot_id: {
        type: "string",
        description: "Filter by snapshot ID"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_BACKUP_TOOL = {
  name: "slide_get_backup",
  description: "Get a specific backup by ID",
  inputSchema: {
    type: "object",
    properties: {
      backup_id: {
        type: "string",
        description: "ID of the backup to retrieve"
      }
    },
    required: ["backup_id"]
  }
};

const START_BACKUP_TOOL = {
  name: "slide_start_backup",
  description: "Start a new backup for an agent",
  inputSchema: {
    type: "object",
    properties: {
      agent_id: {
        type: "string",
        description: "ID of the agent to backup"
      }
    },
    required: ["agent_id"]
  }
};

// Snapshot tools
const LIST_SNAPSHOTS_TOOL = {
  name: "slide_list_snapshots",
  description: "List snapshots with pagination and filtering options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      agent_id: {
        type: "string",
        description: "Filter by agent ID"
      },
      snapshot_location: {
        type: "string",
        description: "Filter by location or deleted status"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_SNAPSHOT_TOOL = {
  name: "slide_get_snapshot",
  description: "Get a specific snapshot by ID",
  inputSchema: {
    type: "object",
    properties: {
      snapshot_id: {
        type: "string",
        description: "ID of the snapshot to retrieve"
      }
    },
    required: ["snapshot_id"]
  }
};

// Client tools
const LIST_CLIENTS_TOOL = {
  name: "slide_list_clients",
  description: "List clients with pagination and sorting options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_CLIENT_TOOL = {
  name: "slide_get_client",
  description: "Get a specific client by ID",
  inputSchema: {
    type: "object",
    properties: {
      client_id: {
        type: "string",
        description: "ID of the client to retrieve"
      }
    },
    required: ["client_id"]
  }
};

const CREATE_CLIENT_TOOL = {
  name: "slide_create_client",
  description: "Create a new client",
  inputSchema: {
    type: "object",
    properties: {
      name: {
        type: "string",
        description: "Name of the client"
      },
      comments: {
        type: "string",
        description: "Comments about the client"
      }
    },
    required: ["name"]
  }
};

const UPDATE_CLIENT_TOOL = {
  name: "slide_update_client",
  description: "Update client properties",
  inputSchema: {
    type: "object",
    properties: {
      client_id: {
        type: "string",
        description: "ID of the client to update"
      },
      name: {
        type: "string",
        description: "New name for the client"
      },
      comments: {
        type: "string",
        description: "New comments about the client"
      }
    },
    required: ["client_id"]
  }
};

const DELETE_CLIENT_TOOL = {
  name: "slide_delete_client",
  description: "Delete a client",
  inputSchema: {
    type: "object",
    properties: {
      client_id: {
        type: "string",
        description: "ID of the client to delete"
      }
    },
    required: ["client_id"]
  }
};

// File Restore tools
const LIST_FILE_RESTORES_TOOL = {
  name: "slide_list_file_restores",
  description: "List file restores with pagination and sorting options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_FILE_RESTORE_TOOL = {
  name: "slide_get_file_restore",
  description: "Get a specific file restore by ID",
  inputSchema: {
    type: "object",
    properties: {
      file_restore_id: {
        type: "string",
        description: "ID of the file restore to retrieve"
      }
    },
    required: ["file_restore_id"]
  }
};

const CREATE_FILE_RESTORE_TOOL = {
  name: "slide_create_file_restore",
  description: "Create a new file restore from a snapshot",
  inputSchema: {
    type: "object",
    properties: {
      snapshot_id: {
        type: "string",
        description: "ID of the snapshot to restore from"
      },
      device_id: {
        type: "string",
        description: "ID of the device to restore to"
      }
    },
    required: ["snapshot_id", "device_id"]
  }
};

const DELETE_FILE_RESTORE_TOOL = {
  name: "slide_delete_file_restore",
  description: "Delete a file restore",
  inputSchema: {
    type: "object",
    properties: {
      file_restore_id: {
        type: "string",
        description: "ID of the file restore to delete"
      }
    },
    required: ["file_restore_id"]
  }
};

const BROWSE_FILE_RESTORE_TOOL = {
  name: "slide_browse_file_restore",
  description: "Browse the files in a file restore",
  inputSchema: {
    type: "object",
    properties: {
      file_restore_id: {
        type: "string",
        description: "ID of the file restore to browse"
      },
      path: {
        type: "string",
        description: "Path to browse within the restore"
      },
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      }
    },
    required: ["file_restore_id", "path"]
  }
};

// Image Export tools
const LIST_IMAGE_EXPORTS_TOOL = {
  name: "slide_list_image_exports",
  description: "List image exports with pagination and sorting options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_IMAGE_EXPORT_TOOL = {
  name: "slide_get_image_export",
  description: "Get a specific image export by ID",
  inputSchema: {
    type: "object",
    properties: {
      image_export_id: {
        type: "string",
        description: "ID of the image export to retrieve"
      }
    },
    required: ["image_export_id"]
  }
};

const CREATE_IMAGE_EXPORT_TOOL = {
  name: "slide_create_image_export",
  description: "Create a new image export from a snapshot",
  inputSchema: {
    type: "object",
    properties: {
      snapshot_id: {
        type: "string",
        description: "ID of the snapshot to export"
      },
      device_id: {
        type: "string",
        description: "ID of the device to export to"
      },
      image_type: {
        type: "string",
        description: "Type of image to export (vhdx, vhdx-dynamic, vhd, raw)"
      },
      boot_mods: {
        type: "array",
        description: "Optional boot mods to enable on the export"
      }
    },
    required: ["snapshot_id", "device_id", "image_type"]
  }
};

const DELETE_IMAGE_EXPORT_TOOL = {
  name: "slide_delete_image_export",
  description: "Delete an image export",
  inputSchema: {
    type: "object",
    properties: {
      image_export_id: {
        type: "string",
        description: "ID of the image export to delete"
      }
    },
    required: ["image_export_id"]
  }
};

const BROWSE_IMAGE_EXPORT_TOOL = {
  name: "slide_browse_image_export",
  description: "Browse the disk images in an image export",
  inputSchema: {
    type: "object",
    properties: {
      image_export_id: {
        type: "string",
        description: "ID of the image export to browse"
      },
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      }
    },
    required: ["image_export_id"]
  }
};

// Virtual Machine tools
const LIST_VIRTUAL_MACHINES_TOOL = {
  name: "slide_list_virtual_machines",
  description: "List virtual machines with pagination and sorting options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_VIRTUAL_MACHINE_TOOL = {
  name: "slide_get_virtual_machine",
  description: "Get a specific virtual machine by ID",
  inputSchema: {
    type: "object",
    properties: {
      virt_id: {
        type: "string",
        description: "ID of the virtual machine to retrieve"
      }
    },
    required: ["virt_id"]
  }
};

const CREATE_VIRTUAL_MACHINE_TOOL = {
  name: "slide_create_virtual_machine",
  description: "Create a new virtual machine from a snapshot",
  inputSchema: {
    type: "object",
    properties: {
      snapshot_id: {
        type: "string",
        description: "ID of the snapshot to virtualize"
      },
      device_id: {
        type: "string",
        description: "ID of the device to create the VM on"
      },
      cpu_count: {
        type: "number",
        description: "Number of CPUs to allocate"
      },
      memory_in_mb: {
        type: "number",
        description: "Amount of memory in MB to allocate"
      },
      disk_bus: {
        type: "string",
        description: "Disk bus type (sata, virtio)"
      },
      network_type: {
        type: "string",
        description: "Network type for the VM"
      },
      network_source: {
        type: "string",
        description: "Network ID for the VM when using network-id type"
      },
      boot_mods: {
        type: "array",
        description: "Optional boot mods to enable on the VM"
      }
    },
    required: ["snapshot_id", "device_id"]
  }
};

const UPDATE_VIRTUAL_MACHINE_TOOL = {
  name: "slide_update_virtual_machine",
  description: "Update virtual machine properties or state",
  inputSchema: {
    type: "object",
    properties: {
      virt_id: {
        type: "string",
        description: "ID of the virtual machine to update"
      },
      state: {
        type: "string",
        description: "New state for the VM (running, stopped, paused)"
      },
      expires_at: {
        type: "string",
        description: "New expiration time for the VM"
      },
      memory_in_mb: {
        type: "number",
        description: "New memory allocation in MB"
      },
      cpu_count: {
        type: "number",
        description: "New CPU count"
      }
    },
    required: ["virt_id"]
  }
};

const DELETE_VIRTUAL_MACHINE_TOOL = {
  name: "slide_delete_virtual_machine",
  description: "Delete a virtual machine",
  inputSchema: {
    type: "object",
    properties: {
      virt_id: {
        type: "string",
        description: "ID of the virtual machine to delete"
      }
    },
    required: ["virt_id"]
  }
};

// Network tools
const LIST_NETWORKS_TOOL = {
  name: "slide_list_networks",
  description: "List networks with pagination and sorting options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_NETWORK_TOOL = {
  name: "slide_get_network",
  description: "Get a specific network by ID",
  inputSchema: {
    type: "object",
    properties: {
      network_id: {
        type: "string",
        description: "ID of the network to retrieve"
      }
    },
    required: ["network_id"]
  }
};

const CREATE_NETWORK_TOOL = {
  name: "slide_create_network",
  description: "Create a new network",
  inputSchema: {
    type: "object",
    properties: {
      name: {
        type: "string",
        description: "Name of the network"
      },
      type: {
        type: "string",
        description: "Type of network (standard, bridge-lan)"
      },
      client_id: {
        type: "string",
        description: "Client ID to associate with the network"
      },
      bridge_device_id: {
        type: "string",
        description: "Device ID for bridge network types"
      },
      router_prefix: {
        type: "string",
        description: "Router IP and netmask for standard networks"
      },
      dhcp: {
        type: "boolean",
        description: "Whether DHCP should be enabled (standard networks)"
      },
      dhcp_range_start: {
        type: "string",
        description: "DHCP range start for standard networks"
      },
      dhcp_range_end: {
        type: "string",
        description: "DHCP range end for standard networks"
      },
      nameservers: {
        type: "string",
        description: "Comma-separated DNS servers for standard networks"
      },
      internet: {
        type: "boolean",
        description: "Whether internet access should be enabled"
      }
    },
    required: ["name", "type"]
  }
};

const UPDATE_NETWORK_TOOL = {
  name: "slide_update_network",
  description: "Update network properties",
  inputSchema: {
    type: "object",
    properties: {
      network_id: {
        type: "string",
        description: "ID of the network to update"
      },
      name: {
        type: "string",
        description: "New name for the network"
      },
      comments: {
        type: "string",
        description: "New comments about the network"
      },
      dhcp: {
        type: "boolean",
        description: "Whether DHCP should be enabled"
      },
      dhcp_range_start: {
        type: "string",
        description: "New DHCP range start"
      },
      dhcp_range_end: {
        type: "string",
        description: "New DHCP range end"
      },
      nameservers: {
        type: "string",
        description: "New comma-separated DNS servers"
      },
      router_prefix: {
        type: "string",
        description: "New router IP and netmask"
      },
      internet: {
        type: "boolean",
        description: "Whether internet access should be enabled"
      }
    },
    required: ["network_id"]
  }
};

const DELETE_NETWORK_TOOL = {
  name: "slide_delete_network",
  description: "Delete a network",
  inputSchema: {
    type: "object",
    properties: {
      network_id: {
        type: "string",
        description: "ID of the network to delete"
      }
    },
    required: ["network_id"]
  }
};

// Alert tools
const LIST_ALERTS_TOOL = {
  name: "slide_list_alerts",
  description: "List alerts with pagination and filtering options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      device_id: {
        type: "string",
        description: "Filter by device ID"
      },
      agent_id: {
        type: "string",
        description: "Filter by agent ID"
      },
      resolved: {
        type: "boolean",
        description: "Filter by resolved status"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_ALERT_TOOL = {
  name: "slide_get_alert",
  description: "Get a specific alert by ID",
  inputSchema: {
    type: "object",
    properties: {
      alert_id: {
        type: "string",
        description: "ID of the alert to retrieve"
      }
    },
    required: ["alert_id"]
  }
};

const UPDATE_ALERT_TOOL = {
  name: "slide_update_alert",
  description: "Update alert properties",
  inputSchema: {
    type: "object",
    properties: {
      alert_id: {
        type: "string",
        description: "ID of the alert to update"
      },
      resolved: {
        type: "boolean",
        description: "Whether the alert is resolved"
      }
    },
    required: ["alert_id", "resolved"]
  }
};

// User tools
const LIST_USERS_TOOL = {
  name: "slide_list_users",
  description: "List users with pagination and sorting options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_USER_TOOL = {
  name: "slide_get_user",
  description: "Get a specific user by ID",
  inputSchema: {
    type: "object",
    properties: {
      user_id: {
        type: "string",
        description: "ID of the user to retrieve"
      }
    },
    required: ["user_id"]
  }
};

// Account tools
const LIST_ACCOUNTS_TOOL = {
  name: "slide_list_accounts",
  description: "List accounts with pagination and sorting options",
  inputSchema: {
    type: "object",
    properties: {
      limit: {
        type: "number",
        description: "Number of results per page (max 50)"
      },
      offset: {
        type: "number",
        description: "Pagination offset"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      }
    }
  }
};

const GET_ACCOUNT_TOOL = {
  name: "slide_get_account",
  description: "Get a specific account by ID",
  inputSchema: {
    type: "object",
    properties: {
      account_id: {
        type: "string",
        description: "ID of the account to retrieve"
      }
    },
    required: ["account_id"]
  }
};

const UPDATE_ACCOUNT_TOOL = {
  name: "slide_update_account",
  description: "Update account properties",
  inputSchema: {
    type: "object",
    properties: {
      account_id: {
        type: "string",
        description: "ID of the account to update"
      },
      alert_emails: {
        type: "array",
        description: "New alert email list for the account"
      }
    },
    required: ["account_id"]
  }
};

// Function to check if args are valid for the list_devices tool
function isListDevicesArgs(args: unknown): args is { 
  limit?: number; 
  offset?: number;
  client_id?: string;
  sort_asc?: boolean;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Type checking functions for device tools
function isGetDeviceArgs(args: unknown): args is {
  device_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "device_id" in args &&
    typeof (args as any).device_id === "string"
  );
}

function isUpdateDeviceArgs(args: unknown): args is {
  device_id: string;
  display_name?: string;
  hostname?: string;
  client_id?: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "device_id" in args &&
    typeof (args as any).device_id === "string"
  );
}

// Type checking functions for agent tools
function isListAgentsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  device_id?: string;
  client_id?: string;
  sort_asc?: boolean;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

function isGetAgentArgs(args: unknown): args is {
  agent_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "agent_id" in args &&
    typeof (args as any).agent_id === "string"
  );
}

function isCreateAgentArgs(args: unknown): args is {
  device_id: string;
  display_name: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "device_id" in args &&
    typeof (args as any).device_id === "string" &&
    "display_name" in args &&
    typeof (args as any).display_name === "string"
  );
}

function isPairAgentArgs(args: unknown): args is {
  device_id: string;
  pair_code: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "device_id" in args &&
    typeof (args as any).device_id === "string" &&
    "pair_code" in args &&
    typeof (args as any).pair_code === "string"
  );
}

function isUpdateAgentArgs(args: unknown): args is {
  agent_id: string;
  display_name?: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "agent_id" in args &&
    typeof (args as any).agent_id === "string"
  );
}

// Type checking functions for backup tools
function isListBackupsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  agent_id?: string;
  device_id?: string;
  snapshot_id?: string;
  sort_asc?: boolean;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

function isGetBackupArgs(args: unknown): args is {
  backup_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "backup_id" in args &&
    typeof (args as any).backup_id === "string"
  );
}

function isStartBackupArgs(args: unknown): args is {
  agent_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "agent_id" in args &&
    typeof (args as any).agent_id === "string"
  );
}

// Type checking functions for snapshot tools
function isListSnapshotsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  agent_id?: string;
  snapshot_location?: string;
  sort_asc?: boolean;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

function isGetSnapshotArgs(args: unknown): args is {
  snapshot_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "snapshot_id" in args &&
    typeof (args as any).snapshot_id === "string"
  );
}

// Type checking functions for client tools
function isListClientsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

function isGetClientArgs(args: unknown): args is {
  client_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "client_id" in args &&
    typeof (args as any).client_id === "string"
  );
}

function isCreateClientArgs(args: unknown): args is {
  name: string;
  comments?: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "name" in args &&
    typeof (args as any).name === "string"
  );
}

function isUpdateClientArgs(args: unknown): args is {
  client_id: string;
  name?: string;
  comments?: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "client_id" in args &&
    typeof (args as any).client_id === "string"
  );
}

function isDeleteClientArgs(args: unknown): args is {
  client_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "client_id" in args &&
    typeof (args as any).client_id === "string"
  );
}

// Type checking functions for file restore tools
function isListFileRestoresArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

function isGetFileRestoreArgs(args: unknown): args is {
  file_restore_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "file_restore_id" in args &&
    typeof (args as any).file_restore_id === "string"
  );
}

function isCreateFileRestoreArgs(args: unknown): args is {
  snapshot_id: string;
  device_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "snapshot_id" in args &&
    typeof (args as any).snapshot_id === "string" &&
    "device_id" in args &&
    typeof (args as any).device_id === "string"
  );
}

function isDeleteFileRestoreArgs(args: unknown): args is {
  file_restore_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "file_restore_id" in args &&
    typeof (args as any).file_restore_id === "string"
  );
}

function isBrowseFileRestoreArgs(args: unknown): args is {
  file_restore_id: string;
  path: string;
  limit?: number;
  offset?: number;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    "file_restore_id" in args &&
    typeof (args as any).file_restore_id === "string" &&
    "path" in args &&
    typeof (args as any).path === "string"
  );
}

// Function to list devices
async function listDevices(args: { 
  limit?: number; 
  offset?: number;
  client_id?: string;
  sort_asc?: boolean;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.client_id) {
      queryParams.append('client_id', args.client_id);
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    // Default sort by hostname
    queryParams.append('sort_by', 'hostname');
    
    const response = await axios.get<PaginatedResponse<Device>>(
      `${API_BASE_URL}/v1/device?${queryParams.toString()}`,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response) {
        const statusCode = axiosError.response.status;
        const errorData = axiosError.response.data as any;
        
        throw new Error(`API Error (${statusCode}): ${errorData.message || 'Unknown error'}`);
      }
      
      throw new Error(`Network Error: ${error.message}`);
    }
    
    throw new Error(`Error: ${(error as Error).message}`);
  }
}

// Function to get a device by ID
async function getDevice(args: { device_id: string }) {
  try {
    const response = await axios.get<Device>(
      `${API_BASE_URL}/v1/device/${args.device_id}`,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response) {
        const statusCode = axiosError.response.status;
        const errorData = axiosError.response.data as any;
        
        throw new Error(`API Error (${statusCode}): ${errorData.message || 'Unknown error'}`);
      }
      
      throw new Error(`Network Error: ${error.message}`);
    }
    
    throw new Error(`Error: ${(error as Error).message}`);
  }
}

// Function to update a device
async function updateDevice(args: {
  device_id: string;
  display_name?: string;
  hostname?: string;
  client_id?: string;
}) {
  try {
    const payload: any = {};
    
    if (args.display_name !== undefined) {
      payload.display_name = args.display_name;
    }
    
    if (args.hostname !== undefined) {
      payload.hostname = args.hostname;
    }
    
    if (args.client_id !== undefined) {
      payload.client_id = args.client_id;
    }
    
    const response = await axios.patch<Device>(
      `${API_BASE_URL}/v1/device/${args.device_id}`,
      payload,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response) {
        const statusCode = axiosError.response.status;
        const errorData = axiosError.response.data as any;
        
        throw new Error(`API Error (${statusCode}): ${errorData.message || 'Unknown error'}`);
      }
      
      throw new Error(`Network Error: ${error.message}`);
    }
    
    throw new Error(`Error: ${(error as Error).message}`);
  }
}

// Function to list agents
async function listAgents(args: {
  limit?: number;
  offset?: number;
  device_id?: string;
  client_id?: string;
  sort_asc?: boolean;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.device_id) {
      queryParams.append('device_id', args.device_id);
    }
    
    if (args.client_id) {
      queryParams.append('client_id', args.client_id);
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    // Default sort by hostname
    queryParams.append('sort_by', 'hostname');
    
    const response = await axios.get<PaginatedResponse<Agent>>(
      `${API_BASE_URL}/v1/agent?${queryParams.toString()}`,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response) {
        const statusCode = axiosError.response.status;
        const errorData = axiosError.response.data as any;
        
        throw new Error(`API Error (${statusCode}): ${errorData.message || 'Unknown error'}`);
      }
      
      throw new Error(`Network Error: ${error.message}`);
    }
    
    throw new Error(`Error: ${(error as Error).message}`);
  }
}

// Function to get an agent by ID
async function getAgent(args: { agent_id: string }) {
  try {
    const response = await axios.get<Agent>(
      `${API_BASE_URL}/v1/agent/${args.agent_id}`,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response) {
        const statusCode = axiosError.response.status;
        const errorData = axiosError.response.data as any;
        
        throw new Error(`API Error (${statusCode}): ${errorData.message || 'Unknown error'}`);
      }
      
      throw new Error(`Network Error: ${error.message}`);
    }
    
    throw new Error(`Error: ${(error as Error).message}`);
  }
}

// Function to create a new agent
async function createAgent(args: {
  device_id: string;
  display_name: string;
}) {
  try {
    const response = await axios.post<AgentPairCode>(
      `${API_BASE_URL}/v1/agent`,
      {
        device_id: args.device_id,
        display_name: args.display_name
      },
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response) {
        const statusCode = axiosError.response.status;
        const errorData = axiosError.response.data as any;
        
        throw new Error(`API Error (${statusCode}): ${errorData.message || 'Unknown error'}`);
      }
      
      throw new Error(`Network Error: ${error.message}`);
    }
    
    throw new Error(`Error: ${(error as Error).message}`);
  }
}

// Function to pair an agent
async function pairAgent(args: {
  device_id: string;
  pair_code: string;
}) {
  try {
    const response = await axios.post<Agent>(
      `${API_BASE_URL}/v1/agent/pair`,
      {
        device_id: args.device_id,
        pair_code: args.pair_code
      },
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response) {
        const statusCode = axiosError.response.status;
        const errorData = axiosError.response.data as any;
        
        throw new Error(`API Error (${statusCode}): ${errorData.message || 'Unknown error'}`);
      }
      
      throw new Error(`Network Error: ${error.message}`);
    }
    
    throw new Error(`Error: ${(error as Error).message}`);
  }
}

// Function to update an agent
async function updateAgent(args: {
  agent_id: string;
  display_name?: string;
}) {
  try {
    const payload: any = {};
    
    if (args.display_name !== undefined) {
      payload.display_name = args.display_name;
    }
    
    const response = await axios.patch<Agent>(
      `${API_BASE_URL}/v1/agent/${args.agent_id}`,
      payload,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response) {
        const statusCode = axiosError.response.status;
        const errorData = axiosError.response.data as any;
        
        throw new Error(`API Error (${statusCode}): ${errorData.message || 'Unknown error'}`);
      }
      
      throw new Error(`Network Error: ${error.message}`);
    }
    
    throw new Error(`Error: ${(error as Error).message}`);
  }
}

// Server implements the listTools request
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: [
    LIST_DEVICES_TOOL,
    GET_DEVICE_TOOL,
    UPDATE_DEVICE_TOOL,
    LIST_AGENTS_TOOL,
    GET_AGENT_TOOL,
    CREATE_AGENT_TOOL,
    PAIR_AGENT_TOOL,
    UPDATE_AGENT_TOOL,
    LIST_BACKUPS_TOOL,
    GET_BACKUP_TOOL,
    START_BACKUP_TOOL,
    LIST_SNAPSHOTS_TOOL,
    GET_SNAPSHOT_TOOL,
    LIST_CLIENTS_TOOL,
    GET_CLIENT_TOOL,
    CREATE_CLIENT_TOOL,
    UPDATE_CLIENT_TOOL,
    DELETE_CLIENT_TOOL,
    LIST_FILE_RESTORES_TOOL,
    GET_FILE_RESTORE_TOOL,
    CREATE_FILE_RESTORE_TOOL,
    DELETE_FILE_RESTORE_TOOL,
    BROWSE_FILE_RESTORE_TOOL,
    LIST_IMAGE_EXPORTS_TOOL,
    GET_IMAGE_EXPORT_TOOL,
    CREATE_IMAGE_EXPORT_TOOL,
    DELETE_IMAGE_EXPORT_TOOL,
    BROWSE_IMAGE_EXPORT_TOOL,
    LIST_VIRTUAL_MACHINES_TOOL,
    GET_VIRTUAL_MACHINE_TOOL,
    CREATE_VIRTUAL_MACHINE_TOOL,
    UPDATE_VIRTUAL_MACHINE_TOOL,
    DELETE_VIRTUAL_MACHINE_TOOL,
    LIST_NETWORKS_TOOL,
    GET_NETWORK_TOOL,
    CREATE_NETWORK_TOOL,
    UPDATE_NETWORK_TOOL,
    DELETE_NETWORK_TOOL,
    LIST_ALERTS_TOOL,
    GET_ALERT_TOOL,
    UPDATE_ALERT_TOOL,
    LIST_USERS_TOOL,
    GET_USER_TOOL,
    LIST_ACCOUNTS_TOOL,
    GET_ACCOUNT_TOOL,
    UPDATE_ACCOUNT_TOOL
  ],
}));

// Handle tool calls
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  try {
    const { name, arguments: args } = request.params;

    if (!args) {
      throw new Error("No arguments provided");
    }

    switch (name) {
      // DEVICE TOOLS
      case "slide_list_devices": {
        if (!isListDevicesArgs(args)) {
          throw new Error("Invalid arguments for slide_list_devices");
        }
        
        const result = await listDevices(args);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }
      
      case "slide_get_device": {
        if (!isGetDeviceArgs(args)) {
          throw new Error("Invalid arguments for slide_get_device");
        }
        
        const result = await getDevice(args);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }
      
      case "slide_update_device": {
        if (!isUpdateDeviceArgs(args)) {
          throw new Error("Invalid arguments for slide_update_device");
        }
        
        const result = await updateDevice(args);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }
      
      // AGENT TOOLS
      case "slide_list_agents": {
        if (!isListAgentsArgs(args)) {
          throw new Error("Invalid arguments for slide_list_agents");
        }
        
        const result = await listAgents(args);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }
      
      case "slide_get_agent": {
        if (!isGetAgentArgs(args)) {
          throw new Error("Invalid arguments for slide_get_agent");
        }
        
        const result = await getAgent(args);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }
      
      case "slide_create_agent": {
        if (!isCreateAgentArgs(args)) {
          throw new Error("Invalid arguments for slide_create_agent");
        }
        
        const result = await createAgent(args);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }
      
      case "slide_pair_agent": {
        if (!isPairAgentArgs(args)) {
          throw new Error("Invalid arguments for slide_pair_agent");
        }
        
        const result = await pairAgent(args);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }
      
      case "slide_update_agent": {
        if (!isUpdateAgentArgs(args)) {
          throw new Error("Invalid arguments for slide_update_agent");
        }
        
        const result = await updateAgent(args);
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }
      
      // Add placeholder for other tools - will implement their functions in future PRs
      case "slide_list_backups":
      case "slide_get_backup":
      case "slide_start_backup":
      case "slide_list_snapshots":
      case "slide_get_snapshot":
      case "slide_list_clients":
      case "slide_get_client":
      case "slide_create_client":
      case "slide_update_client":
      case "slide_delete_client":
      case "slide_list_file_restores":
      case "slide_get_file_restore":
      case "slide_create_file_restore":
      case "slide_delete_file_restore":
      case "slide_browse_file_restore":
      case "slide_list_image_exports":
      case "slide_get_image_export":
      case "slide_create_image_export":
      case "slide_delete_image_export":
      case "slide_browse_image_export":
      case "slide_list_virtual_machines":
      case "slide_get_virtual_machine":
      case "slide_create_virtual_machine":
      case "slide_update_virtual_machine":
      case "slide_delete_virtual_machine":
      case "slide_list_networks":
      case "slide_get_network":
      case "slide_create_network":
      case "slide_update_network":
      case "slide_delete_network":
      case "slide_list_alerts":
      case "slide_get_alert":
      case "slide_update_alert":
      case "slide_list_users":
      case "slide_get_user":
      case "slide_list_accounts":
      case "slide_get_account":
      case "slide_update_account": {
        return {
          content: [{ type: "text", text: `Tool ${name} is defined but not yet implemented. It will be available in a future release.` }],
          isError: true,
        };
      }

      default:
        return {
          content: [{ type: "text", text: `Unknown tool: ${name}` }],
          isError: true,
        };
    }
  } catch (error) {
    return {
      content: [
        {
          type: "text",
          text: `Error: ${error instanceof Error ? error.message : String(error)}`,
        },
      ],
      isError: true,
    };
  }
});

async function main() {
  try {
    const transport = new StdioServerTransport();
    await server.connect(transport);
    console.error("Slide MCP Server running on stdio");
  } catch (error) {
    console.error("Error starting server:", error);
    process.exit(1);
  }
}

main();

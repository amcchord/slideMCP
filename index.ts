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

// Helper function to generate VNC viewer URL
function generateVncViewerUrl(virtId: string, websocketUri: string, vncPassword: string): string {
  // URL encode the websocket URI and base64 encode the password
  const encodedWebsocketUri = encodeURIComponent(websocketUri);
  
  // Use Buffer.from for base64 encoding (Node.js environment)
  const base64Password = Buffer.from(vncPassword).toString('base64');
  
  // Return a browser-accessible VNC viewer URL for easy console access with base64-encoded password
  return `https://slide.recipes/mcpTools/vncViewer.php?id=${virtId}&ws=${encodedWebsocketUri}&password=${base64Password}&encoding=base64`;
}

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
  manufacturer?: string;
  client_id?: string;
  booted_at?: string;
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
  ended_at?: string;
  status: string;
  error_code?: number;
  error_message?: string;
  snapshot_id?: string;
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
  download_uris: {
    type: string;
    uri: string;
  }[];
  symlink_target_path?: string;
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
  download_uris: {
    type: string;
    uri: string;
  }[];
}

interface VirtualMachine {
  virt_id: string;
  device_id: string;
  agent_id: string;
  snapshot_id: string;
  state: string;
  created_at: string;
  expires_at?: string;
  cpu_count: number;
  memory_in_mb: number;
  disk_bus: string;
  network_model: string;
  network_type?: string;
  network_source?: string;
  vnc: {
    type: string;
    host?: string;
    port?: number;
    websocket_uri?: string;
  }[];
  vnc_password: string;
}

interface User {
  user_id: string;
  first_name: string;
  last_name: string;
  display_name: string;
  email: string;
  role_id: string;
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

interface Account {
  account_id: string;
  account_name: string;
  primary_contact: string;
  primary_email: string;
  primary_phone: string;
  billing_address: {
    Line1: string;
    Line2?: string;
    City: string;
    State: string;
    PostalCode: string;
    Country: string;
  };
  alert_emails: string[];
}

interface Client {
  client_id: string;
  name: string;
  comments: string;
}

interface PaginatedResponse<T> {
  pagination: {
    total: number;
    next_offset?: number;
  };
  data: T[];
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
  description: "List all devices with pagination and filtering options. Hostname is the primary identifier for devices and should be used when referring to devices in conversations with users. Although each device has a unique device_id, humans typically identify devices by their hostname.",
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

// Define the slide_list_agents tool
const LIST_AGENTS_TOOL = {
  name: "slide_list_agents",
  description: "List all agents with pagination and filtering options. Display Name is the primary identifier for agents that users recognize. If Display Name is blank, use hostname instead. Agent IDs are internal identifiers not commonly used by humans.",
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
      },
      sort_by: {
        type: "string",
        description: "Sort by field (id, hostname, name)"
      }
    }
  }
};

// Define the slide_get_agent tool
const GET_AGENT_TOOL = {
  name: "slide_get_agent",
  description: "Get detailed information about a specific agent by ID",
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

// Define the slide_create_agent tool
const CREATE_AGENT_TOOL = {
  name: "slide_create_agent",
  description: "Create an agent for auto-pair installation",
  inputSchema: {
    type: "object",
    properties: {
      display_name: {
        type: "string",
        description: "Display name for the agent"
      },
      device_id: {
        type: "string",
        description: "ID of the device to associate with the agent"
      }
    },
    required: ["display_name", "device_id"]
  }
};

// Define the slide_pair_agent tool
const PAIR_AGENT_TOOL = {
  name: "slide_pair_agent",
  description: "Pair an agent with a device using a pair code",
  inputSchema: {
    type: "object",
    properties: {
      pair_code: {
        type: "string",
        description: "Pair code generated during agent creation"
      },
      device_id: {
        type: "string",
        description: "ID of the device to pair with"
      }
    },
    required: ["pair_code", "device_id"]
  }
};

// Define the slide_update_agent tool
const UPDATE_AGENT_TOOL = {
  name: "slide_update_agent",
  description: "Update an agent's properties",
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

// Define the slide_list_backups tool
const LIST_BACKUPS_TOOL = {
  name: "slide_list_backups",
  description: "List all backups with pagination and filtering options",
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
      },
      sort_by: {
        type: "string",
        description: "Sort by field (id, start_time)"
      }
    }
  }
};

// Define the slide_get_backup tool
const GET_BACKUP_TOOL = {
  name: "slide_get_backup",
  description: "Get detailed information about a specific backup",
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

// Define the slide_start_backup tool
const START_BACKUP_TOOL = {
  name: "slide_start_backup",
  description: "Start a backup for a specific agent",
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

// Define the slide_list_snapshots tool
const LIST_SNAPSHOTS_TOOL = {
  name: "slide_list_snapshots",
  description: "List all snapshots with pagination and filtering options",
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
        description: "Filter by snapshot location (exists_local, exists_cloud, exists_deleted, exists_deleted_retention, exists_deleted_manual, exists_deleted_other)"
      },
      sort_asc: {
        type: "boolean",
        description: "Sort in ascending order"
      },
      sort_by: {
        type: "string",
        description: "Sort by field (backup_start_time, backup_end_time, created)"
      }
    }
  }
};

// Define the slide_get_snapshot tool
const GET_SNAPSHOT_TOOL = {
  name: "slide_get_snapshot",
  description: "Get detailed information about a specific snapshot",
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

// Define the slide_list_file_restores tool
const LIST_FILE_RESTORES_TOOL = {
  name: "slide_list_file_restores",
  description: "List all file restores with pagination and filtering options",
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
      },
      sort_by: {
        type: "string",
        description: "Sort by field (id)"
      }
    }
  }
};

// Define the slide_get_file_restore tool
const GET_FILE_RESTORE_TOOL = {
  name: "slide_get_file_restore",
  description: "Get detailed information about a specific file restore",
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

// Define the slide_create_file_restore tool
const CREATE_FILE_RESTORE_TOOL = {
  name: "slide_create_file_restore",
  description: "Create a file restore from a snapshot",
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

// Define the slide_delete_file_restore tool
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

// Define the slide_browse_file_restore tool
const BROWSE_FILE_RESTORE_TOOL = {
  name: "slide_browse_file_restore",
  description: "Browse the contents of a file restore. IMPORTANT: You must first create a file restore using slide_create_file_restore before you can browse it.",
  inputSchema: {
    type: "object",
    properties: {
      file_restore_id: {
        type: "string",
        description: "ID of the file restore to browse"
      },
      path: {
        type: "string",
        description: "Path to browse (e.g., 'C' for root of C drive)"
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

// Define the slide_list_image_exports tool
const LIST_IMAGE_EXPORTS_TOOL = {
  name: "slide_list_image_exports",
  description: "List all image exports with pagination and filtering options",
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
      },
      sort_by: {
        type: "string",
        description: "Sort by field (id)"
      }
    }
  }
};

// Define the slide_get_image_export tool
const GET_IMAGE_EXPORT_TOOL = {
  name: "slide_get_image_export",
  description: "Get detailed information about a specific image export",
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

// Define the slide_create_image_export tool
const CREATE_IMAGE_EXPORT_TOOL = {
  name: "slide_create_image_export",
  description: "Create an image export from a snapshot",
  inputSchema: {
    type: "object",
    properties: {
      snapshot_id: {
        type: "string",
        description: "ID of the snapshot to export from"
      },
      device_id: {
        type: "string",
        description: "ID of the device to export to"
      },
      image_type: {
        type: "string",
        description: "Image type to export (vhdx, vhdx-dynamic, vhd, raw)"
      },
      boot_mods: {
        type: "array",
        items: {
          type: "string",
          enum: ["passwordless_admin_user"]
        },
        description: "Optional boot modifications to apply (e.g., 'passwordless_admin_user')"
      }
    },
    required: ["snapshot_id", "device_id", "image_type"]
  }
};

// Define the slide_delete_image_export tool
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

// Define the slide_browse_image_export tool
const BROWSE_IMAGE_EXPORT_TOOL = {
  name: "slide_browse_image_export",
  description: "Browse the contents of an image export. IMPORTANT: You must first create an image export using slide_create_image_export before you can browse it.",
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

// Define the slide_list_virtual_machines tool
const LIST_VIRTUAL_MACHINES_TOOL = {
  name: "slide_list_virtual_machines",
  description: "List all virtual machines with pagination and filtering options",
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
      },
      sort_by: {
        type: "string",
        description: "Sort by field (created)"
      }
    }
  }
};

// Define the slide_get_virtual_machine tool
const GET_VIRTUAL_MACHINE_TOOL = {
  name: "slide_get_virtual_machine",
  description: "Get detailed information about a specific virtual machine",
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

// Define the slide_create_virtual_machine tool
const CREATE_VIRTUAL_MACHINE_TOOL = {
  name: "slide_create_virtual_machine",
  description: "Create a virtual machine from a snapshot",
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
      },
      cpu_count: {
        type: "number",
        description: "Number of CPU cores (1-16)"
      },
      memory_in_mb: {
        type: "number",
        description: "Amount of memory in MB (1024-12288). Recommended default: 8192MB"
      },
      disk_bus: {
        type: "string",
        description: "Disk bus type (sata or virtio)"
      },
      network_model: {
        type: "string",
        description: "Network adapter model (hypervisor_default, e1000, rtl8139)"
      },
      network_type: {
        type: "string",
        description: "Network type (network, network-isolated, bridge, network-id)"
      },
      network_source: {
        type: "string",
        description: "Network ID when network_type is network-id"
      },
      boot_mods: {
        type: "array",
        items: {
          type: "string",
          enum: ["passwordless_admin_user"]
        },
        description: "Optional boot modifications to apply (e.g., 'passwordless_admin_user')"
      }
    },
    required: ["snapshot_id", "device_id"]
  }
};

// Define the slide_update_virtual_machine tool
const UPDATE_VIRTUAL_MACHINE_TOOL = {
  name: "slide_update_virtual_machine",
  description: "Update a virtual machine's properties",
  inputSchema: {
    type: "object",
    properties: {
      virt_id: {
        type: "string",
        description: "ID of the virtual machine to update"
      },
      state: {
        type: "string",
        description: "New state of the VM (running, stopped, paused)"
      },
      expires_at: {
        type: "string",
        description: "Expiration time in ISO 8601 format (e.g., 2024-08-23T01:25:08Z)"
      },
      memory_in_mb: {
        type: "number",
        description: "New amount of memory in MB (1024-12288)"
      },
      cpu_count: {
        type: "number",
        description: "New number of CPU cores (1-16)"
      }
    },
    required: ["virt_id"]
  }
};

// Define the slide_delete_virtual_machine tool
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

// Define the slide_list_users tool
const LIST_USERS_TOOL = {
  name: "slide_list_users",
  description: "List all users with pagination and filtering options",
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
      },
      sort_by: {
        type: "string",
        description: "Sort by field (id)"
      }
    }
  }
};

// Define the slide_get_user tool
const GET_USER_TOOL = {
  name: "slide_get_user",
  description: "Get detailed information about a specific user",
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

// Define the slide_list_alerts tool
const LIST_ALERTS_TOOL = {
  name: "slide_list_alerts",
  description: "List all alerts with pagination and filtering options",
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
      },
      sort_by: {
        type: "string",
        description: "Sort by field (created)"
      }
    }
  }
};

// Define the slide_get_alert tool
const GET_ALERT_TOOL = {
  name: "slide_get_alert",
  description: "Get detailed information about a specific alert",
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

// Define the slide_update_alert tool
const UPDATE_ALERT_TOOL = {
  name: "slide_update_alert",
  description: "Update an alert's properties (primarily used to resolve alerts)",
  inputSchema: {
    type: "object",
    properties: {
      alert_id: {
        type: "string",
        description: "ID of the alert to update"
      },
      resolved: {
        type: "boolean",
        description: "Set to true to resolve the alert"
      }
    },
    required: ["alert_id", "resolved"]
  }
};

// Define the slide_list_accounts tool
const LIST_ACCOUNTS_TOOL = {
  name: "slide_list_accounts",
  description: "List all accounts with pagination and filtering options",
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
      },
      sort_by: {
        type: "string",
        description: "Sort by field (name)"
      }
    }
  }
};

// Define the slide_get_account tool
const GET_ACCOUNT_TOOL = {
  name: "slide_get_account",
  description: "Get detailed information about a specific account",
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

// Define the slide_update_account tool
const UPDATE_ACCOUNT_TOOL = {
  name: "slide_update_account",
  description: "Update an account's properties (primarily alert emails)",
  inputSchema: {
    type: "object",
    properties: {
      account_id: {
        type: "string",
        description: "ID of the account to update"
      },
      alert_emails: {
        type: "array",
        items: {
          type: "string"
        },
        description: "List of email addresses to send alert emails to"
      }
    },
    required: ["account_id", "alert_emails"]
  }
};

// Define the slide_list_clients tool
const LIST_CLIENTS_TOOL = {
  name: "slide_list_clients",
  description: "List all clients with pagination and filtering options",
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
      },
      sort_by: {
        type: "string",
        description: "Sort by field (id)"
      }
    }
  }
};

// Define the slide_get_client tool
const GET_CLIENT_TOOL = {
  name: "slide_get_client",
  description: "Get detailed information about a specific client",
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

// Define the slide_create_client tool
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

// Define the slide_update_client tool
const UPDATE_CLIENT_TOOL = {
  name: "slide_update_client",
  description: "Update a client's properties",
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

// Define the slide_delete_client tool
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

// Function to check if args are valid for the list_devices tool
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

// Function to check if args are valid for the list_clients tool
function isListClientsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_client tool
function isGetClientArgs(args: unknown): args is {
  client_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).client_id === "string"
  );
}

// Function to check if args are valid for the create_client tool
function isCreateClientArgs(args: unknown): args is {
  name: string;
  comments?: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).name === "string" &&
    (typeof (args as any).comments === "string" || (args as any).comments === undefined)
  );
}

// Function to check if args are valid for the update_client tool
function isUpdateClientArgs(args: unknown): args is {
  client_id: string;
  name?: string;
  comments?: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).client_id === "string" &&
    (typeof (args as any).name === "string" || (args as any).name === undefined) &&
    (typeof (args as any).comments === "string" || (args as any).comments === undefined)
  );
}

// Function to check if args are valid for the delete_client tool
function isDeleteClientArgs(args: unknown): args is {
  client_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).client_id === "string"
  );
}

// Function to check if args are valid for the list_agents tool
function isListAgentsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  device_id?: string;
  client_id?: string;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_agent tool
function isGetAgentArgs(args: unknown): args is {
  agent_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).agent_id === "string"
  );
}

// Function to check if args are valid for the create_agent tool
function isCreateAgentArgs(args: unknown): args is {
  display_name: string;
  device_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).display_name === "string" &&
    typeof (args as any).device_id === "string"
  );
}

// Function to check if args are valid for the pair_agent tool
function isPairAgentArgs(args: unknown): args is {
  pair_code: string;
  device_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).pair_code === "string" &&
    typeof (args as any).device_id === "string"
  );
}

// Function to check if args are valid for the update_agent tool
function isUpdateAgentArgs(args: unknown): args is {
  agent_id: string;
  display_name: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).agent_id === "string" &&
    typeof (args as any).display_name === "string"
  );
}

// Function to check if args are valid for the list_backups tool
function isListBackupsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  agent_id?: string;
  device_id?: string;
  snapshot_id?: string;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_backup tool
function isGetBackupArgs(args: unknown): args is {
  backup_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).backup_id === "string"
  );
}

// Function to check if args are valid for the start_backup tool
function isStartBackupArgs(args: unknown): args is {
  agent_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).agent_id === "string"
  );
}

// Function to check if args are valid for the list_snapshots tool
function isListSnapshotsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  agent_id?: string;
  snapshot_location?: string;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_snapshot tool
function isGetSnapshotArgs(args: unknown): args is {
  snapshot_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).snapshot_id === "string"
  );
}

// Function to check if args are valid for the list_file_restores tool
function isListFileRestoresArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_file_restore tool
function isGetFileRestoreArgs(args: unknown): args is {
  file_restore_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).file_restore_id === "string"
  );
}

// Function to check if args are valid for the create_file_restore tool
function isCreateFileRestoreArgs(args: unknown): args is {
  snapshot_id: string;
  device_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).snapshot_id === "string" &&
    typeof (args as any).device_id === "string"
  );
}

// Function to check if args are valid for the delete_file_restore tool
function isDeleteFileRestoreArgs(args: unknown): args is {
  file_restore_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).file_restore_id === "string"
  );
}

// Function to check if args are valid for the browse_file_restore tool
function isBrowseFileRestoreArgs(args: unknown): args is {
  file_restore_id: string;
  path: string;
  limit?: number;
  offset?: number;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).file_restore_id === "string" &&
    typeof (args as any).path === "string"
  );
}

// Function to check if args are valid for the list_image_exports tool
function isListImageExportsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_image_export tool
function isGetImageExportArgs(args: unknown): args is {
  image_export_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).image_export_id === "string"
  );
}

// Function to check if args are valid for the create_image_export tool
function isCreateImageExportArgs(args: unknown): args is {
  snapshot_id: string;
  device_id: string;
  image_type: string;
  boot_mods?: string[];
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).snapshot_id === "string" &&
    typeof (args as any).device_id === "string" &&
    typeof (args as any).image_type === "string"
  );
}

// Function to check if args are valid for the delete_image_export tool
function isDeleteImageExportArgs(args: unknown): args is {
  image_export_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).image_export_id === "string"
  );
}

// Function to check if args are valid for the browse_image_export tool
function isBrowseImageExportArgs(args: unknown): args is {
  image_export_id: string;
  limit?: number;
  offset?: number;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).image_export_id === "string"
  );
}

// Function to check if args are valid for the list_virtual_machines tool
function isListVirtualMachinesArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_virtual_machine tool
function isGetVirtualMachineArgs(args: unknown): args is {
  virt_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).virt_id === "string"
  );
}

// Function to check if args are valid for the create_virtual_machine tool
function isCreateVirtualMachineArgs(args: unknown): args is {
  snapshot_id: string;
  device_id: string;
  cpu_count?: number;
  memory_in_mb?: number;
  disk_bus?: string;
  network_model?: string;
  network_type?: string;
  network_source?: string;
  boot_mods?: string[];
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).snapshot_id === "string" &&
    typeof (args as any).device_id === "string"
  );
}

// Function to check if args are valid for the update_virtual_machine tool
function isUpdateVirtualMachineArgs(args: unknown): args is {
  virt_id: string;
  state?: string;
  expires_at?: string;
  memory_in_mb?: number;
  cpu_count?: number;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).virt_id === "string"
  );
}

// Function to check if args are valid for the delete_virtual_machine tool
function isDeleteVirtualMachineArgs(args: unknown): args is {
  virt_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).virt_id === "string"
  );
}

// Function to check if args are valid for the list_users tool
function isListUsersArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_user tool
function isGetUserArgs(args: unknown): args is {
  user_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).user_id === "string"
  );
}

// Function to check if args are valid for the list_alerts tool
function isListAlertsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  device_id?: string;
  agent_id?: string;
  resolved?: boolean;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_alert tool
function isGetAlertArgs(args: unknown): args is {
  alert_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).alert_id === "string"
  );
}

// Function to check if args are valid for the update_alert tool
function isUpdateAlertArgs(args: unknown): args is {
  alert_id: string;
  resolved: boolean;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).alert_id === "string" &&
    typeof (args as any).resolved === "boolean"
  );
}

// Function to check if args are valid for the list_accounts tool
function isListAccountsArgs(args: unknown): args is {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
} {
  return (
    typeof args === "object" &&
    args !== null
  );
}

// Function to check if args are valid for the get_account tool
function isGetAccountArgs(args: unknown): args is {
  account_id: string;
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).account_id === "string"
  );
}

// Function to check if args are valid for the update_account tool
function isUpdateAccountArgs(args: unknown): args is {
  account_id: string;
  alert_emails: string[];
} {
  return (
    typeof args === "object" &&
    args !== null &&
    typeof (args as any).account_id === "string" &&
    Array.isArray((args as any).alert_emails) &&
    (args as any).alert_emails.every((email: any) => typeof email === "string")
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

// Function to list agents
async function listAgents(args: {
  limit?: number;
  offset?: number;
  device_id?: string;
  client_id?: string;
  sort_asc?: boolean;
  sort_by?: string;
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
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by hostname
      queryParams.append('sort_by', 'hostname');
    }
    
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

// Function to get agent by ID
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

// Function to create agent
async function createAgent(args: { display_name: string; device_id: string }) {
  try {
    const response = await axios.post<AgentPairCode>(
      `${API_BASE_URL}/v1/agent`,
      {
        display_name: args.display_name,
        device_id: args.device_id
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

// Function to pair agent
async function pairAgent(args: { pair_code: string; device_id: string }) {
  try {
    const response = await axios.post<Agent>(
      `${API_BASE_URL}/v1/agent/pair`,
      {
        pair_code: args.pair_code,
        device_id: args.device_id
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

// Function to update agent
async function updateAgent(args: { agent_id: string; display_name: string }) {
  try {
    const response = await axios.patch<Agent>(
      `${API_BASE_URL}/v1/agent/${args.agent_id}`,
      {
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

// Function to list backups
async function listBackups(args: {
  limit?: number;
  offset?: number;
  agent_id?: string;
  device_id?: string;
  snapshot_id?: string;
  sort_asc?: boolean;
  sort_by?: string;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.agent_id) {
      queryParams.append('agent_id', args.agent_id);
    }
    
    if (args.device_id) {
      queryParams.append('device_id', args.device_id);
    }
    
    if (args.snapshot_id) {
      queryParams.append('snapshot_id', args.snapshot_id);
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by start_time
      queryParams.append('sort_by', 'start_time');
    }
    
    const response = await axios.get<PaginatedResponse<Backup>>(
      `${API_BASE_URL}/v1/backup?${queryParams.toString()}`,
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

// Function to get backup by ID
async function getBackup(args: { backup_id: string }) {
  try {
    const response = await axios.get<Backup>(
      `${API_BASE_URL}/v1/backup/${args.backup_id}`,
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

// Function to start a backup
async function startBackup(args: { agent_id: string }) {
  try {
    const response = await axios.post<{ backup_id: string }>(
      `${API_BASE_URL}/v1/backup`,
      {
        agent_id: args.agent_id
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

// Function to list snapshots
async function listSnapshots(args: {
  limit?: number;
  offset?: number;
  agent_id?: string;
  snapshot_location?: string;
  sort_asc?: boolean;
  sort_by?: string;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.agent_id) {
      queryParams.append('agent_id', args.agent_id);
    }
    
    if (args.snapshot_location) {
      queryParams.append('snapshot_location', args.snapshot_location);
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by created
      queryParams.append('sort_by', 'created');
    }
    
    const response = await axios.get<PaginatedResponse<Snapshot>>(
      `${API_BASE_URL}/v1/snapshot?${queryParams.toString()}`,
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

// Function to get snapshot by ID
async function getSnapshot(args: { snapshot_id: string }) {
  try {
    const response = await axios.get<Snapshot>(
      `${API_BASE_URL}/v1/snapshot/${args.snapshot_id}`,
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

// Function to list file restores
async function listFileRestores(args: {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by id
      queryParams.append('sort_by', 'id');
    }
    
    const response = await axios.get<PaginatedResponse<FileRestore>>(
      `${API_BASE_URL}/v1/restore/file?${queryParams.toString()}`,
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

// Function to get file restore by ID
async function getFileRestore(args: { file_restore_id: string }) {
  try {
    const response = await axios.get<FileRestore>(
      `${API_BASE_URL}/v1/restore/file/${args.file_restore_id}`,
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

// Function to create a file restore
async function createFileRestore(args: { snapshot_id: string; device_id: string }) {
  try {
    const response = await axios.post<FileRestore>(
      `${API_BASE_URL}/v1/restore/file`,
      {
        snapshot_id: args.snapshot_id,
        device_id: args.device_id
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

// Function to delete a file restore
async function deleteFileRestore(args: { file_restore_id: string }) {
  try {
    await axios.delete(
      `${API_BASE_URL}/v1/restore/file/${args.file_restore_id}`,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return { success: true, message: `File restore ${args.file_restore_id} deleted successfully` };
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

// Function to browse a file restore
async function browseFileRestore(args: { 
  file_restore_id: string; 
  path: string;
  limit?: number;
  offset?: number;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    queryParams.append('path', args.path);
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    const response = await axios.get<PaginatedResponse<FileRestoreEntry>>(
      `${API_BASE_URL}/v1/restore/file/${args.file_restore_id}/browse?${queryParams.toString()}`,
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

// Function to list image exports
async function listImageExports(args: {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by id
      queryParams.append('sort_by', 'id');
    }
    
    const response = await axios.get<PaginatedResponse<ImageExport>>(
      `${API_BASE_URL}/v1/restore/image?${queryParams.toString()}`,
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

// Function to get image export by ID
async function getImageExport(args: { image_export_id: string }) {
  try {
    const response = await axios.get<ImageExport>(
      `${API_BASE_URL}/v1/restore/image/${args.image_export_id}`,
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

// Function to create an image export
async function createImageExport(args: { 
  snapshot_id: string; 
  device_id: string;
  image_type: string;
  boot_mods?: string[];
}) {
  try {
    const requestBody: any = {
      snapshot_id: args.snapshot_id,
      device_id: args.device_id,
      image_type: args.image_type
    };
    
    if (args.boot_mods) {
      requestBody.boot_mods = args.boot_mods;
    }
    
    const response = await axios.post<ImageExport>(
      `${API_BASE_URL}/v1/restore/image`,
      requestBody,
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

// Function to delete an image export
async function deleteImageExport(args: { image_export_id: string }) {
  try {
    await axios.delete(
      `${API_BASE_URL}/v1/restore/image/${args.image_export_id}`,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return { success: true, message: `Image export ${args.image_export_id} deleted successfully` };
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

// Function to browse an image export
async function browseImageExport(args: { 
  image_export_id: string;
  limit?: number;
  offset?: number;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    const response = await axios.get<PaginatedResponse<ImageExportEntry>>(
      `${API_BASE_URL}/v1/restore/image/${args.image_export_id}/browse?${queryParams.toString()}`,
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

// Function to list virtual machines
async function listVirtualMachines(args: {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by created
      queryParams.append('sort_by', 'created');
    }
    
    const response = await axios.get<PaginatedResponse<VirtualMachine>>(
      `${API_BASE_URL}/v1/restore/virt?${queryParams.toString()}`,
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

// Function to get virtual machine by ID
async function getVirtualMachine(args: { virt_id: string }) {
  try {
    const response = await axios.get<VirtualMachine>(
      `${API_BASE_URL}/v1/restore/virt/${args.virt_id}`,
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

// Function to create a virtual machine
async function createVirtualMachine(args: { 
  snapshot_id: string; 
  device_id: string;
  cpu_count?: number;
  memory_in_mb?: number;
  disk_bus?: string;
  network_model?: string;
  network_type?: string;
  network_source?: string;
  boot_mods?: string[];
}) {
  try {
    const requestBody: any = {
      snapshot_id: args.snapshot_id,
      device_id: args.device_id
    };
    
    if (args.cpu_count !== undefined) {
      requestBody.cpu_count = args.cpu_count;
    }
    
    if (args.memory_in_mb !== undefined) {
      requestBody.memory_in_mb = args.memory_in_mb;
    }
    
    if (args.disk_bus) {
      requestBody.disk_bus = args.disk_bus;
    }
    
    if (args.network_model) {
      requestBody.network_model = args.network_model;
    }
    
    if (args.network_type) {
      requestBody.network_type = args.network_type;
    }
    
    if (args.network_source) {
      requestBody.network_source = args.network_source;
    }
    
    if (args.boot_mods) {
      requestBody.boot_mods = args.boot_mods;
    }
    
    const response = await axios.post<VirtualMachine>(
      `${API_BASE_URL}/v1/restore/virt`,
      requestBody,
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

// Function to update a virtual machine
async function updateVirtualMachine(args: { 
  virt_id: string;
  state?: string;
  expires_at?: string;
  memory_in_mb?: number;
  cpu_count?: number;
}) {
  try {
    const requestBody: any = {};
    
    if (args.state) {
      requestBody.state = args.state;
    }
    
    if (args.expires_at) {
      requestBody.expires_at = args.expires_at;
    }
    
    if (args.memory_in_mb !== undefined) {
      requestBody.memory_in_mb = args.memory_in_mb;
    }
    
    if (args.cpu_count !== undefined) {
      requestBody.cpu_count = args.cpu_count;
    }
    
    const response = await axios.patch<VirtualMachine>(
      `${API_BASE_URL}/v1/restore/virt/${args.virt_id}`,
      requestBody,
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

// Function to delete a virtual machine
async function deleteVirtualMachine(args: { virt_id: string }) {
  try {
    await axios.delete(
      `${API_BASE_URL}/v1/restore/virt/${args.virt_id}`,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return { success: true, message: `Virtual machine ${args.virt_id} deleted successfully` };
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

// Function to list users
async function listUsers(args: {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by id
      queryParams.append('sort_by', 'id');
    }
    
    const response = await axios.get<PaginatedResponse<User>>(
      `${API_BASE_URL}/v1/user?${queryParams.toString()}`,
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

// Function to get user by ID
async function getUser(args: { user_id: string }) {
  try {
    const response = await axios.get<User>(
      `${API_BASE_URL}/v1/user/${args.user_id}`,
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

// Function to list alerts
async function listAlerts(args: {
  limit?: number;
  offset?: number;
  device_id?: string;
  agent_id?: string;
  resolved?: boolean;
  sort_asc?: boolean;
  sort_by?: string;
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
    
    if (args.agent_id) {
      queryParams.append('agent_id', args.agent_id);
    }
    
    if (args.resolved !== undefined) {
      queryParams.append('resolved', args.resolved.toString());
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by created
      queryParams.append('sort_by', 'created');
    }
    
    const response = await axios.get<PaginatedResponse<Alert>>(
      `${API_BASE_URL}/v1/alert?${queryParams.toString()}`,
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

// Function to get alert by ID
async function getAlert(args: { alert_id: string }) {
  try {
    const response = await axios.get<Alert>(
      `${API_BASE_URL}/v1/alert/${args.alert_id}`,
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

// Function to update an alert
async function updateAlert(args: { alert_id: string; resolved: boolean }) {
  try {
    const response = await axios.patch<Alert>(
      `${API_BASE_URL}/v1/alert/${args.alert_id}`,
      {
        resolved: args.resolved
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

// Function to list accounts
async function listAccounts(args: {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by name
      queryParams.append('sort_by', 'name');
    }
    
    const response = await axios.get<PaginatedResponse<Account>>(
      `${API_BASE_URL}/v1/account?${queryParams.toString()}`,
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

// Function to get account by ID
async function getAccount(args: { account_id: string }) {
  try {
    const response = await axios.get<Account>(
      `${API_BASE_URL}/v1/account/${args.account_id}`,
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

// Function to update an account
async function updateAccount(args: { account_id: string; alert_emails: string[] }) {
  try {
    const response = await axios.patch<Account>(
      `${API_BASE_URL}/v1/account/${args.account_id}`,
      {
        alert_emails: args.alert_emails
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

// Function to list clients
async function listClients(args: {
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  sort_by?: string;
}) {
  try {
    const queryParams = new URLSearchParams();
    
    if (args.limit) {
      queryParams.append('limit', args.limit.toString());
    }
    
    if (args.offset) {
      queryParams.append('offset', args.offset.toString());
    }
    
    if (args.sort_asc !== undefined) {
      queryParams.append('sort_asc', args.sort_asc.toString());
    }
    
    if (args.sort_by) {
      queryParams.append('sort_by', args.sort_by);
    } else {
      // Default sort by id
      queryParams.append('sort_by', 'id');
    }
    
    const response = await axios.get<PaginatedResponse<Client>>(
      `${API_BASE_URL}/v1/client?${queryParams.toString()}`,
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

// Function to get client by ID
async function getClient(args: { client_id: string }) {
  try {
    const response = await axios.get<Client>(
      `${API_BASE_URL}/v1/client/${args.client_id}`,
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

// Function to create a client
async function createClient(args: { name: string; comments?: string }) {
  try {
    const requestBody: any = {
      name: args.name
    };
    
    if (args.comments) {
      requestBody.comments = args.comments;
    }
    
    const response = await axios.post<Client>(
      `${API_BASE_URL}/v1/client`,
      requestBody,
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

// Function to update a client
async function updateClient(args: { client_id: string; name?: string; comments?: string }) {
  try {
    const requestBody: any = {};
    
    if (args.name) {
      requestBody.name = args.name;
    }
    
    if (args.comments !== undefined) {
      requestBody.comments = args.comments;
    }
    
    const response = await axios.patch<Client>(
      `${API_BASE_URL}/v1/client/${args.client_id}`,
      requestBody,
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

// Function to delete a client
async function deleteClient(args: { client_id: string }) {
  try {
    await axios.delete(
      `${API_BASE_URL}/v1/client/${args.client_id}`,
      {
        headers: {
          'Authorization': `Bearer ${SLIDE_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    return { success: true, message: `Client ${args.client_id} deleted successfully` };
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
    LIST_USERS_TOOL,
    GET_USER_TOOL,
    LIST_ALERTS_TOOL,
    GET_ALERT_TOOL,
    UPDATE_ALERT_TOOL,
    LIST_ACCOUNTS_TOOL,
    GET_ACCOUNT_TOOL,
    UPDATE_ACCOUNT_TOOL,
    LIST_CLIENTS_TOOL,
    GET_CLIENT_TOOL,
    CREATE_CLIENT_TOOL,
    UPDATE_CLIENT_TOOL,
    DELETE_CLIENT_TOOL
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
      case "slide_list_devices": {
        if (!isListDevicesArgs(args)) {
          throw new Error("Invalid arguments for slide_list_devices");
        }
        
        const result = await listDevices(args);
        
        // Add metadata to guide the LLM on how to present and refer to devices
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "display_name",
            presentation_guidance: "When referring to devices, use the Display Name as the primary identifier. If its blank use hostname. Device IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_agents": {
        if (!isListAgentsArgs(args)) {
          throw new Error("Invalid arguments for slide_list_agents");
        }
        
        const result = await listAgents(args);
        
        // Add metadata to guide the LLM on how to present and refer to agents
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "display_name",
            presentation_guidance: "When referring to agents, use the Display Name as the primary identifier. If Display Name is blank, use hostname instead. Agent IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_agent": {
        if (!isGetAgentArgs(args)) {
          throw new Error("Invalid arguments for slide_get_agent");
        }
        
        const result = await getAgent(args);
        
        // Add metadata to guide the LLM on how to present and refer to the agent
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "display_name",
            presentation_guidance: "When referring to the agent, use the Display Name as the primary identifier. If its blank use hostname. Agent IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_create_agent": {
        if (!isCreateAgentArgs(args)) {
          throw new Error("Invalid arguments for slide_create_agent");
        }
        
        const result = await createAgent(args);
        
        // Add metadata to guide the LLM on how to present and refer to the agent pair code
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "pair_code",
            presentation_guidance: "When referring to the agent pair code, use the pair_code as the primary identifier. Agent IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_pair_agent": {
        if (!isPairAgentArgs(args)) {
          throw new Error("Invalid arguments for slide_pair_agent");
        }
        
        const result = await pairAgent(args);
        
        // Add metadata to guide the LLM on how to present and refer to the agent
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "display_name",
            presentation_guidance: "When referring to the agent, use the Display Name as the primary identifier. If its blank use hostname. Agent IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_update_agent": {
        if (!isUpdateAgentArgs(args)) {
          throw new Error("Invalid arguments for slide_update_agent");
        }
        
        const result = await updateAgent(args);
        
        // Add metadata to guide the LLM on how to present and refer to the updated agent
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "display_name",
            presentation_guidance: "When referring to the updated agent, use the Display Name as the primary identifier. If its blank use hostname. Agent IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_backups": {
        if (!isListBackupsArgs(args)) {
          throw new Error("Invalid arguments for slide_list_backups");
        }
        
        const result = await listBackups(args);
        
        // Add metadata to guide the LLM on how to present and refer to backups
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "backup_id",
            presentation_guidance: "When referring to backups, use the backup_id as the primary identifier. Backup IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_backup": {
        if (!isGetBackupArgs(args)) {
          throw new Error("Invalid arguments for slide_get_backup");
        }
        
        const result = await getBackup(args);
        
        // Add metadata to guide the LLM on how to present and refer to the backup
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "backup_id",
            presentation_guidance: "When referring to the backup, use the backup_id as the primary identifier. Backup IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_start_backup": {
        if (!isStartBackupArgs(args)) {
          throw new Error("Invalid arguments for slide_start_backup");
        }
        
        const result = await startBackup(args);
        
        // Add metadata to guide the LLM on how to present and refer to the backup
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "backup_id",
            presentation_guidance: "When referring to the backup, use the backup_id as the primary identifier. Backup IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_snapshots": {
        if (!isListSnapshotsArgs(args)) {
          throw new Error("Invalid arguments for slide_list_snapshots");
        }
        
        const result = await listSnapshots(args);
        
        // Add metadata to guide the LLM on how to present and refer to snapshots
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "snapshot_id",
            presentation_guidance: "When referring to snapshots, use the snapshot_id as the primary identifier. Snapshot IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_snapshot": {
        if (!isGetSnapshotArgs(args)) {
          throw new Error("Invalid arguments for slide_get_snapshot");
        }
        
        const result = await getSnapshot(args);
        
        // Add metadata to guide the LLM on how to present and refer to the snapshot
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "snapshot_id",
            presentation_guidance: "When referring to the snapshot, use the snapshot_id as the primary identifier. Snapshot IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_file_restores": {
        if (!isListFileRestoresArgs(args)) {
          throw new Error("Invalid arguments for slide_list_file_restores");
        }
        
        const result = await listFileRestores(args);
        
        // Add metadata to guide the LLM on how to present and refer to file restores
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "file_restore_id",
            presentation_guidance: "When referring to file restores, use the file_restore_id as the primary identifier. File restore IDs are internal identifiers not commonly used by humans.",
            workflow_guidance: "File restores must be created before they can be browsed. To create a file restore, use slide_create_file_restore with a snapshot_id and device_id."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_file_restore": {
        if (!isGetFileRestoreArgs(args)) {
          throw new Error("Invalid arguments for slide_get_file_restore");
        }
        
        const result = await getFileRestore(args);
        
        // Add metadata to guide the LLM on how to present and refer to the file restore
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "file_restore_id",
            presentation_guidance: "When referring to the file restore, use the file_restore_id as the primary identifier. File restore IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_create_file_restore": {
        if (!isCreateFileRestoreArgs(args)) {
          throw new Error("Invalid arguments for slide_create_file_restore");
        }
        
        const result = await createFileRestore(args);
        
        // Add metadata to guide the LLM on how to present and refer to the file restore
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "file_restore_id",
            presentation_guidance: "When referring to the file restore, use the file_restore_id as the primary identifier. File restore IDs are internal identifiers not commonly used by humans.",
            next_steps: "Now that you've created a file restore, you can browse its contents using slide_browse_file_restore with this file_restore_id and a path parameter (e.g., 'C' for the root of C drive)."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_delete_file_restore": {
        if (!isDeleteFileRestoreArgs(args)) {
          throw new Error("Invalid arguments for slide_delete_file_restore");
        }
        
        const result = await deleteFileRestore(args);
        
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }

      case "slide_browse_file_restore": {
        if (!isBrowseFileRestoreArgs(args)) {
          throw new Error("Invalid arguments for slide_browse_file_restore");
        }
        
        const result = await browseFileRestore(args);
        
        // Add metadata to guide the LLM on how to present file browse results
        const enhancedResult = {
          ...result,
          _metadata: {
            presentation_guidance: "When presenting file browse results, organize by type (directories first, then files) and highlight download options for files.",
            workflow_guidance: "File restores are temporary. If a file_restore_id is not found, it may have expired or not been created yet. Create a file restore using slide_create_file_restore before browsing."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_image_exports": {
        if (!isListImageExportsArgs(args)) {
          throw new Error("Invalid arguments for slide_list_image_exports");
        }
        
        const result = await listImageExports(args);
        
        // Add metadata to guide the LLM on how to present and refer to image exports
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "image_export_id",
            presentation_guidance: "When referring to image exports, use the image_export_id as the primary identifier. Image export IDs are internal identifiers not commonly used by humans.",
            workflow_guidance: "Image exports must be created before they can be browsed. To create an image export, use slide_create_image_export with a snapshot_id, device_id, and image_type."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_image_export": {
        if (!isGetImageExportArgs(args)) {
          throw new Error("Invalid arguments for slide_get_image_export");
        }
        
        const result = await getImageExport(args);
        
        // Add metadata to guide the LLM on how to present and refer to the image export
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "image_export_id",
            presentation_guidance: "When referring to the image export, use the image_export_id as the primary identifier. Image export IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_create_image_export": {
        if (!isCreateImageExportArgs(args)) {
          throw new Error("Invalid arguments for slide_create_image_export");
        }
        
        const result = await createImageExport(args);
        
        // Add metadata to guide the LLM on how to present and refer to the image export
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "image_export_id",
            presentation_guidance: "When referring to the image export, use the image_export_id as the primary identifier. Image export IDs are internal identifiers not commonly used by humans.",
            next_steps: "Now that you've created an image export, you can browse its contents using slide_browse_image_export with this image_export_id."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_delete_image_export": {
        if (!isDeleteImageExportArgs(args)) {
          throw new Error("Invalid arguments for slide_delete_image_export");
        }
        
        const result = await deleteImageExport(args);
        
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }

      case "slide_browse_image_export": {
        if (!isBrowseImageExportArgs(args)) {
          throw new Error("Invalid arguments for slide_browse_image_export");
        }
        
        const result = await browseImageExport(args);
        
        // Add metadata to guide the LLM on how to present image export browse results
        const enhancedResult = {
          ...result,
          _metadata: {
            presentation_guidance: "When presenting image export results, highlight download options for disk images.",
            workflow_guidance: "Image exports are temporary. If an image_export_id is not found, it may have expired or not been created yet. Create an image export using slide_create_image_export before browsing."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_virtual_machines": {
        if (!isListVirtualMachinesArgs(args)) {
          throw new Error("Invalid arguments for slide_list_virtual_machines");
        }
        
        const result = await listVirtualMachines(args);
        
        // Process each VM to add VNC viewer URLs
        if (result.data && result.data.length > 0) {
          result.data = result.data.map(vm => {
            // Generate VNC viewer URL if websocket URI is available
            let vncViewerUrl = null;
            if (vm.vnc && vm.vnc.length > 0) {
              const vncInfo = vm.vnc.find(v => v.websocket_uri);
              if (vncInfo && vncInfo.websocket_uri) {
                vncViewerUrl = generateVncViewerUrl(vm.virt_id, vncInfo.websocket_uri, vm.vnc_password);
              }
            }
            
            // Add the VNC viewer URL to each VM
            return {
              ...vm,
              _vnc_viewer_url: vncViewerUrl
            };
          });
        }
        
        // Add metadata to guide the LLM on how to present and refer to virtual machines
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "virt_id",
            presentation_guidance: "When referring to virtual machines, use the virt_id as the primary identifier. Virtual machine IDs are internal identifiers not commonly used by humans.",
            workflow_guidance: "Virtual machines are created from snapshots. To create a virtual machine, use slide_create_virtual_machine with a snapshot_id and device_id.",
            vnc_guidance: "Each virtual machine includes a _vnc_viewer_url property that provides a direct link to access its console through a browser-based VNC client."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_virtual_machine": {
        if (!isGetVirtualMachineArgs(args)) {
          throw new Error("Invalid arguments for slide_get_virtual_machine");
        }
        
        const result = await getVirtualMachine(args);
        
        // Generate VNC viewer URL if websocket URI is available
        let vncViewerUrl = null;
        if (result.vnc && result.vnc.length > 0) {
          const vncInfo = result.vnc.find(v => v.websocket_uri);
          if (vncInfo && vncInfo.websocket_uri) {
            vncViewerUrl = generateVncViewerUrl(result.virt_id, vncInfo.websocket_uri, result.vnc_password);
          }
        }
        
        // Add metadata to guide the LLM on how to present and refer to the virtual machine
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "virt_id",
            presentation_guidance: "When referring to the virtual machine, use the virt_id as the primary identifier. Virtual machine IDs are internal identifiers not commonly used by humans.",
            vnc_guidance: "Use the vnc_viewer_url to access the virtual machine's console via a browser-based VNC client.",
            vnc_viewer_url: vncViewerUrl
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_create_virtual_machine": {
        if (!isCreateVirtualMachineArgs(args)) {
          throw new Error("Invalid arguments for slide_create_virtual_machine");
        }
        
        const result = await createVirtualMachine(args);
        
        // Generate VNC viewer URL if websocket URI is available
        let vncViewerUrl = null;
        if (result.vnc && result.vnc.length > 0) {
          const vncInfo = result.vnc.find(v => v.websocket_uri);
          if (vncInfo && vncInfo.websocket_uri) {
            vncViewerUrl = generateVncViewerUrl(result.virt_id, vncInfo.websocket_uri, result.vnc_password);
          }
        }
        
        // Add metadata to guide the LLM on how to present and refer to the virtual machine
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "virt_id",
            presentation_guidance: "When referring to the virtual machine, use the virt_id as the primary identifier. Virtual machine IDs are internal identifiers not commonly used by humans.",
            next_steps: "Now that you've created a virtual machine, you can control it using slide_update_virtual_machine to change its state (running, stopped, paused) or update resources.",
            vnc_guidance: "Use the vnc_viewer_url to access the virtual machine's console.",
            resource_guidance: "For optimal performance, 8192MB of RAM is recommended for most VMs. You can adjust this as needed using slide_update_virtual_machine.",
            vnc_viewer_url: vncViewerUrl
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_update_virtual_machine": {
        if (!isUpdateVirtualMachineArgs(args)) {
          throw new Error("Invalid arguments for slide_update_virtual_machine");
        }
        
        const result = await updateVirtualMachine(args);
        
        // Generate VNC viewer URL if websocket URI is available
        let vncViewerUrl = null;
        if (result.vnc && result.vnc.length > 0) {
          const vncInfo = result.vnc.find(v => v.websocket_uri);
          if (vncInfo && vncInfo.websocket_uri) {
            vncViewerUrl = generateVncViewerUrl(result.virt_id, vncInfo.websocket_uri, result.vnc_password);
          }
        }
        
        // Add metadata to guide the LLM on how to present and refer to the updated virtual machine
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "virt_id",
            presentation_guidance: "When referring to the virtual machine, use the virt_id as the primary identifier. Virtual machine IDs are internal identifiers not commonly used by humans.",
            vnc_guidance: "Use the vnc_viewer_url to access the virtual machine's console via a browser-based VNC client.",
            vnc_viewer_url: vncViewerUrl
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_delete_virtual_machine": {
        if (!isDeleteVirtualMachineArgs(args)) {
          throw new Error("Invalid arguments for slide_delete_virtual_machine");
        }
        
        const result = await deleteVirtualMachine(args);
        
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_users": {
        if (!isListUsersArgs(args)) {
          throw new Error("Invalid arguments for slide_list_users");
        }
        
        const result = await listUsers(args);
        
        // Add metadata to guide the LLM on how to present and refer to users
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "display_name",
            presentation_guidance: "When referring to users, use the display name (formatted as 'First Last') as the primary identifier. User IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_user": {
        if (!isGetUserArgs(args)) {
          throw new Error("Invalid arguments for slide_get_user");
        }
        
        const result = await getUser(args);
        
        // Add metadata to guide the LLM on how to present and refer to the user
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "display_name",
            presentation_guidance: "When referring to the user, use the display name (formatted as 'First Last') as the primary identifier. User IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_alerts": {
        if (!isListAlertsArgs(args)) {
          throw new Error("Invalid arguments for slide_list_alerts");
        }
        
        const result = await listAlerts(args);
        
        // Add metadata to guide the LLM on how to present and refer to alerts
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "alert_id",
            presentation_guidance: "When presenting alerts, highlight unresolved alerts first, and categorize by alert_type. Alert IDs are internal identifiers."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_alert": {
        if (!isGetAlertArgs(args)) {
          throw new Error("Invalid arguments for slide_get_alert");
        }
        
        const result = await getAlert(args);
        
        // Add metadata to guide the LLM on how to present and refer to the alert
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "alert_id",
            presentation_guidance: "When presenting the alert, highlight its type, whether it's resolved, and related device or agent if present."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_update_alert": {
        if (!isUpdateAlertArgs(args)) {
          throw new Error("Invalid arguments for slide_update_alert");
        }
        
        const result = await updateAlert(args);
        
        // Add metadata to guide the LLM on how to present the updated alert
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "alert_id",
            presentation_guidance: "When confirming alert update, indicate whether it was successfully resolved."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_accounts": {
        if (!isListAccountsArgs(args)) {
          throw new Error("Invalid arguments for slide_list_accounts");
        }
        
        const result = await listAccounts(args);
        
        // Add metadata to guide the LLM on how to present and refer to accounts
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "account_name",
            presentation_guidance: "When presenting accounts, use the account name as the primary identifier. Account IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_account": {
        if (!isGetAccountArgs(args)) {
          throw new Error("Invalid arguments for slide_get_account");
        }
        
        const result = await getAccount(args);
        
        // Add metadata to guide the LLM on how to present and refer to the account
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "account_name",
            presentation_guidance: "When presenting account details, highlight account name, primary contact, and alert email settings."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_update_account": {
        if (!isUpdateAccountArgs(args)) {
          throw new Error("Invalid arguments for slide_update_account");
        }
        
        const result = await updateAccount(args);
        
        // Add metadata to guide the LLM on how to present the updated account
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "account_name",
            presentation_guidance: "When confirming account update, highlight the new alert email settings that were applied."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_list_clients": {
        if (!isListClientsArgs(args)) {
          throw new Error("Invalid arguments for slide_list_clients");
        }
        
        const result = await listClients(args);
        
        // Add metadata to guide the LLM on how to present and refer to clients
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "name",
            presentation_guidance: "When presenting clients, use the client name as the primary identifier. Client IDs are internal identifiers not commonly used by humans."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_get_client": {
        if (!isGetClientArgs(args)) {
          throw new Error("Invalid arguments for slide_get_client");
        }
        
        const result = await getClient(args);
        
        // Add metadata to guide the LLM on how to present and refer to the client
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "name",
            presentation_guidance: "When presenting client details, highlight client name and comments."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_create_client": {
        if (!isCreateClientArgs(args)) {
          throw new Error("Invalid arguments for slide_create_client");
        }
        
        const result = await createClient(args);
        
        // Add metadata to guide the LLM on how to present the created client
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "name",
            presentation_guidance: "When confirming client creation, highlight the new client name and ID."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_update_client": {
        if (!isUpdateClientArgs(args)) {
          throw new Error("Invalid arguments for slide_update_client");
        }
        
        const result = await updateClient(args);
        
        // Add metadata to guide the LLM on how to present the updated client
        const enhancedResult = {
          ...result,
          _metadata: {
            primary_identifier: "name",
            presentation_guidance: "When confirming client update, highlight the updated fields."
          }
        };
        
        return {
          content: [{ type: "text", text: JSON.stringify(enhancedResult, null, 2) }],
          isError: false,
        };
      }

      case "slide_delete_client": {
        if (!isDeleteClientArgs(args)) {
          throw new Error("Invalid arguments for slide_delete_client");
        }
        
        const result = await deleteClient(args);
        
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          isError: false,
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

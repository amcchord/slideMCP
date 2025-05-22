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

// Server implements the listTools request
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: [LIST_DEVICES_TOOL, LIST_AGENTS_TOOL, GET_AGENT_TOOL, CREATE_AGENT_TOOL, PAIR_AGENT_TOOL, UPDATE_AGENT_TOOL, LIST_BACKUPS_TOOL, GET_BACKUP_TOOL, START_BACKUP_TOOL],
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
            presentation_guidance: "When referring to agents, use the Display Name as the primary identifier. If its blank use hostname. Agent IDs are internal identifiers not commonly used by humans."
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

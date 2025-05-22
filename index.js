/**
 * Slide API Model Configuration Profile (MCP)
 * This module provides an interface for LLMs to interact with the Slide API
 */

const axios = require('axios');
const fs = require('fs');
require('dotenv').config();

// For debugging - log to a file
function logToFile(message) {
  fs.appendFileSync('/tmp/slide-mcp.log', new Date().toISOString() + ' ' + message + '\n');
}

class SlideClient {
  constructor(apiKey, baseURL = 'https://api.slide.tech', version = 'v1') {
    this.apiKey = apiKey;
    this.baseURL = baseURL;
    this.version = version;
    
    this.client = axios.create({
      baseURL: `${this.baseURL}/${this.version}`,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
      }
    });
  }

  // Helper method for GET requests
  async get(endpoint, params = {}) {
    try {
      const response = await this.client.get(endpoint, { params });
      return response.data;
    } catch (error) {
      this._handleError(error);
    }
  }

  // Helper method for POST requests
  async post(endpoint, data = {}) {
    try {
      const response = await this.client.post(endpoint, data);
      return response.data;
    } catch (error) {
      this._handleError(error);
    }
  }

  // Helper method for PATCH requests
  async patch(endpoint, data = {}) {
    try {
      const response = await this.client.patch(endpoint, data);
      return response.data;
    } catch (error) {
      this._handleError(error);
    }
  }

  // Helper method for DELETE requests
  async delete(endpoint) {
    try {
      const response = await this.client.delete(endpoint);
      return response.data;
    } catch (error) {
      this._handleError(error);
    }
  }

  // Error handler
  _handleError(error) {
    if (error.response) {
      // The request was made and the server responded with a status code
      // that falls out of the range of 2xx
      throw new Error(`API Error: ${JSON.stringify(error.response.data)}`);
    } else if (error.request) {
      // The request was made but no response was received
      throw new Error('No response received from API');
    } else {
      // Something happened in setting up the request that triggered an Error
      throw new Error(`Error: ${error.message}`);
    }
  }

  // Devices API
  async getDevices(params = {}) {
    return this.get('/device', params);
  }

  async getDevice(deviceId) {
    return this.get(`/device/${deviceId}`);
  }

  async updateDevice(deviceId, data) {
    return this.patch(`/device/${deviceId}`, data);
  }

  // Agents API
  async getAgents(params = {}) {
    return this.get('/agent', params);
  }

  async getAgent(agentId) {
    return this.get(`/agent/${agentId}`);
  }

  async updateAgent(agentId, data) {
    return this.patch(`/agent/${agentId}`, data);
  }

  async createAgentPair(data) {
    return this.post('/agent', data);
  }

  async pairAgent(data) {
    return this.post('/agent/pair', data);
  }

  // Backups API
  async getBackups(params = {}) {
    return this.get('/backup', params);
  }

  async getBackup(backupId) {
    return this.get(`/backup/${backupId}`);
  }

  async startBackup(data) {
    return this.post('/backup', data);
  }

  // Snapshots API
  async getSnapshots(params = {}) {
    return this.get('/snapshot', params);
  }

  async getSnapshot(snapshotId) {
    return this.get(`/snapshot/${snapshotId}`);
  }

  // File Restores API
  async getFileRestores(params = {}) {
    return this.get('/restore/file', params);
  }

  async getFileRestore(fileRestoreId) {
    return this.get(`/restore/file/${fileRestoreId}`);
  }

  async createFileRestore(data) {
    return this.post('/restore/file', data);
  }

  async deleteFileRestore(fileRestoreId) {
    return this.delete(`/restore/file/${fileRestoreId}`);
  }

  async browseFileRestore(fileRestoreId, params = {}) {
    return this.get(`/restore/file/${fileRestoreId}/browse`, params);
  }

  // Image Exports API
  async getImageExports(params = {}) {
    return this.get('/restore/image', params);
  }

  async getImageExport(imageExportId) {
    return this.get(`/restore/image/${imageExportId}`);
  }

  async createImageExport(data) {
    return this.post('/restore/image', data);
  }

  async deleteImageExport(imageExportId) {
    return this.delete(`/restore/image/${imageExportId}`);
  }

  async browseImageExport(imageExportId, params = {}) {
    return this.get(`/restore/image/${imageExportId}/browse`, params);
  }

  // Virtual Machines API
  async getVirtualMachines(params = {}) {
    return this.get('/restore/virt', params);
  }

  async getVirtualMachine(virtId) {
    return this.get(`/restore/virt/${virtId}`);
  }

  async createVirtualMachine(data) {
    return this.post('/restore/virt', data);
  }

  async updateVirtualMachine(virtId, data) {
    return this.patch(`/restore/virt/${virtId}`, data);
  }

  async deleteVirtualMachine(virtId) {
    return this.delete(`/restore/virt/${virtId}`);
  }

  // Networks API
  async getNetworks(params = {}) {
    return this.get('/network', params);
  }

  async getNetwork(networkId) {
    return this.get(`/network/${networkId}`);
  }

  async createNetwork(data) {
    return this.post('/network', data);
  }

  async updateNetwork(networkId, data) {
    return this.patch(`/network/${networkId}`, data);
  }

  async deleteNetwork(networkId) {
    return this.delete(`/network/${networkId}`);
  }

  // Network Port Forwards API
  async createNetworkPortForward(networkId, data) {
    return this.post(`/network/${networkId}/port-forwards`, data);
  }

  async deleteNetworkPortForward(networkId, data) {
    return this.delete(`/network/${networkId}/port-forwards`, data);
  }

  // Network WireGuard Peers API
  async createNetworkWGPeer(networkId, data) {
    return this.post(`/network/${networkId}/wg-peers`, data);
  }

  async updateNetworkWGPeer(networkId, data) {
    return this.patch(`/network/${networkId}/wg-peers`, data);
  }

  async deleteNetworkWGPeer(networkId, data) {
    return this.delete(`/network/${networkId}/wg-peers`, data);
  }

  // Users API
  async getUsers(params = {}) {
    return this.get('/user', params);
  }

  async getUser(userId) {
    return this.get(`/user/${userId}`);
  }

  // Alerts API
  async getAlerts(params = {}) {
    return this.get('/alert', params);
  }

  async getAlert(alertId) {
    return this.get(`/alert/${alertId}`);
  }

  async updateAlert(alertId, data) {
    return this.patch(`/alert/${alertId}`, data);
  }

  // Account API
  async getAccounts(params = {}) {
    return this.get('/account', params);
  }

  async getAccount(accountId) {
    return this.get(`/account/${accountId}`);
  }

  async updateAccount(accountId, data) {
    return this.patch(`/account/${accountId}`, data);
  }
}

// Factory function to create a new SlideClient instance
function createClient(apiKey) {
  if (!apiKey) {
    throw new Error('API key is required');
  }
  return new SlideClient(apiKey);
}

// Handle tool calls for MCP
async function handleToolCall(client, toolName, args) {
  logToFile(`Handling tool call: ${toolName} with args: ${JSON.stringify(args)}`);
  
  switch(toolName) {
    // Devices
    case 'slide_get_devices':
      return await client.getDevices(args);
    case 'slide_get_device':
      return await client.getDevice(args.device_id);
    
    // Agents
    case 'slide_get_agents':
      return await client.getAgents(args);
    case 'slide_get_agent':
      return await client.getAgent(args.agent_id);
    
    // Backups
    case 'slide_get_backups':
      return await client.getBackups(args);
    case 'slide_start_backup':
      return await client.startBackup({ agent_id: args.agent_id });
    
    // Snapshots
    case 'slide_get_snapshots':
      return await client.getSnapshots(args);
    
    // Virtual Machines
    case 'slide_create_virtual_machine':
      return await client.createVirtualMachine(args);
    
    default:
      throw new Error(`Unknown tool: ${toolName}`);
  }
}

// Handle MCP protocol if this script is called directly
if (require.main === module) {
  // Create a client using API key from environment
  const apiKey = process.env.SLIDE_API_KEY;
  if (!apiKey) {
    logToFile('ERROR: No API key found. Please set SLIDE_API_KEY environment variable');
    process.exit(1);
  }
  
  try {
    logToFile(`Starting MCP server with API key: ${apiKey}`);
    const slideClient = createClient(apiKey);
    
    // Set up the line-by-line reader for stdin
    const readline = require('readline');
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
      terminal: false
    });
    
    // Handle each line of input as a separate request
    rl.on('line', async (line) => {
      if (!line.trim()) return;
      
      try {
        logToFile(`Received message: ${line}`);
        const message = JSON.parse(line);
        
        // Handle different message types
        if (message.jsonrpc === '2.0') {
          // This is a JSON-RPC message
          const { id, method, params } = message;
          
          if (method === 'initialize') {
            // Handle initialization request
            console.log(JSON.stringify({
              jsonrpc: '2.0',
              id,
              result: {
                protocolVersion: '2025-03-26',
                serverInfo: {
                  name: 'slideMCP',
                  version: '1.0.0'
                },
                capabilities: {
                  runTool: true
                },
                tools: [
                  {
                    name: 'slide_get_devices',
                    description: 'Get a list of Slide devices',
                    schema: {
                      type: 'object',
                      properties: {
                        limit: { type: 'integer' },
                        offset: { type: 'integer' }
                      }
                    }
                  },
                  {
                    name: 'slide_get_device',
                    description: 'Get details of a specific device by ID',
                    schema: {
                      type: 'object',
                      properties: {
                        device_id: { type: 'string' }
                      },
                      required: ['device_id']
                    }
                  },
                  {
                    name: 'slide_get_agents',
                    description: 'Get a list of Slide agents',
                    schema: {
                      type: 'object',
                      properties: {
                        device_id: { type: 'string' },
                        limit: { type: 'integer' },
                        offset: { type: 'integer' }
                      }
                    }
                  },
                  {
                    name: 'slide_get_agent',
                    description: 'Get details of a specific agent by ID',
                    schema: {
                      type: 'object',
                      properties: {
                        agent_id: { type: 'string' }
                      },
                      required: ['agent_id']
                    }
                  },
                  {
                    name: 'slide_get_backups',
                    description: 'Get a list of backups',
                    schema: {
                      type: 'object',
                      properties: {
                        agent_id: { type: 'string' },
                        limit: { type: 'integer' },
                        offset: { type: 'integer' }
                      }
                    }
                  },
                  {
                    name: 'slide_start_backup',
                    description: 'Start a new backup for a specific agent',
                    schema: {
                      type: 'object',
                      properties: {
                        agent_id: { type: 'string' }
                      },
                      required: ['agent_id']
                    }
                  },
                  {
                    name: 'slide_get_snapshots',
                    description: 'Get a list of snapshots',
                    schema: {
                      type: 'object',
                      properties: {
                        agent_id: { type: 'string' },
                        limit: { type: 'integer' },
                        offset: { type: 'integer' }
                      }
                    }
                  },
                  {
                    name: 'slide_create_virtual_machine',
                    description: 'Create a new virtual machine from a snapshot',
                    schema: {
                      type: 'object',
                      properties: {
                        snapshot_id: { type: 'string' },
                        device_id: { type: 'string' },
                        cpu_count: { type: 'integer' },
                        memory_in_mb: { type: 'integer' }
                      },
                      required: ['snapshot_id', 'device_id']
                    }
                  }
                ]
              }
            }));
          } else if (method === 'run') {
            // Handle tool execution
            try {
              const { name, arguments: args } = params;
              const result = await handleToolCall(slideClient, name, args || {});
              console.log(JSON.stringify({
                jsonrpc: '2.0',
                id,
                result
              }));
            } catch (error) {
              logToFile(`Tool execution error: ${error.message}`);
              console.log(JSON.stringify({
                jsonrpc: '2.0',
                id,
                error: {
                  code: -32000,
                  message: error.message
                }
              }));
            }
          } else if (method === 'listFunctions') {
            // Respond with list of supported tools
            console.log(JSON.stringify({
              jsonrpc: '2.0',
              id,
              result: [
                'slide_get_devices',
                'slide_get_device',
                'slide_get_agents',
                'slide_get_agent',
                'slide_get_backups',
                'slide_start_backup',
                'slide_get_snapshots',
                'slide_create_virtual_machine'
              ]
            }));
          } else {
            // Unknown method
            console.log(JSON.stringify({
              jsonrpc: '2.0',
              id,
              error: {
                code: -32601,
                message: `Method not found: ${method}`
              }
            }));
          }
        } else {
          // Legacy MCP message format
          const { name, arguments: args } = message;
          
          if (name && args) {
            try {
              const result = await handleToolCall(slideClient, name, args);
              console.log(JSON.stringify({
                jsonrpc: '2.0',
                id: message.id || 0,
                result
              }));
            } catch (error) {
              logToFile(`Tool execution error: ${error.message}`);
              console.log(JSON.stringify({
                jsonrpc: '2.0',
                id: message.id || 0,
                error: {
                  code: -32000,
                  message: error.message
                }
              }));
            }
          } else {
            logToFile('Invalid message format');
            console.log(JSON.stringify({
              jsonrpc: '2.0',
              id: message.id || 0,
              error: {
                code: -32600,
                message: 'Invalid message format'
              }
            }));
          }
        }
      } catch (error) {
        logToFile(`Error processing message: ${error.message}`);
        // Generic error response with fallback ID to ensure we always have required fields
        console.log(JSON.stringify({
          jsonrpc: '2.0',
          id: 0,
          error: {
            code: -32603,
            message: `Internal error: ${error.message}`
          }
        }));
      }
    });
    
    // Handle end of input
    rl.on('close', () => {
      logToFile('Input stream closed, exiting');
      process.exit(0);
    });
    
    // Keep the process alive
    process.stdin.resume();
    
    // Log that we're ready
    logToFile('MCP server ready to process requests');
  } catch (error) {
    logToFile(`Fatal error: ${error.message}`);
    process.exit(1);
  }
}

module.exports = {
  SlideClient,
  createClient
}; 
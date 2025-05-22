#!/usr/bin/env node

const readline = require('readline');
const fs = require('fs');
const { createClient } = require('./index');

// For debugging
function logToFile(message) {
  fs.appendFileSync('/tmp/slide-mcp-test.log', new Date().toISOString() + ' ' + message + '\n');
}

// Create Slide client using API key from environment
const apiKey = process.env.SLIDE_API_KEY;
if (!apiKey) {
  logToFile('ERROR: No API key found');
  process.exit(1);
}

const slideClient = createClient(apiKey);

// Define tools
const tools = [
  {
    name: 'devices_list',
    description: 'Get a list of Slide devices',
    inputSchema: {
      type: 'object',
      properties: {
        limit: { 
          type: 'integer',
          description: 'Maximum number of devices to return'
        },
        offset: { 
          type: 'integer',
          description: 'Starting index for pagination'
        }
      }
    }
  },
  {
    name: 'devices_get',
    description: 'Get details of a specific device by ID',
    inputSchema: {
      type: 'object',
      properties: {
        device_id: { 
          type: 'string',
          description: 'ID of the device to retrieve'
        }
      },
      required: ['device_id']
    }
  }
];

// Create interface for reading line by line
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

// Handle JSON-RPC requests
rl.on('line', async (line) => {
  if (!line.trim()) return;
  
  try {
    logToFile(`Received: ${line}`);
    const request = JSON.parse(line);
    
    if (!request.jsonrpc || request.jsonrpc !== '2.0') {
      logToFile('Not a JSON-RPC 2.0 request');
      return;
    }
    
    const id = request.id;
    const method = request.method;
    const params = request.params || {};
    
    // Skip handling notifications (methods without an id)
    if (method && method.startsWith('notifications/')) {
      logToFile(`Received notification: ${method}`);
      return; // Don't respond to notifications
    }
    
    // Initialize - respond with capabilities
    if (method === 'initialize') {
      const response = {
        jsonrpc: '2.0',
        id,
        result: {
          protocolVersion: '2025-03-26',
          serverInfo: {
            name: 'slideMCP',
            version: '1.0.0'
          },
          capabilities: {
            runTool: true,
            tools: {
              get: true,
              run: true
            }
          },
          tools: tools
        }
      };
      
      logToFile(`Sending initialize response: ${JSON.stringify(response)}`);
      console.log(JSON.stringify(response));
      return;
    }
    
    // Handle tools.list request
    if (method === 'tools/list') {
      const response = {
        jsonrpc: '2.0',
        id,
        result: {
          tools: tools
        }
      };
      
      logToFile(`Sending tools list response: ${JSON.stringify(response)}`);
      console.log(JSON.stringify(response));
      return;
    }
    
    // Handle direct tool calls - slide_get_devices
    if (method === 'slide_get_devices' || method === 'devices_list') {
      try {
        logToFile(`Direct tool call: ${method} with args: ${JSON.stringify(params)}`);
        const result = await slideClient.getDevices(params);
        const response = {
          jsonrpc: '2.0',
          id,
          result
        };
        
        logToFile(`Sending ${method} response: ${JSON.stringify(response)}`);
        console.log(JSON.stringify(response));
        return;
      } catch (error) {
        logToFile(`Error in ${method}: ${error.message}`);
        const response = {
          jsonrpc: '2.0',
          id,
          error: {
            code: -32000,
            message: error.message
          }
        };
        
        console.log(JSON.stringify(response));
        return;
      }
    }
    
    // Handle direct tool calls - slide_get_device
    if (method === 'slide_get_device' || method === 'devices_get') {
      try {
        logToFile(`Direct tool call: ${method} with args: ${JSON.stringify(params)}`);
        const result = await slideClient.getDevice(params.device_id);
        const response = {
          jsonrpc: '2.0',
          id,
          result
        };
        
        logToFile(`Sending ${method} response: ${JSON.stringify(response)}`);
        console.log(JSON.stringify(response));
        return;
      } catch (error) {
        logToFile(`Error in ${method}: ${error.message}`);
        const response = {
          jsonrpc: '2.0',
          id,
          error: {
            code: -32000,
            message: error.message
          }
        };
        
        console.log(JSON.stringify(response));
        return;
      }
    }
    
    // Handle MCP prefixed tool calls - mcp_slideMCP_slide_get_devices
    if (method === 'mcp_slideMCP_slide_get_devices' || method === 'mcp_slideMCP_devices_list') {
      try {
        logToFile(`MCP tool call: ${method} with args: ${JSON.stringify(params)}`);
        const result = await slideClient.getDevices(params);
        const response = {
          jsonrpc: '2.0',
          id,
          result
        };
        
        logToFile(`Sending ${method} response: ${JSON.stringify(response)}`);
        console.log(JSON.stringify(response));
        return;
      } catch (error) {
        logToFile(`Error in ${method}: ${error.message}`);
        const response = {
          jsonrpc: '2.0',
          id,
          error: {
            code: -32000,
            message: error.message
          }
        };
        
        console.log(JSON.stringify(response));
        return;
      }
    }
    
    // Handle MCP prefixed tool calls - mcp_slideMCP_slide_get_device
    if (method === 'mcp_slideMCP_slide_get_device' || method === 'mcp_slideMCP_devices_get') {
      try {
        logToFile(`MCP tool call: ${method} with args: ${JSON.stringify(params)}`);
        const result = await slideClient.getDevice(params.device_id);
        const response = {
          jsonrpc: '2.0',
          id,
          result
        };
        
        logToFile(`Sending ${method} response: ${JSON.stringify(response)}`);
        console.log(JSON.stringify(response));
        return;
      } catch (error) {
        logToFile(`Error in ${method}: ${error.message}`);
        const response = {
          jsonrpc: '2.0',
          id,
          error: {
            code: -32000,
            message: error.message
          }
        };
        
        console.log(JSON.stringify(response));
        return;
      }
    }
    
    // Handle MCP prefixed tool calls - mcp_slideMCP_devices_list
    if (method === 'mcp_slideMCP_devices_list') {
      try {
        logToFile(`MCP tool call: ${method} with args: ${JSON.stringify(params)}`);
        const result = await slideClient.getDevices(params);
        const response = {
          jsonrpc: '2.0',
          id,
          result
        };
        
        logToFile(`Sending ${method} response: ${JSON.stringify(response)}`);
        console.log(JSON.stringify(response));
        return;
      } catch (error) {
        logToFile(`Error in ${method}: ${error.message}`);
        const response = {
          jsonrpc: '2.0',
          id,
          error: {
            code: -32000,
            message: error.message
          }
        };
        
        console.log(JSON.stringify(response));
        return;
      }
    }
    
    // Handle MCP prefixed tool calls - mcp_slideMCP_devices_get
    if (method === 'mcp_slideMCP_devices_get') {
      try {
        logToFile(`MCP tool call: ${method} with args: ${JSON.stringify(params)}`);
        const result = await slideClient.getDevice(params.device_id);
        const response = {
          jsonrpc: '2.0',
          id,
          result
        };
        
        logToFile(`Sending ${method} response: ${JSON.stringify(response)}`);
        console.log(JSON.stringify(response));
        return;
      } catch (error) {
        logToFile(`Error in ${method}: ${error.message}`);
        const response = {
          jsonrpc: '2.0',
          id,
          error: {
            code: -32000,
            message: error.message
          }
        };
        
        console.log(JSON.stringify(response));
        return;
      }
    }
    
    // Run tool
    if (method === 'run') {
      try {
        const { name, arguments: args } = params;
        let result;
        
        if (name === 'slide_get_devices' || name === 'devices_list' || 
            name === 'mcp_slideMCP_slide_get_devices' || name === 'mcp_slideMCP_devices_list') {
          result = await slideClient.getDevices(args || {});
        } else if (name === 'slide_get_device' || name === 'devices_get' ||
                   name === 'mcp_slideMCP_slide_get_device' || name === 'mcp_slideMCP_devices_get') {
          result = await slideClient.getDevice(args.device_id);
        } else {
          throw new Error(`Unknown tool: ${name}`);
        }
        
        const response = {
          jsonrpc: '2.0',
          id,
          result
        };
        
        logToFile(`Sending run response: ${JSON.stringify(response)}`);
        console.log(JSON.stringify(response));
        return;
      } catch (error) {
        const response = {
          jsonrpc: '2.0',
          id,
          error: {
            code: -32000,
            message: error.message
          }
        };
        
        logToFile(`Sending error response: ${JSON.stringify(response)}`);
        console.log(JSON.stringify(response));
        return;
      }
    }
    
    // Handle missing id (notifications)
    if (id === undefined) {
      logToFile(`Ignoring notification method: ${method}`);
      return;
    }
    
    // Any other method - respond with method not found
    const response = {
      jsonrpc: '2.0',
      id,
      error: {
        code: -32601,
        message: `Method not found: ${method}`
      }
    };
    
    logToFile(`Sending method not found response: ${JSON.stringify(response)}`);
    console.log(JSON.stringify(response));
    
  } catch (error) {
    // Parse error or other unexpected error
    logToFile(`Error: ${error.message}`);
    console.log(JSON.stringify({
      jsonrpc: '2.0',
      id: null,
      error: {
        code: -32700,
        message: `Parse error: ${error.message}`
      }
    }));
  }
});

// Handle end of input
rl.on('close', () => {
  logToFile('Input stream closed');
  process.exit(0);
});

// Keep the process alive
process.stdin.resume();
logToFile('MCP test server ready'); 
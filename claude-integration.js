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
  }
];

// Tool execution handler - this would be connected to Claude's tool calling
async function handleToolCall(toolName, args) {
  console.log(`Claude called tool: ${toolName} with args:`, args);
  
  switch(toolName) {
    case "slide_get_devices":
      return await slideClient.getDevices(args);
    
    case "slide_get_agents":
      return await slideClient.getAgents(args);
    
    case "slide_start_backup":
      return await slideClient.startBackup(args);
    
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
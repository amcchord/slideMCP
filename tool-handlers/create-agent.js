#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_create_agent
 * This script is executed by Claude Desktop when the slide_create_agent tool is called
 */

require('dotenv').config({ path: require('path').resolve(__dirname, '../.env') });
const { createClient } = require('../index');

// Parse input arguments from Claude Desktop
const args = process.argv.length > 2 ? JSON.parse(process.argv[2]) : {};

// Get API key from environment variables
const apiKey = process.env.SLIDE_API_KEY;

if (!apiKey) {
  console.error('ERROR: No API key found. Please create a .env file with SLIDE_API_KEY');
  process.exit(1);
}

// Validate required parameters
if (!args.display_name) {
  console.error('ERROR: display_name is required');
  process.exit(1);
}

if (!args.device_id) {
  console.error('ERROR: device_id is required');
  process.exit(1);
}

// Create agent data object
const agentData = {
  display_name: args.display_name,
  device_id: args.device_id
};

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function createAgent() {
  try {
    const result = await slideClient.createAgentPair(agentData);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      success: true,
      agent_id: result.agent_id,
      device_id: args.device_id,
      display_name: args.display_name,
      pair_code: result.pair_code,
      install_instructions: {
        windows: result.install_urls?.windows || "No Windows install URL provided",
        linux: result.install_urls?.linux || "No Linux install URL provided",
        mac: result.install_urls?.mac || "No Mac install URL provided"
      }
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
createAgent(); 
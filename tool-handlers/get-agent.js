#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_agent
 * This script is executed by Claude Desktop when the slide_get_agent tool is called
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
if (!args.agent_id) {
  console.error('ERROR: agent_id is required');
  process.exit(1);
}

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function getAgent() {
  try {
    const result = await slideClient.getAgent(args.agent_id);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      id: result.agent_id,
      device_id: result.device_id,
      name: result.display_name || result.hostname || 'Unnamed Agent',
      hostname: result.hostname,
      platform: result.platform,
      os: result.os,
      os_version: result.os_version,
      agent_version: result.agent_version,
      addresses: result.addresses,
      last_seen: result.last_seen_at
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
getAgent(); 
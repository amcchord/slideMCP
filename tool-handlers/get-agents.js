#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_agents
 * This script is executed by Claude Desktop when the slide_get_agents tool is called
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

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function getAgents() {
  try {
    const result = await slideClient.getAgents(args);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      agents: result.data.map(agent => ({
        id: agent.agent_id,
        device_id: agent.device_id,
        hostname: agent.hostname,
        os: agent.platform || `${agent.os} ${agent.os_version}`,
        ip_addresses: agent.ip_addresses,
        last_seen: agent.last_seen_at
      })),
      total: result.pagination.total
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
getAgents(); 
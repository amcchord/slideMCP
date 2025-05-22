#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_network
 * This script is executed by Claude Desktop when the slide_get_network tool is called
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
if (!args.network_id) {
  console.error('ERROR: network_id is required');
  process.exit(1);
}

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function getNetwork() {
  try {
    const result = await slideClient.getNetwork(args.network_id);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      id: result.network_id,
      name: result.name,
      type: result.type,
      client_id: result.client_id,
      bridge_device_id: result.bridge_device_id,
      router_prefix: result.router_prefix,
      dhcp: result.dhcp,
      dhcp_range: result.dhcp ? {
        start: result.dhcp_range_start,
        end: result.dhcp_range_end
      } : null,
      nameservers: result.nameservers,
      internet: result.internet,
      wg: result.wg,
      comments: result.comments
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
getNetwork(); 
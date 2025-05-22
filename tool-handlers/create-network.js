#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_create_network
 * This script is executed by Claude Desktop when the slide_create_network tool is called
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
if (!args.name) {
  console.error('ERROR: name is required');
  process.exit(1);
}

if (!args.type) {
  console.error('ERROR: type is required');
  process.exit(1);
}

// Create network data object
const networkData = {
  name: args.name,
  type: args.type
};

// Add optional parameters if provided
if (args.client_id) networkData.client_id = args.client_id;
if (args.comments) networkData.comments = args.comments;
if (args.bridge_device_id) networkData.bridge_device_id = args.bridge_device_id;
if (args.router_prefix) networkData.router_prefix = args.router_prefix;
if (args.dhcp !== undefined) networkData.dhcp = args.dhcp;
if (args.dhcp_range_start) networkData.dhcp_range_start = args.dhcp_range_start;
if (args.dhcp_range_end) networkData.dhcp_range_end = args.dhcp_range_end;
if (args.nameservers) networkData.nameservers = args.nameservers;
if (args.internet !== undefined) networkData.internet = args.internet;

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function createNetwork() {
  try {
    const result = await slideClient.createNetwork(networkData);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      success: true,
      network_id: result.network_id,
      name: args.name,
      type: args.type
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
createNetwork(); 
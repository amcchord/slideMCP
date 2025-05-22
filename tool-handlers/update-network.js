#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_update_network
 * This script is executed by Claude Desktop when the slide_update_network tool is called
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

// Create update data object
const updateData = {};

// Add optional parameters if provided
if (args.name) updateData.name = args.name;
if (args.comments !== undefined) updateData.comments = args.comments;
if (args.router_prefix) updateData.router_prefix = args.router_prefix;
if (args.dhcp !== undefined) updateData.dhcp = args.dhcp;
if (args.dhcp_range_start) updateData.dhcp_range_start = args.dhcp_range_start;
if (args.dhcp_range_end) updateData.dhcp_range_end = args.dhcp_range_end;
if (args.nameservers) updateData.nameservers = args.nameservers;
if (args.internet !== undefined) updateData.internet = args.internet;
if (args.wg !== undefined) updateData.wg = args.wg;

// Check if we have data to update
if (Object.keys(updateData).length === 0) {
  console.error('ERROR: At least one updateable parameter must be provided');
  process.exit(1);
}

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function updateNetwork() {
  try {
    const result = await slideClient.updateNetwork(args.network_id, updateData);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      success: true,
      network_id: args.network_id,
      updated: updateData
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
updateNetwork(); 
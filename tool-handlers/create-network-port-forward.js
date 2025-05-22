#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_create_network_port_forward
 * This script is executed by Claude Desktop when the slide_create_network_port_forward tool is called
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

if (!args.proto) {
  console.error('ERROR: proto is required');
  process.exit(1);
}

if (!args.dest) {
  console.error('ERROR: dest is required');
  process.exit(1);
}

// Create port forward data object
const portForwardData = {
  proto: args.proto,
  dest: args.dest
};

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function createNetworkPortForward() {
  try {
    const result = await slideClient.createNetworkPortForward(args.network_id, portForwardData);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      success: true,
      network_id: args.network_id,
      port_forward: {
        proto: args.proto,
        dest: args.dest,
        port: result.port
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
createNetworkPortForward(); 
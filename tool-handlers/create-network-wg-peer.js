#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_create_network_wg_peer
 * This script is executed by Claude Desktop when the slide_create_network_wg_peer tool is called
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

if (!args.peer_name) {
  console.error('ERROR: peer_name is required');
  process.exit(1);
}

// Create WG peer data object
const wgPeerData = {
  peer_name: args.peer_name
};

// Add optional parameters if provided
if (args.remote_networks) wgPeerData.remote_networks = args.remote_networks;

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function createNetworkWGPeer() {
  try {
    const result = await slideClient.createNetworkWGPeer(args.network_id, wgPeerData);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      success: true,
      network_id: args.network_id,
      wg_peer: {
        peer_name: args.peer_name,
        wg_address: result.wg_address,
        wg_public_key: result.wg_public_key,
        wg_private_key: result.wg_private_key,
        wg_endpoint: result.wg_endpoint,
        wg_allowed_ips: result.wg_allowed_ips
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
createNetworkWGPeer(); 
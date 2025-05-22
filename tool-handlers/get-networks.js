#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_networks
 * This script is executed by Claude Desktop when the slide_get_networks tool is called
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
async function getNetworks() {
  try {
    const result = await slideClient.getNetworks(args);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      networks: result.data.map(network => ({
        id: network.network_id,
        name: network.name,
        type: network.type,
        client_id: network.client_id,
        bridge_device_id: network.bridge_device_id,
        router_prefix: network.router_prefix,
        dhcp: network.dhcp,
        dhcp_range: network.dhcp ? {
          start: network.dhcp_range_start,
          end: network.dhcp_range_end
        } : null,
        nameservers: network.nameservers,
        internet: network.internet,
        wg: network.wg,
        comments: network.comments
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
getNetworks(); 
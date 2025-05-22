#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_create_virtual_machine
 * This script is executed by Claude Desktop when the slide_create_virtual_machine tool is called
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
async function createVirtualMachine() {
  try {
    // Required parameters check
    if (!args.snapshot_id) {
      throw new Error('snapshot_id is required');
    }
    if (!args.device_id) {
      throw new Error('device_id is required');
    }

    // Create payload
    const payload = {
      snapshot_id: args.snapshot_id,
      device_id: args.device_id
    };
    
    // Optional parameters
    if (args.cpu_count) payload.cpu_count = args.cpu_count;
    if (args.memory_in_mb) payload.memory_in_mb = args.memory_in_mb;
    if (args.disk_bus) payload.disk_bus = args.disk_bus;
    if (args.network_model) payload.network_model = args.network_model;
    if (args.network_type) payload.network_type = args.network_type;
    if (args.network_source) payload.network_source = args.network_source;
    if (args.boot_mods) payload.boot_mods = args.boot_mods;

    const result = await slideClient.createVirtualMachine(payload);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      virt_id: result.virt_id,
      snapshot_id: result.snapshot_id,
      device_id: result.device_id,
      state: result.state,
      created_at: result.created_at
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
createVirtualMachine(); 
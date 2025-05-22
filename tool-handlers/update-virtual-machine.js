#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_update_virtual_machine
 * This script is executed by Claude Desktop when the slide_update_virtual_machine tool is called
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
async function updateVirtualMachine() {
  try {
    // Required parameters check
    if (!args.virt_id) {
      throw new Error('virt_id is required');
    }
    
    // Create payload - at least one field must be present
    const payload = {};
    if (args.state) payload.state = args.state;
    if (args.expires_at) payload.expires_at = args.expires_at;
    if (args.memory_in_mb) payload.memory_in_mb = args.memory_in_mb;
    if (args.cpu_count) payload.cpu_count = args.cpu_count;
    
    if (Object.keys(payload).length === 0) {
      throw new Error('At least one of state, expires_at, memory_in_mb, or cpu_count is required');
    }

    const result = await slideClient.updateVirtualMachine(args.virt_id, payload);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      virt_id: result.virt_id,
      state: result.state,
      memory_mb: result.memory_in_mb,
      cpu_count: result.cpu_count,
      expires_at: result.expires_at,
      updated_at: result.updated_at
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
updateVirtualMachine(); 
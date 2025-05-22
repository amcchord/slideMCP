#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_device
 * This script is executed by Claude Desktop when the slide_get_device tool is called
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
if (!args.device_id) {
  console.error('ERROR: device_id is required');
  process.exit(1);
}

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function getDevice() {
  try {
    const result = await slideClient.getDevice(args.device_id);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      id: result.device_id,
      name: result.display_name || result.hostname || 'Unnamed Device',
      hostname: result.hostname,
      model: result.hardware_model_name,
      image_version: result.image_version,
      package_version: result.package_version,
      storage: {
        used: result.storage_used_bytes,
        total: result.storage_total_bytes
      },
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
getDevice(); 
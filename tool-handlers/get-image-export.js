#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_image_export
 * This script is executed by Claude Desktop when the slide_get_image_export tool is called
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
async function getImageExport() {
  try {
    // Required parameters check
    if (!args.image_export_id) {
      throw new Error('image_export_id is required');
    }

    const result = await slideClient.getImageExport(args.image_export_id);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      id: result.image_export_id,
      snapshot_id: result.snapshot_id,
      device_id: result.device_id,
      image_type: result.image_type,
      status: result.status,
      created_at: result.created_at,
      download_url: result.download_url,
      file_size: result.file_size,
      boot_mods: result.boot_mods
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
getImageExport(); 
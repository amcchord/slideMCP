#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_image_exports
 * This script is executed by Claude Desktop when the slide_get_image_exports tool is called
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
async function getImageExports() {
  try {
    // Optional parameters
    const params = {};
    if (args.limit) params.limit = args.limit;
    if (args.offset) params.offset = args.offset;
    if (args.sort_by) params.sort_by = args.sort_by;
    if (args.sort_asc) params.sort_asc = args.sort_asc;

    const result = await slideClient.getImageExports(params);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      image_exports: result.data.map(export_item => ({
        id: export_item.image_export_id,
        snapshot_id: export_item.snapshot_id,
        device_id: export_item.device_id,
        image_type: export_item.image_type,
        status: export_item.status,
        created_at: export_item.created_at
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
getImageExports(); 
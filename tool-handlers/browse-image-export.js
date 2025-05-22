#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_browse_image_export
 * This script is executed by Claude Desktop when the slide_browse_image_export tool is called
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
async function browseImageExport() {
  try {
    // Required parameters check
    if (!args.image_export_id) {
      throw new Error('image_export_id is required');
    }

    // Optional parameters
    const params = {};
    if (args.limit) params.limit = args.limit;
    if (args.offset) params.offset = args.offset;

    const result = await slideClient.browseImageExport(args.image_export_id, params);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      files: result.data.map(file => ({
        name: file.name,
        path: file.path,
        size: file.size,
        type: file.type,
        modified: file.modified
      })),
      total: result.pagination?.total || result.data.length
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
browseImageExport(); 
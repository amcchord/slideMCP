#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_browse_file_restore
 * This script is executed by Claude Desktop when the slide_browse_file_restore tool is called
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
async function browseFileRestore() {
  try {
    // Required parameters check
    if (!args.file_restore_id) {
      throw new Error('file_restore_id is required');
    }

    // Optional parameters
    const params = {};
    if (args.limit) params.limit = args.limit;
    if (args.offset) params.offset = args.offset;
    if (args.path) params.path = args.path;

    const result = await slideClient.browseFileRestore(args.file_restore_id, params);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      files: result.data.map(file => ({
        name: file.name,
        path: file.path,
        size: file.size,
        type: file.type,
        modified: file.modified
      })),
      current_path: result.current_path || '/',
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
browseFileRestore(); 
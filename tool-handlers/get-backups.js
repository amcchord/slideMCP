#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_backups
 * This script is executed by Claude Desktop when the slide_get_backups tool is called
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
async function getBackups() {
  try {
    const result = await slideClient.getBackups(args);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      backups: result.data.map(backup => ({
        id: backup.backup_id,
        agent_id: backup.agent_id,
        snapshot_id: backup.snapshot_id,
        started_at: backup.started_at,
        ended_at: backup.ended_at,
        status: backup.status,
        error: backup.error_message ? {
          code: backup.error_code,
          message: backup.error_message
        } : null
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
getBackups(); 
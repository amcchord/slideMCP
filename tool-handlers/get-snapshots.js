#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_snapshots
 * This script is executed by Claude Desktop when the slide_get_snapshots tool is called
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
async function getSnapshots() {
  try {
    const result = await slideClient.getSnapshots(args);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      snapshots: result.data.map(snapshot => ({
        id: snapshot.snapshot_id,
        agent_id: snapshot.agent_id,
        backup_started_at: snapshot.backup_started_at,
        backup_ended_at: snapshot.backup_ended_at,
        locations: snapshot.locations,
        verify_boot_status: snapshot.verify_boot_status,
        verify_fs_status: snapshot.verify_fs_status,
        deleted: snapshot.deleted
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
getSnapshots(); 
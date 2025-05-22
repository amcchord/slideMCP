#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_update_alert
 * This script is executed by Claude Desktop when the slide_update_alert tool is called
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
async function updateAlert() {
  try {
    // Required parameters check
    if (!args.alert_id) {
      throw new Error('alert_id is required');
    }
    
    // Create payload
    const payload = {};
    if (args.resolved !== undefined) payload.resolved = args.resolved;
    
    if (Object.keys(payload).length === 0) {
      throw new Error('At least one of resolved is required');
    }

    const result = await slideClient.updateAlert(args.alert_id, payload);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      id: result.alert_id,
      resolved: result.resolved,
      resolved_at: result.resolved_at,
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
updateAlert(); 
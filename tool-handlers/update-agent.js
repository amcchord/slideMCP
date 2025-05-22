#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_update_agent
 * This script is executed by Claude Desktop when the slide_update_agent tool is called
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
if (!args.agent_id) {
  console.error('ERROR: agent_id is required');
  process.exit(1);
}

// Create update data object
const updateData = {};
if (args.display_name) updateData.display_name = args.display_name;

// Check if we have data to update
if (Object.keys(updateData).length === 0) {
  console.error('ERROR: display_name must be provided');
  process.exit(1);
}

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function updateAgent() {
  try {
    const result = await slideClient.updateAgent(args.agent_id, updateData);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      success: true,
      updated: updateData,
      agent_id: args.agent_id
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
updateAgent(); 
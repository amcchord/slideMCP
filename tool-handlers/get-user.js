#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_user
 * This script is executed by Claude Desktop when the slide_get_user tool is called
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
async function getUser() {
  try {
    // Required parameters check
    if (!args.user_id) {
      throw new Error('user_id is required');
    }

    const result = await slideClient.getUser(args.user_id);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      id: result.user_id,
      name: result.name,
      email: result.email,
      role: result.role,
      created_at: result.created_at,
      last_login_at: result.last_login_at
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
getUser(); 
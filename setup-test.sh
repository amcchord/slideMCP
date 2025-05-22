#!/bin/bash

# setup-test.sh
# This script creates a temporary .env file for testing the Slide API client
# and cleans it up afterward to prevent accidental commits

# Check if API key was provided
if [ $# -eq 0 ]; then
    echo "Error: No API key provided"
    echo "Usage: ./setup-test.sh YOUR_API_KEY"
    exit 1
fi

API_KEY=$1

# Create .env file with the API key
echo "Creating temporary .env file..."
echo "SLIDE_API_KEY=$API_KEY" > .env
echo "SLIDE_API_URL=https://api.slide.tech" >> .env
echo "SLIDE_API_VERSION=v1" >> .env

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "Error: npm is not installed"
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
npm install

# Run the test
echo "Running API test..."
node test.js

# Check if test was successful
TEST_STATUS=$?

# Clean up .env file
echo "Cleaning up temporary .env file..."
rm .env

# Return test status
if [ $TEST_STATUS -eq 0 ]; then
    echo "Test completed successfully!"
    exit 0
else
    echo "Test failed!"
    exit 1
fi 
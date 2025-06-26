#!/bin/bash

echo "=== Slide MCP Server Notarization Setup ==="
echo ""
echo "This script will help you set up notarization for macOS binaries."
echo ""
echo "You need:"
echo "1. Your Apple ID email address"
echo "2. An app-specific password from https://appleid.apple.com/"
echo "3. Your Team ID: 7PTN7E8EDS"
echo ""

read -p "Enter your Apple ID email: " APPLE_ID
read -s -p "Enter your app-specific password: " APP_PASSWORD
echo ""

echo "Setting up keychain profile..."
xcrun notarytool store-credentials \
    --apple-id "$APPLE_ID" \
    --team-id "7PTN7E8EDS" \
    --password "$APP_PASSWORD" \
    slidemcp-notarization

if [ $? -eq 0 ]; then
    echo "✅ Notarization credentials stored successfully!"
    echo "You can now run: make notarize-macos KEYCHAIN_PROFILE=slidemcp-notarization"
else
    echo "❌ Failed to store credentials. Please check your Apple ID and password."
    exit 1
fi 
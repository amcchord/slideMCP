#!/bin/bash

# Test script for documentation context improvements

echo "Testing Documentation Context Improvements"
echo "========================================"
echo

# Test 1: List sections with descriptions
echo "Test 1: Listing documentation sections with descriptions"
echo "--------------------------------------------------------"
curl -s -X POST http://localhost:8000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "slide_docs",
      "arguments": {
        "operation": "list_sections"
      }
    },
    "id": 1
  }' | jq -r '.result.content | fromjson | .sections[] | "\(.name): \(.description)"'

echo
echo

# Test 2: Get topics with descriptions for ambiguous sections
echo "Test 2: Getting topics for 'Slide Console' section"
echo "------------------------------------------------"
curl -s -X POST http://localhost:8000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "slide_docs",
      "arguments": {
        "operation": "get_topics",
        "section": "Slide Console"
      }
    },
    "id": 2
  }' | jq -r '.result.content | fromjson | .topics[] | if .description then "\(.name): \(.description)" else .name end'

echo
echo

# Test 3: Search for "network" to see contextual results
echo "Test 3: Searching for 'network' with contextual results"
echo "------------------------------------------------------"
curl -s -X POST http://localhost:8000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "slide_docs",
      "arguments": {
        "operation": "search_docs",
        "query": "network"
      }
    },
    "id": 3
  }' | jq -r '.result.content | fromjson | .results[] | select(.type == "topic_match") | "Section: \(.section) - \(.section_description)\nTopic: \(.topic)\(.topic_description // "")\n"'

echo "Test complete!" 
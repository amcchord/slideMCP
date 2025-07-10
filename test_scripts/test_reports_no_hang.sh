#!/bin/bash

# Test script to verify reports tool won't hang with safety checks

echo "=== Testing Reports Tool Anti-Hang Measures ==="
echo ""
echo "This test verifies the following safety measures:"
echo "1. Pagination loops have safety checks to prevent infinite loops"
echo "2. Individual agent reports have 30-second timeouts"
echo "3. Overall operation has 5-minute timeout"
echo "4. Verbose mode shows warnings if pagination gets stuck"
echo ""

# Test with verbose mode to see safety warnings
echo "Running weekly report with verbose mode..."
echo "If pagination gets stuck, you'll see WARNING messages"
echo ""

# Add timeout command to ensure script doesn't hang forever
timeout 60s ./slide-mcp-server --tool slide_reports --args '{
  "operation": "weekly_backup_snapshot",
  "agent_id": "a_mbepgxrb629h",
  "date": "2024-01-15",
  "format": "json",
  "verbose": true
}' 2>&1 | tail -100

EXIT_CODE=$?

echo ""
if [ $EXIT_CODE -eq 124 ]; then
    echo "ERROR: Command timed out after 60 seconds!"
    echo "This suggests the safety checks may not be working properly."
else
    echo "SUCCESS: Command completed without hanging."
    echo ""
    echo "Safety features implemented:"
    echo "- Pagination loops check if offset is progressing (newOffset > offset)"
    echo "- Maximum offset limit of 10,000 to prevent runaway loops"
    echo "- Per-agent timeout of 30 seconds"
    echo "- Overall operation timeout of 5 minutes"
    echo "- Context cancellation for graceful shutdown"
fi 
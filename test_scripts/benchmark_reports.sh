#!/bin/bash

# Simple benchmark to demonstrate performance improvements

echo "=== Reports Performance Benchmark ==="
echo ""
echo "Testing report generation performance..."
echo ""

# Test 1: Single agent report
echo "Test 1: Single Agent Daily Report"
echo "--------------------------------"
time ./slide-mcp-server --tool slide_reports --args '{
  "operation": "daily_backup_snapshot",
  "agent_id": "YOUR_AGENT_ID_HERE",
  "date": "2024-01-15",
  "format": "json"
}' > /dev/null 2>&1
echo ""

# Test 2: All agents report (demonstrates parallel processing)
echo "Test 2: All Agents Daily Report (Parallel Processing)"
echo "---------------------------------------------------"
echo "This will process multiple agents concurrently..."
time ./slide-mcp-server --tool slide_reports --args '{
  "operation": "daily_backup_snapshot",
  "date": "2024-01-15",
  "format": "json"
}' > /dev/null 2>&1
echo ""

# Test 3: Weekly report (demonstrates caching benefits)
echo "Test 3: Weekly Report (Cache Benefits)"
echo "------------------------------------"
echo "Processing 7 days of data with caching..."
time ./slide-mcp-server --tool slide_reports --args '{
  "operation": "weekly_backup_snapshot",
  "date": "2024-01-15",
  "format": "json"
}' > /dev/null 2>&1
echo ""

echo "=== Performance Notes ==="
echo "- Single agent reports benefit from parallel backup/snapshot stats"
echo "- Multiple agent reports show dramatic improvement with concurrency"
echo "- Weekly/monthly reports benefit most from caching"
echo "- Actual improvement depends on number of agents and API latency" 
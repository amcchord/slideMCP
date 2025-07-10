#!/bin/bash

# Example of the monthly calendar report

API_KEY="tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"

echo "Example: Monthly Calendar Report"
echo "================================"
echo ""
echo "This example shows the monthly backup report with a visual calendar."
echo "The calendar displays success indicators for each day:"
echo "  ✓ = ≥90% success rate"
echo "  ~ = 50-89% success rate"
echo "  ✗ = <50% success rate"
echo ""

# Get the monthly report with calendar
MONTHLY_REPORT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_reports","arguments":{"operation":"monthly_backup_snapshot","format":"markdown"}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1)

# Extract and display the report up to the Daily Details section
echo "$MONTHLY_REPORT" | jq '.result.content[0].text' -r | sed -n '1,/## Daily Details/p' | head -n -1

echo ""
echo "================================"
echo "Note: The full report includes detailed daily breakdowns below the calendar."
echo "Use this report to quickly identify:"
echo "- Days with backup failures"
echo "- Patterns in backup activity"
echo "- Overall monthly performance" 
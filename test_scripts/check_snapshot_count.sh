#!/bin/bash

# Simple script to check snapshot counts

echo "Checking snapshot counts via API..."

# Count total snapshots
echo -n "Total snapshots in system: "
curl -s -H "Authorization: Bearer $SLIDE_API_KEY" "https://api.slide.tech/v1/snapshot?limit=1" | jq '.pagination.total' 2>/dev/null || echo "Error"

# Count active snapshots (cloud)
echo -n "Active snapshots (cloud): "
curl -s -H "Authorization: Bearer $SLIDE_API_KEY" "https://api.slide.tech/v1/snapshot?limit=1&snapshot_location=exists_cloud" | jq '.pagination.total' 2>/dev/null || echo "Error"

# Count deleted snapshots
echo -n "Deleted snapshots: "
curl -s -H "Authorization: Bearer $SLIDE_API_KEY" "https://api.slide.tech/v1/snapshot?limit=1&snapshot_location=exists_deleted" | jq '.pagination.total' 2>/dev/null || echo "Error"

# Get a sample of snapshots to see data size
echo -e "\nGetting sample snapshot data..."
SAMPLE=$(curl -s -H "Authorization: Bearer $SLIDE_API_KEY" "https://api.slide.tech/v1/snapshot?limit=5" | jq '.')
echo "Sample data size for 5 snapshots: ${#SAMPLE} characters"

# Calculate average size per snapshot
if [ ${#SAMPLE} -gt 0 ]; then
    AVG_SIZE=$((${#SAMPLE} / 5))
    echo "Average size per snapshot: ~$AVG_SIZE characters"
fi 
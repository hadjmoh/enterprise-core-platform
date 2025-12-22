#!/bin/bash
# Build and package connector apps

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONNECTORS_DIR="$SCRIPT_DIR/../sample_apps/connectors"
OUTPUT_DIR="$SCRIPT_DIR/../sample_apps"

echo "Building connector packages..."

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Package each connector
for connector in aws-cloudtrail-connector slack-connector github-connector jira-connector; do
    echo "Packaging $connector..."
    
    cd "$CONNECTORS_DIR"
    
    # Validate manifest exists
    if [ ! -f "$connector/prospect.yaml" ]; then
        echo "ERROR: Missing prospect.yaml for $connector"
        exit 1
    fi
    
    # Create tarball
    tar -czf "$OUTPUT_DIR/$connector.tar.gz" "$connector"
    
    echo "✓ Created $connector.tar.gz"
done

echo ""
echo "Package Summary:"
ls -lh "$OUTPUT_DIR"/*.tar.gz

echo ""
echo "✓ All connectors packaged successfully!"

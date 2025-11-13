#!/bin/bash
# Watch for changes and restart MCP server
# This provides a fast feedback loop for development

set -e

cd "$(dirname "$0")/.."

echo "==> Watching for Go file changes..."
echo "==> Will run tests on each change"
echo ""

# Use watchexec if available, otherwise fall back to basic loop
if command -v watchexec &> /dev/null; then
  watchexec -cr -e go -w . -i cmd/test_client -- go run ./cmd/test_client greet '{"name":"Developer"}'
else
  echo "Warning: watchexec not found. Install with: brew install watchexec"
  echo "Running without watch mode..."
  go run ./cmd/test_client greet '{"name":"Developer"}'
fi


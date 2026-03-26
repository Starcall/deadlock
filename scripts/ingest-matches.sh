#!/bin/bash
# Ingest match data from deadlock-api.com
set -euo pipefail
cd "$(dirname "$0")/.."
COUNT=${1:-1000}
echo "Ingesting $COUNT matches..."
go run ./cmd/ingest --count="$COUNT"
echo "Done."

#!/bin/bash
# Seed heroes and items from deadlock-api.com into the database
set -euo pipefail
cd "$(dirname "$0")/.."
echo "Seeding assets..."
go run ./cmd/ingest --count=0
echo "Done."

#!/bin/bash
# Train model and compute WPA
set -euo pipefail
cd "$(dirname "$0")/.."
echo "Training model and computing WPA..."
go run ./cmd/compute
echo "Done."

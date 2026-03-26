# Deadlock WPA Analytics Tool

## What This Repo Does
Win Probability Added (WPA) analytics dashboard for Valve's Deadlock game. Computes debiased item effectiveness metrics by measuring the change in estimated win probability after item purchases, contextualized by rank, game phase, and hero.

## Architecture
- **Go backend**: REST API server (`cmd/server`), data ingestion (`cmd/ingest`), model training + WPA computation (`cmd/compute`)
- **SQLite database**: Single-file persistence for matches, snapshots, and pre-computed WPA results
- **Next.js frontend**: Dashboard at `web/` showing item WPA tables, charts, and model diagnostics

## Data Flow
1. `cmd/ingest` fetches match data from deadlock-api.com → SQLite
2. `cmd/compute` trains logistic regression model on game state snapshots, then computes WPA per hero/item/context → SQLite
3. `cmd/server` serves pre-computed WPA data via REST API
4. Next.js frontend consumes the API and renders the dashboard

## Key Directories
- `cmd/` — CLI entry points (server, ingest, compute)
- `internal/api/` — REST API handlers
- `internal/deadlockapi/` — deadlock-api.com HTTP client
- `internal/model/` — Win probability model (logistic regression, calibration)
- `internal/wpa/` — WPA computation engine
- `internal/store/` — SQLite persistence layer
- `internal/domain/` — Shared domain types
- `web/` — Next.js frontend dashboard

## Commands
```bash
# Backend
go run ./cmd/ingest --count=1000      # Ingest match data
go run ./cmd/compute                   # Train model + compute WPA
go run ./cmd/server                    # Start API server on :8080

# Frontend
cd web && npm run dev                  # Start dev server on :3000
cd web && npm run build                # Production build
```

## Toolchain
- Go 1.22+ with modernc.org/sqlite (pure Go), gonum
- Node 20+ / npm for frontend
- Next.js 14, TypeScript, Tailwind CSS, Recharts

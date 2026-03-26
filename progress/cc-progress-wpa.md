# Deadlock WPA Analytics — Progress

## Current State
All phases complete. Project builds cleanly (Go + Next.js).

## Completed
- [x] Project scaffolding (directories, git, config files)
- [x] Domain types (internal/domain/)
- [x] deadlock-api.com client (internal/deadlockapi/)
- [x] SQLite store with migrations (internal/store/)
- [x] Data ingestion CLI (cmd/ingest/)
- [x] Win probability model (internal/model/)
- [x] WPA engine + compute CLI (cmd/compute/, internal/wpa/)
- [x] REST API server (cmd/server/, internal/api/)
- [x] Next.js frontend dashboard (web/)
- [x] Scripts, documentation, polish

## Architecture
- Go backend with 3 CLI commands (server, ingest, compute)
- SQLite database (single file, zero-config)
- Logistic regression model (67 features) with Platt scaling calibration
- Next.js 14 frontend dashboard with Recharts

## Usage
```bash
go run ./cmd/ingest --count=1000   # Ingest match data
go run ./cmd/compute               # Train model + compute WPA
go run ./cmd/server                # Start API on :8080
cd web && npm run dev              # Start frontend on :3000
```

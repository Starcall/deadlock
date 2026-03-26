# Deadlock WPA Analytics

Win Probability Added (WPA) analytics dashboard for Valve's Deadlock. Instead of raw win rates (biased by when items are bought), WPA measures the **change in estimated win probability** after item purchases — giving a debiased view of which items genuinely help each hero win.

## Quick Start

### Prerequisites
- Go 1.22+
- Node.js 20+
- npm

### Setup

```bash
# Clone and configure
cp .env.example .env

# Install Go dependencies
go mod download

# Ingest match data (start with 1000 for testing)
go run ./cmd/ingest --count=1000

# Train model and compute WPA
go run ./cmd/compute

# Start the API server
go run ./cmd/server
```

In a separate terminal:

```bash
# Start the frontend
cd web
npm install
npm run dev
```

Open http://localhost:3000 to view the dashboard.

## How It Works

### Win Probability Model
A logistic regression model estimates `P(win | game_state)` using features like team net worth difference, kill/death/assist differences, and per-player stats. Items are **excluded** from features to avoid circular evaluation.

### WPA Computation
For each item purchase, WPA = `w(state_after) - w(state_before)`, where `w` is the win probability model. This is averaged across all purchases of each item by each hero, segmented by rank bracket and game phase.

### Key Metrics
| Metric | Meaning |
|--------|---------|
| **ΔW̄** (Mean WPA) | Average win probability change after buying the item |
| **W̄** (Mean Initial W) | Average win probability when the item is bought (reveals selection bias) |
| **Win Rate** | Raw outcome rate (for comparison) |
| **K** (Sample Size) | Number of purchases observed |

## Architecture

```
Next.js Dashboard (:3000) ←→ Go REST API (:8080) ←→ SQLite Database
                                    ↑
                            Offline batch jobs
                            (ingest, compute)
                                    ↑
                          deadlock-api.com
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/heroes` | All heroes |
| `GET /api/items` | All items |
| `GET /api/wpa/hero/:heroId` | Item WPA table for a hero |
| `GET /api/wpa/hero/:heroId/item/:itemId` | Detailed WPA across contexts |
| `GET /api/model/stats` | Model accuracy and calibration |
| `GET /api/status` | Data freshness |

## Data Source

Match data is sourced from [deadlock-api.com](https://deadlock-api.com) (no auth required).

## Methodology

Based on the win probability estimation methodology described in xPetu's thesis on WPA in esports, adapted for Deadlock's team-based mechanics.

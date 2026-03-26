# Deadlock WPA Analytics

Win Probability Added (WPA) analytics dashboard for Valve's Deadlock. Instead of raw win rates (biased by when items are bought), WPA measures the **change in estimated win probability** after item purchases — giving a debiased view of which items genuinely help each hero win.

## Architecture

```
Next.js Dashboard (:3000) ←→ Go REST API (:8080) ←→ SQLite Database
                                    ↑
                            Offline batch jobs
                            (ingest, compute)
                                    ↑
                          deadlock-api.com (public, no auth)
```

**Components:**
- `cmd/server` — Go REST API serving pre-computed WPA data
- `cmd/ingest` — Fetches ranked match data from deadlock-api.com into SQLite
- `cmd/compute` — Trains logistic regression win probability model, then computes WPA per hero/item/context
- `web/` — Next.js 14 + TypeScript + Tailwind + Recharts frontend dashboard

**Data flow:** Ingest → Compute → Serve → Display

**Database:** Single SQLite file, no external DB needed. Tables: matches, match_players, match_player_items, match_snapshots, items, heroes, wpa_item_results, model_metadata.

## Deploy with Docker (Proxmox / Home Server)

### Prerequisites
- Docker + Docker Compose on the host

### Steps

1. **Clone the repo:**
```bash
git clone git@github.com:starcall/deadlock.git
cd deadlock
```

2. **Configure environment:**
```bash
cp .env.example .env
```

Edit `.env` — set your server's IP/hostname:
```env
PUBLIC_API_URL=http://YOUR_SERVER_IP:8080
PUBLIC_FRONTEND_URL=http://YOUR_SERVER_IP:3000
```

`PUBLIC_API_URL` is called by the **browser** (not internal Docker networking), so it must be reachable from whoever views the dashboard. `PUBLIC_FRONTEND_URL` is used for CORS on the backend.

3. **Build and start:**
```bash
docker compose up -d --build
```

This starts 3 containers:
| Container | Purpose | Port |
|-----------|---------|------|
| `backend` | Go REST API | 8080 |
| `frontend` | Next.js dashboard | 3000 |
| `cron` | Ingests 1000 matches + recomputes WPA every 15 min | — |

The `cron` container runs an initial ingest+compute on startup, then loops every 15 minutes. It uses `--patch-days=14` to only include matches from the last 2 weeks (approximating the current game patch).

4. **Verify:**
```bash
docker compose logs -f cron       # Watch ingest/compute progress
docker compose logs -f backend    # Watch API
curl http://YOUR_SERVER_IP:8080/api/status   # Check data
```

Open `http://YOUR_SERVER_IP:3000` in a browser.

### Data Persistence

SQLite database is stored in a Docker volume (`db-data`). Survives container restarts and rebuilds. To wipe and start fresh:
```bash
docker compose down -v   # -v removes volumes
docker compose up -d --build
```

### Updating

```bash
git pull
docker compose up -d --build
```

Data is preserved across rebuilds (volume is not deleted).

## Local Development (without Docker)

### Prerequisites
- Go 1.22+
- Node.js 20+, npm

### Setup

```bash
cp .env.example .env
go mod download

# 1. Ingest match data
go run ./cmd/ingest --count=1000 --patch-days=14

# 2. Train model + compute WPA
go run ./cmd/compute --patch-days=14

# 3. Start API server (terminal 1)
go run ./cmd/server

# 4. Start frontend (terminal 2)
cd web && npm install && npm run dev
```

Open http://localhost:3000

## CLI Flags

### `cmd/ingest`
| Flag | Default | Description |
|------|---------|-------------|
| `--count` | 1000 | Number of matches to ingest |
| `--patch-days` | 14 | Only ingest matches from last N days |
| `--rate` | 8 | API requests/sec (limit: 100/10s) |
| `--min-duration` | 900 | Min match duration in seconds |
| `--workers` | 8 | Concurrent fetch workers |
| `--db` | `./deadlock.db` | SQLite path |

### `cmd/compute`
| Flag | Default | Description |
|------|---------|-------------|
| `--patch-days` | 14 | Only use matches from last N days |
| `--snapshot-interval` | 300 | Snapshot sampling interval (sec) |
| `--test-split` | 0.2 | Test set fraction |
| `--lambda` | 1.0 | L2 regularization |
| `--lr` | 0.01 | Learning rate |
| `--epochs` | 1000 | Training epochs |
| `--db` | `./deadlock.db` | SQLite path |

### `cmd/server`
| Flag | Default | Description |
|------|---------|-------------|
| `--port` | 8080 | API port |
| `--frontend-url` | `http://localhost:3000` | CORS origin |
| `--db` | `./deadlock.db` | SQLite path |

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/heroes` | All heroes |
| `GET /api/items` | All shop items |
| `GET /api/wpa/hero/:heroId?context=all&min_sample_size=30` | Item WPA for a hero |
| `GET /api/wpa/hero/:heroId/item/:itemId` | WPA across all contexts for a hero/item pair |
| `GET /api/wpa/contexts` | Available filter contexts |
| `GET /api/model/stats` | Model accuracy and calibration |
| `GET /api/model/reliability` | Reliability diagram data |
| `GET /api/status` | Match count, model accuracy, data freshness |

### Context Keys

Used with `?context=` on the WPA endpoint:

| Key | Description |
|-----|-------------|
| `all` | All matches (default) |
| `rank:low` | Badge 0-30 |
| `rank:mid` | Badge 31-70 |
| `rank:high` | Badge 71+ |
| `phase:early` | Items bought 0-10 min |
| `phase:mid` | Items bought 10-25 min |
| `phase:late` | Items bought 25+ min |
| `before:5m` | Items bought before 5 min |
| `before:8m` | Items bought before 8 min |
| `before:10m` | Items bought before 10 min |
| `before:15m` | Items bought before 15 min |
| `before:20m` | Items bought before 20 min |

## How WPA Works

A logistic regression model estimates `P(win | game_state)` using team-level features (net worth diff, kill/death/assist diffs, etc.). Items are excluded from features to avoid circular evaluation.

For each item purchase: **WPA = w(state_after) - w(state_before)**

Results are averaged per hero/item/context combination.

| Metric | Meaning |
|--------|---------|
| **ΔW̄** | Mean win probability change after purchase. Green = positive, Red = negative |
| **W̄** | Mean win probability when item is bought. Yellow = ahead, Blue = behind (reveals selection bias) |
| **Win Rate** | Raw outcome rate for comparison |
| **K** | Sample size (purchases observed) |

## Key Design Decisions

- **Ranked only**: Filters by non-null average_badge fields (API uses "Unranked" for all public matches)
- **Normal game mode only**: Excludes brawl, coop, private lobbies
- **Patch window**: Default 14 days to keep WPA relevant to current game balance
- **Shop items only**: WPA computation filters to known shop items (abilities/upgrades excluded)
- **SQLite single-writer**: `SetMaxOpenConns(1)` — concurrent goroutines queue at Go level, not SQLite level
- **Rate limiting**: 8 req/s with exponential backoff on 429 (API limit: 100/10s)

## Methodology

Based on win probability estimation methodology described in xPetu's thesis on WPA in esports, adapted for Deadlock's team-based item purchasing mechanics.

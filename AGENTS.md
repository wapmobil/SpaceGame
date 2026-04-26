# SpaceGame — Agent Instructions

## Quick Start

### Prerequisites
- Go 1.22+
- Flutter 3.x (WASM target for web builds)
- PostgreSQL 16+

### 1. PostgreSQL

PostgreSQL must be running. Ensure the `spacegame` database exists:

```bash
psql -h localhost -U postgres -c "SELECT 1 FROM pg_database WHERE datname='spacegame'"
# If not found, create it:
createdb -h localhost -U postgres spacegame
```

### 2. Deploy (builds frontend, backend, starts server)

```bash
./deploy.sh
```

Stops any running server, builds Flutter web (`--release --wasm`), copies frontend to `server/web/`, builds Go binary, starts server in background. Logs to `spacegame.log`.

Env vars (defaults): `DB_HOST=localhost`, `DB_PORT=5432`, `DB_USER=postgres`, `DB_PASSWORD=postgres`, `DB_NAME=spacegame`, `DB_SSLMODE=disable`, `PORT=8088`.

### 3. Dev mode

```bash
# Backend only (needs server/web/ populated from a prior deploy)
cd server && go run ./cmd/server/
```

### 4. Open

```
http://localhost:8088/
```

## Repo layout

```
server/          Go backend (module `spacegame`)
  cmd/server/main.go          entrypoint
  internal/api/               HTTP router (chi), handlers, WebSocket
  internal/db/                PostgreSQL, embedded migrations
  internal/game/              Core game engine + subpackages
  internal/scheduler/         Ticked background tasks
  internal/web/               Runtime FS loader for server/web/
client/          Flutter app
  lib/main.dart               entrypoint
```

Handler files in `internal/api/` are split by domain:
- `auth_handlers.go` — health, register, login, planets CRUD
- `building_handlers.go` — buildings, toggle, confirm
- `research_handlers.go` — research
- `fleet_handlers.go` — fleet, ship building
- `expedition_handlers.go` — expeditions
- `market_handlers.go` — market orders, global market, sell food
- `drill_handlers.go` — drill mini-game
- `garden_bed_handlers.go` — garden bed
- `other_handlers.go` — ratings, stats, events

## Commands

**Backend (run from `server/`):**
```
go run ./cmd/server/                     # dev server (needs PostgreSQL + server/web/)
go test -timeout 10s ./...               # all tests (always use -timeout 10s)
go test -timeout 10s ./internal/game/... # single package
go build -o SpaceGameServer ./cmd/server/
```

**Frontend (run from `client/`):**
```
flutter test
flutter build web --release --wasm       # production web build
```

## Architecture

- **Game engine** (`internal/game/game.go`): in-memory singleton via `SetInstance()`/`Instance()`. Manages planets. Each planet ticks every second.
- **Persistence**: Planet state saved to PostgreSQL every 10 ticks. Save covers resources, buildings, research, fleet, shipyard, energy buffer, garden bed.
- **Scheduler** (`internal/scheduler/scheduler.go`): four goroutines — game tick (1s), ratings update (5 min), random events (1s), market NPC refresh (20 min).
- **WebSocket** (`internal/api/websocket.go`): per-player connections, message queue for offline players (5 min retention, 50 msg cap), rate limiting, ping/pong heartbeat (30s).
- **Frontend serving**: NOT `//go:embed`. The `internal/web/` package loads `server/web/` from disk at runtime (relative to executable). Deploy copies Flutter build output there.
- **Auth**: Custom UUID v7 + session tokens. WebSocket auth via `?token=` query param. DB lookups against `players.auth_token`.
- **Migrations**: `//go:embed migrations/*.up.sql` in `internal/db/database.go`. Auto-apply on startup via `schema_migrations` table. Naming: `NNN_description.up.sql`. Add `.up.sql` files to `server/internal/db/migrations/`.

## Testing

- Pure unit tests, no fixtures, no integration tests.
- Tests needing `Game` call `game.New()` directly (no DB).
- Test files: `planet_test.go`, `ratings_test.go`, `marketplace_test.go`, `drill_test.go`, `garden_bed_test.go`, `battle/battle_test.go`, `api/handler_test.go`, `api/websocket_test.go`.
- Garden bed tests call `ClearGardenBedCooldown()` between assertions.

## Gotchas

- **No Makefile / no CI / no pre-commit** — raw `go` / `flutter` commands only.
- **Circular import guard**: `game` package cannot import `api`. Uses callback pattern (`RegisterBroadcastHandler`) for WebSocket broadcast.
- **TriggerRandomEvents() called twice**: once in `game.Tick()` and once in scheduler's `randomEventsTick()`. Be careful modifying either.
- **Drill mini-game**: cooldown is flat 30s (`GetDrillCooldown()` in `drill.go`). Active sessions cleaned up on WebSocket disconnect.
- **Garden bed**: 5s cooldown between actions (`GardenBedActionCooldown` in `garden_bed.go`). State persisted to both `garden_bed` table and `planets.garden_bed_grid`.
- **Marketplace**: NPC traders refresh every 20 minutes. NPC order count scales with total market building level across all planets.
- **Energy system**: `calculateEnergy()` returns (production, consumption). Solar buildings produce negative consumption. Auto-disable logic turns off non-essential buildings when energy is insufficient.
- **`server/web/` is gitignored**: must be populated by deploy or manual Flutter build. Running `go run` without it means no frontend served.

# SpaceGame — Agent Instructions

## Quick Start

### Prerequisites
- Go 1.23+
- Flutter 3.41+
- PostgreSQL 16+

### 1. Setup PostgreSQL

```bash
# Install PostgreSQL
sudo apt install postgresql
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'postgres';"
sudo -u postgres createdb spacegame
```

### 2. Deploy (builds frontend, backend, starts server)

```bash
./deploy.sh
```

This script stops any running server, builds Flutter web, copies frontend to `server/web/`, builds the Go binary, and starts the server in the background with logs written to `spacegame.log`.

Environment variables (optional, defaults shown):
`DB_HOST=localhost`, `DB_PORT=5432`, `DB_USER=postgres`, `DB_PASSWORD=postgres`, `DB_NAME=spacegame`, `DB_SSLMODE=disable`, `PORT=8088`.

### 3. Open in browser

```
http://localhost:8088/
```

Public address: `http://home.signalmodelling.ru:8088/`

### Useful commands

```bash
# View logs
tail -f spacegame.log

# Deploy (full rebuild + restart)
./deploy.sh

# Dev mode — run backend directly (hot reload not available, rebuild manually)
cd server && go run ./cmd/server/
```

## Repo layout

```
server/          Go backend (module `spacegame`)
  cmd/server/main.go          entrypoint
  internal/api/               HTTP router (chi), handlers, WebSocket
  internal/db/                PostgreSQL, embedded migrations (auto-apply on startup)
  internal/game/              Core game engine + subpackages (building, ship, research, battle, expedition, mining)
  internal/scheduler/         Ticked background tasks
  migrations/                 .up/.down SQL files (kept in sync with internal/db/migrations/)
client/          Flutter app
  lib/main.dart               entrypoint
  lib/core/websocket_manager.dart
  lib/providers/              state providers (Provider package)
  lib/screens/                UI screens
  lib/models/                 Data models
  lib/services/api_service.dart
  lib/widgets/                Reusable widgets
```

## Commands

**Backend (run from `server/`):**
```
go run ./cmd/server/          # start dev server (depends on PostgreSQL)
go test -timeout 10s ./...    # all tests (always use -timeout 10s to catch infinite loops)
go test -timeout 10s ./internal/game/...   # a single package
go build ./cmd/server/        # build binary to server/server
```

**Frontend (run from `client/`):**
```
flutter test                  # unit / widget tests
flutter run                   # dev run
flutter build linux           # release build
```

**Environment variables (server):**
`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` (defaults: localhost, 5432, postgres, postgres, spacegame, disable). Server port: `PORT` (default 8080).

## Architecture notes

- **Game engine** (`internal/game/game.go`): in-memory singleton (`game.Instance()`), manages planets. Each planet ticks every second via `scheduler`.
- **Persistence**: Planet state saved to PostgreSQL every 10 ticks. Migrations are `//go:embed` in `internal/db/database.go` — add `.up.sql` files there, not in the top-level `migrations/` directory (the top-level copies are for reference).
- **Scheduler** (`internal/scheduler/scheduler.go`): three goroutines — game tick (1s), ratings update (5 min), random events (checked every 1s).
- **WebSocket**: `internal/api/websocket.go` — per-player connections, broadcast channel for state updates.
- **Migrations**: auto-apply on server startup via `schema_migrations` table. Naming convention: `NNN_description.up.sql`.

## Testing

- Tests are pure unit tests, no fixtures, no integration tests.
- Tests that need `Game` call `game.New()` directly (no DB).
- `ratings_test.go`, `battle_test.go`, `marketplace_test.go`, `mining_test.go`, `research_test.go`, `ship_test.go`, `expedition_test.go`, `planet_test.go` are the test files.

## Gotchas

- **deploy.sh** — single command to build frontend, backend, and start the server (`./deploy.sh`).
- **No Makefile / no CI / no pre-commit** — everything else is raw `go` / `flutter` commands.
- **Circular import guard**: `game` package cannot import `api` — uses callback pattern (`RegisterBroadcastHandler`) for WebSocket broadcast.
- **Battle cooldown**: auto-battle checks `battleCooldownTicks = 60` (1 minute). The `getBattleTick()` stub always returns true; per-planet cooldown logic is in `processAutoBattle`.
- **Mining mini-game**: cooldown is 30s in production, 5 min in development (check `mining.go` for the flag).
- **Marketplace matching**: NPC traders refresh every 15 seconds.

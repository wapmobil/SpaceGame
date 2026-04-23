# SpaceGame

A real-time multiplayer space strategy game with asynchronous multiplayer, built with Go and Flutter.

## Features

- **Planet Management** — Build structures, manage resources and energy production
- **Research System** — Two technology trees (Standard & Alien) with 13+ technologies
- **Ship Fleet** — Design and command 6 types of ships
- **Auto-Battles** — Fleet combat with damage calculations and looting
- **Expeditions** — Explore space, discover NPC planets, trade and support missions
- **Marketplace** — Global trading with buy/sell orders and NPC traders
- **Mining Mini-game** — Dungeon crawler procedural maze with monsters and loot
- **Real-time Updates** — WebSocket-powered live state synchronization
- **Leaderboards** — Competitive rankings across multiple categories
- **Random Events** — Dynamic events like short circuits, thefts, and storage collapses

## Tech Stack

- **Backend:** Go 1.23+, PostgreSQL 16+, WebSocket
- **Frontend:** Flutter 3.41+, Provider state management
- **Deployment:** Self-hosted, Linux

## Quick Start

### Prerequisites

- Go 1.23+
- Flutter 3.41+
- PostgreSQL 16+

### 1. Setup PostgreSQL

```bash
sudo apt install postgresql
sudo systemctl start postgresql
sudo systemctl enable postgresql
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'postgres';"
sudo -u postgres createdb spacegame
```

### 2. Deploy

```bash
./deploy.sh
```

This builds the Flutter frontend, copies it to the server, builds the Go binary, and starts the server.

Environment variables (with defaults):

| Variable | Default |
|---|---|
| `DB_HOST` | `localhost` |
| `DB_PORT` | `5432` |
| `DB_USER` | `postgres` |
| `DB_PASSWORD` | `postgres` |
| `DB_NAME` | `spacegame` |
| `DB_SSLMODE` | `disable` |
| `PORT` | `8088` |

### 3. Open

```
http://localhost:8088/
```

### Useful Commands

```bash
# View logs
tail -f spacegame.log

# Dev mode (backend only)
cd server && go run ./cmd/server/

# Run tests
cd server && go test -timeout 10s ./...
cd client && flutter test
```

## Project Structure

```
server/          Go backend
  cmd/server/    Entry point
  internal/api/  HTTP router (chi), handlers, WebSocket
  internal/db/   PostgreSQL, embedded migrations
  internal/game/ Core game engine (buildings, ships, research, battle, expedition, mining)
  internal/scheduler/ Ticked background tasks
  migrations/    SQL migration files

client/          Flutter frontend
  lib/main.dart  Entry point
  lib/core/      WebSocket manager
  lib/providers/ State management (Provider)
  lib/screens/   UI screens (home, planet, battle, research, etc.)
  lib/models/    Data models
  lib/services/  API service
  lib/widgets/   Reusable UI components
```

## Game Systems

### Resources

7 resources: `food`, `composite`, `mechanisms`, `reagents`, `energy`, `money`, `alien_tech`

### Buildings

8 building types: `farm`, `solar`, `storage`, `base`, `factory`, `energy_storage`, `shipyard`, `comcenter`

### Ships

6 ship types: `trade`, `small`, `interceptor`, `corvette`, `frigate`, `cruiser`

### Technology Trees

- **Standard Research** (10 technologies) — Planet Exploration, Energy Storage, Trade, Ships, etc.
- **Alien Technology** (3 technologies) — Alien Technologies, Additional Expedition, Super Energy Storage

## Architecture

- **Game Engine** — In-memory singleton, manages planets. Each planet ticks every second.
- **Persistence** — Planet state saved to PostgreSQL every 10 ticks. Migrations auto-apply on startup.
- **Scheduler** — Three goroutines: game tick (1s), ratings update (5 min), random events.
- **WebSocket** — Per-player connections with broadcast channel for state updates.
- **Auth** — Custom auth with UUID v7 + session tokens.

## Testing

```bash
# All backend tests
cd server && go test -timeout 10s ./...

# A single package
cd server && go test -timeout 10s ./internal/game/...

# Frontend tests
cd client && flutter test
```

## License

MIT

> **Общий план:** [surface_expeditions_design.md](../surface_expeditions_design.md)

# Фаза 1: DB migration + old code cleanup

## Цель
Создать новую БД схему для expedition event chains, удалить старую схему surface expeditions (JSONB в planets, таблица surface_expeditions).

## Что работает после деплоя
Без surface expeditions. Без expedition events. Старые данные (locations, location_buildings) сохраняются.

## Задачи

### 6.1 Создать migration — удаление старой схемы
**Файл:** `server/internal/db/migrations/025_cleanup_surface_expeditions.up.sql`

Удаление старых JSONB-колонок из planets:
```sql
ALTER TABLE planets DROP COLUMN IF EXISTS surface_expeditions;
ALTER TABLE planets DROP COLUMN IF EXISTS expedition_history;
ALTER TABLE planets DROP COLUMN IF EXISTS range_stats;
ALTER TABLE planets DROP COLUMN IF EXISTS max_expeditions;
```

Удаление старых таблиц:
```sql
DROP TABLE IF EXISTS surface_expeditions;
DROP TABLE IF EXISTS expedition_history;
```

### 6.2 Создать migration — новая схема expedition chains
**Файл:** `server/internal/db/migrations/025_cleanup_surface_expeditions.up.sql` (продолжение)

Таблица цепочек экспедиций:
```sql
CREATE TABLE IF NOT EXISTS expedition_chains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'completed', 'failed')),
    event_count INTEGER NOT NULL DEFAULT 0,
    current_event_index INTEGER NOT NULL DEFAULT 0,
    discovered_location_id UUID REFERENCES surface_locations(id) ON DELETE SET NULL,
    inventory JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_expedition_chains_planet ON expedition_chains(planet_id);
CREATE INDEX idx_expedition_chains_owner ON expedition_chains(owner_id);
```

Таблица событий экспедиции:
```sql
CREATE TABLE IF NOT EXISTS expedition_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id UUID NOT NULL REFERENCES expedition_chains(id) ON DELETE CASCADE,
    event_id TEXT NOT NULL,
    description TEXT NOT NULL,
    choices JSONB NOT NULL DEFAULT '[]',
    immediate_reward JSONB NOT NULL DEFAULT '{}',
    is_end BOOLEAN NOT NULL DEFAULT false,
    location_reward TEXT,
    player_choice INTEGER,
    rewards_received JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_expedition_events_chain ON expedition_events(chain_id);
```

**Примечание:** Храним весь JSON события (включая description, choices, rewards). Для UI истории нужно: event_id, description, choices (для извлечения label выбранного варианта), player_choice, rewards_received.

### 6.3 Удалить старые файлы planet_survey/
**Удалить:** `server/internal/game/planet_survey/surface_expedition.go`
- Все функции: NewSurfaceExpedition, Tick, IsExpired, GetCostPerMin, CalculateCost, CalculateDiscoveryChance, GetResourceChance, CalculateResourceRecovery

**Удалить:** `server/internal/game/planet_survey/surface_expedition_test.go`

**Удалить:** `server/internal/game/planet_survey/history.go`
- ExpeditionHistoryEntry — больше не нужен, история теперь в БД

### 6.4 Удалить вызовы TickSurfaceExpeditions
**Файл:** `server/internal/game/planet_tick.go`

Удалить функцию `TickSurfaceExpeditions()` полностью (строки 83-215).
Удалить вызов `TickSurfaceExpeditions()` из `Tick()`.

Оставить `TickLocationBuildings()` — он не зависит от surface expeditions.

### 6.5 Обновить Planet struct
**Файл:** `server/internal/game/planet.go`

Удалить поля:
```go
// УДАЛИТЬ:
SurfaceExpeditions []*planet_survey.SurfaceExpedition
ExpeditionHistory  []planet_survey.ExpeditionHistoryEntry
RangeStats         map[string]*planet_survey.ExpeditionRangeStats
MaxExpeditions     int
```

Добавить:
```go
// ДОБАВИТЬ:
ExpeditionChains []*ExpeditionChain
```

В `NewPlanet()`:
- Удалить инициализацию SurfaceExpeditions, ExpeditionHistory, RangeStats, MaxExpeditions
- Добавить: `ExpeditionChains: make([]*ExpeditionChain, 0)`

### 6.6 Обновить applyResearchEffects
**Файл:** `server/internal/game/planet.go`

Удалить блок `advanced_exploration` → `MaxExpeditions`.
Advanced exploration больше не нужен в текущей форме — лимит экспедиций будет управляться иначе.

### 6.7 Обновить savePlanet
**Файл:** `server/internal/game/game.go`

Удалить сохранение:
- surface_expeditions JSONB
- expedition_history JSONB
- range_stats JSONB

Удалить функции:
- `loadSurfaceExpeditions()`
- `loadExpeditionHistory()`
- `loadRangeStats()`

Оставить:
- `loadLocations()` — locations остаются в JSONB
- `loadLocationBuildings()` — location_buildings остаются в JSONB

### 6.8 Обновить loadPlanetFromDB / loadPlanetsFromDB
**Файл:** `server/internal/game/game.go`

Удалить вызовы:
- `loadSurfaceExpeditions(p)`
- `loadExpeditionHistory(p)`
- `loadRangeStats(p)`

Оставить:
- `loadLocations(p)`

### 6.9 Обновить planet_state.go
**Файл:** `server/internal/game/planet_state.go`

В `GetState()` удалить:
- `"surface_expeditions"`
- `"max_expeditions"`
- `"can_start_planet_survey"` (заменится на `can_start_expedition`)

В `GetBuildDetails()` удалить:
- `CanPlanetSurvey`

### 6.10 Обновить GetExpeditionState / GetActiveExpeditionsCount
**Файл:** `server/internal/game/planet_state.go`

Удалить `GetExpeditionState()` — если он возвращает surface expeditions.
Удалить `GetActiveExpeditionsCount()` — если он считает surface expeditions.
Удалить `GetMaxExpeditions()`.

### 6.11 Тесты
**Запуск:**
```bash
go test -timeout 10s ./internal/game/planet_survey/...
```
Ожидаем ошибку — файлы удалены. Это нормально.

```bash
go build ./cmd/server/
```
Проверяем что бэкенд собирается.

## Деплой
```bash
./deploy.sh
go build ./cmd/server/
```

## Риски
- Если БД не миграция — сервер запустится без expedition_chains таблицы. Миграция применится автоматически при старте.
- Старые данные surface_expeditions будут потеряны. Это намеренно.
- Locations и location_buildings сохраняются через JSONB — не затрагиваются.

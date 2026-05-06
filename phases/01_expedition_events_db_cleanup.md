# Фаза 1: DB migration + old code cleanup

## Цель
Создать новую БД схему для expedition event chains, удалить старую схему surface expeditions (JSONB в planets, таблицы surface_expeditions).

## Что работает после деплоя
Без surface expeditions. Без expedition events. Старые данные (locations, location_buildings) сохраняются.

## Задачи

### 1.1 Создать migration — новая схема expedition chains

**Файл:** `server/internal/db/migrations/024_expedition_chains.up.sql` (NEW)

Сначала создаём новые таблицы, потом удаляем старые — если создание не пройдёт, старые данные не потеряны.

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
CREATE INDEX idx_expedition_chains_status ON expedition_chains(status);
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

**Примечание:** Храним весь JSON события (description, choices, rewards). Для UI истории нужно: event_id, description, choices (для извлечения label выбранного варианта), player_choice, rewards_received.

### 1.2 Создать migration — удаление старой схемы

**Файл:** `server/internal/db/migrations/024_expedition_chains.up.sql` (продолжение, после CREATE TABLE)

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

### 1.3 Создать down migration

**Файл:** `server/internal/db/migrations/024_expedition_chains.down.sql` (NEW)

```sql
DROP TABLE IF EXISTS expedition_events;
DROP TABLE IF EXISTS expedition_chains;
ALTER TABLE planets ADD COLUMN IF NOT EXISTS surface_expeditions JSONB DEFAULT '[]';
ALTER TABLE planets ADD COLUMN IF NOT EXISTS expedition_history JSONB DEFAULT '[]';
ALTER TABLE planets ADD COLUMN IF NOT EXISTS range_stats JSONB DEFAULT '{}';
ALTER TABLE planets ADD COLUMN IF NOT EXISTS max_expeditions INTEGER DEFAULT 1;
```

### 1.4 Удалить старые файлы planet_survey/

**Удалить:** `server/internal/game/planet_survey/surface_expedition.go`
- Все функции: NewSurfaceExpedition, Tick, IsExpired, GetCostPerMin, CalculateCost, CalculateDiscoveryChance, GetResourceChance, CalculateResourceRecovery

**Удалить:** `server/internal/game/planet_survey/surface_expedition_test.go`

**Удалить:** `server/internal/game/planet_survey/history.go`
- ExpeditionHistoryEntry — больше не нужен, история теперь в БД

### 1.5 Удалить TickSurfaceExpeditions

**Файл:** `server/internal/game/planet_tick.go`

Удалить функцию `TickSurfaceExpeditions()` полностью (строки 83-215).
Удалить вызов `TickSurfaceExpeditions()` из `Tick()` (строка 69).

Оставить `TickLocationBuildings()` — он не зависит от surface expeditions.

### 1.6 Обновить Planet struct

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
ExpeditionChains []*planet_survey.ExpeditionChain
```

В `NewPlanet()`:
- Удалить инициализацию SurfaceExpeditions, ExpeditionHistory, RangeStats, MaxExpeditions
- Добавить: `ExpeditionChains: make([]*planet_survey.ExpeditionChain, 0)`

### 1.7 Обновить applyResearchEffects

**Файл:** `server/internal/game/planet.go`

Удалить блок `advanced_exploration` → `MaxExpeditions` из `applyResearchEffects()`.
Лимит экспедиций теперь управляется через константу (по умолчанию 1 активная цепочка на планету).

### 1.8 Обновить savePlanet

**Файл:** `server/internal/game/game.go`

Удалить блоки сохранения:
- surface_expeditions JSONB (строки 520-536)
- expedition_history JSONB (строки 570-586)
- range_stats JSONB (строки 589-605)
- max_expeditions (строки 558-567)

Удалить функции:
- `loadSurfaceExpeditions()` (строки 652-666)
- `loadExpeditionHistory()` (строки 684-698)
- `loadRangeStats()` (строки 700-714)

Оставить:
- `loadLocations()` — locations остаются в JSONB

### 1.9 Обновить loadPlanetFromDB / loadPlanetsFromDB

**Файл:** `server/internal/game/game.go`

Удалить вызовы:
- `loadSurfaceExpeditions(p)` (строка 168)
- `loadExpeditionHistory(p)` (строка 178)
- `loadRangeStats(p)` (строка 183)
- Блок загрузки max_expeditions (строки 187-194)

Оставить:
- `loadLocations(p)`

### 1.10 Обновить planet_state.go

**Файл:** `server/internal/game/planet_state.go`

В `GetState()` удалить ключи:
- `"surface_expeditions"` (строка 93)
- `"range_stats"` (строка 96)
- `"expedition_history"` (строка 97)
- `"max_surface_expeditions"` (строка 98)
- `"can_start_planet_survey"` (строка 99)

Удалить функции:
- `GetSurfaceExpeditionState()` (строки 168-197)
- `GetRangeStatsState()` (строки 248-257)
- `CanStartPlanetSurvey()` (строки 259-276)
- `GetMaxSurfaceExpeditions()` (строки 278-286)

В `BuildDetails` struct удалить поля:
- `MaxSurfaceExpeditions` (строка 122)
- `CanPlanetSurvey` (строка 125)
- `PlanetSurveyUnlocked` (строка 126)

В `GetBuildDetails()` удалить:
- Вычисление `planetSurveyUnlocked` (строки 142-147)
- Поля `MaxSurfaceExpeditions`, `CanPlanetSurvey`, `PlanetSurveyUnlocked` в возвращаемом struct

### 1.11 Тесты

**Запуск:**
```bash
go test -timeout 10s ./internal/game/planet_survey/...
```
Ожидаем ошибку — файлы удалены. Это нормально.

```bash
go build -o /dev/null ./cmd/server/
```
Проверяем что бэкенд собирается.

## Деплой

```bash
./deploy.sh
go build -o /dev/null ./cmd/server/
```

## Риски

- Старые данные surface_expeditions будут потеряны. Это намеренно.
- Locations и location_buildings сохраняются через JSONB — не затрагиваются.
- Migration 024 идёт сразу после 023 — номер корректен.

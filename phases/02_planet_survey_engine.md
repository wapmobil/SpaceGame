> **Общий план:** [surface_expeditions_design.md](../surface_expeditions_design.md)

# Фаза 2: Planet Survey engine (core)

## Цель
Создать полностью рабочую систему surface expeditions на бэкенде — запуск, тик, обнаружение, ресурсы. Без API и UI.

## Что работает после деплоя
Полностью рабочая система surface expeditions на бэкенде. Без API и UI — работает через прямой вызов методов.

## Задачи

### 2.1 Создать planet_survey/ package
**Директория:** `server/internal/game/planet_survey/`

### 2.2 location.go — Location, discovery logic
**Файл:** `server/internal/game/planet_survey/location.go`

Структуры:
```go
type Location struct {
    ID              string
    PlanetID        string
    OwnerID         string
    Type            string    // "pond", "river", "forest", etc.
    Name            string
    BuildingType    string
    BuildingLevel   int
    BuildingActive  bool
    SourceResource  string
    SourceAmount    float64
    SourceRemaining float64
    Active          bool
    DiscoveredAt    time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type LocationBuildingDef struct {
    BuildingType       string
    Level1Production   building.ProductionResult
    Level2Production   building.ProductionResult
    Level3Production   building.ProductionResult
    SourceConsumption  float64
}

type LocationType struct {
    Type          string
    Name          string
    Buildings     []LocationBuildingDef
    SourceResource string
    AmountRange   [2]float64
    RarityWeight  int
}
```

Функции:
- `GetLocationTypes() []LocationType` — 20 типов с rarity weights (30/20/12/6)
- `SelectLocationType(planetType PlanetResourceType) string` — взвешенный рандом с ResourceType bias (×6 match, ×1 others)
- `GenerateName(locType string) string` — генерация имени
- `CalculateSourceAmount(locType string, planetType PlanetResourceType) float64` — rarity multiplier × ResourceType bias

Rarity weights:
- Обычные (30): pond, river, forest, mineral_deposit, dry_valley
- Необычные (20): waterfall, cave, thermal_spring, salt_lake, wind_pass
- Редкие (12): crystal_cave, meteor_crater, sunken_city, glacier, mushroom_forest
- Экзотические (6): crystal_field, cloud_island, underground_lake, radioactive_zone, anomaly_zone

ResourceType bias:
- Match: ×6 chance, ×1.5 source amount
- No match: ×1 chance, ×0.5 source amount

Rarity multiplier:
- Обычная: ×1.0, Необычная: ×1.2, Редкая: ×1.5, Экзотическая: ×2.0

### 2.3 buildings.go — Building definitions
**Файл:** `server/internal/game/planet_survey/buildings.go`

Static map: 20 location types → 1-3 building options each.

Примеры:
```go
// pond: fish_farm, water_purifier
// river: fish_farm, irrigation_system, water_plant
// forest: lumber_mill, herb_garden, resin_tap
// mineral_deposit: mineral_extractor, smelter
// dry_valley: solar_farm, wind_turbine
// ... и т.д. для всех 20 типов
```

Каждое здание: production per level (lvl 1/2/3), source consumption per level.

Полная таблица из design.md (строки 338-385).

### 2.4 surface_expedition.go — PlanetSurveyExpedition
**Файл:** `server/internal/game/planet_survey/surface_expedition.go`

Структура:
```go
type SurfaceExpedition struct {
    ID            string
    PlanetID      string
    Status        string    // "active", "discovered", "completed", "failed"
    Progress      float64
    Duration      float64
    ElapsedTime   float64
    Discovered    *Location
    Range         string    // "300s", "600s", "1200s"
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type ExpeditionRangeStats struct {
    TotalExpeditions int
    LocationsFound   int
}

type CostPerMin struct {
    Food  float64
    Iron  float64
    Money float64
}
```

Base level → max duration + cost per min:
- Lvl 1: 300s, 100/100/10
- Lvl 2: 600s, 200/200/20
- Lvl 3: 1200s, 400/400/40

Функции:
- `NewSurfaceExpedition(id, planetID, range string, baseLevel int) *SurfaceExpedition`
- `Tick(exp *SurfaceExpedition, dt float64)` — advance elapsed time
- `IsExpired(exp *SurfaceExpedition) bool` — elapsed >= duration
- `GetCostPerMin(baseLevel int) CostPerMin`
- `CalculateCost(baseLevel int, duration float64) (food, iron, money float64)`
- `CalculateDiscoveryChance(count int) float64` — 45% × 1/(1+(count/3)^2)
- `GetResourceChance(count int) float64` — 100%/count if count > 0, else 100%
- `CalculateResourceRecovery(baseLevel int, count int, isSuccess bool) map[string]float64`

Resource recovery:
- Failure: resourceChance = 100%/max(1,count), each resource independently
- Success: resourceChance = resourceChance/5.0
- maxAmount: food/iron=1000, money=250, reagents/composite/mechanisms=100
- amount = rand.Float64() * maxAmount * baseLevel

### 2.5 location_building.go — LocationBuilding tick/depletion
**Файл:** `server/internal/game/planet_survey/location_building.go`

Структура:
```go
type LocationBuilding struct {
    ID            string
    LocationID    string
    BuildingType  string
    Level         int
    Active        bool
    BuildProgress float64
    BuildTime     float64
    CostFood      float64
    CostIron      float64
    CostMoney     float64
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

Функции:
- `TickLocationBuildings(location *Location, buildings []*LocationBuilding, planetResources *PlanetResources) float64` — production + depletion
- `Depletion logic`: production = ProductionPerLevel[building_type, level], add to planet resources, source_remaining -= source_consumption * level
- When source_remaining <= 0: building_active = false, location marked "depleted"

### 2.6 history.go — ExpeditionHistoryEntry
**Файл:** `server/internal/game/planet_survey/history.go`

```go
type ExpeditionHistoryEntry struct {
    ID             string
    PlanetID       string
    ExpeditionType string // "surface"
    Status         string
    Result         string // "success", "failed", "abandoned"
    Discovered     string // location name
    ResourcesGained map[string]float64
    CreatedAt      time.Time
    CompletedAt    time.Time
}
```

### 2.7 Добавить поля в Planet struct
**Файл:** `server/internal/game/planet.go`

Добавить:
```go
SurfaceExpeditions []*planet_survey.SurfaceExpedition
Locations          []*planet_survey.Location
ExpeditionHistory  []planet_survey.ExpeditionHistoryEntry
RangeStats         map[string]*planet_survey.ExpeditionRangeStats
ExpeditionCooldown int64         // unix timestamp
MaxLocations       int
```

В `NewPlanet()`:
```go
SurfaceExpeditions: make([]*planet_survey.SurfaceExpedition, 0),
Locations:          make([]*planet_survey.Location, 0),
ExpeditionHistory:  make([]planet_survey.ExpeditionHistoryEntry, 0),
RangeStats:         make(map[string]*planet_survey.ExpeditionRangeStats),
ExpeditionCooldown: 0,
MaxLocations:       1,
```

### 2.8 TickSurfaceExpeditions()
**Файл:** `server/internal/game/planet_tick.go`

Логика:
1. Для каждой активной экспедиции: tick (advance elapsed time)
2. Если expired:
   - Get range stats for this range
   - Check discovery chance every 60 ticks
   - If discovery: create Location, set Discovered, status = "discovered"
   - If no discovery: status = "failed"
   - Calculate resource recovery
   - Apply resources to planet
   - Add to history
   - Remove from SurfaceExpeditions
   - Update range stats
   - Set cooldown (30s)
3. Check base.IsWorking() — if not working, pause expeditions (don't tick)
4. If base starts working again, resume

### 2.9 TickLocationBuildings()
**Файл:** `server/internal/game/planet_tick.go`

Логика:
1. Для каждой location с building:
   - If building active and source_remaining > 0:
     - Calculate production per level
     - Add to planet resources
     - source_remaining -= source_consumption * level
   - If source_remaining <= 0: building_active = false

### 2.10 Удалить random building unlock
**Файл:** `server/internal/game/planet.go`

В `applyResearchEffects()`:
- Удалить case "planet_exploration" block

**Файл:** `server/internal/game/game.go`

В `LoadPlanetFromDB()` и `LoadPlanetsFromDB()`:
- Удалить блок "If planet_exploration is completed but no random unlock"

### 2.11 Обновить IsBuildingUnlocked()
**Файл:** `server/internal/game/building_entry.go`

- `RandomUnlockBuildings` уже пустой (из предфазы)
- Упростить функцию — убрать проверку random unlocks

### 2.12 Update GetState()
**Файл:** `server/internal/game/planet_state.go`

Добавить:
```go
"surface_expeditions": p.GetSurfaceExpeditionState(),
"locations":           p.GetLocationState(),
"max_locations":       p.MaxLocations,
"range_stats":         p.GetRangeStatsState(),
"surface_expedition_cooldown": p.ExpeditionCooldown,
"can_start_planet_survey": p.CanStartPlanetSurvey(),
"can_start_space_expedition": p.CanStartSpaceExpedition(),
```

### 2.13 Update GetBuildDetails()
**Файл:** `server/internal/game/planet_state.go`

В BuildDetails добавить:
```go
CanPlanetSurvey bool
```

В `GetBuildDetails()`:
```go
details.CanPlanetSurvey = baseOperational && p.Research.GetCompleted()["planet_exploration"] > 0
```

### 2.14 Update savePlanet()
**Файл:** `server/internal/game/game.go`

Добавить сохранение:
```go
// Save surface expeditions
if len(p.SurfaceExpeditions) > 0 {
    expData, _ := json.Marshal(p.SurfaceExpeditions)
    g.db.Exec(`UPDATE planets SET surface_expeditions = $1::jsonb WHERE id = $2`, string(expData), p.ID)
}

// Save locations
if len(p.Locations) > 0 {
    locData, _ := json.Marshal(p.Locations)
    g.db.Exec(`UPDATE planets SET locations = $1::jsonb WHERE id = $2`, string(locData), p.ID)
}

// Save location buildings
// Save history
// Save range stats
```

### 2.15 Update LoadPlanetFromDB() / LoadPlanetsFromDB()
**Файл:** `server/internal/game/game.go`

Добавить загрузку:
```go
// Load surface expeditions
// Load locations
// Load history
// Load range stats
```

### 2.16 Migration — surface expeditions tables
**Файл:** `server/internal/db/migrations/020_add_surface_expeditions.up.sql`

```sql
-- Add columns to planets
ALTER TABLE planets ADD COLUMN surface_expeditions JSONB NOT NULL DEFAULT '[]';
ALTER TABLE planets ADD COLUMN locations JSONB NOT NULL DEFAULT '[]';
ALTER TABLE planets ADD COLUMN location_buildings JSONB NOT NULL DEFAULT '[]';
ALTER TABLE planets ADD COLUMN expedition_history JSONB NOT NULL DEFAULT '[]';
ALTER TABLE planets ADD COLUMN range_stats JSONB NOT NULL DEFAULT '{}';

-- New tables
CREATE TABLE IF NOT EXISTS surface_expeditions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active',
    progress REAL NOT NULL DEFAULT 0,
    duration REAL NOT NULL DEFAULT 1800,
    elapsed_time REAL NOT NULL DEFAULT 0,
    discovered_location_id UUID REFERENCES surface_locations(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS surface_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    building_type TEXT,
    building_level INTEGER NOT NULL DEFAULT 1,
    building_active BOOLEAN NOT NULL DEFAULT false,
    source_resource TEXT NOT NULL,
    source_amount REAL NOT NULL,
    source_remaining REAL NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    discovered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS location_buildings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_id UUID NOT NULL REFERENCES surface_locations(id) ON DELETE CASCADE,
    building_type TEXT NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL DEFAULT false,
    build_progress REAL NOT NULL DEFAULT 0,
    build_time REAL NOT NULL DEFAULT 0,
    cost_food REAL NOT NULL DEFAULT 0,
    cost_iron REAL NOT NULL DEFAULT 0,
    cost_money REAL NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS expedition_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    planet_id UUID NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    expedition_type TEXT NOT NULL,
    status TEXT NOT NULL,
    result TEXT NOT NULL,
    discovered TEXT,
    resources_gained JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- expedition_type column on expeditions table
ALTER TABLE expeditions ADD COLUMN expedition_type TEXT NOT NULL DEFAULT 'space_exploration';
UPDATE expeditions SET expedition_type = 'space_exploration' WHERE expedition_type = 'exploration';
UPDATE expeditions SET expedition_type = 'space_trade' WHERE expedition_type = 'trade';
UPDATE expeditions SET expedition_type = 'space_support' WHERE expedition_type = 'support';
```

### 2.17 Тесты
**Файл:** `server/internal/game/planet_survey/surface_expedition_test.go`
- Тест: `NewSurfaceExpedition()` — correct duration/cost
- Тест: `Tick()` — elapsed increases
- Тест: `IsExpired()` — returns true when elapsed >= duration
- Тест: `GetCostPerMin()` — correct for each base level
- Тест: `CalculateDiscoveryChance()` — correct formula
- Тест: `GetResourceChance()` — correct formula
- Тест: `CalculateResourceRecovery()` — resource amounts

**Файл:** `server/internal/game/planet_survey/location_test.go`
- Тест: `SelectLocationType()` — weighted random
- Тест: `SelectLocationType()` — ResourceType bias
- Тест: `CalculateSourceAmount()` — rarity + ResourceType bias
- Тест: `TickLocationBuildings()` — production applied
- Тест: `TickLocationBuildings()` — depletion over time

**Файл:** `server/internal/game/planet_survey/buildings_test.go`
- Тест: Все 20 location types определены
- Тест: Все buildings имеют production per level
- Тест: Все buildings имеют source consumption

**Файл:** `server/internal/game/planet_test.go`
- Тест: Planet struct has SurfaceExpeditions field
- Тест: Planet struct has Locations field
- Тест: Planet struct has RangeStats field
- Тест: Save/Load roundtrip

**Запуск:**
```bash
go test -timeout 10s ./internal/game/planet_survey/...
go test -timeout 10s ./internal/game/...
```

## Деплой
```bash
./deploy.sh
go test -timeout 10s ./...
```

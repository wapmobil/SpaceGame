> **Общий план:** [surface_expeditions_design.md](../surface_expeditions_design.md)

# Фаза 4: API + WebSocket

## Цель
Создать полностью рабочий бэкенд API для surface expeditions и locations.

## Что работает после деплоя
Полностью рабочий бэкенд API. Фронтенд может интегрироваться.

## Задачи

### 4.1 API models
**Файл:** `server/internal/api/models.go`

Добавить request/response типы:
```go
type StartPlanetSurveyRequest struct {
    Duration int `json:"duration"`
}

type BuildOnLocationRequest struct {
    BuildingType string `json:"building_type"`
}

type LocationResponse struct {
    ID             string                  `json:"id"`
    Type           string                  `json:"type"`
    Name           string                  `json:"name"`
    BuildingType   string                  `json:"building_type,omitempty"`
    BuildingLevel  int                     `json:"building_level"`
    BuildingActive bool                    `json:"building_active"`
    SourceResource string                  `json:"source_resource"`
    SourceAmount   float64                 `json:"source_amount"`
    SourceRemaining float64                `json:"source_remaining"`
    Active         bool                    `json:"active"`
    DiscoveredAt   time.Time               `json:"discovered_at"`
}

type RangeStatsResponse map[string]ExpeditionRangeStatsResponse
type ExpeditionRangeStatsResponse struct {
    TotalExpeditions int `json:"total_expeditions"`
    LocationsFound   int `json:"locations_found"`
}

type ExpeditionHistoryResponse struct {
    ID              string            `json:"id"`
    Status          string            `json:"status"`
    Result          string            `json:"result"`
    Discovered      string            `json:"discovered"`
    ResourcesGained map[string]float64 `json:"resources_gained"`
    CreatedAt       time.Time         `json:"created_at"`
    CompletedAt     time.Time         `json:"completed_at"`
}
```

### 4.2 planet_survey_handlers.go — все endpoints
**Файл:** `server/internal/api/planet_survey_handlers.go` (NEW)

**handleStartPlanetSurvey** — `POST /api/planets/{id}/planet-survey`
- Validate: base.IsWorking(), planet_exploration completed, cooldown expired
- Validate: duration <= maxDuration for base level
- Validate: len(locations) < max_locations
- Validate: resources affordable (food, iron, money)
- Deduct resources
- Create SurfaceExpedition
- Set cooldown
- Return 200 with expedition info

**handleGetPlanetSurvey** — `GET /api/planets/{id}/planet-survey`
- Return expeditions list + range_stats + max_duration + cost_per_min

**handleGetLocations** — `GET /api/planets/{id}/locations`
- Return locations list with building info

**handleBuildOnLocation** — `POST /api/planets/{id}/locations/{id}/build`
- Validate: location exists, player owns it
- Validate: location_buildings research completed
- Validate: no building on location yet
- Validate: resources affordable
- Validate: building type valid for location
- Deduct resources
- Create LocationBuilding
- Update location building_type/building_level
- Return 200

**handleRemoveBuilding** — `DELETE /api/planets/{id}/locations/{id}/building`
- Validate: location exists
- Delete LocationBuilding
- Update location (building_type = null, building_level = 0)
- Return 200

**handleAbandonLocation** — `POST /api/planets/{id}/locations/{id}/abandon`
- Validate: location exists
- Delete LocationBuilding + Location
- Decrement max_locations effective count
- Return 200

**handleGetExpeditionHistory** — `GET /api/planets/{id}/expedition-history`
- Return expedition history list

### 4.3 router.go — routes
**Файл:** `server/internal/api/router.go`

Добавить:
```go
rr := planetGroup.Group("/planet-survey")
rr.Post("", handleStartPlanetSurvey(db))
rr.Get("", handleGetPlanetSurvey(db))

rr2 := planetGroup.Group("/locations")
rr2.Get("", handleGetLocations(db))
rr2.POST("/:id/build", handleBuildOnLocation(db))
rr2.DELETE("/:id/building", handleRemoveBuilding(db))
rr2.POST("/:id/abandon", handleAbandonLocation(db))

rr3 := planetGroup.Group("/expedition-history")
rr3.Get("", handleGetExpeditionHistory(db))
```

Rename expedition routes:
```go
rr := planetGroup.Group("/space-expeditions")
rr.Post("", handleCreateSpaceExpedition(db))
rr.Get("", handleGetSpaceExpeditions(db))

rr2 := authGroup.Group("/space-expeditions")
rr2.POST("/:id/action", handleSpaceExpeditionAction(db))
```

### 4.4 websocket.go — new message types
**Файл:** `server/internal/api/websocket.go`

Добавить:
```go
// Message types
"planet_survey_update"
"location_update"

// Notification types
"location_discovered"
```

### 4.5 Frontend — api_service.dart
**Файл:** `client/lib/services/api_service.dart`

Добавить методы:
```dart
Future<Map<String, dynamic>> startPlanetSurvey(String planetId, int duration);
Future<Map<String, dynamic>> getPlanetSurvey(String planetId);
Future<Map<String, dynamic>> getLocations(String planetId);
Future<Map<String, dynamic>> buildOnLocation(String planetId, String locationId, String buildingType);
Future<Map<String, dynamic>> removeBuilding(String planetId, String locationId);
Future<Map<String, dynamic>> abandonLocation(String planetId, String locationId);
Future<Map<String, dynamic>> getExpeditionHistory(String planetId);
```

### 4.6 Тесты
**Файл:** `server/internal/api/planet_survey_test.go` (NEW)
- Тест: POST /planet-survey — creates expedition
- Тест: POST /planet-survey — rejects if base not working
- Тест: POST /planet-survey — rejects if cooldown active
- Тест: POST /planet-survey — rejects if no resources
- Тест: GET /planet-survey — returns expeditions + range_stats
- Тест: GET /locations — returns locations
- Тест: POST /locations/:id/build — creates building
- Тест: POST /locations/:id/build — rejects if no research
- Тест: DELETE /locations/:id/building — removes building
- Тест: POST /locations/:id/abandon — removes location
- Тест: GET /expedition-history — returns history

**Запуск:**
```bash
go test -timeout 10s ./internal/api/...
```

## Деплой
```bash
./deploy.sh
go test -timeout 10s ./...
```

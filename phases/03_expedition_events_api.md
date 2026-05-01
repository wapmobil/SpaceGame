> **Общий план:** [surface_expeditions_design.md](../surface_expeditions_design.md)

# Фаза 3: API + handlers + routes + broadcast

## Цель
Создать полностью рабочий REST API для expedition event chains. WebSocket broadcast событий.

## Что работает после деплоя
Полностью рабочий API. Frontend может интегрироваться.

## Задачи

### 8.1 API models
**Файл:** `server/internal/api/models.go`

Добавить request типы:
```go
type StartExpeditionRequest struct {
    Inventory map[string]float64 `json:"inventory"`
}

type ExpeditionChoiceRequest struct {
    ChoiceIndex int `json:"choice_index"`
}
```

Добавить response типы:
```go
type ExpeditionChainResponse struct {
    ID                 string                  `json:"id"`
    PlanetID           string                  `json:"planet_id"`
    OwnerID            string                  `json:"owner_id"`
    Status             string                  `json:"status"`
    EventCount         int                     `json:"event_count"`
    CurrentEventIndex  int                     `json:"current_event_index"`
    Inventory          map[string]float64      `json:"inventory"`
    DiscoveredLocation *LocationResponse       `json:"discovered_location,omitempty"`
    CreatedAt          time.Time               `json:"created_at"`
    UpdatedAt          time.Time               `json:"updated_at"`
}

type ExpeditionEventResponse struct {
    EventID           string                  `json:"event_id"`
    Description       string                  `json:"description"`
    ImmediateReward   map[string]float64      `json:"immediate_reward"`
    Choices           []ExpeditionChoiceResp  `json:"choices"`
    IsEnd             bool                    `json:"is_end"`
    LocationReward    string                  `json:"location_reward,omitempty"`
}

type ExpeditionChoiceResp struct {
    Label       string                  `json:"label"`
    Description string                  `json:"description"`
    Reward      map[string]float64      `json:"reward"`
    NextEventID string                  `json:"next_event_id"`
}

type ExpeditionChainListResponse struct {
    Chains   []ExpeditionChainResponse `json:"chains"`
    Total    int                       `json:"total"`
}

type ExpeditionChoiceResult struct {
    Event            *ExpeditionEventResponse `json:"event"`
    Chain            ExpeditionChainResponse  `json:"chain"`
    Inventory        map[string]float64       `json:"inventory"`
    Completed        bool                     `json:"completed"`
    Failed           bool                     `json:"failed"`
    Location         *LocationResponse        `json:"location,omitempty"`
    LocationReward   string                   `json:"location_reward,omitempty"`
    Error            string                   `json:"error,omitempty"`
}
```

Удалить:
- `StartPlanetSurveyRequest`
- `ExpeditionHistoryResponse`
- `ExpeditionRangeStatsResponse`

### 8.2 planet_survey_handlers.go — переработка
**Файл:** `server/internal/api/planet_survey_handlers.go`

#### handleStartExpedition
**Path:** `POST /api/planets/{id}/expeditions`

```go
func handleStartExpedition(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        planetID := PlanetIDFromContext(r)
        
        var req StartExpeditionRequest
        DecodeJSON(r.Body, &req)
        
        p := ensurePlanetLoaded(planetID)
        if p == nil { return Error(...) }
        
        // Проверка: can start expedition
        if !p.CanStartExpedition() {
            if !p.BaseOperational() { return Error("base_not_operational") }
            if _, ok := p.Research.GetCompleted()["planet_exploration"]; !ok {
                return Error("research_not_completed")
            }
        }
        
        // Валидация inventory
        if len(req.Inventory) == 0 {
            return Error("inventory required")
        }
        
        // Списываем ресурсы с планеты
        for res, amount := range req.Inventory {
            switch res {
            case "food": p.Resources.Food -= amount
            case "iron": p.Resources.Iron -= amount
            case "composite": p.Resources.Composite -= amount
            case "mechanisms": p.Resources.Mechanisms -= amount
            case "reagents": p.Resources.Reagents -= amount
            }
        }
        
        // Создаём цепочку
        chain, event, err := p.StartExpeditionChain(req.Inventory)
        if err != nil {
            // Возвращаем ресурсы
            p.ReturnExpeditionInventory(chainID)
            return Error(...)
        }
        
        // Сохраняем в БД
        SaveChainToDB(chain, db)
        SaveEventsToDB(chain.ID, GetEventHistory(chain), db)
        
        // Broadcast event
        wsBroadcast.BroadcastExpeditionEvent(p.OwnerID, map[string]interface{}{
            "chain_id": chain.ID,
            "event":    toMap(event),
            "inventory": chain.Inventory,
            "event_count": chain.EventCount,
        })
        
        JSON(w, 200, map[string]interface{}{
            "chain_id": chain.ID,
            "event":    toMap(event),
            "inventory": chain.Inventory,
            "event_count": chain.EventCount,
        })
    }
}
```

#### handleGetExpeditionChains
**Path:** `GET /api/planets/{id}/expeditions`

```go
func handleGetExpeditionChains(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        planetID := PlanetIDFromContext(r)
        p := ensurePlanetLoaded(planetID)
        
        chains := p.GetExpeditionChains()
        response := make([]ExpeditionChainResponse, 0, len(chains))
        
        for _, c := range chains {
            resp := ExpeditionChainResponse{...}
            if c.DiscoveredLocation != nil {
                resp.DiscoveredLocation = locationToResponse(c.DiscoveredLocation)
            }
            response = append(response, resp)
        }
        
        JSON(w, 200, ExpeditionChainListResponse{Chains: response, Total: len(response)})
    }
}
```

#### handleGetExpeditionEvent
**Path:** `GET /api/planets/{id}/expeditions/{chainID}/event`

```go
func handleGetExpeditionEvent(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        chainID := chi.URLParam(r, "chainID")
        p := ensurePlanetLoaded(planetID)
        
        chain := p.GetActiveExpeditionChain(chainID)
        if chain == nil { return Error("chain_not_found_or_not_active") }
        
        event := GetCurrentEvent(chain)
        if event == nil { return Error("no_current_event") }
        
        JSON(w, 200, ExpeditionEventResponse{...})
    }
}
```

#### handleExpeditionChoice
**Path:** `POST /api/planets/{id}/expeditions/{chainID}/choice`

```go
func handleExpeditionChoice(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        chainID := chi.URLParam(r, "chainID")
        
        var req ExpeditionChoiceRequest
        DecodeJSON(r.Body, &req)
        
        p := ensurePlanetLoaded(planetID)
        chain := p.GetActiveExpeditionChain(chainID)
        if chain == nil { return Error("chain_not_found_or_not_active") }
        
        // Resolve choice
        event, err := p.ResolveExpeditionChoice(chainID, req.ChoiceIndex)
        if err != nil {
            // LLM failed after retries → fail chain + return inventory
            p.ReturnExpeditionInventory(chainID)
            SaveChainToDB(chain, db)
            
            wsBroadcast.BroadcastExpeditionComplete(p.OwnerID, map[string]interface{}{
                "chain_id": chainID,
                "status":   "failed",
                "error":    err.Error(),
            })
            
            JSON(w, 200, ExpeditionChoiceResult{
                Chain:   chainToResponse(chain),
                Failed:  true,
                Error:   err.Error(),
            })
            return
        }
        
        // Save to DB
        SaveChainToDB(chain, db)
        SaveEventsToDB(chain.ID, GetEventHistory(chain), db)
        
        var resp ExpeditionChoiceResult
        if event != nil {
            // Still active — broadcast next event
            resp.Event = eventToResponse(event)
            wsBroadcast.BroadcastExpeditionEvent(p.OwnerID, map[string]interface{}{
                "chain_id": chainID,
                "event":    eventToMap(event),
                "inventory": chain.Inventory,
                "event_count": chain.EventCount,
            })
        } else {
            // Completed
            resp.Completed = true
            resp.Chain = chainToResponse(chain)
            resp.Inventory = chain.Inventory
            if chain.DiscoveredLocation != nil {
                resp.Location = locationToResponse(chain.DiscoveredLocation)
            }
            
            // Inventory возвращается на планету
            ReturnInventoryToPlanet(p, chain.Inventory)
            
            wsBroadcast.BroadcastExpeditionComplete(p.OwnerID, map[string]interface{}{
                "chain_id": chainID,
                "status":   "completed",
                "inventory": chain.Inventory,
                "location": locationToMap(chain.DiscoveredLocation),
            })
        }
        
        JSON(w, 200, resp)
    }
}
```

#### handleGetExpeditionEvents
**Path:** `GET /api/planets/{id}/expeditions/{chainID}/events`

```go
func handleGetExpeditionEvents(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        chainID := chi.URLParam(r, "chainID")
        p := ensurePlanetLoaded(planetID)
        
        chain := p.GetActiveExpeditionChain(chainID)
        if chain == nil {
            // Load from DB
            events, err := LoadChainEvents(chainID, db)
            if err != nil { return Error(...) }
            JSON(w, 200, events)
            return
        }
        
        events := GetEventHistory(chain)
        JSON(w, 200, events)
    }
}
```

#### Оставить без изменений
- `handleGetLocations`
- `handleBuildOnLocation`
- `handleRemoveBuilding`
- `handleAbandonLocation`

### 8.3 router.go — маршруты
**Файл:** `server/internal/api/router.go`

Заменить:
```go
// БЫЛО:
rr.Post("/{id}/planet-survey", handleStartPlanetSurvey(db))
rr.Get("/{id}/planet-survey", handleGetPlanetSurvey(db))
rr.Get("/{id}/expedition-history", handleGetExpeditionHistory(db))

// СТАЛО:
rr.Post("/{id}/expeditions", handleStartExpedition(db))
rr.Get("/{id}/expeditions", handleGetExpeditionChains(db))
rr.Get("/{id}/expeditions/{chainID}/event", handleGetExpeditionEvent(db))
rr.Post("/{id}/expeditions/{chainID}/choice", handleExpeditionChoice(db))
rr.Get("/{id}/expeditions/{chainID}/events", handleGetExpeditionEvents(db))
```

### 8.4 websocket_broadcast.go — broadcast методы
**Файл:** `server/internal/api/websocket_broadcast.go`

Добавить:
```go
func (bs *WSBroadcastService) BroadcastExpeditionEvent(playerID string, data map[string]interface{}) {
    bs.cm.SendToPlayer(playerID, WSMessage{
        Type: "expedition_event",
        Data: json.RawMessage(toJSON(data)),
    })
}

func (bs *WSBroadcastService) BroadcastExpeditionComplete(playerID string, data map[string]interface{}) {
    bs.cm.SendToPlayer(playerID, WSMessage{
        Type: "expedition_complete",
        Data: json.RawMessage(toJSON(data)),
    })
}
```

### 8.5 Тесты
**Файл:** `server/internal/api/planet_survey_test.go` (extend существующий)

- `TestHandleStartExpedition` — creates chain with valid inventory
- `TestHandleStartExpedition` — rejects if base not operational
- `TestHandleStartExpedition` — rejects if research not completed
- `TestHandleStartExpedition` — rejects if inventory > 1000
- `TestHandleGetExpeditionChains` — returns active chains
- `TestHandleGetExpeditionChains` — returns completed chains
- `TestHandleExpeditionChoice` — processes valid choice
- `TestHandleExpeditionChoice` — returns error for invalid choice index
- `TestHandleGetExpeditionEvents` — returns event history

**Запуск:**
```bash
go test -timeout 10s ./internal/api/...
```

## Деплой
```bash
./deploy.sh
go test -timeout 10s ./internal/api/...
go test -timeout 10s ./internal/game/...
```

## Зависимости
- Фаза 7 (expedition_chain.go) должна быть реализована

# Фаза 3: API + handlers + routes + broadcast

## Цель
Создать полностью рабочий REST API для expedition event chains. WebSocket broadcast событий.

## Что работает после деплоя
Полностью рабочий API. Frontend может интегрироваться.

## Задачи

### 3.1 API models

**Файл:** `server/internal/api/models.go`

Добавить типы:
```go
// StartExpeditionChainRequest is the request body for starting an expedition chain.
type StartExpeditionChainRequest struct {
    Inventory map[string]float64 `json:"inventory"`
}

// ExpeditionChoiceRequest is the request body for resolving a choice.
type ExpeditionChoiceRequest struct {
    ChoiceIndex int `json:"choice_index"`
}

// ExpeditionChainResponse is the API response for an expedition chain.
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

// ExpeditionEventResponse is the API response for a single event.
type ExpeditionEventResponse struct {
    EventID         string                  `json:"event_id"`
    Description     string                  `json:"description"`
    ImmediateReward map[string]float64      `json:"immediate_reward"`
    Choices         []ExpeditionChoiceResp  `json:"choices"`
    IsEnd           bool                    `json:"is_end"`
    LocationReward  string                  `json:"location_reward,omitempty"`
}

// ExpeditionChoiceResp is a single choice within an event.
type ExpeditionChoiceResp struct {
    Label       string             `json:"label"`
    Description string             `json:"description"`
    Reward      map[string]float64 `json:"reward"`
    NextEventID string             `json:"next_event_id"`
}

// ExpeditionChainListResponse lists chains for a planet.
type ExpeditionChainListResponse struct {
    Chains []ExpeditionChainResponse `json:"chains"`
    Total  int                       `json:"total"`
}

// ExpeditionChoiceResult is the response after resolving a choice.
type ExpeditionChoiceResult struct {
    Event          *ExpeditionEventResponse `json:"event,omitempty"`
    Chain          ExpeditionChainResponse  `json:"chain"`
    Inventory      map[string]float64       `json:"inventory"`
    Completed      bool                     `json:"completed"`
    Failed         bool                     `json:"failed"`
    Location       *LocationResponse        `json:"location,omitempty"`
    LocationReward string                   `json:"location_reward,omitempty"`
    Error          string                   `json:"error,omitempty"`
}

// ExpeditionEventLogEntry is a history entry for the event log.
type ExpeditionEventLogEntry struct {
    EventID         string             `json:"event_id"`
    Description     string             `json:"description"`
    PlayerChoice    int                `json:"player_choice"`
    ChoiceLabel     string             `json:"choice_label"`
    RewardsReceived map[string]float64 `json:"rewards_received"`
    CreatedAt       time.Time          `json:"created_at"`
}
```

Удалить типы:
- `StartPlanetSurveyRequest`
- `ExpeditionHistoryResponse`
- `ExpeditionRangeStatsResponse`

### 3.2 HTTP timeout для LLM маршрутов

**Файл:** `server/internal/api/router.go`

Маршруты с LLM вызовом (`handleStartExpedition`, `handleExpeditionChoice`) требуют увеличенного timeout. Добавить middleware:

```go
// llmTimeout wraps a handler with an extended write timeout for LLM operations.
func llmTimeout(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // LLM может занять до 300s + retry
        ctx, cancel := context.WithTimeout(r.Context(), 330*time.Second)
        defer cancel()
        h.ServeHTTP(w, r.WithContext(ctx))
    }
}
```

Применить к маршрутам в router.go:
```go
rr.Post("/{id}/expeditions", llmTimeout(handleStartExpedition(db)))
rr.Post("/{id}/expeditions/{chainID}/choice", llmTimeout(handleExpeditionChoice(db)))
```

### 3.3 planet_survey_handlers.go — переработка

**Файл:** `server/internal/api/planet_survey_handlers.go`

Удалить старые handler'ы:
- `handleStartPlanetSurvey`
- `handleGetPlanetSurvey`
- `handleGetExpeditionHistory`

Удалить helper'ы:
- `getMaxDurationForBaseLevel`
- `getRangeForDuration`

Оставить без изменений:
- `handleGetLocations`
- `handleBuildOnLocation`
- `handleRemoveBuilding`
- `handleAbandonLocation`
- `findLocation`
- `getLocationBuildingTypes`

Добавить новые handler'ы:

#### handleStartExpedition
**Path:** `POST /api/planets/{id}/expeditions`

```go
func handleStartExpedition(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        planetID := PlanetIDFromContext(r)

        var req StartExpeditionChainRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            Error(w, http.StatusBadRequest, "Invalid request body")
            return
        }

        p := ensurePlanetLoaded(planetID)
        if p == nil {
            Error(w, http.StatusInternalServerError, "Failed to load planet")
            return
        }

        if !p.CanStartExpedition() {
            if !p.BaseOperational() {
                Error(w, http.StatusBadRequest, "Planet base requires food to operate")
                return
            }
            if _, ok := p.Research.GetCompleted()["planet_exploration"]; !ok {
                Error(w, http.StatusBadRequest, "Planet exploration research not completed")
                return
            }
            Error(w, http.StatusBadRequest, "Active expedition chain already exists")
            return
        }

        // Валидация inventory
        if err := planet_survey.ValidateInventory(req.Inventory); err != nil {
            Error(w, http.StatusBadRequest, err.Error())
            return
        }

        // Проверка достаточности ресурсов
        if !hasResources(p, req.Inventory) {
            Error(w, http.StatusConflict, "Insufficient resources")
            return
        }

        // Списываем ресурсы
        deductResources(p, req.Inventory)

        // Создаём цепочку (LLM вызов внутри, может занять до 300s)
        chain, event, err := p.StartExpeditionChain(req.Inventory)
        if err != nil {
            // Возвращаем ресурсы
            planet_survey.ReturnInventoryToPlanet(p, req.Inventory)
            game.Instance().SavePlanet(p)
            Error(w, http.StatusInternalServerError, "Failed to start expedition: "+err.Error())
            return
        }

        // Сохраняем в БД
        if err := planet_survey.SaveChainToDB(chain, db); err != nil {
            log.Printf("Error saving chain %s: %v", chain.ID, err)
        }
        if err := planet_survey.SaveEventsToDB(chain.ID, planet_survey.GetEventHistory(chain), db); err != nil {
            log.Printf("Error saving events for chain %s: %v", chain.ID, err)
        }

        game.Instance().SavePlanet(p)

        // Broadcast
        wsBroadcast.BroadcastExpeditionEvent(p.OwnerID, map[string]interface{}{
            "chain_id":    chain.ID,
            "event":       eventToResponse(event),
            "inventory":   chain.Inventory,
            "event_count": chain.EventCount,
        })

        JSON(w, http.StatusOK, map[string]interface{}{
            "chain_id":    chain.ID,
            "event":       eventToResponse(event),
            "inventory":   chain.Inventory,
            "event_count": chain.EventCount,
        })
    }
}
```

Helper'ы:
```go
func hasResources(p *game.Planet, inventory map[string]float64) bool {
    for res, amount := range inventory {
        switch res {
        case "food":
            if p.Resources.Food < amount { return false }
        case "iron":
            if p.Resources.Iron < amount { return false }
        case "composite":
            if p.Resources.Composite < amount { return false }
        case "mechanisms":
            if p.Resources.Mechanisms < amount { return false }
        case "reagents":
            if p.Resources.Reagents < amount { return false }
        }
    }
    return true
}

func deductResources(p *game.Planet, inventory map[string]float64) {
    for res, amount := range inventory {
        switch res {
        case "food":      p.Resources.Food -= amount
        case "iron":      p.Resources.Iron -= amount
        case "composite": p.Resources.Composite -= amount
        case "mechanisms": p.Resources.Mechanisms -= amount
        case "reagents":  p.Resources.Reagents -= amount
        }
    }
}
```

#### handleGetExpeditionChains
**Path:** `GET /api/planets/{id}/expeditions`

Возвращает список цепочек (active + recent completed). Для completed цепочек показывает итоговый статус и обнаруженную локацию.

#### handleGetExpeditionEvent
**Path:** `GET /api/planets/{id}/expeditions/{chainID}/event`

Возвращает текущее активное событие цепочки. Если цепочка не активна или события нет — ошибка.

#### handleExpeditionChoice
**Path:** `POST /api/planets/{id}/expeditions/{chainID}/choice`

```go
func handleExpeditionChoice(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        chainID := chi.URLParam(r, "chainID")

        var req ExpeditionChoiceRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            Error(w, http.StatusBadRequest, "Invalid request body")
            return
        }

        p := ensurePlanetLoaded(PlanetIDFromContext(r))
        if p == nil {
            Error(w, http.StatusInternalServerError, "Failed to load planet")
            return
        }

        chain := p.GetActiveExpeditionChain(chainID)
        if chain == nil {
            Error(w, http.StatusNotFound, "Chain not found or not active")
            return
        }

        // Resolve choice (LLM вызов внутри при генерации следующего события)
        event, err := p.ResolveExpeditionChoice(chainID, req.ChoiceIndex)

        if err != nil {
            // LLM failed after retries → fail chain + return inventory
            p.ReturnExpeditionInventory(chainID)
            planet_survey.SaveChainToDB(chain, db)
            game.Instance().SavePlanet(p)

            wsBroadcast.BroadcastExpeditionComplete(p.OwnerID, map[string]interface{}{
                "chain_id": chainID,
                "status":   "failed",
                "error":    err.Error(),
            })

            JSON(w, http.StatusOK, ExpeditionChoiceResult{
                Chain:   chainToResponse(chain),
                Failed:  true,
                Error:   err.Error(),
            })
            return
        }

        // Save to DB
        planet_survey.SaveChainToDB(chain, db)
        planet_survey.SaveEventsToDB(chain.ID, planet_survey.GetEventHistory(chain), db)
        game.Instance().SavePlanet(p)

        if event != nil {
            // Still active — next event generated
            wsBroadcast.BroadcastExpeditionEvent(p.OwnerID, map[string]interface{}{
                "chain_id":    chainID,
                "event":       eventToResponse(event),
                "inventory":   chain.Inventory,
                "event_count": chain.EventCount,
            })

            JSON(w, http.StatusOK, ExpeditionChoiceResult{
                Event:     eventToResponse(event),
                Chain:     chainToResponse(chain),
                Inventory: chain.Inventory,
            })
        } else {
            // Completed
            var locResp *LocationResponse
            if chain.DiscoveredLocation != nil {
                locResp = locationToResponse(chain.DiscoveredLocation)
            }

            // Inventory возвращается на планету
            planet_survey.ReturnInventoryToPlanet(p, chain.Inventory)
            game.Instance().SavePlanet(p)

            wsBroadcast.BroadcastExpeditionComplete(p.OwnerID, map[string]interface{}{
                "chain_id": chainID,
                "status":   "completed",
                "inventory": chain.Inventory,
                "location": locResp,
            })

            JSON(w, http.StatusOK, ExpeditionChoiceResult{
                Chain:          chainToResponse(chain),
                Inventory:      chain.Inventory,
                Completed:      true,
                Location:       locResp,
                LocationReward: chain.Events[len(chain.Events)-1].LocationReward,
            })
        }
    }
}
```

#### handleGetExpeditionEvents
**Path:** `GET /api/planets/{id}/expeditions/{chainID}/events`

Возвращает все события цепочки. Если chain в памяти — из памяти, иначе из БД.

#### handleGetExpeditionEventLog
**Path:** `GET /api/planets/{id}/expeditions/{chainID}/event-log`

Возвращает упрощённый лог событий для истории: event_id, description, player_choice, choice_label, rewards_received, created_at.

Типизированный ответ через `ExpeditionEventLogEntry`.

#### Conversion helper'ы

Добавить private функции:
- `eventToResponse(*planet_survey.ExpeditionEvent) *ExpeditionEventResponse`
- `chainToResponse(*planet_survey.ExpeditionChain) ExpeditionChainResponse`
- `locationToResponse(*planet_survey.Location) *LocationResponse` (если ещё нет)

### 3.4 router.go — маршруты

**Файл:** `server/internal/api/router.go`

Заменить:
```go
// БЫЛО:
rr.Post("/{id}/planet-survey", handleStartPlanetSurvey(db))
rr.Get("/{id}/planet-survey", handleGetPlanetSurvey(db))
rr.Get("/{id}/expedition-history", handleGetExpeditionHistory(db))

// СТАЛО:
rr.Post("/{id}/expeditions", llmTimeout(handleStartExpedition(db)))
rr.Get("/{id}/expeditions", handleGetExpeditionChains(db))
rr.Get("/{id}/expeditions/{chainID}/event", handleGetExpeditionEvent(db))
rr.Post("/{id}/expeditions/{chainID}/choice", llmTimeout(handleExpeditionChoice(db)))
rr.Get("/{id}/expeditions/{chainID}/events", handleGetExpeditionEvents(db))
rr.Get("/{id}/expeditions/{chainID}/event-log", handleGetExpeditionEventLog(db))
```

### 3.5 websocket_broadcast.go — broadcast методы

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

### 3.6 Тесты

**Файл:** `server/internal/api/planet_survey_test.go` (extend существующий)

- `TestHandleStartExpedition` — creates chain with valid inventory
- `TestHandleStartExpedition_BaseNotOperational` — returns 400
- `TestHandleStartExpedition_ResearchNotCompleted` — returns 400
- `TestHandleStartExpedition_InsufficientResources` — returns 409
- `TestHandleStartExpedition_InvalidInventory` — returns 400
- `TestHandleGetExpeditionChains` — returns active chains
- `TestHandleGetExpeditionChains_Completed` — returns completed chains
- `TestHandleExpeditionChoice` — processes valid choice
- `TestHandleExpeditionChoice_InvalidIndex` — returns error
- `TestHandleExpeditionChoice_ChainNotFound` — returns 404
- `TestHandleGetExpeditionEvents` — returns event history
- `TestHandleGetExpeditionEventLog` — returns log with choice labels

**Запуск:**
```bash
go test -timeout 10s ./internal/api/...
go build -o /dev/null ./cmd/server/
```

## Деплой

```bash
./deploy.sh
go test -timeout 10s ./internal/api/...
go test -timeout 10s ./internal/game/...
```

## Зависимости

- Фаза 2 (expedition_chain.go) должна быть реализована

## Примечания

- `llmTimeout` middleware увеличивает context timeout до 330s для маршрутов с LLM.
- `StartExpeditionRequest` (space expeditions) в `models.go` не затрагивается — это другой тип экспедиций.
- Все handler'ы сохраняют chain в БД и вызывают `SavePlanet` для обновления в-memory → DB.

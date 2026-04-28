> **Общий план:** [surface_expeditions_design.md](../surface_expeditions_design.md)

# Фаза 3: Space Expedition rename

## Цель
Переименовать все space expedition типы и референсы. Старые данные не мигрируются (БД пуста).

## Что работает после деплоя
Все space expedition функции работают с новыми типами.

## Задачи

### 3.1 Обновить expedition/expedition.go — типы
**Файл:** `server/internal/game/expedition/expedition.go`

Константы:
```go
// Было:
TypeExploration Type = "exploration"
TypeTrade       Type = "trade"
TypeSupport     Type = "support"

// Стало:
TypeExploration Type = "space_exploration"
TypeTrade       Type = "space_trade"
TypeSupport     Type = "space_support"
```

### 3.2 Обновить planet_expedition.go
**Файл:** `server/internal/game/planet_expedition.go`

- Обновить все проверки типа экспедиции
- Обновить все строковые литералы "exploration", "trade", "support" → новые значения

### 3.3 Обновить api handlers — rename file
**Файл:** `server/internal/api/space_expedition_handlers.go` (RENAMED от expedition_handlers.go)

- Переименовать файл
- Обновить все референсы на новые типы
- Обновить handleCreateExpedition → handleCreateSpaceExpedition
- Обновить handleGetExpeditions → handleGetSpaceExpeditions
- Обновить handleExpeditionAction → handleSpaceExpeditionAction

### 3.4 Обновить WS message type
**Файл:** `server/internal/api/websocket.go`

- `expedition_update` → `space_expedition_update`
- notification type: `"discovery"` → `"space_discovery"`

### 3.5 Обновить frontend — expedition.dart
**Файл:** `client/lib/models/expedition.dart`

- Обновить expedition types в константах/парсинге
- Убедиться что новые типы корректно парсятся

### 3.6 Обновить frontend — constants.dart
**Файл:** `client/lib/utils/constants.dart`

- Обновить `Constants.expeditionTypes` — новые значения

### 3.7 Обновить frontend — game_provider.dart
**Файл:** `client/lib/providers/game_provider.dart`

- Обновить WS handler: `_handleExpeditionUpdate` → `_handleSpaceExpeditionUpdate`
- Обновить все референсы на новые типы

### 3.8 Тесты
**Файл:** `server/internal/game/expedition/expedition_test.go`
- Тест: Новые типы корректно определены
- Тест: Создаётся экспедиция с типом "space_exploration"
- Тест: Создаётся экспедиция с типом "space_trade"
- Тест: Создаётся экспедиция с типом "space_support"

**Файл:** `server/internal/api/space_expedition_test.go` (NEW)
- Тест: handleCreateSpaceExpedition — создаёт экспедицию
- Тест: handleGetSpaceExpeditions — возвращает список
- Тест: handleSpaceExpeditionAction — выполняет действие

**Файл:** `server/internal/api/websocket_test.go`
- Тест: WS message type "space_expedition_update"
- Тест: notification type "space_discovery"

**Запуск:**
```bash
go test -timeout 10s ./internal/game/expedition/...
go test -timeout 10s ./internal/api/...
```

## Деплой
```bash
./deploy.sh
go test -timeout 10s ./...
flutter test
```

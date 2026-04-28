> **Общий план:** [surface_expeditions_design.md](../surface_expeditions_design.md)

# Предфаза: Минорные изменения (ground work)

## Цель
Подготовить ground work для последующих фаз: добавить command_center consumption, переименовать tech expeditions → space_expeditions, обновить описания.

## Что работает после деплоя
Всё как было, но с добавленным `command_center` consumption и renamed tech ID. Никаких новых фич.

## Задачи

### 0.1 Добавить command_center energy/food consumption
**Файл:** `server/internal/game/building/building.go`
- Добавить `command_center` в `EnergyConsumption()`: `level * 50`
- Добавить `command_center` в `Production()`: food consumption `level * 10` (negative)

### 0.2 Переименовать tech expeditions → space_expeditions
**Файл:** `server/internal/game/research/tech.go`
- ID: `"expeditions"` → `"space_expeditions"`
- Description: "Открывает космические экспедиции"

### 0.3 Обновить BuildingResearchRequirements
**Файл:** `server/internal/game/building_entry.go`
- `command_center` уже мапится на `"space_expeditions"` — проверить что совпадает

### 0.4 Обновить planet_exploration description
**Файл:** `server/internal/game/research/tech.go`
- Description: "Unlocks random buildings" → "Открывает систему планетарной разведки"

### 0.5 Тесты
**Файл:** `server/internal/game/building/building_test.go` (NEW)
- Тест: `EnergyConsumption("command_center", 1)` == 50
- Тест: `EnergyConsumption("command_center", 3)` == 150
- Тест: `Production("command_center", 1).Food` == -10
- Тест: `Production("command_center", 3).Food` == -30
- Запуск: `go test -timeout 10s ./internal/game/building/...`

### 0.6 Миграция
**Файл:** `server/internal/db/migrations/020_add_surface_expeditions.up.sql`
```sql
ALTER TABLE planets ADD COLUMN resource_type TEXT NOT NULL DEFAULT 'composite';
ALTER TABLE planets ADD COLUMN max_locations INTEGER NOT NULL DEFAULT 1;
```

## Деплой
```bash
./deploy.sh
go test -timeout 10s ./...
```

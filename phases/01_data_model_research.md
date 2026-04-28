> **Общий план:** [surface_expeditions_design.md](../surface_expeditions_design.md)

# Фаза 1: Data model + Research techs

## Цель
Добавить ResourceType на Planet и новые research techs (location_buildings, advanced_exploration).

## Что работает после деплоя
Новые технологии доступны для исследования. Планеты получают случайный `resource_type`. Без surface expeditions API.

## Задачи

### 1.1 Добавить ResourceType в Planet struct
**Файл:** `server/internal/game/planet.go`

Добавить тип и константы:
```go
type PlanetResourceType string

const (
    ResourceComposite  PlanetResourceType = "composite"
    ResourceMechanisms PlanetResourceType = "mechanisms"
    ResourceReagents   PlanetResourceType = "reagents"
)
```

Добавить поле в Planet struct:
```go
ResourceType PlanetResourceType
```

В `NewPlanet()` — случайный выбор:
```go
types := []PlanetResourceType{ResourceComposite, ResourceMechanisms, ResourceReagents}
p.ResourceType = types[rand.Intn(len(types))]
```

### 1.2 Добавить resource_type в GetState()
**Файл:** `server/internal/game/planet_state.go`

В `GetState()` добавить:
```go
"resource_type": string(p.ResourceType),
```

### 1.3 Добавить tech location_buildings
**Файл:** `server/internal/game/research/tech.go`

В `AllTechs()` добавить:
```go
{
    ID:          "location_buildings",
    Name:        "Location Buildings",
    Description: "Позволяет строить здания на локациях",
    CostFood:    800,
    CostMoney:   600,
    BuildTime:   250,
    Tree:        TreeStandard,
    MaxLevel:    1,
    DependsOn:   []string{"planet_exploration"},
},
```

### 1.4 Добавить tech advanced_exploration
**Файл:** `server/internal/game/research/tech.go`

В `AllTechs()` добавить:
```go
{
    ID:          "advanced_exploration",
    Name:        "Advanced Exploration",
    Description: "+1 слот локаций за уровень (макс. 4)",
    CostFood:    1000,
    CostMoney:   800,
    BuildTime:   300,
    Tree:        TreeStandard,
    MaxLevel:    3,
    DependsOn:   []string{"location_buildings"},
},
```

### 1.5 Тесты
**Файл:** `server/internal/game/planet_test.go` (NEW или extend существующий)
- Тест: `NewPlanet()` — ResourceType один из трёх
- Тест: `planet.ResourceType` не пустой
- Тест: `GetState()` содержит `resource_type`

**Файл:** `server/internal/game/research/research_test.go` (NEW или extend существующий)
- Тест: `GetTechByID("location_buildings")` не nil
- Тест: `GetTechByID("advanced_exploration")` не nil
- Тест: `advanced_exploration.DependsOn` содержит `location_buildings`
- Тест: `advanced_exploration.MaxLevel` == 3

**Файл:** `server/internal/game/building/building_test.go` (создать в предфазе)
- Тест: `EnergyConsumption("command_center", 1)` == 50
- Тест: `EnergyConsumption("command_center", 3)` == 150
- Тест: `Production("command_center", 1).Food` == -10
- Тест: `Production("command_center", 3).Food` == -30

**Запуск:**
```bash
go test -timeout 10s ./internal/game/...
```

## Деплой
```bash
./deploy.sh
go test -timeout 10s ./...
```

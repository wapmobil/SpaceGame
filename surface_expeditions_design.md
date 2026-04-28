# Planet Survey & Space Expedition System — Design Document

## 1. Обзор

Добавляются два типа экспедиций и система локаций:

| Тип | Описание | Исследование |
|-----|----------|-------------|
| **Space Expedition** | Межпланетные экспедиции — используют корабли, обнаруживают NPC-планеты | `space_expeditions` (переименовано из `expeditions`) |
| **Planet Survey** | Локальные экспедиции по планете — без кораблей, обнаруживают локации для строительства | `planet_exploration` (переименовано, было random building unlocks) |

Каждая локация:
- Имеет 1-3 варианта зданий на выбор
- Производит ресурсы планеты, пока не исчерпает свой **исчерпаемый ресурс**
- Может быть заброшена (здание снимается, слот освобождается)
- Сохраняется в БД между сессиями

### Зависимость от зданий

| Тип экспедиций | Требует работающий | UI навигация |
|---------------|-------------------|--------------|
| **Planet Survey** | `base` (база) | Кнопка "экспедиции" в navigation chips появляется когда `base.BuildingLevel > 0` |
| **Space Expedition** | `command_center` (командный центр) | Отдельная кнопка "космические экспедиции" когда `command_center.BuildingLevel > 0` |

- Запуск Planet Survey возможен только когда `base.IsWorking() == true`
- Запуск Space Expedition возможен только когда `command_center.IsWorking() == true`
- Если `base` перестаёт работать (food = 0), planet survey экспедиции **приостанавливаются**
- Если `command_center` перестаёт работать (энергия = 0), space expedition экспедиции **приостанавливаются**
- При восстановлении соответствующего здания — экспедиции **продолжают работу**

---

## 2. Планета: ResourceType

Новое поле на планете определяет её **специализацию** — какой производственный ресурс она производит лучше всего.

```go
type PlanetResourceType string

const (
    ResourceComposite PlanetResourceType = "composite"
    ResourceMechanisms PlanetResourceType = "mechanisms"
    ResourceReagents   PlanetResourceType = "reagents"
)
```

### Влияние на обнаружение локаций

Когда планета имеет ResourceType = `composite`:

| Тип локации | Шанс за тик | Диапазон запаса ресурса |
|-------------|------------|------------------------|
| **Совпадающий** (composite) | **×6** | **100–500** |
| **Другие** (mechanisms, reagents) | **×1** | **20–100** |

Это означает, что планета с композитами будет находить локации для производства композитов в ~6 раз чаще, чем локации для других ресурсов.

---

## 3. Space Expedition (переименование)

### Исследование
| Было | Стало | Cost | BuildTime | DependsOn | MaxLevel |
|------|-------|------|-----------|-----------|----------|
| `expeditions` | `space_expeditions` | 1500 food, 1000 money | 300s | `trade` | 1 |

### Типы экспедиций
| Было | Стало |
|------|-------|
| `exploration` | `space_exploration` |
| `trade` | `space_trade` |
| `support` | `space_support` |

### WS-сообщения
| Было | Стало |
|------|-------|
| `expedition_update` | `space_expedition_update` |
| notification type: `"discovery"` | notification type: `"space_discovery"` |

---

## 4. Planet Survey (surface expedition)

### Исследование
| Было | Стало | Cost | BuildTime | DependsOn | MaxLevel |
|------|-------|------|-----------|-----------|----------|
| `planet_exploration` | `planet_exploration` | 100 food, 100 money | 60s | — | 1 | Открывает **planet survey** систему |

> `planet_exploration` ранее давал random building unlocks. Теперь он открывает **planet survey** систему.

### Структура

```go
type PlanetSurveyExpedition struct {
    ID              string    `json:"id"`
    PlanetID        string    `json:"planet_id"`
    Status          string    `json:"status"` // "active", "discovered", "completed", "failed"
    Progress        float64   `json:"progress"`
    Duration        float64   `json:"duration"` // seconds
    ElapsedTime     float64   `json:"elapsed_time"`
    Discovered      *Location `json:"discovered,omitempty"`
    Range           string    `json:"range"`    // "300s", "600s", "1200s"
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

### Дальность экспедиции (зависит от base level)

| Base lvl | Max duration | Cost per min (food/iron/money) |
|----------|-------------|-------------------------------|
| 1 | 5 min (300s) | 100 / 100 / 10 |
| 2 | 10 min (600s) | 200 / 200 / 20 |
| 3 | 20 min (1200s) | 400 / 400 / 40 |

Стоимость: `costPerMin × durationInMinutes`

| Base lvl | 5 min | 10 min | 20 min |
|----------|-------|--------|--------|
| 1 | 500/500/50 | — | — |
| 2 | 1000/1000/100 | 2000/2000/200 | — |
| 3 | 2000/2000/200 | 4000/4000/400 | 8000/8000/800 |

### Cooldown
30 секунд между запуском экспедиций (как drill mini-game).

### Зависимость от base
- Запуск возможен только когда `base.IsWorking() == true`
- Если base перестаёт работать (food = 0) — экспедиции **приостанавливаются**
- При восстановлении food > 0 — **продолжают работу**

### Лимит локаций
- Максимальное количество активных локаций: `max_locations` (по умолчанию 1)
- `max_locations` увеличивается исследованием `advanced_exploration`: +1 за уровень (макс. 4)
- Когда `len(locations) >= max_locations` — шанс обнаружения новой локации = 0

### Механика обнаружения локации

Проверка происходит **раз в 60 тиков** (1 минуту):

```
checkChance = 45% × diminishingFactor(count)

где diminishingFactor(count) = 1 / (1 + (count / 3)^2)
count = totalExpeditions в данном диапазоне дальности
```

| Экспедиций в диапазоне | diminishingFactor | Итоговый шанс за проверку |
|------------------------|------------------|-------------------------|
| 0 | 1.00 | 45% |
| 1 | 0.90 | 40.5% |
| 2 | 0.75 | 33.75% |
| 3 | 0.50 | 22.5% |
| 5 | 0.28 | 12.6% |
| 10 | 0.09 | 4.05% |

### Ресурсы при завершении экспедиции

Каждая экспедиция (успешная или неудачная) может вернуть ресурсы.

**resourceChance:**
```
resourceChance = 100% / count  если count > 0, иначе 100%
где count = totalExpeditions в данном диапазоне
```

**При неудаче (локация не найдена):**
```
resourceChance = 100% / max(1, count)
Каждый ресурс определяется независимо:
  if rand.Float64() < resourceChance {
    amount = rand.Float64() * maxAmount * baseLevel
  }

  food:       maxAmount = 1000
  iron:       maxAmount = 1000
  money:      maxAmount = 250
  reagents:   maxAmount = 100
  composite:  maxAmount = 100
  mechanisms: maxAmount = 100
```

**При успехе (локация найдена):**
```
resourceChance = resourceChance / 5.0
Те же диапазоны maxAmount, но шанс в 5 раз меньше.
```

| Base lvl | Неудача (count=0) | Успех (count=0) |
|----------|------------------|-----------------|
| 1 | food: 0–1000 (100%) | food: 0–1000 (20%) |
| 2 | food: 0–2000 (100%) | food: 0–2000 (20%) |
| 3 | food: 0–3000 (100%) | food: 0–3000 (20%) |

| Base lvl | Успех (count=3) |
|----------|-----------------|
| 1 | food: 0–1000 (4%) |
| 2 | food: 0–2000 (4%) |
| 3 | food: 0–3000 (4%) |

### Статистика экспедиций

```go
type ExpeditionRangeStats struct {
    TotalExpeditions int     // всего завершённых в этом диапазоне
    LocationsFound   int     // сколько локаций найдено
}
```

Хранится в Planet: `rangeStats map[string]*ExpeditionRangeStats`
Ключи: `"300s"`, `"600s"`, `"1200s"`

Обновляется при завершении экспедиции:
- `TotalExpeditions++`
- Если найдена локация: `LocationsFound++`

---

## 5. Location System

### Структура

```go
type Location struct {
    ID              string    `json:"id"`
    PlanetID        string    `json:"planet_id"`
    OwnerID         string    `json:"owner_id"`
    Type            string    `json:"type"`
    Name            string    `json:"name"`
    BuildingType    string    `json:"building_type,omitempty"`
    BuildingLevel   int       `json:"building_level"`
    BuildingActive  bool      `json:"building_active"`
    SourceResource  string    `json:"source_resource"`
    SourceAmount    float64   `json:"source_amount"`
    SourceRemaining float64   `json:"source_remaining"`
    Active          bool      `json:"active"`
    DiscoveredAt    time.Time `json:"discovered_at"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

### Структура LocationBuilding

```go
type LocationBuilding struct {
    ID             string    `json:"id"`
    LocationID     string    `json:"location_id"`
    BuildingType   string    `json:"building_type"`
    Level          int       `json:"level"`
    Active         bool      `json:"active"`
    BuildProgress  float64   `json:"build_progress"`
    BuildTime      float64   `json:"build_time"`
    CostFood       float64   `json:"cost_food"`
    CostIron       float64   `json:"cost_iron"`
    CostMoney      float64   `json:"cost_money"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

> **Примечание:** `Location.building_type`, `building_level`, `building_active` — кэшированные поля для быстрого доступа. Авторитетный источник — таблица `location_buildings`. При построении/удалении/изменении уровня здания на локации обновляются обе таблицы.

### Механика исчерпания

Каждый тик, если здание активно и `source_remaining > 0`:
1. Рассчитать производство: `production = ProductionPerLevel[building_type, building_level]`
2. Добавить производство в ресурсы планеты (через `calculateResourceProduction()` — energy учитывается в общем балансе)
3. Потребить ресурс локации: `source_remaining -= source_consumption_per_level * level`

Когда `source_remaining <= 0`:
- `building_active = false`
- Локация помечается как "depleted"
- Игрок может снести здание или отказаться от локации

### Действия с локацией

| Действие | Результат |
|----------|----------|
| Снести здание | Здание удаляется, локация остаётся (building_type = null) |
| Построить новое здание | Старое должно быть удалено, выбирается новый тип |
| Забрать локацию | Здание + локация удаляются, слот экспедиции освобождается |

---

## 6. Таблица локаций (20 типов)

### Вес rarity и ResourceType bias

| Rarity | Вес | Примеры |
|--------|-----|---------|
| Обычные | 30 | pond, river, forest, mineral_deposit, valley |
| Необычные | 20 | waterfall, cave, thermal_spring, salt_lake, wind_pass |
| Редкие | 12 | crystal_cave, meteor_crater, sunken_city, glacier, mushroom_forest |
| Экзотические | 6 | crystal_field, cloud_island, underground_lake, radioactive_zone, anomaly_zone |

### Полная таблица

| # | Тип | Название | Здания (выбор 1) | Source Resource | Amount Range | Rarity | Produces |
|---|-----|---------|-----------------|-----------------|-------------|--------|----------|
| 1 | `pond` | Пруд | fish_farm, water_purifier | fish | 400–700 | Обычная | food: 2/level |
| 2 | `river` | Река | fish_farm, irrigation_system, water_plant | water | 800–1200 | Обычная | food: 3/level, reagents: 1/level |
| 3 | `forest` | Лес | lumber_mill, herb_garden, resin_tap | wood | 1200–1800 | Обычная | composite: 1.5/level, food: 1/level |
| 4 | `mineral_deposit` | Залежи руды | mineral_extractor, smelter | ore | 1500–2500 | Обычная | iron: 3/level |
| 5 | `dry_valley` | Сухая долина | solar_farm, wind_turbine | solar_radiation | 800–1200 | Обычная | energy: 20/level |
| 6 | `waterfall` | Водопад | hydro_plant, turbine_station | water | 600–1000 | Необычная | energy: 15/level |
| 7 | `cave` | Пещера | crystal_mine, cave_lab | crystals | 200–400 | Необычная | composite: 1/level, money: 2/level |
| 8 | `thermal_spring` | Термальный источник | geothermal_plant, hot_spring_lab | heat | 500–700 | Необычная | energy: 12/level, reagents: 0.5/level |
| 9 | `salt_lake` | Соляное озеро | salt_pans, chemical_plant | salt | 600–1000 | Необычная | reagents: 2/level, money: 1/level |
| 10 | `wind_pass` | Ветреный перевал | wind_farm, storm_collector | wind | 500–900 | Необычная | energy: 10/level |
| 11 | `crystal_cave` | Кристальная пещера | crystal_harvester, crystal_lab | rare_crystals | 150–250 | Редкая | composite: 3/level, mechanisms: 2/level |
| 12 | `meteor_crater` | Метеоритный кратер | meteor_science_lab, alloy_forge | meteorite_metal | 200–400 | Редкая | mechanisms: 2.5/level, iron: 2/level |
| 13 | `sunken_city` | Затонувший город | salvage_station, ruins_archive, deep_bore | artifacts | 100–200 | Редкая | money: 5/level, reagents: 1/level |
| 14 | `glacier` | Ледниковая пустошь | ice_mine, cryo_lab, frost_turbine | ice | 1000–1400 | Редкая | reagents: 1.5/level, energy: 5/level |
| 15 | `mushroom_forest` | Грибной лес | mushroom_farm, spore_extractor, bio_lab | spores | 500–700 | Редкая | food: 4/level, reagents: 0.5/level |
| 16 | `crystal_field` | Кристальное поле | crystal_array, resonance_amplifier | resonance_crystals | 200–300 | Экзотическая | energy: 25/level, composite: 2/level |
| 17 | `cloud_island` | Облачный остров | cloud_harvester, aerial_platform | clouds | 400–600 | Экзотическая | energy: 18/level, money: 2/level |
| 18 | `underground_lake` | Подземное озеро | aquaculture_base, underground_irrigation, deep_well | underground_water | 700–1100 | Экзотическая | food: 3/level, reagents: 2/level |
| 19 | `radioactive_zone` | Радиоактивная зона | radiation_filter, isotope_plant, waste_converter | radioactive_material | 300–500 | Экзотическая | reagents: 3/level, alien_tech: 0.5/level |
| 20 | `anomaly_zone` | Аномальная зона | anomaly_siphon, containment_unit, anomaly_lab | anomaly_energy | 80–150 | Экзотическая | alien_tech: 1/level, money: 0.5/level |

### Примечания по балансу

- **Обычные локации** — производят 1 основной ресурс + иногда бонусный
- **Необычные** — 1 основной + бонусный ресурс
- **Редкие** — производят 2 ресурса, больше бонусов
- **Экзотические** — производят 2 ресурса, некоторые дают alien_tech

---

## 7. Здания на локациях

### Production per level

> Здания на локациях производят ресурсы планеты так же, как обычные здания: через `calculateResourceProduction()` в planet tick. Энергия зданий на локациях учитывается в общем энергетическом балансе планеты. Каждое здание на локации можно включить/отключить через API.

| Здание | Уровень 1 | Уровень 2 | Уровень 3 | Source Consumption |
|--------|----------|----------|----------|-------------------|
| **fish_farm** | food: 2 | food: 5 | food: 9 | −1 source/level |
| **water_purifier** | food: 1, money: 1 | food: 3, money: 2 | food: 5, money: 4 | −1.5 source/level |
| **irrigation_system** | food: 3, reagents: 1 | food: 6, reagents: 2 | food: 10, reagents: 4 | −2 source/level |
| **water_plant** | reagents: 2, energy: 2 | reagents: 4, energy: 4 | reagents: 7, energy: 7 | −1.5 source/level |
| **lumber_mill** | composite: 1.5 | composite: 3.5 | composite: 6 | −1.5 source/level |
| **herb_garden** | food: 1, composite: 0.5 | food: 2.5, composite: 1 | food: 4, composite: 2 | −2 source/level |
| **resin_tap** | composite: 1, money: 1 | composite: 2.5, money: 2 | composite: 4, money: 4 | −2 source/level |
| **mineral_extractor** | iron: 3 | iron: 6 | iron: 10 | −2 source/level |
| **smelter** | iron: 2, money: 1 | iron: 5, money: 2 | iron: 8, money: 4 | −2.5 source/level |
| **solar_farm** | energy: 18 | energy: 40 | energy: 70 | −2 source/level |
| **wind_turbine** | energy: 12 | energy: 25 | energy: 45 | −1.5 source/level |
| **hydro_plant** | energy: 14 | energy: 30 | energy: 52 | −1.5 source/level |
| **turbine_station** | energy: 18, reagents: 0.5 | energy: 38, reagents: 1 | energy: 65, reagents: 2 | −2 source/level |
| **crystal_mine** | composite: 1, money: 2 | composite: 2.5, money: 4 | composite: 4, money: 7 | −1.5 source/level |
| **cave_lab** | composite: 2, mechanisms: 1 | composite: 4.5, mechanisms: 2 | composite: 8, mechanisms: 4 | −2 source/level |
| **geothermal_plant** | energy: 10, reagents: 0.5 | energy: 22, reagents: 1 | energy: 38, reagents: 2 | −1.5 source/level |
| **hot_spring_lab** | reagents: 2, energy: 3 | reagents: 4, energy: 6 | reagents: 7, energy: 10 | −2 source/level |
| **salt_pans** | reagents: 2, money: 1 | reagents: 4, money: 2 | reagents: 7, money: 4 | −1.5 source/level |
| **chemical_plant** | reagents: 3, money: 1 | reagents: 6, money: 2 | reagents: 10, money: 4 | −2 source/level |
| **wind_farm** | energy: 8 | energy: 18 | energy: 32 | −1 source/level |
| **storm_collector** | energy: 14, money: 0.5 | energy: 30, money: 1 | energy: 52, money: 2 | −1.5 source/level |
| **crystal_harvester** | composite: 2.5, mechanisms: 1.5 | composite: 5.5, mechanisms: 3 | composite: 9, mechanisms: 5 | −1.5 source/level |
| **crystal_lab** | composite: 1.5, mechanisms: 2 | composite: 3.5, mechanisms: 4 | composite: 6, mechanisms: 7 | −2 source/level |
| **meteor_science_lab** | mechanisms: 2, iron: 1 | mechanisms: 4.5, iron: 2 | mechanisms: 8, iron: 4 | −2 source/level |
| **alloy_forge** | mechanisms: 1.5, iron: 2 | mechanisms: 3.5, iron: 4 | mechanisms: 6, iron: 7 | −2.5 source/level |
| **salvage_station** | money: 4, reagents: 1 | money: 9, reagents: 2 | money: 15, reagents: 4 | −1.5 source/level |
| **ruins_archive** | money: 2, composite: 1 | money: 5, composite: 2 | money: 8, composite: 4 | −1 source/level |
| **deep_bore** | reagents: 2, iron: 1 | reagents: 4, iron: 2 | reagents: 7, iron: 4 | −2 source/level |
| **ice_mine** | reagents: 1.5 | reagents: 3.5 | reagents: 6 | −1.5 source/level |
| **cryo_lab** | reagents: 1, energy: 3 | reagents: 2.5, energy: 6 | reagents: 4, energy: 10 | −2 source/level |
| **frost_turbine** | energy: 5, reagents: 1 | energy: 12, reagents: 2 | energy: 20, reagents: 4 | −2 source/level |
| **mushroom_farm** | food: 3 | food: 7 | food: 12 | −1.5 source/level |
| **spore_extractor** | food: 1.5, reagents: 0.5 | food: 3.5, reagents: 1 | food: 6, reagents: 2 | −2 source/level |
| **bio_lab** | food: 1, reagents: 1, alien_tech: 0.2 | food: 2.5, reagents: 2, alien_tech: 0.5 | food: 4, reagents: 3.5, alien_tech: 1 | −2 source/level |
| **crystal_array** | energy: 22 | energy: 48 | energy: 82 | −1.5 source/level |
| **resonance_amplifier** | energy: 15, composite: 1.5 | energy: 32, composite: 3 | energy: 55, composite: 5 | −2 source/level |
| **cloud_harvester** | energy: 15, money: 1 | energy: 32, money: 2 | energy: 55, money: 4 | −1.5 source/level |
| **aerial_platform** | energy: 10, money: 2 | energy: 22, money: 4 | energy: 38, money: 7 | −2 source/level |
| **aquaculture_base** | food: 2.5, reagents: 1.5 | food: 5.5, reagents: 3 | food: 9, reagents: 5 | −1.5 source/level |
| **underground_irrigation** | food: 3, reagents: 1 | food: 6.5, reagents: 2 | food: 11, reagents: 4 | −2 source/level |
| **deep_well** | food: 1.5, reagents: 2 | food: 3, reagents: 4 | food: 5, reagents: 7 | −2 source/level |
| **radiation_filter** | reagents: 2.5 | reagents: 5.5 | reagents: 9 | −1.5 source/level |
| **isotope_plant** | reagents: 2, alien_tech: 0.3 | reagents: 4.5, alien_tech: 0.6 | reagents: 8, alien_tech: 1.2 | −2 source/level |
| **waste_converter** | reagents: 1.5, alien_tech: 0.5 | reagents: 3.5, alien_tech: 1 | reagents: 6, alien_tech: 2 | −2.5 source/level |
| **anomaly_siphon** | alien_tech: 0.8 | alien_tech: 1.8 | alien_tech: 3 | −1 source/level |
| **containment_unit** | alien_tech: 0.5, money: 0.3 | alien_tech: 1.2, money: 0.6 | alien_tech: 2, money: 1 | −1.5 source/level |
| **anomaly_lab** | alien_tech: 1, money: 0.5 | alien_tech: 2.2, money: 1 | alien_tech: 3.8, money: 1.8 | −2 source/level |

---

## 8. Стоимость строительства на локациях

### Базовая стоимость (для уровня 1)

| Rarity | Food | Iron | Money | Build Time |
|--------|------|------|-------|------------|
| Обычная локация | 100 | 50 | 200 | 600s |
| Необычная локация | 200 | 100 | 400 | 900s |
| Редкая локация | 400 | 200 | 800 | 1200s |
| Экзотическая локация | 600 | 300 | 1200 | 1800s |

### Повышение уровня
Каждый следующий уровень стоит в 2.5× больше предыдущего.

### Особенности
- Здания на локациях **НЕ требуют** prereq chain (farm → solar → ...)
- Здания на локациях **НЕ потребляют** энергию планеты
- Max level: 3 (как обычные здания)
- Cooldown между экспедициями: 30 секунд

---

## 9. Исследования

### Существующие (переименование/изменение)

| ID | Name | Было | Cost | BuildTime | DependsOn | MaxLevel | Effect |
|----|------|------|------|-----------|-----------|----------|--------|
| `planet_exploration` | Planet Exploration | random building unlocks | 100 food, 100 money | 60s | — | 1 | Открывает **planet survey** систему |
| `expeditions` | Space Expeditions | — | 1500 food, 1000 money | 300s | `trade` | 1 | Открывает **space expedition** систему (переименовано из `expeditions`) |

### Новые

| ID | Name | Cost | BuildTime | DependsOn | MaxLevel | Effect |
|----|------|------|-----------|-----------|----------|--------|
| `location_buildings` | Location Buildings | 800 food, 600 money | 250s | `planet_exploration` | 1 | Позволяет строить здания на локациях |
| `advanced_exploration` | Advanced Exploration | 1000 food, 800 money | 300s | `location_buildings` | 3 | +1 слот за уровень (всего 1–4) |

### Дерево зависимостей

```
planet_exploration (planet survey)
├── location_buildings
│   └── advanced_exploration (lvl 1-3)

trade
└── space_expeditions (space expedition, переименовано из expeditions)
```

---

## 10. Механика ResourceType

### Генерация планеты

При создании новой планеты случайным выбирается ResourceType из трёх:
```go
types := []PlanetResourceType{"composite", "mechanisms", "reagents"}
planet.ResourceType = types[rand.Intn(3)]
```

### Шанс обнаружения по ResourceType

При обнаружении локации рассчитывается множитель:

| Ситуация | Множитель source amount | Множитель шанса |
|----------|------------------------|-----------------|
| Тип локации совпадает с ResourceType планеты | ×1.5 | ×6 |
| Тип локации не совпадает | ×1 | ×1 |

### Пример

Планета с ResourceType = `composite`:
- Локация `forest` (produces composite): базовый шанс 30 × 6 = **180**
- Локация `cave` (produces composite): базовый шанс 20 × 6 = **120**
- Локация `crystal_cave` (produces composite): базовый шанс 12 × 6 = **72**
- Локация `river` (produces reagents): базовый шанс 30 × 1 = **30**
- Локация `salt_lake` (produces reagents): базовый шанс 20 × 1 = **20**
- Локация `waterfall` (produces energy): базовый шанс 20 × 1 = **20**

### Source amount bias

Сначала применяется **rarity multiplier** к базовому диапазону из таблицы:

| Rarity | Rarity multiplier |
|--------|-------------------|
| Обычная | ×1.0 |
| Необычная | ×1.2 |
| Редкая | ×1.5 |
| Экзотическая | ×2.0 |

Затем применяется **ResourceType bias**:

Для совпадающих локаций: `rarity_adjusted_range × 1.5`
- `forest` (Обычная, composite совпадает): 1200–1800 → ×1.0 → **1200–1800** → ×1.5 → **1800–2700**
- `crystal_cave` (Редкая, composite совпадает): 150–250 → ×1.5 → **225–375** → ×1.5 → **337–562**

Для несовпадающих: `rarity_adjusted_range × 0.5`
- `river` (Обычная, reagents не совпадает): 800–1200 → ×1.0 → **800–1200** → ×0.5 → **400–600**
- `glacier` (Редкая, reagents совпадает): 1000–1400 → ×1.5 → **1500–2100** → ×1.5 → **2250–3150**

---

## 11. История экспедиций

```go
type ExpeditionHistoryEntry struct {
    ID              string
    PlanetID        string
    ExpeditionType  string // "space" | "surface"
    Status          string
    Result          string // "success" | "failed" | "abandoned"
    Discovered      string // location name / NPC planet name
    ResourcesGained map[string]float64
    CreatedAt       time.Time
    CompletedAt     time.Time
}
```

Сохраняется в БД. Доступна через API для отображения истории.

---

## 12. БД: новые таблицы

### surface_expeditions

```sql
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
```

### surface_locations

```sql
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
```

### location_buildings

```sql
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
```

### expedition_history

```sql
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
```

### Изменения в planets

```sql
ALTER TABLE planets ADD COLUMN resource_type TEXT NOT NULL DEFAULT 'composite';
ALTER TABLE planets ADD COLUMN max_locations INTEGER NOT NULL DEFAULT 1;
```

### Изменения в expeditions

```sql
ALTER TABLE expeditions ADD COLUMN expedition_type TEXT NOT NULL DEFAULT 'space_exploration';
-- Обновить существующие данные:
UPDATE expeditions SET expedition_type = 'space_exploration' WHERE expedition_type = 'exploration';
```

---

## 13. UI: Navigation Chips и экран экспедиций

### Navigation chips

| Здание | Кнопка | Появляется когда |
|--------|--------|-----------------|
| **base** (база) | **"разведка"** (planet survey) | `base.BuildingLevel > 0` |
| **command_center** (командный центр) | **"космические экспедиции"** (space expedition) | `command_center.BuildingLevel > 0` |

### Экран экспедиций

Открывается по кнопке "экспедиции" (base). Содержит:

| Вкладка | Содержимое |
|---------|-----------|
| **Planet Survey** | Список активных экспедиций, кнопка "запустить", история, список локаций, доступные здания |
| **Space Expedition** | Список космических экспедиций, кнопка "запустить", NPC-планы, действия |

### Состояние зданий

| Здание | Состояние | Поведение |
|--------|-----------|----------|
| **base** | Working (food > 0) | Planet survey тикают, можно запускать новые |
| **base** | Paused (food = 0) | Planet survey приостановлены, таймер не идёт, нельзя запускать новые |
| **base** | Destroyed (level = 0) | Planet survey приостановлены, кнопка "запустить" неактивна |
| **command_center** | Working (energy > 0) | Space expedition тикают, можно запускать новые |
| **command_center** | Paused (energy = 0) | Space expedition приостановлены, таймер не идёт, нельзя запускать новые |
| **command_center** | Destroyed (level = 0) | Space expedition приостановлены, кнопка "запустить" неактивна |

---

## 14. Building Definition: command_center

Новое здание для планеты — `command_center` (командный центр).

### Характеристики

| Параметр | Значение |
|----------|---------|
| Type | `command_center` |
| Level | 1-3 |
| Build Time | 600s |
| Energy Consumption | level × 50 |
| Food Consumption | level × 10 |
| Research Requirements | `space_expeditions` |

### Назначение
- Открывает возможность строить space expeditions
- Без работающего command_center невозможно запускать космические экспедиции
- При остановке food > 0 — base перестаёт работать, planet survey приостанавливаются
- При остановке energy > 0 — command_center перестаёт работать, space expedition приостанавливаются

### Влияние на planet state
- `command_center_level` — уровень здания (0 если не построено)
- `can_start_space_expedition` — true когда `command_center_level > 0` и `command_center.IsWorking()`

---

## 15. API Endpoints

| Method | Path | Handler | Описание |
|--------|------|---------|----------|
| `POST` | `/api/planets/{id}/planet-survey` | `handleStartPlanetSurvey` | Запустить planet survey (body: `{duration: 300}`) |
| `GET` | `/api/planets/{id}/planet-survey` | `handleGetPlanetSurvey` | Список экспедиций + rangeStats |
| `GET` | `/api/planets/{id}/locations` | `handleGetLocations` | Список локаций |
| `POST` | `/api/planets/{id}/locations/{id}/build` | `handleBuildOnLocation` | Построить здание на локации |
| `DELETE` | `/api/planets/{id}/locations/{id}/building` | `handleRemoveBuilding` | Снести здание |
| `POST` | `/api/planets/{id}/locations/{id}/abandon` | `handleAbandonLocation` | Забрать локацию |
| `GET` | `/api/planets/{id}/expedition-history` | `handleGetExpeditionHistory` | История экспедиций |
| `POST` | `/api/planets/{id}/space-expeditions` | `handleStartSpaceExpedition` | Запустить космическую экспедицию (переименовано) |
| `GET` | `/api/planets/{id}/space-expeditions` | `handleGetSpaceExpeditions` | Список космических экспедиций (переименовано) |
| `POST` | `/api/planets/{id}/space-expeditions/{id}/action` | `handleSpaceExpeditionAction` | Действие в точке интереса (переименовано) |

### POST /planet-survey request

```json
{
    "duration": 300  // seconds, must be <= maxDuration для base level
}
```

### Response (start)

```json
{
    "status": "started",
    "expedition_id": "...",
    "duration": 300,
    "cost_food": 500,
    "cost_iron": 500,
    "cost_money": 50
}
```

### Response (get)

```json
{
    "expeditions": [...],
    "range_stats": {
        "300s": { "total_expeditions": 5, "locations_found": 2 },
        "600s": { "total_expeditions": 3, "locations_found": 1 },
        "1200s": { "total_expeditions": 0, "locations_found": 0 }
    },
    "max_duration": 300,
    "cost_per_min": { "food": 100, "iron": 100, "money": 10 }
}
```

---

## 16. WebSocket сообщения

| Тип | Структура | Когда |
|-----|-----------|-------|
| `planet_survey_update` | `{expedition_id, progress, status, discovered_location?}` | Каждый тик активной экспедиции |
| `space_expedition_update` | `{expedition_id, progress, status}` | Каждый тик космической экспедиции (переименовано) |
| `location_update` | `{location_id, building_type, source_remaining, building_active}` | Изменение локации |
| `notification` (type: `location_discovered`) | `{message, location_id, location_type}` | Обнаружена новая локация |
| `notification` (type: `space_discovery`) | `{message, expedition_id, npc_type}` | Обнаружена NPC-планета (переименовано) |

---

## 17. План реализации

### Фаза 1: Data model и определения

**1.1. Planet ResourceType**
- Добавить `ResourceType` в `Planet` struct
- Добавить в `NewPlanet()` — случайный выбор
- Добавить в `planet_state.go` — сериализация
- Добавить в `game.go` — save/load (JSONB в `planets`)
- Миграция: `ALTER TABLE planets ADD COLUMN resource_type TEXT NOT NULL DEFAULT 'composite'`

**1.2. Planet survey package**
- Создать `server/internal/game/planet_survey/planet_survey.go`
- Определить `PlanetSurveyExpedition` struct, `Tick()`, `IsExpired()`
- Создать `location.go` — `Location` struct, discovery logic
- Создать `location_building.go` — `LocationBuilding` struct (building_type, level, active, build_progress, build_time, cost_food, cost_iron, cost_money), tick logic, depletion
- Создать `buildings.go` — static map LocationType → []LocationBuildingDef

**1.3. Building definitions**
- Определить все 30-40 зданий для локаций в `buildings.go`
- Каждое: cost, build_time, production per level, source_consumption per level, max_level

**1.4. Discovery logic**
- `ExpeditionRangeStats` struct — хранение статистики по диапазонам
- `CalculateDiscoveryChance(count)` — `45% × 1/(1+(count/3)^2)`
- `SelectLocationType(weightedTypes, planetResourceType)` — взвешенный рандом с rarity bias + ResourceType bias (×6 совпадение, ×0.5 source amount)
- `GenerateName(locType)` — генерация имени
- `GetResourceChance(count)` — `100% / count` если count > 0, иначе 100%
- `CalculateResourceRecovery(baseLevel, count, isSuccess)` — случайные ресурсы с diminishing

**1.5. Location building logic**
- `TickLocationBuildings()` — production + depletion
- `RemoveBuilding(locationID)` — удалить здание
- `AbandonLocation(locationID)` — удалить здание + локацию

**1.6. Building dependency integration**
- Добавить `IsWorking()` на BuildingEntry — возвращает true если building работает (enabled, level > 0, not building, not complete)
- Добавить проверку `base.IsWorking()` перед запуском planet survey экспедиций
- Добавить проверку `command_center.IsWorking()` перед запуском space expedition экспедиций
- При food = 0 (base не работает) — приостановить planet survey экспедиции (не завершать, просто не тикать)
- При восстановлении food > 0 — продолжить planet survey экспедиции
- При energy = 0 (command_center не работает) — приостановить space expedition экспедиции
- При восстановлении energy > 0 — продолжить space expedition экспедиции

**1.7. Command Center building**
- Добавить `command_center` в `BuildingsOrder` (building_entry.go)
- Добавить `command_center` в `BuildingResearchRequirements`: `"command_center": "space_expeditions"`
- Определить cost, build_time (600s), production, energy_consumption (level × 50), food_consumption (level × 10) в `building/building.go`

**1.8. New research techs**
- Добавить `location_buildings` и `advanced_exploration` в `AllTechs()` (tech.go)
- `advanced_exploration.MaxLevel = 3`, `DependsOn: ["location_buildings"]`

**1.9. Remove old planet_exploration random unlock code**
- Удалить генерацию random building unlocks из `applyResearchEffects()` (planet.go:116-121)
- Удалить логику генерации random unlock из `LoadPlanetFromDB` / `LoadPlanetsFromDB` (game.go:157-161)
- Удалить `RandomUnlockBuildings` переменную из building_entry.go (строка 38)
- Обновить `IsBuildingUnlocked()` — убрать проверку random unlocks
- Обновить описание `planet_exploration` tech (tech.go:32): `"Открывает систему планетарной разведки"`

### Фаза 2: Интеграция с Planet

**2.1. Planet struct**
- Добавить `PlanetSurveyExpeditions []*PlanetSurveyExpedition`
- Добавить `Locations []*Location`
- Добавить `ExpeditionHistory []ExpeditionHistoryEntry`
- Добавить `RangeStats map[string]*ExpeditionRangeStats` — статистика по дальностям
- Добавить `ExpeditionCooldown int64` — unix timestamp для cooldown (30s)

**2.2. Planet tick**
- Добавить `TickPlanetSurvey()` — между step 6 и step 7
- Добавить `TickLocationBuildings()` — в tickResources после calculateResourceProduction
- `TickPlanetSurvey()`: для каждой активной экспедиции — каждые 60 тиков проверять шанс обнаружения (`45% × diminishingFactor(count)`)
- `CompleteExpedition(exp, success bool)`: обновить статистику, начислить ресурсы (resourceChance = 100%/count, ranges по baseLevel), снять cooldown

**2.3. Resource recovery**
- `ApplyResourceRecovery(baseLevel, count, isSuccess)` — каждый ресурс: шанс = resourceChance / (5 если success), amount = rand(0, maxAmount * baseLevel)
- maxAmount: food/iron=1000, money=250, reagents/composite/mechanisms=100
- Ресурсы начисляются сразу при завершении экспедиции

**2.3. Planet state**
- Добавить в `GetState()`: `planet_survey`, `locations`, `max_locations`, `resource_type`, `base_level`, `command_center_level`

**2.4. Space expedition rename**
- Обновить типы в `expedition/expedition.go`
- `expedition.Type` → `SpaceExpeditionType`
- Значения: `space_exploration`, `space_trade`, `space_support`
- Обновить все референсы в `api/`, `game/`, `scheduler/`

**2.5. UI: Navigation chips**
- Добавить в `GetState()`: `base_level`, `can_start_planet_survey`, `can_start_space_expedition`
- Frontend: когда `base_level > 0` показывать кнопку "разведка" (planet survey) в navigation chips
- Frontend: когда `command_center_level > 0` показывать кнопку "космические экспедиции" (space expedition) в navigation chips
- Экран экспедиций с двумя вкладками: Planet Survey / Space Expedition

### Фаза 3: Persistence

**3.1. Миграции**
- Создать `NNN_add_planet_survey_system.up.sql`:
  - `planet_survey` table
  - `surface_locations` table
  - `location_buildings` table
  - `expedition_history` table
  - `resource_type` column
  - `max_locations` column
  - `expedition_type` column + data migration (exploration → space_exploration, trade → space_trade, support → space_support)
- Создать `NNN_update_research_table.up.sql`:
  - Переименовать `expeditions` → `space_expeditions` в `tech.go`
  - Обновить `planet_exploration` эффект (убрать random building unlocks)
- Создать `NNN_add_location_buildings_to_buildings_check.up.sql` — добавить новые building types в CHECK constraint

**3.2. Save/Load**
- `savePlanet()` — сохранять planet survey expeditions, locations, buildings, history, rangeStats (JSONB в `planets`)
- `loadPlanet()` — загружать из DB
- Space expeditions — начать сохранять в DB (были in-memory)
- Range stats — сохранять/загружать как JSONB в `planets.range_stats`

### Фаза 4: API

**4.1. Handlers**
- `server/internal/api/planet_survey_handlers.go`:
  - `handleStartPlanetSurvey` — POST, принимает `duration` (в секундах), проверяет base lvl → costPerMin, deducts resources
  - `handleGetPlanetSurvey` — GET, список экспедиций + rangeStats
  - `handleGetLocations` — GET, список локаций
  - `handleBuildOnLocation` — POST, построить здание на локации
  - `handleRemoveBuilding` — DELETE, снести здание
  - `handleAbandonLocation` — POST, забрать локацию
  - `handleGetExpeditionHistory` — GET, история экспедиций

**4.2. Router**
- Добавить новые маршруты
- Переименовать существующие (expedition → space_expedition)

**4.3. WebSocket**
- Обновить типы сообщений
- `expedition_update` → `space_expedition_update`
- Добавить `surface_expedition_update`, `location_update`

### Фаза 5: Тесты

- `surface_expedition/surface_expedition_test.go` — tick, discovery
- `surface_expedition/location_test.go` — depletion, building production
- `api/surface_expedition_test.go` — API endpoints

### Фаза 6: Frontend (отдельно)

- Новая секция "Surface Expeditions" на странице планеты
- Карточки локаций
- Прогресс-бары экспедиций
- Модальные окна выбора зданий

---

## 18. Итоговая структура файлов

```
server/
├── internal/
│   ├── game/
│   │   ├── planet_survey/
│   │   │   ├── planet_survey.go            # PlanetSurveyExpedition, Tick
│   │   │   ├── location.go                 # Location, discovery
│   │   │   ├── location_building.go        # LocationBuilding, depletion
│   │   │   ├── buildings.go                # Static building definitions
│   │   │   └── history.go                  # ExpeditionHistory
│   │   ├── planet.go                       # + ResourceType, PlanetSurveyExpeditions, Locations
│   │   ├── planet_tick.go                  # + TickPlanetSurvey, TickLocationBuildings
│   │   ├── planet_state.go                 # + planet_survey, locations, command_center_level
│   │   ├── expedition/
│   │   │   └── expedition.go               # Rename: Type→SpaceExpeditionType
│   │   └── game.go                         # + save/load planet survey system
│   ├── api/
│   │   ├── planet_survey_handlers.go       # NEW: все planet survey endpoints
│   │   ├── space_expedition_handlers.go    # RENAMED: expedition_handlers.go
│   │   ├── router.go                       # + новые маршруты, rename
│   │   └── websocket.go                    # + новые WS типы
│   └── db/
│       └── migrations/
│           ├── NNN_add_planet_survey_system.up.sql
│           ├── NNN_update_research_table.up.sql
│           └── NNN_add_location_buildings_check.up.sql
└── migrations/                              # Дубликат для deploy
```

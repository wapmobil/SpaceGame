# SpaceGame - План реализации

## Архитектура
- **Backend:** Go + PostgreSQL + WebSocket
- **Frontend:** Flutter
- **Мультиплеер:** Асинхронный
- **Авторизация:** Custom auth (unique player ID)

---

## Фаза 1: Backend foundation ✅
- [x] Go проект структура (`server/`)
- [x] PostgreSQL миграции
- [x] Custom auth (UUID v7 + session tokens)
- [x] Player/planet CRUD
- [x] WebSocket handler
- [x] Scheduler (game tick, ratings)
- [x] Flutter проект структура (`client/`)
- [x] WebSocket manager в Flutter
- [x] Базовые модели данных

## Фаза 2: Core game logic ✅
- [x] 7 ресурсов (food, composite, mechanisms, reagents, energy, money, alien_tech)
- [x] 8 типов зданий (farm, solar, storage, base, factory, energy_storage, shipyard, comcenter)
- [x] Система энергии (production vs consumption)
- [x] Production tick (каждую секунду)
- [x] Storage capacity
- [x] Building construction progress
- [x] 14 unit тестов

## Фаза 3: Research system ✅
- [x] Дерево 1 — Standard Research (10 технологий)
  - Planet Exploration, Energy Storage, Energy Saving (4 уровня)
  - Trade, Ships, Upgraded Energy Storage (3 уровня)
  - Fast Construction (3 уровня), Compact Storage (3 уровня)
  - Expeditions, Command Center
- [x] Дерево 2 — Alien Technology (3 технологии)
  - Alien Technologies, Additional Expedition, Super Energy Storage (5 уровней)
- [x] Проверка пререквизитов
- [x] Стоимость в ресурсах + время постройки
- [x] API: GET/POST research
- [x] 12 unit тестов

## Фаза 4: Ships & Fleet ✅
- [x] 6 типов кораблей (trade, small, interceptor, corvette, frigate, cruiser)
- [x] Характеристики кораблей (slots, cargo, energy, hp, armor, weapon)
- [x] Стоимость постройки каждого типа
- [x] Fleet management (очередь постройки, энергия, лимиты)
- [x] Shipyard integration
- [x] API: GET fleet, POST build ship
- [x] Unit тесты

---

## Фаза 5: Battle System ✅
**Автоматические бои между флотами (auto-battle)**

### Что реализовать:
- [x] Упрощённая структура боя (attacker/defender, result, loot)
- [x] Auto-battle режим (без пошаговой сетки)
- [x] Расчёт урона (weapon damage - armor, минимум 0)
- [x] Распределение урона пропорционально HP между типами кораблей
- [x] Ресайкл уничтоженных кораблей (возврат 50% ресурсов)
- [x] Tick-based запуск боёв (каждые 60 секунд)
- [x] API endpoint: GET /api/planets/:id/battles
- [x] Unit тесты (15+ тестов)

### Reference:
- `/home/andrey/projects/SpaceTimeBot/battle.qs` — боевая система

---

## Фаза 6: Expeditions ✅
**Экспедиции: разведка, торговля, поддержка**

### Что реализовать:
- [x] Expedition struct (planet_id, fleet_id, target, progress, status)
- [x] 3 типа экспедиций:
  - Exploration — обнаружение NPC планет
  - Trade — торговля через маркетплейс
  - Support — помощь другим игрокам
- [x] NPC планеты (генерация, ресурсы, вражеские корабли)
- [x] Механика обнаружения (шанс зависит от длительности и размера флота)
- [x] Точки интереса: abandoned stations, debris, asteroids, unknown planets, alien bases
- [x] Действия в точках: loot, attack, wait for reinforcements
- [x] Energy cost для экспедиций
- [x] API endpoints:
  - POST /api/planets/:id/expeditions — создать экспедицию
  - GET /api/planets/:id/expeditions — список экспедиций
  - POST /api/expeditions/:id/action — действие в экспедиции
- [x] Unit тесты (11 тестов)

### Reference:
- `/home/andrey/projects/SpaceTimeBot/main.qs` — expedition handling
- `/home/andrey/projects/SpaceTimeBot/npcplanets.qs` — NPC планеты

---

## Фаза 7: Marketplace ✅
**Маркетплейс с ордерами buy/sell**

### Что реализовать:
- [x] MarketOrder struct (planet_id, resource, type, amount, price)
- [x] Public и private ордера (private — по ссылке)
- [x] Резервирование ресурсов при создании ордера
- [x] Cost: 50 energy для создания/удаления ордера
- [x] NPC traders (3 AI игрока с ордерами)
- [x] Global market view (агрегация всех видимых ордеров)
- [x] Matching engine (автоматическое исполнение ордеров)
- [x] API endpoints:
  - POST /api/planets/:id/market/orders — создать ордер
  - GET /api/planets/:id/market/orders — мои ордера
  - GET /api/market — глобальный рынок
  - DELETE /api/market/orders/:id — удалить ордер
- [x] Unit тесты (38 тестов)

### Reference:
- `/home/andrey/projects/SpaceTimeBot/stock.qs` — маркетплейс

---

## Фаза 8: Mining Mini-game ✅
**Dungeon crawler мини-игра**

### Что реализовать:
- [x] Procedural maze generation (tracer algorithm, 13x13)
- [x] Player navigation (стрелки: single step / slide to wall)
- [x] Entities: walls, empty space, monsters (rats, bats, aliens), hearts, bombs, money
- [x] Monster encounters (урон, награда деньгами)
- [x] Bombs для разрушения стен
- [x] Exit (🚪) для сбора денег
- [x] Cooldown: 30 сек (production) / 5 мин (development)
- [x] Bonus multiplier based on Base level
- [x] API endpoints:
  - POST /api/planets/:id/mining/start — начать забег
  - POST /api/planets/:id/mining/move — движение
  - GET /api/planets/:id/mining — состояние
- [x] Unit тесты (16 тестов)

### Reference:
- `/home/andrey/projects/SpaceTimeBot/mininig.qs` — dungeon crawler

---

## Фаза 9: WebSocket Integration ✅
**Real-time state sync**

### Что реализовать:
- [x] WebSocket connection management (per player)
- [x] Broadcast planet state updates (resources, buildings, energy)
- [x] Broadcast battle updates (grid, phase, turn)
- [x] Broadcast expedition updates (progress, discoveries)
- [x] Broadcast notifications (random events, alerts)
- [x] Client reconnection handling
- [x] Message queue for offline state (50 messages, 5-min window)
- [x] Rate limiting (10 msg/sec per client)
- [x] Health check endpoint

### WebSocket protocol:
```json
// Client → Server
{"type": "build", "data": {"building": "farm", "level": 1}}
{"type": "research", "data": {"tech": "energy_storage"}}
{"type": "battle_action", "data": {"action": "move", "ship": 3, "pos": [2,5]}}
{"type": "build_ship", "data": {"ship_type": "corvette"}}
{"type": "start_expedition", "data": {"fleet_id": "...", "target": "..."}}

// Server → Client
{"type": "state_update", "data": {"planet_id": "...", "resources": {...}}}
{"type": "battle_update", "data": {"battle_id": "...", "grid": [...], "phase": "move"}}
{"type": "notification", "data": {"message": "Enemy fleet spotted!", "type": "alert"}}
{"type": "expedition_update", "data": {"expedition_id": "...", "progress": 0.5}}
```

---

## Фаза 10: Flutter Frontend ✅
**Все экраны приложения**

### Экраны:
- [x] **HomeScreen** — список планет, общие ресурсы, навигация
- [x] **PlanetScreen** — здания, производство, энергия, ресурсы
- [x] **ShipyardScreen** — постройка кораблей, очередь, флот
- [x] **ResearchScreen** — дерево технологий, прогресс
- [x] **BattleScreen** — пошаговый бой на сетке 7x7
- [x] **ExpeditionScreen** — экспедиции, карта NPC планет
- [x] **MarketScreen** — маркетплейс, ордера
- [x] **MiningScreen** — dungeon mini-game
- [x] **SettingsScreen** — настройки, профиль

### Компоненты:
- [x] WebSocket client manager
- [x] Resource display widgets
- [x] Building grid widget
- [x] Battle grid widget (7x7)
- [x] Research tree widget
- [x] Ship card widget
- [x] Market order list widget
- [x] Mining maze widget

---

## Фаза 11: Ratings, Events, Polish ✅
**Финальная полировка**

### Что реализовать:
- [x] Leaderboards (money, food, ships, buildings, total resources)
- [x] Обновление рейтингов каждые 5 минут
- [x] Random events:
  - Short circuit (reset energy, cost to fix)
  - Theft (lose 5-20% money)
  - Storage roof collapse (lose 5-20% resources)
  - Mine collapse (lose mine level)
- [x] Statistics tracking (30+ metrics)
- [x] Daily stats reset (6:00)
- [x] NPC market refresh (every 15 sec)
- [x] UI polish и анимации
- [x] Error handling
- [x] Performance optimization

---

## Итого
- **Сделано:** 11/11 фаз ✅
- **Осталось:** 0 фаз
- **Тестов пройдено:** 120+ (Go)

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

## Фаза 5: Battle System
**Пошаговые тактические бои на сетке 7x7**

### Что реализовать:
- [ ] Структура боя (attacker/defender, grid, phase, turn)
- [ ] 2 фазы хода: Move → Attack
- [ ] Перемещение кораблей (left, right, back, front, center, flanking)
- [ ] Атака с разным уроном (зависит от дистанции)
- [ ] Расчёт урона (weapon damage - armor)
- [ ] Распределение урона пропорционально HP между типами кораблей
- [ ] Ресайкл уничтоженных кораблей (возврат 50% ресурсов)
- [ ] AI для NPC кораблей
- [ ] Auto-battle режим
- [ ] Таймер хода (60 сек для NPC)
- [ ] WebSocket events для обновления боя
- [ ] API endpoints:
  - POST /api/planets/:id/battle/start — начать бой
  - POST /api/planets/:id/battle/:id/action — действие в бою
  - GET /api/planets/:id/battles — список активных боёв
- [ ] Unit тесты (расчёт урона, перемещение, завершение боя)

### Reference:
- `/home/andrey/projects/SpaceTimeBot/battle.qs` — боевая система

---

## Фаза 6: Expeditions
**Экспедиции: разведка, торговля, поддержка**

### Что реализовать:
- [ ] Expedition struct (planet_id, fleet_id, target, progress, status)
- [ ] 3 типа экспедиций:
  - Exploration — обнаружение NPC планет
  - Trade — торговля через маркетплейс
  - Support — помощь другим игрокам
- [ ] NPC планеты (генерация, ресурсы, вражеские корабли)
- [ ] Механика обнаружения (шанс зависит от длительности и размера флота)
- [ ] Точки интереса: abandoned stations, debris, asteroids, unknown planets, alien bases
- [ ] Действия в точках: loot, attack, wait for reinforcements
- [ ] Energy cost для экспедиций
- [ ] API endpoints:
  - POST /api/planets/:id/expeditions — создать экспедицию
  - GET /api/planets/:id/expeditions — список экспедиций
  - POST /api/expeditions/:id/action — действие в экспедиции
- [ ] Unit тесты

### Reference:
- `/home/andrey/projects/SpaceTimeBot/main.qs` — expedition handling
- `/home/andrey/projects/SpaceTimeBot/npcplanets.qs` — NPC планеты

---

## Фаза 7: Marketplace
**Маркетплейс с ордерами buy/sell**

### Что реализовать:
- [ ] MarketOrder struct (planet_id, resource, type, amount, price)
- [ ] Public и private ордера (private — по ссылке)
- [ ] Резервирование ресурсов при создании ордера
- [ ] Cost: 50 energy для создания/удаления ордера
- [ ] NPC traders (3 AI игрока с ордерами)
- [ ] Global market view (агрегация всех видимых ордеров)
- [ ] Matching engine (автоматическое исполнение ордеров)
- [ ] API endpoints:
  - POST /api/planets/:id/market/orders — создать ордер
  - GET /api/planets/:id/market/orders — мои ордера
  - GET /api/market — глобальный рынок
  - DELETE /api/market/orders/:id — удалить ордер
- [ ] Unit тесты

### Reference:
- `/home/andrey/projects/SpaceTimeBot/stock.qs` — маркетплейс

---

## Фаза 8: Mining Mini-game
**Dungeon crawler мини-игра**

### Что реализовать:
- [ ] Procedural maze generation (tracer algorithm, 13x13)
- [ ] Player navigation (стрелки: single step / slide to wall)
- [ ] Entities: walls, empty space, monsters (rats, bats, aliens), hearts, bombs, money
- [ ] Monster encounters (урон, награда деньгами)
- [ ] Bombs для разрушения стен
- [ ] Exit (🚪) для сбора денег
- [ ] Cooldown: 30 сек (production) / 5 мин (development)
- [ ] Bonus multiplier based on Base level
- [ ] API endpoints:
  - POST /api/planets/:id/mining/start — начать забег
  - POST /api/planets/:id/mining/:id/move — движение
  - GET /api/planets/:id/mining — состояние
- [ ] Unit тесты (генерация лабиринта, коллизии)

### Reference:
- `/home/andrey/projects/SpaceTimeBot/mininig.qs` — dungeon crawler

---

## Фаза 9: WebSocket Integration
**Real-time state sync**

### Что реализовать:
- [ ] WebSocket connection management (per player)
- [ ] Broadcast planet state updates (resources, buildings, energy)
- [ ] Broadcast battle updates (grid, phase, turn)
- [ ] Broadcast expedition updates (progress, discoveries)
- [ ] Broadcast notifications (random events, alerts)
- [ ] Client reconnection handling
- [ ] Message queue for offline state
- [ ] Rate limiting
- [ ] Health check endpoint

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

## Фаза 10: Flutter Frontend
**Все экраны приложения**

### Экраны:
- [ ] **HomeScreen** — список планет, общие ресурсы, навигация
- [ ] **PlanetScreen** — здания, производство, энергия, ресурсы
- [ ] **ShipyardScreen** — постройка кораблей, очередь, флот
- [ ] **ResearchScreen** — дерево технологий, прогресс
- [ ] **BattleScreen** — пошаговый бой на сетке 7x7
- [ ] **ExpeditionScreen** — экспедиции, карта NPC планет
- [ ] **MarketScreen** — маркетплейс, ордера
- [ ] **MiningScreen** — dungeon mini-game
- [ ] **SettingsScreen** — настройки, профиль

### Компоненты:
- [ ] WebSocket client manager
- [ ] Resource display widgets
- [ ] Building grid widget
- [ ] Battle grid widget (7x7)
- [ ] Research tree widget
- [ ] Ship card widget
- [ ] Market order list widget
- [ ] Mining maze widget

---

## Фаза 11: Ratings, Events, Polish
**Финальная полировка**

### Что реализовать:
- [ ] Leaderboards (money, food, ships, buildings, total resources)
- [ ] Обновление рейтингов каждые 5 минут
- [ ] Random events:
  - Short circuit (reset energy, cost to fix)
  - Theft (lose 5-20% money)
  - Storage roof collapse (lose 5-20% resources)
  - Mine collapse (lose mine level)
- [ ] Statistics tracking (30+ metrics)
- [ ] Daily stats reset (6:00)
- [ ] NPC market refresh (every 15 sec)
- [ ] UI polish и анимации
- [ ] Error handling
- [ ] Performance optimization

---

## Итого
- **Сделано:** 4/11 фаз
- **Осталось:** 7 фаз
- **Тестов пройдено:** 27+ (Go)

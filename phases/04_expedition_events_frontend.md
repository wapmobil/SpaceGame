> **Общий план:** [surface_expeditions_design.md](../surface_expeditions_design.md)

# Фаза 4: Frontend UI

## Цель
Создать полностью готовый UI для expedition event chains + inventory dialog.

## Что работает после деплоя
Полностью готовый UI. Полная фича готова.

## Задачи

### 9.1 expedition_chain.dart — модели
**Файл:** `client/lib/models/expedition_chain.dart` (NEW)

```dart
class ExpeditionChain {
  final String id;
  final String planetId;
  final String ownerId;
  final String status; // "active" | "completed" | "failed"
  final int eventCount;
  final int currentEventIndex;
  final Map<String, double> inventory; // только 5 ресурсов
  final Location? discoveredLocation;
  final DateTime createdAt;
  final DateTime updatedAt;

  bool get isActive => status == 'active';
  bool get isCompleted => status == 'completed';
  bool get isFailed => status == 'failed';
  double get totalInventory => inventory.values.fold(0, (sum, v) => sum + v);

  factory ExpeditionChain.fromJson(Map<String, dynamic> json);
  Map<String, dynamic> toJson();
}

class ExpeditionEvent {
  final String eventId;
  final String description;
  final Map<String, double> immediateReward; // flat: {"reagents": 30, "food": -10}
  final List<ExpeditionChoice> choices;
  final bool isEnd;
  final String? locationReward;

  factory ExpeditionEvent.fromJson(Map<String, dynamic> json);
  Map<String, dynamic> toJson();
}

class ExpeditionChoice {
  final String label;
  final String description;
  final Map<String, double> reward; // flat: {"iron": 50, "composite": 20}
  final String nextEventId;

  factory ExpeditionChoice.fromJson(Map<String, dynamic> json);
  Map<String, dynamic> toJson();
}

class ExpeditionEventLogEntry {
  final String eventId;
  final String description;
  final int playerChoice; // индекс выбранного варианта
  final String choiceLabel; // label выбранного варианта
  final Map<String, double> rewardsReceived; // flat: {"iron": 50, "food": -10}
  final DateTime createdAt;

  factory ExpeditionEventLogEntry.fromJson(Map<String, dynamic> json);
  Map<String, dynamic> toJson();
}

class ExpeditionChoiceResult {
  final ExpeditionEvent? event;
  final ExpeditionChain chain;
  final Map<String, double> inventory;
  final bool completed;
  final bool failed;
  final Location? location;
  final String? locationReward;
  final String? error;

  factory ExpeditionChoiceResult.fromJson(Map<String, dynamic> json);
}
```

### 9.2 expedition_provider.dart — state management
**Файл:** `client/lib/providers/expedition_provider.dart` (NEW)

```dart
class ExpeditionProvider extends ChangeNotifier {
  String? _authToken;
  String? baseUrl;

  List<ExpeditionChain> _chains = [];
  List<ExpeditionChain> _activeChains = [];
  List<ExpeditionChain> _completedChains = [];
  String? _selectedChainId; // выбранная активная цепочка для отображения
  ExpeditionEvent? _currentEvent;

  // Getters
  List<ExpeditionChain> get chains => _chains;
  List<ExpeditionChain> get activeChains => _activeChains;
  List<ExpeditionChain> get completedChains => _completedChains;
  ExpeditionChain? get selectedChain => _selectedChainId != null
      ? _activeChains.where((c) => c.id == _selectedChainId).firstOrNull
      : null;
  ExpeditionEvent? get currentEvent => _currentEvent;

  // API methods
  Future<void> loadExpeditionChains(String planetId);
  Future<ExpeditionEvent?> getExpeditionEvent(String planetId, String chainId);
  Future<ExpeditionChoiceResult> makeChoice(String planetId, String chainId, int choiceIndex);
  Future<List<ExpeditionEventLogEntry>> getExpeditionEventLog(String planetId, String chainId);
  // GET /api/planets/{id}/expeditions/{chainID}/event-log
  void selectChain(String chainId);

  // Inventory helpers
  double getInventoryTotal(Map<String, double> inventory);
  int getRemainingCapacity(double currentTotal); // 1000 - currentTotal

  // WebSocket handlers
  void onExpeditionEvent(Map<String, dynamic> data);
  void onExpeditionComplete(Map<String, dynamic> data);

  // Helpers
  void setAuthToken(String token);
  Map<String, String> _authHeaders();
}
```

### 9.3 inventory_dialog.dart — диалог старта
**Файл:** `client/lib/widgets/inventory_dialog.dart` (NEW)

Диалог при нажатии "Новая экспедиция":

Компоненты:
- Заголовок "Подготовка экспедиции"
- Описание: "Распределите ресурсы между экспедицией (макс. 1000 суммарно)"
- Для каждого ресурса (food, iron, composite, mechanisms, reagents):
  - Slider с range 0..min(available, 1000 - current_total)
  - Текущее значение (целое число)
  - Доступно: {planet_resources[res]}
- Total: {sum of all} / 1000 (красный если > 1000)
- Кнопки:
  - "Отмена" — закрывает диалог
  - "Отправить" — создаёт экспедицию

Логика слайдеров:
- Каждый слайдер ограничен: min(доступно на планете, 1000 - сумма остальных слайдеров)
- Если сумма слайдеров > 1000 — показать предупреждение (красный текст)
- Кнопка "Отправить" неактивна если total == 0 или total > 1000

### 9.4 expedition_events_screen.dart — основной экран
**Файл:** `client/lib/screens/expedition_events_screen.dart` (NEW)

Экран с тремя секциями:

#### Секция 1: Активные экспедиции
- Если несколько активных цепочек — показать список карточек для переключения
- Карточка активной экспедиции:
  - Номер события: "Событие {index+1}/{eventCount}"
  - Описание события (текст, 2-4 абзаца)
  - Мгновенная награда (если есть): чипы "+X 🍍, -Y 🪨"
  - Инвентарь экспедиции: горизонтальный список ресурсов с иконками
  - Кнопки выбора (2-4 варианта):
    - Label выбора
    - Description последствий
    - Награда за выбор (чипы)
  - При нажатии выбора → спиннер "Обработка выбора..." (LLM 5-300 сек)
- Если is_end:
  - Показать обнаруженную локацию
  - Показать возвращённый инвентарь
  - Кнопка "Закрыть"

#### Секция 2: История событий текущей экспедиции
- Список карточек завершённых событий
- Каждая карточка:
  - Текст события (описание)
  - Выбранный вариант (label)
  - Чипы расходов/доходов ресурсов (+X 🍍, -Y 🪨)
- Карточки отображаются в хронологическом порядке (сверху вниз)

#### Секция 3: Кнопка запуска
- "Новая экспедиция" → открывает InventoryDialog
- После отправки → навигация к активному событию

### 9.5 Обновить planet_survey.dart — модели
**Файл:** `client/lib/models/planet_survey.dart`

Удалить:
- `SurfaceExpedition` class
- `ExpeditionHistoryEntry` class
- `RangeStats` class
- `PlanetSurveyState` class (заменить на ExpeditionChainsState)

Оставить:
- `Location` class
- `LocationBuilding` class
- `LocationsResponse` class

### 9.6 Обновить planet_survey_provider.dart
**Файл:** `client/lib/providers/planet_survey_provider.dart`

Удалить:
- `_surfaceExpeditions` field
- `surfaceExpeditions` getter
- `rangeStats` field (если используется только для surface expeditions)
- `startPlanetSurvey()` method
- `loadPlanetSurveyData()` method
- `onPlanetSurveyUpdate()` method
- `applyStateUpdate()` — убрать surface_expeditions, range_stats

Оставить:
- `_locations` — locations остаются
- `buildOnLocation()`
- `removeBuilding()`
- `abandonLocation()`
- `onLocationUpdate()`
- `applyBuildDetails()`

### 9.7 Обновить planet_survey_screen.dart
**Файл:** `client/lib/screens/planet_survey_screen.dart`

Заменить:
- `_buildStartExpeditionWithSurvey()` → кнопка "Новая экспедиция" → `Navigator.push` к `ExpeditionEventsScreen`
- `_buildExpeditionListWithSurvey()` → убрать (теперь в ExpeditionEventsScreen)
- `_buildRangeStatsWithSurvey()` → убрать

Оставить:
- `_buildLocationsListWithSurvey()` — locations остаются
- `_buildHistoryWithSurvey()` → заменить на историю expedition chains

### 9.8 Обновить game_provider.dart
**Файл:** `client/lib/providers/game_provider.dart`

Добавить:
- `ExpeditionProvider` как поле
- Инициализация в конструкторе

Добавить WebSocket handlers:
```dart
void _handleExpeditionEvent(Map<String, dynamic> data) {
  final chainId = data['chain_id'] as String;
  final event = ExpeditionEvent.fromJson(data['event'] as Map<String, dynamic>);
  final inventory = Map<String, double>.from(data['inventory'] as Map);
  
  expeditionProvider.onExpeditionEvent({
    'chain_id': chainId,
    'event': event,
    'inventory': inventory,
  });
  
  // Показать уведомление
  showNotification('Новое событие экспедиции', event.description.substring(0, 100) + '...');
}

void _handleExpeditionComplete(Map<String, dynamic> data) {
  final chainId = data['chain_id'] as String;
  final status = data['status'] as String;
  final inventory = Map<String, double>.from(data['inventory'] as Map);
  
  expeditionProvider.onExpeditionComplete({
    'chain_id': chainId,
    'status': status,
    'inventory': inventory,
  });
  
  if (status == 'completed') {
    showNotification('Экспедиция завершена!', 'Обнаружена локация: ${data['location']['name']}');
  } else {
    showNotification('Экспедиция провалена', data['error'] as String? ?? 'Ошибка LLM');
  }
}
```

В `_handleWsMessage()`:
```dart
case 'expedition_event':
  _handleExpeditionEvent(data);
  break;
case 'expedition_complete':
  _handleExpeditionComplete(data);
  break;
```

### 9.9 constants.dart — обновление
**Файл:** `client/lib/utils/constants.dart`

Убедиться что:
- `planet_exploration` tech определён
- Нет упоминаний surface expedition duration (300s/600s/1200s)
- Определены 5 ресурсов инвентаря: food, iron, composite, mechanisms, reagents
- Определены иконки для ресурсов инвентаря

### 9.10 Тесты
**Запуск:**
```bash
flutter test
```

Проверить что все существующие тесты проходят.

## Деплой
```bash
./deploy.sh
flutter test
```

## Зависимости
- Фаза 8 (API) должна быть реализована
- WebSocket broadcast методы должны быть добавлены

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
  final Map<String, double> inventory;
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
  final Map<String, double> immediateReward;
  final List<ExpeditionChoice> choices;
  final bool isEnd;
  final String? locationReward;

  factory ExpeditionEvent.fromJson(Map<String, dynamic> json);
  Map<String, dynamic> toJson();
}

class ExpeditionChoice {
  final String label;
  final String description;
  final Map<String, double> reward;
  final String nextEventId;

  factory ExpeditionChoice.fromJson(Map<String, dynamic> json);
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
  ExpeditionChain? _activeChain;
  ExpeditionEvent? _currentEvent;

  // Getters
  List<ExpeditionChain> get chains => _chains;
  ExpeditionChain? get activeChain => _activeChain;
  ExpeditionEvent? get currentEvent => _currentEvent;

  // API methods
  Future<void> loadExpeditionChains(String planetId);
  Future<ExpeditionEvent?> getExpeditionEvent(String planetId, String chainId);
  Future<ExpeditionChoiceResult> makeChoice(String planetId, String chainId, int choiceIndex);
  Future<List<ExpeditionEvent>> getExpeditionEvents(String planetId, String chainId);
  
  // Inventory helper
  Map<String, double> getMaxInventory(Map<String, double> planetResources);
  // Возвращает inventory где sum <= 1000 и каждый ресурс <= planetResources[res]

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
  - Slider с range 0..min(available, 1000)
  - Текущее значение
  - Доступно: {planet_resources[res]}
- Total: {sum of all} / 1000 (красный если > 1000)
- Кнопки:
  - "Заполнить максимум" — заполняет все ресурсы пропорционально до 1000
  - "Очистить" — сбрасывает всё
  - "Отмена" — закрывает диалог
  - "Отправить" — создаёт экспедицию

Логика "Заполнить максимум":
```dart
void fillMax() {
  final totalAvailable = resources.values.fold(0.0, (s, v) => s + v);
  if (totalAvailable <= 1000) {
    // Просто берём всё
    inventory = Map.from(resources);
  } else {
    // Пропорциональное заполнение до 1000
    final ratio = 1000.0 / totalAvailable;
    inventory = resources.map((k, v) => MapEntry(k, v * ratio));
  }
}
```

### 9.4 expedition_events_screen.dart — основной экран
**Файл:** `client/lib/screens/expedition_events_screen.dart` (NEW)

Экран с тремя секциями:

#### Секция 1: Активная экспедиция
- Показывает текущее событие если есть активная цепочка
- Карточка события:
  - Номер: "Событие {index+1}/{eventCount}"
  - Описание события (текст)
  - Мгновенная награда (если есть): "Немедленно: +X food, +Y iron"
  - Инвентарь экспедиции: список ресурсов с иконками
  - Кнопки выбора (2-4 варианта):
    - Label выбора
    - Description последствий
    - Награда за выбор
- Если is_end:
  - Показать обнаруженную локацию
  - Показать возвращённый инвентарь
  - Кнопка "Закрыть"

#### Секция 2: История экспедиций
- Список завершённых/проваленных цепочек
- Каждая: статус, дата, результат (успех/провал), обнаруженная локация

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
  
  expeditionProvider.onExpeditionComplete({
    'chain_id': chainId,
    'status': status,
    'inventory': Map<String, double>.from(data['inventory'] as Map),
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

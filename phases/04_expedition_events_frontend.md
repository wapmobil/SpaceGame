# Фаза 4: Frontend UI

## Цель
Создать полностью готовый UI для expedition event chains + inventory dialog.

## Что работает после деплоя
Полностью готовый UI. Полная фича готова.

## Задачи

### 4.1 expedition_chain.dart — модели

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

class ExpeditionEventLogEntry {
  final String eventId;
  final String description;
  final int playerChoice;
  final String choiceLabel;
  final Map<String, double> rewardsReceived;
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

### 4.2 expedition_provider.dart — state management

**Файл:** `client/lib/providers/expedition_provider.dart` (NEW)

Примечание: существующий `expedition_provider.dart` — для space expeditions. Новый файл переименовать существующий или создать как `expedition_chain_provider.dart`.

```dart
class ExpeditionChainProvider extends ChangeNotifier {
  String? _authToken;
  String? baseUrl;

  List<ExpeditionChain> _chains = [];
  String? _selectedChainId;
  ExpeditionEvent? _currentEvent;
  bool _isLoading = false;

  List<ExpeditionChain> get chains => _chains;
  List<ExpeditionChain> get activeChains =>
      _chains.where((c) => c.isActive).toList();
  List<ExpeditionChain> get completedChains =>
      _chains.where((c) => c.isCompleted || c.isFailed).toList();
  ExpeditionChain? get selectedChain => _selectedChainId != null
      ? _chains.where((c) => c.id == _selectedChainId).firstOrNull
      : null;
  ExpeditionEvent? get currentEvent => _currentEvent;
  bool get isLoading => _isLoading;

  // API methods
  Future<void> loadExpeditionChains(String planetId);
  Future<ExpeditionEvent?> getExpeditionEvent(String planetId, String chainId);
  Future<ExpeditionChoiceResult> makeChoice(String planetId, String chainId, int choiceIndex);
  Future<List<ExpeditionEventLogEntry>> getExpeditionEventLog(String planetId, String chainId);
  Future<ExpeditionChoiceResult> startExpedition(String planetId, Map<String, double> inventory);
  void selectChain(String chainId);

  // Inventory helpers
  double getInventoryTotal(Map<String, double> inventory);
  int getRemainingCapacity(double currentTotal);

  // WebSocket handlers
  void onExpeditionEvent(Map<String, dynamic> data);
  void onExpeditionComplete(Map<String, dynamic> data);

  // Helpers
  void setAuthToken(String token);
  Map<String, String> _authHeaders();
}
```

### 4.3 inventory_dialog.dart — диалог старта

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

### 4.4 expedition_events_screen.dart — основной экран

**Файл:** `client/lib/screens/expedition_events_screen.dart` (NEW)

Экран с тремя секциями:

#### Секция 1: Активные экспедиции
- Если несколько активных цепочек — показать список карточек для переключения
- Карточка активной экспедиции:
  - Номер события: "Событие {index+1}/{eventCount}"
  - Описание события (текст, 2-4 абзаца)
  - Мгновенная награда (если есть): чипы "+X, -Y" с иконками ресурсов
  - Инвентарь экспедиции: горизонтальный список ресурсов с иконками
  - Кнопки выбора (2-4 варианта):
    - Label выбора
    - Description последствий
    - Награда за выбор (чипы)
  - При нажатии выбора → спиннер "Обработка выбора..." (LLM может занять время)
- Если is_end:
  - Показать обнаруженную локацию
  - Показать возвращённый инвентарь
  - Кнопка "Закрыть"

#### Секция 2: История событий текущей экспедиции
- Список карточек завершённых событий
- Каждая карточка:
  - Текст события (описание)
  - Выбранный вариант (label)
  - Чипы расходов/доходов ресурсов
- Карточки отображаются в хронологическом порядке (сверху вниз)

#### Секция 3: Кнопка запуска
- "Новая экспедиция" → открывает InventoryDialog
- После отправки → навигация к активному событию

### 4.5 Обновить planet_survey.dart — модели

**Файл:** `client/lib/models/planet_survey.dart`

Удалить:
- `SurfaceExpedition` class
- `ExpeditionHistoryEntry` class
- `RangeStats` class
- `PlanetSurveyState` class

Оставить:
- `Location` class
- `LocationBuilding` class
- `LocationsResponse` class

### 4.6 Обновить planet_survey_provider.dart

**Файл:** `client/lib/providers/planet_survey_provider.dart`

Удалить:
- `_surfaceExpeditions` field
- `surfaceExpeditions` getter
- `rangeStats` field
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

### 4.7 Обновить planet_survey_screen.dart

**Файл:** `client/lib/screens/planet_survey_screen.dart`

Заменить:
- `_buildStartExpeditionWithSurvey()` → кнопка "Новая экспедиция" → `Navigator.push` к `ExpeditionEventsScreen`
- `_buildExpeditionListWithSurvey()` → убрать (теперь в ExpeditionEventsScreen)
- `_buildRangeStatsWithSurvey()` → убрать

Оставить:
- `_buildLocationsListWithSurvey()` — locations остаются
- `_buildHistoryWithSurvey()` → заменить на историю expedition chains

### 4.8 Обновить game_provider.dart

**Файл:** `client/lib/providers/game_provider.dart`

Добавить:
- `ExpeditionChainProvider` как поле
- Инициализация в конструкторе

Добавить WebSocket handlers:
```dart
void _handleExpeditionEvent(Map<String, dynamic> data) {
  final chainId = data['chain_id'] as String;
  final event = ExpeditionEvent.fromJson(data['event'] as Map<String, dynamic>);
  final inventory = Map<String, double>.from(data['inventory'] as Map);

  expeditionChainProvider.onExpeditionEvent({
    'chain_id': chainId,
    'event': event,
    'inventory': inventory,
  });

  showNotification('Новое событие экспедиции',
      event.description.substring(0, min(100, event.description.length)) + '...');
}

void _handleExpeditionComplete(Map<String, dynamic> data) {
  final chainId = data['chain_id'] as String;
  final status = data['status'] as String;
  final inventory = Map<String, double>.from(data['inventory'] as Map);

  expeditionChainProvider.onExpeditionComplete({
    'chain_id': chainId,
    'status': status,
    'inventory': inventory,
  });

  if (status == 'completed') {
    final locName = data['location'] != null ? data['location']['name'] : 'неизвестна';
    showNotification('Экспедиция завершена!', 'Обнаружена локация: $locName');
  } else {
    showNotification('Экспедиция провалена', data['error'] as String? ?? 'Ошибка');
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

### 4.9 constants.dart — обновление

**Файл:** `client/lib/utils/constants.dart`

Убедиться что:
- `planet_exploration` tech определён
- Нет упоминаний surface expedition duration (300s/600s/1200s)
- Определены 5 ресурсов инвентаря: food, iron, composite, mechanisms, reagents
- Определены иконки для ресурсов инвентаря

### 4.10 Тесты

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

- Фаза 3 (API) должна быть реализована
- WebSocket broadcast методы должны быть добавлены

## Примечания

- Новый provider назван `ExpeditionChainProvider` чтобы не конфликтовать со существующим `ExpeditionProvider` для space expeditions.
- HTTP запросы к `makeChoice` и `startExpedition` могут занять до 330s — нужен загрузочный индикатор и обработка timeout.

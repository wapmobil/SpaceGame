> **Общий план:** [surface_expeditions_design.md](../surface_expeditions_design.md)

# Фаза 5: Frontend UI

## Цель
Создать полностью готовый UI для surface expeditions и locations.

## Что работает после деплоя
Полностью готовый UI. Полная фича готова.

## Задачи

### 5.1 planet_survey.dart — models
**Файл:** `client/lib/models/planet_survey.dart` (NEW)

```dart
class SurfaceExpedition {
  final String id;
  final String planetId;
  final String status; // "active", "discovered", "completed", "failed"
  final double progress;
  final double duration;
  final double elapsedTime;
  final String range; // "300s", "600s", "1200s"
  final DateTime createdAt;
  final DateTime updatedAt;
  
  // From factory
  bool get isActive => status == 'active';
  bool get isComplete => status == 'completed';
  bool get isFailed => status == 'failed';
  bool get isDiscovered => status == 'discovered';
  double get remainingTime => duration - elapsedTime;
  
  factory SurfaceExpedition.fromJson(Map<String, dynamic> json);
  Map<String, dynamic> toJson();
}

class Location {
  final String id;
  final String type;
  final String name;
  final String? buildingType;
  final int buildingLevel;
  final bool buildingActive;
  final String sourceResource;
  final double sourceAmount;
  final double sourceRemaining;
  final bool active;
  final DateTime discoveredAt;
  
  factory Location.fromJson(Map<String, dynamic> json);
  Map<String, dynamic> toJson();
}

class LocationBuilding {
  final String id;
  final String locationId;
  final String buildingType;
  final int level;
  final bool active;
  final double buildProgress;
  final double buildTime;
  
  factory LocationBuilding.fromJson(Map<String, dynamic> json);
  Map<String, dynamic> toJson();
}

class ExpeditionHistoryEntry {
  final String id;
  final String status;
  final String result; // "success", "failed", "abandoned"
  final String discovered;
  final Map<String, double> resourcesGained;
  final DateTime createdAt;
  final DateTime completedAt;
  
  factory ExpeditionHistoryEntry.fromJson(Map<String, dynamic> json);
  Map<String, dynamic> toJson();
}

class RangeStats {
  final int totalExpeditions;
  final int locationsFound;
  
  factory RangeStats.fromJson(Map<String, dynamic> json);
}

class PlanetSurveyState {
  final List<SurfaceExpedition> expeditions;
  final Map<String, RangeStats> rangeStats;
  final int maxDuration;
  final Map<String, double> costPerMin;
  
  factory PlanetSurveyState.fromJson(Map<String, dynamic> json);
}

class LocationsResponse {
  final List<Location> locations;
  
  factory LocationsResponse.fromJson(Map<String, dynamic> json);
}
```

### 5.2 planet_survey_provider.dart — state management
**Файл:** `client/lib/providers/planet_survey_provider.dart` (NEW)

```dart
class PlanetSurveyProvider extends ChangeNotifier {
  List<SurfaceExpedition> expeditions = [];
  List<Location> locations = [];
  Map<String, RangeStats> rangeStats = {};
  ExpeditionHistoryEntry? selectedHistory;
  int? surfaceExpeditionCooldown;
  bool canStartPlanetSurvey = false;
  bool canStartSpaceExpedition = false;
  
  Future<void> loadPlanetSurvey(String planetId);
  Future<void> loadLocations(String planetId);
  Future<void> startPlanetSurvey(String planetId, int duration);
  Future<void> buildOnLocation(String planetId, String locationId, String buildingType);
  Future<void> removeBuilding(String planetId, String locationId);
  Future<void> abandonLocation(String planetId, String locationId);
  Future<void> loadExpeditionHistory(String planetId);
  
  void handlePlanetSurveyUpdate(Map<String, dynamic> data);
  void handleLocationUpdate(Map<String, dynamic> data);
  
  List<LocationBuildingDef> getAvailableBuildingsForLocation(String locationType);
  int getMaxDurationForBaseLevel(int baseLevel);
  Map<String, double> getCostPerMinForBaseLevel(int baseLevel);
}
```

### 5.3 planet_survey_screen.dart — main screen
**Файл:** `client/lib/screens/planet_survey_screen.dart` (NEW)

Экран с двумя вкладками:
- **Planet Survey** — активные экспедиции, кнопка запуска, история, список локаций
- **Space Expedition** — космические экспедиции (переиспользовать существующий ExpeditionScreen)

Компоненты:
- `_buildStartExpedition()` — выбор duration (300/600/1200s), cost display
- `_buildExpeditionList()` — список активных экспедиций с прогресс-барами
- `_buildLocationList()` — список локаций с building info
- `_buildHistory()` — история экспедиций
- `_buildLocationCard()` — карточка локации

### 5.4 location_card.dart — location card widget
**Файл:** `client/lib/widgets/location_card.dart` (NEW)

```dart
class LocationCard extends StatelessWidget {
  final Location location;
  final VoidCallback? onBuild;
  final VoidCallback? onRemove;
  final VoidCallback? onAbandon;
  
  // Display:
  // - Location name + type (rarity color)
  // - Progress bar (source_remaining / source_amount)
  // - Building info (if built)
  // - Action buttons (build/remove/abandon)
}
```

### 5.5 planet.dart — update model
**Файл:** `client/lib/models/planet.dart`

Добавить поля:
```dart
final String? resourceType;
final bool canStartPlanetSurvey;
final bool canStartSpaceExpedition;
final int? baseLevel;
final int? commandCenterLevel;
final int maxLocations;
```

### 5.6 planet_screen.dart — nav chips
**Файл:** `client/lib/screens/planet_screen.dart`

Добавить navigation chips:
- "Разведка" — появляется когда `baseLevel > 0`
- "Космические экспедиции" — появляется когда `commandCenterLevel > 0`

### 5.7 game_provider.dart — update
**Файл:** `client/lib/providers/game_provider.dart`

Добавить методы:
```dart
Future<void> startPlanetSurvey(String planetId, int duration);
Future<void> buildOnLocation(String planetId, String locationId, String buildingType);
Future<void> removeBuilding(String planetId, String locationId);
Future<void> abandonLocation(String planetId, String locationId);
```

WS handlers:
```dart
void _handlePlanetSurveyUpdate(Map<String, dynamic> data);
void _handleLocationUpdate(Map<String, dynamic> data);
```

### 5.8 constants.dart — update
**Файл:** `client/lib/utils/constants.dart`

Добавить:
- Новые techs: `location_buildings`, `advanced_exploration`
- Surface expedition duration options: [300, 600, 1200]
- Base level → duration/cost mapping

### 5.9 Тесты
**Запуск:**
```bash
flutter test
```

Все фреймворк тесты должны проходить.

## Деплой
```bash
./deploy.sh
flutter test
```

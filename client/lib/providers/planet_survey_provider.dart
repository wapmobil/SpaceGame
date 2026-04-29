import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import '../models/planet_survey.dart';
import '../services/api_service.dart';

class LocationBuildingDef {
  final String type;
  final String name;
  final double costFood;
  final double costIron;
  final double costMoney;
  final double buildTime;

  LocationBuildingDef({
    required this.type,
    required this.name,
    required this.costFood,
    required this.costIron,
    required this.costMoney,
    required this.buildTime,
  });
}

class PlanetSurveyProvider extends ChangeNotifier {
  final String baseUrl;
  final String? authToken;
  final ApiService _apiService;

  List<SurfaceExpedition> expeditions = [];
  List<Location> locations = [];
  Map<String, RangeStats> rangeStats = {};
  ExpeditionHistoryEntry? selectedHistory;
  int? surfaceExpeditionCooldown;
  bool canStartPlanetSurvey = false;
  bool canStartSpaceExpedition = false;
  int? baseLevel;
  int? commandCenterLevel;
  int? maxLocations;
  PlanetSurveyState? surveyState;
  List<ExpeditionHistoryEntry> history = [];
  Map<String, List<LocationBuildingDef>> availableBuildingsMap = {};

  PlanetSurveyProvider({required this.baseUrl, this.authToken})
      : _apiService = ApiService(baseUrl: baseUrl, authToken: authToken);

  void setAuthToken(String token) {
    _apiService.authToken = token;
  }

  Future<void> loadPlanetSurvey(String planetId) async {
    try {
      final data = await _apiService.getPlanetSurvey(planetId);
      surveyState = PlanetSurveyState.fromJson(data);
      expeditions = surveyState!.expeditions;
      rangeStats = surveyState!.rangeStats;
      canStartPlanetSurvey = data['can_start_planet_survey'] as bool? ?? false;
      canStartSpaceExpedition = data['can_start_space_expedition'] as bool? ?? false;
      baseLevel = data['base_level'] as int?;
      commandCenterLevel = data['command_center_level'] as int?;
      maxLocations = data['max_locations'] as int? ?? 1;
      surfaceExpeditionCooldown = data['surface_expedition_cooldown'] as int?;
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to load planet survey: $e');
    }
  }

  Future<void> loadLocations(String planetId) async {
    try {
      final data = await _apiService.getLocations(planetId);
      locations = data.map((e) => Location.fromJson(e)).toList();
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to load locations: $e');
    }
  }

  Future<void> startPlanetSurvey(String planetId, int duration) async {
    try {
      final data = await _apiService.startPlanetSurvey(planetId, duration);
      if (data['status'] == 'started') {
        surveyState = PlanetSurveyState.fromJson(data);
        expeditions = surveyState!.expeditions;
        rangeStats = surveyState!.rangeStats;
        canStartPlanetSurvey = data['can_start_planet_survey'] as bool? ?? false;
        surfaceExpeditionCooldown = data['surface_expedition_cooldown'] as int?;
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to start planet survey: $e');
    }
  }

  Future<void> buildOnLocation(String planetId, String locationId, String buildingType) async {
    try {
      await _apiService.buildOnLocation(planetId, locationId, buildingType);
      loadLocations(planetId);
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to build on location: $e');
    }
  }

  Future<void> removeBuilding(String planetId, String locationId) async {
    try {
      await _apiService.removeBuilding(planetId, locationId);
      loadLocations(planetId);
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to remove building: $e');
    }
  }

  Future<void> abandonLocation(String planetId, String locationId) async {
    try {
      await _apiService.abandonLocation(planetId, locationId);
      loadLocations(planetId);
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to abandon location: $e');
    }
  }

  Future<void> loadExpeditionHistory(String planetId) async {
    try {
      final data = await _apiService.getExpeditionHistory(planetId);
      history = data.map((e) => ExpeditionHistoryEntry.fromJson(e)).toList();
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to load expedition history: $e');
    }
  }

  void handlePlanetSurveyUpdate(Map<String, dynamic> data) {
    if (data['expedition_id'] == null) return;

    final expId = data['expedition_id'] as String;
    final idx = expeditions.indexWhere((e) => e.id == expId);

    if (idx >= 0) {
      final old = expeditions[idx];
      final updated = SurfaceExpedition(
        id: old.id,
        planetId: old.planetId,
        status: data['status'] as String? ?? old.status,
        progress: (data['progress'] as num?)?.toDouble() ?? old.progress,
        duration: old.duration,
        elapsedTime: (data['elapsed_time'] as num?)?.toDouble() ?? old.elapsedTime,
        range: old.range,
        createdAt: old.createdAt,
        updatedAt: DateTime.now().toUtc(),
      );
      expeditions[idx] = updated;
    } else {
      final newExp = SurfaceExpedition(
        id: expId,
        planetId: data['planet_id'] as String? ?? '',
        status: data['status'] as String? ?? 'active',
        progress: (data['progress'] as num?)?.toDouble() ?? 0,
        duration: (data['duration'] as num?)?.toDouble() ?? 0,
        elapsedTime: (data['elapsed_time'] as num?)?.toDouble() ?? 0,
        range: data['range'] as String? ?? '300s',
        createdAt: DateTime.now().toUtc(),
        updatedAt: DateTime.now().toUtc(),
      );
      expeditions.add(newExp);
    }

    if (data['status'] == 'completed' || data['status'] == 'failed' || data['status'] == 'discovered') {
      loadPlanetSurvey(_getPlanetIdFromExpedition(expId));
      loadExpeditionHistory(_getPlanetIdFromExpedition(expId));
    }

    notifyListeners();
  }

  void handleLocationUpdate(Map<String, dynamic> data) {
    if (data['location_id'] == null) return;

    final locId = data['location_id'] as String;
    final idx = locations.indexWhere((l) => l.id == locId);

    if (idx >= 0) {
      final old = locations[idx];
      final updated = Location(
        id: old.id,
        type: old.type,
        name: old.name,
        buildingType: data['building_type'] as String? ?? old.buildingType,
        buildingLevel: data['building_level'] as int? ?? old.buildingLevel,
        buildingActive: data['building_active'] as bool? ?? old.buildingActive,
        sourceResource: old.sourceResource,
        sourceAmount: old.sourceAmount,
        sourceRemaining: (data['source_remaining'] as num?)?.toDouble() ?? old.sourceRemaining,
        active: data['active'] as bool? ?? old.active,
        discoveredAt: old.discoveredAt,
      );
      locations[idx] = updated;
    }

    notifyListeners();
  }

  String _getPlanetIdFromExpedition(String expId) {
    final exp = expeditions.firstWhere((e) => e.id == expId, orElse: () => expeditions.first);
    return exp.planetId;
  }

  List<LocationBuildingDef> getAvailableBuildingsForLocation(String locationType) {
    if (availableBuildingsMap.containsKey(locationType)) {
      return availableBuildingsMap[locationType]!;
    }

    final buildings = <LocationBuildingDef>[];
    final rarity = _getRarityForLocationType(locationType);

    double getCostMultiplier() {
      switch (rarity) {
        case 'common':
          return 1.0;
        case 'uncommon':
          return 2.0;
        case 'rare':
          return 4.0;
        case 'exotic':
          return 6.0;
        default:
          return 1.0;
      }
    }

    final costMult = getCostMultiplier();

    switch (locationType) {
      case 'pond':
        buildings.addAll([
          LocationBuildingDef(type: 'fish_farm', name: 'Рыбная ферма', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
          LocationBuildingDef(type: 'water_purifier', name: 'Очиститель воды', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        ]);
        break;
      case 'river':
        buildings.addAll([
          LocationBuildingDef(type: 'fish_farm', name: 'Рыбная ферма', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
          LocationBuildingDef(type: 'irrigation_system', name: 'Система орошения', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        ]);
        break;
      case 'forest':
        buildings.addAll([
          LocationBuildingDef(type: 'lumber_mill', name: 'Лесопилка', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
          LocationBuildingDef(type: 'herb_garden', name: 'Травяной сад', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        ]);
        break;
      case 'mineral_deposit':
        buildings.addAll([
          LocationBuildingDef(type: 'mineral_extractor', name: 'Экстрактор минералов', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
          LocationBuildingDef(type: 'smelter', name: 'Плавильня', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        ]);
        break;
      case 'dry_valley':
        buildings.addAll([
          LocationBuildingDef(type: 'solar_farm', name: 'Солнечная ферма', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
          LocationBuildingDef(type: 'wind_turbine', name: 'Ветровая турбина', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        ]);
        break;
      case 'waterfall':
        buildings.addAll([
          LocationBuildingDef(type: 'hydro_plant', name: 'Гидроэлектростанция', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
          LocationBuildingDef(type: 'turbine_station', name: 'Турбинная станция', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
        ]);
        break;
      case 'cave':
        buildings.addAll([
          LocationBuildingDef(type: 'crystal_mine', name: 'Кристальная шахта', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
          LocationBuildingDef(type: 'cave_lab', name: 'Лаборатория пещеры', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
        ]);
        break;
      case 'thermal_spring':
        buildings.addAll([
          LocationBuildingDef(type: 'geothermal_plant', name: 'Геотермальная станция', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
          LocationBuildingDef(type: 'hot_spring_lab', name: 'Лаборатория источников', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
        ]);
        break;
      case 'salt_lake':
        buildings.addAll([
          LocationBuildingDef(type: 'salt_pans', name: 'Соляные копи', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
          LocationBuildingDef(type: 'chemical_plant', name: 'Химический завод', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
        ]);
        break;
      case 'wind_pass':
        buildings.addAll([
          LocationBuildingDef(type: 'wind_farm', name: 'Ветряная ферма', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
          LocationBuildingDef(type: 'storm_collector', name: 'Сборщик штормов', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
        ]);
        break;
      case 'crystal_cave':
        buildings.addAll([
          LocationBuildingDef(type: 'crystal_harvester', name: 'Сборщик кристаллов', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
          LocationBuildingDef(type: 'crystal_lab', name: 'Кристальная лаборатория', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
        ]);
        break;
      case 'meteor_crater':
        buildings.addAll([
          LocationBuildingDef(type: 'meteor_science_lab', name: 'Метеоритная лаборатория', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
          LocationBuildingDef(type: 'alloy_forge', name: 'Кузница сплавов', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
        ]);
        break;
      case 'sunken_city':
        buildings.addAll([
          LocationBuildingDef(type: 'salvage_station', name: 'Станция спасения', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
          LocationBuildingDef(type: 'ruins_archive', name: 'Архив руин', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
        ]);
        break;
      case 'glacier':
        buildings.addAll([
          LocationBuildingDef(type: 'ice_mine', name: 'Ледниковая шахта', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
          LocationBuildingDef(type: 'cryo_lab', name: 'Криолаборатория', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
        ]);
        break;
      case 'mushroom_forest':
        buildings.addAll([
          LocationBuildingDef(type: 'mushroom_farm', name: 'Грибная ферма', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
          LocationBuildingDef(type: 'spore_extractor', name: 'Экстрактор спор', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
        ]);
        break;
      case 'crystal_field':
        buildings.addAll([
          LocationBuildingDef(type: 'crystal_array', name: 'Кристальный массив', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
          LocationBuildingDef(type: 'resonance_amplifier', name: 'Резонансный усилитель', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
        ]);
        break;
      case 'cloud_island':
        buildings.addAll([
          LocationBuildingDef(type: 'cloud_harvester', name: 'Сборщик облаков', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
          LocationBuildingDef(type: 'aerial_platform', name: 'Воздушная платформа', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
        ]);
        break;
      case 'underground_lake':
        buildings.addAll([
          LocationBuildingDef(type: 'aquaculture_base', name: 'Аквакультурная база', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
          LocationBuildingDef(type: 'underground_irrigation', name: 'Подземное орошение', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
        ]);
        break;
      case 'radioactive_zone':
        buildings.addAll([
          LocationBuildingDef(type: 'radiation_filter', name: 'Фильтр радиации', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
          LocationBuildingDef(type: 'isotope_plant', name: 'Изотопный завод', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
        ]);
        break;
      case 'anomaly_zone':
        buildings.addAll([
          LocationBuildingDef(type: 'anomaly_siphon', name: 'Сифон аномалий', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
          LocationBuildingDef(type: 'containment_unit', name: 'Удержание', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
        ]);
        break;
      default:
        buildings.add(
          LocationBuildingDef(type: 'generic_extractor', name: 'Экстрактор', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        );
    }

    availableBuildingsMap[locationType] = buildings;
    return buildings;
  }

  String _getRarityForLocationType(String locationType) {
    final common = ['pond', 'river', 'forest', 'mineral_deposit', 'dry_valley'];
    final uncommon = ['waterfall', 'cave', 'thermal_spring', 'salt_lake', 'wind_pass'];
    final rare = ['crystal_cave', 'meteor_crater', 'sunken_city', 'glacier', 'mushroom_forest'];
    final exotic = ['crystal_field', 'cloud_island', 'underground_lake', 'radioactive_zone', 'anomaly_zone'];

    if (common.contains(locationType)) return 'common';
    if (uncommon.contains(locationType)) return 'uncommon';
    if (rare.contains(locationType)) return 'rare';
    if (exotic.contains(locationType)) return 'exotic';
    return 'common';
  }

  int getMaxDurationForBaseLevel(int baseLevel) {
    switch (baseLevel) {
      case 1:
        return 300;
      case 2:
        return 600;
      case 3:
        return 1200;
      default:
        return 300;
    }
  }

  Map<String, double> getCostPerMinForBaseLevel(int baseLevel) {
    switch (baseLevel) {
      case 1:
        return {'food': 100, 'iron': 100, 'money': 10};
      case 2:
        return {'food': 200, 'iron': 200, 'money': 20};
      case 3:
        return {'food': 400, 'iron': 400, 'money': 40};
      default:
        return {'food': 100, 'iron': 100, 'money': 10};
    }
  }

  double getExpeditionCost(int duration, int baseLevel) {
    final costs = getCostPerMinForBaseLevel(baseLevel);
    final minutes = duration / 60;
    return costs['food']! * minutes;
  }

  String getRarityLabel(String locationType) {
    final rarity = _getRarityForLocationType(locationType);
    switch (rarity) {
      case 'common':
        return 'Обычная';
      case 'uncommon':
        return 'Необычная';
      case 'rare':
        return 'Редкая';
      case 'exotic':
        return 'Экзотическая';
      default:
        return 'Неизвестная';
    }
  }

  Color getRarityColor(String locationType) {
    final rarity = _getRarityForLocationType(locationType);
    switch (rarity) {
      case 'common':
        return const Color(0xFF9e9e9e);
      case 'uncommon':
        return const Color(0xFF4caf50);
      case 'rare':
        return const Color(0xFF2196f3);
      case 'exotic':
        return const Color(0xFF9c27b0);
      default:
        return Colors.white54;
    }
  }

  String getBuildingName(String buildingType) {
    for (final buildings in availableBuildingsMap.values) {
      for (final b in buildings) {
        if (b.type == buildingType) return b.name;
      }
    }
    return buildingType;
  }

  @override
  void dispose() {
    super.dispose();
  }
}

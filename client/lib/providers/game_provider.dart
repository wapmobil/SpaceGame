import 'dart:convert';
import 'package:flutter/foundation.dart';
import '../core/platform_service.dart';
import '../core/platform_service_web.dart' if (dart.library.io) '../core/platform_service_io.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../core/websocket_manager.dart';
import '../models/planet.dart';
import '../models/building.dart';
import '../models/ship.dart';
import '../models/research.dart';
import '../models/expedition.dart';
import '../models/market.dart';
import '../models/drill.dart';
import '../models/rating.dart';
import '../models/player.dart';
import '../providers/garden_bed_provider.dart';
import '../providers/planet_survey_provider.dart';

class GameProvider extends ChangeNotifier {
  final WebSocketManager websocket;
  final GardenBedProvider _gardenBedProvider;
  final PlanetSurveyProvider _planetSurveyProvider;
  String _baseUrl;
  Player? _player;
  List<Planet> _planets = [];
  Planet? _selectedPlanet;
  List<Building> _buildings = [];
  List<Ship> _ships = [];
  ShipyardInfo? _shipyardInfo;
  List<ShipType> _availableShipTypes = [];
  ResearchState? _researchState;
   ExpeditionsListResponse? _expeditions;
  MarketData? _marketData;
  List<MarketOrder> _myOrders = [];
  DrillState? _drillState;
  String? _drillSessionId;
  int? _drillSeed;
  String? _drillPendingDirection;
  bool _drillPendingExtracting = false;
  List<RatingEntry> _ratings = [];
  Map<String, dynamic>? _stats;
  List<Map<String, dynamic>> _events = [];
  String? _errorMessage;
  int _activeConstructions = 0;
  int _maxConstructions = 1;
  Map<String, Map<String, dynamic>> _buildingCosts = {};

  // Energy balance
  double _energyBalanceProduction = 0;
  double _energyBalanceConsumption = 0;
  double _energyBalanceNet = 0;

  // Production totals
  double _productionFood = 0;
  double _productionIron = 0;
  double _productionComposite = 0;
  double _productionMechanisms = 0;
  double _productionReagents = 0;
  double _productionEnergy = 0;
  double _productionMoney = 0;
  double _productionAlienTech = 0;

  // Storage
  double _storageCapacity = 0;

  // Base operational flags
  bool _baseOperational = true;
  bool _canResearch = true;
  bool _canExpedition = false;
  
  // Research unlocks (planet_exploration random unlock)
  String _researchUnlocks = '';
  
  // Research paused status
  bool _researchPaused = false;

  // Completed research tech IDs with levels (from WS state updates)
  Map<String, int> _completedResearch = {};

  GameProvider({required this.websocket, String? baseUrl, GardenBedProvider? gardenBedProvider})
      : _baseUrl = baseUrl ?? _getBaseUri(),
        _gardenBedProvider = gardenBedProvider ?? GardenBedProvider(websocket: websocket, baseUrl: baseUrl ?? _getBaseUri()),
        _planetSurveyProvider = PlanetSurveyProvider(baseUrl: baseUrl ?? _getBaseUri(), authToken: null);

  static String _getBaseUri() {
    final service = createPlatformService();
    final host = service.host;
    final scheme = service.scheme;
    if (host.isNotEmpty) {
      return '$scheme://$host';
    }
    return 'http://localhost:8080';
  }

  String get baseUrl => _baseUrl;
  String getBuildDetailsUrl(String planetId) => '$_baseUrl/api/planets/$planetId/build-details';
  Player? get player => _player;
  List<Planet> get planets => _planets;
  Planet? get selectedPlanet => _selectedPlanet;
  List<Building> get buildings => _buildings;
  List<Ship> get ships => _ships;
  ShipyardInfo? get shipyardInfo => _shipyardInfo;
  List<ShipType> get availableShipTypes => _availableShipTypes;
  ResearchState? get researchState => _researchState;
  GardenBedProvider get gardenBedProvider => _gardenBedProvider;
  PlanetSurveyProvider get planetSurveyProvider => _planetSurveyProvider;
    ExpeditionsListResponse? get expeditions => _expeditions;
  MarketData? get marketData => _marketData;
  List<MarketOrder> get myOrders => _myOrders;
  DrillState? get drillState => _drillState;
  String? get drillPendingDirection => _drillPendingDirection;
  bool get drillPendingExtracting => _drillPendingExtracting;

  void clearDrillState() {
    _drillState = null;
    notifyListeners();
  }
  List<RatingEntry> get ratings => _ratings;
  Map<String, dynamic>? get stats => _stats;
  List<Map<String, dynamic>> get events => _events;
  String? get errorMessage => _errorMessage;
  int get activeConstructions => _activeConstructions;
  int get maxConstructions => _maxConstructions;
  Map<String, Map<String, dynamic>> get buildingCosts => _buildingCosts;
  Map<String, int> get completedResearch => _completedResearch;
  bool get isLoggedIn => _player != null;

  // Energy getters (from planet resources)
  double get energyValue => _selectedPlanet?.resources['energy'] as double? ?? 0;
  double get energyMax => _selectedPlanet?.resources['max_energy'] as double? ?? 100;
  bool get energyDeficit => (_selectedPlanet?.resources['energy'] as double? ?? 0) <= 0;

  // Energy balance getters
  double get energyBalanceProduction => _energyBalanceProduction;
  double get energyBalanceConsumption => _energyBalanceConsumption;
  double get energyBalanceNet => _energyBalanceNet;

  // Production getters
  double get productionFood => _productionFood;
  double get productionIron => _productionIron;
  double get productionComposite => _productionComposite;
  double get productionMechanisms => _productionMechanisms;
  double get productionReagents => _productionReagents;
  double get productionEnergy => _productionEnergy;
  double get productionMoney => _productionMoney;
  double get productionAlienTech => _productionAlienTech;

  // Storage getter
  double get storageCapacity => _storageCapacity;

  // Base operational getters
  bool get baseOperational => _baseOperational;
  bool get canResearch => _canResearch;
  bool get canExpedition => _canExpedition;
    String get researchUnlocks => _researchUnlocks;
  bool get researchPaused => _researchPaused;
  String? get authToken => _player?.authToken;

  int getBuildingLevelForPlanet(String planetId, String buildingType) {
    final buildings = planetId == _selectedPlanet?.id ? _buildings : (planets.firstWhere((p) => p.id == planetId, orElse: () => _selectedPlanet!).buildings ?? []);
    for (final b in buildings) {
      if (b.type == buildingType) {
        return b.level;
      }
    }
    return 0;
  }

  void _setError(String msg) {
    _errorMessage = msg;
    notifyListeners();
  }

  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }

  Future<void> login(String name) async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/login'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'name': name}),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        _player = Player.fromJson(data);
        _gardenBedProvider.setAuthToken(_player!.authToken);
        await _savePlayer();
        notifyListeners();
        connectWebSocket();
      } else {
        _setError('Не удалось войти: ${response.statusCode}');
      }
    } catch (e) {
      _setError('Ошибка входа: $e');
    }
  }

  Future<void> logout() async {
    websocket.disconnect();
    _player = null;
    _planets = [];
    _selectedPlanet = null;
    _buildings = [];
    _ships = [];
    _researchState = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.clear();
    notifyListeners();
  }

  Future<void> _savePlayer() async {
    if (_player == null) return;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('player_id', _player!.id);
    await prefs.setString('auth_token', _player!.authToken);
    await prefs.setString('player_name', _player!.name);
    await prefs.setString('base_url', _baseUrl);
  }

  Future<void> loadSavedPlayer() async {
    final prefs = await SharedPreferences.getInstance();
    final id = prefs.getString('player_id');
    final token = prefs.getString('auth_token');
    final name = prefs.getString('player_name');
    final url = prefs.getString('base_url') ?? _baseUrl;

    if (id != null && token != null) {
      _baseUrl = url;
      _player = Player(id: id, authToken: token, name: name ?? '');
      _gardenBedProvider.setAuthToken(token);
      notifyListeners();
      await connectWebSocket();
      await loadPlanets();
    }
  }

  Future<void> connectWebSocket() async {
    if (_player != null) {
      websocket.connect(_baseUrl, _player!.authToken);
      websocket.addMessageListener(_onWebSocketMessage);
    }
  }

  void _onWebSocketMessage(Map<String, dynamic> message) {
    final type = message['type'] as String?;
    final data = message['data'] as Map<String, dynamic>?;

    switch (type) {
      case 'state_update':
        _handleStateUpdate(data);
        break;
      case 'building_update':
        _handleBuildingUpdate(data);
        break;
      case 'research_update':
        _handleResearchUpdate(data);
        break;
       case 'space_expedition_update':
        _handleSpaceExpeditionUpdate(data);
        break;
      case 'market_update':
        _handleMarketUpdate(data);
        break;
            case 'notification':
        _handleNotification(data);
        break;
      case 'drill_update':
        onDrillUpdate(data ?? {});
        break;
       case 'garden_bed_update':
        _gardenBedProvider.onGardenBedUpdate(data ?? {});
        notifyListeners();
        break;
      case 'planet_survey_update':
        _handlePlanetSurveyUpdate(data);
        break;
      case 'location_update':
        _handleLocationUpdate(data);
        break;
      default:
        if (message['type'] == 'garden_bed_action_result') {
          _gardenBedProvider.onGardenBedWSActionResult(data ?? {});
          notifyListeners();
        }
    }
  }

 void _handleStateUpdate(Map<String, dynamic>? data) {
    if (data != null && _selectedPlanet != null) {
      // Backend sends {"planet_id": "...", "state": {...}}
      final stateData = data['state'] as Map<String, dynamic>? ?? data;
      final planetId = stateData['id'] as String?;
      if (planetId != _selectedPlanet!.id) return;

      // Update resources
      final resources = stateData['resources'] as Map<String, dynamic>?;
      if (resources != null) {
        _selectedPlanet = _selectedPlanet!.copyWith(resources: resources);
      }

      // Update energy net from energy_balance
      final energyBalance = stateData['energy_balance'];
      if (energyBalance != null) {
        _productionEnergy = (energyBalance as num?)?.toDouble() ?? 0;
      }

      // Update storage capacity
      final storageCap = stateData['storage_capacity'];
      if (storageCap != null) {
        _storageCapacity = storageCap is double ? storageCap : (storageCap as num).toDouble();
      }

      // Update buildings from slice
      final buildingsJson = stateData['buildings'] as List<dynamic>?;
      if (buildingsJson != null) {
        _buildings = buildingsJson.map((b) => Building.fromJson(b as Map<String, dynamic>)).toList();
      }

      // Update construction limits
      final activeConstr = stateData['active_constructions'];
      if (activeConstr != null) {
        _activeConstructions = activeConstr is int ? activeConstr : (activeConstr as num).toInt();
      }
      final maxConstr = stateData['max_constructions'];
      if (maxConstr != null) {
        _maxConstructions = maxConstr is int ? maxConstr : (maxConstr as num).toInt();
      }

      // Update research paused status
      final resPaused = stateData['research_paused'];
      if (resPaused != null) {
        _researchPaused = resPaused is bool ? resPaused : (resPaused as num).toInt() != 0;
      }

      // Update research progress from state
      final researchList = stateData['research'] as List<dynamic>?;
      final availableResearchList = stateData['available_research'] as List<dynamic>?;
      if (researchList != null && _researchState != null) {
        final updatedResearch = List<ResearchTech>.from(_researchState!.research);
        for (final r in researchList) {
          final techId = r['tech_id'] as String;
          final idx = updatedResearch.indexWhere((t) => t.techId == techId);
          if (idx >= 0) {
            updatedResearch[idx] = updatedResearch[idx].copyWith(
              completed: r['completed'] as bool? ?? updatedResearch[idx].completed,
              inProgress: r['in_progress'] as bool? ?? updatedResearch[idx].inProgress,
              progress: (r['progress'] as num?)?.toDouble() ?? updatedResearch[idx].progress,
              totalTime: (r['total_time'] as num?)?.toDouble() ?? updatedResearch[idx].totalTime,
              progressPct: (r['progress_pct'] as num?)?.toDouble() ?? updatedResearch[idx].progressPct,
            );
          }
        }
        var updatedAvailable = _researchState!.available;
        if (availableResearchList != null && availableResearchList.isNotEmpty) {
          updatedAvailable = availableResearchList.map((a) => ResearchTech.fromMap(a as Map<String, dynamic>)).toList();
        }
        _researchState = ResearchState(
          research: updatedResearch,
          available: updatedAvailable,
        );
      }

      // Update completed research map (tech_id -> level)
      final completedResearch = stateData['completed_research'] as Map<String, dynamic>?;
      if (completedResearch != null) {
        _completedResearch = {};
        completedResearch.forEach((key, value) {
          _completedResearch[key] = (value as num).toInt();
        });
      }

      // Update garden bed state
      final farmStateJson = stateData['garden_bed_state'] as Map<String, dynamic>?;
      if (farmStateJson != null && farmStateJson.isNotEmpty) {
        _gardenBedProvider.onGardenBedUpdate(farmStateJson);
      }

      notifyListeners();
    }
  }

  void _handleBuildingUpdate(Map<String, dynamic>? data) {
    if (data != null && _selectedPlanet != null) {
      loadBuildDetails(_selectedPlanet!.id);
    }
  }

  void _handleResearchUpdate(Map<String, dynamic>? data) {
    if (data != null) {
      loadResearch(_selectedPlanet?.id ?? '');
      notifyListeners();
    }
  }

    void _handleSpaceExpeditionUpdate(Map<String, dynamic>? data) {
    if (data != null && _selectedPlanet != null) {
      loadExpeditions(_selectedPlanet!.id);
      notifyListeners();
    }
  }

  void _handleMarketUpdate(Map<String, dynamic>? data) {
    if (data != null) {
      loadMarketData(_selectedPlanet?.id ?? '');
      loadMyOrders(_selectedPlanet?.id ?? '');
      notifyListeners();
    }
  }

    void _handleNotification(Map<String, dynamic>? data) {
    // Notifications are handled by the UI layer
    debugPrint('Notification: ${data?['message']}');
  }

  void _handlePlanetSurveyUpdate(Map<String, dynamic>? data) {
    if (data != null && _selectedPlanet != null) {
      _planetSurveyProvider.handlePlanetSurveyUpdate(data);
      notifyListeners();
    }
  }

  void _handleLocationUpdate(Map<String, dynamic>? data) {
    if (data != null && _selectedPlanet != null) {
      _planetSurveyProvider.handleLocationUpdate(data);
      notifyListeners();
    }
  }

  Future<void> loadPlanets() async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        _planets = (data as List?)?.map((e) => Planet.fromJson(e as Map<String, dynamic>)).toList() ?? [];
        notifyListeners();
      }
    } catch (e) {
       _setError('Не удалось загрузить планеты: $e');
    }
  }

  Future<void> createPlanet(String name) async {
    if (_player == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({'name': name}),
      );

      if (response.statusCode == 201) {
        await loadPlanets();
      } else {
        _setError('Не удалось создать планету: ${response.statusCode}');
      }
    } catch (e) {
      _setError('Ошибка создания планеты: $e');
    }
  }

  void selectPlanet(Planet planet) {
    _selectedPlanet = planet;
    websocket.subscribe(planet.id);
    loadBuildDetails(planet.id);
    loadShips(planet.id);
    loadResearch(planet.id);
     loadExpeditions(planet.id);
    loadMarketData(planet.id);
    loadMyOrders(planet.id);
    _gardenBedProvider.getGardenBed(planet.id);
    _planetSurveyProvider.loadPlanetSurvey(planet.id);
    _planetSurveyProvider.loadLocations(planet.id);
    _planetSurveyProvider.loadExpeditionHistory(planet.id);
    _planetSurveyProvider.setAuthToken(_player!.authToken);
        notifyListeners();
  }

  Future<void> loadPlanetDetail(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final planet = Planet.fromJson(data);
        _planets = _planets.map((p) => p.id == planet.id ? planet : p).toList();
        if (_selectedPlanet?.id == planet.id) {
          _selectedPlanet = planet;
        }
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load planet detail: $e');
    }
  }

  Future<void> loadBuildDetails(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse(getBuildDetailsUrl(planetId)),
        headers: _authHeaders(),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        _updateFromBuildDetails(data);
      }
    } catch (e) {
      debugPrint('Error loading build details: $e');
    }
  }

  Map<String, String> _authHeaders() {
    return {'X-Auth-Token': _player!.authToken};
  }

  void _updateFromBuildDetails(Map<String, dynamic> data) {
    if (_selectedPlanet == null) return;

    // Update energy balance
    final energyBalance = data['energy_balance'] as Map<String, dynamic>?;
    if (energyBalance != null) {
      _energyBalanceProduction = (energyBalance['production'] as num?)?.toDouble() ?? 0;
      _energyBalanceConsumption = (energyBalance['consumption'] as num?)?.toDouble() ?? 0;
      _energyBalanceNet = (energyBalance['net'] as num?)?.toDouble() ?? 0;
    }

    // Update storage capacity
    final resources = data['resources'] as Map<String, dynamic>?;
    if (resources != null) {
      _storageCapacity = (resources['storage_capacity'] as num?)?.toDouble() ?? 0;
    }

    // Update buildings
    final buildingsJson = data['buildings'] as List<dynamic>?;
    if (buildingsJson != null) {
      _buildings = buildingsJson.map((b) => Building.fromJson(b as Map<String, dynamic>)).toList();
    }

    // Update production
    final production = data['production'] as Map<String, dynamic>?;
    if (production != null) {
      _productionFood = (production['food'] as num?)?.toDouble() ?? 0;
      _productionIron = (production['iron'] as num?)?.toDouble() ?? 0;
      _productionComposite = (production['composite'] as num?)?.toDouble() ?? 0;
      _productionMechanisms = (production['mechanisms'] as num?)?.toDouble() ?? 0;
      _productionReagents = (production['reagents'] as num?)?.toDouble() ?? 0;
      _productionEnergy = (production['energy_net'] as num?)?.toDouble() ?? 0;
      _productionMoney = (production['money'] as num?)?.toDouble() ?? 0;
      _productionAlienTech = (production['alien_tech'] as num?)?.toDouble() ?? 0;
    }

    // Update construction limits
    _activeConstructions = (data['active_constructions'] as num?)?.toInt() ?? 0;
    _maxConstructions = (data['max_constructions'] as num?)?.toInt() ?? 1;

    // Update building costs for unbuilt buildings
    final costsJson = data['building_costs'] as Map<String, dynamic>?;
    if (costsJson != null) {
      _buildingCosts = {};
      costsJson.forEach((key, value) {
        final v = value as Map<String, dynamic>;
        _buildingCosts[key] = {
          'food': (v['cost']?['food'] as num?)?.toDouble() ?? 0,
          'iron': (v['cost']?['iron'] as num?)?.toDouble() ?? 0,
          'money': (v['cost']?['money'] as num?)?.toDouble() ?? 0,
          'production': v['production'] as Map<String, dynamic>? ?? {},
          'deltas': v['deltas'] as Map<String, dynamic>? ?? {},
          'next_production': v['next_production'] as Map<String, dynamic>? ?? {},
        };
      });
    }

    // Update base operational flags
    _baseOperational = data['base_operational'] as bool? ?? true;
    _canResearch = data['can_research'] as bool? ?? true;
    _canExpedition = data['can_expedition'] as bool? ?? false;
    
    // Update research unlocks
    _researchUnlocks = data['research_unlocks'] as String? ?? '';

    // Update farm state
    final farmStateJson = data['garden_bed_state'] as Map<String, dynamic>?;
    if (farmStateJson != null && farmStateJson.isNotEmpty) {
      _gardenBedProvider.onGardenBedUpdate(farmStateJson);
    }

    // Update planet survey fields
    _planetSurveyProvider.canStartPlanetSurvey = data['can_start_planet_survey'] as bool? ?? false;
    _planetSurveyProvider.canStartSpaceExpedition = data['can_start_space_expedition'] as bool? ?? false;
    _planetSurveyProvider.baseLevel = data['base_level'] as int?;
    _planetSurveyProvider.commandCenterLevel = data['command_center_level'] as int?;
    _planetSurveyProvider.maxLocations = data['max_locations'] as int? ?? 1;

    notifyListeners();
  }

  Future<void> buildStructure(String buildingType) async {
    if (_selectedPlanet == null) return;

    _errorMessage = null;

    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/buildings'),
        headers: _authHeaders(),
        body: jsonEncode({'type': buildingType}),
      );

      if (response.statusCode == 201) {
        await loadBuildDetails(_selectedPlanet!.id);
      } else if (response.statusCode == 400) {
        final errorData = jsonDecode(response.body) as Map<String, dynamic>;
        _errorMessage = errorData['error'] as String? ?? 'Не удалось построить';
        final errorMsg = _errorMessage ?? '';
        if (errorMsg.contains('max_constructions')) {
          _errorMessage = 'Достигнут лимит строительства. Исследуйте "Параллельное строительство", чтобы открыть больше.';
        } else if (errorMsg.contains('prerequisite_missing')) {
          _errorMessage = errorData['extra'] as String? ?? 'Не выполнены требования.';
        }
      } else {
        _errorMessage = 'Не удалось построить, статус ${response.statusCode}';
      }
    } catch (e) {
      _errorMessage = 'Ошибка сети: $e';
    }

    notifyListeners();
  }

  BuildingUpgradeInfo getBuildingUpgradeInfo(Building building) {
    if (_selectedPlanet == null) return BuildingUpgradeInfo(canUpgrade: false, reason: 'Планета не выбрана');

    final isReady = building.isBuildComplete;
    if (isReady) return BuildingUpgradeInfo(canUpgrade: false, reason: 'Нажмите чтобы открыть');

    final isBuilding = building.isBuilding;
    if (isBuilding) return BuildingUpgradeInfo(canUpgrade: false, reason: 'Уже строится');

    final nextCostFood = building.nextCostFood;
    final nextCostIron = building.nextCostIron;
    final nextCostMoney = building.nextCostMoney;
    if (nextCostFood <= 0 && nextCostIron <= 0 && nextCostMoney <= 0) return BuildingUpgradeInfo(canUpgrade: false, reason: 'Максимальный уровень');

    final currentFood = (_selectedPlanet!.resources['food'] ?? 0) as num;
    final currentIron = (_selectedPlanet!.resources['iron'] ?? 0) as num;
    final currentMoney = (_selectedPlanet!.resources['money'] ?? 0) as num;
    final canAffordFood = currentFood.toDouble() >= nextCostFood;
    final canAffordIron = currentIron.toDouble() >= nextCostIron;
    final canAffordMoney = currentMoney.toDouble() >= nextCostMoney;
    if (!canAffordFood || !canAffordIron || !canAffordMoney) {
      final missing = <String>[];
      if (!canAffordFood) missing.add('еда');
      if (!canAffordIron) missing.add('железо');
      if (!canAffordMoney) missing.add('деньги');
      return BuildingUpgradeInfo(canUpgrade: false, reason: 'Не хватает ${missing.join(" и ")}');
    }

    if (activeConstructions >= maxConstructions) return BuildingUpgradeInfo(canUpgrade: false, reason: 'Достигнут лимит строительства');

    return BuildingUpgradeInfo(canUpgrade: true, reason: null);
  }

  Future<void> confirmBuilding(String buildingType) async {
    if (_selectedPlanet == null) return;

    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/buildings/$buildingType/confirm'),
        headers: _authHeaders(),
      );

      if (response.statusCode == 200) {
        await loadBuildDetails(_selectedPlanet!.id);
      }
    } catch (e) {
      debugPrint('Error confirming building: $e');
    }

    notifyListeners();
  }

  Future<void> toggleBuilding(String buildingType) async {
    if (_selectedPlanet == null) return;

    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/buildings/$buildingType/toggle'),
        headers: _authHeaders(),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final enabled = data['enabled'] as bool;
        final idx = _buildings.indexWhere((b) => b.type == buildingType);
        if (idx >= 0) {
          final old = _buildings[idx];
          _buildings[idx] = Building(
            type: old.type,
            level: old.level,
            buildProgress: old.buildProgress,
            enabled: enabled,
            buildTime: old.buildTime,
            costFood: old.costFood,
            costIron: old.costIron,
            costMoney: old.costMoney,
            nextCostFood: old.nextCostFood,
            nextCostIron: old.nextCostIron,
            nextCostMoney: old.nextCostMoney,
            productionFood: old.productionFood,
            productionIron: old.productionIron,
            productionComposite: old.productionComposite,
            productionMechanisms: old.productionMechanisms,
            productionReagents: old.productionReagents,
            productionEnergy: old.productionEnergy,
            productionMoney: old.productionMoney,
            productionAlienTech: old.productionAlienTech,
            consumption: old.consumption,
          );
        }
      }
    } catch (e) {
      debugPrint('Error toggling building: $e');
    }

    notifyListeners();
  }

  Future<void> loadShips(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/fleet'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        _shipyardInfo = ShipyardInfo.fromJson(data);

        final shipsMap = data['ships'] as Map<String, dynamic>? ?? {};
        _ships = shipsMap.entries
            .where((e) => e.value != null)
            .map((e) => Ship.fromJson({
                  'id': e.key,
                  'planet_id': planetId,
                  'type': e.value['type'] as String? ?? 'unknown',
                  'hp': e.value['hp'] as int? ?? 100,
                  'armor': e.value['armor'] as int? ?? 0,
                  'weapons': e.value['weapons'] as List? ?? [],
                  'cargo': (e.value['cargo'] as num?)?.toDouble() ?? 0,
                  'energy': (e.value['energy'] as num?)?.toDouble() ?? 0,
                }))
            .toList();
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load ships: $e');
    }
  }

  Future<void> loadAvailableShipTypes(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/ships/available'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final shipTypesData = data['ship_types'] as List? ?? [];
        _availableShipTypes = shipTypesData
            .map((e) => ShipType.fromJson(e as Map<String, dynamic>))
            .toList();
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load ship types: $e');
    }
  }

  Future<void> buildShip(String shipType) async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/ship/build'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({'ship_type': shipType}),
      );

      if (response.statusCode == 201) {
        await loadShips(_selectedPlanet!.id);
        await loadAvailableShipTypes(_selectedPlanet!.id);
      } else {
        _setError('Не удалось построить корабль: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка постройки корабля: $e');
    }
  }

  Future<void> loadResearch(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/research'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        _researchState = ResearchState.fromJson(data);
        _researchPaused = data['research_paused'] as bool? ?? false;
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load research: $e');
    }
  }

  Future<void> startResearch(String techId) async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/research/start'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({'tech_id': techId}),
      );

      if (response.statusCode == 201) {
        await loadResearch(_selectedPlanet!.id);
      } else {
        _setError('Не удалось начать исследование: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка исследования: $e');
    }
  }

    Future<void> loadExpeditions(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/expeditions'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        _expeditions = ExpeditionsListResponse.fromJson(
            jsonDecode(response.body) as Map<String, dynamic>);
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load expeditions: $e');
    }
  }

  Future<void> startExpedition({
    required String expeditionType,
    String? target,
    List<String>? shipTypes,
    List<int>? shipCounts,
    double? duration,
  }) async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      final body = <String, dynamic>{
        'expedition_type': expeditionType,
        'duration': duration ?? 3600,
      };
      if (target != null) body['target'] = target;
      if (shipTypes != null) body['ship_types'] = shipTypes;
      if (shipCounts != null) body['ship_counts'] = shipCounts;

      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/expeditions'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode(body),
      );

      if (response.statusCode == 201) {
        await loadExpeditions(_selectedPlanet!.id);
      } else {
        _setError('Не удалось начать экспедицию: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка экспедиции: $e');
    }
  }

  Future<void> expeditionAction(String expeditionId, String action) async {
    if (_player == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/expeditions/$expeditionId/action'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({'action': action}),
      );

      if (response.statusCode == 200) {
        await loadExpeditions(_selectedPlanet?.id ?? '');
      } else {
         _setError('Не удалось выполнить действие экспедиции: ${response.body}');
       }
    } catch (e) {
      _setError('Ошибка действия экспедиции: $e');
    }
  }

  Future<void> loadMarketData(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/market'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        _marketData = MarketData.fromJson(
            jsonDecode(response.body) as Map<String, dynamic>);
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load market: $e');
    }
  }

  Future<void> loadMyOrders(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/market/orders'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as List;
        _myOrders = data.map((e) => MarketOrder.fromJson(e as Map<String, dynamic>)).toList();
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load my orders: $e');
    }
  }

  Future<void> createMarketOrder({
    required String resource,
    required String orderType,
    required double amount,
    required double price,
    bool isPrivate = false,
  }) async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/market/orders'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({
          'resource': resource,
          'order_type': orderType,
          'amount': amount,
          'price': price,
          'is_private': isPrivate,
        }),
      );

      if (response.statusCode == 201) {
        await loadMarketData(_selectedPlanet!.id);
        await loadMyOrders(_selectedPlanet!.id);
      } else {
        _setError('Не удалось создать ордер: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка создания ордера: $e');
    }
  }

  Future<void> deleteMarketOrder(String orderId) async {
    if (_player == null) return;
    try {
      final response = await http.delete(
        Uri.parse('$_baseUrl/api/market/orders/$orderId'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        await loadMarketData(_selectedPlanet?.id ?? '');
        await loadMyOrders(_selectedPlanet?.id ?? '');
      } else {
        _setError('Не удалось удалить ордер: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка удаления ордера: $e');
    }
  }

  Future<bool> sellFood(String planetId, double amount) async {
    if (_player == null) return false;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/$planetId/sell-food'),
        headers: {'X-Auth-Token': _player!.authToken},
        body: jsonEncode({'amount': amount}),
      );

      if (response.statusCode == 200) {
        await loadPlanetDetail(planetId);
        return true;
      } else {
        _setError('Не удалось продать еду: ${response.body}');
        return false;
      }
    } catch (e) {
      _setError('Ошибка продажи еды: $e');
      return false;
    }
  }

  Future<bool> sellIron(String planetId, double amount) async {
    if (_player == null) return false;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/$planetId/sell-iron'),
        headers: {'X-Auth-Token': _player!.authToken},
        body: jsonEncode({'amount': amount}),
      );

      if (response.statusCode == 200) {
        await loadPlanetDetail(planetId);
        return true;
      } else {
        _setError('Не удалось продать железо: ${response.body}');
        return false;
      }
    } catch (e) {
      _setError('Ошибка продажи железа: $e');
      return false;
    }
  }



  Future<void> loadRatings({String category = 'score'}) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/ratings?category=$category'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final rating = RatingResponse.fromJson(data);
        _ratings = rating.entries;
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load ratings: $e');
    }
  }

  Future<void> loadStats(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/stats'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        _stats = jsonDecode(response.body) as Map<String, dynamic>;
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load stats: $e');
    }
  }

  Future<void> loadEvents(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/events'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final eventsData = data['events'] as List? ?? [];
        _events = eventsData.map((e) => e as Map<String, dynamic>).toList();
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load events: $e');
    }
  }

  Future<void> resolveEvent(String eventType) async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/events/resolve'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({'event_type': eventType}),
      );

      if (response.statusCode == 200) {
        await loadEvents(_selectedPlanet!.id);
      } else {
        _setError('Не удалось решить событие: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка решения события: $e');
    }
  }

  @override
  void dispose() {
    websocket.removeMessageListener(_onWebSocketMessage);
    super.dispose();
  }

 Future<void> startDrill() async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/drill/start'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
      );

      if (response.statusCode == 201) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final startResp = DrillStartResponse.fromJson(data);
        _drillSessionId = startResp.sessionId;
        _drillSeed = startResp.seed;
        _drillState = DrillState(
          sessionId: startResp.sessionId,
          planetId: _selectedPlanet!.id,
          drillHp: startResp.drillHp,
          drillMaxHp: startResp.drillMaxHp,
          depth: startResp.depth,
          drillX: startResp.drillX,
          worldWidth: 5,
          world: [],
          resources: [],
          status: startResp.status,
          totalEarned: 0,
          createdAt: DateTime.now().toUtc().toIso8601String(),
          seed: startResp.seed,
        );
        notifyListeners();
      } else {
        _setError('Не удалось начать бурение: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка бурения: $e');
    }
  }

  void drillCommand({String? direction, bool? extract}) {
    if (_drillState == null || _drillState!.status != 'active') return;
    if (direction != null) _drillPendingDirection = direction;
    if (extract != null) _drillPendingExtracting = extract;
    _drillState = DrillState(
      sessionId: _drillState!.sessionId,
      planetId: _drillState!.planetId,
      drillHp: _drillState!.drillHp,
      drillMaxHp: _drillState!.drillMaxHp,
      depth: _drillState!.depth,
      drillX: _drillState!.drillX,
      worldWidth: _drillState!.worldWidth,
      world: _drillState!.world,
      resources: _drillState!.resources,
      status: _drillState!.status,
      totalEarned: _drillState!.totalEarned,
      createdAt: _drillState!.createdAt,
      completedAt: _drillState!.completedAt,
      seed: _drillState!.seed,
      pendingDirection: _drillPendingDirection,
      pendingExtracting: _drillPendingExtracting,
    );
    notifyListeners();
    websocket.sendDrillCommand(direction: direction, extract: extract);
  }

  void onDrillUpdate(Map<String, dynamic> data) {
    if (_player == null) return;
    try {
      final update = DrillUpdate.fromJson(data);
      _drillSessionId = update.sessionId;
      _drillPendingDirection = null;
      _drillState = DrillState(
        sessionId: update.sessionId,
        planetId: _selectedPlanet?.id ?? '',
        drillHp: update.drillHp,
        drillMaxHp: update.drillMaxHp,
        depth: update.depth,
        drillX: update.drillX,
        worldWidth: update.world.isNotEmpty ? update.world[0].length : 5,
        world: update.world,
        resources: update.resources,
        status: update.gameEnded ? 'failed' : update.status,
        totalEarned: update.totalEarned,
        createdAt: _drillState?.createdAt ?? DateTime.now().toUtc().toIso8601String(),
        completedAt: update.gameEnded ? DateTime.now().toUtc().toIso8601String() : _drillState?.completedAt,
        seed: _drillSeed,
        pendingDirection: null,
        pendingExtracting: false,
      );
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to process drill update: $e');
    }
  }

  Future<void> completeDrill() async {
    if (_player == null || _selectedPlanet == null || _drillState == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/drill/complete'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
      );
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        List<List<DrillCell>> world = [];
        if (data['world'] != null) {
          world = (data['world'] as List).map((row) {
            return (row as List).map((cell) => DrillCell.fromJson(cell as Map<String, dynamic>)).toList();
          }).toList();
        }
        List<DrillResource> resources = [];
        if (data['resources'] != null) {
          resources = (data['resources'] as List).map((r) => DrillResource.fromJson(r as Map<String, dynamic>)).toList();
        }
        _drillState = DrillState(
          sessionId: _drillState!.sessionId,
          planetId: _drillState!.planetId,
          drillHp: data['drill_hp'] as int? ?? 0,
          drillMaxHp: data['drill_max_hp'] as int? ?? 0,
          depth: data['depth'] as int? ?? 0,
          drillX: data['drill_x'] as int? ?? 0,
          worldWidth: world.isNotEmpty ? world[0].length : _drillState!.worldWidth,
          world: world,
          resources: resources,
          status: 'completed',
          totalEarned: (data['total_earned'] as num?)?.toInt() ?? 0,
          createdAt: _drillState!.createdAt,
          completedAt: DateTime.now().toUtc().toIso8601String(),
          seed: _drillSeed,
        );
        notifyListeners();
      } else {
        _setError('Не удалось завершить бурение: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка завершения: $e');
    }
  }

  Future<void> destroyDrill() async {
    if (_player == null || _selectedPlanet == null || _drillState == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/drill/destroy'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
      );
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        List<List<DrillCell>> world = [];
        if (data['world'] != null) {
          world = (data['world'] as List).map((row) {
            return (row as List).map((cell) => DrillCell.fromJson(cell as Map<String, dynamic>)).toList();
          }).toList();
        }
        List<DrillResource> resources = [];
        if (data['resources'] != null) {
          resources = (data['resources'] as List).map((r) => DrillResource.fromJson(r as Map<String, dynamic>)).toList();
        }
        _drillState = DrillState(
          sessionId: _drillState!.sessionId,
          planetId: _drillState!.planetId,
          drillHp: data['drill_hp'] as int? ?? 0,
          drillMaxHp: data['drill_max_hp'] as int? ?? 0,
          depth: data['depth'] as int? ?? 0,
          drillX: data['drill_x'] as int? ?? 0,
          worldWidth: world.isNotEmpty ? world[0].length : _drillState!.worldWidth,
          world: world,
          resources: resources,
          status: 'failed',
          totalEarned: (data['total_earned'] as num?)?.toInt() ?? 0,
          createdAt: _drillState!.createdAt,
          completedAt: DateTime.now().toUtc().toIso8601String(),
          seed: _drillSeed,
        );
        notifyListeners();
      }
    } catch (e) {
      _drillState = null;
      notifyListeners();
    }
  }

  Future<void> cleanupDrill() async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/drill/cleanup'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
      );
    } catch (e) {
      // ignore
    }
  }

  Future<void> startPlanetSurvey(String planetId, int duration) async {
    if (_player == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/$planetId/planet-survey'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({'duration': duration}),
      );

      if (response.statusCode == 200) {
        await _planetSurveyProvider.loadPlanetSurvey(planetId);
        notifyListeners();
      } else {
        _setError('Не удалось начать разведку: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка разведки: $e');
    }
  }

  Future<void> buildOnLocation(String planetId, String locationId, String buildingType) async {
    if (_player == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/$planetId/locations/$locationId/build'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({'building_type': buildingType}),
      );

      if (response.statusCode == 200) {
        await _planetSurveyProvider.loadLocations(planetId);
        notifyListeners();
      } else {
        _setError('Не удалось построить: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка строительства: $e');
    }
  }

  Future<void> removeBuilding(String planetId, String locationId) async {
    if (_player == null) return;
    try {
      final response = await http.delete(
        Uri.parse('$_baseUrl/api/planets/$planetId/locations/$locationId/building'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        await _planetSurveyProvider.loadLocations(planetId);
        notifyListeners();
      } else {
        _setError('Не удалось снести: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка: $e');
    }
  }

  Future<void> abandonLocation(String planetId, String locationId) async {
    if (_player == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/$planetId/locations/$locationId/abandon'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
      );

      if (response.statusCode == 200) {
        await _planetSurveyProvider.loadLocations(planetId);
        notifyListeners();
      } else {
        _setError('Не удалось забрать: ${response.body}');
      }
    } catch (e) {
      _setError('Ошибка: $e');
    }
  }
}

class ShipyardInfo {
  final Map<String, dynamic> ships;
  final int totalShips;
  final int totalSlots;
  final int maxSlots;
  final double totalCargo;
  final double totalEnergy;
  final double totalDamage;
  final double totalHP;
  final int shipyardLevel;
  final int shipyardQueueLen;
  final double shipyardProgress;

  ShipyardInfo({
    this.ships = const {},
    this.totalShips = 0,
    this.totalSlots = 0,
    this.maxSlots = 0,
    this.totalCargo = 0,
    this.totalEnergy = 0,
    this.totalDamage = 0,
    this.totalHP = 0,
    this.shipyardLevel = 0,
    this.shipyardQueueLen = 0,
    this.shipyardProgress = 0,
  });

  factory ShipyardInfo.fromJson(Map<String, dynamic> json) {
    return ShipyardInfo(
      ships: json['ships'] as Map<String, dynamic>? ?? {},
      totalShips: json['total_ships'] as int? ?? 0,
      totalSlots: json['total_slots'] as int? ?? 0,
      maxSlots: json['max_slots'] as int? ?? 0,
      totalCargo: (json['total_cargo'] as num?)?.toDouble() ?? 0,
      totalEnergy: (json['total_energy'] as num?)?.toDouble() ?? 0,
      totalDamage: (json['total_damage'] as num?)?.toDouble() ?? 0,
      totalHP: (json['total_hp'] as num?)?.toDouble() ?? 0,
      shipyardLevel: json['shipyard_level'] as int? ?? 0,
      shipyardQueueLen: json['shipyard_queue_len'] as int? ?? 0,
      shipyardProgress: (json['shipyard_progress'] as num?)?.toDouble() ?? 0,
    );
  }

  int get availableSlots => maxSlots - totalSlots;
  double get buildProgressPercent => (shipyardProgress / (shipyardQueueLen > 0 ? 1 : 100)) * 100;
}

class BuildingUpgradeInfo {
  final bool canUpgrade;
  final String? reason;
  const BuildingUpgradeInfo({required this.canUpgrade, this.reason});
}

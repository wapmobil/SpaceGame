import 'dart:convert';
import 'dart:js_interop';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../core/server_config.dart';
import '../core/websocket_manager.dart';
import '../models/planet.dart';
import '../models/building.dart';
import '../models/ship.dart';
import '../models/research.dart';
import '../models/player.dart';
import '../providers/garden_bed_provider.dart';
import '../providers/drill_provider.dart';
import '../providers/planet_survey_provider.dart';
import '../providers/rating_provider.dart';
import '../providers/research_provider.dart';
import '../providers/market_provider.dart';
import '../providers/expedition_provider.dart';
import '../providers/expedition_chain_provider.dart';
import '../providers/shipyard_info.dart';
import '../providers/building_upgrade_info.dart';

@JS('window.location.origin')
external JSString? get _origin;

class GameProvider extends ChangeNotifier {
  final WebSocketManager websocket;
  final GardenBedProvider _gardenBedProvider;
  final DrillProvider _drillProvider;
  final PlanetSurveyProvider _surveyProvider;
  final RatingProvider _ratingProvider;
  final ResearchProvider _researchProvider;
  final MarketProvider _marketProvider;
  final ExpeditionProvider _expeditionProvider;
  final ExpeditionChainProvider _expeditionChainProvider;
  String _baseUrl;
  Player? _player;
  List<Planet> _planets = [];
  Planet? _selectedPlanet;
  List<Building> _buildings = [];
  List<Ship> _ships = [];
  ShipyardInfo? _shipyardInfo;
  List<ShipType> _availableShipTypes = [];
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
  bool _planetSurveyUnlocked = false;
  
  // Research unlocks (planet_exploration random unlock)
  String _researchUnlocks = '';

  final List<Map<String, String>> _notifications = [];

  GameProvider({required this.websocket, String? baseUrl, GardenBedProvider? gardenBedProvider})
       : _baseUrl = baseUrl ?? _getBaseUri(),
         _gardenBedProvider = gardenBedProvider ?? GardenBedProvider(websocket: websocket, baseUrl: baseUrl ?? _getBaseUri()),
         _drillProvider = DrillProvider(),
         _surveyProvider = PlanetSurveyProvider(),
         _ratingProvider = RatingProvider(),
         _researchProvider = ResearchProvider(),
         _marketProvider = MarketProvider(),
         _expeditionProvider = ExpeditionProvider(),
         _expeditionChainProvider = ExpeditionChainProvider() {
    _expeditionChainProvider.addListener(() {
      notifyListeners();
    });
    _initProviders(baseUrl);
  }

  void _initProviders(String? baseUrl) {
    _drillProvider.baseUrl = baseUrl ?? _getBaseUri();
    _drillProvider.websocket = websocket;
    _surveyProvider.baseUrl = baseUrl ?? _getBaseUri();
    _ratingProvider.baseUrl = baseUrl ?? _getBaseUri();
    _researchProvider.baseUrl = baseUrl ?? _getBaseUri();
    _marketProvider.baseUrl = baseUrl ?? _getBaseUri();
    _expeditionProvider.baseUrl = baseUrl ?? _getBaseUri();
    _expeditionChainProvider.baseUrl = baseUrl ?? _getBaseUri();
  }

  static String _getBaseUri() {
    if (kIsWeb) {
      final origin = _origin?.toDart;
      if (origin != null && origin.isNotEmpty) return origin;
    }
    return ServerConfig.baseUri;
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
  ResearchProvider get researchProvider => _researchProvider;
  GardenBedProvider get gardenBedProvider => _gardenBedProvider;
  DrillProvider get drillProvider => _drillProvider;
  PlanetSurveyProvider get surveyProvider => _surveyProvider;
  RatingProvider get ratingProvider => _ratingProvider;
  MarketProvider get marketProvider => _marketProvider;
  ExpeditionProvider get expeditionProvider => _expeditionProvider;
  ExpeditionChainProvider get expeditionChainProvider => _expeditionChainProvider;
  String? get errorMessage => _errorMessage;
  int get activeConstructions => _activeConstructions;
  int get maxConstructions => _maxConstructions;
  Map<String, Map<String, dynamic>> get buildingCosts => _buildingCosts;
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
  bool get planetSurveyUnlocked => _planetSurveyUnlocked;
    String get researchUnlocks => _researchUnlocks;
  ResearchState? get researchState => _researchProvider.researchState;
  bool get researchPaused => _researchProvider.researchPaused;
  Map<String, int> get completedResearch => _researchProvider.completedResearch;
  String? get authToken => _player?.authToken;

  List<Map<String, String>> get notifications => _notifications;

   void addNotification(String title, String message) {
     _notifications.add({'title': title, 'message': message});
     notifyListeners();
   }

   void clearNotifications() {
     _notifications.clear();
     notifyListeners();
   }

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
        _drillProvider.setAuthToken(_player!.authToken);
        _surveyProvider.setAuthToken(_player!.authToken);
        _ratingProvider.setAuthToken(_player!.authToken);
        _researchProvider.setAuthToken(_player!.authToken);
        _marketProvider.setAuthToken(_player!.authToken);
        _expeditionProvider.setAuthToken(_player!.authToken);
        _expeditionChainProvider.setAuthToken(_player!.authToken);
        await _savePlayer();
        notifyListeners();
        connectWebSocket();
        await loadPlanets();
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
        _drillProvider.setAuthToken(token);
        _surveyProvider.setAuthToken(token);
        _ratingProvider.setAuthToken(token);
        _researchProvider.setAuthToken(token);
        _marketProvider.setAuthToken(token);
        _expeditionProvider.setAuthToken(token);
        _expeditionChainProvider.setAuthToken(token);
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
        _researchProvider.onResearchUpdate(data);
        notifyListeners();
        break;
       case 'space_expedition_update':
        _expeditionProvider.onExpeditionUpdate();
        notifyListeners();
        break;
      case 'market_update':
        _marketProvider.onMarketUpdate();
        notifyListeners();
        break;
            case 'notification':
        _ratingProvider.onNotification(data);
        notifyListeners();
        break;
      case 'drill_update':
        _drillProvider.onDrillUpdate(data ?? {});
        notifyListeners();
        break;
       case 'garden_bed_update':
        _gardenBedProvider.onGardenBedUpdate(data ?? {});
        notifyListeners();
        break;
      case 'planet_survey_update':
        _handlePlanetSurveyUpdate(data);
        break;
      case 'expedition_event':
        _handleExpeditionEvent(data);
        notifyListeners();
        break;
      case 'expedition_complete':
        _handleExpeditionComplete(data);
        notifyListeners();
        break;
      case 'location_update':
        _surveyProvider.onLocationUpdate(data);
        notifyListeners();
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

      // Update research state from WS
      _researchProvider.applyResearchStateUpdate(stateData);

      // Update garden bed state
      final farmStateJson = stateData['garden_bed_state'] as Map<String, dynamic>?;
      if (farmStateJson != null && farmStateJson.isNotEmpty) {
        _gardenBedProvider.onGardenBedUpdate(farmStateJson);
      }

      // Update planet survey data from state
      _surveyProvider.applyStateUpdate(stateData);

      // Update expedition data
      _expeditionProvider.applyStateUpdate(stateData);

      _planetSurveyUnlocked = _researchProvider.completedResearch['planet_exploration'] != null;
      _baseOperational = stateData['base_operational'] as bool? ?? true;

      notifyListeners();
    }
  }

  void _handleBuildingUpdate(Map<String, dynamic>? data) {
    if (data != null && _selectedPlanet != null) {
      loadBuildDetails(_selectedPlanet!.id);
    }
  }

  void _handlePlanetSurveyUpdate(Map<String, dynamic>? data) {
    if (data != null) {
      _expeditionChainProvider.onExpeditionEvent(data);
    }
  }

  void _handleExpeditionEvent(Map<String, dynamic>? data) {
    if (data != null) {
      _expeditionChainProvider.onExpeditionEvent(data);
      final event = _expeditionChainProvider.currentEvent;
      if (event != null && event.description.isNotEmpty) {
        addNotification('Событие экспедиции', event.description.length > 100 ? event.description.substring(0, 100) + '...' : event.description);
      }
    }
  }

  void _handleExpeditionComplete(Map<String, dynamic>? data) {
    if (data != null) {
      _expeditionChainProvider.onExpeditionComplete(data);
      final status = data['status'] as String?;
      if (status == 'completed') {
        final location = data['location'] as Map<String, dynamic>?;
        final locName = location?['name'] as String? ?? 'неизвестна';
        addNotification('Экспедиция завершена', 'Обнаружена локация: $locName');
      } else if (status == 'failed') {
        final error = data['error'] as String? ?? 'Неизвестная ошибка';
        addNotification('Экспедиция провалена', error);
      }
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
    _researchProvider.setPlanetId(planet.id);
    _researchProvider.loadResearch(planet.id);
    _marketProvider.setPlanetId(planet.id);
    _marketProvider.loadMarketData();
    _marketProvider.loadMyOrders();
    _expeditionProvider.setPlanetId(planet.id);
    _gardenBedProvider.getGardenBed(planet.id);
    _drillProvider.setDrillPlanetId(planet.id);
    _surveyProvider.notifyListeners();
    _expeditionChainProvider.loadExpeditionChains(planet.id);
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
    _planetSurveyUnlocked = data['planet_survey_unlocked'] as bool? ?? false;

    // Update research unlocks
    _researchUnlocks = data['research_unlocks'] as String? ?? '';

    // Update farm state
    final farmStateJson = data['garden_bed_state'] as Map<String, dynamic>?;
    if (farmStateJson != null && farmStateJson.isNotEmpty) {
      _gardenBedProvider.onGardenBedUpdate(farmStateJson);
    }

     // Update planet survey fields
      _surveyProvider.applyBuildDetails(data);

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
    if (_selectedPlanet == null) return const BuildingUpgradeInfo(canUpgrade: false, reason: 'Планета не выбрана');

    final isReady = building.isBuildComplete;
    if (isReady) return const BuildingUpgradeInfo(canUpgrade: false, reason: 'Нажмите чтобы открыть');

    final isBuilding = building.isBuilding;
    if (isBuilding) return const BuildingUpgradeInfo(canUpgrade: false, reason: 'Уже строится');

    final nextCostFood = building.nextCostFood;
    final nextCostIron = building.nextCostIron;
    final nextCostMoney = building.nextCostMoney;
    if (nextCostFood <= 0 && nextCostIron <= 0 && nextCostMoney <= 0) return const BuildingUpgradeInfo(canUpgrade: false, reason: 'Максимальный уровень');

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

    if (activeConstructions >= maxConstructions) return const BuildingUpgradeInfo(canUpgrade: false, reason: 'Достигнут лимит строительства');

    return const BuildingUpgradeInfo(canUpgrade: true, reason: null);
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


  @override
  void dispose() {
    websocket.removeMessageListener(_onWebSocketMessage);
    super.dispose();
  }
}

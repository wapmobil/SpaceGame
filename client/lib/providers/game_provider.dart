import 'dart:convert';
import 'dart:html' show window;
import 'dart:math' as math;
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../core/websocket_manager.dart';
import '../models/planet.dart';
import '../models/building.dart';
import '../models/ship.dart';
import '../models/research.dart';
import '../models/expedition.dart';
import '../models/market.dart';
import '../models/mining.dart';
import '../models/battle.dart';
import '../models/rating.dart';
import '../models/player.dart';

class GameProvider extends ChangeNotifier {
  final WebSocketManager websocket;
  String _baseUrl;
  Player? _player;
  List<Planet> _planets = [];
  Planet? _selectedPlanet;
  List<Building> _buildings = [];
  List<Ship> _ships = [];
  ShipyardInfo? _shipyardInfo;
  List<ShipType> _availableShipTypes = [];
  ResearchState? _researchState;
  List<BattleEntry> _battles = [];
  ExpeditionsListResponse? _expeditions;
  MarketData? _marketData;
  List<MarketOrder> _myOrders = [];
  MiningState? _miningState;
  List<RatingEntry> _ratings = [];
  Map<String, dynamic>? _stats;
  List<Map<String, dynamic>> _events = [];
  String? _errorMessage;

  GameProvider({required this.websocket, String? baseUrl})
      : _baseUrl = baseUrl ?? _getBaseUri();

  static String _getBaseUri() {
    if (kIsWeb) {
      final host = window.location.host;
      final scheme = window.location.protocol == 'https:' ? 'https' : 'http';
      return '$scheme://$host';
    }
    return 'http://localhost:8080';
  }

  String get baseUrl => _baseUrl;
  Player? get player => _player;
  List<Planet> get planets => _planets;
  Planet? get selectedPlanet => _selectedPlanet;
  List<Building> get buildings => _buildings;
  List<Ship> get ships => _ships;
  ShipyardInfo? get shipyardInfo => _shipyardInfo;
  List<ShipType> get availableShipTypes => _availableShipTypes;
  ResearchState? get researchState => _researchState;
  List<BattleEntry> get battles => _battles;
  ExpeditionsListResponse? get expeditions => _expeditions;
  MarketData? get marketData => _marketData;
  List<MarketOrder> get myOrders => _myOrders;
  MiningState? get miningState => _miningState;
  List<RatingEntry> get ratings => _ratings;
  Map<String, dynamic>? get stats => _stats;
  List<Map<String, dynamic>> get events => _events;
  String? get errorMessage => _errorMessage;
  bool get isLoggedIn => _player != null;
  String? get authToken => _player?.authToken;

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
        await _savePlayer();
        notifyListeners();
        connectWebSocket();
      } else {
        _setError('Failed to login: ${response.statusCode}');
      }
    } catch (e) {
      _setError('Login error: $e');
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
      case 'battle_update':
        _handleBattleUpdate(data);
        break;
      case 'expedition_update':
        _handleExpeditionUpdate(data);
        break;
      case 'market_update':
        _handleMarketUpdate(data);
        break;
      case 'mining_update':
        _handleMiningUpdate(data);
        break;
      case 'notification':
        _handleNotification(data);
        break;
    }
  }

  void _handleStateUpdate(Map<String, dynamic>? data) {
    if (data != null && _selectedPlanet != null) {
      final resources = data['resources'] as Map<String, dynamic>? ??
          data['state']?['resources'] as Map<String, dynamic>?;
      if (resources != null) {
        _selectedPlanet = _selectedPlanet!.copyWith(resources: resources);
      }

      final state = data['state'] as Map<String, dynamic>? ?? data;
      final buildProgress = state['build_progress'] as Map<String, dynamic>?;
      final buildingsMap = state['buildings'] as Map<String, dynamic>?;
      final buildSpeed = (state['build_speed'] as num?)?.toDouble() ?? 1.0;

      if (buildingsMap != null && _selectedPlanet != null) {
        final planetId = _selectedPlanet!.id;
        final updatedBuildings = <Building>[];

        for (final entry in buildingsMap.entries) {
          final type = entry.key as String;
          final level = (entry.value as num).toInt();
          final remainingTicks = buildProgress?[type] as num? ?? 0;
          final remainingTicksDouble = remainingTicks.toDouble();

          final totalBuildTime = _getTotalBuildTime(type, level, buildSpeed);
          final isUnderConstruction = buildProgress != null && buildProgress.containsKey(type);
          final progress = isUnderConstruction
              ? (totalBuildTime > 0 && remainingTicksDouble > 0
                  ? 1.0 - (remainingTicksDouble / totalBuildTime)
                  : 1.0)
              : 1.0;

          final existing = _buildings.where((b) => b.type == type).toList();
          if (existing.isNotEmpty) {
            updatedBuildings.add(Building(
              id: existing.first.id,
              planetId: planetId,
              type: type,
              level: level,
              buildProgress: progress,
              totalBuildTime: totalBuildTime,
            ));
          } else {
            updatedBuildings.add(Building(
              planetId: planetId,
              type: type,
              level: level,
              buildProgress: progress,
              totalBuildTime: totalBuildTime,
            ));
          }
        }

        _buildings = updatedBuildings;
        notifyListeners();
      }
    }
  }

  double _getTotalBuildTime(String type, int level, double buildSpeed) {
    double raw = 100.0;
    switch (type) {
      case 'farm':
        raw = (level * level * level * 20 + 100).toDouble();
        break;
      case 'solar':
        raw = (level * level * 200 + 80).toDouble();
        break;
      case 'storage':
        raw = ((level * level + 1) * 100).toDouble();
        break;
      case 'base':
        raw = math.pow(2, level + 3) + 100;
        break;
      case 'factory':
        raw = ((level * 2 + 1) * 100000).toDouble();
        break;
      case 'energy_storage':
        raw = (level * level + 1000).toDouble();
        break;
      case 'shipyard':
        raw = math.pow(2, level + 7) + 3000;
        break;
      case 'comcenter':
        raw = level == 0 ? 10000000.0 : (10000000 * level).toDouble();
        break;
      case 'composite_drone':
      case 'mechanism_factory':
      case 'reagent_lab':
        raw = ((level * level + 1) * 100).toDouble();
        break;
    }
    return raw / buildSpeed;
  }

  void _handleBuildingUpdate(Map<String, dynamic>? data) {
    if (data != null && _selectedPlanet != null) {
      final buildingType = data['building'] as String?;
      final newLevel = (data['level'] as num?)?.toInt() ?? 0;
      if (buildingType != null && newLevel > 0) {
        final planetId = _selectedPlanet!.id;
        final totalBuildTime = _getTotalBuildTime(buildingType, newLevel, 1.0);
        final existing = _buildings.where((b) => b.type == buildingType).toList();
        if (existing.isNotEmpty) {
          final idx = _buildings.indexWhere((b) => b.type == buildingType);
          if (idx >= 0) {
            _buildings[idx] = Building(
              id: existing.first.id,
              planetId: planetId,
              type: buildingType,
              level: newLevel,
              buildProgress: 1.0,
              totalBuildTime: totalBuildTime,
            );
          }
        } else {
          _buildings.add(Building(
            planetId: planetId,
            type: buildingType,
            level: newLevel,
            buildProgress: 1.0,
            totalBuildTime: totalBuildTime,
          ));
        }
        notifyListeners();
      }
    }
  }

  void _handleResearchUpdate(Map<String, dynamic>? data) {
    if (data != null) {
      loadResearch(_selectedPlanet?.id ?? '');
      notifyListeners();
    }
  }

  void _handleBattleUpdate(Map<String, dynamic>? data) {
    if (data != null) {
      loadBattles(_selectedPlanet?.id ?? '');
      notifyListeners();
    }
  }

  void _handleExpeditionUpdate(Map<String, dynamic>? data) {
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

  void _handleMiningUpdate(Map<String, dynamic>? data) {
    if (data != null) {
      loadMiningState(_selectedPlanet?.id ?? '');
      notifyListeners();
    }
  }

  void _handleNotification(Map<String, dynamic>? data) {
    // Notifications are handled by the UI layer
    debugPrint('Notification: ${data?['message']}');
  }

  Future<void> loadPlanets() async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as List;
        _planets = data.map((e) => Planet.fromJson(e as Map<String, dynamic>)).toList();
        notifyListeners();
      }
    } catch (e) {
      _setError('Failed to load planets: $e');
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
        _setError('Failed to create planet: ${response.statusCode}');
      }
    } catch (e) {
      _setError('Create planet error: $e');
    }
  }

  void selectPlanet(Planet planet) {
    _selectedPlanet = planet;
    websocket.subscribe(planet.id);
    loadBuildings(planet.id);
    loadShips(planet.id);
    loadResearch(planet.id);
    loadBattles(planet.id);
    loadExpeditions(planet.id);
    loadMarketData(planet.id);
    loadMyOrders(planet.id);
    loadMiningState(planet.id);
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

  Future<void> loadBuildings(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/buildings'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as List;
        _buildings = data.map((e) => Building.fromJson(e as Map<String, dynamic>)).toList();
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load buildings: $e');
    }
  }

  Future<void> buildStructure(String buildingType) async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/buildings'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({'type': buildingType}),
      );

      if (response.statusCode == 201) {
        await loadBuildings(_selectedPlanet!.id);
        await loadPlanetDetail(_selectedPlanet!.id);
      } else {
        _setError('Build failed: ${response.body}');
      }
    } catch (e) {
      _setError('Build error: $e');
    }
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
        _setError('Failed to build ship: ${response.body}');
      }
    } catch (e) {
      _setError('Build ship error: $e');
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
        _researchState = ResearchState.fromJson(
            jsonDecode(response.body) as Map<String, dynamic>);
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
        _setError('Failed to start research: ${response.body}');
      }
    } catch (e) {
      _setError('Research error: $e');
    }
  }

  Future<void> loadBattles(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/battles'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final battlesData = data['battles'] as List? ?? [];
        _battles = battlesData
            .map((e) => BattleEntry.fromJson(e as Map<String, dynamic>))
            .toList();
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load battles: $e');
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
        _setError('Failed to start expedition: ${response.body}');
      }
    } catch (e) {
      _setError('Expedition error: $e');
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
        _setError('Expedition action failed: ${response.body}');
      }
    } catch (e) {
      _setError('Expedition action error: $e');
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
        _setError('Failed to create order: ${response.body}');
      }
    } catch (e) {
      _setError('Create order error: $e');
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
        _setError('Failed to delete order: ${response.body}');
      }
    } catch (e) {
      _setError('Delete order error: $e');
    }
  }

  Future<void> loadMiningState(String planetId) async {
    if (_player == null) return;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/planets/$planetId/mining'),
        headers: {'X-Auth-Token': _player!.authToken},
      );

      if (response.statusCode == 200) {
        _miningState = MiningState.fromJson(
            jsonDecode(response.body) as Map<String, dynamic>);
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load mining state: $e');
    }
  }

  Future<void> startMining() async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/mining/start'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
      );

      if (response.statusCode == 201) {
        _miningState = MiningState.fromJson(
            jsonDecode(response.body) as Map<String, dynamic>);
        notifyListeners();
      } else {
        _setError('Failed to start mining: ${response.body}');
      }
    } catch (e) {
      _setError('Mining error: $e');
    }
  }

  Future<void> miningMove(String direction, {bool slide = false}) async {
    if (_player == null || _selectedPlanet == null) return;
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/api/planets/${_selectedPlanet!.id}/mining/move'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _player!.authToken,
        },
        body: jsonEncode({'direction': direction, 'slide': slide}),
      );

      if (response.statusCode == 200) {
        _miningState = MiningState.fromJson(
            jsonDecode(response.body) as Map<String, dynamic>);
        notifyListeners();
      } else {
        _setError('Mining move failed: ${response.body}');
      }
    } catch (e) {
      _setError('Mining move error: $e');
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
        _setError('Failed to resolve event: ${response.body}');
      }
    } catch (e) {
      _setError('Resolve event error: $e');
    }
  }

  @override
  void dispose() {
    websocket.removeMessageListener(_onWebSocketMessage);
    super.dispose();
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

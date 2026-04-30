import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

class PlanetSurveyProvider extends ChangeNotifier {
  String? _authToken;
  String? baseUrl;

  List<Map<String, dynamic>> _surfaceExpeditions = [];
  List<Map<String, dynamic>> _locations = [];
  Map<String, dynamic> _rangeStats = {};
  List<Map<String, dynamic>> _expeditionHistory = [];
  int? _baseLevel;
  int? _commandCenterLevel;
  int? _maxSurfaceExpeditions;
  bool? _canStartPlanetSurvey;
  bool? _canStartSpaceExpedition;

  String? get authToken => _authToken;

  void setAuthToken(String token) {
    _authToken = token;
  }

  List<Map<String, dynamic>> get surfaceExpeditions => _surfaceExpeditions;
  List<Map<String, dynamic>> get locations => _locations;
  Map<String, dynamic> get rangeStats => _rangeStats;
  List<Map<String, dynamic>> get expeditionHistory => _expeditionHistory;
  int? get baseLevel => _baseLevel;
  int? get commandCenterLevel => _commandCenterLevel;
  int? get maxSurfaceExpeditions => _maxSurfaceExpeditions;
  bool? get canStartPlanetSurvey => _canStartPlanetSurvey;
  bool? get canStartSpaceExpedition => _canStartSpaceExpedition;

  Future<void> loadPlanetSurveyData(String planetId) async {
    if (_authToken == null) return;
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/planets/$planetId/expedition-history'),
        headers: _authHeaders(),
      );
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as List;
        _expeditionHistory = data.map((e) => e as Map<String, dynamic>).toList();
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load expedition history: $e');
    }
  }

  Future<void> startPlanetSurvey(String planetId, int duration) async {
    if (_authToken == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$planetId/planet-survey'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode({'duration': duration}),
      );

      if (response.statusCode == 200) {
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Ошибка разведки: $e');
    }
  }

  Future<void> buildOnLocation(String planetId, String locationId, String buildingType) async {
    if (_authToken == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$planetId/locations/$locationId/build'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode({'building_type': buildingType}),
      );

      if (response.statusCode == 200) {
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Ошибка строительства: $e');
    }
  }

  Future<void> removeBuilding(String planetId, String locationId) async {
    if (_authToken == null) return;
    try {
      final response = await http.delete(
        Uri.parse('${baseUrl!}/api/planets/$planetId/locations/$locationId/building'),
        headers: _authHeaders(),
      );

      if (response.statusCode == 200) {
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Ошибка: $e');
    }
  }

  Future<void> abandonLocation(String planetId, String locationId) async {
    if (_authToken == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$planetId/locations/$locationId/abandon'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
      );

      if (response.statusCode == 200) {
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Ошибка: $e');
    }
  }

  void onPlanetSurveyUpdate(Map<String, dynamic>? data) {
    if (data == null) return;
    if (data['expedition_id'] != null) {
      final expId = data['expedition_id'] as String;
      final idx = _surfaceExpeditions.indexWhere((e) => e['id'] == expId);
      if (idx >= 0) {
        _surfaceExpeditions[idx] = {..._surfaceExpeditions[idx], ...data};
      } else {
        _surfaceExpeditions.add(data);
      }
    }
    notifyListeners();
  }

  void onLocationUpdate(Map<String, dynamic>? data) {
    if (data == null) return;
    if (data['location_id'] != null) {
      final locId = data['location_id'] as String;
      final idx = _locations.indexWhere((l) => l['id'] == locId);
      if (idx >= 0) {
        _locations[idx] = {..._locations[idx], ...data};
      } else {
        _locations.add(data);
      }
    }
    notifyListeners();
  }

  void applyStateUpdate(Map<String, dynamic> stateData) {
    _surfaceExpeditions = (stateData['surface_expeditions'] as List?)
        ?.map((e) => e as Map<String, dynamic>)
        .toList() ?? [];
    _locations = (stateData['locations'] as List?)
        ?.map((e) => e as Map<String, dynamic>)
        .toList() ?? [];
    _rangeStats = stateData['range_stats'] as Map<String, dynamic>? ?? {};
    if (stateData['expedition_history'] is List &&
        (stateData['expedition_history'] as List).length > _expeditionHistory.length) {
      _expeditionHistory = (stateData['expedition_history'] as List)
          .map((e) => e as Map<String, dynamic>)
          .toList();
    }
    _maxSurfaceExpeditions = stateData['max_surface_expeditions'] as int?;
    _canStartPlanetSurvey = stateData['can_start_planet_survey'] as bool?;
    _canStartSpaceExpedition = stateData['can_start_space_expedition'] as bool?;
    notifyListeners();
  }

  void applyBuildDetails(Map<String, dynamic> data) {
    _canStartPlanetSurvey = data['can_start_planet_survey'] as bool?;
    _canStartSpaceExpedition = data['can_start_space_expedition'] as bool?;
    _baseLevel = data['base_level'] as int?;
    _commandCenterLevel = data['command_center_level'] as int?;
    _maxSurfaceExpeditions = data['max_surface_expeditions'] as int? ?? 1;
    notifyListeners();
  }

  Map<String, String> _authHeaders() {
    return {'X-Auth-Token': _authToken!};
  }
}

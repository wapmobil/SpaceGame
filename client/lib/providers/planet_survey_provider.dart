import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

class PlanetSurveyProvider extends ChangeNotifier {
  String? _authToken;
  String? baseUrl;

  List<Map<String, dynamic>> _locations = [];
  int? _baseLevel;
  int? _commandCenterLevel;
  bool? _canStartExpedition;
  bool? _canStartSpaceExpedition;
  bool? _canStartPlanetSurvey;
  String? _currentPlanetId;

  String? get authToken => _authToken;

  void setAuthToken(String token) {
    _authToken = token;
  }

  List<Map<String, dynamic>> get locations => _locations;
  int? get baseLevel => _baseLevel;
  int? get commandCenterLevel => _commandCenterLevel;
  bool? get canStartExpedition => _canStartExpedition;
  bool? get canStartSpaceExpedition => _canStartSpaceExpedition;
  bool? get canStartPlanetSurvey => _canStartPlanetSurvey;
  String? get currentPlanetId => _currentPlanetId;

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
    _locations = (stateData['locations'] as List?)
        ?.map((e) => e as Map<String, dynamic>)
        .toList() ?? [];
    _canStartExpedition = stateData['can_start_expedition'] as bool?;
    _canStartSpaceExpedition = stateData['can_start_space_expedition'] as bool?;
    notifyListeners();
  }

  void applyBuildDetails(Map<String, dynamic> data) {
    _canStartExpedition = data['can_start_expedition'] as bool?;
    _canStartSpaceExpedition = data['can_start_space_expedition'] as bool?;
    _baseLevel = data['base_level'] as int?;
    _commandCenterLevel = data['command_center_level'] as int?;
    notifyListeners();
  }

  Map<String, String> _authHeaders() {
    return {'X-Auth-Token': _authToken!};
  }
}

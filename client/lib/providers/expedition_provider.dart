import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../models/expedition.dart';

class ExpeditionProvider extends ChangeNotifier {
  String? _authToken;
  String? baseUrl;
  String? _planetId;

  ExpeditionsListResponse? _expeditions;
  String? _errorMessage;

  ExpeditionsListResponse? get expeditions => _expeditions;
  String? get errorMessage => _errorMessage;

  void setAuthToken(String token) {
    _authToken = token;
  }

  void setPlanetId(String planetId) {
    _planetId = planetId;
  }

  Future<void> startExpedition({
    required String expeditionType,
    String? target,
    List<String>? shipTypes,
    List<int>? shipCounts,
    double? duration,
  }) async {
    if (_authToken == null || _planetId == null) return;
    try {
      final body = <String, dynamic>{
        'expedition_type': expeditionType,
        'duration': duration ?? 3600,
      };
      if (target != null) body['target'] = target;
      if (shipTypes != null) body['ship_types'] = shipTypes;
      if (shipCounts != null) body['ship_counts'] = shipCounts;

      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$_planetId/expeditions'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode(body),
      );

      if (response.statusCode == 201) {
        notifyListeners();
      } else {
        _errorMessage = 'Не удалось начать экспедицию: ${response.body}';
        notifyListeners();
      }
    } catch (e) {
      _errorMessage = 'Ошибка экспедиции: $e';
      notifyListeners();
    }
  }

  Future<void> expeditionAction(String expeditionId, String action) async {
    if (_authToken == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/expeditions/$expeditionId/action'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode({'action': action}),
      );

      if (response.statusCode == 200) {
        notifyListeners();
      } else {
        _errorMessage = 'Не удалось выполнить действие экспедиции: ${response.body}';
        notifyListeners();
      }
    } catch (e) {
      _errorMessage = 'Ошибка действия экспедиции: $e';
      notifyListeners();
    }
  }

  void onExpeditionUpdate() {
    notifyListeners();
  }

  void applyStateUpdate(Map<String, dynamic> stateData) {
    final expeditionsJson = stateData['expeditions'] as List? ?? [];
    final activeCount = stateData['active_expeditions'] as int? ?? 0;
    final maxExpeditions = stateData['max_expeditions'] as int? ?? 1;
    final canStartNew = stateData['can_start_space_expedition'] as bool? ?? false;
    final expeditionsUnlocked = stateData['can_expedition'] as bool? ?? false;

    final parsedExpeditions = expeditionsJson
        .map((e) => Expedition.fromJson(e as Map<String, dynamic>))
        .toList();

    _expeditions = ExpeditionsListResponse(
      expeditions: parsedExpeditions,
      activeCount: activeCount,
      maxExpeditions: maxExpeditions,
      canStartNew: canStartNew,
      expeditionsUnlocked: expeditionsUnlocked,
    );
    notifyListeners();
  }

  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }
}

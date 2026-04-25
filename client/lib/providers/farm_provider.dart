import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../models/farm.dart';
import '../core/websocket_manager.dart';

class FarmProvider extends ChangeNotifier {
  final WebSocketManager websocket;
  final String baseUrl;
  String? _authToken;
  FarmState? _farmState;
  DateTime? _lastActionTime;
  bool _isLoading = false;
  String? _errorMessage;

  FarmProvider({required this.websocket, required this.baseUrl, String? authToken})
      : _authToken = authToken;

  FarmState? get farmState => _farmState;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;
  bool get canAct => _lastActionTime == null || DateTime.now().difference(_lastActionTime!).inSeconds >= 5;
  int get remainingCooldown => _lastActionTime == null
      ? 0
      : (5 - DateTime.now().difference(_lastActionTime!).inSeconds).clamp(0, 5);

  void setAuthToken(String token) {
    _authToken = token;
  }

  void _setError(String msg) {
    _errorMessage = msg;
    notifyListeners();
  }

  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }

  Future<void> getFarm(String planetId) async {
    if (_authToken == null) return;
    _isLoading = true;
    notifyListeners();

    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/planets/$planetId/farm'),
        headers: {'X-Auth-Token': _authToken!},
      );

      if (response.statusCode == 200) {
        if (response.body.isEmpty) {
          _farmState = null;
          _isLoading = false;
          _errorMessage = null;
          notifyListeners();
          return;
        }
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        _farmState = FarmState.fromJson(data);
        _isLoading = false;
        _errorMessage = null;
        notifyListeners();
      } else if (response.statusCode == 404) {
        _farmState = null;
        _isLoading = false;
        _errorMessage = null;
        notifyListeners();
      } else {
        _setError('Failed to load farm: ${response.statusCode}');
        _isLoading = false;
        notifyListeners();
      }
    } catch (e) {
      _setError('Error loading farm: $e');
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> farmAction(String planetId, String action, int rowIndex, {String? plantType}) async {
    if (_authToken == null) return;
    _isLoading = true;
    _lastActionTime = DateTime.now();
    notifyListeners();

    try {
      final body = <String, dynamic>{
        'action': action,
        'row_index': rowIndex,
      };
      if (plantType != null) {
        body['plant_type'] = plantType;
      }

      final response = await http.post(
        Uri.parse('$baseUrl/api/planets/$planetId/farm/action'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final rowsData = data['rows'] as List? ?? [];
        final rows = rowsData.map((r) => FarmRow.fromJson(r as Map<String, dynamic>)).toList();
        _farmState = FarmState(
          rows: rows,
          lastTick: data['last_tick'] as int? ?? _farmState?.lastTick ?? 0,
          rowCount: _farmState?.rowCount ?? rows.length,
        );
        _isLoading = false;
        _errorMessage = null;
        notifyListeners();
      } else if (response.statusCode == 400) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final error = data['error'] as String? ?? 'Action failed';
        _setError(error);
        _isLoading = false;
        notifyListeners();
        // Revert cooldown on error
        _lastActionTime = null;
        notifyListeners();
      } else {
        _setError('Failed to perform action: ${response.statusCode}');
        _isLoading = false;
        notifyListeners();
        _lastActionTime = null;
        notifyListeners();
      }
    } catch (e) {
      _setError('Error: $e');
      _isLoading = false;
      notifyListeners();
      _lastActionTime = null;
      notifyListeners();
    }
  }

  void onFarmUpdate(Map<String, dynamic> data) {
    try {
      final rowsData = data['rows'] as List? ?? [];
      final rows = rowsData.map((r) => FarmRow.fromJson(r as Map<String, dynamic>)).toList();

      _farmState = FarmState(
        rows: rows,
        lastTick: data['last_tick'] as int? ?? _farmState?.lastTick ?? 0,
        rowCount: _farmState?.rowCount ?? rows.length,
      );

      // Reset cooldown and error on successful update
      _lastActionTime = null;
      _errorMessage = null;
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to process farm update: $e');
    }
  }

  void onFarmWSActionResult(Map<String, dynamic> data) {
    try {
      final success = data['success'] as bool? ?? false;
      if (success) {
        final rowsData = data['rows'] as List? ?? [];
        final rows = rowsData.map((r) => FarmRow.fromJson(r as Map<String, dynamic>)).toList();

        _farmState = FarmState(
          rows: rows,
          lastTick: data['last_tick'] as int? ?? _farmState?.lastTick ?? 0,
          rowCount: _farmState?.rowCount ?? rows.length,
        );
        _lastActionTime = null;
        _errorMessage = null;
        notifyListeners();
      } else {
        _setError(data['error'] as String? ?? 'Action failed');
        _lastActionTime = null;
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to process farm WS action result: $e');
    }
  }

  String getPlantName(String type) {
    switch (type) {
      case 'wheat':
        return 'Пшеница';
      case 'berries':
        return 'Ягоды';
      case 'melon':
        return 'Космическая дыня';
      default:
        return 'Растение';
    }
  }

  String getPlantIcon(String type) {
    switch (type) {
      case 'wheat':
        return '🌾';
      case 'berries':
        return '🫐';
      case 'melon':
        return '🍈';
      default:
        return '🌱';
    }
  }

  String getStageName(int stage) {
    switch (stage) {
      case 0:
        return 'Семя';
      case 1:
        return 'Росток';
      case 2:
        return 'Созрело';
      default:
        return 'Неизвестно';
    }
  }

  String getRowStatusText(FarmRow row) {
    if (row.isEmpty) return 'Пусто';
    if (row.isMature) return 'Собрать';
    if (row.isPlanted) {
      final plantName = getPlantName(row.plantType ?? '');
      final stageName = getStageName(row.stage ?? 0);
      return '$plantName - $stageName';
    }
    return '';
  }

  double getRowProgress(FarmRow row) {
    if (!row.isPlanted || row.isMature) return 1.0;
    final stage = row.stage ?? 0;
    return (stage + 1) / 3.0;
  }

  void dispose() {
    super.dispose();
  }
}

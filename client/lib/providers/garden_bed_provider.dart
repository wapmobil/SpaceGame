import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../models/garden_bed.dart';
import '../core/websocket_manager.dart';

class GardenBedProvider extends ChangeNotifier {
  final WebSocketManager websocket;
  final String baseUrl;
  String? _authToken;
  GardenBedState? _gardenBedState;
  bool _isLoading = false;
  String? _errorMessage;

  GardenBedProvider({required this.websocket, required this.baseUrl, String? authToken})
      : _authToken = authToken;

  GardenBedState? get gardenBedState => _gardenBedState;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;

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

  // Plant definitions
  Map<String, Map<String, dynamic>> get allPlants => {
    'wheat': {'type': 'wheat', 'name': 'Пшеница', 'icon': '🌾', 'seedCost': 5, 'moneyReward': 15, 'foodReward': 5, 'unlockLevel': 1, 'weedCost': 2, 'waterCost': 1, 'growthTicks': 36},
    'berries': {'type': 'berries', 'name': 'Ягоды', 'icon': '🫐', 'seedCost': 15, 'moneyReward': 45, 'foodReward': 15, 'unlockLevel': 2, 'weedCost': 5, 'waterCost': 3, 'growthTicks': 90},
    'raspberry': {'type': 'raspberry', 'name': 'Малина', 'icon': '🪴', 'seedCost': 25, 'moneyReward': 80, 'foodReward': 25, 'unlockLevel': 3, 'weedCost': 15, 'waterCost': 5, 'growthTicks': 180},
    'rose': {'type': 'rose', 'name': 'Космическая роза', 'icon': '🌷', 'seedCost': 60, 'moneyReward': 200, 'foodReward': 50, 'unlockLevel': 5, 'weedCost': 25, 'waterCost': 10, 'growthTicks': 270},
    'sunflower': {'type': 'sunflower', 'name': 'Космический подсолнух', 'icon': '🌻', 'seedCost': 120, 'moneyReward': 400, 'foodReward': 80, 'unlockLevel': 7, 'weedCost': 20, 'waterCost': 30, 'growthTicks': 360},
    'melon': {'type': 'melon', 'name': 'Космическая дыня', 'icon': '🍈', 'seedCost': 250, 'moneyReward': 800, 'foodReward': 120, 'unlockLevel': 9, 'weedCost': 30, 'waterCost': 20, 'growthTicks': 540},
    'banana': {'type': 'banana', 'name': 'Лунный банан', 'icon': '🌙', 'seedCost': 500, 'moneyReward': 1700, 'foodReward': 150, 'unlockLevel': 11, 'weedCost': 50, 'waterCost': 50, 'growthTicks': 1080},
    'blueberry': {'type': 'blueberry', 'name': 'Звёздная голубика', 'icon': '🫐', 'seedCost': 1000, 'moneyReward': 3500, 'foodReward': 300, 'unlockLevel': 13, 'weedCost': 80, 'waterCost': 50, 'growthTicks': 2880},
  };

  Map<String, dynamic>? getPlant(String type) {
    return allPlants[type];
  }

  List<Map<String, dynamic>> getAvailablePlants(int farmLevel) {
    return allPlants.values.where((p) => (p['unlockLevel'] as int) <= farmLevel).toList();
  }

  bool isPlantUnlocked(String type, int farmLevel) {
    final plant = allPlants[type];
    if (plant == null) return false;
    return (plant['unlockLevel'] as int) <= farmLevel;
  }

  double getSeedCost(String type) {
    return allPlants[type]?['seedCost']?.toDouble() ?? 0;
  }

  double getMoneyReward(String type) {
    return allPlants[type]?['moneyReward']?.toDouble() ?? 0;
  }

  double getFoodReward(String type) {
    return allPlants[type]?['foodReward']?.toDouble() ?? 0;
  }

  int getUnlockLevel(String type) {
    return allPlants[type]?['unlockLevel'] ?? 1;
  }

  double getWeedCost(String type) {
    return allPlants[type]?['weedCost']?.toDouble() ?? 2;
  }

  double getWaterCost(String type) {
    return allPlants[type]?['waterCost']?.toDouble() ?? 1;
  }

  int getGrowthTicks(String type) {
    return allPlants[type]?['growthTicks'] ?? 36;
  }

  Future<void> getGardenBed(String planetId) async {
    if (_authToken == null) return;
    _isLoading = true;
    notifyListeners();

    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/planets/$planetId/garden_bed'),
        headers: {'X-Auth-Token': _authToken!},
      );

      if (response.statusCode == 200) {
        if (response.body.isEmpty) {
          _gardenBedState = null;
          _isLoading = false;
          _errorMessage = null;
          notifyListeners();
          return;
        }
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        _gardenBedState = GardenBedState.fromJson(data);
        _isLoading = false;
        _errorMessage = null;
        notifyListeners();
      } else if (response.statusCode == 404) {
        _gardenBedState = null;
        _isLoading = false;
        _errorMessage = null;
        notifyListeners();
      } else {
        _setError('Failed to load garden bed: ${response.statusCode}');
        _isLoading = false;
        notifyListeners();
      }
    } catch (e) {
      _setError('Error loading garden bed: $e');
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> gardenBedAction(String planetId, String action, int rowIndex, {String? plantType}) async {
    if (_authToken == null) return;
    _isLoading = true;
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
        Uri.parse('$baseUrl/api/planets/$planetId/garden_bed/action'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final rowsData = data['rows'] as List? ?? [];
        final rows = rowsData.map((r) => GardenBedRow.fromJson(r as Map<String, dynamic>)).toList();
        _gardenBedState = GardenBedState(
          rows: rows,
          lastTick: data['last_tick'] as int? ?? _gardenBedState?.lastTick ?? 0,
          rowCount: _gardenBedState?.rowCount ?? rows.length,
        );
        _isLoading = false;
        _errorMessage = null;
        notifyListeners();
      } else if (response.statusCode == 400) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final rowsData = data['rows'] as List? ?? [];
        final rows = rowsData.map((r) => GardenBedRow.fromJson(r as Map<String, dynamic>)).toList();
        if (rows.isNotEmpty) {
          _gardenBedState = GardenBedState(
            rows: rows,
            lastTick: data['last_tick'] as int? ?? _gardenBedState?.lastTick ?? 0,
            rowCount: _gardenBedState?.rowCount ?? rows.length,
          );
        }
        final error = data['error'] as String? ?? 'Action failed';
        _setError(error);
        _isLoading = false;
        notifyListeners();
      } else {
        _setError('Failed to perform action: ${response.statusCode}');
        _isLoading = false;
        notifyListeners();
      }
    } catch (e) {
      _setError('Error: $e');
      _isLoading = false;
      notifyListeners();
    }
  }

  void onGardenBedUpdate(Map<String, dynamic> data) {
    try {
      final rowsData = data['rows'] as List? ?? [];
      final rows = rowsData.map((r) => GardenBedRow.fromJson(r as Map<String, dynamic>)).toList();

      _gardenBedState = GardenBedState(
        rows: rows,
        lastTick: data['last_tick'] as int? ?? _gardenBedState?.lastTick ?? 0,
        rowCount: _gardenBedState?.rowCount ?? rows.length,
      );

      _errorMessage = null;
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to process garden bed update: $e');
    }
  }

  void onGardenBedWSActionResult(Map<String, dynamic> data) {
    try {
      final success = data['success'] as bool? ?? false;
      final rowsData = data['rows'] as List? ?? [];
      final rows = rowsData.map((r) => GardenBedRow.fromJson(r as Map<String, dynamic>)).toList();

      if (rows.isNotEmpty) {
        _gardenBedState = GardenBedState(
          rows: rows,
          lastTick: data['last_tick'] as int? ?? _gardenBedState?.lastTick ?? 0,
          rowCount: _gardenBedState?.rowCount ?? rows.length,
        );
      }

      if (success) {
        _errorMessage = null;
      } else {
        _setError(data['error'] as String? ?? 'Action failed');
      }
      notifyListeners();
    } catch (e) {
      debugPrint('Failed to process garden bed WS action result: $e');
    }
  }

  String getPlantName(String type) {
    final plant = allPlants[type];
    if (plant != null) return plant['name'] as String;
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
    final plant = allPlants[type];
    if (plant != null) return plant['icon'] as String;
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

  String getRowStatusText(GardenBedRow row) {
    if (row.isEmpty) return 'Пусто';
    if (row.isMature) return 'Собрать';
    if (row.isWithered) return 'Увядшее';
    if (row.isPlanted) {
      final plantName = getPlantName(row.plantType ?? '');
      final stageName = getStageName(row.stage ?? 0);
      return '$plantName - $stageName';
    }
    return '';
  }

  double getRowProgress(GardenBedRow row) {
    if (!row.isPlanted) return 0.0;
    if (row.isMature) return 1.0;
    final plantType = row.plantType ?? 'wheat';
    final growthTicks = getGrowthTicks(plantType);
    const stages = 3;
    final ticksPerStage = growthTicks ~/ (stages - 1);
    const maxStage = stages - 1;
    final totalTicks = maxStage * ticksPerStage;
    final completed = (row.stage ?? 0) * ticksPerStage + row.stageProgress;
    return (completed / totalTicks).clamp(0.0, 1.0);
  }

  String getTicksToMatureText(int ticks) {
    if (ticks <= 0) return '';
    // Each farm tick = 10 seconds
    final totalSeconds = ticks * 10;
    final minutes = totalSeconds ~/ 60;
    final seconds = totalSeconds % 60;
    if (minutes > 0 && seconds > 0) {
      return '$minutesм $secondsс';
    } else if (minutes > 0) {
      return '$minutesм';
    }
    return '$secondsс';
  }

  String getWitherStatusText(GardenBedRow row) {
    if (!row.isMature && !row.isWithered) return '';
    if (row.isWithered) return '🥀 Увядшее';
    final witherTimer = row.witherTimer;
    if (witherTimer <= 0) return '';
    return '⏱ $witherTimer';
  }

}

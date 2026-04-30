import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../models/drill.dart';
import '../core/websocket_manager.dart';

class DrillProvider extends ChangeNotifier {
  WebSocketManager? websocket;
  String? baseUrl;
  String? _authToken;

  DrillState? _drillState;
  int? _drillSeed;
  String? _drillPendingDirection;
  bool _drillPendingExtracting = false;

  DrillState? get drillState => _drillState;
  String? get drillPendingDirection => _drillPendingDirection;
  bool get drillPendingExtracting => _drillPendingExtracting;

  void setAuthToken(String token) {
    _authToken = token;
  }

  void clearDrillState() {
    _drillState = null;
    _drillSeed = null;
    _drillPendingDirection = null;
    _drillPendingExtracting = false;
    notifyListeners();
  }

  Future<void> startDrill(String planetId, {int speed = 1}) async {
    if (_authToken == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$planetId/drill/start'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode({'speed': speed}),
      );

      if (response.statusCode == 201) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final startResp = DrillStartResponse.fromJson(data);
        _drillSeed = startResp.seed;
        _drillState = DrillState(
          sessionId: startResp.sessionId,
          planetId: planetId,
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
      }
    } catch (e) {
      debugPrint('Ошибка бурения: $e');
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
    websocket?.sendDrillCommand(direction: direction, extract: extract);
  }

  void setDrillPlanetId(String planetId) {
    if (_drillState != null) {
      _drillState = DrillState(
        sessionId: _drillState!.sessionId,
        planetId: planetId,
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
        pendingDirection: _drillState!.pendingDirection,
        pendingExtracting: _drillState!.pendingExtracting,
      );
      notifyListeners();
    }
  }

  void onDrillUpdate(Map<String, dynamic> data) {
    try {
      final update = DrillUpdate.fromJson(data);
      _drillPendingDirection = null;
      _drillState = DrillState(
        sessionId: update.sessionId,
        planetId: _drillState?.planetId ?? '',
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

  Future<void> completeDrill(String planetId) async {
    if (_authToken == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$planetId/drill/complete'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
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
          planetId: planetId,
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
      }
    } catch (e) {
      debugPrint('Ошибка завершения: $e');
    }
  }

  Future<void> destroyDrill(String planetId) async {
    if (_authToken == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$planetId/drill/destroy'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
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
          planetId: planetId,
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

  Future<void> cleanupDrill(String planetId) async {
    if (_authToken == null) return;
    try {
      await http.post(
        Uri.parse('${baseUrl!}/api/planets/$planetId/drill/cleanup'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
      );
    } catch (e) {
      // ignore
    }
  }
}

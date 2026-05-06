import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../models/expedition_chain.dart';

class ExpeditionChainProvider extends ChangeNotifier {
  String? _authToken;
  String? baseUrl;

  List<ExpeditionChain> _chains = [];
  String? _selectedChainId;
  ExpeditionEvent? _currentEvent;
  bool _isLoading = false;

  List<ExpeditionChain> get chains => _chains;
  List<ExpeditionChain> get activeChains =>
      _chains.where((c) => c.isActive).toList();
  List<ExpeditionChain> get completedChains =>
      _chains.where((c) => c.isCompleted || c.isFailed).toList();
  ExpeditionChain? get selectedChain => _selectedChainId != null
      ? _chains.where((c) => c.id == _selectedChainId).firstOrNull
      : null;
  ExpeditionEvent? get currentEvent => _currentEvent;
  bool get isLoading => _isLoading;

  void setAuthToken(String token) {
    _authToken = token;
  }

  Future<void> loadExpeditionChains(String planetId) async {
    if (_authToken == null) return;
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/planets/$planetId/expeditions'),
        headers: _authHeaders(),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final chainsJson = data['chains'] as List? ?? [];
        _chains = chainsJson
            .map((c) => ExpeditionChain.fromJson(c as Map<String, dynamic>))
            .toList();

        if (_chains.isNotEmpty) {
          final active = _chains.firstWhere(
            (c) => c.isActive,
            orElse: () => _chains.first,
          );
          _selectedChainId = active.id;
          await getExpeditionEvent(planetId, active.id);
        }
      }
    } catch (e) {
      debugPrint('Failed to load expedition chains: $e');
    }
  }

  Future<ExpeditionEvent?> getExpeditionEvent(
    String planetId,
    String chainId,
  ) async {
    if (_authToken == null) return null;
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/planets/$planetId/expeditions/$chainId/event'),
        headers: _authHeaders(),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        _currentEvent = ExpeditionEvent.fromJson(data);
        notifyListeners();
        return _currentEvent;
      }
    } catch (e) {
      debugPrint('Failed to get expedition event: $e');
    }
    return null;
  }

  Future<ExpeditionChoiceResult> makeChoice(
    String planetId,
    String chainId,
    int choiceIndex,
  ) async {
    if (_authToken == null) {
      throw Exception('Not authenticated');
    }

    _isLoading = true;
    notifyListeners();

    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$planetId/expeditions/$chainId/choice'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode({'choice_index': choiceIndex}),
        // Extended timeout for LLM processing
      ).timeout(const Duration(seconds: 330));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final result = ExpeditionChoiceResult.fromJson(data);

        if (result.completed || result.failed) {
          final chainIdx = _chains.indexWhere((c) => c.id == chainId);
          if (chainIdx >= 0) {
            _chains[chainIdx] = result.chain;
          }
          _currentEvent = null;
        } else if (result.event != null) {
          _currentEvent = result.event;
          final chainIdx = _chains.indexWhere((c) => c.id == chainId);
          if (chainIdx >= 0) {
            _chains[chainIdx] = result.chain;
          }
        }

        notifyListeners();
        return result;
      }
    } catch (e) {
      debugPrint('Failed to make choice: $e');
      rethrow;
    } finally {
      _isLoading = false;
      notifyListeners();
    }

    throw Exception('Unexpected response');
  }

  Future<List<ExpeditionEventLogEntry>> getExpeditionEventLog(
    String planetId,
    String chainId,
  ) async {
    if (_authToken == null) return [];
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/planets/$planetId/expeditions/$chainId/event-log'),
        headers: _authHeaders(),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as List;
        return data
            .map((e) =>
                ExpeditionEventLogEntry.fromJson(e as Map<String, dynamic>))
            .toList();
      }
    } catch (e) {
      debugPrint('Failed to get event log: $e');
    }
    return [];
  }

  Future<ExpeditionChoiceResult> startExpedition(
    String planetId,
    Map<String, double> inventory,
  ) async {
    if (_authToken == null) {
      throw Exception('Not authenticated');
    }

    _isLoading = true;
    notifyListeners();

    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$planetId/expeditions'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode({'inventory': inventory}),
      ).timeout(const Duration(seconds: 330));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final event = ExpeditionEvent.fromJson(data['event'] as Map<String, dynamic>);
        final inventoryJson = data['inventory'] as Map<String, dynamic>;
        final inventory = <String, double>{};
        inventoryJson.forEach((key, value) {
          inventory[key] = (value as num).toDouble();
        });

        _currentEvent = event;
        await loadExpeditionChains(planetId);

        return ExpeditionChoiceResult(
          event: event,
          chain: _chains.firstWhere(
            (c) => c.id == data['chain_id'],
            orElse: () => _chains.first,
          ),
          inventory: inventory,
          completed: false,
          failed: false,
        );
      }
    } catch (e) {
      debugPrint('Failed to start expedition: $e');
      rethrow;
    } finally {
      _isLoading = false;
      notifyListeners();
    }

    throw Exception('Unexpected response');
  }

  void selectChain(String chainId) {
    _selectedChainId = chainId;
    notifyListeners();
  }

  double getInventoryTotal(Map<String, double> inventory) {
    return inventory.values.fold(0.0, (sum, v) => sum + v);
  }

  int getRemainingCapacity(double currentTotal) {
    return (1000 - currentTotal).toInt();
  }

  void onExpeditionEvent(Map<String, dynamic> data) {
    final chainId = data['chain_id'] as String?;
    if (chainId == null) return;

    final eventJson = data['event'] as Map<String, dynamic>?;
    if (eventJson != null) {
      _currentEvent = ExpeditionEvent.fromJson(eventJson);
    }

    final chainIdx = _chains.indexWhere((c) => c.id == chainId);
    if (chainIdx >= 0 && data['inventory'] != null) {
      final inventoryJson = data['inventory'] as Map<String, dynamic>;
      final inventory = <String, double>{};
      inventoryJson.forEach((key, value) {
        inventory[key] = (value as num).toDouble();
      });
      _chains[chainIdx] = ExpeditionChain(
        id: _chains[chainIdx].id,
        planetId: _chains[chainIdx].planetId,
        ownerId: _chains[chainIdx].ownerId,
        status: _chains[chainIdx].status,
        eventCount: (data['event_count'] as int?) ?? _chains[chainIdx].eventCount,
        currentEventIndex: _chains[chainIdx].currentEventIndex,
        inventory: inventory,
        events: _chains[chainIdx].events,
        discoveredLocation: _chains[chainIdx].discoveredLocation,
        createdAt: _chains[chainIdx].createdAt,
        updatedAt: DateTime.now().toUtc(),
      );
    }

    notifyListeners();
  }

  void onExpeditionComplete(Map<String, dynamic> data) {
    final chainId = data['chain_id'] as String?;
    if (chainId == null) return;

    final status = data['status'] as String?;
    if (status == 'completed' || status == 'failed') {
      final chainIdx = _chains.indexWhere((c) => c.id == chainId);
      if (chainIdx >= 0) {
        _chains[chainIdx] = ExpeditionChain(
          id: _chains[chainIdx].id,
          planetId: _chains[chainIdx].planetId,
          ownerId: _chains[chainIdx].ownerId,
          status: status!,
          eventCount: _chains[chainIdx].eventCount,
          currentEventIndex: _chains[chainIdx].currentEventIndex,
          inventory: _chains[chainIdx].inventory,
          events: _chains[chainIdx].events,
          discoveredLocation: _chains[chainIdx].discoveredLocation,
          createdAt: _chains[chainIdx].createdAt,
          updatedAt: DateTime.now().toUtc(),
        );
      }
      _currentEvent = null;
      notifyListeners();
    }
  }

  Map<String, String> _authHeaders() {
    return {'X-Auth-Token': _authToken!};
  }
}

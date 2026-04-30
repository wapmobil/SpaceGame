import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../models/research.dart';

class ResearchProvider extends ChangeNotifier {
  String? _authToken;
  String? baseUrl;
  String? _planetId;

  ResearchState? _researchState;
  bool _researchPaused = false;
  Map<String, int> _completedResearch = {};
  String? _errorMessage;

  ResearchState? get researchState => _researchState;
  bool get researchPaused => _researchPaused;
  Map<String, int> get completedResearch => _completedResearch;
  String? get errorMessage => _errorMessage;

  void setAuthToken(String token) {
    _authToken = token;
  }

  void setPlanetId(String planetId) {
    _planetId = planetId;
  }

  Future<void> loadResearch(String planetId) async {
    if (_authToken == null) return;
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/planets/$planetId/research'),
        headers: {'X-Auth-Token': _authToken!},
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
    if (_authToken == null || _planetId == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$_planetId/research/start'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode({'tech_id': techId}),
      );

      if (response.statusCode == 201) {
        await loadResearch(_planetId!);
      } else {
        _errorMessage = 'Не удалось начать исследование: ${response.body}';
        notifyListeners();
      }
    } catch (e) {
      _errorMessage = 'Ошибка исследования: $e';
      notifyListeners();
    }
  }

  void onResearchUpdate(Map<String, dynamic>? data) {
    if (data != null && _planetId != null) {
      loadResearch(_planetId!);
      notifyListeners();
    }
  }

  void applyResearchStateUpdate(Map<String, dynamic> stateData) {
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

    // Update research paused status
    final resPaused = stateData['research_paused'];
    if (resPaused != null) {
      _researchPaused = resPaused is bool ? resPaused : (resPaused as num).toInt() != 0;
    }
  }
}

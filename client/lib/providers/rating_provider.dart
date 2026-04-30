import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../models/rating.dart';

class RatingProvider extends ChangeNotifier {
  String? _authToken;
  String? baseUrl;

  List<RatingEntry> _ratings = [];
  Map<String, dynamic>? _stats;
  List<Map<String, dynamic>> _events = [];

  List<RatingEntry> get ratings => _ratings;
  Map<String, dynamic>? get stats => _stats;
  List<Map<String, dynamic>> get events => _events;

  void setAuthToken(String token) {
    _authToken = token;
  }

  Future<void> loadRatings({String category = 'score'}) async {
    if (_authToken == null) return;
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/ratings?category=$category'),
        headers: {'X-Auth-Token': _authToken!},
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
    if (_authToken == null) return;
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/planets/$planetId/stats'),
        headers: {'X-Auth-Token': _authToken!},
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
    if (_authToken == null) return;
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/planets/$planetId/events'),
        headers: {'X-Auth-Token': _authToken!},
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
    if (_authToken == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/events/resolve'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode({'event_type': eventType}),
      );

      if (response.statusCode == 200) {
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to resolve event: $e');
    }
  }

  void onNotification(Map<String, dynamic>? data) {
    if (data == null) return;
    notifyListeners();
  }
}

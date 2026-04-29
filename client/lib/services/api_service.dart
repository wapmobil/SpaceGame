import 'dart:convert';
import 'package:http/http.dart' as http;

class ApiService {
  final String baseUrl;
  String? authToken;

  ApiService({required this.baseUrl, this.authToken});

  Future<Map<String, dynamic>> registerPlayer(String name) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/register'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'name': name}),
    );

    if (response.statusCode == 201) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    throw Exception('Не удалось зарегистрировать игрока: ${response.statusCode}');
  }

  Future<List<Map<String, dynamic>>> getPlanets() async {
    final response = await http.get(
      Uri.parse('$baseUrl/api/planets'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': authToken ?? '',
      },
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body) as List;
      return List<Map<String, dynamic>>.from(data);
    }
    throw Exception('Не удалось загрузить планеты: ${response.statusCode}');
  }

  Future<Map<String, dynamic>> createPlanet(String name) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/planets/create'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': authToken ?? '',
      },
      body: jsonEncode({'name': name}),
    );

    if (response.statusCode == 201) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    throw Exception('Не удалось создать планету: ${response.statusCode}');
  }

  Future<Map<String, dynamic>> startPlanetSurvey(String planetId, int duration) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/planets/$planetId/planet-survey'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': authToken ?? '',
      },
      body: jsonEncode({'duration': duration}),
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    throw Exception('Failed to start planet survey: ${response.statusCode}');
  }

  Future<Map<String, dynamic>> getPlanetSurvey(String planetId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/api/planets/$planetId/planet-survey'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': authToken ?? '',
      },
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    throw Exception('Failed to get planet survey: ${response.statusCode}');
  }

  Future<List<Map<String, dynamic>>> getLocations(String planetId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/api/planets/$planetId/locations'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': authToken ?? '',
      },
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body) as List;
      return List<Map<String, dynamic>>.from(data);
    }
    throw Exception('Failed to get locations: ${response.statusCode}');
  }

  Future<Map<String, dynamic>> buildOnLocation(String planetId, String locationId, String buildingType) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/planets/$planetId/locations/$locationId/build'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': authToken ?? '',
      },
      body: jsonEncode({'building_type': buildingType}),
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    throw Exception('Failed to build on location: ${response.statusCode}');
  }

  Future<Map<String, dynamic>> removeBuilding(String planetId, String locationId) async {
    final response = await http.delete(
      Uri.parse('$baseUrl/api/planets/$planetId/locations/$locationId/building'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': authToken ?? '',
      },
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    throw Exception('Failed to remove building: ${response.statusCode}');
  }

  Future<Map<String, dynamic>> abandonLocation(String planetId, String locationId) async {
    final response = await http.post(
      Uri.parse('$baseUrl/api/planets/$planetId/locations/$locationId/abandon'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': authToken ?? '',
      },
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    throw Exception('Failed to abandon location: ${response.statusCode}');
  }

  Future<List<Map<String, dynamic>>> getExpeditionHistory(String planetId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/api/planets/$planetId/expedition-history'),
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': authToken ?? '',
      },
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body) as List;
      return List<Map<String, dynamic>>.from(data);
    }
    throw Exception('Failed to get expedition history: ${response.statusCode}');
  }
}

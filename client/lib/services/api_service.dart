import 'dart:convert';
import 'package:http/http.dart' as http;

class ApiService {
  final String baseUrl;
  final String? authToken;

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
    throw Exception('Failed to register player: ${response.statusCode}');
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
    throw Exception('Failed to load planets: ${response.statusCode}');
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
    throw Exception('Failed to create planet: ${response.statusCode}');
  }
}

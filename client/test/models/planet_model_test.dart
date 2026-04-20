import 'package:flutter_test/flutter_test.dart';
import 'package:spacegame/models/planet.dart';

void main() {
  group('Planet model', () {
    test('fromJson creates planet with default values', () {
      final planet = Planet.fromJson({
        'id': 'test-id',
        'player_id': 'player-1',
        'name': 'Test Planet',
        'level': 2,
        'resources': {'food': 10, 'money': 50},
      });

      expect(planet.id, 'test-id');
      expect(planet.playerId, 'player-1');
      expect(planet.name, 'Test Planet');
      expect(planet.level, 2);
      expect(planet.resources['food'], 10);
    });

    test('fromJson uses defaults for missing fields', () {
      final planet = Planet.fromJson({
        'id': 'test-id',
        'player_id': 'player-1',
        'name': 'Test Planet',
      });

      expect(planet.level, 1);
    });

    test('fromJson with empty resources uses defaults', () {
      final planet = Planet.fromJson({
        'id': 'test-id',
        'player_id': 'player-1',
        'name': 'Test Planet',
        'resources': {},
      });

      expect(planet.resources['food'], 0);
      expect(planet.resources['max_energy'], 100);
    });

    test('toJson produces expected format', () {
      final planet = Planet(
        id: 'test-id',
        playerId: 'player-1',
        name: 'Test Planet',
        level: 1,
      );

      final json = planet.toJson();
      expect(json['id'], 'test-id');
      expect(json['player_id'], 'player-1');
      expect(json['name'], 'Test Planet');
      expect(json['level'], 1);
    });
  });
}

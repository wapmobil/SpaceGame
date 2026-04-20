class Planet {
  final String id;
  final String playerId;
  final String name;
  final int level;
  final Map<String, dynamic> resources;
  final DateTime? createdAt;

  Planet({
    required this.id,
    required this.playerId,
    required this.name,
    this.level = 1,
    Map<String, dynamic>? resources,
    this.createdAt,
  }) : resources = {...defaultResources, ...?resources};

  static const Map<String, dynamic> defaultResources = {
    'food': 0,
    'composite': 0,
    'mechanisms': 0,
    'reagents': 0,
    'energy': 0,
    'max_energy': 100,
    'money': 0,
    'alien_tech': 0,
  };

  factory Planet.fromJson(Map<String, dynamic> json) {
    return Planet(
      id: json['id'] as String,
      playerId: json['player_id'] as String,
      name: json['name'] as String,
      level: json['level'] as int? ?? 1,
      resources: json['resources'] != null
          ? Map<String, dynamic>.from(json['resources'] as Map)
          : {},
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'player_id': playerId,
      'name': name,
      'level': level,
      'resources': resources,
    };
  }

  Planet copyWith({
    String? id,
    String? playerId,
    String? name,
    int? level,
    Map<String, dynamic>? resources,
  }) {
    return Planet(
      id: id ?? this.id,
      playerId: playerId ?? this.playerId,
      name: name ?? this.name,
      level: level ?? this.level,
      resources: resources ?? this.resources,
      createdAt: createdAt,
    );
  }
}

class Building {
  final String id;
  final String planetId;
  final String type;
  final int level;
  final double buildProgress;

  Building({
    required this.id,
    required this.planetId,
    required this.type,
    this.level = 1,
    this.buildProgress = 0,
  });

  factory Building.fromJson(Map<String, dynamic> json) {
    return Building(
      id: json['id'] as String,
      planetId: json['planet_id'] as String,
      type: json['type'] as String,
      level: json['level'] as int? ?? 1,
      buildProgress: (json['build_progress'] as num?)?.toDouble() ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'planet_id': planetId,
      'type': type,
      'level': level,
      'build_progress': buildProgress,
    };
  }
}

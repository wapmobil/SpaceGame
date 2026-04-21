class Building {
  final String id;
  final String planetId;
  final String type;
  final int level;
  final double buildProgress;
  final double totalBuildTime;
  final DateTime? createdAt;

  Building({
    this.id = '',
    this.planetId = '',
    required this.type,
    this.level = 0,
    this.buildProgress = 0,
    this.totalBuildTime = 0,
    this.createdAt,
  });

  factory Building.fromJson(Map<String, dynamic> json) {
    final progress = (json['build_progress'] as num?)?.toDouble() ?? 0;
    final totalBuildTime = (json['total_build_time'] as num?)?.toDouble() ?? 0;
    return Building(
      id: json['id'] as String? ?? '',
      planetId: json['planet_id'] as String? ?? '',
      type: json['type'] as String,
      level: json['level'] as int? ?? 0,
      buildProgress: totalBuildTime > 0 ? 1.0 - (progress / totalBuildTime) : 1.0,
      totalBuildTime: totalBuildTime,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : null,
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

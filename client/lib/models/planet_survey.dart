class SurfaceExpedition {
  final String id;
  final String planetId;
  final String status;
  final double progress;
  final double duration;
  final double elapsedTime;
  final String range;
  final DateTime createdAt;
  final DateTime updatedAt;

  SurfaceExpedition({
    required this.id,
    required this.planetId,
    required this.status,
    required this.progress,
    required this.duration,
    required this.elapsedTime,
    required this.range,
    required this.createdAt,
    required this.updatedAt,
  });

  bool get isActive => status == 'active';
  bool get isComplete => status == 'completed';
  bool get isFailed => status == 'failed';
  bool get isDiscovered => status == 'discovered';
  double get remainingTime => duration - elapsedTime;

  factory SurfaceExpedition.fromJson(Map<String, dynamic> json) {
    return SurfaceExpedition(
      id: json['id'] as String,
      planetId: json['planet_id'] as String,
      status: json['status'] as String? ?? 'active',
      progress: (json['progress'] as num?)?.toDouble() ?? 0,
      duration: (json['duration'] as num?)?.toDouble() ?? 0,
      elapsedTime: (json['elapsed_time'] as num?)?.toDouble() ?? 0,
      range: json['range'] as String? ?? '300s',
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now().toUtc(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'] as String)
          : DateTime.now().toUtc(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'planet_id': planetId,
      'status': status,
      'progress': progress,
      'duration': duration,
      'elapsed_time': elapsedTime,
      'range': range,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }
}

class Location {
  final String id;
  final String type;
  final String name;
  final String? buildingType;
  final int buildingLevel;
  final bool buildingActive;
  final String sourceResource;
  final double sourceAmount;
  final double sourceRemaining;
  final bool active;
  final DateTime discoveredAt;

  Location({
    required this.id,
    required this.type,
    required this.name,
    this.buildingType,
    this.buildingLevel = 0,
    this.buildingActive = false,
    required this.sourceResource,
    required this.sourceAmount,
    required this.sourceRemaining,
    required this.active,
    required this.discoveredAt,
  });

  factory Location.fromJson(Map<String, dynamic> json) {
    return Location(
      id: json['id'] as String,
      type: json['type'] as String? ?? 'unknown',
      name: json['name'] as String? ?? '',
      buildingType: json['building_type'] as String?,
      buildingLevel: json['building_level'] as int? ?? 0,
      buildingActive: json['building_active'] as bool? ?? false,
      sourceResource: json['source_resource'] as String? ?? '',
      sourceAmount: (json['source_amount'] as num?)?.toDouble() ?? 0,
      sourceRemaining: (json['source_remaining'] as num?)?.toDouble() ?? 0,
      active: json['active'] as bool? ?? true,
      discoveredAt: json['discovered_at'] != null
          ? DateTime.parse(json['discovered_at'] as String)
          : DateTime.now().toUtc(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'type': type,
      'name': name,
      'building_type': buildingType,
      'building_level': buildingLevel,
      'building_active': buildingActive,
      'source_resource': sourceResource,
      'source_amount': sourceAmount,
      'source_remaining': sourceRemaining,
      'active': active,
      'discovered_at': discoveredAt.toIso8601String(),
    };
  }

  double get depletionPercent {
    if (sourceAmount <= 0) return 100;
    return ((sourceAmount - sourceRemaining) / sourceAmount) * 100;
  }

  bool get isDepleted => sourceRemaining <= 0 && sourceAmount > 0;
}

class LocationBuilding {
  final String id;
  final String locationId;
  final String buildingType;
  final int level;
  final bool active;
  final double buildProgress;
  final double buildTime;

  LocationBuilding({
    required this.id,
    required this.locationId,
    required this.buildingType,
    this.level = 1,
    this.active = false,
    this.buildProgress = 0,
    this.buildTime = 0,
  });

  factory LocationBuilding.fromJson(Map<String, dynamic> json) {
    return LocationBuilding(
      id: json['id'] as String,
      locationId: json['location_id'] as String,
      buildingType: json['building_type'] as String,
      level: json['level'] as int? ?? 1,
      active: json['active'] as bool? ?? false,
      buildProgress: (json['build_progress'] as num?)?.toDouble() ?? 0,
      buildTime: (json['build_time'] as num?)?.toDouble() ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'location_id': locationId,
      'building_type': buildingType,
      'level': level,
      'active': active,
      'build_progress': buildProgress,
      'build_time': buildTime,
    };
  }

  bool get isBuilding => buildTime > 0 && buildProgress > 0;
  double get buildProgressPercent {
    if (buildTime <= 0) return 0;
    return (buildProgress / buildTime).clamp(0.0, 1.0);
  }
}

class ExpeditionHistoryEntry {
  final String id;
  final String status;
  final String result;
  final String discovered;
  final Map<String, double> resourcesGained;
  final DateTime createdAt;
  final DateTime completedAt;

  ExpeditionHistoryEntry({
    required this.id,
    required this.status,
    required this.result,
    required this.discovered,
    required this.resourcesGained,
    required this.createdAt,
    required this.completedAt,
  });

  factory ExpeditionHistoryEntry.fromJson(Map<String, dynamic> json) {
    final resourcesJson = json['resources_gained'] as Map<String, dynamic>? ?? {};
    final resourcesGained = <String, double>{};
    resourcesJson.forEach((key, value) {
      resourcesGained[key] = (value as num).toDouble();
    });

    return ExpeditionHistoryEntry(
      id: json['id'] as String,
      status: json['status'] as String? ?? 'completed',
      result: json['result'] as String? ?? 'failed',
      discovered: json['discovered'] as String? ?? '',
      resourcesGained: resourcesGained,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now().toUtc(),
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : DateTime.now().toUtc(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'status': status,
      'result': result,
      'discovered': discovered,
      'resources_gained': resourcesGained,
      'created_at': createdAt.toIso8601String(),
      'completed_at': completedAt.toIso8601String(),
    };
  }
}

class RangeStats {
  final int totalExpeditions;
  final int locationsFound;

  RangeStats({
    this.totalExpeditions = 0,
    this.locationsFound = 0,
  });

  factory RangeStats.fromJson(Map<String, dynamic> json) {
    return RangeStats(
      totalExpeditions: json['total_expeditions'] as int? ?? 0,
      locationsFound: json['locations_found'] as int? ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'total_expeditions': totalExpeditions,
      'locations_found': locationsFound,
    };
  }
}

class PlanetSurveyState {
  final List<SurfaceExpedition> expeditions;
  final Map<String, RangeStats> rangeStats;
  final int maxDuration;
  final Map<String, double> costPerMin;

  PlanetSurveyState({
    this.expeditions = const [],
    Map<String, RangeStats>? rangeStats,
    this.maxDuration = 300,
    Map<String, double>? costPerMin,
  })  : rangeStats = rangeStats ?? {},
        costPerMin = costPerMin ?? {};

  factory PlanetSurveyState.fromJson(Map<String, dynamic> json) {
    final expeditionsJson = json['expeditions'] as List? ?? [];
    final expeditions = expeditionsJson
        .map((e) => SurfaceExpedition.fromJson(e as Map<String, dynamic>))
        .toList();

    final rangeStatsJson = json['range_stats'] as Map<String, dynamic>? ?? {};
    final rangeStats = <String, RangeStats>{};
    rangeStatsJson.forEach((key, value) {
      rangeStats[key] = RangeStats.fromJson(value as Map<String, dynamic>);
    });

    final costPerMinJson = json['cost_per_min'] as Map<String, dynamic>? ?? {};
    final costPerMin = <String, double>{};
    costPerMinJson.forEach((key, value) {
      costPerMin[key] = (value as num).toDouble();
    });

    return PlanetSurveyState(
      expeditions: expeditions,
      rangeStats: rangeStats,
      maxDuration: json['max_duration'] as int? ?? 300,
      costPerMin: costPerMin,
    );
  }
}

class LocationsResponse {
  final List<Location> locations;

  LocationsResponse({this.locations = const []});

  factory LocationsResponse.fromJson(Map<String, dynamic> json) {
    final locationsJson = json['locations'] as List? ?? json as List;
    return LocationsResponse(
      locations: locationsJson
          .map((e) => Location.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}

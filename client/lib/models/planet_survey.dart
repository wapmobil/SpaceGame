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

import 'package:spacegame/models/building.dart';

class Planet {
  final String id;
  final String playerId;
  final String name;
  final int level;
  final Map<String, dynamic> resources;
  final DateTime? createdAt;

  // Energy buffer fields
  final double? energyBufferValue;
  final double? energyBufferMax;
  final bool? energyBufferDeficit;

  // Base operational flags
  final bool? baseOperational;
  final bool? canResearch;
  final bool? canExpedition;
  
  // Building list from API
  final List<Building>? buildings;

  // Production info
  final double? productionFood;
  final double? productionComposite;
  final double? productionMechanisms;
  final double? productionReagents;
  final double? productionEnergy;
  final double? productionMoney;
  final double? productionAlienTech;

  // Energy balance
  final double? energyBalanceProduction;
  final double? energyBalanceConsumption;
  final double? energyBalanceNet;

  // Construction limits
  final int? activeConstructions;
  final int? maxConstructions;

  // Storage
  final double? storageCapacity;

  // Planet survey fields
  final String? resourceType;
  final bool canStartPlanetSurvey;
  final bool canStartSpaceExpedition;
  final int? baseLevel;
  final int? commandCenterLevel;
  final int maxLocations;

  Planet({
    required this.id,
    required this.playerId,
    required this.name,
    this.level = 1,
    Map<String, dynamic>? resources,
    this.createdAt,
    this.energyBufferValue,
    this.energyBufferMax,
    this.energyBufferDeficit,
    this.baseOperational,
    this.canResearch,
    this.canExpedition,

    this.buildings,
    this.productionFood,
    this.productionComposite,
    this.productionMechanisms,
    this.productionReagents,
    this.productionEnergy,
    this.productionMoney,
    this.productionAlienTech,
    this.energyBalanceProduction,
    this.energyBalanceConsumption,
    this.energyBalanceNet,
    this.activeConstructions,
    this.maxConstructions,
    this.storageCapacity,
    this.resourceType,
    this.canStartPlanetSurvey = false,
    this.canStartSpaceExpedition = false,
    this.baseLevel,
    this.commandCenterLevel,
    this.maxLocations = 1,
  }) : resources = {...defaultResources, ...?resources};

  static const Map<String, dynamic> defaultResources = {
    'food': 0,
    'iron': 0,
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
      energyBufferValue: (json['energy_buffer_value'] as num?)?.toDouble(),
      energyBufferMax: (json['energy_buffer_max'] as num?)?.toDouble(),
      energyBufferDeficit: json['energy_buffer_deficit'] as bool?,
      baseOperational: json['base_operational'] as bool?,
      canResearch: json['can_research'] as bool?,
      canExpedition: json['can_expedition'] as bool?,
            buildings: json['buildings'] != null
          ? List<Building>.from((json['buildings'] as List).map((b) => Building.fromJson(b as Map<String, dynamic>)))
          : null,
      productionFood: (json['production_food'] as num?)?.toDouble(),
      productionComposite: (json['production_composite'] as num?)?.toDouble(),
      productionMechanisms: (json['production_mechanisms'] as num?)?.toDouble(),
      productionReagents: (json['production_reagents'] as num?)?.toDouble(),
      productionEnergy: (json['production_energy'] as num?)?.toDouble(),
      productionMoney: (json['production_money'] as num?)?.toDouble(),
      productionAlienTech: (json['production_alien_tech'] as num?)?.toDouble(),
      energyBalanceProduction: (json['energy_balance_production'] as num?)?.toDouble(),
      energyBalanceConsumption: (json['energy_balance_consumption'] as num?)?.toDouble(),
      energyBalanceNet: (json['energy_balance_net'] as num?)?.toDouble(),
      activeConstructions: json['active_constructions'] as int?,
      maxConstructions: json['max_constructions'] as int?,
      storageCapacity: (json['storage_capacity'] as num?)?.toDouble(),
      resourceType: json['resource_type'] as String?,
      canStartPlanetSurvey: json['can_start_planet_survey'] as bool? ?? false,
      canStartSpaceExpedition: json['can_start_space_expedition'] as bool? ?? false,
      baseLevel: json['base_level'] as int?,
      commandCenterLevel: json['command_center_level'] as int?,
      maxLocations: json['max_locations'] as int? ?? 1,
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
    List<Building>? buildings,
    double? productionFood,
    double? productionComposite,
    double? productionMechanisms,
    double? productionReagents,
    double? productionEnergy,
    double? productionMoney,
    double? productionAlienTech,
    double? energyBalanceProduction,
    double? energyBalanceConsumption,
    double? energyBalanceNet,
    double? energyBufferValue,
    double? energyBufferMax,
    bool? energyBufferDeficit,
    bool? baseOperational,
    bool? canResearch,
    bool? canExpedition,
        int? activeConstructions,
    int? maxConstructions,
    double? storageCapacity,
    String? resourceType,
    bool? canStartPlanetSurvey,
    bool? canStartSpaceExpedition,
    int? baseLevel,
    int? commandCenterLevel,
    int? maxLocations,
  }) {
    return Planet(
      id: id ?? this.id,
      playerId: playerId ?? this.playerId,
      name: name ?? this.name,
      level: level ?? this.level,
      resources: resources ?? this.resources,
      createdAt: createdAt,
      energyBufferValue: energyBufferValue ?? this.energyBufferValue,
      energyBufferMax: energyBufferMax ?? this.energyBufferMax,
      energyBufferDeficit: energyBufferDeficit ?? this.energyBufferDeficit,
      baseOperational: baseOperational ?? this.baseOperational,
      canResearch: canResearch ?? this.canResearch,
      canExpedition: canExpedition ?? this.canExpedition,
            buildings: buildings ?? this.buildings,
      productionFood: productionFood ?? this.productionFood,
      productionComposite: productionComposite ?? this.productionComposite,
      productionMechanisms: productionMechanisms ?? this.productionMechanisms,
      productionReagents: productionReagents ?? this.productionReagents,
      productionEnergy: productionEnergy ?? this.productionEnergy,
      productionMoney: productionMoney ?? this.productionMoney,
      productionAlienTech: productionAlienTech ?? this.productionAlienTech,
      energyBalanceProduction: energyBalanceProduction ?? this.energyBalanceProduction,
      energyBalanceConsumption: energyBalanceConsumption ?? this.energyBalanceConsumption,
      energyBalanceNet: energyBalanceNet ?? this.energyBalanceNet,
      activeConstructions: activeConstructions ?? this.activeConstructions,
      maxConstructions: maxConstructions ?? this.maxConstructions,
      storageCapacity: storageCapacity ?? this.storageCapacity,
      resourceType: resourceType ?? this.resourceType,
      canStartPlanetSurvey: canStartPlanetSurvey ?? this.canStartPlanetSurvey,
      canStartSpaceExpedition: canStartSpaceExpedition ?? this.canStartSpaceExpedition,
      baseLevel: baseLevel ?? this.baseLevel,
      commandCenterLevel: commandCenterLevel ?? this.commandCenterLevel,
      maxLocations: maxLocations ?? this.maxLocations,
    );
  }
}

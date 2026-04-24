class Building {
  final String type;
  final int level;
  final double buildProgress; // -1 = working, 0..buildTime = under construction, 0 = ready for confirmation
  final bool enabled;
  final double buildTime; // total build time in seconds
  final double costFood;
  final double costIron;
  final double costMoney;
  final double nextCostFood;
  final double nextCostIron;
  final double nextCostMoney;
  final double productionFood;
  final double productionIron;
  final double productionComposite;
  final double productionMechanisms;
  final double productionReagents;
  final double productionEnergy;
  final double productionMoney;
  final double productionAlienTech;
  final double consumption;
  final double nextProductionFood;
  final double nextProductionIron;
  final double nextProductionComposite;
  final double nextProductionMechanisms;
  final double nextProductionReagents;
  final double nextProductionEnergy;
  final double nextProductionMoney;
  final double nextProductionAlienTech;
  final double deltaFood;
  final double deltaIron;
  final double deltaComposite;
  final double deltaMechanisms;
  final double deltaReagents;
  final double deltaEnergy;
  final double deltaMoney;
  final double deltaAlienTech;

  Building({
    required this.type,
    this.level = 0,
    this.buildProgress = 0,
    this.enabled = true,
    this.buildTime = 0,
    this.costFood = 0,
    this.costIron = 0,
    this.costMoney = 0,
    this.nextCostFood = 0,
    this.nextCostIron = 0,
    this.nextCostMoney = 0,
    this.productionFood = 0,
    this.productionIron = 0,
    this.productionComposite = 0,
    this.productionMechanisms = 0,
    this.productionReagents = 0,
    this.productionEnergy = 0,
    this.productionMoney = 0,
    this.productionAlienTech = 0,
    this.consumption = 0,
    this.nextProductionFood = 0,
    this.nextProductionIron = 0,
    this.nextProductionComposite = 0,
    this.nextProductionMechanisms = 0,
    this.nextProductionReagents = 0,
    this.nextProductionEnergy = 0,
    this.nextProductionMoney = 0,
    this.nextProductionAlienTech = 0,
    this.deltaFood = 0,
    this.deltaIron = 0,
    this.deltaComposite = 0,
    this.deltaMechanisms = 0,
    this.deltaReagents = 0,
    this.deltaEnergy = 0,
    this.deltaMoney = 0,
    this.deltaAlienTech = 0,
  });

  bool get isBuilding => buildTime > 0 && buildProgress > 0;
  bool get isBuildComplete => buildTime > 0 && buildProgress == 0;
  bool get isWorking => !isBuilding && !isBuildComplete && enabled && level > 0;

  factory Building.fromJson(Map<String, dynamic> json) {
    return Building(
      type: json['type'] as String? ?? '',
      level: json['level'] as int? ?? 0,
      buildProgress: (json['build_progress'] as num?)?.toDouble() ?? 0,
      enabled: json['enabled'] as bool? ?? true,
      buildTime: (json['build_time'] as num?)?.toDouble() ?? 0,
      costFood: (json['cost']?['food'] as num?)?.toDouble() ?? 0,
      costIron: (json['cost']?['iron'] as num?)?.toDouble() ?? 0,
      costMoney: (json['cost']?['money'] as num?)?.toDouble() ?? 0,
      nextCostFood: (json['next_cost']?['food'] as num?)?.toDouble() ?? 0,
      nextCostIron: (json['next_cost']?['iron'] as num?)?.toDouble() ?? 0,
      nextCostMoney: (json['next_cost']?['money'] as num?)?.toDouble() ?? 0,
      productionFood: (json['production']?['food'] as num?)?.toDouble() ?? 0,
      productionIron: (json['production']?['iron'] as num?)?.toDouble() ?? 0,
      productionComposite: (json['production']?['composite'] as num?)?.toDouble() ?? 0,
      productionMechanisms: (json['production']?['mechanisms'] as num?)?.toDouble() ?? 0,
      productionReagents: (json['production']?['reagents'] as num?)?.toDouble() ?? 0,
      productionEnergy: (json['production']?['energy'] as num?)?.toDouble() ?? 0,
      productionMoney: (json['production']?['money'] as num?)?.toDouble() ?? 0,
      productionAlienTech: (json['production']?['alien_tech'] as num?)?.toDouble() ?? 0,
      consumption: (json['consumption'] as num?)?.toDouble() ?? 0,
      nextProductionFood: (json['next_production']?['food'] as num?)?.toDouble() ?? 0,
      nextProductionIron: (json['next_production']?['iron'] as num?)?.toDouble() ?? 0,
      nextProductionComposite: (json['next_production']?['composite'] as num?)?.toDouble() ?? 0,
      nextProductionMechanisms: (json['next_production']?['mechanisms'] as num?)?.toDouble() ?? 0,
      nextProductionReagents: (json['next_production']?['reagents'] as num?)?.toDouble() ?? 0,
      nextProductionEnergy: (json['next_production']?['energy'] as num?)?.toDouble() ?? 0,
      nextProductionMoney: (json['next_production']?['money'] as num?)?.toDouble() ?? 0,
      nextProductionAlienTech: (json['next_production']?['alien_tech'] as num?)?.toDouble() ?? 0,
      deltaFood: (json['deltas']?['food'] as num?)?.toDouble() ?? 0,
      deltaIron: (json['deltas']?['iron'] as num?)?.toDouble() ?? 0,
      deltaComposite: (json['deltas']?['composite'] as num?)?.toDouble() ?? 0,
      deltaMechanisms: (json['deltas']?['mechanisms'] as num?)?.toDouble() ?? 0,
      deltaReagents: (json['deltas']?['reagents'] as num?)?.toDouble() ?? 0,
      deltaEnergy: (json['deltas']?['energy'] as num?)?.toDouble() ?? 0,
      deltaMoney: (json['deltas']?['money'] as num?)?.toDouble() ?? 0,
      deltaAlienTech: (json['deltas']?['alien_tech'] as num?)?.toDouble() ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'level': level,
      'build_progress': buildProgress,
      'enabled': enabled,
      'build_time': buildTime,
      'cost': {'food': costFood, 'iron': costIron, 'money': costMoney},
      'next_cost': {'food': nextCostFood, 'iron': nextCostIron, 'money': nextCostMoney},
      'production': {
        'food': productionFood,
        'iron': productionIron,
        'composite': productionComposite,
        'mechanisms': productionMechanisms,
        'reagents': productionReagents,
        'energy': productionEnergy,
        'money': productionMoney,
        'alien_tech': productionAlienTech,
      },
      'consumption': consumption,
      'next_production': {
        'food': nextProductionFood,
        'iron': nextProductionIron,
        'composite': nextProductionComposite,
        'mechanisms': nextProductionMechanisms,
        'reagents': nextProductionReagents,
        'energy': nextProductionEnergy,
        'money': nextProductionMoney,
        'alien_tech': nextProductionAlienTech,
      },
    };
  }
}

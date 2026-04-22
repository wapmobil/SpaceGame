class Building {
  final String type;
  final int level;
  final double buildProgress; // remaining seconds
  final bool pending;
  final bool enabled;
  final double buildTime; // total build time in seconds
  final double costFood;
  final double costMoney;
  final double nextCostFood;
  final double nextCostMoney;
  final double productionFood;
  final double productionComposite;
  final double productionMechanisms;
  final double productionReagents;
  final double productionEnergy;
  final double productionMoney;
  final double productionAlienTech;
  final double consumption;

  Building({
    required this.type,
    this.level = 0,
    this.buildProgress = 0,
    this.pending = false,
    this.enabled = true,
    this.buildTime = 0,
    this.costFood = 0,
    this.costMoney = 0,
    this.nextCostFood = 0,
    this.nextCostMoney = 0,
    this.productionFood = 0,
    this.productionComposite = 0,
    this.productionMechanisms = 0,
    this.productionReagents = 0,
    this.productionEnergy = 0,
    this.productionMoney = 0,
    this.productionAlienTech = 0,
    this.consumption = 0,
  });

  factory Building.fromJson(Map<String, dynamic> json) {
    return Building(
      type: json['type'] as String? ?? '',
      level: json['level'] as int? ?? 0,
      buildProgress: (json['build_progress'] as num?)?.toDouble() ?? 0,
      pending: json['pending'] as bool? ?? false,
      enabled: json['enabled'] as bool? ?? true,
      buildTime: (json['build_time'] as num?)?.toDouble() ?? 0,
      costFood: (json['cost']?['food'] as num?)?.toDouble() ?? 0,
      costMoney: (json['cost']?['money'] as num?)?.toDouble() ?? 0,
      nextCostFood: (json['next_cost']?['food'] as num?)?.toDouble() ?? 0,
      nextCostMoney: (json['next_cost']?['money'] as num?)?.toDouble() ?? 0,
      productionFood: (json['production']?['food'] as num?)?.toDouble() ?? 0,
      productionComposite: (json['production']?['composite'] as num?)?.toDouble() ?? 0,
      productionMechanisms: (json['production']?['mechanisms'] as num?)?.toDouble() ?? 0,
      productionReagents: (json['production']?['reagents'] as num?)?.toDouble() ?? 0,
      productionEnergy: (json['production']?['energy'] as num?)?.toDouble() ?? 0,
      productionMoney: (json['production']?['money'] as num?)?.toDouble() ?? 0,
      productionAlienTech: (json['production']?['alien_tech'] as num?)?.toDouble() ?? 0,
      consumption: (json['consumption'] as num?)?.toDouble() ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'level': level,
      'build_progress': buildProgress,
      'pending': pending,
      'enabled': enabled,
      'build_time': buildTime,
      'cost': {'food': costFood, 'money': costMoney},
      'next_cost': {'food': nextCostFood, 'money': nextCostMoney},
      'production': {
        'food': productionFood,
        'composite': productionComposite,
        'mechanisms': productionMechanisms,
        'reagents': productionReagents,
        'energy': productionEnergy,
        'money': productionMoney,
        'alien_tech': productionAlienTech,
      },
      'consumption': consumption,
    };
  }
}

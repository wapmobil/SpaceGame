class GardenBedRow {
  final String status; // "empty", "planted", "mature", "withered"
  final String? plantType;
  final int? stage;
  final int weeds;
  final int waterTimer;
  final int lastTick;
  final int witherTimer;
  final int ticksToMature;
  final int stageProgress;

  GardenBedRow({
    required this.status,
    this.plantType,
    this.stage = 0,
    this.weeds = 0,
    this.waterTimer = 0,
    this.lastTick = 0,
    this.witherTimer = 0,
    this.ticksToMature = 0,
    this.stageProgress = 0,
  });

  factory GardenBedRow.fromJson(Map<String, dynamic> json) {
    String status = json['status'] as String? ?? 'empty';
    if (status.isEmpty) status = 'empty';
    return GardenBedRow(
      status: status,
      plantType: json['plant_type'] as String?,
      stage: json['stage'] as int? ?? 0,
      weeds: json['weeds'] as int? ?? 0,
      waterTimer: json['water_timer'] as int? ?? 0,
      lastTick: json['last_tick'] as int? ?? 0,
      witherTimer: json['wither_timer'] as int? ?? 0,
      ticksToMature: json['ticks_to_mature'] as int? ?? 0,
      stageProgress: json['stage_progress'] as int? ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'status': status,
      if (plantType != null) 'plant_type': plantType,
      if (stage != null) 'stage': stage,
      'weeds': weeds,
      'water_timer': waterTimer,
      'last_tick': lastTick,
      if (witherTimer > 0) 'wither_timer': witherTimer,
      if (ticksToMature > 0) 'ticks_to_mature': ticksToMature,
    };
  }

  bool get isEmpty => status == 'empty';
  bool get isPlanted => status == 'planted';
  bool get isMature => status == 'mature';
  bool get isWithered => status == 'withered';
  bool get isWeedyEmpty => isEmpty && weeds > 0;
  bool get isGrowing => isPlanted && !isMature;
  bool get isWeedy => weeds > 0;
  bool get isAtMaxWeeds => weeds >= 3;
  bool get isWatered => waterTimer > 0;

  GardenBedRow copyWith({
    String? status,
    String? plantType,
    int? stage,
    int? weeds,
    int? waterTimer,
    int? lastTick,
    int? witherTimer,
    int? stageProgress,
  }) {
    return GardenBedRow(
      status: status ?? this.status,
      plantType: plantType ?? this.plantType,
      stage: stage ?? this.stage,
      weeds: weeds ?? this.weeds,
      waterTimer: waterTimer ?? this.waterTimer,
      lastTick: lastTick ?? this.lastTick,
      witherTimer: witherTimer ?? this.witherTimer,
      stageProgress: stageProgress ?? this.stageProgress,
    );
  }
}

class GardenBedState {
  final List<GardenBedRow> rows;
  final int lastTick;
  final int rowCount;

  GardenBedState({
    required this.rows,
    this.lastTick = 0,
    this.rowCount = 0,
  });

  factory GardenBedState.fromJson(Map<String, dynamic> json) {
    final rowsData = json['rows'] as List? ?? [];
    final rows = rowsData.map((r) => GardenBedRow.fromJson(r as Map<String, dynamic>)).toList();

    return GardenBedState(
      rows: rows,
      lastTick: json['last_tick'] as int? ?? 0,
      rowCount: json['row_count'] as int? ?? rows.length,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'rows': rows.map((r) => r.toJson()).toList(),
      'last_tick': lastTick,
      'row_count': rowCount,
    };
  }
}

class GardenBedPlant {
  final String type;
  final String name;
  final String icon;
  final double seedCost;
  final double moneyReward;
  final double foodReward;
  final int unlockLevel;
  final List<String> stageNames;
  final double weedCost;
  final double waterCost;
  final int growthTicks;

  GardenBedPlant({
    required this.type,
    required this.name,
    required this.icon,
    required this.seedCost,
    required this.moneyReward,
    required this.foodReward,
    required this.unlockLevel,
    required this.stageNames,
    required this.weedCost,
    required this.waterCost,
    required this.growthTicks,
  });

  factory GardenBedPlant.fromJson(Map<String, dynamic> json) {
    return GardenBedPlant(
      type: json['type'] as String? ?? '',
      name: json['name'] as String? ?? '',
      icon: json['icon'] as String? ?? '',
      seedCost: (json['seed_cost'] as num?)?.toDouble() ?? 0,
      moneyReward: (json['money_reward'] as num?)?.toDouble() ?? 0,
      foodReward: (json['food_reward'] as num?)?.toDouble() ?? 0,
      unlockLevel: json['unlock_level'] as int? ?? 1,
      stageNames: (json['stage_names'] as List?)?.map((s) => s as String).toList() ?? ['Семя', 'Росток', 'Созрело'],
      weedCost: (json['weed_cost'] as num?)?.toDouble() ?? 0,
      waterCost: (json['water_cost'] as num?)?.toDouble() ?? 0,
      growthTicks: json['growth_ticks'] as int? ?? 36,
    );
  }

  String get currentStageName {
    if (stageNames.isEmpty) return '';
    return stageNames[0];
  }
}

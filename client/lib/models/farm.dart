class FarmRow {
  final String status; // "empty", "planted", "mature"
  final String? plantType;
  final int? stage;
  final int weeds;
  final int waterTimer;
  final int lastTick;

  FarmRow({
    required this.status,
    this.plantType,
    this.stage = 0,
    this.weeds = 0,
    this.waterTimer = 0,
    this.lastTick = 0,
  });

  factory FarmRow.fromJson(Map<String, dynamic> json) {
    return FarmRow(
      status: json['status'] as String? ?? 'empty',
      plantType: json['plant_type'] as String?,
      stage: json['stage'] as int? ?? 0,
      weeds: json['weeds'] as int? ?? 0,
      waterTimer: json['water_timer'] as int? ?? 0,
      lastTick: json['last_tick'] as int? ?? 0,
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
    };
  }

  bool get isEmpty => status == 'empty';
  bool get isPlanted => status == 'planted';
  bool get isMature => status == 'mature';
  bool get isGrowing => isPlanted && !isMature;
  bool get isWeedy => weeds > 0;
  bool get isAtMaxWeeds => weeds >= 3;
  bool get isWatered => waterTimer > 0;

  FarmRow copyWith({
    String? status,
    String? plantType,
    int? stage,
    int? weeds,
    int? waterTimer,
    int? lastTick,
  }) {
    return FarmRow(
      status: status ?? this.status,
      plantType: plantType ?? this.plantType,
      stage: stage ?? this.stage,
      weeds: weeds ?? this.weeds,
      waterTimer: waterTimer ?? this.waterTimer,
      lastTick: lastTick ?? this.lastTick,
    );
  }
}

class FarmState {
  final List<FarmRow> rows;
  final int lastTick;
  final int rowCount;

  FarmState({
    required this.rows,
    this.lastTick = 0,
    this.rowCount = 0,
  });

  factory FarmState.fromJson(Map<String, dynamic> json) {
    final rowsData = json['rows'] as List? ?? [];
    final rows = rowsData.map((r) => FarmRow.fromJson(r as Map<String, dynamic>)).toList();

    return FarmState(
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

class FarmPlant {
  final String type;
  final String name;
  final String icon;
  final double foodReward;
  final List<String> stageNames;

  FarmPlant({
    required this.type,
    required this.name,
    required this.icon,
    required this.foodReward,
    required this.stageNames,
  });

  factory FarmPlant.fromJson(Map<String, dynamic> json) {
    return FarmPlant(
      type: json['type'] as String? ?? '',
      name: json['name'] as String? ?? '',
      icon: json['icon'] as String? ?? '',
      foodReward: (json['food_reward'] as num?)?.toDouble() ?? 0,
      stageNames: (json['stage_names'] as List?)?.map((s) => s as String).toList() ?? ['Семя', 'Росток', 'Созрело'],
    );
  }

  String get currentStageName {
    if (stageNames.isEmpty) return '';
    return stageNames[0];
  }
}

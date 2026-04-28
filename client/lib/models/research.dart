class ResearchTech {
  final String techId;
  final String name;
  final String description;
  final double costFood;
  final double costMoney;
  final double costAlienTech;
  final double buildTime;
  final int maxLevel;
  final List<String> dependsOn;
  final int level;
  final bool completed;
  final bool inProgress;
  final double progress;
  final double totalTime;
  final double progressPct;

  ResearchTech({
    required this.techId,
    required this.name,
    required this.description,
    this.costFood = 0,
    this.costMoney = 0,
    this.costAlienTech = 0,
    this.buildTime = 0,
    this.maxLevel = 1,
    this.dependsOn = const [],
    this.level = 0,
    this.completed = false,
    this.inProgress = false,
    this.progress = 0,
    this.totalTime = 0,
    this.progressPct = 0,
  });

  factory ResearchTech.fromJson(Map<String, dynamic> json) {
    return ResearchTech(
      techId: json['tech_id'] as String,
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      costFood: (json['cost_food'] as num?)?.toDouble() ?? 0,
      costMoney: (json['cost_money'] as num?)?.toDouble() ?? 0,
      costAlienTech: (json['cost_alien_tech'] as num?)?.toDouble() ?? 0,
      buildTime: (json['build_time'] as num?)?.toDouble() ?? 0,
      maxLevel: json['max_level'] as int? ?? 1,
      dependsOn: (json['depends_on'] as List?)
              ?.map((e) => e as String)
              .toList() ??
          [],
      level: json['level'] as int? ?? 0,
      completed: json['completed'] as bool? ?? false,
      inProgress: json['in_progress'] as bool? ?? false,
      progress: (json['progress'] as num?)?.toDouble() ?? 0,
      totalTime: (json['total_time'] as num?)?.toDouble() ?? 0,
      progressPct: (json['progress_pct'] as num?)?.toDouble() ?? 0,
    );
  }

  factory ResearchTech.fromMap(Map<String, dynamic> map) {
    return ResearchTech(
      techId: map['tech_id'] as String,
      name: map['name'] as String? ?? '',
      description: map['description'] as String? ?? '',
      costFood: (map['cost_food'] as num?)?.toDouble() ?? 0,
      costMoney: (map['cost_money'] as num?)?.toDouble() ?? 0,
      costAlienTech: (map['cost_alien_tech'] as num?)?.toDouble() ?? 0,
      buildTime: (map['build_time'] as num?)?.toDouble() ?? 0,
      maxLevel: map['max_level'] as int? ?? 1,
      dependsOn: (map['depends_on'] as List?)
              ?.map((e) => e as String)
              .toList() ??
          [],
      level: 0,
      completed: false,
      inProgress: false,
      progress: 0,
      totalTime: 0,
      progressPct: 0,
    );
  }

  double get remainingTime => totalTime - progress;

  bool get isUnlocked => dependsOn.isEmpty;
  bool get canResearch => !completed && !inProgress && isUnlocked;

  ResearchTech copyWith({
    String? techId,
    String? name,
    String? description,
    double? costFood,
    double? costMoney,
    double? costAlienTech,
    double? buildTime,
    int? maxLevel,
    List<String>? dependsOn,
    int? level,
    bool? completed,
    bool? inProgress,
    double? progress,
    double? totalTime,
    double? progressPct,
  }) {
    return ResearchTech(
      techId: techId ?? this.techId,
      name: name ?? this.name,
      description: description ?? this.description,
      costFood: costFood ?? this.costFood,
      costMoney: costMoney ?? this.costMoney,
      costAlienTech: costAlienTech ?? this.costAlienTech,
      buildTime: buildTime ?? this.buildTime,
      maxLevel: maxLevel ?? this.maxLevel,
      dependsOn: dependsOn ?? this.dependsOn,
      level: level ?? this.level,
      completed: completed ?? this.completed,
      inProgress: inProgress ?? this.inProgress,
      progress: progress ?? this.progress,
      totalTime: totalTime ?? this.totalTime,
      progressPct: progressPct ?? this.progressPct,
    );
  }
}

class ResearchState {
  final List<ResearchTech> research;
  final List<ResearchTech> available;

  ResearchState({this.research = const [], this.available = const []});

  factory ResearchState.fromJson(Map<String, dynamic> json) {
    final researchList =
        (json['research'] as List?)?.map((e) => ResearchTech.fromJson(e as Map<String, dynamic>)).toList() ?? [];
    final availableList =
        (json['available'] as List?)?.map((e) => ResearchTech.fromJson(e as Map<String, dynamic>)).toList() ?? [];
    return ResearchState(research: researchList, available: availableList);
  }
}

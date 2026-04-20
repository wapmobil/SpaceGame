class Expedition {
  final String id;
  final String planetId;
  final String target;
  final double progress;
  final String status;
  final String expeditionType;
  final double duration;
  final double elapsedTime;
  final Map<String, int> fleetShips;
  final int fleetTotal;
  final double fleetCargo;
  final double fleetEnergy;
  final double fleetDamage;
  final NPCPlanet? discoveredNPC;
  final List<ExpeditionAction> actions;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  Expedition({
    required this.id,
    required this.planetId,
    required this.target,
    this.progress = 0,
    this.status = 'queued',
    required this.expeditionType,
    this.duration = 3600,
    this.elapsedTime = 0,
    this.fleetShips = const {},
    this.fleetTotal = 0,
    this.fleetCargo = 0,
    this.fleetEnergy = 0,
    this.fleetDamage = 0,
    this.discoveredNPC,
    this.actions = const [],
    this.createdAt,
    this.updatedAt,
  });

  factory Expedition.fromJson(Map<String, dynamic> json) {
    return Expedition(
      id: json['id'] as String,
      planetId: json['planet_id'] as String,
      target: json['target'] as String,
      progress: (json['progress'] as num?)?.toDouble() ?? 0,
      status: json['status'] as String? ?? 'queued',
      expeditionType: json['expedition_type'] as String? ?? 'exploration',
      duration: (json['duration'] as num?)?.toDouble() ?? 3600,
      elapsedTime: (json['elapsed_time'] as num?)?.toDouble() ?? 0,
      fleetShips: (json['fleet_ships'] as Map?)
              ?.map((k, v) => MapEntry(k as String, v as int)) ??
          {},
      fleetTotal: json['fleet_total'] as int? ?? 0,
      fleetCargo: (json['fleet_cargo'] as num?)?.toDouble() ?? 0,
      fleetEnergy: (json['fleet_energy'] as num?)?.toDouble() ?? 0,
      fleetDamage: (json['fleet_damage'] as num?)?.toDouble() ?? 0,
      discoveredNPC: json['discovered_npc'] != null
          ? NPCPlanet.fromJson(json['discovered_npc'] as Map<String, dynamic>)
          : null,
      actions: (json['actions'] as List?)
              ?.map((e) => ExpeditionAction.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : null,
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'] as String)
          : null,
    );
  }

  bool get isActive => status == 'active' || status == 'in_progress';
  bool get isComplete => status == 'completed';
  bool get canAct => isActive && actions.isNotEmpty;
  double get remainingProgress => duration - elapsedTime;
}

class ExpeditionAction {
  final String id;
  final String type;
  final String label;
  final String? required;

  ExpeditionAction({
    required this.id,
    required this.type,
    required this.label,
    this.required,
  });

  factory ExpeditionAction.fromJson(Map<String, dynamic> json) {
    return ExpeditionAction(
      id: json['id'] as String,
      type: json['type'] as String,
      label: json['label'] as String,
      required: json['required'] as String?,
    );
  }
}

class NPCPlanet {
  final String id;
  final String name;
  final String type;
  final Map<String, double> resources;
  final double totalResources;
  final bool hasCombat;
  final double fleetStrength;
  final Map<String, dynamic>? enemyFleet;

  NPCPlanet({
    required this.id,
    required this.name,
    required this.type,
    this.resources = const {},
    this.totalResources = 0,
    this.hasCombat = false,
    this.fleetStrength = 0,
    this.enemyFleet,
  });

  factory NPCPlanet.fromJson(Map<String, dynamic> json) {
    return NPCPlanet(
      id: json['id'] as String,
      name: json['name'] as String,
      type: json['type'] as String,
      resources: (json['resources'] as Map?)
              ?.map((k, v) => MapEntry(k as String, (v as num).toDouble())) ??
          {},
      totalResources: (json['total_resources'] as num?)?.toDouble() ?? 0,
      hasCombat: json['has_combat'] as bool? ?? false,
      fleetStrength: (json['fleet_strength'] as num?)?.toDouble() ?? 0,
      enemyFleet: json['enemy_fleet'] as Map<String, dynamic>?,
    );
  }
}

class ExpeditionsListResponse {
  final List<Expedition> expeditions;
  final int activeCount;
  final int maxExpeditions;
  final bool canStartNew;
  final bool expeditionsUnlocked;

  ExpeditionsListResponse({
    this.expeditions = const [],
    this.activeCount = 0,
    this.maxExpeditions = 1,
    this.canStartNew = true,
    this.expeditionsUnlocked = false,
  });

  factory ExpeditionsListResponse.fromJson(Map<String, dynamic> json) {
    return ExpeditionsListResponse(
      expeditions: (json['expeditions'] as List?)
              ?.map((e) => Expedition.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      activeCount: json['active_count'] as int? ?? 0,
      maxExpeditions: json['max_expeditions'] as int? ?? 1,
      canStartNew: json['can_start_new'] as bool? ?? true,
      expeditionsUnlocked: json['expeditions_unlocked'] as bool? ?? false,
    );
  }
}

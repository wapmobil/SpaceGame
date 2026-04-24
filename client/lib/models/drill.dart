class DrillResource {
  final String type;
  final String name;
  final String icon;
  final double amount;
  final double value;

  DrillResource({
    required this.type,
    required this.name,
    required this.icon,
    required this.amount,
    required this.value,
  });

  factory DrillResource.fromJson(Map<String, dynamic> json) {
    return DrillResource(
      type: json['type'] as String? ?? '',
      name: json['name'] as String? ?? '',
      icon: json['icon'] as String? ?? '',
      amount: (json['amount'] as num?)?.toDouble() ?? 0,
      value: (json['value'] as num?)?.toDouble() ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'name': name,
      'icon': icon,
      'amount': amount,
      'value': value,
    };
  }
}

class DrillCell {
  final int x;
  final int y;
  final String cellType;
  final String? resourceType;
  final double resourceAmount;
  final double resourceValue;
  final bool extracted;

  DrillCell({
    required this.x,
    required this.y,
    required this.cellType,
    this.resourceType,
    this.resourceAmount = 0,
    this.resourceValue = 0,
    this.extracted = false,
  });

  factory DrillCell.fromJson(Map<String, dynamic> json) {
    return DrillCell(
      x: json['x'] as int? ?? 0,
      y: json['y'] as int? ?? 0,
      cellType: json['cell_type'] as String? ?? 'empty',
      resourceType: json['resource_type'] as String?,
      resourceAmount: (json['resource_amount'] as num?)?.toDouble() ?? 0,
      resourceValue: (json['resource_value'] as num?)?.toDouble() ?? 0,
      extracted: json['extracted'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'x': x,
      'y': y,
      'cell_type': cellType,
      'resource_type': resourceType,
      'resource_amount': resourceAmount,
      'resource_value': resourceValue,
      'extracted': extracted,
    };
  }
}

class DrillHitResource {
  final String type;
  final String name;
  final String icon;
  final double amount;
  final double value;

  DrillHitResource({
    required this.type,
    required this.name,
    required this.icon,
    required this.amount,
    required this.value,
  });

  factory DrillHitResource.fromJson(Map<String, dynamic> json) {
    return DrillHitResource(
      type: json['type'] as String? ?? '',
      name: json['name'] as String? ?? '',
      icon: json['icon'] as String? ?? '',
      amount: (json['amount'] as num?)?.toDouble() ?? 0,
      value: (json['value'] as num?)?.toDouble() ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'name': name,
      'icon': icon,
      'amount': amount,
      'value': value,
    };
  }
}

class DrillState {
  final String sessionId;
  final String planetId;
  final int drillHp;
  final int drillMaxHp;
  final int depth;
  final int drillX;
  final int worldWidth;
  final List<List<DrillCell>> world;
  final List<DrillResource> resources;
  final String status;
  final double totalEarned;
  final String createdAt;
  final String? completedAt;
  final int? seed;
  final String? pendingDirection;
  final bool pendingExtracting;

  DrillState({
    required this.sessionId,
    required this.planetId,
    required this.drillHp,
    required this.drillMaxHp,
    required this.depth,
    required this.drillX,
    required this.worldWidth,
    required this.world,
    required this.resources,
    required this.status,
    required this.totalEarned,
    required this.createdAt,
    this.completedAt,
    this.seed,
    this.pendingDirection,
    this.pendingExtracting = false,
  });

  factory DrillState.fromJson(Map<String, dynamic> json) {
    List<List<DrillCell>> world = [];
    if (json['world'] != null) {
      world = (json['world'] as List).map((row) {
        return (row as List).map((cell) => DrillCell.fromJson(cell as Map<String, dynamic>)).toList();
      }).toList();
    }

    List<DrillResource> resources = [];
    if (json['resources'] != null) {
      resources = (json['resources'] as List).map((r) => DrillResource.fromJson(r as Map<String, dynamic>)).toList();
    }

    return DrillState(
      sessionId: json['session_id'] as String? ?? '',
      planetId: json['planet_id'] as String? ?? '',
      drillHp: json['drill_hp'] as int? ?? 0,
      drillMaxHp: json['drill_max_hp'] as int? ?? 0,
      depth: json['depth'] as int? ?? 0,
      drillX: json['drill_x'] as int? ?? 0,
      worldWidth: json['world_width'] as int? ?? 5,
      world: world,
      resources: resources,
      status: json['status'] as String? ?? 'no_session',
      totalEarned: (json['total_earned'] as num?)?.toDouble() ?? 0,
      createdAt: json['created_at'] as String? ?? '',
      completedAt: json['completed_at'] as String?,
      seed: json['seed'] as int?,
      pendingDirection: json['pending_direction'] as String?,
      pendingExtracting: json['pending_extracting'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'session_id': sessionId,
      'planet_id': planetId,
      'drill_hp': drillHp,
      'drill_max_hp': drillMaxHp,
      'depth': depth,
      'drill_x': drillX,
      'world_width': worldWidth,
      'world': world.map((row) => row.map((cell) => cell.toJson()).toList()).toList(),
      'resources': resources.map((r) => r.toJson()).toList(),
      'status': status,
      'total_earned': totalEarned,
      'created_at': createdAt,
      'completed_at': completedAt,
      'seed': seed,
      'pending_direction': pendingDirection,
      'pending_extracting': pendingExtracting,
    };
  }

  bool get isActive => status == 'active';
  bool get isCompleted => status == 'completed';
  bool get isFailed => status == 'failed';
  bool get isGameEnded => isCompleted || isFailed;
  double get hpPercent => drillMaxHp > 0 ? drillHp / drillMaxHp : 0;
}

class DrillStartResponse {
  final String sessionId;
  final int seed;
  final int drillHp;
  final int drillMaxHp;
  final int depth;
  final int drillX;
  final String status;

  DrillStartResponse({
    required this.sessionId,
    required this.seed,
    required this.drillHp,
    required this.drillMaxHp,
    required this.depth,
    required this.drillX,
    required this.status,
  });

  factory DrillStartResponse.fromJson(Map<String, dynamic> json) {
    return DrillStartResponse(
      sessionId: json['session_id'] as String? ?? '',
      seed: json['seed'] as int? ?? 0,
      drillHp: json['drill_hp'] as int? ?? 0,
      drillMaxHp: json['drill_max_hp'] as int? ?? 0,
      depth: json['depth'] as int? ?? 0,
      drillX: json['drill_x'] as int? ?? 0,
      status: json['status'] as String? ?? 'active',
    );
  }
}

class DrillCommandResponse {
  final String status;

  DrillCommandResponse({required this.status});

  factory DrillCommandResponse.fromJson(Map<String, dynamic> json) {
    return DrillCommandResponse(
      status: json['status'] as String? ?? '',
    );
  }
}

class DrillChunkResponse {
  final String sessionId;
  final int seed;
  final List<List<DrillCell>> world;

  DrillChunkResponse({
    required this.sessionId,
    required this.seed,
    required this.world,
  });

  factory DrillChunkResponse.fromJson(Map<String, dynamic> json) {
    List<List<DrillCell>> world = [];
    if (json['world'] != null) {
      world = (json['world'] as List).map((row) {
        return (row as List).map((cell) => DrillCell.fromJson(cell as Map<String, dynamic>)).toList();
      }).toList();
    }

    return DrillChunkResponse(
      sessionId: json['session_id'] as String? ?? '',
      seed: json['seed'] as int? ?? 0,
      world: world,
    );
  }
}

class DrillUpdate {
  final String sessionId;
  final int drillHp;
  final int drillMaxHp;
  final int depth;
  final int drillX;
  final List<List<DrillCell>> world;
  final List<DrillResource> resources;
  final double totalEarned;
  final String status;
  final bool gameEnded;
  final String? endReason;
  final DrillHitResource? newResource;
  final double extracted;

  DrillUpdate({
    required this.sessionId,
    required this.drillHp,
    required this.drillMaxHp,
    required this.depth,
    required this.drillX,
    required this.world,
    required this.resources,
    required this.totalEarned,
    required this.status,
    required this.gameEnded,
    this.endReason,
    this.newResource,
    this.extracted = 0,
  });

  factory DrillUpdate.fromJson(Map<String, dynamic> json) {
    List<List<DrillCell>> world = [];
    if (json['world'] != null) {
      world = (json['world'] as List).map((row) {
        return (row as List).map((cell) => DrillCell.fromJson(cell as Map<String, dynamic>)).toList();
      }).toList();
    }

    List<DrillResource> resources = [];
    if (json['resources'] != null) {
      resources = (json['resources'] as List).map((r) => DrillResource.fromJson(r as Map<String, dynamic>)).toList();
    }

    DrillHitResource? newResource;
    if (json['new_resource'] != null) {
      newResource = DrillHitResource.fromJson(json['new_resource'] as Map<String, dynamic>);
    }

    return DrillUpdate(
      sessionId: json['session_id'] as String? ?? '',
      drillHp: json['drill_hp'] as int? ?? 0,
      drillMaxHp: json['drill_max_hp'] as int? ?? 0,
      depth: json['depth'] as int? ?? 0,
      drillX: json['drill_x'] as int? ?? 0,
      world: world,
      resources: resources,
      totalEarned: (json['total_earned'] as num?)?.toDouble() ?? 0,
      status: json['status'] as String? ?? 'active',
      gameEnded: json['game_ended'] as bool? ?? false,
      endReason: json['end_reason'] as String?,
      newResource: newResource,
      extracted: (json['extracted'] as num?)?.toDouble() ?? 0,
    );
  }
}

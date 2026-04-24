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
      worldWidth: json['world_width'] as int? ?? 20,
      world: world,
      resources: resources,
      status: json['status'] as String? ?? 'no_session',
      totalEarned: (json['total_earned'] as num?)?.toDouble() ?? 0,
      createdAt: json['created_at'] as String? ?? '',
      completedAt: json['completed_at'] as String?,
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
    };
  }

  bool get isActive => status == 'active';
  bool get isCompleted => status == 'completed';
  bool get isFailed => status == 'failed';
  bool get isGameEnded => isCompleted || isFailed;
  double get hpPercent => drillMaxHp > 0 ? drillHp / drillMaxHp : 0;
}

class DrillMoveResponse {
  final bool success;
  final String? message;
  final int drillHp;
  final int drillMaxHp;
  final int depth;
  final int drillX;
  final List<DrillResource> resources;
  final double totalEarned;
  final bool gameEnded;
  final String? endReason;
  final DrillHitResource? newResource;
  final double extracted;

  DrillMoveResponse({
    required this.success,
    this.message,
    required this.drillHp,
    required this.drillMaxHp,
    required this.depth,
    required this.drillX,
    required this.resources,
    required this.totalEarned,
    required this.gameEnded,
    this.endReason,
    this.newResource,
    required this.extracted,
  });

  factory DrillMoveResponse.fromJson(Map<String, dynamic> json) {
    List<DrillResource> resources = [];
    if (json['resources'] != null) {
      resources = (json['resources'] as List).map((r) => DrillResource.fromJson(r as Map<String, dynamic>)).toList();
    }

    return DrillMoveResponse(
      success: json['success'] as bool? ?? false,
      message: json['message'] as String?,
      drillHp: json['drill_hp'] as int? ?? 0,
      drillMaxHp: json['drill_max_hp'] as int? ?? 0,
      depth: json['depth'] as int? ?? 0,
      drillX: json['drill_x'] as int? ?? 0,
      resources: resources,
      totalEarned: (json['total_earned'] as num?)?.toDouble() ?? 0,
      gameEnded: json['game_ended'] as bool? ?? false,
      endReason: json['end_reason'] as String?,
      newResource: json['new_resource'] != null 
          ? DrillHitResource.fromJson(json['new_resource'] as Map<String, dynamic>) 
          : null,
      extracted: (json['extracted'] as num?)?.toDouble() ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'success': success,
      'message': message,
      'drill_hp': drillHp,
      'drill_max_hp': drillMaxHp,
      'depth': depth,
      'drill_x': drillX,
      'resources': resources.map((r) => r.toJson()).toList(),
      'total_earned': totalEarned,
      'game_ended': gameEnded,
      'end_reason': endReason,
      'new_resource': newResource?.toJson(),
      'extracted': extracted,
    };
  }
}

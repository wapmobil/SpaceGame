class MiningMonster {
  final String id;
  final String type;
  final String name;
  final String icon;
  final int x;
  final int y;
  final int hp;
  final int maxHp;
  final int damage;
  final double reward;
  final bool alive;

  MiningMonster({
    required this.id,
    required this.type,
    required this.name,
    required this.icon,
    required this.x,
    required this.y,
    required this.hp,
    required this.maxHp,
    required this.damage,
    required this.reward,
    this.alive = true,
  });

  factory MiningMonster.fromJson(Map<String, dynamic> json) {
    return MiningMonster(
      id: json['id'] as String,
      type: json['type'] as String,
      name: json['name'] as String,
      icon: json['icon'] as String? ?? '👾',
      x: json['x'] as int? ?? 0,
      y: json['y'] as int? ?? 0,
      hp: json['hp'] as int? ?? 0,
      maxHp: json['max_hp'] as int? ?? 0,
      damage: json['damage'] as int? ?? 0,
      reward: (json['reward'] as num?)?.toDouble() ?? 0,
      alive: json['alive'] as bool? ?? true,
    );
  }
}

class MiningEncounter {
  final String monsterId;
  final String monsterName;
  final String monsterIcon;
  final int damage;
  final double reward;
  final bool killed;

  MiningEncounter({
    required this.monsterId,
    required this.monsterName,
    required this.monsterIcon,
    required this.damage,
    required this.reward,
    this.killed = false,
  });

  factory MiningEncounter.fromJson(Map<String, dynamic> json) {
    return MiningEncounter(
      monsterId: json['monster_id'] as String,
      monsterName: json['monster_name'] as String,
      monsterIcon: json['monster_icon'] as String? ?? '👾',
      damage: json['damage'] as int? ?? 0,
      reward: (json['reward'] as num?)?.toDouble() ?? 0,
      killed: json['killed'] as bool? ?? false,
    );
  }
}

class MiningState {
  final String sessionId;
  final String planetId;
  final List<List<String>> maze;
  final int playerX;
  final int playerY;
  final int playerHp;
  final int playerMaxHp;
  final int playerBombs;
  final double moneyCollected;
  final String status;
  final int exitX;
  final int exitY;
  final int baseLevel;
  final List<MiningMonster> monsters;
  final List<String> availableMoves;
  final MiningEncounter? encounter;
  final bool gameEnded;
  final String? endReason;
  final DateTime? startTime;
  final DateTime? completedAt;

  MiningState({
    required this.sessionId,
    required this.planetId,
    required this.maze,
    required this.playerX,
    required this.playerY,
    this.playerHp = 100,
    this.playerMaxHp = 100,
    this.playerBombs = 3,
    this.moneyCollected = 0,
    this.status = 'active',
    required this.exitX,
    required this.exitY,
    this.baseLevel = 1,
    this.monsters = const [],
    this.availableMoves = const [],
    this.encounter,
    this.gameEnded = false,
    this.endReason,
    this.startTime,
    this.completedAt,
  });

  factory MiningState.fromJson(Map<String, dynamic> json) {
    return MiningState(
      sessionId: json['session_id'] as String,
      planetId: json['planet_id'] as String,
      maze: (json['maze'] as List?)
              ?.map((row) => (row as List).map((e) => e as String).toList())
              .toList() ??
          [],
      playerX: json['player_x'] as int? ?? 0,
      playerY: json['player_y'] as int? ?? 0,
      playerHp: json['player_hp'] as int? ?? 100,
      playerMaxHp: json['player_max_hp'] as int? ?? 100,
      playerBombs: json['player_bombs'] as int? ?? 3,
      moneyCollected: (json['money_collected'] as num?)?.toDouble() ?? 0,
      status: json['status'] as String? ?? 'active',
      exitX: json['exit_x'] as int? ?? 0,
      exitY: json['exit_y'] as int? ?? 0,
      baseLevel: json['base_level'] as int? ?? 1,
      monsters: (json['monsters'] as List?)
              ?.map((e) => MiningMonster.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      availableMoves:
          (json['available_moves'] as List?)?.map((e) => e as String).toList() ??
              [],
      encounter: json['encounter'] != null
          ? MiningEncounter.fromJson(json['encounter'] as Map<String, dynamic>)
          : null,
      gameEnded: json['game_ended'] as bool? ?? false,
      endReason: json['end_reason'] as String?,
      startTime: json['start_time'] != null
          ? DateTime.parse(json['start_time'] as String)
          : null,
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
    );
  }

  bool get canMove => status == 'active' && !gameEnded;
}

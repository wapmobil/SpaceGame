class BattleEntry {
  final String id;
  final String planetId;
  final String opponent;
  final String status;
  final double playerDamage;
  final double opponentDamage;
  final List<BattleGridCell>? grid;
  final String phase;
  final DateTime? createdAt;
  final DateTime? completedAt;

  BattleEntry({
    required this.id,
    required this.planetId,
    required this.opponent,
    this.status = 'pending',
    this.playerDamage = 0,
    this.opponentDamage = 0,
    this.grid,
    this.phase = 'idle',
    this.createdAt,
    this.completedAt,
  });

  factory BattleEntry.fromJson(Map<String, dynamic> json) {
    return BattleEntry(
      id: json['id'] as String,
      planetId: json['planet_id'] as String,
      opponent: json['opponent'] as String? ?? 'Unknown',
      status: json['status'] as String? ?? 'pending',
      playerDamage: (json['player_damage'] as num?)?.toDouble() ?? 0,
      opponentDamage: (json['opponent_damage'] as num?)?.toDouble() ?? 0,
      grid: (json['grid'] as List?)
              ?.map((e) => BattleGridCell.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      phase: json['phase'] as String? ?? 'idle',
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : null,
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
    );
  }
}

class BattleGridCell {
  final int row;
  final int col;
  final String cellType;
  final String? shipId;
  final String? shipType;
  final int? hp;
  final int? maxHp;
  final bool isPlayer;
  final bool isEnemy;
  final bool isWall;
  final bool isExit;
  final bool isExplored;

  BattleGridCell({
    required this.row,
    required this.col,
    this.cellType = 'empty',
    this.shipId,
    this.shipType,
    this.hp,
    this.maxHp,
    this.isPlayer = false,
    this.isEnemy = false,
    this.isWall = false,
    this.isExit = false,
    this.isExplored = false,
  });

  factory BattleGridCell.fromJson(Map<String, dynamic> json) {
    return BattleGridCell(
      row: json['row'] as int? ?? 0,
      col: json['col'] as int? ?? 0,
      cellType: json['type'] as String? ?? 'empty',
      shipId: json['ship_id'] as String?,
      shipType: json['ship_type'] as String?,
      hp: json['hp'] as int?,
      maxHp: json['max_hp'] as int?,
      isPlayer: json['is_player'] as bool? ?? false,
      isEnemy: json['is_enemy'] as bool? ?? false,
      isWall: json['is_wall'] as bool? ?? false,
      isExit: json['is_exit'] as bool? ?? false,
      isExplored: json['is_explored'] as bool? ?? false,
    );
  }
}

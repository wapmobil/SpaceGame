class ShipyardInfo {
  final Map<String, dynamic> ships;
  final int totalShips;
  final int totalSlots;
  final int maxSlots;
  final double totalCargo;
  final double totalEnergy;
  final double totalDamage;
  final double totalHP;
  final int shipyardLevel;
  final int shipyardQueueLen;
  final double shipyardProgress;

  ShipyardInfo({
    this.ships = const {},
    this.totalShips = 0,
    this.totalSlots = 0,
    this.maxSlots = 0,
    this.totalCargo = 0,
    this.totalEnergy = 0,
    this.totalDamage = 0,
    this.totalHP = 0,
    this.shipyardLevel = 0,
    this.shipyardQueueLen = 0,
    this.shipyardProgress = 0,
  });

  factory ShipyardInfo.fromJson(Map<String, dynamic> json) {
    return ShipyardInfo(
      ships: json['ships'] as Map<String, dynamic>? ?? {},
      totalShips: json['total_ships'] as int? ?? 0,
      totalSlots: json['total_slots'] as int? ?? 0,
      maxSlots: json['max_slots'] as int? ?? 0,
      totalCargo: (json['total_cargo'] as num?)?.toDouble() ?? 0,
      totalEnergy: (json['total_energy'] as num?)?.toDouble() ?? 0,
      totalDamage: (json['total_damage'] as num?)?.toDouble() ?? 0,
      totalHP: (json['total_hp'] as num?)?.toDouble() ?? 0,
      shipyardLevel: json['shipyard_level'] as int? ?? 0,
      shipyardQueueLen: json['shipyard_queue_len'] as int? ?? 0,
      shipyardProgress: (json['shipyard_progress'] as num?)?.toDouble() ?? 0,
    );
  }

  int get availableSlots => maxSlots - totalSlots;
  double get buildProgressPercent => (shipyardProgress / (shipyardQueueLen > 0 ? 1 : 100)) * 100;
}

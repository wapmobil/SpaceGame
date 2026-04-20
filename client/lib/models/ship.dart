class Ship {
  final String id;
  final String planetId;
  final String type;
  final int hp;
  final int armor;
  final List<dynamic> weapons;

  Ship({
    required this.id,
    required this.planetId,
    required this.type,
    this.hp = 100,
    this.armor = 0,
    List<dynamic>? weapons,
  }) : weapons = weapons ?? [];

  factory Ship.fromJson(Map<String, dynamic> json) {
    return Ship(
      id: json['id'] as String,
      planetId: json['planet_id'] as String,
      type: json['type'] as String,
      hp: json['hp'] as int? ?? 100,
      armor: json['armor'] as int? ?? 0,
      weapons: json['weapons'] as List<dynamic>? ?? [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'planet_id': planetId,
      'type': type,
      'hp': hp,
      'armor': armor,
      'weapons': weapons,
    };
  }
}

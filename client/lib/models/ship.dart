class Ship {
  final String id;
  final String planetId;
  final String type;
  final int hp;
  final int maxHp;
  final int armor;
  final List<dynamic> weapons;
  final double cargo;
  final double energy;

  Ship({
    required this.id,
    required this.planetId,
    required this.type,
    this.hp = 100,
    this.maxHp = 100,
    this.armor = 0,
    List<dynamic>? weapons,
    this.cargo = 0,
    this.energy = 0,
  }) : weapons = weapons ?? [];

  factory Ship.fromJson(Map<String, dynamic> json) {
    return Ship(
      id: json['id'] as String,
      planetId: json['planet_id'] as String,
      type: json['type'] as String,
      hp: json['hp'] as int? ?? 100,
      maxHp: json['max_hp'] as int? ?? json['hp'] as int? ?? 100,
      armor: json['armor'] as int? ?? 0,
      weapons: json['weapons'] as List<dynamic>? ?? [],
      cargo: (json['cargo'] as num?)?.toDouble() ?? 0,
      energy: (json['energy'] as num?)?.toDouble() ?? 0,
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

class ShipType {
  final String typeId;
  final String name;
  final String description;
  final int slots;
  final double cargo;
  final double energy;
  final double hp;
  final double armor;
  final double weaponMinDmg;
  final double weaponMaxDmg;
  final ShipCost cost;
  final double buildTime;
  final int minShipyard;
  final bool canBuild;

  ShipType({
    required this.typeId,
    required this.name,
    required this.description,
    required this.slots,
    required this.cargo,
    required this.energy,
    required this.hp,
    required this.armor,
    required this.weaponMinDmg,
    required this.weaponMaxDmg,
    required this.cost,
    required this.buildTime,
    required this.minShipyard,
    this.canBuild = false,
  });

  factory ShipType.fromJson(Map<String, dynamic> json) {
    return ShipType(
      typeId: json['type_id'] as String,
      name: json['name'] as String,
      description: json['description'] as String,
      slots: json['slots'] as int? ?? 0,
      cargo: (json['cargo'] as num?)?.toDouble() ?? 0,
      energy: (json['energy'] as num?)?.toDouble() ?? 0,
      hp: (json['hp'] as num?)?.toDouble() ?? 0,
      armor: (json['armor'] as num?)?.toDouble() ?? 0,
      weaponMinDmg: (json['weapon_min_damage'] as num?)?.toDouble() ?? 0,
      weaponMaxDmg: (json['weapon_max_damage'] as num?)?.toDouble() ?? 0,
      cost: ShipCost.fromJson((json['cost'] as Map?)?.cast<String, dynamic>() ?? {}),
      buildTime: (json['build_time'] as num?)?.toDouble() ?? 0,
      minShipyard: json['min_shipyard_level'] as int? ?? 1,
      canBuild: json['can_build'] as bool? ?? false,
    );
  }

  bool get isPeaceful => weaponMinDmg == 0 && weaponMaxDmg == 0;
}

class ShipCost {
  final double food;
  final double composite;
  final double mechanisms;
  final double reagents;
  final double money;

  ShipCost({
    this.food = 0,
    this.composite = 0,
    this.mechanisms = 0,
    this.reagents = 0,
    this.money = 0,
  });

  factory ShipCost.fromJson(Map<String, dynamic> json) {
    return ShipCost(
      food: (json['food'] as num?)?.toDouble() ?? 0,
      composite: (json['composite'] as num?)?.toDouble() ?? 0,
      mechanisms: (json['mechanisms'] as num?)?.toDouble() ?? 0,
      reagents: (json['reagents'] as num?)?.toDouble() ?? 0,
      money: (json['money'] as num?)?.toDouble() ?? 0,
    );
  }
}

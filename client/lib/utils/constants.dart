class AppConfig {
  static const String appName = 'SpaceGame';
  static const String defaultBaseUrl = 'http://localhost:8080';
  static const int defaultPort = 8080;
}

class Constants {
  static const String appName = 'SpaceGame';
  static const String apiBaseUrl = 'http://localhost:8080';
  static const String wsUrl = 'ws://localhost:8080/ws';

  static const resourceNames = {
    'food': 'Food',
    'composite': 'Composite',
    'mechanisms': 'Mechanisms',
    'reagents': 'Reagents',
    'energy': 'Energy',
    'money': 'Money',
    'alien_tech': 'Alien Tech',
  };

  static const resourceIcons = {
    'food': '🍖',
    'composite': '🧬',
    'mechanisms': '⚙️',
    'reagents': '🧪',
    'energy': '⚡',
    'money': '💰',
    'alien_tech': '👾',
  };

  static const resourceColors = {
    'food': 0xFFff9800,
    'composite': 0xFF8bc34a,
    'mechanisms': 0xFF607d8b,
    'reagents': 0xFF9c27b0,
    'energy': 0xFFFFeb3b,
    'money': 0xFFffd700,
    'alien_tech': 0xFF00bcd4,
  };

  static const buildingTypes = {
    'base': {'name': 'Base', 'icon': '🏠', 'description': 'Headquarters'},
    'farm': {'name': 'Farm', 'icon': '🌾', 'description': 'Produces food'},
    'composite_drone': {'name': 'Composite Drone', 'icon': '🤖', 'description': 'Produces composite'},
    'mechanism_factory': {'name': 'Mechanism Factory', 'icon': '🏭', 'description': 'Produces mechanisms'},
    'reagent_lab': {'name': 'Reagent Lab', 'icon': '🔬', 'description': 'Produces reagents'},
    'solar': {'name': 'Solar Panel', 'icon': '☀️', 'description': 'Produces energy'},
    'energy_storage': {'name': 'Energy Storage', 'icon': '🔋', 'description': 'Stores energy'},
    'storage': {'name': 'Storage', 'icon': '📦', 'description': 'Increases storage capacity'},
    'factory': {'name': 'Factory', 'icon': '🏗️', 'description': 'Advanced production'},
    'shipyard': {'name': 'Shipyard', 'icon': '🚀', 'description': 'Builds ships'},
    'comcenter': {'name': 'Comm Center', 'icon': '📡', 'description': 'Communication'},
  };

  static const shipIcons = {
    'trade_ship': '🚢',
    'small_ship': '🛸',
    'interceptor': '✈️',
    'corvette': '🚀',
    'frigate': '🛩️',
    'cruiser': '🛶',
  };

  static const expeditionTypes = {
    'exploration': {'name': 'Exploration', 'icon': '🗺️', 'description': 'Discover new systems'},
    'trade': {'name': 'Trade', 'icon': '💰', 'description': 'Generate revenue'},
    'support': {'name': 'Support', 'icon': '🏥', 'description': 'Assist allied forces'},
  };

  static const miningIcons = {
    'wall': '🧱',
    'floor': '⬛',
    'player': '🧑‍🚀',
    'exit': '🚪',
    'money': '💰',
    'monster': '👾',
    'bomb': '💣',
  };

  static const techList = [
    {'id': 'planet_exploration', 'name': 'Planet Exploration', 'description': 'Unlocks Factory building', 'cost_food': 100, 'cost_money': 100, 'build_time': 60, 'max_level': 1, 'depends_on': []},
    {'id': 'energy_storage', 'name': 'Energy Storage', 'description': 'Unlocks Energy Storage building', 'cost_food': 200, 'cost_money': 150, 'build_time': 90, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'energy_saving', 'name': 'Energy Saving', 'description': '-10% energy consumption per level', 'cost_food': 300, 'cost_money': 200, 'build_time': 120, 'max_level': 4, 'depends_on': ['energy_storage']},
    {'id': 'trade', 'name': 'Trade', 'description': 'Unlocks Marketplace', 'cost_food': 400, 'cost_money': 300, 'build_time': 120, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'ships', 'name': 'Ships', 'description': 'Unlocks Shipyard', 'cost_food': 500, 'cost_money': 400, 'build_time': 150, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'upgraded_energy_storage', 'name': 'Upgraded Energy Storage', 'description': '+20% energy capacity per level', 'cost_food': 600, 'cost_money': 500, 'build_time': 180, 'max_level': 3, 'depends_on': ['energy_saving']},
    {'id': 'fast_construction', 'name': 'Fast Construction', 'description': 'Building speed bonus per level', 'cost_food': 800, 'cost_money': 600, 'build_time': 200, 'max_level': 3, 'depends_on': ['ships']},
    {'id': 'parallel_construction', 'name': 'Parallel Construction', 'description': '+1 simultaneous construction per level (up to 3 levels)', 'cost_food': 2000, 'cost_money': 1500, 'build_time': 300, 'max_level': 3, 'depends_on': ['fast_construction']},
    {'id': 'compact_storage', 'name': 'Compact Storage', 'description': '2x storage capacity per level', 'cost_food': 1000, 'cost_money': 800, 'build_time': 240, 'max_level': 3, 'depends_on': ['ships']},
    {'id': 'expeditions', 'name': 'Expeditions', 'description': 'Unlocks expedition system', 'cost_food': 1500, 'cost_money': 1000, 'build_time': 300, 'max_level': 1, 'depends_on': ['trade']},
    {'id': 'command_center', 'name': 'Command Center', 'description': 'Unlocks alien technology tree', 'cost_food': 5000, 'cost_money': 3000, 'build_time': 600, 'max_level': 1, 'depends_on': ['expeditions']},
    {'id': 'alien_technologies', 'name': 'Alien Technologies', 'description': 'Unlocks alien technology tree', 'cost_alien_tech': 10, 'build_time': 300, 'max_level': 1, 'depends_on': ['command_center']},
    {'id': 'additional_expedition', 'name': 'Additional Expedition', 'description': '+1 concurrent expedition', 'cost_alien_tech': 15, 'build_time': 200, 'max_level': 1, 'depends_on': ['alien_technologies']},
    {'id': 'super_energy_storage', 'name': 'Super Energy Storage', 'description': '+20% energy capacity per level', 'cost_alien_tech': 20, 'build_time': 300, 'max_level': 5, 'depends_on': ['alien_technologies']},
  ];

  static String formatResources(double amount) {
    if (amount >= 1000000) {
      return '${(amount / 1000000).toStringAsFixed(1)}M';
    }
    if (amount >= 1000) {
      return '${(amount / 1000).toStringAsFixed(1)}K';
    }
    return amount.toStringAsFixed(amount == amount.toInt() ? 0 : 1);
  }

  static String formatTime(double seconds) {
    final mins = seconds ~/ 60;
    final secs = (seconds % 60).toInt();
    if (mins > 0) {
      return '$mins:${secs.toString().padLeft(2, '0')}';
    }
    return '$secs s';
  }

  static String formatDateTime(DateTime? dt) {
    if (dt == null) return 'N/A';
    return '${dt.month}/${dt.day} ${dt.hour}:${dt.minute.toString().padLeft(2, '0')}';
  }
}

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
    'food': 'Еда',
    'composite': 'Композит',
    'mechanisms': 'Механизмы',
    'reagents': 'Реагенты',
    'energy': 'Энергия',
    'money': 'Деньги',
    'alien_tech': 'Чертежи',
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
    'base': {'name': 'Центр исследований', 'icon': '🏠', 'description': 'Штаб-квартира'},
    'farm': {'name': 'Ферма', 'icon': '🌾', 'description': 'Производит еду'},
    'composite_drone': {'name': 'Дрон-композит', 'icon': '🤖', 'description': 'Производит композит'},
    'mechanism_factory': {'name': 'Механическая фабрика', 'icon': '🏭', 'description': 'Производит механизмы'},
    'reagent_lab': {'name': 'Лаборатория реагентов', 'icon': '🔬', 'description': 'Производит реагенты'},
    'solar': {'name': 'Солнечная панель', 'icon': '☀️', 'description': 'Производит энергию'},
    'energy_storage': {'name': 'Аккумулятор', 'icon': '🔋', 'description': 'Накапливает энергию'},
    'storage': {'name': 'Склад', 'icon': '📦', 'description': 'Увеличивает вместимость хранилища'},
    'factory': {'name': 'Фабрика', 'icon': '🏗️', 'description': 'Расширенное производство'},
    'shipyard': {'name': 'Верфь', 'icon': '🚀', 'description': 'Строит корабли'},
    'comcenter': {'name': 'Центр связи', 'icon': '📡', 'description': 'Связь'},
    'dynamo': {'name': 'Динамомашинa', 'icon': '⚡', 'description': 'Производит энергию, потребждает еду'},
  };

  static const researchRequirements = {
    'energy_storage': 'energy_storage',
    'shipyard': 'ships',
    'comcenter': 'expeditions',
  };

  static const researchRandomUnlockBuildings = [
    'composite_drone',
    'mechanism_factory',
    'reagent_lab',
  ];

  static const researchNames = {
    'planet_exploration': 'Исследование планет',
    'energy_storage': 'Аккумулятор',
    'ships': 'Корабли',
    'expeditions': 'Экспедиции',
    'fast_construction': 'Быстрое строительство',
  };

  static const buildingNames = {
    'composite_drone': 'Дрон-композит',
    'mechanism_factory': 'Механическая фабрика',
    'reagent_lab': 'Лаборатория реагентов',
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
    'exploration': {'name': 'Разведка', 'icon': '🗺️', 'description': 'Открывает новые системы'},
    'trade': {'name': 'Торговля', 'icon': '💰', 'description': 'Генерирует доход'},
    'support': {'name': 'Поддержка', 'icon': '🏥', 'description': 'Помогает союзным силам'},
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
    {'id': 'planet_exploration', 'name': 'Исследование планет', 'description': 'Открывает здание Фабрики', 'cost_food': 100, 'cost_money': 100, 'build_time': 60, 'max_level': 1, 'depends_on': []},
    {'id': 'energy_storage', 'name': 'Аккумулятор', 'description': 'Открывает здание Аккумулятора', 'cost_food': 200, 'cost_money': 150, 'build_time': 90, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'energy_saving', 'name': 'Энергосбережение', 'description': '-10% расхода энергии за уровень', 'cost_food': 300, 'cost_money': 200, 'build_time': 120, 'max_level': 4, 'depends_on': ['energy_storage']},
    {'id': 'trade', 'name': 'Торговля', 'description': 'Открывает Рынок', 'cost_food': 400, 'cost_money': 300, 'build_time': 120, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'ships', 'name': 'Корабли', 'description': 'Открывает Верфь', 'cost_food': 500, 'cost_money': 400, 'build_time': 150, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'upgraded_energy_storage', 'name': 'Улучшенный накопитель', 'description': '+20% вместимости энергии за уровень', 'cost_food': 600, 'cost_money': 500, 'build_time': 180, 'max_level': 3, 'depends_on': ['energy_saving']},
    {'id': 'fast_construction', 'name': 'Быстрое строительство', 'description': 'Бонус скорости строительства за уровень', 'cost_food': 800, 'cost_money': 600, 'build_time': 200, 'max_level': 3, 'depends_on': ['ships']},
    {'id': 'parallel_construction', 'name': 'Параллельное строительство', 'description': '+1 одновременное строительство за уровень (до 3 уровней)', 'cost_food': 2000, 'cost_money': 1500, 'build_time': 300, 'max_level': 3, 'depends_on': ['fast_construction']},
    {'id': 'compact_storage', 'name': 'Компактное хранилище', 'description': '2x вместимость хранилища за уровень', 'cost_food': 1000, 'cost_money': 800, 'build_time': 240, 'max_level': 3, 'depends_on': ['ships']},
    {'id': 'expeditions', 'name': 'Экспедиции', 'description': 'Открывает систему экспедиций', 'cost_food': 1500, 'cost_money': 1000, 'build_time': 300, 'max_level': 1, 'depends_on': ['trade']},
    {'id': 'command_center', 'name': 'Командный центр', 'description': 'Открывает древо чертежей', 'cost_food': 5000, 'cost_money': 3000, 'build_time': 600, 'max_level': 1, 'depends_on': ['expeditions']},
    {'id': 'alien_technologies', 'name': 'Чертежи', 'description': 'Открывает древо чертежей', 'cost_alien_tech': 10, 'build_time': 300, 'max_level': 1, 'depends_on': ['command_center']},
    {'id': 'additional_expedition', 'name': 'Дополнительная экспедиция', 'description': '+1 одновременная экспедиция', 'cost_alien_tech': 15, 'build_time': 200, 'max_level': 1, 'depends_on': ['alien_technologies']},
    {'id': 'super_energy_storage', 'name': 'Супер накопитель', 'description': '+20% вместимости энергии за уровень', 'cost_alien_tech': 20, 'build_time': 300, 'max_level': 5, 'depends_on': ['alien_technologies']},
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
    return '$secs с';
  }

  static String formatDateTime(DateTime? dt) {
    if (dt == null) return 'Н/Д';
    return '${dt.month}/${dt.day} ${dt.hour}:${dt.minute.toString().padLeft(2, '0')}';
  }
}

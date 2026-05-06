import 'package:flutter/material.dart';

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
    'iron': 'Железо',
    'composite': 'Композит',
    'mechanisms': 'Механизмы',
    'reagents': 'Реагенты',
    'energy': 'Энергия',
    'money': 'Деньги',
    'alien_tech': 'Чертежи',
  };

  static const resourceIcons = {
    'food': '🍍',
    'iron': '🪨',
    'composite': '🧬',
    'mechanisms': '⚙️',
    'reagents': '🧪',
    'energy': '⚡',
    'money': '💰',
    'alien_tech': '👾',
  };

  static const resourceColors = {
    'food': Color(0xFFFF9800),
    'iron': Color(0xFFA1887F),
    'composite': Color(0xFF8BC34A),
    'mechanisms': Color(0xFF607D8B),
    'reagents': Color(0xFF9C27B0),
    'energy': Color(0xFFFFEB3B),
    'money': Color(0xFFFFD700),
    'alien_tech': Color(0xFF00BCD4),
  };

 static const buildingTypes = {
    'base': {'name': 'Центр исследований', 'icon': '🏠', 'description': 'Штаб-квартира'},
    'farm': {'name': 'Ферма', 'icon': '🌾', 'description': 'Производит еду'},
    'solar': {'name': 'Солнечная панель', 'icon': '☀️', 'description': 'Производит энергию'},
    'energy_storage': {'name': 'Аккумулятор', 'icon': '🔋', 'description': 'Накапливает энергию'},
    'storage': {'name': 'Склад', 'icon': '📦', 'description': 'Увеличивает вместимость хранилища'},
    'mine': {'name': 'Шахта', 'icon': '⛏️', 'description': 'Добывает железо'},
    'shipyard': {'name': 'Верфь', 'icon': '🚀', 'description': 'Строит корабли'},
    'market': {'name': 'Рынок', 'icon': '🏪', 'description': 'Открывает доступ к Рынку'},
    'dynamo': {'name': 'Динамомашинa', 'icon': '⚡', 'description': 'Производит энергию, потребщает еду'},
  };

  static const researchRequirements = {
    'energy_storage': 'energy_storage',
    'shipyard': 'ships',
    'market': 'trade',
  };

   static const researchNames = {
    'planet_exploration': 'Исследование планет',
    'energy_storage': 'Аккумулятор',
    'ships': 'Корабли',
    'expeditions': 'Экспедиции',
    'fast_construction': 'Быстрое строительство',
    'location_buildings': 'Здания на локациях',
    'advanced_exploration': 'Углублённая разведка',
    'command_center': 'Командный центр',
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
    'space_exploration': {'name': 'Разведка', 'icon': '🗺️', 'description': 'Открывает новые системы'},
    'space_trade': {'name': 'Торговля', 'icon': '💰', 'description': 'Генерирует доход'},
    'space_support': {'name': 'Поддержка', 'icon': '🏥', 'description': 'Помогает союзным силам'},
  };

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

  static String getLocationRarityLabel(String type) {
    final common = ['pond', 'river', 'forest', 'mineral_deposit', 'dry_valley'];
    final uncommon = ['waterfall', 'cave', 'thermal_spring', 'salt_lake', 'wind_pass'];
    final rare = ['crystal_cave', 'meteor_crater', 'sunken_city', 'glacier', 'mushroom_forest'];
    final exotic = ['crystal_field', 'cloud_island', 'underground_lake', 'radioactive_zone', 'anomaly_zone'];

    if (common.contains(type)) return 'Обычная';
    if (uncommon.contains(type)) return 'Необычная';
    if (rare.contains(type)) return 'Редкая';
    if (exotic.contains(type)) return 'Экзотическая';
    return 'Неизвестная';
  }

  static Color getLocationRarityColor(String type) {
    final rarity = getLocationRarityLabel(type);
    switch (rarity) {
      case 'Обычная':
        return const Color(0xFF9e9e9e);
      case 'Необычная':
        return const Color(0xFF4caf50);
      case 'Редкая':
        return const Color(0xFF2196f3);
      case 'Экзотическая':
        return const Color(0xFF9c27b0);
      default:
        return Colors.white54;
    }
  }
}

import 'package:flutter/material.dart';
import '../core/app_theme.dart';

class ResearchTree extends StatelessWidget {
  final List research;
  final List available;
  final Function(String techId) onResearch;

  const ResearchTree({
    super.key,
    required this.research,
    required this.available,
    required this.onResearch,
  });

  @override
  Widget build(BuildContext context) {
    if (research.isEmpty && available.isEmpty) {
      return const Center(child: Text('Нет данных об исследованиях', style: TextStyle(color: Colors.white38)));
    }

    final availableIds = available.map((t) => t.techId).toSet();
    final completedIds = research.where((t) => t.completed).map((t) => t.techId).toSet();
    final inProgressIds = research.where((t) => t.inProgress).map((t) => t.techId).toSet();

    return Column(
      children: Constants.techList.map((tech) {
        final techMap = Map<String, dynamic>.from(tech);
        final techId = techMap['id'] as String;
        final name = techMap['name'] as String;
        final description = techMap['description'] as String;
        final dependsOn = (techMap['depends_on'] as List).map((e) => e as String).toList();

        final isCompleted = completedIds.contains(techId);
        final isInProgress = inProgressIds.contains(techId);
        final isAvailable = availableIds.contains(techId);
        final hasPrerequisites = dependsOn.every((dep) => completedIds.contains(dep));

        Color statusColor;
        if (isCompleted) statusColor = AppTheme.successColor;
        else if (isInProgress) statusColor = AppTheme.warningColor;
        else if (isAvailable && hasPrerequisites) statusColor = AppTheme.accentColor;
        else statusColor = Colors.white24;

        return Padding(
          padding: const EdgeInsets.only(bottom: 4),
          child: _TechNode(
            techId: techId,
            name: name,
            description: description,
            dependsOn: dependsOn,
            statusColor: statusColor,
            isCompleted: isCompleted,
            isInProgress: isInProgress,
            isAvailable: isAvailable && hasPrerequisites,
            onResearch: () => onResearch(techId),
          ),
        );
      }).toList(),
    );
  }
}

class _TechNode extends StatelessWidget {
  final String techId;
  final String name;
  final String description;
  final List<String> dependsOn;
  final Color statusColor;
  final bool isCompleted;
  final bool isInProgress;
  final bool isAvailable;
  final VoidCallback onResearch;

  const _TechNode({
    required this.techId,
    required this.name,
    required this.description,
    required this.dependsOn,
    required this.statusColor,
    required this.isCompleted,
    required this.isInProgress,
    required this.isAvailable,
    required this.onResearch,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: statusColor, width: 2),
        borderRadius: BorderRadius.circular(8),
        color: statusColor.withValues(alpha: 0.1),
      ),
      child: ListTile(
        dense: true,
        leading: CircleAvatar(
          radius: 12,
          backgroundColor: statusColor,
          child: isCompleted
              ? const Icon(Icons.check, size: 14, color: Colors.white)
              : isInProgress
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                    )
                  : const Icon(Icons.science, size: 14, color: Colors.white),
        ),
        title: Text(
          name,
          style: TextStyle(
            color: isCompleted ? AppTheme.successColor : Colors.white,
            fontWeight: isAvailable ? FontWeight.bold : FontWeight.normal,
          ),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(description, style: const TextStyle(fontSize: 11, color: Colors.white54)),
            if (isInProgress) ...[
              const SizedBox(height: 4),
              LinearProgressIndicator(
                minHeight: 4,
                borderRadius: BorderRadius.circular(2),
                color: AppTheme.warningColor,
              ),
            ],
            if (dependsOn.isNotEmpty)
              Text(
                'Требуется: ${dependsOn.join(", ")}',
                style: const TextStyle(fontSize: 10, color: Colors.white38),
              ),
          ],
        ),
        trailing: isAvailable
            ? ElevatedButton(
                onPressed: onResearch,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  backgroundColor: AppTheme.accentColor,
                ),
                child: const Text('Исследовать', style: TextStyle(fontSize: 11)),
              )
            : null,
      ),
    );
  }
}

class Constants {
  static const techList = [
    {'id': 'planet_exploration', 'name': '🌍Разведка планеты', 'description': 'Открывает здание Фабрики', 'cost_food': 100, 'cost_money': 100, 'build_time': 60, 'max_level': 1, 'depends_on': []},
    {'id': 'energy_storage', 'name': '🔋Аккумуляторы', 'description': 'Открывает здание Аккумулятора', 'cost_food': 200, 'cost_money': 150, 'build_time': 90, 'max_level': 5, 'depends_on': ['planet_exploration']},
    {'id': 'energy_saving', 'name': '🔌Экономия энергии', 'description': '-10% расхода энергии за уровень', 'cost_food': 300, 'cost_money': 200, 'build_time': 120, 'max_level': 4, 'depends_on': ['energy_storage']},
    {'id': 'upgraded_energy_storage', 'name': '🔋Улучшенные аккумуляторы', 'description': '+20% вместимости энергии за уровень', 'cost_food': 600, 'cost_money': 500, 'build_time': 180, 'max_level': 3, 'depends_on': ['energy_saving']},
    {'id': 'upgraded_energy_storage_2', 'name': '🔋Улучшенные аккумуляторы 2', 'description': 'Максимальный буст энергии', 'cost_food': 800, 'cost_money': 700, 'build_time': 200, 'max_level': 1, 'depends_on': ['upgraded_energy_storage']},
    {'id': 'trade', 'name': '💸Торговля', 'description': 'Открывает Рынок', 'cost_food': 400, 'cost_money': 300, 'build_time': 120, 'max_level': 2, 'depends_on': ['planet_exploration']},
    {'id': 'trade_connections', 'name': '💵Торговые связи', 'description': 'Открывает расширенные опции торговли', 'cost_food': 600, 'cost_money': 450, 'build_time': 150, 'max_level': 1, 'depends_on': ['trade']},
    {'id': 'ships', 'name': '🚀Корабли', 'description': 'Открывает Верфь', 'cost_food': 500, 'cost_money': 400, 'build_time': 150, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'expeditions', 'name': '👣Экспедиции', 'description': 'Открывает систему экспедиций', 'cost_food': 1500, 'cost_money': 1000, 'build_time': 300, 'max_level': 1, 'depends_on': ['trade']},
    {'id': 'command_center', 'name': '🏪Командный центр', 'description': 'Открывает древо инопланетных технологий', 'cost_food': 5000, 'cost_money': 3000, 'build_time': 600, 'max_level': 1, 'depends_on': ['expeditions']},
    {'id': 'fast_construction', 'name': '🛠Быстрое строительство', 'description': 'Бонус скорости строительства за уровень', 'cost_food': 800, 'cost_money': 600, 'build_time': 200, 'max_level': 3, 'depends_on': ['ships']},
    {'id': 'compact_storage', 'name': '📦Компактное хранение', 'description': '2x вместимость хранилища за уровень', 'cost_food': 1000, 'cost_money': 800, 'build_time': 240, 'max_level': 3, 'depends_on': ['fast_construction']},
    {'id': 'fast_construction_2', 'name': '🛠Быстрое строительство 2', 'description': 'Дополнительный буст скорости', 'cost_food': 1200, 'cost_money': 900, 'build_time': 250, 'max_level': 1, 'depends_on': ['fast_construction']},
    {'id': 'compact_storage_2', 'name': '📦Компактное хранение 2', 'description': '4x вместимость хранилища', 'cost_food': 1500, 'cost_money': 1200, 'build_time': 300, 'max_level': 1, 'depends_on': ['compact_storage', 'fast_construction_2']},
    {'id': 'fast_construction_3', 'name': '🛠Быстрое строительство 3', 'description': 'Максимальный буст скорости', 'cost_food': 2000, 'cost_money': 1500, 'build_time': 350, 'max_level': 1, 'depends_on': ['fast_construction_2']},
    {'id': 'compact_storage_3', 'name': '📦Компактное хранение 3', 'description': '8x вместимость хранилища', 'cost_food': 2500, 'cost_money': 2000, 'build_time': 400, 'max_level': 1, 'depends_on': ['compact_storage_2', 'fast_construction_3']},
    {'id': 'parallel_construction', 'name': '🔧Параллельное строительство', 'description': '+1 одновременный проект за уровень', 'cost_food': 2000, 'cost_money': 1500, 'build_time': 300, 'max_level': 3, 'depends_on': ['fast_construction', 'compact_storage']},
    {'id': 'alien_technologies', 'name': '👽Инопланетные технологии', 'description': 'Открывает древо инопланетных технологий', 'cost_alien_tech': 10, 'build_time': 300, 'max_level': 1, 'depends_on': ['command_center']},
    {'id': 'additional_expedition', 'name': '🛸Дополнительная экспедиция', 'description': '+1 одновременная экспедиция', 'cost_alien_tech': 15, 'build_time': 200, 'max_level': 1, 'depends_on': ['alien_technologies']},
    {'id': 'super_energy_storage', 'name': '⚡Супер накопитель', 'description': '+20% вместимости энергии за уровень', 'cost_alien_tech': 20, 'build_time': 300, 'max_level': 5, 'depends_on': ['alien_technologies']},
  ];
}

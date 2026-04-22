import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';

class ResearchScreen extends StatelessWidget {
  const ResearchScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Исследования')),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          final planet = gameProvider.selectedPlanet;
          if (planet == null)    return const Center(child: Text('Планета не выбрана'));

          return RefreshIndicator(
            onRefresh: () async => gameProvider.loadResearch(planet.id),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildTreeSection(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildAllTechsSection(context, gameProvider),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildTreeSection(BuildContext context, GameProvider gameProvider) {
    final state = gameProvider.researchState;
    if (state == null) return const Center(child: CircularProgressIndicator());

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Древо исследований', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 12),
            _buildResearchTree(context, state.research, state.available, gameProvider),
          ],
        ),
      ),
    );
  }

  Widget _buildResearchTree(BuildContext context, List research, List available, GameProvider gameProvider) {
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
            onResearch: () => gameProvider.startResearch(techId),
          ),
        );
      }).toList(),
    );
  }

  Widget _buildAllTechsSection(BuildContext context, GameProvider gameProvider) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Все технологии', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 8),
            ...Constants.techList.map((tech) {
              final techMap = Map<String, dynamic>.from(tech);
              return Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Text(
                  '• ${techMap['name']} ${techMap['depends_on'].isNotEmpty ? '(требуется: ${(techMap['depends_on'] as List).join(", ")})' : ''}',
                  style: const TextStyle(fontSize: 12, color: Colors.white70),
                ),
              );
            }),
          ],
        ),
      ),
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
    {'id': 'planet_exploration', 'name': 'Исследование планет', 'description': 'Открывает здание Фабрики', 'cost_food': 100, 'cost_money': 100, 'build_time': 60, 'max_level': 1, 'depends_on': []},
    {'id': 'energy_storage', 'name': 'Накопитель энергии', 'description': 'Открывает здание Накопителя энергии', 'cost_food': 200, 'cost_money': 150, 'build_time': 90, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'energy_saving', 'name': 'Энергосбережение', 'description': '-10% расхода энергии за уровень', 'cost_food': 300, 'cost_money': 200, 'build_time': 120, 'max_level': 4, 'depends_on': ['energy_storage']},
    {'id': 'trade', 'name': 'Торговля', 'description': 'Открывает Рынок', 'cost_food': 400, 'cost_money': 300, 'build_time': 120, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'ships', 'name': 'Корабли', 'description': 'Открывает Верфь', 'cost_food': 500, 'cost_money': 400, 'build_time': 150, 'max_level': 1, 'depends_on': ['planet_exploration']},
    {'id': 'upgraded_energy_storage', 'name': 'Улучшенный накопитель', 'description': '+20% вместимости энергии за уровень', 'cost_food': 600, 'cost_money': 500, 'build_time': 180, 'max_level': 3, 'depends_on': ['energy_saving']},
    {'id': 'fast_construction', 'name': 'Быстрое строительство', 'description': 'Бонус скорости строительства за уровень', 'cost_food': 800, 'cost_money': 600, 'build_time': 200, 'max_level': 3, 'depends_on': ['ships']},
    {'id': 'compact_storage', 'name': 'Компактное хранилище', 'description': '2x вместимость хранилища за уровень', 'cost_food': 1000, 'cost_money': 800, 'build_time': 240, 'max_level': 3, 'depends_on': ['ships']},
    {'id': 'expeditions', 'name': 'Экспедиции', 'description': 'Открывает систему экспедиций', 'cost_food': 1500, 'cost_money': 1000, 'build_time': 300, 'max_level': 1, 'depends_on': ['trade']},
    {'id': 'command_center', 'name': 'Командный центр', 'description': 'Открывает древо чужих технологий', 'cost_food': 5000, 'cost_money': 3000, 'build_time': 600, 'max_level': 1, 'depends_on': ['expeditions']},
    {'id': 'alien_technologies', 'name': 'Чужие технологии', 'description': 'Открывает древо чужих технологий', 'cost_alien_tech': 10, 'build_time': 300, 'max_level': 1, 'depends_on': ['command_center']},
    {'id': 'additional_expedition', 'name': 'Дополнительная экспедиция', 'description': '+1 одновременная экспедиция', 'cost_alien_tech': 15, 'build_time': 200, 'max_level': 1, 'depends_on': ['alien_technologies']},
    {'id': 'super_energy_storage', 'name': 'Супер накопитель', 'description': '+20% вместимости энергии за уровень', 'cost_alien_tech': 20, 'build_time': 300, 'max_level': 5, 'depends_on': ['alien_technologies']},
  ];
}

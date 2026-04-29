import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';
import '../models/research.dart';

class ResearchScreen extends StatelessWidget {
  const ResearchScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Исследования')),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          final planet = gameProvider.selectedPlanet;
          if (planet == null) return const Center(child: Text('Планета не выбрана'));

          return RefreshIndicator(
            onRefresh: () async => gameProvider.loadResearch(planet.id),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (gameProvider.researchPaused) ...[
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                      decoration: BoxDecoration(
                        color: Colors.orange.withValues(alpha: 0.15),
                        border: Border.all(color: Colors.orange.withValues(alpha: 0.4)),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Row(
                        children: [
                          const Icon(Icons.info_outline, color: Colors.orange, size: 20),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              'Исследования приостановлены — Центр исследований отключён',
                              style: const TextStyle(fontSize: 12, color: Colors.orange),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 16),
                  ],
                  _buildTreeSection(context, gameProvider),
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

    final completedIds = state.research.where((t) => t.completed).map((t) => t.techId).toSet();
    final inProgressIds = state.research.where((t) => t.inProgress).map((t) => t.techId).toSet();
    final availableIds = state.available.map((t) => t.techId).toSet();
    final researchMap = {for (var t in state.research) t.techId: t};
    final resources = gameProvider.selectedPlanet?.resources ?? {};
    final currentFood = (resources['food'] ?? 0) as num;
    final currentMoney = (resources['money'] ?? 0) as num;
    final currentAlien = (resources['alien_tech'] ?? 0) as num;
    final researchPaused = gameProvider.researchPaused;

    final availableTechs = <Widget>[];
    final completedTechs = <Widget>[];

    for (final tech in Constants.techList) {
      final techMap = Map<String, dynamic>.from(tech);
      final techId = techMap['id'] as String;
      final dependsOn = (techMap['depends_on'] as List).map((e) => e as String).toList();

      final isCompleted = completedIds.contains(techId);
      final isInProgress = inProgressIds.contains(techId);
      final isAvailable = availableIds.contains(techId);
      final hasPrerequisites = dependsOn.every((dep) => completedIds.contains(dep));

      if (!isCompleted && !isInProgress && !hasPrerequisites) continue;

      final costFood = (techMap['cost_food'] as num?)?.toDouble() ?? 0;
      final costMoney = (techMap['cost_money'] as num?)?.toDouble() ?? 0;
      final costAlien = (techMap['cost_alien_tech'] as num?)?.toDouble() ?? 0;
      final buildTime = (techMap['build_time'] as num?)?.toDouble() ?? 0;
      final canAfford = currentFood >= costFood && currentMoney >= costMoney && currentAlien >= costAlien;
      final maxLevel = ((techMap['max_level'] as num?)?.toInt()) ?? 1;

      final research = researchMap[techId];
      final progressPct = research != null ? research.progressPct : 0.0;
      final totalTime = research != null ? research.totalTime : 0.0;
      final progress = research != null ? research.progress : 0.0;
      final currentLevel = research != null ? research.level : 0;

      if (isCompleted) {
        completedTechs.add(
          _ResearchCard(
            techId: techId,
            name: techMap['name'] as String,
            description: techMap['description'] as String,
            statusColor: AppTheme.successColor,
            isCompleted: true,
            currentLevel: currentLevel,
            maxLevel: maxLevel,
            buildTime: buildTime,
          ),
        );
      } else if (isInProgress) {
        availableTechs.add(
          _ResearchCard(
            techId: techId,
            name: techMap['name'] as String,
            description: techMap['description'] as String,
            statusColor: AppTheme.warningColor,
            isCompleted: false,
            isInProgress: true,
            buildTime: buildTime,
            progressPct: progressPct,
            totalTime: totalTime,
            progress: progress,
            researchPaused: researchPaused,
          ),
        );
      } else if (isAvailable && hasPrerequisites) {
        availableTechs.add(
          _ResearchCard(
            techId: techId,
            name: techMap['name'] as String,
            description: techMap['description'] as String,
            statusColor: AppTheme.accentColor,
            isCompleted: false,
            isAvailable: true,
            canAfford: canAfford,
            costFood: costFood,
            costMoney: costMoney,
            costAlien: costAlien,
            buildTime: buildTime,
            onResearch: () => gameProvider.startResearch(techId),
            researchPaused: researchPaused,
          ),
        );
      }
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (availableTechs.isNotEmpty) ...[
          const Text('Доступные', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
          const SizedBox(height: 8),
          ...availableTechs,
          const SizedBox(height: 20),
        ],
        if (completedTechs.isNotEmpty) ...[
          const Text('Завершённые', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
          const SizedBox(height: 8),
          ...completedTechs,
        ],
        if (availableTechs.isEmpty && completedTechs.isEmpty)
          const Center(child: Text('Нет доступных исследований', style: TextStyle(color: Colors.white54))),
      ],
    );
  }
}

class _ResearchCard extends StatelessWidget {
  final String techId;
  final String name;
  final String description;
  final Color statusColor;
  final bool isCompleted;
  final bool isInProgress;
  final bool isAvailable;
  final bool canAfford;
  final int currentLevel;
  final int maxLevel;
  final double costFood;
  final double costMoney;
  final double costAlien;
  final double buildTime;
  final double progressPct;
  final double totalTime;
  final double progress;
  final bool researchPaused;
  final VoidCallback? onResearch;

  const _ResearchCard({
    required this.techId,
    required this.name,
    required this.description,
    required this.statusColor,
    required this.isCompleted,
    this.isInProgress = false,
    this.isAvailable = false,
    this.canAfford = true,
    this.currentLevel = 0,
    this.maxLevel = 1,
    this.costFood = 0,
    this.costMoney = 0,
    this.costAlien = 0,
    required this.buildTime,
    this.progressPct = 0,
    this.totalTime = 0,
    this.progress = 0,
    this.researchPaused = false,
    this.onResearch,
  });

  String _formatCost() {
    final parts = <String>[];
    if (costFood > 0) parts.add('🍍${costFood.toInt()}');
    if (costMoney > 0) parts.add('💰${costMoney.toInt()}');
    if (costAlien > 0) parts.add('👽${costAlien.toInt()}');
    return parts.join(' ');
  }

  String _formatRemainingTime(double seconds) {
    if (seconds <= 0) return '0 сек';
    final mins = (seconds / 60).floor();
    final secs = (seconds % 60).floor();
    if (mins > 0) {
      return '${mins} мин ${secs.toString().padLeft(2, '0')} сек';
    }
    return '${secs} сек';
  }

  @override
  Widget build(BuildContext context) {
    final isDimmed = researchPaused && !isCompleted;
    final textColor = isCompleted ? AppTheme.successColor : (isDimmed ? Colors.white38 : Colors.white);
    final cardColor = isCompleted ? statusColor.withValues(alpha: 0.1) : (isDimmed ? Colors.white.withValues(alpha: 0.03) : statusColor.withValues(alpha: 0.1));
    final borderColor = isCompleted ? statusColor : (isDimmed ? Colors.white24 : statusColor);
    final iconColor = isCompleted ? Colors.white : (isDimmed ? Colors.white38 : Colors.white);

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      decoration: BoxDecoration(
        border: Border.all(color: borderColor, width: 2),
        borderRadius: BorderRadius.circular(8),
        color: cardColor,
      ),
      child: ListTile(
        dense: true,
        leading: CircleAvatar(
          radius: 12,
          backgroundColor: isDimmed ? Colors.white24 : statusColor,
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
            color: textColor,
            fontWeight: (isAvailable || isInProgress) && !isDimmed ? FontWeight.bold : FontWeight.normal,
          ),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(description, style: TextStyle(fontSize: 11, color: isDimmed ? Colors.white24 : Colors.white54)),
            if (isCompleted) ...[
              const SizedBox(height: 2),
              Text(
                'Уровень ${currentLevel}/${maxLevel}',
                style: TextStyle(fontSize: 11, color: AppTheme.successColor),
              ),
              if (currentLevel > 0 && buildTime > 0) ...[
                const SizedBox(height: 2),
                Text('⏱ ${(buildTime / 60).toStringAsFixed(1)} мин', style: TextStyle(fontSize: 10, color: isDimmed ? Colors.white24 : Colors.white70)),
              ],
            ],
            if (isInProgress) ...[
              const SizedBox(height: 2),
              Row(
                children: [
                  Expanded(
                    child: LinearProgressIndicator(
                      value: progressPct / 100,
                      minHeight: 6,
                      borderRadius: BorderRadius.circular(3),
                      color: researchPaused ? Colors.orange : AppTheme.warningColor,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(
                    '${progressPct.toStringAsFixed(0)}%',
                    style: const TextStyle(fontSize: 10, color: Colors.white70, fontWeight: FontWeight.bold),
                  ),
                ],
              ),
              Text(
                '${researchPaused ? '⏸' : '⏱'} ⏳ ${_formatRemainingTime(totalTime - progress)}',
                style: TextStyle(fontSize: 10, color: researchPaused ? Colors.orange : Colors.white70),
              ),
            ],
            if (!isCompleted && !isInProgress && isAvailable) ...[
              const SizedBox(height: 2),
              Text(
                _formatCost(),
                style: TextStyle(fontSize: 10, color: isDimmed ? Colors.white24 : Colors.white70),
              ),
              const SizedBox(height: 2),
              Text(
                '⏱ ${(buildTime / 60).toStringAsFixed(1)} мин',
                style: TextStyle(fontSize: 10, color: isDimmed ? Colors.white24 : Colors.white70),
              ),
            ],
          ],
        ),
        trailing: isAvailable && !isCompleted
            ? ElevatedButton(
                onPressed: canAfford && !isDimmed ? onResearch : null,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  backgroundColor: canAfford && !isDimmed ? AppTheme.accentColor : Colors.white24,
                ),
                child: Text(canAfford && !isDimmed ? 'Исследовать' : 'Нет ресурсов', style: TextStyle(fontSize: 11, color: canAfford && !isDimmed ? Colors.white : Colors.white54)),
              )
            : null,
      ),
    );
  }
}

class Constants {
  static const techList = [
    {'id': 'planet_exploration', 'name': '🌍Разведка планеты', 'description': 'Открывает здание Фабрики', 'cost_food': 100, 'cost_money': 100, 'build_time': 60, 'max_level': 1, 'depends_on': []},
    {'id': 'energy_storage', 'name': '🔋Аккумуляторы', 'description': 'Открывает здание Аккумулятора', 'cost_food': 200, 'cost_money': 150, 'build_time': 90, 'max_level': 5, 'depends_on': []},
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

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

    for (final tech in state.research) {
      final techId = tech.techId;
      final dependsOn = tech.dependsOn;

      final isCompleted = completedIds.contains(techId);
      final isInProgress = inProgressIds.contains(techId);
      final isAvailable = availableIds.contains(techId);
      final hasPrerequisites = dependsOn.every((dep) => completedIds.contains(dep));

      if (!isCompleted && !isInProgress && !hasPrerequisites) continue;

      final costFood = tech.costFood;
      final costMoney = tech.costMoney;
      final costAlien = tech.costAlienTech;
      final buildTime = tech.buildTime;
      final canAfford = currentFood >= costFood && currentMoney >= costMoney && currentAlien >= costAlien;
      final maxLevel = tech.maxLevel;

      final research = researchMap[techId];
      final progressPct = research != null ? research.progressPct : 0.0;
      final totalTime = research != null ? research.totalTime : 0.0;
      final progress = research != null ? research.progress : 0.0;
      final currentLevel = research != null ? research.level : 0;

      if (isCompleted) {
        completedTechs.add(
          _ResearchCard(
            techId: techId,
            name: tech.name,
            description: tech.description,
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
            name: tech.name,
            description: tech.description,
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
            name: tech.name,
            description: tech.description,
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

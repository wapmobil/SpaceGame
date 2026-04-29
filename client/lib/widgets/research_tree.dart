import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../models/research.dart';

class ResearchTree extends StatelessWidget {
  final List<ResearchTech> research;
  final List<ResearchTech> available;
  final Function(String techId) onResearch;
  final double currentFood;
  final double currentMoney;
  final double currentAlien;

  const ResearchTree({
    super.key,
    required this.research,
    required this.available,
    required this.onResearch,
    this.currentFood = 0,
    this.currentMoney = 0,
    this.currentAlien = 0,
  });

  bool _canAfford(double currentFood, double currentMoney, double currentAlien, double costFood, double costMoney, double costAlien) {
    return currentFood >= costFood && currentMoney >= costMoney && currentAlien >= costAlien;
  }

  @override
  Widget build(BuildContext context) {
    if (research.isEmpty && available.isEmpty) {
      return const Center(child: Text('Нет данных об исследованиях', style: TextStyle(color: Colors.white38)));
    }

    final availableIds = available.map((t) => t.techId).toSet();
    final completedIds = research.where((t) => t.completed).map((t) => t.techId).toSet();
    final inProgressIds = research.where((t) => t.inProgress).map((t) => t.techId).toSet();
    final researchMap = {for (var t in research) t.techId: t};

    final techList = research.map((t) => {
      'id': t.techId,
      'name': t.name,
      'description': t.description,
      'depends_on': t.dependsOn,
      'cost_food': t.costFood,
      'cost_money': t.costMoney,
      'cost_alien_tech': t.costAlienTech,
      'build_time': t.buildTime,
      'max_level': t.maxLevel,
    }).toList();

    return Column(
      children: _buildTree(context, techList, completedIds, inProgressIds, availableIds, researchMap, onResearch, null, 0, currentFood, currentMoney, currentAlien),
    );
  }

  List<Widget> _buildTree(
    BuildContext context,
    List techList,
    Set<String> completedIds,
    Set<String> inProgressIds,
    Set<String> availableIds,
    Map<String, ResearchTech> researchMap,
    Function(String) onResearch,
    String? parentId,
    int depth,
    double currentFood,
    double currentMoney,
    double currentAlien,
  ) {
    final children = <Widget>[];

    for (final tech in techList) {
      final techMap = Map<String, dynamic>.from(tech);
      final techId = techMap['id'] as String;
      final dependsOn = (techMap['depends_on'] as List).map((e) => e as String).toList();

      // Filter by parent or show root nodes
      if (parentId != null) {
        if (!dependsOn.contains(parentId)) continue;
      } else {
        if (dependsOn.isNotEmpty) continue;
      }

      final isCompleted = completedIds.contains(techId);
      final isInProgress = inProgressIds.contains(techId);
      final isAvailable = availableIds.contains(techId);
      final hasPrerequisites = dependsOn.every((dep) => completedIds.contains(dep));

      // Hide if not visible (not completed, not in progress, prerequisites not met)
      if (!isCompleted && !isInProgress && !hasPrerequisites) continue;

      Color statusColor;
      if (isCompleted) statusColor = AppTheme.successColor;
      else if (isInProgress) statusColor = AppTheme.warningColor;
      else if (isAvailable && hasPrerequisites) statusColor = AppTheme.accentColor;
      else statusColor = Colors.white24;

      final research = researchMap[techId];
      final progressPct = research != null ? research.progressPct : 0.0;
      final totalTime = research != null ? research.totalTime : 0.0;
      final progress = research != null ? research.progress : 0.0;

      final costFood = (techMap['cost_food'] as num?)?.toDouble() ?? 0;
      final costMoney = (techMap['cost_money'] as num?)?.toDouble() ?? 0;
      final costAlien = (techMap['cost_alien_tech'] as num?)?.toDouble() ?? 0;
      final canAfford = _canAfford(currentFood, currentMoney, currentAlien, costFood, costMoney, costAlien);

      children.add(
        Padding(
          padding: EdgeInsets.only(left: depth * 24.0, bottom: 4),
          child: _TechNode(
            techId: techId,
            name: techMap['name'] as String,
            description: techMap['description'] as String,
            dependsOn: dependsOn.toList(),
            statusColor: statusColor,
            isCompleted: isCompleted,
            isInProgress: isInProgress,
            isAvailable: isAvailable && hasPrerequisites,
            canAfford: canAfford,
            onResearch: () => onResearch(techId),
            costFood: costFood,
            costMoney: costMoney,
            costAlien: costAlien,
            buildTime: (techMap['build_time'] as num?)?.toDouble() ?? 0,
            progressPct: progressPct,
            totalTime: totalTime,
            progress: progress,
          ),
        ),
      );

      // Recursively render children
      final subChildren = _buildTree(context, techList, completedIds, inProgressIds, availableIds, researchMap, onResearch, techId, depth + 1, currentFood, currentMoney, currentAlien);
      children.addAll(subChildren);
    }

    return children;
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
  final bool canAfford;
  final VoidCallback onResearch;
  final double costFood;
  final double costMoney;
  final double costAlien;
  final double buildTime;
  final double progressPct;
  final double totalTime;
  final double progress;

  const _TechNode({
    required this.techId,
    required this.name,
    required this.description,
    required this.dependsOn,
    required this.statusColor,
    required this.isCompleted,
    required this.isInProgress,
    required this.isAvailable,
    required this.canAfford,
    required this.onResearch,
    required this.costFood,
    required this.costMoney,
    required this.costAlien,
    required this.buildTime,
    required this.progressPct,
    required this.totalTime,
    required this.progress,
  });

  String _formatCost() {
    final parts = <String>[];
    if (costFood > 0) parts.add('🍞${costFood.toInt()}');
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
              const SizedBox(height: 2),
              Row(
                children: [
                  Expanded(
                    child: LinearProgressIndicator(
                      value: progressPct / 100,
                      minHeight: 6,
                      borderRadius: BorderRadius.circular(3),
                      color: AppTheme.warningColor,
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
                '⏱ ⏳ ${_formatRemainingTime(totalTime - progress)}',
                style: const TextStyle(fontSize: 10, color: Colors.white70),
              ),
            ],
            if (!isCompleted && !isInProgress) ...[
              const SizedBox(height: 2),
              Text(
                _formatCost(),
                style: const TextStyle(fontSize: 10, color: Colors.white70),
              ),
              const SizedBox(height: 2),
              Text(
                '⏱ ${(buildTime / 60).toStringAsFixed(1)} мин',
                style: const TextStyle(fontSize: 10, color: Colors.white70),
              ),
            ],
          if (dependsOn.isNotEmpty && !isCompleted && !isAvailable)
               const Text(
                 '🔒 Заблокировано',
                 style: TextStyle(fontSize: 10, color: Colors.white24),
               ),
          ],
        ),
        trailing: isAvailable && !isCompleted
            ? ElevatedButton(
                onPressed: canAfford ? onResearch : null,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  backgroundColor: canAfford ? AppTheme.accentColor : Colors.white24,
                ),
                child: Text(canAfford ? 'Исследовать' : 'Нет ресурсов', style: TextStyle(fontSize: 11, color: canAfford ? Colors.white : Colors.white54)),
              )
            : null,
      ),
    );
  }
}

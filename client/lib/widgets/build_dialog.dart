import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/building.dart';
import '../providers/game_provider.dart';

class BuildDialog extends StatelessWidget {
  final GameProvider gameProvider;

  const BuildDialog({super.key, required this.gameProvider});

  @override
  Widget build(BuildContext context) {
    final allBuildings = Constants.buildingTypes.keys.toList();
    final allBuildingsList = gameProvider.buildings;
    final hasFarm = allBuildingsList.where((b) => b.type == 'farm').any((b) => !b.pending && b.level > 0);
    final hasSolar = allBuildingsList.where((b) => b.type == 'solar').any((b) => !b.pending && b.level > 0);

    return Dialog(
      backgroundColor: AppTheme.cardColor,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Container(
        constraints: BoxConstraints(maxHeight: MediaQuery.of(context).size.height * 0.7),
        padding: const EdgeInsets.all(16),
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Построить сооружение',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white),
              ),
              const SizedBox(height: 4),
              Text(
                'Строительство: ${gameProvider.activeConstructions}/${gameProvider.maxConstructions}',
                style: TextStyle(fontSize: 12, color: Colors.white54),
              ),
              if (gameProvider.activeConstructions >= gameProvider.maxConstructions) ...[
                const SizedBox(height: 4),
                Text(
                  'Исследуйте "Параллельное строительство", чтобы открыть больше.',
                  style: TextStyle(fontSize: 10, color: Colors.orange),
                ),
              ],
              if (gameProvider.errorMessage != null && gameProvider.errorMessage!.isNotEmpty) ...[
                const SizedBox(height: 4),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
                  decoration: BoxDecoration(
                    color: Colors.red.withValues(alpha: 0.15),
                    border: Border.all(color: Colors.red.withValues(alpha: 0.4)),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    gameProvider.errorMessage!,
                    style: const TextStyle(fontSize: 11, color: Colors.red),
                  ),
                ),
                const SizedBox(height: 4),
              ],
              const SizedBox(height: 8),
              ...allBuildings.where((key) {
                if (key != 'farm' && !hasFarm) return false;
                if (key != 'farm' && key != 'solar' && !hasSolar) return false;
                return true;
              }).toList().map((key) => _buildItem(context, key, allBuildingsList)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildItem(BuildContext context, String key, List<Building> allBuildingsList) {
    final info = Constants.buildingTypes[key]!;
    final existing = allBuildingsList.where((b) => b.type == key).toList();
    final currentLevel = existing.isNotEmpty ? existing.first.level : 0;
    final isBuilding = existing.isNotEmpty && existing.first.buildTime > 0 && existing.first.buildProgress > 0 && existing.first.buildProgress <= existing.first.buildTime;
    final isPending = existing.isNotEmpty && existing.first.pending == true && existing.first.buildProgress <= 0;
    double nextCostFood, nextCostMoney;
    if (existing.isNotEmpty) {
      nextCostFood = existing.first.nextCostFood;
      nextCostMoney = existing.first.nextCostMoney;
    } else {
      final cost = gameProvider.buildingCosts[key];
      nextCostFood = cost?['food'] ?? 0;
      nextCostMoney = cost?['money'] ?? 0;
    }
    final canAfford = gameProvider.selectedPlanet != null &&
         ((gameProvider.selectedPlanet!.resources['food'] ?? 0) as num).toDouble() >= nextCostFood &&
         ((gameProvider.selectedPlanet!.resources['money'] ?? 0) as num).toDouble() >= nextCostMoney;

    return ListTile(
      leading: Text(info['icon'] as String, style: const TextStyle(fontSize: 24)),
      title: Text(info['name'] as String, style: const TextStyle(color: Colors.white)),
      subtitle: Text(
        isBuilding
            ? 'Строится... Ур.$currentLevel'
            : isPending
                ? 'Ожидает подтверждения - нажмите на здание, чтобы забрать'
                : existing.isNotEmpty
                    ? 'Ур.$currentLevel → ${currentLevel + 1} | 🍖$nextCostFood 💰$nextCostMoney'
                    : '${info['description'] as String} | 🍖$nextCostFood 💰$nextCostMoney',
      ),
      enabled: !isBuilding && !isPending && canAfford && gameProvider.activeConstructions < gameProvider.maxConstructions,
      onTap: () {
        Navigator.pop(context);
        gameProvider.buildStructure(key);
      },
    );
  }
}

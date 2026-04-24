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
    final hasFarm = allBuildingsList.where((b) => b.type == 'farm').any((b) => b.level > 0);
    final hasSolar = allBuildingsList.where((b) => b.type == 'solar').any((b) => b.level > 0);

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
              if (gameProvider.researchUnlocks.isNotEmpty) ...[
                const SizedBox(height: 4),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
                  decoration: BoxDecoration(
                    color: Colors.green.withValues(alpha: 0.15),
                    border: Border.all(color: Colors.green.withValues(alpha: 0.4)),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    'Исследование планет: доступен ${Constants.buildingNames[gameProvider.researchUnlocks] ?? gameProvider.researchUnlocks}',
                    style: const TextStyle(fontSize: 11, color: Colors.green),
                  ),
                ),
                const SizedBox(height: 4),
              ],
              const SizedBox(height: 8),
              ...allBuildings.where((key) {
                 if (key != 'farm' && !hasFarm) return false;
                 if (key != 'farm' && key != 'solar' && !hasSolar) return false;
                 final existing = allBuildingsList.where((b) => b.type == key).toList();
                 if (existing.isNotEmpty && existing.first.isBuilding) return false;
                 final req = Constants.researchRequirements[key];
                 if (req != null) {
                   final isUnlocked = gameProvider.researchState?.research
                           .any((r) => r.techId == req && r.completed) ??
                       false;
                   if (!isUnlocked) return false;
                 }
                 if (Constants.researchRandomUnlockBuildings.contains(key)) {
                   if (gameProvider.researchUnlocks.isEmpty ||
                       gameProvider.researchUnlocks != key) {
                     return false;
                   }
                 }
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
    final isBuilding = existing.isNotEmpty && existing.first.isBuilding;
    final isReady = existing.isNotEmpty && existing.first.isBuildComplete;
   double nextCostFood, nextCostIron, nextCostMoney;
    if (existing.isNotEmpty) {
      nextCostFood = existing.first.nextCostFood;
      nextCostIron = existing.first.nextCostIron;
      nextCostMoney = existing.first.nextCostMoney;
    } else {
      final cost = gameProvider.buildingCosts[key];
      nextCostFood = cost?['food'] ?? 0;
      nextCostIron = cost?['iron'] ?? 0;
      nextCostMoney = cost?['money'] ?? 0;
    }
    final canAfford = gameProvider.selectedPlanet != null &&
          ((gameProvider.selectedPlanet!.resources['food'] ?? 0) as num).toDouble() >= nextCostFood &&
          ((gameProvider.selectedPlanet!.resources['iron'] ?? 0) as num).toDouble() >= nextCostIron &&
          ((gameProvider.selectedPlanet!.resources['money'] ?? 0) as num).toDouble() >= nextCostMoney;

    return ListTile(
      leading: Text(info['icon'] as String, style: const TextStyle(fontSize: 24)),
      title: Text(info['name'] as String, style: const TextStyle(color: Colors.white)),
      subtitle: _buildSubtitle(context, key, info, existing, currentLevel, isBuilding, isReady, nextCostFood, nextCostIron, nextCostMoney, canAfford),
      enabled: !isBuilding && !isReady && canAfford && gameProvider.activeConstructions < gameProvider.maxConstructions,
      onTap: () {
        Navigator.pop(context);
        gameProvider.buildStructure(key);
      },
    );
  }

  Widget _buildSubtitle(BuildContext context, String key, Map<String, dynamic> info, List<Building> existing, int currentLevel, bool isBuilding, bool isReady, double costFood, double costIron, double costMoney, bool canAfford) {
    if (isBuilding) {
      return Text('Строится... Ур.$currentLevel', style: const TextStyle(fontSize: 10, color: Colors.orange));
    }
    if (isReady) {
      return const Text('Ожидает подтверждения - нажмите на здание', style: TextStyle(fontSize: 10, color: AppTheme.accentColor));
    }

    if (existing.isNotEmpty) {
      final deltas = _getDeltas(existing.first);
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            children: [
              Text('Ур.$currentLevel → ${currentLevel + 1}', style: const TextStyle(fontSize: 10, color: Colors.white70)),
              if (deltas.isNotEmpty) ...[
                const SizedBox(width: 8),
                Wrap(
                  spacing: 4,
                  runSpacing: 4,
                  children: deltas.map((d) => _deltaChip(d['value'] as double, d['text'] as String)).toList(),
                ),
              ],
            ],
          ),
          const SizedBox(height: 4),
          _bigCostChip(costFood, costIron, costMoney, canAfford),
        ],
      );
    }

    final cost = gameProvider.buildingCosts[key] ?? {};
    final prod = cost['production'] as Map<String, dynamic>? ?? {};
    final desc = info['description'] as String;
    final prodParts = <String>[];
    num? numVal;
    numVal = prod['food'];
    if (numVal != null) {
      final v = numVal.toDouble();
      if (v.abs() > 0.01) prodParts.add('${v >= 0 ? '+' : ''}🍖${v.toInt()}');
    }
    numVal = prod['iron'];
    if (numVal != null) {
      final v = numVal.toDouble();
      if (v.abs() > 0.01) prodParts.add('${v >= 0 ? '+' : ''}⛏️${v.toInt()}');
    }
    numVal = prod['energy'];
    if (numVal != null) {
      final v = numVal.toDouble();
      if (v.abs() > 0.01) prodParts.add('${v >= 0 ? '+' : ''}⚡${v.toInt()}');
    }
    numVal = prod['composite'];
    if (numVal != null) {
      final v = numVal.toDouble();
      if (v.abs() > 0.01) prodParts.add('${v >= 0 ? '+' : ''}🧬${v.toInt()}');
    }
    numVal = prod['mechanisms'];
    if (numVal != null) {
      final v = numVal.toDouble();
      if (v.abs() > 0.01) prodParts.add('${v >= 0 ? '+' : ''}⚙️${v.toInt()}');
    }
    numVal = prod['reagents'];
    if (numVal != null) {
      final v = numVal.toDouble();
      if (v.abs() > 0.01) prodParts.add('${v >= 0 ? '+' : ''}🧪${v.toInt()}');
    }
    final prodText = prodParts.join('  ');

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(desc, style: const TextStyle(fontSize: 10, color: Colors.white70)),
        if (prodText.isNotEmpty) const SizedBox(height: 2),
        if (prodText.isNotEmpty) Text(prodText, style: const TextStyle(fontSize: 10, color: Colors.white54)),
        const SizedBox(height: 4),
        _bigCostChip(costFood, costIron, costMoney, canAfford),
      ],
    );
  }

   Widget _bigCostChip(double food, double iron, double money, bool canAfford) {
    final color = canAfford ? Colors.amber : Colors.red;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.withValues(alpha: 0.4), width: 1.5),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (food > 0) ...[
            const Text('🍖', style: TextStyle(fontSize: 14)),
            const SizedBox(width: 4),
            Text(
              food.toInt().toString(),
              style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: color),
            ),
          ],
          if (food > 0 && iron > 0) const SizedBox(width: 12),
          if (iron > 0) ...[
            const Text('⛏️', style: TextStyle(fontSize: 14)),
            const SizedBox(width: 4),
            Text(
              iron.toInt().toString(),
              style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: color),
            ),
          ],
          if ((food > 0 || iron > 0) && money > 0) const SizedBox(width: 12),
          if (money > 0) ...[
            const Text('💰', style: TextStyle(fontSize: 14)),
            const SizedBox(width: 4),
            Text(
              money.toInt().toString(),
              style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: color),
            ),
          ],
        ],
      ),
    );
  }

  List<Map<String, dynamic>> _getDeltas(Building building) {
    final result = <Map<String, dynamic>>[];
    if (building.deltaFood.abs() > 0.01) {
      result.add({'text': '${building.deltaFood > 0 ? '+' : ''}🍖${building.deltaFood.toInt()}', 'value': building.deltaFood});
    }
    if (building.deltaIron.abs() > 0.01) {
      result.add({'text': '${building.deltaIron > 0 ? '+' : ''}⛏️${building.deltaIron.toInt()}', 'value': building.deltaIron});
    }
    if (building.deltaEnergy.abs() > 0.01) {
      result.add({'text': '${building.deltaEnergy > 0 ? '+' : ''}⚡${building.deltaEnergy.toInt()}', 'value': building.deltaEnergy});
    }
    if (building.deltaComposite.abs() > 0.01) {
      result.add({'text': '${building.deltaComposite > 0 ? '+' : ''}🧬${building.deltaComposite.toInt()}', 'value': building.deltaComposite});
    }
    if (building.deltaMechanisms.abs() > 0.01) {
      result.add({'text': '${building.deltaMechanisms > 0 ? '+' : ''}⚙️${building.deltaMechanisms.toInt()}', 'value': building.deltaMechanisms});
    }
    if (building.deltaReagents.abs() > 0.01) {
      result.add({'text': '${building.deltaReagents > 0 ? '+' : ''}🧪${building.deltaReagents.toInt()}', 'value': building.deltaReagents});
    }
    return result;
  }

  Widget _deltaChip(double value, String text) {
    final isNegative = value < 0;
    final color = isNegative ? Colors.red : Colors.green;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Text(text, style: TextStyle(fontSize: 9, color: color)),
    );
  }
}

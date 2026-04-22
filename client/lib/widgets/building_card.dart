import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/building.dart';
import '../providers/game_provider.dart';

class BuildingCard extends StatelessWidget {
  final Building building;
  final GameProvider gameProvider;
  final VoidCallback? onTap;

  const BuildingCard({
    super.key,
    required this.building,
    required this.gameProvider,
    this.onTap,
  });

  bool get isPending => building.pending == true && building.buildProgress <= 0;
  bool get isBuilding => building.buildTime > 0 && building.buildProgress > 0 && building.buildProgress <= building.buildTime && !isPending;
  bool get energyDeficit => gameProvider.energyBufferDeficit;

  String? get statusText {
    if (isPending) return 'Tap to claim!';
    if (isBuilding) {
      final remaining = (building.buildTime - building.buildProgress).toInt();
      return 'Building... ${remaining}s';
    }
    if (energyDeficit) return '⚠ Energy deficit';
    if (building.level == 0) return 'Not built';
    return 'Operational';
  }

  Color? get statusColor {
    if (isPending) return AppTheme.accentColor;
    if (isBuilding) return Colors.orange;
    if (energyDeficit) return Colors.red;
    if (building.level == 0) return Colors.white54;
    return Colors.green;
  }

  List<Widget> get prodLines {
    final lines = <Widget>[];
    if (building.productionFood.abs() > 0.01) {
      lines.add(_prodRow('🍖', building.productionFood));
    }
    if (building.productionEnergy.abs() > 0.01) {
      lines.add(_prodRow('⚡', building.productionEnergy));
    }
    if (building.productionComposite.abs() > 0.01) {
      lines.add(_prodRow('🧬', building.productionComposite));
    }
    if (building.productionMechanisms.abs() > 0.01) {
      lines.add(_prodRow('⚙️', building.productionMechanisms));
    }
    if (building.productionReagents.abs() > 0.01) {
      lines.add(_prodRow('🧪', building.productionReagents));
    }
    if (building.consumption > 0 && !energyDeficit) {
      lines.add(Row(mainAxisSize: MainAxisSize.min, children: [
        const Text('⚡', style: TextStyle(fontSize: 11)),
        Text('-${building.consumption.toInt()}', style: const TextStyle(fontSize: 10, color: Colors.orange)),
      ]));
    }
    return lines;
  }

  Widget _prodRow(String icon, double value) {
    return Row(mainAxisSize: MainAxisSize.min, children: [
      Text(icon, style: const TextStyle(fontSize: 11)),
      Text('${value >= 0 ? "+" : ""}${value.toInt()}',
          style: TextStyle(fontSize: 10, color: value >= 0 ? Colors.green : Colors.red)),
    ]);
  }

  Widget? _buildUpgradeButton() {
    final nextCostFood = building.nextCostFood;
    final nextCostMoney = building.nextCostMoney;
    if (building.level == 0 && nextCostFood <= 0 && nextCostMoney <= 0) return null;

    final maxLevel = nextCostFood <= 0 && nextCostMoney <= 0;
    if (maxLevel) return null;

    final upgradeInfo = gameProvider.getBuildingUpgradeInfo(building);
    final canUpgrade = upgradeInfo.canUpgrade;
    final hasResources = nextCostFood <= 0 || nextCostMoney <= 0 ||
        (gameProvider.selectedPlanet != null &&
            ((gameProvider.selectedPlanet!.resources['food'] ?? 0) as num).toDouble() >= nextCostFood &&
            ((gameProvider.selectedPlanet!.resources['money'] ?? 0) as num).toDouble() >= nextCostMoney);

    return ElevatedButton(
      onPressed: (canUpgrade && hasResources) ? () => gameProvider.buildStructure(building.type) : null,
      style: ElevatedButton.styleFrom(
        backgroundColor: (canUpgrade && hasResources) ? AppTheme.accentColor : Colors.grey[700],
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        minimumSize: const Size(0, 28),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      ),
      child: Text(
        'Upgrade → ${building.level + 1}\n🍖${nextCostFood.toInt()} 💰${nextCostMoney.toInt()}',
        style: const TextStyle(fontSize: 9, color: Colors.white),
        textAlign: TextAlign.center,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final info = Constants.buildingTypes[building.type] ??
        {'name': building.type, 'icon': '🏗️', 'description': 'Unknown'};
    final name = info['name'] ?? building.type;
    final icon = info['icon'] ?? '🏗️';
    final upgradeBtn = _buildUpgradeButton();

    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.white.withValues(alpha: 0.15)),
        color: AppTheme.cardColor,
      ),
      child: InkWell(
        onTap: isPending ? () => gameProvider.confirmBuilding(building.type) : onTap,
        borderRadius: BorderRadius.circular(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 10, 12, 8),
              child: Row(
                children: [
                  Text(icon, style: const TextStyle(fontSize: 24)),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Text(name, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: Colors.white)),
                            const SizedBox(width: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                              decoration: BoxDecoration(
                                color: AppTheme.accentColor.withValues(alpha: 0.2),
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(
                                'Lv. ${building.level}',
                                style: const TextStyle(fontSize: 10, color: AppTheme.accentColor, fontWeight: FontWeight.bold),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 2),
                        Row(
                          children: [
                            Icon(Icons.circle, size: 8, color: statusColor),
                            const SizedBox(width: 4),
                            Text(statusText!, style: TextStyle(fontSize: 10, color: statusColor)),
                          ],
                        ),
                      ],
                    ),
                  ),
                  if (upgradeBtn != null) upgradeBtn,
                ],
              ),
            ),
            if (prodLines.isNotEmpty)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                child: Wrap(
                  spacing: 10,
                  runSpacing: 2,
                  children: prodLines,
                ),
              ),
            if (isBuilding)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 24),
                child: LinearProgressIndicator(
                  value: 1.0 - (building.buildProgress / building.buildTime),
                  minHeight: 3,
                  color: AppTheme.accentColor,
                  backgroundColor: Colors.grey[800],
                ),
              ),
          ],
        ),
      ),
    );
  }
}

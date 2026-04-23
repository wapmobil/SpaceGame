import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/building.dart';
import '../providers/game_provider.dart';

class BuildingCard extends StatefulWidget {
  final Building building;
  final GameProvider gameProvider;
  final VoidCallback? onTap;

  const BuildingCard({
    super.key,
    required this.building,
    required this.gameProvider,
    this.onTap,
  });

  @override
  State<BuildingCard> createState() => _BuildingCardState();
}

class _BuildingCardState extends State<BuildingCard> with SingleTickerProviderStateMixin {
  late final AnimationController _pulseController;
  late final AnimationController _opacityController;

  bool get isBuilding => widget.building.isBuilding;

  bool get isPending => widget.building.isBuildComplete;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 500),
    );
    _opacityController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 500),
    );
    if (isBuilding || isPending) {
      _pulseController.repeat(reverse: true);
      _opacityController.repeat(reverse: true);
    }
  }

  @override
  void didUpdateWidget(BuildingCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    final wasBuilding = oldWidget.building.isBuilding;
    final wasPending = oldWidget.building.isBuildComplete;
    if ((isBuilding || isPending) && !(wasBuilding || wasPending)) {
      _pulseController.repeat(reverse: true);
      _opacityController.repeat(reverse: true);
    } else if (!isBuilding && !isPending) {
      _pulseController.stop();
      _pulseController.value = 0;
      _opacityController.stop();
      _opacityController.value = 0;
    }
  }

  @override
  void dispose() {
    _pulseController.dispose();
    _opacityController.dispose();
    super.dispose();
  }

  String? get statusText {
    if (isPending) return 'Нажмите чтобы открыть';
    if (isBuilding) return 'Строится...';
    if (!widget.building.enabled) return 'Отключено';
    if (widget.building.level == 0) return 'Не построено';
    return 'Работает';
  }

  Color? get statusColor {
    if (isPending) return AppTheme.accentColor;
    if (isBuilding) return Colors.orange;
    if (!widget.building.enabled) return Colors.grey;
    if (widget.building.level == 0) return Colors.white54;
    return Colors.green;
  }

  List<Widget> get prodLines {
    final lines = <Widget>[];
    if (widget.building.productionFood.abs() > 0.01) {
      lines.add(_prodChip('🍖', widget.building.productionFood));
    }
    if (widget.building.productionEnergy.abs() > 0.01) {
      lines.add(_prodChip('⚡', widget.building.productionEnergy));
    }
    if (widget.building.productionComposite.abs() > 0.01) {
      lines.add(_prodChip('🧬', widget.building.productionComposite));
    }
    if (widget.building.productionMechanisms.abs() > 0.01) {
      lines.add(_prodChip('⚙️', widget.building.productionMechanisms));
    }
    if (widget.building.productionReagents.abs() > 0.01) {
      lines.add(_prodChip('🧪', widget.building.productionReagents));
    }
    if (widget.building.consumption > 0) {
      lines.add(_prodChip('⚡', -widget.building.consumption, isConsumption: true));
    }
    return lines;
  }

  Widget _prodChip(String icon, double value, {bool isConsumption = false}) {
    final color = isConsumption ? Colors.orange : (value >= 0 ? AppTheme.successColor : AppTheme.dangerColor);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(icon, style: const TextStyle(fontSize: 12)),
          const SizedBox(width: 4),
          Text(
            '${value >= 0 ? "+" : ""}${value.toInt()}',
            style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: color),
          ),
        ],
      ),
    );
  }

  Widget? _buildUpgradeButton() {
    final nextCostFood = widget.building.nextCostFood;
    final nextCostMoney = widget.building.nextCostMoney;
    if (widget.building.level == 0 && nextCostFood <= 0 && nextCostMoney <= 0) return null;

    final maxLevel = nextCostFood <= 0 && nextCostMoney <= 0;
    if (maxLevel) return null;

    final upgradeInfo = gameProvider.getBuildingUpgradeInfo(widget.building);
    final canUpgrade = upgradeInfo.canUpgrade;
    final hasResources = nextCostFood <= 0 || nextCostMoney <= 0 ||
        (gameProvider.selectedPlanet != null &&
            ((gameProvider.selectedPlanet!.resources['food'] ?? 0) as num).toDouble() >= nextCostFood &&
            ((gameProvider.selectedPlanet!.resources['money'] ?? 0) as num).toDouble() >= nextCostMoney);

    final hasAll = canUpgrade && hasResources;

    return Tooltip(
      message: hasResources ? '' : 'Не хватает ресурсов',
      child: ElevatedButton(
        onPressed: hasAll ? () => gameProvider.buildStructure(widget.building.type) : null,
        style: ElevatedButton.styleFrom(
          backgroundColor: hasAll ? AppTheme.accentColor : Colors.grey[700],
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          minimumSize: const Size(0, 28),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          elevation: 0,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('⬆️', style: TextStyle(fontSize: 10)),
            const SizedBox(width: 3),
            Text(
              'Lv.${widget.building.level + 1}\n🍖${nextCostFood.toInt()} 💰${nextCostMoney.toInt()}',
              style: const TextStyle(fontSize: 9, color: Colors.white, height: 1.2),
            ),
          ],
        ),
      ),
    );
  }

  GameProvider get gameProvider => widget.gameProvider;
  Building get building => widget.building;

  @override
  Widget build(BuildContext context) {
    final info = Constants.buildingTypes[building.type] ??
        {'name': building.type, 'icon': '🏗️', 'description': 'Неизвестно'};
    final name = info['name'] ?? building.type;
    final icon = info['icon'] ?? '🏗️';
    final upgradeBtn = _buildUpgradeButton();

    final remainingTime = isBuilding ? building.buildProgress.toInt().clamp(0, 999) : 0;
    final progressValue = building.buildTime > 0
        ? (building.buildProgress / building.buildTime).clamp(0.0, 1.0)
        : 0.0;

    final glow = isBuilding ? AppTheme.accentColor.withValues(alpha: 0.25) : Colors.transparent;
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: !building.enabled
              ? Colors.grey.withValues(alpha: 0.15)
              : isBuilding
                  ? AppTheme.accentColor.withValues(alpha: 0.4)
                  : AppTheme.accentColor.withValues(alpha: 0.2),
          width: isBuilding ? 1.5 : 1,
        ),
        color: !building.enabled ? Colors.black.withValues(alpha: 0.2) : AppTheme.cardColor,
        boxShadow: !building.enabled
            ? []
            : [
                BoxShadow(
                  color: glow,
                  blurRadius: isBuilding ? 12 : 6,
                  spreadRadius: isBuilding ? 2 : 0,
                ),
              ],
      ),
      child: InkWell(
        onTap: isPending ? () => gameProvider.confirmBuilding(building.type) : widget.onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header row
              Row(
                children: [
                  // Icon
                  Container(
                    width: 44,
                    height: 44,
                    decoration: BoxDecoration(
                      color: AppTheme.accentColor.withValues(alpha: 0.08),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Center(child: Text(icon, style: const TextStyle(fontSize: 22))),
                  ),
                  const SizedBox(width: 12),
                  // Name + level + status
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Text(name, style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: Colors.white)),
                            const SizedBox(width: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
                              decoration: BoxDecoration(
                                color: AppTheme.accentColor.withValues(alpha: 0.15),
                                borderRadius: BorderRadius.circular(6),
                                border: Border.all(color: AppTheme.accentColor.withValues(alpha: 0.3)),
                              ),
                              child: Text(
                                'Lv. ${building.level}',
                                style: const TextStyle(fontSize: 10, color: AppTheme.accentColor, fontWeight: FontWeight.bold),
                              ),
                            ),
                            if (upgradeBtn != null) ...[
                              const SizedBox(width: 6),
                              upgradeBtn,
                            ],
                          ],
                        ),
                        const SizedBox(height: 4),
                        Row(
                          children: [
                            _statusIndicator(),
                            const SizedBox(width: 5),
                            Text(
                              statusText!,
                              style: TextStyle(fontSize: 11, color: statusColor),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  // Toggle
                  if (!isBuilding && building.level > 0)
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Switch(
                          value: building.enabled,
                          onChanged: (_) => gameProvider.toggleBuilding(building.type),
                          activeThumbColor: AppTheme.accentColor,
                          inactiveThumbColor: Colors.grey,
                          inactiveTrackColor: Colors.grey[800],
                          materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                          trackOutlineColor: const WidgetStatePropertyAll(Colors.transparent),
                        ),
                      ],
                    ),
                ],
              ),
              // Separator
              const SizedBox(height: 10),
              Divider(color: Colors.white.withValues(alpha: 0.08), thickness: 1),
              const SizedBox(height: 10),
              // Production lines
              if (prodLines.isNotEmpty)
                Wrap(
                  spacing: 6,
                  runSpacing: 6,
                  children: prodLines,
                ),
              // Progress + remaining time
              if (isBuilding) ...[
                const SizedBox(height: 12),
                Row(
                  children: [
                    Text(
                      'Осталось: $remainingTimeс',
                      style: const TextStyle(fontSize: 11, color: AppTheme.accentColor, fontWeight: FontWeight.w500),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: 1.0 - progressValue,
                    minHeight: 6,
                    color: AppTheme.accentColor,
                    backgroundColor: Colors.grey[800]?.withValues(alpha: 0.5),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Widget _statusIndicator() {
    if (isBuilding || isPending) {
      return AnimatedBuilder(
        animation: Listenable.merge([_pulseController, _opacityController]),
        builder: (context, _) {
          final scale = 0.8 + 0.5 * _pulseController.value;
          final opacity = 0.4 + 0.6 * _opacityController.value;
          return Container(
            width: 10 * scale,
            height: 10 * scale,
            decoration: BoxDecoration(
              color: (statusColor ?? Colors.white).withValues(alpha: opacity),
              shape: BoxShape.circle,
            ),
          );
        },
      );
    }
    return Container(
      width: 8,
      height: 8,
      decoration: BoxDecoration(
        color: statusColor,
        shape: BoxShape.circle,
      ),
    );
  }
}

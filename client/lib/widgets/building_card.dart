import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/building.dart';
import '../providers/game_provider.dart';
import 'planet_action_chip.dart';

class BuildingCard extends StatefulWidget {
  final Building building;
  final GameProvider gameProvider;
  final VoidCallback? onTap;
  final void Function(String buildingType, String action)? onNavigateBuilding;

  const BuildingCard({
    super.key,
    required this.building,
    required this.gameProvider,
    this.onTap,
    this.onNavigateBuilding,
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
    if (widget.building.productionIron.abs() > 0.01) {
      lines.add(_prodChip('⛏️', widget.building.productionIron));
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
    final nextCostIron = widget.building.nextCostIron;
    final nextCostMoney = widget.building.nextCostMoney;
    if (widget.building.level == 0 && nextCostFood <= 0 && nextCostIron <= 0 && nextCostMoney <= 0) return null;

    final maxLevel = nextCostFood <= 0 && nextCostIron <= 0 && nextCostMoney <= 0;
    if (maxLevel) return null;

    final upgradeInfo = gameProvider.getBuildingUpgradeInfo(widget.building);
    final canUpgrade = upgradeInfo.canUpgrade;
    final hasResources = nextCostFood <= 0 || nextCostIron <= 0 || nextCostMoney <= 0 ||
        (gameProvider.selectedPlanet != null &&
            ((gameProvider.selectedPlanet!.resources['food'] ?? 0) as num).toDouble() >= nextCostFood &&
            ((gameProvider.selectedPlanet!.resources['iron'] ?? 0) as num).toDouble() >= nextCostIron &&
            ((gameProvider.selectedPlanet!.resources['money'] ?? 0) as num).toDouble() >= nextCostMoney);

    final hasAll = canUpgrade && hasResources;

    return Tooltip(
      message: hasResources ? '' : 'Не хватает ресурсов',
      child: ElevatedButton(
        onPressed: hasAll ? () => gameProvider.buildStructure(widget.building.type) : null,
        style: ElevatedButton.styleFrom(
          backgroundColor: hasAll ? Colors.amber.withValues(alpha: 0.2) : Colors.grey[800],
          foregroundColor: Colors.white,
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          minimumSize: const Size(0, 28),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8), side: BorderSide(color: hasAll ? Colors.amber.withValues(alpha: 0.4) : Colors.grey[700]!, width: 1)),
          elevation: 0,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('▲', style: TextStyle(fontSize: 11, color: Colors.white)),
            const SizedBox(width: 4),
            Flexible(
              child: Text(
                [
                  'Lv.${widget.building.level + 1}',
                  if (nextCostFood > 0) '🍖$nextCostFood',
                  if (nextCostIron > 0) '⛏️$nextCostIron',
                  if (nextCostMoney > 0) '💰$nextCostMoney',
                ].join('  '),
                style: const TextStyle(fontSize: 10, color: Colors.white),
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
    );
  }

  List<Map<String, dynamic>> get _deltas {
    if (widget.building.level == 0) return [];
    final result = <Map<String, dynamic>>[];

    if (widget.building.deltaFood.abs() > 0.01) {
      result.add({'text': '${widget.building.deltaFood > 0 ? '+' : ''}🍖${widget.building.deltaFood.toInt()}', 'value': widget.building.deltaFood});
    }
    if (widget.building.deltaIron.abs() > 0.01) {
      result.add({'text': '${widget.building.deltaIron > 0 ? '+' : ''}⛏️${widget.building.deltaIron.toInt()}', 'value': widget.building.deltaIron});
    }
    if (widget.building.deltaEnergy.abs() > 0.01) {
      result.add({'text': '${widget.building.deltaEnergy > 0 ? '+' : ''}⚡${widget.building.deltaEnergy.toInt()}', 'value': widget.building.deltaEnergy});
    }
    if (widget.building.deltaComposite.abs() > 0.01) {
      result.add({'text': '${widget.building.deltaComposite > 0 ? '+' : ''}🧬${widget.building.deltaComposite.toInt()}', 'value': widget.building.deltaComposite});
    }
    if (widget.building.deltaMechanisms.abs() > 0.01) {
      result.add({'text': '${widget.building.deltaMechanisms > 0 ? '+' : ''}⚙️${widget.building.deltaMechanisms.toInt()}', 'value': widget.building.deltaMechanisms});
    }
    if (widget.building.deltaReagents.abs() > 0.01) {
      result.add({'text': '${widget.building.deltaReagents > 0 ? '+' : ''}🧪${widget.building.deltaReagents.toInt()}', 'value': widget.building.deltaReagents});
    }

    return result;
  }

  Widget _deltaChip(Map<String, dynamic> d) {
    final isNegative = d['value'] < 0;
    final color = isNegative ? Colors.red : Colors.green;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Text(
        d['text']!,
        style: TextStyle(fontSize: 9, color: color),
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
                crossAxisAlignment: CrossAxisAlignment.start,
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
                  // Name + level + actions
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Flexible(
                              child: Text(
                                name,
                                style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: Colors.white),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
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
                  ],
                 ),
              // Upgrade button + deltas
                if (upgradeBtn != null || _deltas.isNotEmpty) ...[
                  const SizedBox(height: 6),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (upgradeBtn != null) ...[
                        upgradeBtn,
                        const SizedBox(width: 8),
                        if (_deltas.isNotEmpty)
                          Wrap(
                            spacing: 4,
                            runSpacing: 4,
                            children: _deltas.map((d) => _deltaChip(d)).toList(),
                          ),
                      ],
                    ],
                  ),
                ],
               // Toggle switch + production chips + navigation chips
                if ((!isBuilding && building.level > 0 && building.type != 'storage') || prodLines.isNotEmpty || (building.isWorking && widget.onNavigateBuilding != null)) ...[
                  const SizedBox(height: 6),
                  OverflowBar(
                    spacing: 6,
                    children: [
                      if (!isBuilding && building.level > 0 && building.type != 'storage')
                        Switch(
                          value: building.enabled,
                          onChanged: (_) => gameProvider.toggleBuilding(building.type),
                          activeThumbColor: AppTheme.accentColor,
                          inactiveThumbColor: Colors.grey,
                          inactiveTrackColor: Colors.grey[800],
                          materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                          trackOutlineColor: const WidgetStatePropertyAll(Colors.transparent),
                        ),
                      ...prodLines,
                      if (building.isWorking && widget.onNavigateBuilding != null)
                        ..._buildNavigationChips(building.type),
                    ],
                  ),
                ],
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

  List<Widget> _buildNavigationChips(String type) {
    final chips = <Map<String, dynamic>>[];
    switch (type) {
      case 'base':
        chips.add({'icon': Icons.science, 'label': 'Исследования', 'action': 'research'});
        break;
      case 'shipyard':
        chips.add({'icon': Icons.rocket_launch, 'label': 'Верфь', 'action': 'shipyard'});
        break;
      case 'comcenter':
        chips.add({'icon': Icons.explore, 'label': 'Экспедиция', 'action': 'expedition'});
        break;
            case 'mine':
        chips.add({'icon': Icons.dns_outlined, 'label': 'Бурение', 'action': 'drill'});
        break;
      case 'market':
        chips.add({'icon': Icons.store, 'label': 'Рынок', 'action': 'market'});
        break;
      case 'farm':
        chips.add({'icon': Icons.eco, 'label': 'Грядки', 'action': 'farm'});
        break;
    }
    return chips.map((chip) {
      return PlanetActionChip(
        icon: chip['icon'] as IconData,
        label: chip['label'] as String,
        onTap: () => widget.onNavigateBuilding?.call(type, chip['action'] as String),
      );
    }).toList();
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

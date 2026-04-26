import 'dart:async';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/garden_bed_provider.dart';
import '../providers/game_provider.dart';
import '../models/garden_bed.dart';
import '../core/app_theme.dart';

class GardenBedScreen extends StatefulWidget {
  final String planetId;

  const GardenBedScreen({super.key, required this.planetId});

  @override
  State<GardenBedScreen> createState() => _GardenBedScreenState();
}

class _GardenBedScreenState extends State<GardenBedScreen> {
  Timer? _ticker;

  @override
  void initState() {
    super.initState();
    _ticker = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted) setState(() {});
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final fp = context.read<GameProvider>().gardenBedProvider;
      fp.clearError();
      fp.getGardenBed(widget.planetId);
    });
  }

  @override
  void dispose() {
    _ticker?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        title: const Text('Грядки'),
        backgroundColor: AppTheme.cardColor,
        foregroundColor: Colors.white,
        elevation: 0,
      ),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          final gardenBedProvider = gameProvider.gardenBedProvider;
          return ListenableBuilder(
            listenable: gardenBedProvider,
            builder: (context, _) {
              if (gardenBedProvider.isLoading && gardenBedProvider.gardenBedState == null) {
                return const Center(child: CircularProgressIndicator());
              }

              final gardenBedState = gardenBedProvider.gardenBedState;
              if (gardenBedState == null) {
                return const Center(child: Text('Грядки не построены'));
              }
              if (gardenBedState.rows.isEmpty) {
                return const Center(child: Text('Грядки не построены'));
              }

              final farmLevel = gameProvider.getBuildingLevelForPlanet(widget.planetId, "farm");

              return Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Error banner
                    if (gardenBedProvider.errorMessage != null)
                      Container(
                        margin: const EdgeInsets.only(bottom: 16),
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: AppTheme.dangerColor.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(color: AppTheme.dangerColor.withValues(alpha: 0.3)),
                        ),
                        child: Row(
                          children: [
                            const Icon(Icons.error_outline, color: AppTheme.dangerColor, size: 20),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                gardenBedProvider.errorMessage!,
                                style: const TextStyle(color: AppTheme.dangerColor, fontSize: 13),
                              ),
                            ),
                            IconButton(
                              icon: const Icon(Icons.close, size: 18, color: AppTheme.dangerColor),
                              onPressed: () => gardenBedProvider.clearError(),
                            ),
                          ],
                        ),
                      ),

                    // Info bar
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: AppTheme.accentColor.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: AppTheme.accentColor.withValues(alpha: 0.3)),
                      ),
                      child: Row(
                        children: [
                          const Icon(Icons.eco_outlined, color: AppTheme.accentColor, size: 20),
                          const SizedBox(width: 8),
                          Text(
                            'Ур. $farmLevel • ${gardenBedState.rowCount} рядов',
                            style: const TextStyle(color: AppTheme.accentColor, fontWeight: FontWeight.w600),
                          ),
                          const Spacer(),
                          if (!gardenBedProvider.canAct)
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                              decoration: BoxDecoration(
                                color: Colors.orange.withValues(alpha: 0.2),
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                '${gardenBedProvider.remainingCooldown}с',
                                style: const TextStyle(color: Colors.orange, fontSize: 12, fontWeight: FontWeight.w600),
                              ),
                            ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 12),

                    // Resources bar
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.white.withValues(alpha: 0.05),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
                      ),
                      child: Row(
                        children: [
                          _buildResourceChip('💰', (gameProvider.selectedPlanet?.resources['money'] as num?)?.toDouble() ?? 0),
                          const SizedBox(width: 16),
                          _buildResourceChip('🍍', (gameProvider.selectedPlanet?.resources['food'] as num?)?.toDouble() ?? 0),
                        ],
                      ),
                    ),
                    const SizedBox(height: 16),

                    // Farm rows
                    Expanded(
                      child: ListView.builder(
                        itemCount: gardenBedState.rows.length,
                        itemBuilder: (context, rowIndex) {
                          return _buildRowCard(context, gardenBedProvider, gameProvider, gardenBedState.rows[rowIndex], rowIndex, farmLevel);
                        },
                      ),
                    ),
                  ],
                ),
              );
            },
          );
        },
      ),
    );
  }

  Widget _buildResourceChip(String icon, double value) {
    return Row(
      children: [
        Text(icon, style: const TextStyle(fontSize: 16)),
        const SizedBox(width: 4),
        Text(
          value.toInt().toString(),
          style: const TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.w600),
        ),
      ],
    );
  }

  Widget _buildRowCard(BuildContext context, GardenBedProvider gardenBedProvider, GameProvider gameProvider, GardenBedRow row, int rowIndex, int farmLevel) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: AppTheme.cardColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: row.isMature
              ? AppTheme.successColor.withValues(alpha: 0.4)
              : row.isWithered
                  ? AppTheme.dangerColor.withValues(alpha: 0.3)
                  : Colors.white.withValues(alpha: 0.08),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(12),
            child: Row(
              children: [
                // Row number
                Container(
                  width: 28,
                  height: 28,
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.08),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Center(
                    child: Text(
                      '${rowIndex + 1}',
                      style: const TextStyle(color: Colors.white70, fontSize: 13, fontWeight: FontWeight.w600),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                // Plant info
                Expanded(
                  child: _buildPlantInfo(context, gardenBedProvider, row),
                ),
              ],
            ),
          ),
          // Action buttons
          if (row.isEmpty) ...[
            const Divider(height: 1, color: Colors.white12),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  if (row.weeds > 0)
                    _buildActionChip(
                      context,
                      gardenBedProvider,
                      'weed',
                      rowIndex,
                      '🌿',
                      'Прополоть',
                      AppTheme.dangerColor,
                      weedCost: farmLevel * 10,
                    ),
                  if (row.weeds == 0)
                    Center(
                      child: _buildActionChip(
                        context,
                        gardenBedProvider,
                        'plant',
                        rowIndex,
                        '+',
                        'Посадить',
                        AppTheme.accentColor,
                      ),
                    ),
                ],
              ),
            ),
          ] else if (row.isWithered) ...[
            const Divider(height: 1, color: Colors.white12),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              child: Center(
                child: _buildActionChip(
                  context,
                  gardenBedProvider,
                  'weed',
                  rowIndex,
                  '🧹',
                  'Очистить',
                  AppTheme.dangerColor,
                  weedCost: farmLevel * 10,
                ),
              ),
            ),
          ] else ...[
            const Divider(height: 1, color: Colors.white12),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  if (row.weeds > 0)
                    _buildActionChip(
                      context,
                      gardenBedProvider,
                      'weed',
                      rowIndex,
                      '🌿',
                      'Прополоть',
                      AppTheme.dangerColor,
                      weedCost: gardenBedProvider.getWeedCost(row.plantType ?? 'wheat') * farmLevel * 10,
                    ),
                  _buildActionChip(
                    context,
                    gardenBedProvider,
                    'water',
                    rowIndex,
                    '💧',
                    'Полить',
                    AppTheme.accentColor,
                    waterCost: gardenBedProvider.getWaterCost(row.plantType ?? 'wheat') * farmLevel * 10,
                  ),
                  if (row.isMature)
                    _buildActionChip(
                      context,
                      gardenBedProvider,
                      'harvest',
                      rowIndex,
                      '🌾',
                      'Собрать',
                      AppTheme.successColor,
                      moneyReward: gardenBedProvider.getMoneyReward(row.plantType ?? 'wheat'),
                      foodReward: gardenBedProvider.getFoodReward(row.plantType ?? 'wheat'),
                    ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildPlantInfo(BuildContext context, GardenBedProvider gardenBedProvider, GardenBedRow row) {
    if (row.isEmpty && row.weeds == 0) {
      return const Text(
        'Пусто',
        style: const TextStyle(color: Colors.white38, fontSize: 13),
      );
    }

    if (row.isEmpty && row.weeds > 0) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Text('🌱', style: TextStyle(fontSize: 16)),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  'Заросло',
                  style: const TextStyle(color: Colors.orange, fontSize: 13, fontWeight: FontWeight.w600),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          if (row.weeds > 0) ...[
            const SizedBox(height: 2),
            Row(
              children: [
                Text(
                  List.filled(row.weeds, '🌿').join(),
                  style: const TextStyle(fontSize: 11),
                ),
                const SizedBox(width: 4),
                Text(
                  '${row.weeds}/3',
                  style: TextStyle(
                    color: row.isAtMaxWeeds ? AppTheme.dangerColor : Colors.orange,
                    fontSize: 10,
                  ),
                ),
              ],
            ),
          ],
        ],
      );
    }

    if (row.isWithered) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Text('🥀', style: TextStyle(fontSize: 16)),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  'Увядшее',
                  style: const TextStyle(color: AppTheme.dangerColor, fontSize: 13, fontWeight: FontWeight.w600),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          const SizedBox(height: 2),
          Text(
            'Нужно очистить',
            style: TextStyle(color: Colors.white54, fontSize: 11),
          ),
        ],
      );
    }

    // Planted or mature
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(
              gardenBedProvider.getPlantIcon(row.plantType ?? ''),
              style: const TextStyle(fontSize: 18),
            ),
            const SizedBox(width: 6),
            Expanded(
              child: Text(
                gardenBedProvider.getPlantName(row.plantType ?? ''),
                style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.w600),
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
        if (row.isPlanted) ...[
          const SizedBox(height: 4),
          Row(
            children: [
              Text(
                gardenBedProvider.getStageName(row.stage ?? 0),
                style: const TextStyle(color: Colors.white70, fontSize: 11),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: TweenAnimationBuilder<double>(
                  tween: Tween(begin: 0.0, end: gardenBedProvider.getRowProgress(row)),
                  duration: const Duration(seconds: 1),
                  curve: Curves.easeInOut,
                  builder: (context, value, _) {
                    return ClipRRect(
                      borderRadius: BorderRadius.circular(3),
                      child: LinearProgressIndicator(
                        value: value,
                        minHeight: 4,
                        color: row.isMature ? AppTheme.successColor : AppTheme.accentColor,
                        backgroundColor: Colors.white.withValues(alpha: 0.1),
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
          if (row.ticksToMature > 0) ...[
            const SizedBox(height: 2),
            Row(
              children: [
                const Text('🕛', style: TextStyle(fontSize: 11)),
                const SizedBox(width: 2),
                Text(
                  gardenBedProvider.getTicksToMatureText(row.ticksToMature),
                  style: const TextStyle(color: AppTheme.accentColor, fontSize: 10),
                ),
              ],
            ),
          ],
        ],
        if (row.isMature) ...[
          const SizedBox(height: 4),
          Row(
            children: [
              const Text('💰', style: TextStyle(fontSize: 12)),
              const SizedBox(width: 2),
              Text(
                gardenBedProvider.getMoneyReward(row.plantType ?? '').toStringAsFixed(0),
                style: const TextStyle(color: Colors.amber, fontSize: 11, fontWeight: FontWeight.w600),
              ),
              const SizedBox(width: 8),
              const Text('🍍', style: TextStyle(fontSize: 12)),
              const SizedBox(width: 2),
              Text(
                gardenBedProvider.getFoodReward(row.plantType ?? '').toStringAsFixed(0),
                style: const TextStyle(color: AppTheme.successColor, fontSize: 11, fontWeight: FontWeight.w600),
              ),
            ],
          ),
          if (row.ticksToMature > 0) ...[
            const SizedBox(height: 2),
            Row(
              children: [
                const Text('⏱', style: TextStyle(fontSize: 11)),
                const SizedBox(width: 2),
                Text(
                  gardenBedProvider.getTicksToMatureText(row.ticksToMature),
                  style: const TextStyle(color: Colors.orange, fontSize: 10),
                ),
              ],
            ),
          ],
        ],
        if (row.weeds > 0) ...[
          const SizedBox(height: 2),
          Row(
            children: [
              Text(
                List.filled(row.weeds, '🌿').join(),
                style: const TextStyle(fontSize: 11),
              ),
              const SizedBox(width: 4),
              Text(
                '${row.weeds}/3',
                style: TextStyle(
                  color: row.isAtMaxWeeds ? AppTheme.dangerColor : Colors.orange,
                  fontSize: 10,
                ),
              ),
            ],
          ),
        ],
        if (row.waterTimer > 0) ...[
          const SizedBox(height: 2),
          Row(
            children: [
              const Text('💧', style: TextStyle(fontSize: 11)),
              const SizedBox(width: 2),
              Text(
                '${row.waterTimer}',
                style: const TextStyle(color: AppTheme.accentColor, fontSize: 10),
              ),
            ],
          ),
        ],
        
      ],
    );
  }

  Widget _buildActionChip(
    BuildContext context,
    GardenBedProvider gardenBedProvider,
    String action,
    int rowIndex,
    String icon,
    String label,
    Color color, {
    double weedCost = 0,
    double waterCost = 0,
    double moneyReward = 0,
    double foodReward = 0,
  }) {
    final canAct = gardenBedProvider.canAct;
    String fullLabel = label;
    if (weedCost > 0) fullLabel = '$label (🍍${weedCost.toInt()})';
    if (waterCost > 0) fullLabel = '$label (🍍${waterCost.toInt()})';
    if (moneyReward > 0 && foodReward > 0) fullLabel = '$label (💰${moneyReward.toInt()} 🍍${foodReward.toInt()})';

    return ElevatedButton(
      onPressed: canAct
          ? () => _handleAction(context, gardenBedProvider, action, rowIndex)
          : null,
      style: ElevatedButton.styleFrom(
        backgroundColor: color.withValues(alpha: canAct ? 0.2 : 0.1),
        foregroundColor: canAct ? color : color.withValues(alpha: 0.5),
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
        minimumSize: const Size(0, 32),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: BorderSide(color: color.withValues(alpha: canAct ? 0.4 : 0.2)),
        ),
        elevation: 0,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(icon, style: const TextStyle(fontSize: 14)),
          const SizedBox(width: 4),
          Text(fullLabel, style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }

  Future<void> _handleAction(BuildContext context, GardenBedProvider gardenBedProvider, String action, int rowIndex) async {
    gardenBedProvider.clearError();
    if (action == 'plant') {
      final selectedPlant = await _showPlantSelectionDialog(context);
      if (selectedPlant != null) {
        await gardenBedProvider.gardenBedAction(widget.planetId, 'plant', rowIndex, plantType: selectedPlant);
      }
    } else if (action == 'weed' && gardenBedProvider.gardenBedState?.rows[rowIndex].isWithered == true) {
      await gardenBedProvider.gardenBedAction(widget.planetId, 'clear', rowIndex);
    } else {
      await gardenBedProvider.gardenBedAction(widget.planetId, action, rowIndex);
    }
  }

  Future<String?> _showPlantSelectionDialog(BuildContext context) async {
    final gardenBedProvider = context.read<GameProvider>().gardenBedProvider;
    final gameProvider = context.read<GameProvider>();
    final farmLevel = gameProvider.getBuildingLevelForPlanet(widget.planetId, "farm");
    final money = (gameProvider.selectedPlanet?.resources['money'] as num?)?.toDouble() ?? 0;

    final availablePlants = gardenBedProvider.getAvailablePlants(farmLevel);

    final result = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: AppTheme.cardColor,
        title: const Text('Выберите растение', style: TextStyle(color: Colors.white)),
        content: Container(
          width: double.maxFinite,
          constraints: const BoxConstraints(maxHeight: 400),
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: availablePlants.map((plant) {
                final type = plant['type'] as String;
                final name = plant['name'] as String;
                final icon = plant['icon'] as String;
                final seedCost = (plant['seedCost'] as num).toDouble();
                final moneyReward = (plant['moneyReward'] as num).toDouble();
                final foodReward = (plant['foodReward'] as num).toDouble();
                final unlockLevel = plant['unlockLevel'] as int;
                final isUnlocked = gardenBedProvider.isPlantUnlocked(type, farmLevel);
                final canAfford = money >= seedCost;

                return InkWell(
                  onTap: isUnlocked && canAfford
                      ? () => Navigator.pop(context, type)
                      : null,
                  borderRadius: BorderRadius.circular(8),
                  child: Container(
                    margin: const EdgeInsets.only(bottom: 8),
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    decoration: BoxDecoration(
                      color: isUnlocked && canAfford
                          ? Colors.white.withValues(alpha: 0.1)
                          : Colors.white.withValues(alpha: 0.03),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: isUnlocked && canAfford
                            ? Colors.white.withValues(alpha: 0.2)
                            : Colors.white.withValues(alpha: 0.05),
                      ),
                    ),
                    child: Row(
                      children: [
                        Text(icon, style: const TextStyle(fontSize: 22)),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Text(
                                    name,
                                    style: TextStyle(
                                      color: isUnlocked && canAfford
                                          ? Colors.white
                                          : Colors.white38,
                                      fontWeight: FontWeight.w600,
                                      fontSize: 14,
                                    ),
                                  ),
                                  if (!isUnlocked) ...[
                                    const SizedBox(width: 6),
                                    const Icon(Icons.lock, size: 14, color: Colors.white38),
                                  ],
                                ],
                              ),
                              const SizedBox(height: 2),
                              Row(
                                children: [
                                  if (isUnlocked) ...[
                                    const Text('💰', style: TextStyle(fontSize: 11)),
                                    const SizedBox(width: 2),
                                    Text(
                                      '-$seedCost',
                                      style: const TextStyle(color: Colors.red, fontSize: 11),
                                    ),
                                    const SizedBox(width: 12),
                                    const Text('💰', style: TextStyle(fontSize: 11)),
                                    const SizedBox(width: 2),
                                    Text(
                                      '+${moneyReward.toInt()}',
                                      style: const TextStyle(color: Colors.amber, fontSize: 11),
                                    ),
                                    const SizedBox(width: 8),
                                    const Text('🍍', style: TextStyle(fontSize: 11)),
                                    const SizedBox(width: 2),
                                    Text(
                                      '+${foodReward.toInt()}',
                                      style: const TextStyle(color: AppTheme.successColor, fontSize: 11),
                                    ),
                                  ] else ...[
                                    Text(
                                      'Требуется ур. $unlockLevel',
                                      style: const TextStyle(color: Colors.white38, fontSize: 11),
                                    ),
                                  ],
                                ],
                              ),
                            ],
                          ),
                        ),
                        if (isUnlocked && canAfford)
                          const Icon(Icons.chevron_right, color: Colors.white38),
                        if (isUnlocked && !canAfford)
                          const Icon(Icons.close, color: Colors.red, size: 18),
                        if (!isUnlocked)
                          const SizedBox(width: 18),
                      ],
                    ),
                  ),
                );
              }).toList(),
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена', style: TextStyle(color: Colors.white54)),
          ),
        ],
      ),
    );
    return result;
  }
}

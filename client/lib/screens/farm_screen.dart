import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/farm_provider.dart';
import '../providers/game_provider.dart';
import '../models/farm.dart';
import '../core/app_theme.dart';

class FarmScreen extends StatefulWidget {
  final String planetId;

  const FarmScreen({super.key, required this.planetId});

  @override
  State<FarmScreen> createState() => _FarmScreenState();
}

class _FarmScreenState extends State<FarmScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final fp = context.read<GameProvider>().farmProvider;
      fp.clearError();
      fp.getFarm(widget.planetId);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        title: const Text('Ферма'),
        backgroundColor: AppTheme.cardColor,
        foregroundColor: Colors.white,
        elevation: 0,
      ),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          final farmProvider = gameProvider.farmProvider;
          return ListenableBuilder(
            listenable: farmProvider,
            builder: (context, _) {
              if (farmProvider.isLoading && farmProvider.farmState == null) {
                return const Center(child: CircularProgressIndicator());
              }

              final farmState = farmProvider.farmState;
              if (farmState == null) {
                return const Center(child: Text('Ферма не построена'));
              }
              if (farmState.rows.isEmpty) {
                return const Center(child: Text('Ферма не построена'));
              }

              final farmLevel = gameProvider.getBuildingLevelForPlanet(widget.planetId, 'farm');

              return Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Error banner
                    if (farmProvider.errorMessage != null)
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
                                farmProvider.errorMessage!,
                                style: const TextStyle(color: AppTheme.dangerColor, fontSize: 13),
                              ),
                            ),
                            IconButton(
                              icon: const Icon(Icons.close, size: 18, color: AppTheme.dangerColor),
                              onPressed: () => farmProvider.clearError(),
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
                            'Ур. $farmLevel • ${farmState.rowCount} рядов',
                            style: const TextStyle(color: AppTheme.accentColor, fontWeight: FontWeight.w600),
                          ),
                          const Spacer(),
                          if (!farmProvider.canAct)
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                              decoration: BoxDecoration(
                                color: Colors.orange.withValues(alpha: 0.2),
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                '${farmProvider.remainingCooldown}с',
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
                        itemCount: farmState.rows.length,
                        itemBuilder: (context, rowIndex) {
                          return _buildRowCard(context, farmProvider, gameProvider, farmState.rows[rowIndex], rowIndex, farmLevel);
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

  Widget _buildRowCard(BuildContext context, FarmProvider farmProvider, GameProvider gameProvider, FarmRow row, int rowIndex, int farmLevel) {
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
                  child: _buildPlantInfo(context, farmProvider, row),
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
                      farmProvider,
                      'weed',
                      rowIndex,
                      '🌿',
                      'Прополоть',
                      AppTheme.dangerColor,
                      weedCost: 1,
                    ),
                  if (row.weeds == 0)
                    Center(
                      child: _buildActionChip(
                        context,
                        farmProvider,
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
                  farmProvider,
                  'weed',
                  rowIndex,
                  '🧹',
                  'Очистить',
                  AppTheme.dangerColor,
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
                      farmProvider,
                      'weed',
                      rowIndex,
                      '🌿',
                      'Прополоть',
                      AppTheme.dangerColor,
                      weedCost: farmProvider.getWeedCost(row.plantType ?? 'wheat'),
                    ),
                  _buildActionChip(
                    context,
                    farmProvider,
                    'water',
                    rowIndex,
                    '💧',
                    'Полить',
                    AppTheme.accentColor,
                    waterCost: farmProvider.getWaterCost(row.plantType ?? 'wheat'),
                  ),
                  if (row.isMature)
                    _buildActionChip(
                      context,
                      farmProvider,
                      'harvest',
                      rowIndex,
                      '🌾',
                      'Собрать',
                      AppTheme.successColor,
                      moneyReward: farmProvider.getMoneyReward(row.plantType ?? 'wheat'),
                      foodReward: farmProvider.getFoodReward(row.plantType ?? 'wheat'),
                    ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildPlantInfo(BuildContext context, FarmProvider farmProvider, FarmRow row) {
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
              farmProvider.getPlantIcon(row.plantType ?? ''),
              style: const TextStyle(fontSize: 18),
            ),
            const SizedBox(width: 6),
            Expanded(
              child: Text(
                farmProvider.getPlantName(row.plantType ?? ''),
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
                farmProvider.getStageName(row.stage ?? 0),
                style: const TextStyle(color: Colors.white70, fontSize: 11),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(3),
                  child: LinearProgressIndicator(
                    value: farmProvider.getRowProgress(row),
                    minHeight: 4,
                    color: row.isMature ? AppTheme.successColor : AppTheme.accentColor,
                    backgroundColor: Colors.white.withValues(alpha: 0.1),
                  ),
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
                  farmProvider.getTicksToMatureText(row.ticksToMature),
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
                farmProvider.getMoneyReward(row.plantType ?? '').toStringAsFixed(0),
                style: const TextStyle(color: Colors.amber, fontSize: 11, fontWeight: FontWeight.w600),
              ),
              const SizedBox(width: 8),
              const Text('🍍', style: TextStyle(fontSize: 12)),
              const SizedBox(width: 2),
              Text(
                farmProvider.getFoodReward(row.plantType ?? '').toStringAsFixed(0),
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
                  farmProvider.getTicksToMatureText(row.ticksToMature),
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
                '$row.waterTimer',
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
    FarmProvider farmProvider,
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
    final canAct = farmProvider.canAct;
    String fullLabel = label;
    if (weedCost > 0) fullLabel = '$label (🍍${weedCost.toInt()})';
    if (waterCost > 0) fullLabel = '$label (🍍${waterCost.toInt()})';
    if (moneyReward > 0 && foodReward > 0) fullLabel = '$label (💰${moneyReward.toInt()} 🍍${foodReward.toInt()})';

    return ElevatedButton(
      onPressed: canAct
          ? () => _handleAction(context, farmProvider, action, rowIndex)
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

  Future<void> _handleAction(BuildContext context, FarmProvider farmProvider, String action, int rowIndex) async {
    if (action == 'plant') {
      final selectedPlant = await _showPlantSelectionDialog(context);
      if (selectedPlant != null) {
        await farmProvider.farmAction(widget.planetId, 'plant', rowIndex, plantType: selectedPlant);
      }
    } else if (action == 'weed' && context.read<FarmProvider>().farmState?.rows[rowIndex].isWithered == true) {
      await farmProvider.farmAction(widget.planetId, 'weed', rowIndex);
    } else {
      await farmProvider.farmAction(widget.planetId, action, rowIndex);
    }
    farmProvider.clearError();
  }

  Future<String?> _showPlantSelectionDialog(BuildContext context) async {
    final farmProvider = context.read<GameProvider>().farmProvider;
    final gameProvider = context.read<GameProvider>();
    final farmLevel = gameProvider.getBuildingLevelForPlanet(widget.planetId, 'farm');
    final money = (gameProvider.selectedPlanet?.resources['money'] as num?)?.toDouble() ?? 0;

    final availablePlants = farmProvider.getAvailablePlants(farmLevel);

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
                final isUnlocked = farmProvider.isPlantUnlocked(type, farmLevel);
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

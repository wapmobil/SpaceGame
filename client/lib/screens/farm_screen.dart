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
      context.read<GameProvider>().farmProvider.getFarm(widget.planetId);
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
          if (farmProvider.isLoading && farmProvider.farmState == null) {
            return const Center(child: CircularProgressIndicator());
          }

          final farmState = farmProvider.farmState;
          if (farmState == null) {
            return const Center(child: Text('Ферма не построена'));
          }

          return Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
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
                        '${farmState.rowCount} рядов',
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
                const SizedBox(height: 16),

                // Farm rows
                Expanded(
                  child: ListView.builder(
                    itemCount: farmState.rows.length,
                    itemBuilder: (context, rowIndex) {
                      return _buildRowCard(context, farmProvider, farmState.rows[rowIndex], rowIndex);
                    },
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildRowCard(BuildContext context, FarmProvider farmProvider, FarmRow row, int rowIndex) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: AppTheme.cardColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: row.isMature
              ? AppTheme.successColor.withValues(alpha: 0.4)
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
          if (!row.isEmpty) ...[
            const Divider(height: 1, color: Colors.white12),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  // Weed button
                  if (row.weeds > 0)
                    _buildActionChip(
                      context,
                      farmProvider,
                      'weed',
                      rowIndex,
                      '🌿',
                      'Прополоть',
                      AppTheme.dangerColor,
                    ),
                  // Water button
                  _buildActionChip(
                    context,
                    farmProvider,
                    'water',
                    rowIndex,
                    '💧',
                    'Полить',
                    AppTheme.accentColor,
                  ),
                  // Harvest button
                  if (row.isMature)
                    _buildActionChip(
                      context,
                      farmProvider,
                      'harvest',
                      rowIndex,
                      '🌾',
                      'Собрать',
                      AppTheme.successColor,
                    ),
                ],
              ),
            ),
          ] else ...[
            const Divider(height: 1, color: Colors.white12),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              child: Center(
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
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildPlantInfo(BuildContext context, FarmProvider farmProvider, FarmRow row) {
    if (row.isEmpty) {
      return const Text(
        'Пусто',
        style: const TextStyle(color: Colors.white38, fontSize: 13),
      );
    }

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
          // Stage progress
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
        ],
        // Weeds indicator
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
        // Water indicator
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
    Color color,
  ) {
    final canAct = farmProvider.canAct;
    return ElevatedButton(
      onPressed: canAct
          ? () => _handleAction(context, farmProvider, action, rowIndex)
          : null,
      style: ElevatedButton.styleFrom(
        backgroundColor: color.withValues(alpha: canAct ? 0.2 : 0.1),
        foregroundColor: canAct ? color : color.withValues(alpha: 0.5),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
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
          Text(label, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }

  Future<void> _handleAction(BuildContext context, FarmProvider farmProvider, String action, int rowIndex) async {
    if (action == 'plant') {
      // Show plant selection dialog
      final selectedPlant = await _showPlantSelectionDialog(context);
      if (selectedPlant != null) {
        await farmProvider.farmAction(widget.planetId, 'plant', rowIndex, plantType: selectedPlant);
      }
    } else {
      await farmProvider.farmAction(widget.planetId, action, rowIndex);
    }
  }

  Future<String?> _showPlantSelectionDialog(BuildContext context) async {
    final farmProvider = context.read<GameProvider>().farmProvider;
    final result = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: AppTheme.cardColor,
        title: const Text('Выберите растение', style: TextStyle(color: Colors.white)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            _buildPlantOption(context, farmProvider, 'wheat', 'Пшеница', '🌾', '+5 еды'),
            const SizedBox(height: 8),
            _buildPlantOption(context, farmProvider, 'berries', 'Ягоды', '🫐', '+15 еды'),
            const SizedBox(height: 8),
            _buildPlantOption(context, farmProvider, 'melon', 'Космическая дыня', '🍈', '+30 еды'),
          ],
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

  Widget _buildPlantOption(
    BuildContext context,
    FarmProvider farmProvider,
    String type,
    String name,
    String icon,
    String reward,
  ) {
    return InkWell(
      onTap: () => Navigator.pop(context, type),
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          color: Colors.white.withValues(alpha: 0.05),
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
        ),
        child: Row(
          children: [
            Text(icon, style: const TextStyle(fontSize: 20)),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    name,
                    style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600, fontSize: 14),
                  ),
                  Text(
                    reward,
                    style: const TextStyle(color: AppTheme.successColor, fontSize: 11),
                  ),
                ],
              ),
            ),
            const Icon(Icons.chevron_right, color: Colors.white38),
          ],
        ),
      ),
    );
  }
}

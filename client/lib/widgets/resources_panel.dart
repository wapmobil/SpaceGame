import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/planet.dart';
import '../providers/game_provider.dart';

enum ResourcesPanelMode { compact, expanded }

class ResourcesPanel extends StatelessWidget {
  final Planet planet;
  final GameProvider gameProvider;
  final ResourcesPanelMode mode;
  final VoidCallback? onTap;

  const ResourcesPanel({
    super.key,
    required this.planet,
    required this.gameProvider,
    this.mode = ResourcesPanelMode.compact,
    this.onTap,
  });

  String _formatEnergyProd(double val) {
    if (val > 0) return '(+${val.toStringAsFixed(1)}/s)';
    if (val < 0) return '(${val.toStringAsFixed(1)}/s)';
    return '';
  }

  String _getProduction(String key) {
    switch (key) {
      case 'food':
        return gameProvider.productionFood.toStringAsFixed(1);
      case 'iron':
        return gameProvider.productionIron.toStringAsFixed(1);
      case 'composite':
        return gameProvider.productionComposite.toStringAsFixed(1);
      case 'mechanisms':
        return gameProvider.productionMechanisms.toStringAsFixed(1);
      case 'reagents':
        return gameProvider.productionReagents.toStringAsFixed(1);
      case 'money':
        return gameProvider.productionMoney.toStringAsFixed(1);
      default:
        return '0';
    }
  }

  Widget _buildCompactView() {
    final resources = planet.resources;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.zero,
        child: Container(
          width: double.infinity,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          decoration: BoxDecoration(
             color: AppTheme.cardColor,
             border: const Border(bottom: BorderSide(color: Color(0xff2196f3))),
          ),
       child: SingleChildScrollView(
             scrollDirection: Axis.horizontal,
             child: Row(
               children: Constants.resourceNames.keys.map((key) {
                final value = resources[key] ?? 0;
                final colorVal = (Constants.resourceColors[key] as Color).toARGB32();
                final icon = Constants.resourceIcons[key] ?? '❓';

                return Padding(
                  padding: const EdgeInsets.only(right: 16),
                  child: Row(
                    children: [
                      Text(icon, style: const TextStyle(fontSize: 14)),
                      const SizedBox(width: 4),
                      Text(
                        value.toInt().toString(),
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: Color(colorVal),
                        ),
                      ),
                    ],
                  ),
                );
              }).toList(),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildExpandedView() {
    final resources = planet.resources;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        InkWell(
          onTap: onTap,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            child: Row(
              children: const [
                Icon(Icons.keyboard_arrow_down, color: Colors.white70, size: 20),
                SizedBox(width: 8),
                Text(
                  'Ресурсы',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70),
                ),
              ],
            ),
          ),
        ),
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: Constants.resourceNames.keys.where((key) => key != 'energy').map((key) {
                    final value = resources[key] ?? 0;
                    final colorVal = (Constants.resourceColors[key] as Color).toARGB32();
                    final icon = Constants.resourceIcons[key] ?? '❓';
                    final production = _getProduction(key);
                    final prodNum = double.tryParse(production) ?? 0;
                    String rateText = '';
                    if (prodNum > 0) {
                      rateText = ' (+$production/s)';
                    } else if (prodNum < 0) {
                      rateText = ' ($production/s)';
                    }
                    final cap = gameProvider.storageCapacity;
                    String capStr = '';
                    if (key != 'energy' && key != 'money' && cap > 0) {
                      capStr = ' / ${cap.toStringAsFixed(0)}';
                    }
                    return Chip(
                      avatar: Text(icon, style: const TextStyle(fontSize: 16)),
                      label: Text(
                        '${Constants.resourceNames[key]}: ${value.toStringAsFixed(0)}$capStr$rateText',
                        style: const TextStyle(fontSize: 12),
                      ),
                      backgroundColor: Color(colorVal).withValues(alpha: 0.2),
                      side: BorderSide.none,
                    );
                  }).toList(),
                ),
                if ((resources['max_energy'] ?? 0) > 0) ...[
                  const SizedBox(height: 8),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          const Text('⚡ Энергия:', style: TextStyle(fontSize: 10, color: Colors.white70)),
                          const SizedBox(width: 4),
                          Text(
                            '${((resources['energy'] ?? 0).toInt())}/${(resources['max_energy'] ?? 0).toInt()}',
                            style: TextStyle(
                              fontSize: 10,
                              color: (resources['energy'] ?? 0) <= 0 ? Colors.red : Colors.white,
                            ),
                          ),
                          const SizedBox(width: 8),
                          Text(
                            _formatEnergyProd(gameProvider.productionEnergy),
                            style: const TextStyle(fontSize: 10, color: Colors.white70),
                          ),
                        ],
                      ),
                      const SizedBox(height: 2),
                      LinearProgressIndicator(
                        value: (resources['max_energy'] ?? 0) > 0
                            ? (resources['energy'] ?? 0) / (resources['max_energy'] ?? 1)
                            : 0,
                        minHeight: 6,
                        borderRadius: BorderRadius.circular(3),
                        valueColor: AlwaysStoppedAnimation(
                          (resources['energy'] ?? 0) <= 0 ? Colors.red : Colors.yellow,
                        ),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          ),
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    if (mode == ResourcesPanelMode.compact) {
      return _buildCompactView();
    }
    return _buildExpandedView();
  }
}

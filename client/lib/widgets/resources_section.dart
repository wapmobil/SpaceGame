import 'package:flutter/material.dart';
import '../utils/constants.dart';
import '../models/planet.dart';
import '../providers/game_provider.dart';

class ResourcesSection extends StatelessWidget {
  final Planet planet;
  final GameProvider gameProvider;

  const ResourcesSection({super.key, required this.planet, required this.gameProvider});

  String _formatEnergyProd(double val) {
    if (val > 0) return '(+${val.toStringAsFixed(1)}/s)';
    if (val < 0) return '(${val.toStringAsFixed(1)}/s)';
    return '';
  }

  String _getProduction(String key) {
    switch (key) {
      case 'food':
        return gameProvider.productionFood.toStringAsFixed(1);
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

  @override
  Widget build(BuildContext context) {
    final resources = planet.resources;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Resources',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70),
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: Constants.resourceNames.keys.where((key) => key != 'energy').map((key) {
                final value = resources[key] ?? 0;
                final colorVal = Constants.resourceColors[key] ?? Colors.white.value;
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
                if (key != 'energy' && cap > 0) {
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
            if (gameProvider.energyBufferMax > 0) ...[
              const SizedBox(height: 8),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Text('⚡ Energy:', style: TextStyle(fontSize: 10, color: Colors.white70)),
                      const SizedBox(width: 4),
                      Text(
                        '${gameProvider.energyBufferValue.toInt()}/${gameProvider.energyBufferMax.toInt()}',
                        style: TextStyle(
                          fontSize: 10,
                          color: gameProvider.energyBufferDeficit ? Colors.red : Colors.white,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        '${_formatEnergyProd(gameProvider.productionEnergy)}',
                        style: const TextStyle(fontSize: 10, color: Colors.white70),
                      ),
                      if (gameProvider.energyBufferDeficit) ...[
                        const SizedBox(width: 4),
                        const Text('(DEFICIT)', style: TextStyle(fontSize: 8, color: Colors.red)),
                      ],
                    ],
                  ),
                  const SizedBox(height: 2),
                  LinearProgressIndicator(
                    value: gameProvider.energyBufferMax > 0 
                        ? gameProvider.energyBufferValue / gameProvider.energyBufferMax 
                        : 0,
                    minHeight: 6,
                    borderRadius: BorderRadius.circular(3),
                    valueColor: AlwaysStoppedAnimation(
                      gameProvider.energyBufferDeficit ? Colors.red : Colors.yellow,
                    ),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}

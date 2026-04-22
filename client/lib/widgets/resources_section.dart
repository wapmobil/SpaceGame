import 'package:flutter/material.dart';
import '../utils/constants.dart';
import '../models/planet.dart';
import '../providers/game_provider.dart';

class ResourcesSection extends StatelessWidget {
  final Planet planet;
  final GameProvider gameProvider;

  const ResourcesSection({super.key, required this.planet, required this.gameProvider});

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
              children: Constants.resourceNames.keys.map((key) {
                final value = resources[key] ?? 0;
                final colorVal = Constants.resourceColors[key] ?? Colors.white.value;
                final icon = Constants.resourceIcons[key] ?? '❓';
                return Chip(
                  avatar: Text(icon, style: const TextStyle(fontSize: 16)),
                  label: Text(
                    '${Constants.resourceNames[key]}: ${value.toStringAsFixed(0)}',
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

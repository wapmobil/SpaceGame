import 'package:flutter/material.dart';
import '../providers/game_provider.dart';

class QuickStatsSection extends StatelessWidget {
  final GameProvider gameProvider;

  const QuickStatsSection({super.key, required this.gameProvider});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Quick Stats',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70),
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 16,
              runSpacing: 8,
              children: [
                _StatItem('Level', gameProvider.selectedPlanet?.level.toString() ?? '1'),
                _StatItem('Ships', gameProvider.ships.length.toString()),
                _StatItem('Research', gameProvider.researchState?.research.length.toString() ?? '0'),
                _StatItem('Expeditions', gameProvider.expeditions?.activeCount.toString() ?? '0'),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _StatItem extends StatelessWidget {
  final String label;
  final String value;

  const _StatItem(this.label, this.value);

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(label, style: const TextStyle(fontSize: 11, color: Colors.white54)),
        Text(value, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white)),
      ],
    );
  }
}

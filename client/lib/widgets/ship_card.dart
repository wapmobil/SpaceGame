import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/ship.dart';

class ShipCard extends StatelessWidget {
  final Ship ship;

  const ShipCard({super.key, required this.ship});

  @override
  Widget build(BuildContext context) {
    final icon = Constants.shipIcons[ship.type] ?? '🚀';
    final hpPercent = ship.maxHp > 0 ? (ship.hp / ship.maxHp).toDouble() : 0.0;

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Card(
        child: ListTile(
          leading: Text(icon, style: const TextStyle(fontSize: 28)),
          title: Text(
            ship.type.split('_').map((w) => w[0].toUpperCase() + w.substring(1)).join(' '),
            style: const TextStyle(fontSize: 14),
          ),
          subtitle: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 4),
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text('HP', style: TextStyle(fontSize: 10, color: Colors.white54)),
                            Text('${ship.hp}/${ship.maxHp}', style: const TextStyle(fontSize: 10, color: Colors.white70)),
                          ],
                        ),
                        const SizedBox(height: 2),
                        LinearProgressIndicator(
                          value: hpPercent,
                          minHeight: 4,
                          borderRadius: BorderRadius.circular(2),
                          color: hpPercent > 0.5
                              ? AppTheme.successColor
                              : hpPercent > 0.25
                                  ? AppTheme.warningColor
                                  : AppTheme.dangerColor,
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 16),
                  _StatIcon(Icons.shield, '${ship.armor}'),
                  _StatIcon(Icons.speed, '${ship.energy.toInt()}'),
                  _StatIcon(Icons.inventory, '${ship.cargo.toInt()}'),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatIcon extends StatelessWidget {
  final IconData icon;
  final String value;

  const _StatIcon(this.icon, this.value);

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Icon(icon, size: 14, color: Colors.white54),
        Text(value, style: const TextStyle(fontSize: 10, color: Colors.white)),
      ],
    );
  }
}

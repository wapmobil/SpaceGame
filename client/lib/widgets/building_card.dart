import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/building.dart';

class BuildingCard extends StatelessWidget {
  final Building building;

  const BuildingCard({super.key, required this.building});

  @override
  Widget build(BuildContext context) {
    final info = Constants.buildingTypes[building.type] ??
        {'name': building.type, 'icon': '🏗️', 'description': 'Unknown building'};

    final name = info['name'] ?? building.type;
    final icon = info['icon'] ?? '🏗️';
    final description = info['description'] ?? '';

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(icon, style: const TextStyle(fontSize: 28)),
            const SizedBox(height: 4),
            Text(
              name,
              style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w500, color: Colors.white),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            if (description.isNotEmpty)
              Text(
                description,
                style: const TextStyle(fontSize: 9, color: Colors.white54),
                textAlign: TextAlign.center,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            const SizedBox(height: 4),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: AppTheme.accentColor.withValues(alpha: 0.2),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                'Lv. ${building.level}',
                style: const TextStyle(fontSize: 10, color: AppTheme.accentColor),
              ),
            ),
            if (building.totalBuildTime > 0 && building.buildProgress > 0 && building.buildProgress < 1) ...[
              const SizedBox(height: 4),
              LinearProgressIndicator(
                value: building.buildProgress,
                minHeight: 3,
                borderRadius: BorderRadius.circular(2),
                color: AppTheme.accentColor,
              ),
              const SizedBox(height: 2),
              Text(
                '${((1 - building.buildProgress) * building.totalBuildTime).toInt()}s',
                style: const TextStyle(fontSize: 8, color: AppTheme.accentColor),
              ),
            ],
            if (building.totalBuildTime > 0 && building.buildProgress >= 1) ...[
              const SizedBox(height: 4),
              Text(
                'Complete',
                style: const TextStyle(fontSize: 8, color: Colors.green),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

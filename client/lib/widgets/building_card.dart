import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/building.dart';

class BuildingCard extends StatelessWidget {
  final Building building;
  final VoidCallback? onTap;

  const BuildingCard({super.key, required this.building, this.onTap});

  @override
  Widget build(BuildContext context) {
    final info = Constants.buildingTypes[building.type] ??
        {'name': building.type, 'icon': '🏗️', 'description': 'Unknown building'};

    final name = info['name'] ?? building.type;
    final icon = info['icon'] ?? '🏗️';
    final description = info['description'] ?? '';

    final isPending = building.pending == true && building.buildProgress >= 1;
    final cost = Constants.getBuildingCost(building.type, building.level + 1);
    final Widget? costWidget = cost['food'] > 0
        ? Column(
            children: [
              const SizedBox(height: 2),
              Text(
                '🍖 ${cost['food'].toInt()}  💰 ${cost['money'].toInt()}',
                style: const TextStyle(fontSize: 7, color: Colors.white54),
              ),
            ],
          )
        : null;

    Widget content = Column(
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
        if (costWidget != null) costWidget,
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
        if (building.totalBuildTime > 0 && building.buildProgress >= 1 && !isPending) ...[
          const SizedBox(height: 4),
          Text(
            'Complete',
            style: const TextStyle(fontSize: 8, color: Colors.green),
          ),
        ],
        if (isPending) ...[
          const SizedBox(height: 4),
          Text(
            'Tap to claim!',
            style: TextStyle(fontSize: 8, color: AppTheme.accentColor, fontWeight: FontWeight.bold),
          ),
        ],
      ],
    );

    if (isPending && onTap != null) {
      return Card(
        child: InkWell(
          onTap: onTap,
          child: Padding(
            padding: const EdgeInsets.all(8),
            child: content,
          ),
        ),
      );
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: content,
      ),
    );
  }
}

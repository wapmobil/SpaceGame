import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/planet.dart';

class ResourceBar extends StatelessWidget {
  final Planet? planet;

  const ResourceBar({super.key, this.planet});

  @override
  Widget build(BuildContext context) {
    if (planet == null) return const SizedBox.shrink();

    final resources = planet!.resources;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: AppTheme.cardColor,
        border: Border(bottom: BorderSide(color: AppTheme.primaryColor)),
      ),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: Constants.resourceNames.keys.map((key) {
            final value = resources[key] ?? 0;
            final colorVal = Constants.resourceColors[key] ?? Colors.white.value;
            final icon = Constants.resourceIcons[key] ?? '❓';

            return Padding(
              padding: const EdgeInsets.only(right: 16),
              child: Row(
                children: [
                  Text(icon, style: const TextStyle(fontSize: 14)),
                  const SizedBox(width: 4),
                  Text(
                    value.toStringAsFixed(value == value.toInt() ? 0 : 1),
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
    );
  }
}

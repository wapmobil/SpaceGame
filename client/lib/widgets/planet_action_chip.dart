import 'package:flutter/material.dart';
import '../core/app_theme.dart';

class PlanetActionChip extends StatelessWidget {
  final IconData? icon;
  final String? label;
  final VoidCallback onTap;

  const PlanetActionChip({
    super.key,
    this.icon,
    this.label,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: ActionChip(
        avatar: icon == null ? null : Icon(icon!, size: 16),
        label: label != null ? Text(label!, style: const TextStyle(fontSize: 11)) : const SizedBox.shrink(),
        onPressed: onTap,
        backgroundColor: AppTheme.cardColor,
        side: const BorderSide(color: AppTheme.primaryColor),
      ),
    );
  }
}

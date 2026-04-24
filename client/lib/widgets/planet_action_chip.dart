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
    return ElevatedButton(
      onPressed: onTap,
      style: ElevatedButton.styleFrom(
        backgroundColor: AppTheme.accentColor.withValues(alpha: 0.2),
        foregroundColor: Colors.white,
        side: BorderSide(color: AppTheme.accentColor.withValues(alpha: 0.4), width: 2),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        minimumSize: const Size(0, 32),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        elevation: 0,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (icon != null) Icon(icon!, size: 14, color: Colors.white),
          if (icon != null && label != null) const SizedBox(width: 4),
          if (label != null) Text(label!, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: Colors.white)),
        ],
      ),
    );
  }
}

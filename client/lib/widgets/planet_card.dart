import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../models/planet.dart';

class PlanetCard extends StatelessWidget {
  final Planet planet;
  final VoidCallback onTap;

  const PlanetCard({super.key, required this.planet, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Container(
                width: 60,
                height: 60,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [
                      AppTheme.primaryColor,
                      AppTheme.secondaryColor,
                    ],
                  ),
                ),
                child: Center(
                  child: Text(
                    planet.name[0].toUpperCase(),
                    style: const TextStyle(
                      fontSize: 28,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      planet.name,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                    const SizedBox(height: 4),
                  Text(
                       'Уровень ${planet.level}',
                       style: const TextStyle(fontSize: 12, color: Colors.white54),
                     ),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        _ResourceDot('🍍', (planet.resources['food'] ?? 0).toInt()),
                        _ResourceDot('⚡', (planet.resources['energy'] ?? 0).toInt()),
                        _ResourceDot('💰', (planet.resources['money'] ?? 0).toInt()),
                      ],
                    ),
                  ],
                ),
              ),
              Icon(Icons.chevron_right, color: Colors.white24),
            ],
          ),
        ),
      ),
    );
  }
}

class _ResourceDot extends StatelessWidget {
  final String icon;
  final int value;

  const _ResourceDot(this.icon, this.value);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: Row(
        children: [
          Text(icon, style: const TextStyle(fontSize: 12)),
          const SizedBox(width: 2),
          Text(
            value.toString(),
            style: const TextStyle(fontSize: 11, color: Colors.white70),
          ),
        ],
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../widgets/resource_bar.dart' as resource_bar;
import '../widgets/building_card.dart';
import 'shipyard_screen.dart' as ship;
import 'research_screen.dart' as research;
import 'battle_screen.dart' as battle;
import 'expedition_screen.dart' as expedition;
import 'market_screen.dart' as market;
import 'mining_screen.dart' as mining;

class PlanetScreen extends StatelessWidget {
  const PlanetScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<GameProvider>(
      builder: (context, gameProvider, _) {
        final planet = gameProvider.selectedPlanet;
        if (planet == null) {
          return const Center(child: Text('Select a planet'));
        }

        return NestedScrollView(
          headerSliverBuilder: (context, innerBoxIsScrolled) => [
            SliverToBoxAdapter(
              child: Column(
                children: [
                  resource_bar.ResourceBar(planet: planet),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(
                            planet.name,
                            style: const TextStyle(
                              fontSize: 24,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                            ),
                          ),
                        ),
                        _PlanetActionChip(
                          icon: Icons.rocket_launch,
                          label: 'Shipyard',
                          onTap: () => _navigateTo(context, const ship.ShipyardScreen()),
                        ),
                        _PlanetActionChip(
                          icon: Icons.science,
                          label: 'Research',
                          onTap: () => _navigateTo(context, const research.ResearchScreen()),
                        ),
                        _PlanetActionChip(
                          icon: Icons.local_fire_department,
                          label: 'Battle',
                          onTap: () => _navigateTo(context, const battle.BattleScreen()),
                        ),
                      ],
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    child: Row(
                      children: [
                        _PlanetActionChip(
                          icon: Icons.explore,
                          label: 'Expedition',
                          onTap: () => _navigateTo(context, const expedition.ExpeditionScreen()),
                        ),
                        _PlanetActionChip(
                          icon: Icons.store,
                          label: 'Market',
                          onTap: () => _navigateTo(context, const market.MarketScreen()),
                        ),
                        _PlanetActionChip(
                          icon: Icons.diamond_outlined,
                          label: 'Mining',
                          onTap: () => _navigateTo(context, const mining.MiningScreen()),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 8),
                ],
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 8)),
          ],
          body: _buildPlanetContent(context, planet, gameProvider),
        );
      },
    );
  }

  Widget _buildPlanetContent(BuildContext context, planet, GameProvider gameProvider) {
    return RefreshIndicator(
      onRefresh: () async {
        if (gameProvider.selectedPlanet != null) {
          final id = gameProvider.selectedPlanet!.id;
          await gameProvider.loadBuildings(id);
          await gameProvider.loadPlanetDetail(id);
        }
      },
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildResourcesSection(context, planet),
            const SizedBox(height: 16),
            _buildBuildingsSection(context, gameProvider),
            const SizedBox(height: 16),
            _buildQuickStats(context, gameProvider),
          ],
        ),
      ),
    );
  }

  Widget _buildResourcesSection(BuildContext context, planet) {
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
            if (resources['max_energy'] != null && resources['energy'] != null)
              Padding(
                padding: const EdgeInsets.only(top: 12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text('Energy', style: TextStyle(fontSize: 12, color: Colors.white70)),
                        Text(
                          '${resources['energy'].toStringAsFixed(0)} / ${resources['max_energy'].toStringAsFixed(0)}',
                          style: const TextStyle(fontSize: 12, color: Colors.white70),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    LinearProgressIndicator(
                      value: resources['max_energy'] != 0
                          ? (resources['energy'] / resources['max_energy']).clamp(0, 1)
                          : 0,
                      minHeight: 6,
                      borderRadius: BorderRadius.circular(3),
                      color: AppTheme.accentColor,
                    ),
                  ],
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildBuildingsSection(BuildContext context, GameProvider gameProvider) {
    final buildings = gameProvider.buildings;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Buildings',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70),
                ),
                TextButton(
                  onPressed: () => _showBuildDialog(context, gameProvider),
                  child: const Text('Build +'),
                ),
              ],
            ),
            if (buildings.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(
                  child: Text('No buildings yet. Build your first structure!', style: TextStyle(color: Colors.white38)),
                ),
              ),
            GridView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                childAspectRatio: 1.3,
                crossAxisSpacing: 8,
                mainAxisSpacing: 8,
              ),
              itemCount: buildings.length,
              itemBuilder: (context, index) {
                return BuildingCard(building: buildings[index]);
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildQuickStats(BuildContext context, GameProvider gameProvider) {
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

  void _showBuildDialog(BuildContext context, GameProvider gameProvider) {
    final allBuildings = Constants.buildingTypes.keys.toList();

    showModalBottomSheet(
      context: context,
      builder: (context) => Container(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Build Structure',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 16),
            ...allBuildings.map((key) {
              final info = Constants.buildingTypes[key]!;
              final existing = gameProvider.buildings.where((b) => b.type == key).toList();
              final currentLevel = existing.isNotEmpty ? existing.first.level : 0;
              final isBuilding = existing.isNotEmpty && existing.first.totalBuildTime > 0 && existing.first.buildProgress < 1;
              return ListTile(
                leading: Text(info['icon'] as String, style: const TextStyle(fontSize: 24)),
                title: Text(info['name'] as String),
                subtitle: Text(
                  isBuilding
                      ? 'Building... Lv.${currentLevel}'
                      : existing.isNotEmpty
                          ? 'Lv.${currentLevel} - Upgrade'
                          : info['description'] as String,
                ),
                enabled: !isBuilding,
                onTap: () {
                  Navigator.pop(context);
                  gameProvider.buildStructure(key);
                },
              );
            }),
          ],
        ),
      ),
    );
  }

  void _navigateTo(BuildContext context, Widget screen) {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => screen),
    );
  }
}

class _PlanetActionChip extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  const _PlanetActionChip({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: ActionChip(
        avatar: Icon(icon, size: 16),
        label: Text(label, style: const TextStyle(fontSize: 11)),
        onPressed: onTap,
        backgroundColor: AppTheme.cardColor,
        side: const BorderSide(color: AppTheme.primaryColor),
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

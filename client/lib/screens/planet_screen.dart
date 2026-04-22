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
                        if (gameProvider.canResearch)
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
                        if (gameProvider.canExpedition)
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
                        if (gameProvider.canMining)
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
          await gameProvider.loadBuildDetails(id);
          await gameProvider.loadPlanetDetail(id);
        }
      },
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildResourcesSection(context, planet, gameProvider),
            if (!gameProvider.baseOperational) ...[
              const SizedBox(height: 8),
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.red.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Row(
                  children: [
                    Icon(Icons.warning, color: Colors.red, size: 16),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        'Base not operational! Produce food to unlock research, expeditions, and mining.',
                        style: TextStyle(fontSize: 10, color: Colors.red),
                      ),
                    ),
                  ],
                ),
              ),
            ],
            const SizedBox(height: 16),
            _buildBuildingsSection(context, gameProvider),
            const SizedBox(height: 16),
            _buildQuickStats(context, gameProvider),
          ],
        ),
      ),
    );
  }

  Widget _buildResourcesSection(BuildContext context, planet, GameProvider gameProvider) {
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
                final building = buildings[index];
                final isPending = building.pending == true && building.buildProgress <= 0;
                return BuildingCard(
                  building: building,
                  onTap: isPending ? () => gameProvider.confirmBuilding(building.type) : null,
                );
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
    final allBuildingsList = gameProvider.buildings;
    final hasFarm = allBuildingsList.where((b) => b.type == 'farm').any((b) => !b.pending && b.level > 0);
    final hasSolar = allBuildingsList.where((b) => b.type == 'solar').any((b) => !b.pending && b.level > 0);

    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: AppTheme.cardColor,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        child: Container(
          constraints: BoxConstraints(maxHeight: MediaQuery.of(context).size.height * 0.7),
          padding: const EdgeInsets.all(16),
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Build Structure',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white),
                ),
                const SizedBox(height: 4),
                Text(
                  'Construction: ${gameProvider.activeConstructions}/${gameProvider.maxConstructions}',
                  style: TextStyle(fontSize: 12, color: Colors.white54),
                ),
                if (gameProvider.activeConstructions >= gameProvider.maxConstructions) ...[
                  const SizedBox(height: 4),
                  Text(
                    'Research Parallel Construction to build more simultaneously',
                    style: TextStyle(fontSize: 10, color: Colors.orange),
                  ),
                ],
                if (gameProvider.errorMessage != null && gameProvider.errorMessage!.isNotEmpty) ...[
                  const SizedBox(height: 4),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
                    decoration: BoxDecoration(
                      color: Colors.red.withValues(alpha: 0.15),
                      border: Border.all(color: Colors.red.withValues(alpha: 0.4)),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      gameProvider.errorMessage!,
                      style: const TextStyle(fontSize: 11, color: Colors.red),
                    ),
                  ),
                  const SizedBox(height: 4),
                ],
                const SizedBox(height: 8),
                ...allBuildings.where((key) {
                  if (key != 'farm' && !hasFarm) return false;
                  if (key != 'farm' && key != 'solar' && !hasSolar) return false;
                  return true;
                }).toList().map((key) {
                  final info = Constants.buildingTypes[key]!;
                  final existing = allBuildingsList.where((b) => b.type == key).toList();
                  final currentLevel = existing.isNotEmpty ? existing.first.level : 0;
                  final isBuilding = existing.isNotEmpty && existing.first.buildTime > 0 && existing.first.buildProgress > 0 && existing.first.buildProgress <= existing.first.buildTime;
                  final isPending = existing.isNotEmpty && existing.first.pending == true && existing.first.buildProgress <= 0;
                  double nextCostFood, nextCostMoney;
                  if (existing.isNotEmpty) {
                    nextCostFood = existing.first.nextCostFood;
                    nextCostMoney = existing.first.nextCostMoney;
                  } else {
                    final cost = gameProvider.buildingCosts[key];
                    nextCostFood = cost?['food'] ?? 0;
                    nextCostMoney = cost?['money'] ?? 0;
                  }
                  final canAfford = gameProvider.selectedPlanet != null &&
                       ((gameProvider.selectedPlanet!.resources['food'] ?? 0) as num).toDouble() >= nextCostFood &&
                       ((gameProvider.selectedPlanet!.resources['money'] ?? 0) as num).toDouble() >= nextCostMoney;
                  return ListTile(
                    leading: Text(info['icon'] as String, style: const TextStyle(fontSize: 24)),
                    title: Text(info['name'] as String, style: const TextStyle(color: Colors.white)),
                    subtitle: Text(
                      isBuilding
                          ? 'Building... Lv.$currentLevel'
                          : isPending
                              ? 'Pending confirmation - tap the building card to claim'
                              : existing.isNotEmpty
                                  ? 'Lv.$currentLevel → ${currentLevel + 1} | 🍖$nextCostFood 💰$nextCostMoney'
                                  : '${info['description'] as String} | 🍖$nextCostFood 💰$nextCostMoney',
                    ),
                    enabled: !isBuilding && !isPending && canAfford && gameProvider.activeConstructions < gameProvider.maxConstructions,
                    onTap: () {
                      Navigator.pop(context);
                      gameProvider.buildStructure(key);
                    },
                  );
                }),
              ],
            ),
          ),
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

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../widgets/resource_bar.dart' as resource_bar;
import '../widgets/planet_action_chip.dart';
import '../widgets/building_card.dart';
import '../widgets/build_dialog.dart';
import '../widgets/resources_section.dart';
import '../widgets/quick_stats_section.dart';
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
          return const Center(child: Text('Выберите планету'));
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
                        PlanetActionChip(
                          icon: Icons.rocket_launch,
                          label: 'Верфь',
                          onTap: () => _navigateTo(context, const ship.ShipyardScreen()),
                        ),
                        if (gameProvider.canResearch)
                          PlanetActionChip(
                            icon: Icons.science,
                            label: 'Исследования',
                            onTap: () => _navigateTo(context, const research.ResearchScreen()),
                          ),
                        PlanetActionChip(
                          icon: Icons.local_fire_department,
                          label: 'Битва',
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
                          PlanetActionChip(
                            icon: Icons.explore,
                            label: 'Экспедиция',
                            onTap: () => _navigateTo(context, const expedition.ExpeditionScreen()),
                          ),
                        PlanetActionChip(
                          icon: Icons.store,
                          label: 'Рынок',
                          onTap: () => _navigateTo(context, const market.MarketScreen()),
                        ),
                        if (gameProvider.canMining)
                          PlanetActionChip(
                            icon: Icons.diamond_outlined,
                            label: 'Горное дело',
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
            ResourcesSection(planet: planet, gameProvider: gameProvider),
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
                        'База не работает! Производите еду, чтобы открыть исследования, экспедиции и горное дело.',
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
            QuickStatsSection(gameProvider: gameProvider),
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
                  'Здания',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70),
                ),
                TextButton(
                  onPressed: () => _showBuildDialog(context, gameProvider),
                  child: const Text('Построить +'),
                ),
              ],
            ),
            if (buildings.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(
                  child: Text('Зданий пока нет. Постройте первое сооружение!', style: TextStyle(color: Colors.white38)),
                ),
              ),
            ListView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: buildings.length,
              itemBuilder: (context, index) {
                return Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: BuildingCard(
                    building: buildings[index],
                    gameProvider: gameProvider,
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showBuildDialog(BuildContext context, GameProvider gameProvider) {
    showDialog(
      context: context,
      builder: (context) => BuildDialog(gameProvider: gameProvider),
    );
  }

  void _navigateTo(BuildContext context, Widget screen) {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => screen),
    );
  }
}

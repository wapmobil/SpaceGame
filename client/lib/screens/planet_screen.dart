import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../widgets/resources_panel.dart';
import '../widgets/building_card.dart';
import '../widgets/build_dialog.dart';
import '../widgets/quick_stats_section.dart';
import 'shipyard_screen.dart' as ship;
import 'research_screen.dart' as research;
import 'expedition_screen.dart' as expedition;
import 'market_screen.dart' as market;
import 'mining_screen.dart' as mining;

class PlanetScreen extends StatefulWidget {
  const PlanetScreen({super.key});

  @override
  State<PlanetScreen> createState() => _PlanetScreenState();
}

class _PlanetScreenState extends State<PlanetScreen> {
  ResourcesPanelMode _mode = ResourcesPanelMode.expanded;

  void _toggleMode() {
    setState(() {
      _mode = _mode == ResourcesPanelMode.expanded
          ? ResourcesPanelMode.compact
          : ResourcesPanelMode.expanded;
    });
  }

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
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                child: Text(
                  planet.name,
                  style: const TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
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
            ResourcesPanel(
              mode: _mode,
              planet: planet,
              gameProvider: gameProvider,
              onTap: _toggleMode,
            ),
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
                    onNavigateBuilding: (buildingType, action) => _navigateTo(context, _getScreenForAction(buildingType, action)),
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _getScreenForAction(String buildingType, String action) {
    switch (action) {
      case 'research':
        return const research.ResearchScreen();
      case 'shipyard':
        return const ship.ShipyardScreen();
      case 'expedition':
        return const expedition.ExpeditionScreen();
      case 'mining':
        return const mining.MiningScreen();
      case 'market':
        return const market.MarketScreen();
      default:
        return const SizedBox.shrink();
    }
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

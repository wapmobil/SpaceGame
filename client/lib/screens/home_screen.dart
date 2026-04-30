import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';

import '../widgets/planet_card.dart';
import '../widgets/resources_panel.dart';
import 'planet_screen.dart';
import 'settings_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _selectedIndex = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<GameProvider>().loadPlanets();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
        index: _selectedIndex,
        children: [
          _buildPlanetsList(),
          const PlanetScreen(),
          const SettingsScreen(),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _selectedIndex,
        onDestinationSelected: (index) {
          setState(() => _selectedIndex = index);
          if (index == 0) {
            context.read<GameProvider>().loadPlanets();
          }
        },
         destinations: const [
          NavigationDestination(icon: Icon(Icons.public), label: 'Планеты'),
          NavigationDestination(icon: Icon(Icons.language), label: 'Планета'),
          NavigationDestination(icon: Icon(Icons.settings), label: 'Настройки'),
        ],
      ),
    );
  }

  Widget _buildPlanetsList() {
    return Consumer<GameProvider>(
      builder: (context, gameProvider, _) {
        if (!gameProvider.isLoggedIn) {
          return const Center(child: Text('Не авторизован'));
        }

        return Column(
          children: [
            if (gameProvider.planets.isNotEmpty && gameProvider.selectedPlanet != null)
              ResourcesPanel(mode: ResourcesPanelMode.compact, planet: gameProvider.selectedPlanet!, gameProvider: gameProvider),
            Expanded(
              child: gameProvider.planets.isEmpty
                  ? _buildEmptyState()
                  : _buildPlanetList(gameProvider),
            ),
          ],
        );
      },
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.public, size: 80, color: Colors.white24),
          const SizedBox(height: 16),
          const Text(
            'Планет пока нет',
            style: TextStyle(fontSize: 20, color: Colors.white54),
          ),
          const SizedBox(height: 8),
          const Text(
            'Создайте свою первую планету для начала',
            style: TextStyle(color: Colors.white38),
          ),
          const SizedBox(height: 24),
          ElevatedButton(
            onPressed: () => _showCreatePlanetDialog(),
            child: const Text('Создать планету'),
          ),
        ],
      ),
    );
  }

  Widget _buildPlanetList(GameProvider gameProvider) {
    return RefreshIndicator(
      onRefresh: () => gameProvider.loadPlanets(),
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: gameProvider.planets.length,
        itemBuilder: (context, index) {
          final planet = gameProvider.planets[index];
          return Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: PlanetCard(
              planet: planet,
              onTap: () {
                gameProvider.selectPlanet(planet);
                setState(() => _selectedIndex = 1);
              },
            ),
          );
        },
      ),
    );
  }

  void _showCreatePlanetDialog() {
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Создать планету'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            hintText: 'Название планеты',
            prefixIcon: Icon(Icons.public),
          ),
          autofocus: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена'),
          ),
          ElevatedButton(
            onPressed: () {
              final name = controller.text.trim();
              if (name.isNotEmpty) {
                context.read<GameProvider>().createPlanet(name);
                Navigator.pop(context);
              }
            },
            child: const Text('Создать'),
          ),
        ],
      ),
    );
  }
}

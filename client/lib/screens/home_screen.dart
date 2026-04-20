import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';

import '../widgets/planet_card.dart';
import '../widgets/resource_bar.dart';
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
        onDestinationSelected: (index) => setState(() => _selectedIndex = index),
        destinations: const [
          NavigationDestination(icon: Icon(Icons.public), label: 'Planets'),
          NavigationDestination(icon: Icon(Icons.language), label: 'Planet'),
          NavigationDestination(icon: Icon(Icons.settings), label: 'Settings'),
        ],
      ),
    );
  }

  Widget _buildPlanetsList() {
    return Consumer<GameProvider>(
      builder: (context, gameProvider, _) {
        if (!gameProvider.isLoggedIn) {
          return const Center(child: Text('Not logged in'));
        }

        return Column(
          children: [
            if (gameProvider.planets.isNotEmpty)
              ResourceBar(planet: gameProvider.selectedPlanet),
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
          Icon(Icons.public, size: 80, color: Colors.white24),
          const SizedBox(height: 16),
          const Text(
            'No planets yet',
            style: TextStyle(fontSize: 20, color: Colors.white54),
          ),
          const SizedBox(height: 8),
          const Text(
            'Create your first planet to begin',
            style: TextStyle(color: Colors.white38),
          ),
          const SizedBox(height: 24),
          ElevatedButton(
            onPressed: () => _showCreatePlanetDialog(),
            child: const Text('Create Planet'),
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
        title: const Text('Create Planet'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            hintText: 'Planet name',
            prefixIcon: Icon(Icons.public),
          ),
          autofocus: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              final name = controller.text.trim();
              if (name.isNotEmpty) {
                context.read<GameProvider>().createPlanet(name);
                Navigator.pop(context);
              }
            },
            child: const Text('Create'),
          ),
        ],
      ),
    );
  }
}

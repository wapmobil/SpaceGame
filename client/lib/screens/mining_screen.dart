import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';

class MiningScreen extends StatelessWidget {
  const MiningScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Mining Dungeon')),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          final planet = gameProvider.selectedPlanet;
          if (planet == null) return const Center(child: Text('No planet selected'));

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildMiningStatus(context, gameProvider),
                const SizedBox(height: 16),
                _buildMiningMaze(context, gameProvider),
                const SizedBox(height: 16),
                _buildMiningControls(context, gameProvider),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildMiningStatus(BuildContext context, GameProvider gameProvider) {
    final mining = gameProvider.miningState;

    if (mining == null) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            children: [
              Icon(Icons.forest_rounded, size: 48, color: Colors.white24),
              const SizedBox(height: 16),
              const Text('No active mining session', style: TextStyle(color: Colors.white54)),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => gameProvider.startMining(),
                child: const Text('Start Mining'),
              ),
            ],
          ),
        ),
      );
    }

    final statusColor = mining.gameEnded
        ? (mining.endReason == 'completed' ? AppTheme.successColor : AppTheme.dangerColor)
        : AppTheme.accentColor;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('Mining Session', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: statusColor.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    mining.gameEnded ? (mining.endReason == 'completed' ? 'Complete' : 'Dead') : 'Active',
                    style: TextStyle(
                      fontSize: 11,
                      color: statusColor,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _MiningStat('💰', '${mining.moneyCollected.toInt()}'),
                _MiningStat('❤️', '${mining.playerHp}/${mining.playerMaxHp}'),
                _MiningStat('💣', '${mining.playerBombs}'),
                _MiningStat('Level', '${mining.baseLevel}'),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMiningMaze(BuildContext context, GameProvider gameProvider) {
    final mining = gameProvider.miningState;
    if (mining == null || mining.maze.isEmpty) return const SizedBox.shrink();

    final rows = mining.maze.length;
    final cols = rows > 0 ? mining.maze[0].length : 0;
    if (cols == 0) return const SizedBox.shrink();

     return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Dungeon Map', style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 8),
            GridView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: cols,
                childAspectRatio: 1,
                crossAxisSpacing: 2,
                mainAxisSpacing: 2,
              ),
              itemCount: rows * cols,
              itemBuilder: (context, index) {
                final row = index ~/ cols;
                final col = index % cols;
                if (row >= mining.maze.length || col >= mining.maze[row].length) {
                  return const SizedBox.shrink();
                }

                final cell = mining.maze[row][col];
                final isPlayer = row == mining.playerY && col == mining.playerX;
                final isExit = row == mining.exitY && col == mining.exitX;
                final monster = mining.monsters.where((m) => m.alive && m.x == col && m.y == row).toList();

                return Container(
                  decoration: BoxDecoration(
                    color: _getCellColor(cell, isPlayer, isExit),
                    border: Border.all(color: Colors.white12, width: 0.5),
                  ),
                  child: Center(
                     child: isPlayer
                         ? const Text('🧑‍🚀', style: TextStyle(fontSize: 14))
                         : isExit
                             ? const Text('🚪', style: TextStyle(fontSize: 14))
                             : monster.isNotEmpty
                                 ? Text(
                                     monster.first.icon,
                                     style: const TextStyle(fontSize: 14),
                                   )
                                 : cell == 'wall'
                                     ? const SizedBox.shrink()
                                     : cell == 'money'
                                         ? const Text('💰', style: TextStyle(fontSize: 10))
                                         : const SizedBox.shrink(),
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  Color _getCellColor(String cell, bool isPlayer, bool isExit) {
    if (isPlayer) return AppTheme.accentColor.withValues(alpha: 0.4);
    if (isExit) return AppTheme.successColor.withValues(alpha: 0.4);
    if (cell == 'wall') return Colors.grey.shade800;
    return AppTheme.cardColor;
  }

  Widget _buildMiningControls(BuildContext context, GameProvider gameProvider) {
    final mining = gameProvider.miningState;
    if (mining == null || !mining.canMove) return const SizedBox.shrink();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Controls', style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 12),
            Center(
              child: Column(
                children: [
                  _DpadButton(
                    icon: Icons.arrow_upward,
                    label: 'Up',
                    onPressed: () => gameProvider.miningMove('up'),
                  ),
                  const SizedBox(height: 4),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      _DpadButton(
                        icon: Icons.arrow_back,
                        label: 'Left',
                        onPressed: () => gameProvider.miningMove('left'),
                      ),
                      const SizedBox(width: 8),
                      _DpadButton(
                        icon: Icons.arrow_forward,
                        label: 'Right',
                        onPressed: () => gameProvider.miningMove('right'),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  _DpadButton(
                    icon: Icons.arrow_downward,
                    label: 'Down',
                    onPressed: () => gameProvider.miningMove('down'),
                  ),
                ],
              ),
            ),
            if (mining.availableMoves.contains('slide')) ...[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton(
                  onPressed: () => gameProvider.miningMove('up', slide: true),
                  child: const Text('Slide (hold)'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _MiningStat extends StatelessWidget {
  final String icon;
  final String value;

  const _MiningStat(this.icon, this.value);

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(icon, style: const TextStyle(fontSize: 18)),
        Text(value, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: Colors.white)),
      ],
    );
  }
}

class _DpadButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onPressed;

  const _DpadButton({
    required this.icon,
    required this.label,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 56,
      height: 56,
      child: ElevatedButton.icon(
        onPressed: onPressed,
        icon: Icon(icon),
        label: Text(label, style: const TextStyle(fontSize: 10)),
        style: ElevatedButton.styleFrom(
          backgroundColor: AppTheme.primaryColor,
          padding: EdgeInsets.zero,
        ),
      ),
    );
  }
}

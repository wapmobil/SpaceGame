import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';

class BattleScreen extends StatelessWidget {
  const BattleScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Сражения')),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          final planet = gameProvider.selectedPlanet;
          if (planet == null) return const Center(child: Text('Планета не выбрана'));

          return RefreshIndicator(
            onRefresh: () async => gameProvider.loadBattles(planet.id),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildBattleGrid(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildBattleHistory(context, gameProvider),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildBattleGrid(BuildContext context, GameProvider gameProvider) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Поле боя', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 12),
            _build7x7Grid(context),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _LegendItem('🧑‍🚀 Игрок', AppTheme.accentColor),
                _LegendItem('👾 Враг', AppTheme.dangerColor),
                _LegendItem('🧱 Стена', Colors.grey),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _build7x7Grid(BuildContext context) {
    const size = 7;
    final cellSize = (MediaQuery.of(context).size.width - 80) / size;

    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: AppTheme.primaryColor, width: 2),
        borderRadius: BorderRadius.circular(8),
      ),
      child: GridView.builder(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: size,
          childAspectRatio: 1,
        ),
        itemCount: size * size,
        itemBuilder: (context, index) {
          final row = index ~/ size;
          final col = index % size;
          final isWall = row == 0 || row == size - 1 || col == 0 || col == size - 1;
          final isPlayer = row == size - 2 && col == size - 2;
          final isEnemy = row == 1 && col == 1;
          final isExit = row == 1 && col == size - 2;

          return Container(
            margin: const EdgeInsets.all(1),
            decoration: BoxDecoration(
              color: isWall
                  ? Colors.grey.shade800
                  : isPlayer
                      ? AppTheme.accentColor.withOpacity(0.3)
                      : isEnemy
                          ? AppTheme.dangerColor.withOpacity(0.3)
                          : isExit
                              ? AppTheme.successColor.withOpacity(0.3)
                              : AppTheme.cardColor,
              border: Border.all(
                color: isWall
                    ? Colors.grey
                    : isPlayer
                        ? AppTheme.accentColor
                        : isEnemy
                            ? AppTheme.dangerColor
                            : AppTheme.primaryColor,
                width: 1,
              ),
            ),
            child: Center(
              child: Text(
                isPlayer ? '🧑‍🚀' : isEnemy ? '👾' : isExit ? '🚪' : '',
                style: TextStyle(
                  fontSize: cellSize * 0.4,
                  color: isWall ? Colors.grey : Colors.white,
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildBattleHistory(BuildContext context, GameProvider gameProvider) {
    final battles = gameProvider.battles;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('История сражений', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            if (battles.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: Text('Сражений пока нет', style: TextStyle(color: Colors.white38))),
              )
            else
              ...battles.map((battle) => Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: ListTile(
                      leading: CircleAvatar(
                        backgroundColor: battle.status == 'completed' ? AppTheme.successColor : AppTheme.warningColor,
                        child: Text(
                          battle.status == 'completed' ? '✓' : '⚔',
                          style: const TextStyle(color: Colors.white),
                        ),
                      ),
                      title: Text('vs ${battle.opponent}'),
                      subtitle: Text(
                        battle.status == 'completed' ? 'Завершено' : 'В ожидании',
                        style: const TextStyle(fontSize: 12, color: Colors.white54),
                      ),
                      trailing: Text(
                        battle.status == 'completed' ? 'Завершено' : 'В ожидании',
                        style: TextStyle(
                          fontSize: 11,
                          color: battle.status == 'completed' ? AppTheme.successColor : AppTheme.warningColor,
                        ),
                      ),
                    ),
                  )),
          ],
        ),
      ),
    );
  }
}

class _LegendItem extends StatelessWidget {
  final String label;
  final Color color;

  const _LegendItem(this.label, this.color);

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(width: 12, height: 12, decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
        const SizedBox(width: 4),
        Text(label, style: const TextStyle(fontSize: 11, color: Colors.white54)),
      ],
    );
  }
}

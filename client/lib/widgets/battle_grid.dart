import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../models/battle.dart';

class BattleGrid extends StatelessWidget {
  final List<BattleGridCell> cells;
  final int gridSize;

  const BattleGrid({super.key, this.cells = const [], this.gridSize = 7});

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.of(context).size;
    final cellSize = (size.width - 80) / gridSize;

    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: AppTheme.primaryColor, width: 2),
        borderRadius: BorderRadius.circular(8),
      ),
      child: GridView.builder(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: gridSize,
          childAspectRatio: 1,
        ),
        itemCount: gridSize * gridSize,
        itemBuilder: (context, index) {
          final row = index ~/ gridSize;
          final col = index % gridSize;

          final cell = cells.where((c) => c.row == row && c.col == col).firstOrNull;
          final isWall = cell?.isWall ?? (row == 0 || row == gridSize - 1 || col == 0 || col == gridSize - 1);
          final isPlayer = cell?.isPlayer ?? false;
          final isEnemy = cell?.isEnemy ?? false;
          final isExit = cell?.isExit ?? false;

          return Container(
            margin: const EdgeInsets.all(1),
            decoration: BoxDecoration(
              color: _getCellColor(isWall, isPlayer, isEnemy, isExit, cell),
              border: Border.all(
                color: _getCellBorderColor(isWall, isPlayer, isEnemy, isExit),
                width: 1,
              ),
            ),
            child: Center(
              child: _getCellContent(isWall, isPlayer, isEnemy, isExit, cell, cellSize),
            ),
          );
        },
      ),
    );
  }

  Color _getCellColor(bool isWall, bool isPlayer, bool isEnemy, bool isExit, BattleGridCell? cell) {
    if (isWall) return Colors.grey.shade800;
    if (isPlayer) return AppTheme.accentColor.withValues(alpha: 0.3);
    if (isEnemy) return AppTheme.dangerColor.withValues(alpha: 0.3);
    if (isExit) return AppTheme.successColor.withValues(alpha: 0.3);
    if (cell?.isExplored ?? false) return AppTheme.cardColor;
    return AppTheme.cardColor.withValues(alpha: 0.5);
  }

  Color _getCellBorderColor(bool isWall, bool isPlayer, bool isEnemy, bool isExit) {
    if (isWall) return Colors.grey;
    if (isPlayer) return AppTheme.accentColor;
    if (isEnemy) return AppTheme.dangerColor;
    if (isExit) return AppTheme.successColor;
    return AppTheme.primaryColor;
  }

  Widget? _getCellContent(bool isWall, bool isPlayer, bool isEnemy, bool isExit, BattleGridCell? cell, double cellSize) {
    if (isWall) return null;
    if (isPlayer) return Text('🧑‍🚀', style: TextStyle(fontSize: cellSize * 0.45));
    if (isEnemy) return Text('👾', style: TextStyle(fontSize: cellSize * 0.45));
    if (isExit) return Text('🚪', style: TextStyle(fontSize: cellSize * 0.45));
    if (cell?.shipId != null) {
      return Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text('🚀', style: TextStyle(fontSize: cellSize * 0.35)),
          if (cell?.hp != null)
            Text(
              '${cell!.hp}',
              style: TextStyle(fontSize: cellSize * 0.2, color: Colors.white70),
            ),
        ],
      );
    }
    return null;
  }
}

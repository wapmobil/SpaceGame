import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../models/mining.dart';

class MiningMaze extends StatelessWidget {
  final MiningState state;
  final Function(String direction, {bool slide}) onMove;

  const MiningMaze({super.key, required this.state, required this.onMove});

  @override
  Widget build(BuildContext context) {
    final rows = state.maze.length;
    final cols = rows > 0 ? state.maze[0].length : 0;
    if (cols == 0) return const SizedBox.shrink();

    final size = MediaQuery.of(context).size;
    final cellSize = (size.width - 80) / cols;

    return Column(
      children: [
        Container(
          decoration: BoxDecoration(
            border: Border.all(color: AppTheme.primaryColor, width: 2),
            borderRadius: BorderRadius.circular(8),
          ),
          child: GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: cols,
              childAspectRatio: 1,
              crossAxisSpacing: 1,
              mainAxisSpacing: 1,
            ),
            itemCount: rows * cols,
            itemBuilder: (context, index) {
              final row = index ~/ cols;
              final col = index % cols;
              if (row >= state.maze.length || col >= state.maze[row].length) {
                return const SizedBox.shrink();
              }

              final cell = state.maze[row][col];
              final isPlayer = row == state.playerY && col == state.playerX;
              final isExit = row == state.exitY && col == state.exitX;
              final monster = state.monsters.where((m) => m.alive && m.x == col && m.y == row).toList();

              return _MazeCell(
                cell: cell,
                isPlayer: isPlayer,
                isExit: isExit,
                monster: monster.isNotEmpty ? monster.first : null,
                cellSize: cellSize,
              );
            },
          ),
        ),
        if (state.canMove) ...[
          const SizedBox(height: 12),
          _buildControls(),
        ],
      ],
    );
  }

  Widget _buildControls() {
    return Column(
      children: [
        _DpadButton(
          icon: Icons.arrow_upward,
          onPressed: () => onMove('up'),
        ),
        const SizedBox(height: 4),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            _DpadButton(
              icon: Icons.arrow_back,
              onPressed: () => onMove('left'),
            ),
            const SizedBox(width: 8),
            _DpadButton(
              icon: Icons.arrow_forward,
              onPressed: () => onMove('right'),
            ),
          ],
        ),
        const SizedBox(height: 4),
        _DpadButton(
          icon: Icons.arrow_downward,
          onPressed: () => onMove('down'),
        ),
      ],
    );
  }
}

class _MazeCell extends StatelessWidget {
  final String cell;
  final bool isPlayer;
  final bool isExit;
  final MiningMonster? monster;
  final double cellSize;

  const _MazeCell({
    required this.cell,
    required this.isPlayer,
    required this.isExit,
    this.monster,
    required this.cellSize,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: _getCellColor(),
        border: Border.all(color: Colors.white12, width: 0.5),
      ),
      child: Center(
        child: _getCellContent(),
      ),
    );
  }

  Color _getCellColor() {
    if (isPlayer) return AppTheme.accentColor.withValues(alpha: 0.4);
    if (isExit) return AppTheme.successColor.withValues(alpha: 0.4);
    if (cell == 'wall') return Colors.grey.shade800;
    return AppTheme.cardColor;
  }

  Widget? _getCellContent() {
    if (isPlayer) return Text('🧑‍🚀', style: TextStyle(fontSize: cellSize * 0.45));
    if (isExit) return Text('🚪', style: TextStyle(fontSize: cellSize * 0.45));
    if (monster != null) return Text(monster!.icon, style: TextStyle(fontSize: cellSize * 0.45));
    if (cell == 'wall') return null;
    if (cell == 'money') return Text('💰', style: TextStyle(fontSize: cellSize * 0.3));
    return null;
  }
}

class _DpadButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback onPressed;

  const _DpadButton({required this.icon, required this.onPressed});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 56,
      height: 56,
      child: ElevatedButton.icon(
        onPressed: onPressed,
        icon: Icon(icon),
        label: const SizedBox.shrink(),
        style: ElevatedButton.styleFrom(
          backgroundColor: AppTheme.primaryColor,
          padding: EdgeInsets.zero,
          minimumSize: const Size(56, 56),
        ),
      ),
    );
  }
}

import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../models/drill.dart';

class DrillScreen extends StatefulWidget {
  final String planetId;
  final VoidCallback? onBack;

  const DrillScreen({
    super.key,
    required this.planetId,
    this.onBack,
  });

  @override
  State<DrillScreen> createState() => _DrillScreenState();
}

class _DrillScreenState extends State<DrillScreen> {
  bool _extracting = false;
  Timer? _extractTimer;
  Timer? _autoTickTimer;
  String _lastMessage = '';
  bool _horizontalCooldown = false;
  bool _resultShown = false;

  @override
  void initState() {
    super.initState();
    _loadDrillState();
  }

  @override
  void dispose() {
    _extractTimer?.cancel();
    _autoTickTimer?.cancel();
    _resultShown = false;
    super.dispose();
  }

  void _startAutoTick() {
    _autoTickTimer?.cancel();
    _autoTickTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      final state = context.read<GameProvider>().drillState;
      if (state == null || !state.isActive) {
        timer.cancel();
        return;
      }
      if (mounted) {
        setState(() {
          _horizontalCooldown = false;
        });
      }
      context.read<GameProvider>().loadDrillState(widget.planetId);
    });
  }

  void _stopAutoTick() {
    _autoTickTimer?.cancel();
    _autoTickTimer = null;
  }

  Future<void> _loadDrillState() async {
    await context.read<GameProvider>().loadDrillState(widget.planetId);
    final state = context.read<GameProvider>().drillState;
    if (state?.isActive == true) {
      _startAutoTick();
    }
  }

  void _startExtracting() {
    setState(() {
      _extracting = true;
    });
    _doExtract();
  }

  void _stopExtracting() {
    setState(() {
      _extracting = false;
    });
  }

  Future<void> _doExtract() async {
    if (_extracting && context.read<GameProvider>().drillState?.isActive == true) {
      await context.read<GameProvider>().drillMove('down', extract: true);
    }
  }

  Future<void> _move(String direction) async {
    if (context.read<GameProvider>().drillState?.isActive != true) return;
    if (direction == 'left' || direction == 'right') {
      if (_horizontalCooldown) return;
      setState(() {
        _horizontalCooldown = true;
      });
    }
    await context.read<GameProvider>().drillMove(direction);
    if (mounted) setState(() {});
  }

  Future<void> _completeDrill() async {
    if (context.read<GameProvider>().drillState?.isActive != true) return;
    await context.read<GameProvider>().completeDrill();
    _stopAutoTick();
  }

  Future<void> _cancelDrill() async {
    if (context.read<GameProvider>().drillState?.isActive != true) return;
    _stopAutoTick();
    await context.read<GameProvider>().destroyDrill();
  }

  Future<void> _startDrill() async {
    await context.read<GameProvider>().startDrill();
    final state = context.read<GameProvider>().drillState;
    if (state?.isActive == true) {
      _startAutoTick();
    }
    if (mounted) setState(() {});
  }

  void _showResult() {
    final state = context.read<GameProvider>().drillState;
    if (state == null) return;

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(state.isCompleted ? 'Добыча завершена!' : 'Бур разрушен!'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Глубина: ${state.depth}'),
            Text('Заработано: \$${state.totalEarned.toStringAsFixed(0)}'),
            const SizedBox(height: 16),
            const Text('Собранные ресурсы:'),
            ...state.resources.map((r) => Padding(
              padding: const EdgeInsets.only(left: 8, top: 4),
              child: Row(
                children: [
                  Text(r.icon),
                  const SizedBox(width: 8),
                  Text('${r.name}: ${r.amount.toStringAsFixed(0)} (\$${r.value.toStringAsFixed(0)})'),
                ],
              ),
            )),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () async {
              await context.read<GameProvider>().cleanupDrill();
              context.read<GameProvider>().clearDrillState();
              _resultShown = false;
              Navigator.of(ctx).pop();
              Navigator.of(context).pop();
            },
            child: const Text('Закрыть'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<GameProvider>().drillState;

    if (state == null || state.status == 'no_session') {
      _stopAutoTick();
      _resultShown = false;
      return _buildStartScreen();
    }

    if (state.isGameEnded && !_resultShown) {
      _stopAutoTick();
      _resultShown = true;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) _showResult();
      });
    } else if (!state.isGameEnded) {
      _resultShown = false;
    }

    return _buildGameScreen(state);
  }

  Widget _buildLoading() {
    return const Scaffold(
      body: Center(child: CircularProgressIndicator()),
    );
  }

  Widget _buildStartScreen() {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Бурение'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.dns, size: 80, color: Colors.grey),
            const SizedBox(height: 24),
            const Text(
              'Шахта готова к бурению',
              style: TextStyle(fontSize: 24),
            ),
            const SizedBox(height: 16),
            const Text(
              'Бур будет погружаться глубоко под поверхность,\nсобирая ценные ресурсы',
              style: TextStyle(fontSize: 16, color: Colors.grey),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            ElevatedButton.icon(
              onPressed: _startDrill,
              icon: const Icon(Icons.play_arrow),
              label: const Text('Начать бурение'),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
                textStyle: const TextStyle(fontSize: 18),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildGameScreen(DrillState state) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Бурение'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            if (state.isActive || state.isGameEnded) {
              context.read<GameProvider>().clearDrillState();
            }
            Navigator.of(context).pop();
          },
        ),
        actions: [
          if (state.isActive)
            IconButton(
              icon: const Icon(Icons.exit_to_app),
              onPressed: _cancelDrill,
              tooltip: 'Выйти',
            ),
        ],
      ),
      body: Column(
        children: [
          _buildHUD(state),
          Expanded(child: _buildWorld(state)),
          _buildControls(state),
        ],
      ),
    );
  }

  Widget _buildHUD(DrillState state) {
    return Container(
      padding: const EdgeInsets.all(8),
      color: Colors.black87,
      child: Column(
        children: [
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Text('🔧 Прочность: ', style: TextStyle(color: Colors.white, fontSize: 12)),
                        Expanded(
                          child: LinearProgressIndicator(
                            value: state.hpPercent,
                            backgroundColor: Colors.grey[800],
                            valueColor: AlwaysStoppedAnimation(
                              state.hpPercent > 0.5 ? Colors.green : state.hpPercent > 0.25 ? Colors.orange : Colors.red,
                            ),
                          ),
                        ),
                        Text('${state.drillHp}/${state.drillMaxHp}',
                            style: const TextStyle(color: Colors.white, fontSize: 12)),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Text('Глубина: ${state.depth}',
                        style: const TextStyle(color: Colors.white70, fontSize: 12)),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.amber.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Text('💰 ', style: TextStyle(fontSize: 14)),
                    Text('\$${state.totalEarned.toStringAsFixed(0)}',
                        style: const TextStyle(color: Colors.amber, fontSize: 14, fontWeight: FontWeight.bold)),
                  ],
                ),
              ),
            ],
          ),
          if (_extracting)
            Container(
              margin: const EdgeInsets.only(top: 4),
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: Colors.orange.withOpacity(0.3),
                borderRadius: BorderRadius.circular(4),
              ),
              child: const Text('⛏️ Добыча...',
                  style: TextStyle(color: Colors.orange, fontSize: 11)),
            ),
          if (state.resources.isNotEmpty)
            Container(
              margin: const EdgeInsets.only(top: 4),
              child: SizedBox(
                height: 28,
                child: ListView.builder(
                  scrollDirection: Axis.horizontal,
                  itemCount: state.resources.length,
                  itemBuilder: (ctx, idx) {
                    final r = state.resources[idx];
                    return Container(
                      margin: EdgeInsets.only(right: idx < state.resources.length - 1 ? 4 : 0),
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(r.icon, style: const TextStyle(fontSize: 14)),
                          const SizedBox(width: 2),
                          Text('${r.amount.toStringAsFixed(0)}',
                              style: const TextStyle(color: Colors.white, fontSize: 10)),
                        ],
                      ),
                    );
                  },
                ),
              ),
            ),
          if (_lastMessage.isNotEmpty)
            Container(
              margin: const EdgeInsets.only(top: 2),
              child: Text(_lastMessage,
                  style: const TextStyle(color: Colors.white70, fontSize: 10)),
            ),
        ],
      ),
    );
  }

  Widget _buildWorld(DrillState state) {
    final world = state.world;
    if (world.isEmpty) {
      return const Center(child: Text('Мир генерируется...'));
    }

    final viewHeight = world.length;
    final worldWidth = state.worldWidth;

    return LayoutBuilder(
      builder: (context, constraints) {
        final cellSize = (constraints.maxWidth / worldWidth).clamp(16.0, 60.0);
        final totalWidth = cellSize * worldWidth;
        final totalHeight = cellSize * viewHeight;

        return SingleChildScrollView(
          child: Center(
            child: SizedBox(
              width: totalWidth,
              height: totalHeight,
              child: Column(
                children: List.generate(viewHeight, (rowIdx) {
                  return Row(
                    children: List.generate(world[rowIdx].length, (colIdx) {
                      final cell = world[rowIdx][colIdx];
                      return _buildCell(cell, cellSize, rowIdx, colIdx, state);
                    }),
                  );
                }),
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildCell(DrillCell cell, double size, int rowIdx, int colIdx, DrillState state) {
    Color bgColor;
    String content;

    final isDrillPos = rowIdx == 0 && colIdx == 2;

    if (isDrillPos) {
      bgColor = Colors.amber;
      content = '⛏️';
    } else if (cell.extracted) {
      bgColor = Colors.grey[900]!;
      content = '';
    } else {
      switch (cell.cellType) {
        case 'dirt':
          bgColor = Colors.brown[400]!;
          content = '';
          break;
        case 'stone':
          bgColor = Colors.grey[600]!;
          content = '';
          break;
        case 'metal':
          bgColor = Colors.grey[800]!;
          content = '';
          break;
        case 'mithril':
          bgColor = Colors.purple[800]!;
          content = '';
          break;
        default:
          bgColor = Colors.black;
          content = '';
      }

      if (cell.resourceType != null && !cell.extracted) {
        content = _getResourceIcon(cell.resourceType!);
      }
    }

    return Container(
      width: size,
      height: size,
      color: bgColor,
      child: Center(
        child: content.isNotEmpty
            ? Text(content, style: TextStyle(fontSize: size * 0.6))
            : null,
      ),
    );
  }

  String _getResourceIcon(String type) {
    switch (type) {
      case 'oil': return '🛢️';
      case 'gas': return '💨';
      case 'copper': return '🟠';
      case 'coal': return '⬛';
      case 'silver': return '⚪';
      case 'gold': return '🟡';
      case 'platinum': return '🔘';
      case 'diamond': return '💎';
      case 'exotic': return '🔮';
      default: return '❓';
    }
  }

  Widget _buildControls(DrillState state) {
    return Container(
      padding: const EdgeInsets.all(16),
      color: Colors.black87,
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              _buildControlButton(
                icon: Icons.arrow_back,
                label: '←',
                onPressed: () => _move('left'),
                enabled: state.isActive && !_horizontalCooldown,
              ),
              const SizedBox(width: 16),
              _buildExtractButton(state),
              const SizedBox(width: 16),
              _buildControlButton(
                icon: Icons.arrow_forward,
                label: '→',
                onPressed: () => _move('right'),
                enabled: state.isActive && !_horizontalCooldown,
              ),
            ],
          ),
          if (_horizontalCooldown)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text('ожидание...',
                  style: TextStyle(color: Colors.grey[500], fontSize: 10)),
            ),
        ],
      ),
    );
  }

  Widget _buildControlButton({
    required IconData icon,
    required String label,
    required VoidCallback onPressed,
    required bool enabled,
  }) {
    return Column(
      children: [
        ElevatedButton.icon(
          onPressed: enabled ? onPressed : null,
          icon: Icon(icon, size: 20),
          label: Text(label, style: const TextStyle(fontSize: 10)),
          style: ElevatedButton.styleFrom(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            backgroundColor: enabled ? Colors.blue[700] : Colors.grey[700],
          ),
        ),
      ],
    );
  }

  Widget _buildExtractButton(DrillState state) {
    return GestureDetector(
      onTapDown: (_) => _startExtracting(),
      onTapUp: (_) => _stopExtracting(),
      onTapCancel: _stopExtracting,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: _extracting ? Colors.orange[700] : Colors.orange[900],
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: _extracting ? Colors.orange : Colors.orange[700]!),
        ),
        child: Column(
          children: [
            const Icon(Icons.build, size: 24, color: Colors.white),
            const SizedBox(height: 2),
            Text(
              _extracting ? 'Добыча...' : 'Добыча',
              style: TextStyle(color: Colors.white, fontSize: 10),
            ),
          ],
        ),
      ),
    );
  }
}

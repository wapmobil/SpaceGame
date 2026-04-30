import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
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
  final String _lastMessage = '';
  bool _resultShown = false;
  final _keyboardFocusNode = FocusNode();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _keyboardFocusNode.requestFocus();
    });
  }

  @override
  void dispose() {
    _keyboardFocusNode.dispose();
    if (_extracting) {
      _stopExtracting();
    }
    _resultShown = false;
    super.dispose();
  }

  void _startExtracting() {
    setState(() {
      _extracting = true;
    });
    context.read<GameProvider>().drillCommand(extract: true);
  }

  void _stopExtracting() {
    setState(() {
      _extracting = false;
    });
    context.read<GameProvider>().drillCommand(extract: false);
  }

  void _move(String direction) {
    if (context.read<GameProvider>().drillState?.isActive != true) return;
    context.read<GameProvider>().drillCommand(direction: direction);
    if (mounted) setState(() {});
  }

  Future<void> _cancelDrill() async {
    if (context.read<GameProvider>().drillState?.isActive != true) return;
    await context.read<GameProvider>().destroyDrill();
  }

  Future<void> _startDrill() async {
    await context.read<GameProvider>().startDrill(speed: 1);
    if (mounted) setState(() {});
  }

  Future<void> _startDrill2x() async {
    await context.read<GameProvider>().startDrill(speed: 2);
    if (mounted) setState(() {});
  }

  void _showResult() {
    final state = context.read<GameProvider>().drillState;
    if (state == null) return;

    showDialog(
      context: context,
      barrierDismissible: false,
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
               final dialogContext = ctx;
               final dialogNavigator = Navigator.of(dialogContext);
               final gameProvider = context.read<GameProvider>();
               await gameProvider.cleanupDrill();
               if (!mounted || !dialogNavigator.canPop()) return;
               gameProvider.clearDrillState();
               _resultShown = false;
               dialogNavigator.pop();
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
      _resultShown = false;
      return _buildStartScreen();
    }

    if (state.isGameEnded && !_resultShown) {
      _resultShown = true;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) _showResult();
      });
    } else if (!state.isGameEnded) {
      _resultShown = false;
    }

    return _buildGameScreen(state);
  }

  Widget _buildStartScreen() {
    final provider = context.watch<GameProvider>();
    final mineLevel = provider.getBuildingLevelForPlanet(widget.planetId, 'mine');
    final cost = 100 * mineLevel;
    final ironAvailable = (provider.selectedPlanet?.resources['iron'] ?? 0) as num? ?? 0;
    final canAfford = ironAvailable >= cost;
    final color = canAfford ? Colors.amber : Colors.red;
    final canAfford2x = canAfford && mineLevel >= 4;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Бурение'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: SingleChildScrollView(
        child: Center(
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
                'Бур будет погружаться глубоко под поверхность,\nсобирая ценные ресурсы. Прочность: 10+100*ур.шахты',
                style: TextStyle(fontSize: 16, color: Colors.grey),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 24),
              ElevatedButton.icon(
                onPressed: canAfford ? _startDrill : null,
                icon: const Icon(Icons.play_arrow),
                label: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Начать бурение'),
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Text('🪨', style: TextStyle(fontSize: 12)),
                        const SizedBox(width: 2),
                        Text(
                          cost.toString(),
                          style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: color),
                        ),
                      ],
                    ),
                  ],
                ),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
                ),
              ),
              if (mineLevel >= 1) ...[
                const SizedBox(height: 12),
                ElevatedButton.icon(
                  onPressed: canAfford2x ? _startDrill2x : null,
                  icon: const Icon(Icons.speed),
                  label: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Text('Начать бурение'),
                          const SizedBox(width: 8),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(
                              color: Colors.orange.shade700,
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: const Text(
                              '2x',
                              style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.white),
                            ),
                          ),
                        ],
                      ),
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Text('🪨', style: TextStyle(fontSize: 12)),
                          const SizedBox(width: 2),
                          Text(
                            cost.toString(),
                            style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: canAfford2x ? Colors.amber : Colors.red),
                          ),
                          const SizedBox(width: 8),
                          Text(
                            mineLevel >= 4 ? '(ур.шахты >= 4)' : '(нужен ур.шахты 4)',
                            style: const TextStyle(fontSize: 11, color: Colors.white54),
                          ),
                        ],
                      ),
                    ],
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: canAfford2x ? Colors.orange.shade700 : Colors.grey.shade700,
                    padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
                  ),
                ),
              ],
              const SizedBox(height: 24),
              Container(
                margin: const EdgeInsets.symmetric(horizontal: 32),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.05),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Управление',
                      style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white),
                    ),
                    const SizedBox(height: 8),
                    _buildInstructionRow(Icons.arrow_back, '← Влево'),
                    _buildInstructionRow(Icons.arrow_forward, 'Вправо →'),
                    _buildInstructionRow(Icons.keyboard, 'Пробел — добыча (удерживать)'),
                    const SizedBox(height: 8),
                    const Text(
                      'Бур спускается автоматически каждые 1 сек.\nНаправления и добыча применяются при спуске.',
                      style: TextStyle(fontSize: 12, color: Colors.white70),
                    ),
                    const SizedBox(height: 12),
                    const Text(
                      'Урон по буру',
                      style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: Colors.white),
                    ),
                    const SizedBox(height: 4),
                    _buildDamageRow('Грязь', '2', Colors.brown),
                    _buildDamageRow('Камень', '5', Colors.grey),
                    _buildDamageRow('Металл', '10', Colors.grey),
                    _buildDamageRow('Мифрил', '15', Colors.purple),
                    const SizedBox(height: 6),
                    const Text(
                      '⚠ Ресурс без добычи: +5\n⚠ Добыча без ресурса: +3',
                      style: TextStyle(fontSize: 11, color: Colors.orange),
                    ),
                    const SizedBox(height: 12),
                    const Text(
                      'Ресурсы',
                      style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: Colors.white),
                    ),
                    const SizedBox(height: 4),
                    _buildResourceRow('🛢️ Нефть', '1\$', '0–1000'),
                    _buildResourceRow('💨 Газ', '2\$', '0–1000'),
                    _buildResourceRow('⬛ Уголь', '5\$', '50–150'),
                    _buildResourceRow('🟠 Медь', '10\$', '50–100'),
                    _buildResourceRow('⚪ Серебро', '15\$', '100–200'),
                    _buildResourceRow('🟡 Золото', '25\$', '150–300'),
                    _buildResourceRow('🔘 Платина', '30\$', '200–400'),
                    _buildResourceRow('💎 Алмазы', '60\$', '300–500'),
                    _buildResourceRow('🔮 Экзотика', '200\$', '500+'),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildInstructionRow(IconData icon, String label) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: Row(
        children: [
          Icon(icon, size: 16, color: Colors.white70),
          const SizedBox(width: 8),
          Text(label, style: const TextStyle(fontSize: 13, color: Colors.white70)),
        ],
      ),
    );
  }

  Widget _buildDamageRow(String name, String damage, Color color) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 3),
      child: Row(
        children: [
          Container(
            width: 16,
            height: 16,
            decoration: BoxDecoration(
              color: color,
              borderRadius: BorderRadius.circular(3),
            ),
          ),
          const SizedBox(width: 8),
          Text(name, style: const TextStyle(fontSize: 12, color: Colors.white70)),
          const Spacer(),
          Text('$damage урона', style: const TextStyle(fontSize: 12, color: Colors.white70)),
        ],
      ),
    );
  }

  Widget _buildResourceRow(String icon, String value, String depth) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 2),
      child: Row(
        children: [
          Text(icon, style: const TextStyle(fontSize: 13)),
          const SizedBox(width: 6),
          Text(value, style: const TextStyle(fontSize: 12, color: Colors.amber)),
          const Spacer(),
          Text(depth, style: const TextStyle(fontSize: 11, color: Colors.white54)),
        ],
      ),
    );
  }

  Widget _buildGameScreen(DrillState state) {
    return KeyboardListener(
      focusNode: _keyboardFocusNode,
      onKeyEvent: (event) {
        if (state.isActive != true) return;
        if (event is KeyDownEvent) {
          if (event.logicalKey == LogicalKeyboardKey.arrowLeft) {
            _move('left');
          } else if (event.logicalKey == LogicalKeyboardKey.arrowRight) {
            _move('right');
          } else if (event.logicalKey == LogicalKeyboardKey.space && !_extracting) {
            _startExtracting();
          }
        } else if (event is KeyUpEvent) {
          if (event.logicalKey == LogicalKeyboardKey.space && _extracting) {
            _stopExtracting();
          }
        }
      },
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Бурение'),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back),
            onPressed: () {
              if (state.isActive || state.isGameEnded) {
                if (_extracting) {
                  _stopExtracting();
                }
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
        body: Focus(
          autofocus: true,
          child: Column(
            children: [
              _buildHUD(state),
              Expanded(child: _buildWorld(state)),
              Container(
                color: Colors.black87,
                child: _buildStatusStrip(state),
              ),
              _buildControls(state),
            ],
          ),
        ),
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
                  color: Colors.amber.withValues(alpha: 0.2),
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
                        color: Colors.white.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(r.icon, style: const TextStyle(fontSize: 14)),
                          const SizedBox(width: 2),
                          Text(r.amount.toStringAsFixed(0),
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

  Widget _buildStatusStrip(DrillState state) {
    final provider = context.read<GameProvider>();
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      color: Colors.black87,
      child: Column(
        children: [
          if (provider.drillPendingExtracting || state.pendingExtracting)
            Container(
              margin: const EdgeInsets.only(bottom: 2),
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: Colors.orange.withValues(alpha: 0.3),
                borderRadius: BorderRadius.circular(4),
              ),
              child: const Text('🛡️ Добыча...',
                  style: TextStyle(color: Colors.orange, fontSize: 11)),
            ),
          if (provider.drillPendingDirection != null && provider.drillPendingDirection != '')
            Container(
              margin: const EdgeInsets.only(bottom: 2),
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: Colors.blue.withValues(alpha: 0.3),
                borderRadius: BorderRadius.circular(4),
              ),
              child: Text('→ ${provider.drillPendingDirection}',
                  style: const TextStyle(color: Colors.blue, fontSize: 11)),
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
      content = '🛡️';
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
                enabled: state.isActive,
              ),
              const SizedBox(width: 16),
              _buildExtractButton(state),
              const SizedBox(width: 16),
              _buildControlButton(
                icon: Icons.arrow_forward,
                label: '→',
                onPressed: () => _move('right'),
                enabled: state.isActive,
              ),
            ],
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
               style: const TextStyle(color: Colors.white, fontSize: 10),
             ),
           ],
        ),
      ),
    );
  }
}

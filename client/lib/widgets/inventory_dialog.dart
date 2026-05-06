import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';

class InventoryDialog extends StatefulWidget {
  final String planetId;
  final VoidCallback onSubmitted;

  const InventoryDialog({
    super.key,
    required this.planetId,
    required this.onSubmitted,
  });

  @override
  State<InventoryDialog> createState() => _InventoryDialogState();
}

class _InventoryDialogState extends State<InventoryDialog> {
  final Map<String, double> _allocations = {
    'food': 0,
    'iron': 0,
    'composite': 0,
    'mechanisms': 0,
    'reagents': 0,
  };

  double get _total => _allocations.values.fold(0.0, (sum, v) => sum + v);

  double _getMaxForResource(String resource) {
    final available = (context.read<GameProvider>().selectedPlanet?.resources[resource] as num?)?.toDouble() ?? 0;
    final remaining = 1000 - (_total - _allocations[resource]!);
    return (available < remaining ? available : remaining).clamp(0, 1000);
  }

  void _onSubmitted() async {
    final gameProvider = context.read<GameProvider>();
    final chainProvider = gameProvider.expeditionChainProvider;

    if (_total <= 0 || _total > 1000) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Распределите ресурсы (сумма от 1 до 1000)'),
          backgroundColor: AppTheme.dangerColor,
        ),
      );
      return;
    }

    try {
      await chainProvider.startExpedition(widget.planetId, _allocations);
      if (mounted) {
        Navigator.of(context).pop(true);
        widget.onSubmitted();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Ошибка запуска экспедиции: $e'),
            backgroundColor: AppTheme.dangerColor,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final planet = context.read<GameProvider>().selectedPlanet;
    final resources = planet?.resources ?? {};

    return AlertDialog(
      backgroundColor: AppTheme.cardColor,
      title: const Text(
        'Подготовка экспедиции',
        style: TextStyle(color: Colors.white, fontSize: 18),
      ),
      content: SizedBox(
        width: double.maxFinite,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Распределите ресурсы между экспедицией (макс. 1000 суммарно)',
              style: TextStyle(color: Colors.white70, fontSize: 13),
            ),
            const SizedBox(height: 16),
            ...Constants.resourceNames.keys.where((k) =>
                k == 'food' || k == 'iron' || k == 'composite' || k == 'mechanisms' || k == 'reagents'
            ).map((resource) {
              final maxVal = _getMaxForResource(resource);
              final currentVal = _allocations[resource]!;
              final available = (resources[resource] as num?)?.toDouble() ?? 0;

              return Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          '${Constants.resourceIcons[resource]} ${Constants.resourceNames[resource] ?? resource}',
                          style: const TextStyle(color: Colors.white, fontSize: 13),
                        ),
                        Text(
                          'Доступно: ${available.toInt()}',
                          style: const TextStyle(color: Colors.white54, fontSize: 11),
                        ),
                      ],
                    ),
                    Row(
                      children: [
                        Expanded(
                          child: Slider(
                            value: currentVal,
                            min: 0,
                            max: maxVal,
                            divisions: maxVal > 0 ? (maxVal * 10).toInt() : 10,
                            label: currentVal.toInt().toString(),
                            activeColor: AppTheme.accentColor,
                            onChanged: (value) {
                              setState(() {
                                _allocations[resource] = value;
                              });
                            },
                          ),
                        ),
                        SizedBox(
                          width: 50,
                          child: Text(
                            '${currentVal.toInt()}',
                            textAlign: TextAlign.end,
                            style: const TextStyle(color: Colors.white, fontSize: 12),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              );
            }),
            const Divider(color: Colors.white24),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Итого:',
                  style: TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.bold),
                ),
                Text(
                  '${_total.toInt()} / 1000',
                  style: TextStyle(
                    color: _total > 1000 ? AppTheme.dangerColor : Colors.white,
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(false),
          child: const Text('Отмена', style: TextStyle(color: Colors.white54)),
        ),
        ElevatedButton(
          onPressed: (_total <= 0 || _total > 1000) ? null : _onSubmitted,
          style: ElevatedButton.styleFrom(
            backgroundColor: AppTheme.accentColor,
            foregroundColor: Colors.white,
            disabledBackgroundColor: Colors.white24,
          ),
          child: const Text('Отправить'),
        ),
      ],
    );
  }
}

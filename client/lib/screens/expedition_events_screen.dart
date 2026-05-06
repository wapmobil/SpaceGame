import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../providers/expedition_chain_provider.dart';
import '../models/expedition_chain.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../widgets/inventory_dialog.dart';
import '../widgets/location_card.dart';

class ExpeditionEventsScreen extends StatefulWidget {
  final String planetId;

  const ExpeditionEventsScreen({super.key, required this.planetId});

  @override
  State<ExpeditionEventsScreen> createState() => _ExpeditionEventsScreenState();
}

class _ExpeditionEventsScreenState extends State<ExpeditionEventsScreen> {
  String? _selectedChainId;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final chains = context.read<ExpeditionChainProvider>().activeChains;
    if (chains.isNotEmpty && _selectedChainId == null) {
      _selectedChainId = chains.first.id;
    }
  }

  void _onExpeditionStarted() {
    final chainProvider = context.read<ExpeditionChainProvider>();
    final chains = chainProvider.activeChains;
    if (chains.isNotEmpty) {
      setState(() {
        _selectedChainId = chains.first.id;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<GameProvider>(
      builder: (context, gameProvider, _) {
        final chainProvider = gameProvider.expeditionChainProvider;
        final activeChains = chainProvider.activeChains;
        final selectedChain = chainProvider.selectedChain;
        final currentEvent = chainProvider.currentEvent;
        final isLoading = chainProvider.isLoading;

        return Scaffold(
          appBar: AppBar(title: const Text('Экспедиции')),
          body: RefreshIndicator(
            onRefresh: () async {
              if (gameProvider.selectedPlanet != null) {
                await chainProvider.loadExpeditionChains(widget.planetId);
              }
            },
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Section 1: Active expeditions
                  _buildActiveExpeditions(
                    activeChains,
                    selectedChain,
                    currentEvent,
                    isLoading,
                    chainProvider,
                  ),
                  const SizedBox(height: 16),

                  // Section 2: Event history
                  _buildEventHistory(selectedChain, chainProvider),
                  const SizedBox(height: 16),

                  // Section 3: Start new expedition
                  _buildStartButton(chainProvider),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildActiveExpeditions(
    List<ExpeditionChain> activeChains,
    ExpeditionChain? selectedChain,
    ExpeditionEvent? currentEvent,
    bool isLoading,
    ExpeditionChainProvider chainProvider,
  ) {
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
                  'Активные экспедиции',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70),
                ),
                Text(
                  '${activeChains.length}/1',
                  style: const TextStyle(fontSize: 12, color: Colors.white54),
                ),
              ],
            ),
            const SizedBox(height: 12),
            if (activeChains.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(
                  child: Text('Нет активных экспедиций', style: TextStyle(color: Colors.white38)),
                ),
              )
            else if (activeChains.length > 1)
              ...activeChains.map((chain) => Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: InkWell(
                      onTap: () {
                        chainProvider.selectChain(chain.id);
                        setState(() {
                          _selectedChainId = chain.id;
                        });
                      },
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        decoration: BoxDecoration(
                          color: chain.id == selectedChain?.id
                              ? AppTheme.accentColor.withValues(alpha: 0.2)
                              : Colors.white10,
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(
                            color: chain.id == selectedChain?.id
                                ? AppTheme.accentColor
                                : Colors.transparent,
                          ),
                        ),
                        child: Text(
                          'Экспедиция #${chain.currentEventIndex + 1}',
                          style: const TextStyle(color: Colors.white, fontSize: 13),
                        ),
                      ),
                    ),
                  )),
            if (selectedChain != null && currentEvent != null)
              _buildEventCard(selectedChain, currentEvent, isLoading, chainProvider)
            else if (selectedChain != null && selectedChain.isGenerating)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 24),
                child: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      CircularProgressIndicator(),
                      SizedBox(height: 12),
                      Text(
                        'Идёт экспедиция...',
                        style: TextStyle(color: Colors.white54, fontSize: 13),
                      ),
                    ],
                  ),
                ),
              )
            else if (selectedChain != null && currentEvent == null && !isLoading)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: Text('Событие не загружено', style: TextStyle(color: Colors.white38))),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildEventCard(
    ExpeditionChain chain,
    ExpeditionEvent event,
    bool isLoading,
    ExpeditionChainProvider chainProvider,
  ) {
    final eventIndex = chain.currentEventIndex;
    final eventCount = chain.eventCount;

    return Card(
      elevation: 0,
      color: AppTheme.cardColor.withValues(alpha: 0.5),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Event header
            Row(
              children: [
                const Text('🔍', style: TextStyle(fontSize: 18)),
                const SizedBox(width: 8),
                Text(
                  'Событие ${eventIndex + 1}/$eventCount',
                  style: const TextStyle(fontWeight: FontWeight.w600, color: Colors.white, fontSize: 14),
                ),
              ],
            ),
            const SizedBox(height: 8),

            // Event description
            Text(
              event.description,
              style: const TextStyle(color: Colors.white70, fontSize: 13),
            ),
            const SizedBox(height: 8),

            // Instant reward chips
            if (event.immediateReward.isNotEmpty) ...[
              Wrap(
                spacing: 4,
                runSpacing: 4,
                children: event.immediateReward.entries.map((entry) {
                  final value = entry.value;
                  final sign = value >= 0 ? '+' : '';
                  return Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: value >= 0
                          ? AppTheme.successColor.withValues(alpha: 0.2)
                          : AppTheme.dangerColor.withValues(alpha: 0.2),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      '$sign${value.toInt()}',
                      style: TextStyle(
                        fontSize: 10,
                        color: value >= 0 ? AppTheme.successColor : AppTheme.dangerColor,
                      ),
                    ),
                  );
                }).toList(),
              ),
              const SizedBox(height: 8),
            ],

            // Expedition inventory
            const Text(
              'Инвентарь экспедиции:',
              style: TextStyle(fontSize: 12, color: Colors.white54),
            ),
            const SizedBox(height: 4),
            SizedBox(
              height: 32,
              child: ListView(
                scrollDirection: Axis.horizontal,
                children: chain.inventory.entries.map((entry) {
                  if (entry.value <= 0) return const SizedBox.shrink();
                  return Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text(Constants.resourceIcons[entry.key] ?? '❓', style: const TextStyle(fontSize: 14)),
                        const SizedBox(width: 2),
                        Text(
                          entry.value.toInt().toString(),
                          style: const TextStyle(fontSize: 11, color: Colors.white70),
                        ),
                      ],
                    ),
                  );
                }).toList(),
              ),
            ),
            const SizedBox(height: 12),

            // Choice buttons or end state
            if (event.isEnd)
              _buildEndState(chain, event)
            else if (isLoading)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      CircularProgressIndicator(),
                      SizedBox(height: 8),
                      Text('Обработка выбора...', style: TextStyle(color: Colors.white54)),
                    ],
                  ),
                ),
              )
            else
              ...event.choices.asMap().entries.map((entry) => _buildChoiceButton(entry.value, entry.key, chain, chainProvider)),
          ],
        ),
      ),
    );
  }

  Widget _buildEndState(ExpeditionChain chain, ExpeditionEvent event) {
    final location = chain.discoveredLocation;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          '🎉 Экспедиция завершена!',
          style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: AppTheme.successColor),
        ),
        const SizedBox(height: 8),
        if (location != null)
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Обнаружена локация:',
                style: TextStyle(fontSize: 12, color: Colors.white70),
              ),
              const SizedBox(height: 4),
              LocationCard(
                location: location,
                onBuild: null,
                onRemove: null,
                onAbandon: null,
              ),
            ],
          ),
        const SizedBox(height: 8),
        const Text(
          'Возвращенный инвентарь:',
          style: TextStyle(fontSize: 12, color: Colors.white54),
        ),
        const SizedBox(height: 4),
        Wrap(
          spacing: 4,
          runSpacing: 4,
          children: chain.inventory.entries.map((entry) {
            if (entry.value <= 0) return const SizedBox.shrink();
            return Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: AppTheme.successColor.withValues(alpha: 0.2),
                borderRadius: BorderRadius.circular(6),
              ),
              child: Text(
                '${Constants.resourceIcons[entry.key]} ${entry.value.toInt()}',
                style: const TextStyle(fontSize: 10, color: AppTheme.successColor),
              ),
            );
          }).toList(),
        ),
        const SizedBox(height: 12),
        Center(
          child: ElevatedButton(
            onPressed: () {
              Navigator.of(context).pop();
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: AppTheme.accentColor,
              foregroundColor: Colors.white,
            ),
            child: const Text('Закрыть'),
          ),
        ),
      ],
    );
  }

  Widget _buildChoiceButton(
    ExpeditionChoice choice,
    int choiceIndex,
    ExpeditionChain chain,
    ExpeditionChainProvider chainProvider,
  ) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: InkWell(
        onTap: () async {
          setState(() {});

          try {
            final result = await chainProvider.makeChoice(
              widget.planetId,
              chain.id,
              choiceIndex,
            );

            if (result.error != null) {
              if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Ошибка: ${result.error}'),
                    backgroundColor: AppTheme.dangerColor,
                  ),
                );
              }
            }
          } catch (e) {
            if (mounted) {
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text('Ошибка выбора: $e'),
                  backgroundColor: AppTheme.dangerColor,
                ),
              );
            }
          }
        },
        child: Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: AppTheme.accentColor.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: AppTheme.accentColor.withValues(alpha: 0.3)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                choice.label,
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: Colors.white,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                choice.description,
                style: const TextStyle(fontSize: 11, color: Colors.white54),
              ),
              if (choice.reward.isNotEmpty) ...[
                const SizedBox(height: 6),
                Wrap(
                  spacing: 4,
                  runSpacing: 4,
                  children: choice.reward.entries.map((entry) {
                    final value = entry.value;
                    final sign = value >= 0 ? '+' : '';
                    return Container(
                      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                      decoration: BoxDecoration(
                        color: value >= 0
                            ? AppTheme.successColor.withValues(alpha: 0.2)
                            : AppTheme.dangerColor.withValues(alpha: 0.2),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Text(
                        '${Constants.resourceIcons[entry.key]} $sign${value.toInt()}',
                        style: TextStyle(
                          fontSize: 9,
                          color: value >= 0 ? AppTheme.successColor : AppTheme.dangerColor,
                        ),
                      ),
                    );
                  }).toList(),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEventHistory(ExpeditionChain? chain, ExpeditionChainProvider chainProvider) {
    if (chain == null || chain.events.isEmpty) {
      return const Card(
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Center(
            child: Text('История событий пуста', style: TextStyle(color: Colors.white38)),
          ),
        ),
      );
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'История событий',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70),
            ),
            const SizedBox(height: 12),
            ...chain.events.map((event) => Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: AppTheme.cardColor.withValues(alpha: 0.5),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          event.description,
                          style: const TextStyle(fontSize: 12, color: Colors.white70),
                        ),
                        const SizedBox(height: 4),
                        if (event.immediateReward.isNotEmpty)
                          Wrap(
                            spacing: 4,
                            runSpacing: 4,
                            children: event.immediateReward.entries.map((entry) {
                              final value = entry.value;
                              final sign = value >= 0 ? '+' : '';
                              return Container(
                                padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                                decoration: BoxDecoration(
                                  color: value >= 0
                                      ? AppTheme.successColor.withValues(alpha: 0.2)
                                      : AppTheme.dangerColor.withValues(alpha: 0.2),
                                  borderRadius: BorderRadius.circular(4),
                                ),
                                child: Text(
                                  '${Constants.resourceIcons[entry.key]} $sign${value.toInt()}',
                                  style: TextStyle(
                                    fontSize: 9,
                                    color: value >= 0 ? AppTheme.successColor : AppTheme.dangerColor,
                                  ),
                                ),
                              );
                            }).toList(),
                          ),
                      ],
                    ),
                  ),
                )),
          ],
        ),
      ),
    );
  }

  Widget _buildStartButton(ExpeditionChainProvider chainProvider) {
    final hasActiveOrGenerating = chainProvider.activeChains.isNotEmpty;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            ElevatedButton.icon(
              onPressed: hasActiveOrGenerating ? null : () async {
                final result = await showDialog<bool>(
                  context: context,
                  builder: (dialogContext) => InventoryDialog(
                    planetId: widget.planetId,
                    onSubmitted: _onExpeditionStarted,
                  ),
                );
                if (result == true) {
                  _onExpeditionStarted();
                }
              },
              icon: const Icon(Icons.explore, size: 18),
              label: const Text('Новая экспедиция'),
              style: ElevatedButton.styleFrom(
                backgroundColor: AppTheme.accentColor,
                foregroundColor: Colors.white,
                minimumSize: const Size(double.infinity, 44),
              ),
            ),
            if (hasActiveOrGenerating)
              const Padding(
                padding: EdgeInsets.only(top: 8),
                child: Text(
                  'Сначала завершите текущую экспедицию',
                  style: TextStyle(fontSize: 11, color: Colors.white38),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

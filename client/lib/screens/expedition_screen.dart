import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';

class ExpeditionScreen extends StatelessWidget {
  const ExpeditionScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Экспедиции')),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          final planet = gameProvider.selectedPlanet;
          if (planet == null) return const Center(child: Text('Планета не выбрана'));

          return RefreshIndicator(
            onRefresh: () async => gameProvider.loadExpeditions(planet.id),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildStartExpedition(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildExpeditionList(context, gameProvider),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildStartExpedition(BuildContext context, GameProvider gameProvider) {
    final expeditions = gameProvider.expeditions;
    final canStart = expeditions?.canStartNew ?? false;
    final unlocked = expeditions?.expeditionsUnlocked ?? false;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('Начать экспедицию', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                Text(
                  '${expeditions?.activeCount ?? 0}/${expeditions?.maxExpeditions ?? 1}',
                  style: const TextStyle(color: Colors.white54),
                ),
              ],
            ),
            if (!unlocked)
              const Padding(
                padding: EdgeInsets.only(top: 8, bottom: 4),
                child: Text('Сначала исследуйте "Экспедиции"', style: TextStyle(fontSize: 12, color: Colors.white38)),
              ),
            const SizedBox(height: 12),
            ...Constants.expeditionTypes.entries.map((entry) {
              final type = entry.key;
              final info = entry.value as Map<String, dynamic>;
              return Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: ElevatedButton.icon(
                  onPressed: canStart && unlocked
                      ? () => gameProvider.startExpedition(expeditionType: type)
                      : null,
                  icon: Text(info['icon'] as String),
                  label: Text(info['name'] as String),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppTheme.primaryColor,
                    foregroundColor: Colors.white,
                    alignment: Alignment.centerLeft,
                  ),
                ),
              );
            }),
          ],
        ),
      ),
    );
  }

  Widget _buildExpeditionList(BuildContext context, GameProvider gameProvider) {
    final expeditions = gameProvider.expeditions;
    if (expeditions == null) return const SizedBox.shrink();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Текущие экспедиции', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            if (expeditions.expeditions.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: Text('Нет активных экспедиций', style: TextStyle(color: Colors.white38))),
              )
            else
              ...expeditions.expeditions.map((exp) => Padding(
                    padding: const EdgeInsets.only(bottom: 12),
                    child: _ExpeditionCard(expedition: exp),
                  )),
          ],
        ),
      ),
    );
  }
}

class _ExpeditionCard extends StatelessWidget {
  final dynamic expedition;

  const _ExpeditionCard({required this.expedition});

  @override
  Widget build(BuildContext context) {
    final icon = Constants.expeditionTypes[expedition.expeditionType]?['icon'] ?? '🗺️';
    final progress = expedition.progress;
    final remaining = expedition.remainingProgress;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Text(icon, style: const TextStyle(fontSize: 24)),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        expedition.target,
                        style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.white),
                      ),
                      Text(
                        '${expedition.expeditionType} | ${expedition.fleetTotal} ships',
                        style: const TextStyle(fontSize: 11, color: Colors.white54),
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: expedition.isActive ? AppTheme.successColor.withOpacity(0.2) : AppTheme.warningColor.withOpacity(0.2),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    expedition.status,
                    style: TextStyle(
                      fontSize: 11,
                      color: expedition.isActive ? AppTheme.successColor : AppTheme.warningColor,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Прогресс: ${(progress * 100).toStringAsFixed(0)}%', style: const TextStyle(fontSize: 11, color: Colors.white54)),
                Text('Время: ${Constants.formatTime(remaining)}', style: const TextStyle(fontSize: 11, color: Colors.white54)),
              ],
            ),
            const SizedBox(height: 4),
            LinearProgressIndicator(
              value: progress.clamp(0.0, 1.0),
              minHeight: 6,
              borderRadius: BorderRadius.circular(3),
              color: AppTheme.accentColor,
            ),
            if (expedition.discoveredNPC != null) ...[
              const SizedBox(height: 8),
              const Divider(),
              _buildNPCInfo(context, expedition.discoveredNPC!),
              if (expedition.canAct) _buildActions(context, expedition),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildNPCInfo(BuildContext context, dynamic npc) {
    return Padding(
      padding: const EdgeInsets.only(top: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Обнаружено: ${npc.name} (${npc.type})', style: const TextStyle(fontSize: 12, color: AppTheme.accentColor)),
          if (npc.hasCombat)
            Text('Сила боевого флота: ${npc.fleetStrength.toInt()}', style: const TextStyle(fontSize: 11, color: Colors.white54)),
          if (npc.resources.isNotEmpty)
            Text('Ресурсы: ${npc.resources.keys.join(", ")}', style: const TextStyle(fontSize: 11, color: Colors.white54)),
        ],
      ),
    );
  }

  Widget _buildActions(BuildContext context, dynamic expedition) {
    return Column(
      children: expedition.actions.map((action) {
        return Padding(
          padding: const EdgeInsets.only(top: 4),
          child: SizedBox(
            width: double.infinity,
            child: OutlinedButton(
              onPressed: () async {
                await context.read<GameProvider>().expeditionAction(expedition.id, action.type);
              },
              child: Text(action.label),
            ),
          ),
        );
      }).toList(),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../providers/planet_survey_provider.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/planet_survey.dart';
import '../widgets/location_card.dart';

class PlanetSurveyScreen extends StatefulWidget {
  final String planetId;

  const PlanetSurveyScreen({super.key, required this.planetId});

  @override
  State<PlanetSurveyScreen> createState() => _PlanetSurveyScreenState();
}

class _PlanetSurveyScreenState extends State<PlanetSurveyScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  PlanetSurveyProvider? _provider;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _provider = PlanetSurveyProvider(baseUrl: context.read<GameProvider>().baseUrl, authToken: context.read<GameProvider>().authToken);
    _loadData();
  }

  @override
  void dispose() {
    _tabController.dispose();
    _provider?.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    if (_provider == null) return;
    await _provider!.loadPlanetSurvey(widget.planetId);
    await _provider!.loadLocations(widget.planetId);
    await _provider!.loadExpeditionHistory(widget.planetId);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Экспедиции'),
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Разведка планеты'),
            Tab(text: 'Космические экспедиции'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildPlanetSurveyTab(),
          const SpaceExpeditionTab(),
        ],
      ),
    );
  }

  Widget _buildPlanetSurveyTab() {
    return Consumer<GameProvider>(
      builder: (context, gameProvider, _) {
        final planet = gameProvider.selectedPlanet;
        if (planet == null) return const Center(child: Text('Планета не выбрана'));

        return RefreshIndicator(
          onRefresh: _loadData,
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildStartExpedition(),
                const SizedBox(height: 16),
                _buildExpeditionList(),
                const SizedBox(height: 16),
                _buildLocationsList(),
                const SizedBox(height: 16),
                _buildHistory(),
                const SizedBox(height: 16),
                _buildRangeStats(),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildStartExpedition() {
    return Consumer<PlanetSurveyProvider>(
      builder: (context, provider, _) {
        final baseLevel = provider.baseLevel ?? 0;
        final canStart = provider.canStartPlanetSurvey;
        final maxDuration = provider.getMaxDurationForBaseLevel(baseLevel);
        final costPerMin = provider.getCostPerMinForBaseLevel(baseLevel);
        final durations = [300, 600, 1200];
        final availableDurations = durations.where((d) => d <= maxDuration).toList();

        if (baseLevel <= 0) {
          return Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Разведка планеты', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                  const SizedBox(height: 8),
                  const Text('Постройте и запустите базу, чтобы начать разведку.', style: TextStyle(fontSize: 12, color: Colors.white38)),
                ],
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
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Запустить экспедицию', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                    Text(
                      'Base Lv.$baseLevel',
                      style: const TextStyle(fontSize: 12, color: Colors.white54),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: availableDurations.map((duration) {
                    final minutes = duration ~/ 60;
                    final canAfford = canStart;
                    return ElevatedButton.icon(
                      onPressed: canAfford && canStart
                          ? () async {
                              await provider.startPlanetSurvey(widget.planetId, duration);
                            }
                          : null,
                      icon: const Icon(Icons.explore, size: 18),
                      label: Text('${minutes} мин'),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppTheme.primaryColor,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                      ),
                    );
                  }).toList(),
                ),
                if (availableDurations.isNotEmpty) ...[
                  const SizedBox(height: 8),
                  Text(
                    'Стоимость: ${costPerMin['food']!.toInt()} еды, ${costPerMin['iron']!.toInt()} железа, ${costPerMin['money']!.toInt()} денег за минуту',
                    style: const TextStyle(fontSize: 11, color: Colors.white54),
                  ),
                ],
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildExpeditionList() {
    return Consumer<PlanetSurveyProvider>(
      builder: (context, provider, _) {
        final activeExpeditions = provider.expeditions.where((e) => e.isActive).toList();

        return Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Активные экспедиции', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                if (activeExpeditions.isEmpty)
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Center(child: Text('Нет активных экспедиций', style: TextStyle(color: Colors.white38))),
                  )
                else
                  ...activeExpeditions.map((exp) => Padding(
                        padding: const EdgeInsets.only(bottom: 12),
                        child: _buildExpeditionCard(exp),
                      )),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildExpeditionCard(SurfaceExpedition exp) {
    final remaining = exp.remainingTime.clamp(0, 999);
    final rangeLabel = exp.range == '300s' ? '5 мин' : exp.range == '600s' ? '10 мин' : '20 мин';

    return Card(
      elevation: 0,
      color: AppTheme.cardColor.withValues(alpha: 0.5),
      child: Padding(
        padding: const EdgeInsets.all(10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Text('🔍', style: TextStyle(fontSize: 18)),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Дальность: $rangeLabel',
                        style: const TextStyle(fontWeight: FontWeight.w600, color: Colors.white, fontSize: 13),
                      ),
                      Text(
                        'Прогресс: ${(exp.progress * 100).toStringAsFixed(0)}%',
                        style: const TextStyle(fontSize: 11, color: Colors.white54),
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: AppTheme.successColor.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    'Активна',
                    style: const TextStyle(fontSize: 11, color: AppTheme.successColor),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 6),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Осталось: ${Constants.formatTime(remaining.toDouble())}', style: const TextStyle(fontSize: 11, color: Colors.white54)),
              ],
            ),
            const SizedBox(height: 4),
            LinearProgressIndicator(
              value: exp.progress.clamp(0.0, 1.0),
              minHeight: 6,
              borderRadius: BorderRadius.circular(3),
              color: AppTheme.accentColor,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLocationsList() {
    return Consumer<PlanetSurveyProvider>(
      builder: (context, provider, _) {
        final baseLevel = provider.baseLevel ?? 0;
        final hasLocationBuildings = baseLevel > 0;

        return Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Локации', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                    Text(
                      '${provider.locations.length}/${provider.maxLocations ?? 1}',
                      style: const TextStyle(fontSize: 12, color: Colors.white54),
                    ),
                  ],
                ),
                if (provider.locations.isEmpty)
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Center(child: Text('Локаций пока нет. Запустите экспедицию!', style: TextStyle(color: Colors.white38))),
                  )
                else
                  ...provider.locations.map((loc) => Padding(
                        padding: const EdgeInsets.only(bottom: 8),
                        child: LocationCard(
                          location: loc,
                          provider: provider,
                          onBuild: hasLocationBuildings ? () => _showBuildDialog(context, loc) : null,
                          onRemove: () => provider.removeBuilding(widget.planetId, loc.id),
                          onAbandon: () => _confirmAbandon(context, loc),
                        ),
                      )),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildHistory() {
    return Consumer<PlanetSurveyProvider>(
      builder: (context, provider, _) {
        return Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('История экспедиций', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                if (provider.history.isEmpty)
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Center(child: Text('История пуста', style: TextStyle(color: Colors.white38))),
                  )
                else
                  ...provider.history.take(10).map((entry) => Padding(
                        padding: const EdgeInsets.only(bottom: 8),
                        child: _buildHistoryEntry(entry),
                      )),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildHistoryEntry(ExpeditionHistoryEntry entry) {
    final resultColor = entry.result == 'success' ? AppTheme.successColor : AppTheme.dangerColor;
    final resultLabel = entry.result == 'success' ? 'Успех' : entry.result == 'abandoned' ? 'Отозвана' : 'Провал';

    return Card(
      elevation: 0,
      color: AppTheme.cardColor.withValues(alpha: 0.5),
      child: Padding(
        padding: const EdgeInsets.all(10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  entry.discovered.isNotEmpty ? entry.discovered : 'Неизвестно',
                  style: const TextStyle(fontWeight: FontWeight.w600, color: Colors.white, fontSize: 13),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: resultColor.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Text(
                    resultLabel,
                    style: TextStyle(fontSize: 10, color: resultColor),
                  ),
                ),
              ],
            ),
            if (entry.resourcesGained.isNotEmpty) ...[
              const SizedBox(height: 4),
              Wrap(
                spacing: 4,
                runSpacing: 4,
                children: entry.resourcesGained.entries.map((e) {
                  if (e.value <= 0) return const SizedBox.shrink();
                  return Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: AppTheme.successColor.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      '${Constants.resourceNames[e.key] ?? e.key}: +${e.value.toInt()}',
                      style: const TextStyle(fontSize: 10, color: AppTheme.successColor),
                    ),
                  );
                }).toList(),
              ),
            ],
            Text(
              'Завершена: ${Constants.formatDateTime(entry.completedAt)}',
              style: const TextStyle(fontSize: 10, color: Colors.white38),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRangeStats() {
    return Consumer<PlanetSurveyProvider>(
      builder: (context, provider, _) {
        if (provider.rangeStats.isEmpty) return const SizedBox.shrink();

        return Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Статистика по дальности', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                const SizedBox(height: 8),
                ...provider.rangeStats.entries.map((e) {
                  final rangeLabel = e.key == '300s' ? '5 мин' : e.key == '600s' ? '10 мин' : '20 мин';
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(rangeLabel, style: const TextStyle(fontSize: 12, color: Colors.white70)),
                        Text(
                          '${e.value.totalExpeditions} экспедиций, ${e.value.locationsFound} локаций',
                          style: const TextStyle(fontSize: 12, color: Colors.white54),
                        ),
                      ],
                    ),
                  );
                }),
              ],
            ),
          ),
        );
      },
    );
  }

  void _showBuildDialog(BuildContext context, Location location) {
    final buildings = _provider!.getAvailableBuildingsForLocation(location.type);
    if (buildings.isEmpty) return;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: AppTheme.cardColor,
        title: const Text('Выберите здание', style: TextStyle(color: Colors.white)),
        content: SizedBox(
          width: double.maxFinite,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: buildings.map((b) {
              return ListTile(
                title: Text(b.name, style: const TextStyle(color: Colors.white)),
                subtitle: Text('${b.costFood.toInt()} еды, ${b.costIron.toInt()} железа, ${b.costMoney.toInt()} денег', style: const TextStyle(color: Colors.white54)),
                trailing: Text('${Constants.formatTime(b.buildTime)}', style: const TextStyle(color: AppTheme.accentColor)),
                onTap: () {
                  Navigator.pop(context);
                  _provider!.buildOnLocation(widget.planetId, location.id, b.type);
                },
              );
            }).toList(),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена', style: TextStyle(color: Colors.white54)),
          ),
        ],
      ),
    );
  }

  void _confirmAbandon(BuildContext context, Location location) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: AppTheme.cardColor,
        title: const Text('Забрать локацию', style: TextStyle(color: Colors.white)),
        content: Text('Вы уверены, что хотите забрать "${location.name}"? Здание и локация будут удалены.', style: const TextStyle(color: Colors.white70)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена', style: TextStyle(color: Colors.white54)),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              _provider!.abandonLocation(widget.planetId, location.id);
            },
            child: const Text('Забрать', style: TextStyle(color: AppTheme.dangerColor)),
          ),
        ],
      ),
    );
  }
}

class SpaceExpeditionTab extends StatelessWidget {
  const SpaceExpeditionTab({super.key});

  @override
  Widget build(BuildContext context) {
    final gameProvider = context.read<GameProvider>();
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
                    child: _SpaceExpeditionCard(expedition: exp),
                  )),
          ],
        ),
      ),
    );
  }
}

class _SpaceExpeditionCard extends StatelessWidget {
  final dynamic expedition;

  const _SpaceExpeditionCard({required this.expedition});

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

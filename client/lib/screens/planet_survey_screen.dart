import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/planet_survey.dart';
import '../widgets/location_card.dart';

class LocationBuildingDef {
  final String type;
  final String name;
  final double costFood;
  final double costIron;
  final double costMoney;
  final double buildTime;

  const LocationBuildingDef({
    required this.type,
    required this.name,
    required this.costFood,
    required this.costIron,
    required this.costMoney,
    required this.buildTime,
  });
}

class PlanetSurveyScreen extends StatefulWidget {
  final String planetId;

  const PlanetSurveyScreen({super.key, required this.planetId});

  @override
  State<PlanetSurveyScreen> createState() => _PlanetSurveyScreenState();
}

class _PlanetSurveyScreenState extends State<PlanetSurveyScreen> {

   Future<void> _loadData(String planetId) async {
     await context.read<GameProvider>().loadPlanetSurveyData(planetId);
   }

   void _checkNotifications() {
     final gameProvider = context.read<GameProvider>();
     if (gameProvider.notifications.isNotEmpty) {
       final notif = gameProvider.notifications.first;
       ScaffoldMessenger.of(context).showSnackBar(
         SnackBar(
           content: Text('${notif['title']}\n${notif['message']}'),
           duration: const Duration(seconds: 5),
           backgroundColor: AppTheme.accentColor,
         ),
       );
       gameProvider.clearNotifications();
     }
   }

   @override
   void didChangeDependencies() {
     super.didChangeDependencies();
     WidgetsBinding.instance.addPostFrameCallback((_) => _checkNotifications());
   }

   @override
   Widget build(BuildContext context) {
     return Consumer<GameProvider>(
       builder: (context, gameProvider, _) {
         final planet = gameProvider.selectedPlanet;
         if (planet == null) return const Center(child: Text('Планета не выбрана'));

         return Scaffold(
           appBar: AppBar(title: const Text('Разведка')),
          body: RefreshIndicator(
            onRefresh: () => _loadData(planet.id),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildStartExpedition(gameProvider),
                  const SizedBox(height: 16),
                  _buildExpeditionList(gameProvider),
                  const SizedBox(height: 16),
                  _buildLocationsList(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildHistory(gameProvider),
                  const SizedBox(height: 16),
                  _buildRangeStats(gameProvider),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildStartExpedition(GameProvider gameProvider) {
    final baseLevel = gameProvider.baseLevel ?? 0;
    final canStart = gameProvider.canStartPlanetSurvey ?? false;
    final maxDuration = _getMaxDurationForBaseLevel(baseLevel);
    final costPerMin = _getCostPerMinForBaseLevel(baseLevel);
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
                return ElevatedButton.icon(
                  onPressed: canStart
                      ? () async {
                          await gameProvider.startPlanetSurvey(widget.planetId, duration);
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
  }

  Widget _buildExpeditionList(GameProvider gameProvider) {
    final activeExpeditions = gameProvider.surfaceExpeditions.where((e) => e['status'] == 'active').toList();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('Активные экспедиции', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                Text(
                  '${activeExpeditions.length}/${gameProvider.maxSurfaceExpeditions ?? 1}',
                  style: const TextStyle(fontSize: 12, color: Colors.white54),
                ),
              ],
            ),
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
  }

  Widget _buildExpeditionCard(Map<String, dynamic> exp) {
    final remaining = (exp['duration'] as num? ?? 0).toInt() - (exp['elapsed_time'] as num? ?? 0).toInt();
    final rangeLabel = exp['range'] == '300s' ? '5 мин' : exp['range'] == '600s' ? '10 мин' : '20 мин';

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
                        'Прогресс: ${(exp['progress'] as num? ?? 0).toDouble() * 100}%${exp['status'] == 'discovered' ? ' | Обнаружено!' : ''}',
                        style: const TextStyle(fontSize: 11, color: Colors.white54),
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: exp['status'] == 'active' ? AppTheme.successColor.withValues(alpha: 0.2) : AppTheme.accentColor.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    exp['status'] == 'active' ? 'Активна' : 'Обнаружено',
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
              value: (exp['progress'] as num? ?? 0).toDouble().clamp(0.0, 1.0),
              minHeight: 6,
              borderRadius: BorderRadius.circular(3),
              color: AppTheme.accentColor,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLocationsList(BuildContext context, GameProvider gameProvider) {
    final baseLevel = gameProvider.baseLevel ?? 0;
    final hasLocationBuildings = baseLevel > 0;
    final locations = gameProvider.locations;

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
                  '${locations.length}/5',
                  style: const TextStyle(fontSize: 12, color: Colors.white54),
                ),
              ],
            ),
            if (locations.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: Text('Локаций пока нет. Запустите экспедицию!', style: TextStyle(color: Colors.white38))),
              )
            else
              ...locations.map((loc) => Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: LocationCard(
                      location: _mapToLocation(loc),
                      onBuild: hasLocationBuildings ? () => _showBuildDialog(context, loc) : null,
                      onRemove: () => gameProvider.removeBuilding(widget.planetId, loc['id'] as String),
                      onAbandon: () => _confirmAbandon(context, loc),
                    ),
                  )),
          ],
        ),
      ),
    );
  }

  Location _mapToLocation(Map<String, dynamic> loc) {
    return Location(
      id: loc['id'] as String? ?? '',
      type: loc['type'] as String? ?? 'generic',
      name: loc['name'] as String? ?? 'Неизвестно',
      buildingType: loc['building_type'] as String? ?? '',
      buildingLevel: loc['building_level'] as int? ?? 0,
      buildingActive: loc['building_active'] as bool? ?? false,
      sourceResource: loc['source_resource'] as String? ?? '',
      sourceAmount: (loc['source_amount'] as num?)?.toDouble() ?? 0,
      sourceRemaining: (loc['source_remaining'] as num?)?.toDouble() ?? 0,
      active: loc['active'] as bool? ?? true,
      discoveredAt: loc['discovered_at'] != null ? DateTime.parse(loc['discovered_at'] as String) : DateTime.now(),
    );
  }

  Widget _buildHistory(GameProvider gameProvider) {
    final history = gameProvider.expeditionHistory;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('История экспедиций', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            if (history.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: Text('История пуста', style: TextStyle(color: Colors.white38))),
              )
            else
              ...history.take(10).map((entry) => Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: _buildHistoryEntry(entry),
                  )),
          ],
        ),
      ),
    );
  }

  Widget _buildHistoryEntry(Map<String, dynamic> entry) {
    final result = entry['result'] as String? ?? 'unknown';
    final resultColor = result == 'success' ? AppTheme.successColor : AppTheme.dangerColor;
    final resultLabel = result == 'success' ? 'Успех' : result == 'abandoned' ? 'Отозвана' : 'Провал';

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
                  entry['discovered'] as String? ?? 'Неизвестно',
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
            if ((entry['resources_gained'] as Map?)?.isNotEmpty ?? false) ...[
              const SizedBox(height: 4),
              Wrap(
                spacing: 4,
                runSpacing: 4,
                children: (entry['resources_gained'] as Map).entries.map((e) {
                  final value = (e.value as num?)?.toInt() ?? 0;
                  if (value <= 0) return const SizedBox.shrink();
                  return Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: AppTheme.successColor.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      '${Constants.resourceNames[e.key as String] ?? e.key}: +$value',
                      style: const TextStyle(fontSize: 10, color: AppTheme.successColor),
                    ),
                  );
                }).toList(),
              ),
            ],
            Text(
              'Завершена: ${entry['completed_at'] ?? 'Неизвестно'}',
              style: const TextStyle(fontSize: 10, color: Colors.white38),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRangeStats(GameProvider gameProvider) {
    final rangeStats = gameProvider.rangeStats;
    if (rangeStats.isEmpty) return const SizedBox.shrink();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Статистика по дальности', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 8),
            ...rangeStats.entries.map((e) {
              final rangeLabel = e.key == '300s' ? '5 мин' : e.key == '600s' ? '10 мин' : '20 мин';
              final stats = e.value as Map<String, dynamic>? ?? {};
              return Padding(
                padding: const EdgeInsets.only(bottom: 6),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(rangeLabel, style: const TextStyle(fontSize: 12, color: Colors.white70)),
                    Text(
                      '${stats['total_expeditions'] ?? 0} экспедиций, ${stats['locations_found'] ?? 0} локаций',
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
  }

  void _showBuildDialog(BuildContext context, Map<String, dynamic> location) {
    final buildings = _getAvailableBuildingsForLocation(location['type'] as String);
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
            children: buildings.map((LocationBuildingDef b) {
              return ListTile(
                title: Text(b.name ?? '', style: const TextStyle(color: Colors.white)),
                subtitle: Text('${(b.costFood ?? 0).toInt()} еды, ${(b.costIron ?? 0).toInt()} железа, ${(b.costMoney ?? 0).toInt()} денег', style: const TextStyle(color: Colors.white54)),
                trailing: Text('${Constants.formatTime(b.buildTime ?? 0)}', style: const TextStyle(color: AppTheme.accentColor)),
                onTap: () {
                  Navigator.pop(context);
                  context.read<GameProvider>().buildOnLocation(widget.planetId, location['id'] as String, b.type ?? '');
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

  void _confirmAbandon(BuildContext context, Map<String, dynamic> location) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: AppTheme.cardColor,
        title: const Text('Забрать локацию', style: TextStyle(color: Colors.white)),
        content: Text('Вы уверены, что хотите забрать "${location['name'] ?? location['type'] ?? 'локацию'}"? Здание и локация будут удалены.', style: const TextStyle(color: Colors.white70)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена', style: TextStyle(color: Colors.white54)),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              context.read<GameProvider>().abandonLocation(widget.planetId, location['id'] as String? ?? '');
            },
            child: const Text('Забрать', style: TextStyle(color: AppTheme.dangerColor)),
          ),
        ],
      ),
    );
  }

  List<LocationBuildingDef> _getAvailableBuildingsForLocation(String locationType) {
    final buildings = <LocationBuildingDef>[];
    final rarity = _getRarityForLocationType(locationType);

    double getCostMultiplier() {
      switch (rarity) {
        case 'common': return 1.0;
        case 'uncommon': return 2.0;
        case 'rare': return 4.0;
        case 'exotic': return 6.0;
        default: return 1.0;
      }
    }

    final costMult = getCostMultiplier();

    switch (locationType) {
      case 'pond':
      case 'river':
        buildings.addAll([
          LocationBuildingDef(type: 'fish_farm', name: 'Рыбная ферма', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
          LocationBuildingDef(type: 'water_purifier', name: 'Очиститель воды', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        ]);
        break;
      case 'forest':
        buildings.addAll([
          LocationBuildingDef(type: 'lumber_mill', name: 'Лесопилка', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
          LocationBuildingDef(type: 'herb_garden', name: 'Травяной сад', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        ]);
        break;
      case 'mineral_deposit':
        buildings.addAll([
          LocationBuildingDef(type: 'mineral_extractor', name: 'Экстрактор минералов', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
          LocationBuildingDef(type: 'smelter', name: 'Плавильня', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        ]);
        break;
      case 'dry_valley':
        buildings.addAll([
          LocationBuildingDef(type: 'solar_farm', name: 'Солнечная ферма', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
          LocationBuildingDef(type: 'wind_turbine', name: 'Ветровая турбина', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        ]);
        break;
      case 'waterfall':
      case 'cave':
      case 'thermal_spring':
      case 'salt_lake':
      case 'wind_pass':
        buildings.addAll([
          LocationBuildingDef(type: 'hydro_plant', name: 'Гидроэлектростанция', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
          LocationBuildingDef(type: 'turbine_station', name: 'Турбинная станция', costFood: 200 * costMult, costIron: 100 * costMult, costMoney: 400 * costMult, buildTime: 900),
        ]);
        break;
      case 'crystal_cave':
      case 'meteor_crater':
      case 'sunken_city':
      case 'glacier':
      case 'mushroom_forest':
        buildings.addAll([
          LocationBuildingDef(type: 'crystal_harvester', name: 'Сборщик кристаллов', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
          LocationBuildingDef(type: 'alloy_forge', name: 'Кузница сплавов', costFood: 400 * costMult, costIron: 200 * costMult, costMoney: 800 * costMult, buildTime: 1200),
        ]);
        break;
      case 'crystal_field':
      case 'cloud_island':
      case 'underground_lake':
      case 'radioactive_zone':
      case 'anomaly_zone':
        buildings.addAll([
          LocationBuildingDef(type: 'crystal_array', name: 'Кристальный массив', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
          LocationBuildingDef(type: 'resonance_amplifier', name: 'Резонансный усилитель', costFood: 600 * costMult, costIron: 300 * costMult, costMoney: 1200 * costMult, buildTime: 1800),
        ]);
        break;
      default:
        buildings.add(
          LocationBuildingDef(type: 'generic_extractor', name: 'Экстрактор', costFood: 100 * costMult, costIron: 50 * costMult, costMoney: 200 * costMult, buildTime: 600),
        );
    }

    return buildings;
  }

  String _getRarityForLocationType(String locationType) {
    final common = ['pond', 'river', 'forest', 'mineral_deposit', 'dry_valley'];
    final uncommon = ['waterfall', 'cave', 'thermal_spring', 'salt_lake', 'wind_pass'];
    final rare = ['crystal_cave', 'meteor_crater', 'sunken_city', 'glacier', 'mushroom_forest'];
    final exotic = ['crystal_field', 'cloud_island', 'underground_lake', 'radioactive_zone', 'anomaly_zone'];

    if (common.contains(locationType)) return 'common';
    if (uncommon.contains(locationType)) return 'uncommon';
    if (rare.contains(locationType)) return 'rare';
    if (exotic.contains(locationType)) return 'exotic';
    return 'common';
  }

  int _getMaxDurationForBaseLevel(int baseLevel) {
    switch (baseLevel) {
      case 1: return 300;
      case 2: return 600;
      case 3: return 1200;
      default: return 300;
    }
  }

  Map<String, double> _getCostPerMinForBaseLevel(int baseLevel) {
    switch (baseLevel) {
      case 1: return {'food': 100, 'iron': 100, 'money': 10};
      case 2: return {'food': 200, 'iron': 200, 'money': 20};
      case 3: return {'food': 400, 'iron': 400, 'money': 40};
      default: return {'food': 100, 'iron': 100, 'money': 10};
    }
  }
}

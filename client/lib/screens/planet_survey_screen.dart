import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../providers/planet_survey_provider.dart';
import '../providers/expedition_chain_provider.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/planet_survey.dart';
import '../widgets/location_card.dart';
import '../widgets/inventory_dialog.dart';
import '../models/expedition_chain.dart';

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
      final gameProvider = context.read<GameProvider>();
      await gameProvider.expeditionChainProvider.loadExpeditionChains(planetId);
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
          final survey = gameProvider.surveyProvider;
          final chainProvider = gameProvider.expeditionChainProvider;
          final activeChains = chainProvider.activeChains;

          return Scaffold(
            appBar: AppBar(title: const Text('Разведка')),
           body: RefreshIndicator(
              onRefresh: () => _loadData(planet.id),
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildStartExpeditionWithSurvey(gameProvider, survey),
                    const SizedBox(height: 16),
                    _buildLocationsListWithSurvey(context, gameProvider, survey),
                    const SizedBox(height: 16),
                    _buildHistoryWithSurvey(chainProvider),
                  ],
                ),
              ),
            ),
          );
        },
      );
    }

  Widget _buildStartExpeditionWithSurvey(GameProvider gameProvider, PlanetSurveyProvider surveyProvider) {
    final baseLevel = surveyProvider.baseLevel ?? 0;

    if (baseLevel <= 0) {
      return const Card(
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Разведка планеты', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
              SizedBox(height: 8),
              Text('Постройте и запустите базу, чтобы начать разведку.', style: TextStyle(fontSize: 12, color: Colors.white38)),
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
            ElevatedButton.icon(
              onPressed: () async {
                final result = await showDialog<bool>(
                  context: context,
                  builder: (dialogContext) => InventoryDialog(
                    planetId: widget.planetId,
                    onSubmitted: () => _loadData(widget.planetId),
                  ),
                );
                if (result == true) {
                  await _loadData(widget.planetId);
                }
              },
              icon: const Icon(Icons.explore, size: 18),
              label: const Text('Новая экспедиция'),
              style: ElevatedButton.styleFrom(
                backgroundColor: AppTheme.primaryColor,
                foregroundColor: Colors.white,
                minimumSize: const Size(double.infinity, 40),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLocationsListWithSurvey(BuildContext context, GameProvider gameProvider, PlanetSurveyProvider surveyProvider) {
    final baseLevel = surveyProvider.baseLevel ?? 0;
    final hasLocationBuildings = baseLevel > 0;
    final locations = surveyProvider.locations;

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
                      onRemove: () => surveyProvider.removeBuilding(widget.planetId, loc['id'] as String),
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

  Widget _buildHistoryWithSurvey(ExpeditionChainProvider chainProvider) {
    final completedChains = chainProvider.completedChains;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('История экспедиций', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            if (completedChains.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: Text('История пуста', style: TextStyle(color: Colors.white38))),
              )
            else
              ...completedChains.where((c) => c.isCompleted).take(10).map((chain) => Padding(
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
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(
                                chain.discoveredLocation?.name ?? 'Неизвестно',
                                style: const TextStyle(fontWeight: FontWeight.w600, color: Colors.white, fontSize: 13),
                              ),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                decoration: BoxDecoration(
                                  color: AppTheme.successColor.withValues(alpha: 0.2),
                                  borderRadius: BorderRadius.circular(10),
                                ),
                                child: const Text(
                                  'Успех',
                                  style: TextStyle(fontSize: 10, color: AppTheme.successColor),
                                ),
                              ),
                            ],
                          ),
                          if (chain.discoveredLocation != null)
                            Text(
                              _formatLocationType(chain.discoveredLocation!.type),
                              style: const TextStyle(fontSize: 11, color: Colors.white70, fontStyle: FontStyle.italic),
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

  String _formatLocationType(String type) {
    final typeMap = <String, String>{
      'pond': 'Пруд',
      'river': 'Река',
      'forest': 'Лес',
      'mineral_deposit': 'Минеральное месторождение',
      'dry_valley': 'Сухая долина',
      'waterfall': 'Водопад',
      'cave': 'Пещера',
      'thermal_spring': 'Горячий источник',
      'salt_lake': 'Соляное озеро',
      'wind_pass': 'Ветровой перевал',
      'crystal_cave': 'Кристальная пещера',
      'meteor_crater': 'Кратер метеорита',
      'sunken_city': 'Затонувший город',
      'glacier': 'Ледник',
      'mushroom_forest': 'Грибной лес',
      'crystal_field': 'Кристальное поле',
      'cloud_island': 'Облачный остров',
      'underground_lake': 'Подземное озеро',
      'radioactive_zone': 'Радиоактивная зона',
      'anomaly_zone': 'Аномальная зона',
    };
    final name = typeMap[type] ?? type;
    return name;
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
                title: Text(b.name, style: const TextStyle(color: Colors.white)),
                subtitle: Text('${b.costFood.toInt()} еды, ${b.costIron.toInt()} железа, ${b.costMoney.toInt()} денег', style: const TextStyle(color: Colors.white54)),
                trailing: Text(Constants.formatTime(b.buildTime), style: const TextStyle(color: AppTheme.accentColor)),
                onTap: () {
                  Navigator.pop(context);
                  context.read<GameProvider>().surveyProvider.buildOnLocation(widget.planetId, location['id'] as String, b.type);
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
        title: const Text('Покинуть локацию', style: TextStyle(color: Colors.white)),
        content: Text('Вы уверены, что хотите покинуть "${location['name'] ?? location['type'] ?? 'локацию'}"? Здание и локация будут удалены.', style: const TextStyle(color: Colors.white70)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена', style: TextStyle(color: Colors.white54)),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              context.read<GameProvider>().surveyProvider.abandonLocation(widget.planetId, location['id'] as String? ?? '');
            },
            child: const Text('Покинуть', style: TextStyle(color: AppTheme.dangerColor)),
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
}

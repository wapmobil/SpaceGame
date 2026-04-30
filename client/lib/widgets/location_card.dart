import 'package:flutter/material.dart';
import '../../core/app_theme.dart';
import '../../models/planet_survey.dart';

class LocationCard extends StatelessWidget {
  final Location location;
  final VoidCallback? onBuild;
  final VoidCallback? onRemove;
  final VoidCallback? onAbandon;

  const LocationCard({
    super.key,
    required this.location,
    this.onBuild,
    this.onRemove,
    this.onAbandon,
  });

  static Color getRarityColor(String locationType) {
    final rarity = _getRarityForLocationType(locationType);
    switch (rarity) {
      case 'common': return const Color(0xFF9e9e9e);
      case 'uncommon': return const Color(0xFF4caf50);
      case 'rare': return const Color(0xFF2196f3);
      case 'exotic': return const Color(0xFF9c27b0);
      default: return Colors.white54;
    }
  }

  static String getRarityLabel(String locationType) {
    final rarity = _getRarityForLocationType(locationType);
    switch (rarity) {
      case 'common': return 'Обычная';
      case 'uncommon': return 'Необычная';
      case 'rare': return 'Редкая';
      case 'exotic': return 'Экзотическая';
      default: return 'Неизвестная';
    }
  }

  static String getBuildingName(String buildingType) {
    const names = {
      'fish_farm': 'Рыбная ферма',
      'water_purifier': 'Очиститель воды',
      'irrigation_system': 'Система орошения',
      'lumber_mill': 'Лесопилка',
      'herb_garden': 'Травяной сад',
      'mineral_extractor': 'Экстрактор минералов',
      'smelter': 'Плавильня',
      'solar_farm': 'Солнечная ферма',
      'wind_turbine': 'Ветровая турбина',
      'hydro_plant': 'Гидроэлектростанция',
      'turbine_station': 'Турбинная станция',
      'crystal_mine': 'Кристальная шахта',
      'cave_lab': 'Лаборатория пещеры',
      'geothermal_plant': 'Геотермальная станция',
      'hot_spring_lab': 'Лаборатория источников',
      'salt_pans': 'Соляные копи',
      'chemical_plant': 'Химический завод',
      'wind_farm': 'Ветряная ферма',
      'storm_collector': 'Сборщик штормов',
      'crystal_harvester': 'Сборщик кристаллов',
      'crystal_lab': 'Кристальная лаборатория',
      'meteor_science_lab': 'Метеоритная лаборатория',
      'alloy_forge': 'Кузница сплавов',
      'salvage_station': 'Станция спасения',
      'ruins_archive': 'Архив руин',
      'ice_mine': 'Ледниковая шахта',
      'cryo_lab': 'Криолаборатория',
      'mushroom_farm': 'Грибная ферма',
      'spore_extractor': 'Экстрактор спор',
      'crystal_array': 'Кристальный массив',
      'resonance_amplifier': 'Резонансный усилитель',
      'cloud_harvester': 'Сборщик облаков',
      'aerial_platform': 'Воздушная платформа',
      'aquaculture_base': 'Аквакультурная база',
      'underground_irrigation': 'Подземное орошение',
      'radiation_filter': 'Фильтр радиации',
      'isotope_plant': 'Изотопный завод',
      'anomaly_siphon': 'Сифон аномалий',
      'containment_unit': 'Удержание',
      'generic_extractor': 'Экстрактор',
    };
    return names[buildingType] ?? buildingType;
  }

  static String _getRarityForLocationType(String locationType) {
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

  @override
  Widget build(BuildContext context) {
    final rarityColor = getRarityColor(location.type);
    final rarityLabel = getRarityLabel(location.type);
    final hasBuilding = location.buildingType != null && location.buildingLevel > 0;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header with name and rarity
            Row(
              children: [
                Container(
                  width: 8,
                  height: 40,
                  decoration: BoxDecoration(
                    color: rarityColor,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Text(
                              location.name,
                              style: const TextStyle(
                                fontSize: 15,
                                fontWeight: FontWeight.w600,
                                color: Colors.white,
                              ),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                          const SizedBox(width: 8),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(
                              color: rarityColor.withValues(alpha: 0.2),
                              borderRadius: BorderRadius.circular(8),
                              border: Border.all(color: rarityColor.withValues(alpha: 0.4)),
                            ),
                            child: Text(
                              rarityLabel,
                              style: TextStyle(
                                fontSize: 10,
                                color: rarityColor,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 2),
                      Text(
                        'Тип: ${location.type}',
                        style: const TextStyle(fontSize: 11, color: Colors.white54),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),

            // Source resource progress
            Text(
              'Ресурс: ${location.sourceResource}',
              style: const TextStyle(fontSize: 11, color: Colors.white70),
            ),
            const SizedBox(height: 4),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: location.sourceAmount > 0 ? location.sourceRemaining / location.sourceAmount : 0,
                minHeight: 6,
                borderRadius: BorderRadius.circular(3),
                color: location.isDepleted ? AppTheme.dangerColor : AppTheme.accentColor,
                backgroundColor: Colors.grey[800]?.withValues(alpha: 0.5),
              ),
            ),
            const SizedBox(height: 4),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '${location.sourceRemaining.toInt()}/${location.sourceAmount.toInt()}',
                  style: const TextStyle(fontSize: 10, color: Colors.white54),
                ),
                if (location.isDepleted)
                  const Text(
                    'Исчерпан',
                    style: TextStyle(fontSize: 10, color: AppTheme.dangerColor),
                  ),
              ],
            ),
            const SizedBox(height: 8),

            // Building info
            if (hasBuilding) ...[
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: AppTheme.accentColor.withValues(alpha: 0.08),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: AppTheme.accentColor.withValues(alpha: 0.2)),
                ),
                child: Row(
                  children: [
                    const Text('🏗️', style: TextStyle(fontSize: 16)),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            LocationCard.getBuildingName(location.buildingType!),
                            style: const TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.w600),
                          ),
                          Text(
                            'Уровень ${location.buildingLevel} | ${location.buildingActive ? "Активно" : "Отключено"}',
                            style: const TextStyle(fontSize: 10, color: Colors.white54),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 8),
            ],

            // Action buttons
            Wrap(
              spacing: 6,
              runSpacing: 6,
              children: [
                if (!hasBuilding && onBuild != null)
                  ElevatedButton.icon(
                    onPressed: onBuild,
                    icon: const Icon(Icons.construction, size: 14),
                    label: const Text('Построить', style: TextStyle(fontSize: 11)),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppTheme.accentColor.withValues(alpha: 0.2),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                      minimumSize: const Size(0, 30),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                  ),
                if (hasBuilding && onRemove != null)
                  OutlinedButton.icon(
                    onPressed: onRemove,
                    icon: const Icon(Icons.delete_outline, size: 14),
                    label: const Text('Снести', style: TextStyle(fontSize: 11)),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: AppTheme.warningColor,
                      side: const BorderSide(color: AppTheme.warningColor),
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                      minimumSize: const Size(0, 30),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                  ),
                if (onAbandon != null)
                  OutlinedButton.icon(
                    onPressed: onAbandon,
                    icon: const Icon(Icons.close, size: 14),
                    label: const Text('Покинуть', style: TextStyle(fontSize: 11)),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: AppTheme.dangerColor,
                      side: const BorderSide(color: AppTheme.dangerColor),
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                      minimumSize: const Size(0, 30),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

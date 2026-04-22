import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../widgets/ship_card.dart';

class ShipyardScreen extends StatelessWidget {
  const ShipyardScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Верфь')),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          final planet = gameProvider.selectedPlanet;
          if (planet == null) return const Center(child: Text('Планета не выбрана'));

          return RefreshIndicator(
            onRefresh: () async {
              await gameProvider.loadShips(planet.id);
              await gameProvider.loadAvailableShipTypes(planet.id);
            },
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildShipyardInfo(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildFleetSection(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildBuildSection(context, gameProvider),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildShipyardInfo(BuildContext context, GameProvider gameProvider) {
    final info = gameProvider.shipyardInfo;
    if (info == null) return const SizedBox.shrink();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Статус верфи', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _InfoTile('Уровень', info.shipyardLevel.toString()),
                _InfoTile('Слоты', '${info.totalSlots}/${info.maxSlots}'),
                _InfoTile('Очередь', '${info.shipyardQueueLen}'),
              ],
            ),
            if (info.shipyardQueueLen > 0) ...[
              const SizedBox(height: 12),
              const Text('Прогресс постройки', style: TextStyle(fontSize: 12, color: Colors.white54)),
              const SizedBox(height: 4),
              LinearProgressIndicator(
                value: info.shipyardProgress,
                minHeight: 6,
                borderRadius: BorderRadius.circular(3),
                color: AppTheme.accentColor,
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildFleetSection(BuildContext context, GameProvider gameProvider) {
    final ships = gameProvider.ships;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('Флот', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                Text('${ships.length} кораблей', style: const TextStyle(color: Colors.white54)),
              ],
            ),
            if (ships.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: Text('Во флоте нет кораблей', style: TextStyle(color: Colors.white38))),
              )
            else
              ...ships.map((ship) => ShipCard(ship: ship)).toList(),
          ],
        ),
      ),
    );
  }

  Widget _buildBuildSection(BuildContext context, GameProvider gameProvider) {
    final shipTypes = gameProvider.availableShipTypes;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Построить корабли', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 12),
            if (shipTypes.isEmpty)
              const Center(child: Text('Загрузка...', style: TextStyle(color: Colors.white38)))
            else
              ...shipTypes.map((shipType) => _ShipBuildTile(
                    shipType: shipType,
                    onBuild: () => gameProvider.buildShip(shipType.typeId),
                  )),
          ],
        ),
      ),
    );
  }
}

class _InfoTile extends StatelessWidget {
  final String label;
  final String value;

  const _InfoTile(this.label, this.value);

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(label, style: const TextStyle(fontSize: 11, color: Colors.white54)),
        Text(value, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white)),
      ],
    );
  }
}

class _ShipBuildTile extends StatelessWidget {
  final dynamic shipType;
  final VoidCallback onBuild;

  const _ShipBuildTile({required this.shipType, required this.onBuild});

  @override
  Widget build(BuildContext context) {
    final icon = Constants.shipIcons[shipType.typeId] ?? '🚀';
    final canBuild = shipType.canBuild;

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Card(
        child: ListTile(
          leading: Text(icon, style: const TextStyle(fontSize: 28)),
          title: Text(shipType.name),
          subtitle: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(shipType.description, style: const TextStyle(fontSize: 11, color: Colors.white70)),
              const SizedBox(height: 4),
              Wrap(
                spacing: 8,
                runSpacing: 2,
                children: [
                  _CostChip('HP', '${shipType.hp.toInt()}'),
                  _CostChip('Броня', '${shipType.armor.toInt()}'),
                  _CostChip('Урон', '${shipType.weaponMinDmg.toInt()}-${shipType.weaponMaxDmg.toInt()}'),
                  _CostChip('⚡', '${shipType.energy.toInt()}'),
                ],
              ),
              const SizedBox(height: 4),
              Wrap(
                spacing: 4,
                runSpacing: 2,
                children: [
                  if (shipType.cost.food > 0) _ResourceCost('🍖', shipType.cost.food),
                  if (shipType.cost.composite > 0) _ResourceCost('🧬', shipType.cost.composite),
                  if (shipType.cost.mechanisms > 0) _ResourceCost('⚙️', shipType.cost.mechanisms),
                  if (shipType.cost.reagents > 0) _ResourceCost('🧪', shipType.cost.reagents),
                  if (shipType.cost.money > 0) _ResourceCost('💰', shipType.cost.money),
                ],
              ),
            ],
          ),
          trailing: ElevatedButton(
            onPressed: canBuild ? onBuild : null,
            style: ElevatedButton.styleFrom(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            ),
            child: Text('${shipType.buildTime.toInt()}s'),
          ),
        ),
      ),
    );
  }
}

class _CostChip extends StatelessWidget {
  final String label;
  final String value;

  const _CostChip(this.label, this.value);

  @override
  Widget build(BuildContext context) {
    return Chip(
      label: Text('$label: $value', style: const TextStyle(fontSize: 10)),
      visualDensity: VisualDensity.compact,
      backgroundColor: AppTheme.primaryColor.withOpacity(0.3),
      side: BorderSide.none,
    );
  }
}

class _ResourceCost extends StatelessWidget {
  final String icon;
  final double amount;

  const _ResourceCost(this.icon, this.amount);

  @override
  Widget build(BuildContext context) {
    return Text(
      '$icon ${amount.toInt()}',
      style: const TextStyle(fontSize: 10, color: Colors.white54),
    );
  }
}

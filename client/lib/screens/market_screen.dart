import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/app_theme.dart';
import '../utils/constants.dart';
import '../models/planet.dart';

class MarketScreen extends StatelessWidget {
  const MarketScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Рынок')),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          final planet = gameProvider.selectedPlanet;
          if (planet == null)     return const Center(child: Text('Планета не выбрана'));

          return RefreshIndicator(
            onRefresh: () async {
              await gameProvider.loadMarketData(planet.id);
              await gameProvider.loadMyOrders(planet.id);
            },
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildMarketOverview(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildQuickSell(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildCreateOrder(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildBuyOrders(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildSellOrders(context, gameProvider),
                  const SizedBox(height: 16),
                  _buildMyOrders(context, gameProvider),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildQuickSell(BuildContext context, GameProvider gameProvider) {
    final planet = gameProvider.selectedPlanet;
    if (planet == null) return const SizedBox.shrink();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Быстрая продажа еды', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 4),
            Text('Курс: 10 еды = 1 деньги', style: const TextStyle(fontSize: 12, color: Colors.white54)),
            const SizedBox(height: 12),
            _QuickSellForm(planet: planet, gameProvider: gameProvider),
          ],
        ),
      ),
    );
  }

  Widget _buildMarketOverview(BuildContext context, GameProvider gameProvider) {
    final market = gameProvider.marketData;
    if (market == null) return const Center(child: CircularProgressIndicator());

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Обзор рынка', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _InfoTile('Покупки', market.buyOrders.values.fold(0, (sum, list) => sum + list.length).toString()),
                _InfoTile('Продажи', market.sellOrders.values.fold(0, (sum, list) => sum + list.length).toString()),
                _InfoTile('NPC торговцы', market.npcTraderCount.toString()),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCreateOrder(BuildContext context, GameProvider gameProvider) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Создать ордер', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 12),
            _OrderForm(gameProvider: gameProvider),
          ],
        ),
      ),
    );
  }

  Widget _buildBuyOrders(BuildContext context, GameProvider gameProvider) {
    final market = gameProvider.marketData;
    if (market == null) return const SizedBox.shrink();

    return _buildOrderList(context, market.buyOrders, 'Покупки', AppTheme.accentColor);
  }

  Widget _buildSellOrders(BuildContext context, GameProvider gameProvider) {
    final market = gameProvider.marketData;
    if (market == null) return const SizedBox.shrink();

    return _buildOrderList(context, market.sellOrders, 'Продажи', AppTheme.successColor);
  }

  Widget _buildOrderList(BuildContext context, Map<String, List> orders, String title, Color color) {
    final allOrders = orders.values.expand((l) => l).toList();
    if (allOrders.isEmpty) return const SizedBox.shrink();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title, style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: color)),
            const SizedBox(height: 8),
            ...allOrders.take(10).map((order) => Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: ListTile(
                    dense: true,
                    leading: CircleAvatar(
                      radius: 12,
                      backgroundColor: color.withOpacity(0.2),
                      child: Text(Constants.resourceIcons[order.resource] ?? '📦'),
                    ),
                    title: Text(
                      '${Constants.resourceNames[order.resource] ?? order.resource}: ${order.amount.toStringAsFixed(0)} @ ${order.price.toStringAsFixed(0)}',
                      style: const TextStyle(fontSize: 12),
                    ),
                    subtitle: Text(
                      'Итого: ${(order.amount * order.price).toStringAsFixed(0)}',
                      style: const TextStyle(fontSize: 10, color: Colors.white54),
                    ),
                  ),
                )),
          ],
        ),
      ),
    );
  }

  Widget _buildMyOrders(BuildContext context, GameProvider gameProvider) {
    final orders = gameProvider.myOrders;
    if (orders.isEmpty) return const SizedBox.shrink();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Мои ордера', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 8),
            ...orders.map((order) => Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: ListTile(
                    dense: true,
                    leading: CircleAvatar(
                      backgroundColor: order.isBuy
                          ? AppTheme.accentColor.withOpacity(0.2)
                          : AppTheme.successColor.withOpacity(0.2),
                      child: Text(order.isBuy ? 'П' : 'П'),
                    ),
                    title: Text(
                      '${order.isBuy ? "Купить" : "Продать"} ${Constants.resourceNames[order.resource] ?? order.resource} ${order.amount.toStringAsFixed(0)} @ ${order.price.toStringAsFixed(0)}',
                      style: const TextStyle(fontSize: 12),
                    ),
                    subtitle: Text(order.status, style: TextStyle(fontSize: 10, color: order.isActive ? AppTheme.successColor : Colors.white54)),
                    trailing: order.isActive
                        ? IconButton(
                            icon: const Icon(Icons.close, size: 16),
                            onPressed: () => gameProvider.deleteMarketOrder(order.id),
                          )
                        : null,
                  ),
                )),
          ],
        ),
      ),
    );
  }
}

class _QuickSellForm extends StatefulWidget {
  final Planet planet;
  final GameProvider gameProvider;

  const _QuickSellForm({required this.planet, required this.gameProvider});

  @override
  State<_QuickSellForm> createState() => _QuickSellFormState();
}

class _QuickSellFormState extends State<_QuickSellForm> {
  final _amountController = TextEditingController();

  @override
  void dispose() {
    _amountController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final foodAvailable = (widget.planet.resources['food'] as num?)?.toDouble() ?? 0.0;

    return Column(
      children: [
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: _amountController,
                decoration: InputDecoration(
                  labelText: 'Количество (кратно 10)',
                  hintText: 'Макс: ${(foodAvailable ~/ 10) * 10}',
                ),
                keyboardType: TextInputType.number,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                'Вы получите: ${(_amountController.text.isNotEmpty ? (double.tryParse(_amountController.text) ?? 0) / 10 : 0).toStringAsFixed(0)} денег',
                style: const TextStyle(fontSize: 12, color: Colors.green),
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: () {
              final amount = double.tryParse(_amountController.text);
              if (amount != null && amount > 0 && amount <= foodAvailable && amount % 10 == 0) {
                widget.gameProvider.sellFood(widget.planet.id, amount);
                _amountController.clear();
              }
            },
            child: const Text('Продать еду'),
          ),
        ),
      ],
    );
  }
}

class _OrderForm extends StatefulWidget {
  final GameProvider gameProvider;

  const _OrderForm({required this.gameProvider});

  @override
  State<_OrderForm> createState() => _OrderFormState();
}

class _OrderFormState extends State<_OrderForm> {
  final _amountController = TextEditingController();
  final _priceController = TextEditingController();
  String _resource = 'food';
  String _orderType = 'buy';

  @override
  void dispose() {
    _amountController.dispose();
    _priceController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Row(
          children: [
            Expanded(
              child: DropdownButtonFormField<String>(
                value: _resource,
                decoration: const InputDecoration(labelText: 'Ресурс'),
                items: Constants.resourceNames.keys
                    .where((k) => k != 'energy' && k != 'money' && k != 'alien_tech')
                    .map((k) => DropdownMenuItem(value: k, child: Text(Constants.resourceNames[k]!)))
                    .toList(),
                onChanged: (v) => setState(() => _resource = v!),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: DropdownButtonFormField<String>(
                value: _orderType,
                decoration: const InputDecoration(labelText: 'Тип'),
                items: const [
                  DropdownMenuItem(value: 'buy', child: Text('Купить')),
                  DropdownMenuItem(value: 'sell', child: Text('Продать')),
                ],
                onChanged: (v) => setState(() => _orderType = v!),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: _amountController,
                decoration: const InputDecoration(labelText: 'Количество'),
                keyboardType: TextInputType.number,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: TextField(
                controller: _priceController,
                decoration: const InputDecoration(labelText: 'Цена'),
                keyboardType: TextInputType.number,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: () {
              final amount = double.tryParse(_amountController.text);
              final price = double.tryParse(_priceController.text);
              if (amount != null && price != null && amount > 0 && price > 0) {
                widget.gameProvider.createMarketOrder(
                  resource: _resource,
                  orderType: _orderType,
                  amount: amount,
                  price: price,
                );
              }
            },
            child: Text('Создать ордер ${_orderType.toUpperCase()}'),
          ),
        ),
      ],
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
        Text(value, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white)),
      ],
    );
  }
}

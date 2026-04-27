import 'package:flutter/material.dart';
import '../core/app_theme.dart';
import '../models/market.dart';

class MarketOrderList extends StatelessWidget {
  final Map<String, List<MarketOrder>> orders;
  final String title;
  final Color color;

  const MarketOrderList({
    super.key,
    required this.orders,
    required this.title,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
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
            ...allOrders.take(10).map((order) => _OrderRow(order: order, color: color)),
          ],
        ),
      ),
    );
  }
}

class _OrderRow extends StatelessWidget {
  final MarketOrder order;
  final Color color;

  const _OrderRow({required this.order, required this.color});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        children: [
          Container(
            width: 32,
            height: 32,
            decoration: BoxDecoration(
              color: color.withValues(alpha: 0.2),
              shape: BoxShape.circle,
            ),
            child: Center(
              child: Text(
                Constants.resourceIcons[order.resource] ?? '📦',
                style: const TextStyle(fontSize: 16),
              ),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  Constants.resourceNames[order.resource] ?? order.resource,
                  style: const TextStyle(fontSize: 13, color: Colors.white),
                ),
                Text(
                  '${order.amount.toStringAsFixed(0)} @ ${order.price.toStringAsFixed(0)}',
                  style: const TextStyle(fontSize: 11, color: Colors.white54),
                ),
              ],
            ),
          ),
          Text(
            'Итого: ${(order.amount * order.price).toStringAsFixed(0)}',
            style: const TextStyle(fontWeight: FontWeight.bold, color: AppTheme.accentColor),
          ),
        ],
      ),
    );
  }
}

class Constants {
  static const resourceIcons = {
    'food': '🍍',
    'composite': '🧬',
    'mechanisms': '⚙️',
    'reagents': '🧪',
    'energy': '⚡',
    'money': '💰',
    'alien_tech': '👾',
  };

  static const resourceNames = {
    'food': 'Еда',
    'composite': 'Композит',
    'mechanisms': 'Механизмы',
    'reagents': 'Реагенты',
    'energy': 'Энергия',
    'money': 'Деньги',
    'alien_tech': 'Чертежи',
  };
}

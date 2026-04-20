class MarketOrder {
  final String id;
  final String planetId;
  final String playerId;
  final String resource;
  final String orderType;
  final double amount;
  final double price;
  final bool isPrivate;
  final String? link;
  final String status;
  final Map<String, double>? reservedResources;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  MarketOrder({
    required this.id,
    required this.planetId,
    required this.playerId,
    required this.resource,
    required this.orderType,
    required this.amount,
    required this.price,
    this.isPrivate = false,
    this.link,
    this.status = 'active',
    this.reservedResources,
    this.createdAt,
    this.updatedAt,
  });

  factory MarketOrder.fromJson(Map<String, dynamic> json) {
    return MarketOrder(
      id: json['id'] as String,
      planetId: json['planet_id'] as String,
      playerId: json['player_id'] as String,
      resource: json['resource'] as String,
      orderType: json['order_type'] as String,
      amount: (json['amount'] as num?)?.toDouble() ?? 0,
      price: (json['price'] as num?)?.toDouble() ?? 0,
      isPrivate: json['is_private'] as bool? ?? false,
      link: json['link'] as String?,
      status: json['status'] as String? ?? 'active',
      reservedResources: (json['reserved_resources'] as Map?)
              ?.map((k, v) => MapEntry(k as String, (v as num).toDouble()))
          ,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : null,
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'] as String)
          : null,
    );
  }

  bool get isActive => status == 'active';
  bool get isBuy => orderType == 'buy';
  bool get isSell => orderType == 'sell';
}

class MarketData {
  final Map<String, List<MarketOrder>> buyOrders;
  final Map<String, List<MarketOrder>> sellOrders;
  final double bestBuyPrice;
  final double bestSellPrice;
  final double totalVolume;
  final int activeOrders;
  final int npcTraderCount;

  MarketData({
    this.buyOrders = const {},
    this.sellOrders = const {},
    this.bestBuyPrice = 0,
    this.bestSellPrice = 0,
    this.totalVolume = 0,
    this.activeOrders = 0,
    this.npcTraderCount = 0,
  });

  factory MarketData.fromJson(Map<String, dynamic> json) {
    final buyOrders = <String, List<MarketOrder>>{};
    final sellOrders = <String, List<MarketOrder>>{};

    (json['buy_orders'] as Map?)?.forEach((key, value) {
      if (value is List) {
        buyOrders[key as String] = value
            .map((e) => MarketOrder.fromJson(e as Map<String, dynamic>))
            .toList();
      }
    });

    (json['sell_orders'] as Map?)?.forEach((key, value) {
      if (value is List) {
        sellOrders[key as String] = value
            .map((e) => MarketOrder.fromJson(e as Map<String, dynamic>))
            .toList();
      }
    });

    return MarketData(
      buyOrders: buyOrders,
      sellOrders: sellOrders,
      bestBuyPrice: (json['best_buy_price'] as num?)?.toDouble() ?? 0,
      bestSellPrice: (json['best_sell_price'] as num?)?.toDouble() ?? 0,
      totalVolume: (json['total_volume'] as num?)?.toDouble() ?? 0,
      activeOrders: json['active_orders'] as int? ?? 0,
      npcTraderCount: json['npc_trader_count'] as int? ?? 0,
    );
  }
}

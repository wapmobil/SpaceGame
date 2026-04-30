import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../models/market.dart';

class MarketProvider extends ChangeNotifier {
  String? _authToken;
  String? baseUrl;
  String? _planetId;

  MarketData? _marketData;
  List<MarketOrder> _myOrders = [];
  String? _errorMessage;

  MarketData? get marketData => _marketData;
  List<MarketOrder> get myOrders => _myOrders;
  String? get errorMessage => _errorMessage;

  void setAuthToken(String token) {
    _authToken = token;
  }

  void setPlanetId(String planetId) {
    _planetId = planetId;
  }

  Future<void> loadMarketData() async {
    if (_authToken == null) return;
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/market'),
        headers: {'X-Auth-Token': _authToken!},
      );

      if (response.statusCode == 200) {
        _marketData = MarketData.fromJson(
            jsonDecode(response.body) as Map<String, dynamic>);
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load market: $e');
    }
  }

  Future<void> loadMyOrders() async {
    if (_authToken == null || _planetId == null) return;
    try {
      final response = await http.get(
        Uri.parse('${baseUrl!}/api/planets/$_planetId/market/orders'),
        headers: {'X-Auth-Token': _authToken!},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as List;
        _myOrders = data.map((e) => MarketOrder.fromJson(e as Map<String, dynamic>)).toList();
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Failed to load my orders: $e');
    }
  }

  Future<void> createMarketOrder({
    required String resource,
    required String orderType,
    required double amount,
    required double price,
    bool isPrivate = false,
  }) async {
    if (_authToken == null || _planetId == null) return;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$_planetId/market/orders'),
        headers: {
          'Content-Type': 'application/json',
          'X-Auth-Token': _authToken!,
        },
        body: jsonEncode({
          'resource': resource,
          'order_type': orderType,
          'amount': amount,
          'price': price,
          'is_private': isPrivate,
        }),
      );

      if (response.statusCode == 201) {
        await loadMarketData();
        await loadMyOrders();
      } else {
        _errorMessage = 'Не удалось создать ордер: ${response.body}';
        notifyListeners();
      }
    } catch (e) {
      _errorMessage = 'Ошибка создания ордера: $e';
      notifyListeners();
    }
  }

  Future<void> deleteMarketOrder(String orderId) async {
    if (_authToken == null) return;
    try {
      final response = await http.delete(
        Uri.parse('${baseUrl!}/api/market/orders/$orderId'),
        headers: {'X-Auth-Token': _authToken!},
      );

      if (response.statusCode == 200) {
        await loadMarketData();
        await loadMyOrders();
      } else {
        _errorMessage = 'Не удалось удалить ордер: ${response.body}';
        notifyListeners();
      }
    } catch (e) {
      _errorMessage = 'Ошибка удаления ордера: $e';
      notifyListeners();
    }
  }

  Future<bool> sellFood(double amount) async {
    if (_authToken == null || _planetId == null) return false;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$_planetId/sell-food'),
        headers: {'X-Auth-Token': _authToken!},
        body: jsonEncode({'amount': amount}),
      );

      if (response.statusCode == 200) {
        return true;
      } else {
        _errorMessage = 'Не удалось продать еду: ${response.body}';
        notifyListeners();
        return false;
      }
    } catch (e) {
      _errorMessage = 'Ошибка продажи еды: $e';
      notifyListeners();
      return false;
    }
  }

  Future<bool> sellIron(double amount) async {
    if (_authToken == null || _planetId == null) return false;
    try {
      final response = await http.post(
        Uri.parse('${baseUrl!}/api/planets/$_planetId/sell-iron'),
        headers: {'X-Auth-Token': _authToken!},
        body: jsonEncode({'amount': amount}),
      );

      if (response.statusCode == 200) {
        return true;
      } else {
        _errorMessage = 'Не удалось продать железо: ${response.body}';
        notifyListeners();
        return false;
      }
    } catch (e) {
      _errorMessage = 'Ошибка продажи железа: $e';
      notifyListeners();
      return false;
    }
  }

  void onMarketUpdate() {
    loadMarketData();
    loadMyOrders();
    notifyListeners();
  }

  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }
}

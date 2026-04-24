import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

typedef MessageListener = void Function(Map<String, dynamic> message);

class WebSocketManager extends ChangeNotifier {
  WebSocketChannel? _channel;
  bool _isConnected = false;
  String? _playerId;
  final List<MessageListener> _listeners = [];
  final List<Map<String, dynamic>> _messageHistory = [];
  Timer? _pingTimer;

  bool get isConnected => _isConnected;
  String? get playerId => _playerId;
  List<Map<String, dynamic>> get messageHistory => List.unmodifiable(_messageHistory);

  void connect(String baseUrl, String authToken) {
    final wsUrl = baseUrl.startsWith('https')
        ? baseUrl.replaceFirst('https', 'wss')
        : baseUrl.replaceFirst('http', 'ws');
    final uri = Uri.parse(wsUrl)
        .replace(path: '/ws', queryParameters: {'token': authToken});

    debugPrint('WebSocket connecting to: $uri');
    try {
      _channel = WebSocketChannel.connect(uri);
      _startPing();
      debugPrint('WebSocket channel created');

      _channel?.stream.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDone,
      );
    } catch (e) {
      debugPrint('WebSocket connection error: $e');
      _isConnected = false;
      notifyListeners();
    }
  }

  void send(Map<String, dynamic> message) {
    if (_channel != null && _isConnected) {
      _channel!.sink.add(jsonEncode(message));
    }
  }

  void subscribe(String planetId) {
    send({'type': 'subscribe', 'data': {'planet_id': planetId}});
  }

  void sendBuild({required String building, int? level}) {
    send({
      'type': 'build',
      'data': {'building': building, 'level': level ?? 1},
    });
  }

  void sendResearch({required String techId}) {
    send({'type': 'research', 'data': {'tech_id': techId}});
  }

  void sendBuildShip({required String shipType}) {
    send({'type': 'build_ship', 'data': {'ship_type': shipType}});
  }

  void sendStartExpedition({
    required String fleetId,
    required String target,
    String? expeditionType,
    double? duration,
  }) {
    send({
      'type': 'start_expedition',
      'data': {
        'fleet_id': fleetId,
        'target': target,
        'expedition_type': expeditionType,
        'duration': duration,
      },
    });
  }

    void sendPing() {
    send({'type': 'ping', 'data': {}});
  }

  void sendDrillCommand({String? direction, bool? extract}) {
    Map<String, dynamic> data = {
      'direction': direction ?? '',
    };
    if (extract != null) {
      data['extract'] = extract;
    }
    send({
      'type': 'drill_command',
      'data': data,
    });
  }

  void disconnect() {
    _pingTimer?.cancel();
    _channel?.sink.close();
    _channel = null;
    _isConnected = false;
    _playerId = null;
    notifyListeners();
  }

  void addMessageListener(MessageListener listener) {
    _listeners.add(listener);
  }

  void removeMessageListener(MessageListener listener) {
    _listeners.remove(listener);
  }

  void _onMessage(dynamic data) {
    final message = _parseMessage(data);
    _messageHistory.add(message);

    if (message['type'] == 'welcome') {
      _playerId = message['player_id'] ?? message['data']?['player_id'];
      _isConnected = true;
      notifyListeners();
    }

    for (final listener in List.from(_listeners)) {
      listener(message);
    }
    notifyListeners();
  }

  void _onError(dynamic error, StackTrace stackTrace) {
    debugPrint('WebSocket error: $error');
    _isConnected = false;
    notifyListeners();
  }

  void _onDone() {
    _isConnected = false;
    _pingTimer?.cancel();
    notifyListeners();
  }

  Map<String, dynamic> _parseMessage(dynamic data) {
    try {
      if (data is String) {
        return jsonDecode(data) as Map<String, dynamic>;
      }
      return {'raw': data};
    } catch (e) {
      return {'raw': data, 'error': e.toString()};
    }
  }

  void _startPing() {
    _pingTimer?.cancel();
    _pingTimer = Timer.periodic(const Duration(seconds: 30), (_) {
      if (_isConnected) {
        sendPing();
      }
    });
  }

  @override
  void dispose() {
    _pingTimer?.cancel();
    disconnect();
    super.dispose();
  }
}

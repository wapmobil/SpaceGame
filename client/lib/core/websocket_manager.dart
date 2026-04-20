import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/io.dart';

class WebSocketManager extends ChangeNotifier {
  WebSocketChannel? _channel;
  bool _isConnected = false;
  String? _playerId;
  final List<Map<String, dynamic>> _messageHistory = [];

  bool get isConnected => _isConnected;
  String? get playerId => _playerId;
  List<Map<String, dynamic>> get messageHistory => List.unmodifiable(_messageHistory);

  void connect(String baseUrl, String authToken) {
    final uri = Uri.parse(baseUrl.replaceFirst('http', 'ws'))
        .replace(
          path: '/ws',
          queryParameters: {'token': authToken},
        );

    _channel = IOWebSocketChannel.connect(uri.toString());
    _isConnected = true;
    notifyListeners();

    _channel?.stream.listen(
      _onMessage,
      onError: _onError,
      onDone: _onDone,
    );
  }

  void send(Map<String, dynamic> message) {
    if (_channel != null && _isConnected) {
      _channel!.sink.add(message.toString());
    }
  }

  void disconnect() {
    _channel?.sink.close();
    _channel = null;
    _isConnected = false;
    _playerId = null;
    notifyListeners();
  }

  void _onMessage(dynamic data) {
    final message = _parseMessage(data);
    _messageHistory.add(message);
    notifyListeners();

    if (message['type'] == 'welcome') {
      _playerId = message['player_id'];
    }
  }

  void _onError(dynamic error, StackTrace stackTrace) {
    _isConnected = false;
    notifyListeners();
  }

  void _onDone() {
    _isConnected = false;
    notifyListeners();
  }

  Map<String, dynamic> _parseMessage(dynamic data) {
    try {
      return {'raw': data};
    } catch (e) {
      return {'raw': data, 'error': e.toString()};
    }
  }
}

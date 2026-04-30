import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ServerConfig extends ChangeNotifier {
  static const _key = 'server_url';
  static final ServerConfig _instance = ServerConfig._internal();
  String _url = 'http://localhost:8088';

  factory ServerConfig() => _instance;

  ServerConfig._internal();

  static ServerConfig get instance => _instance;

  String get url => _url;
  String get host => _url.replaceFirst(RegExp(r'^https?://'), '');
  String get scheme {
    if (_url.startsWith('https://')) return 'https';
    return 'http';
  }

  static String get baseUri => _instance._url;

  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    final saved = prefs.getString(_key);
    if (saved != null && saved.isNotEmpty) {
      _url = saved.startsWith('http') ? saved : 'http://$saved';
      notifyListeners();
    }
  }

  Future<void> setUrl(String value) async {
    final url = value.trim();
    if (url.isEmpty) return;
    _url = url.startsWith('http') ? url : 'http://$url';
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_key, _url);
    notifyListeners();
  }

  void reset() {
    _url = 'http://localhost:8088';
    SharedPreferences.getInstance().then((prefs) {
      prefs.remove(_key);
    });
    notifyListeners();
  }
}

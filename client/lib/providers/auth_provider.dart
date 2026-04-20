import 'package:flutter/material.dart';

class AuthProvider extends ChangeNotifier {
  bool _isLoading = false;
  String? _error;

  bool get isLoading => _isLoading;
  String? get error => _error;

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }

  void _setError(String? msg) {
    _error = msg;
    notifyListeners();
  }

  Future<bool> login(String name, Function onSuccess, Function onError) async {
    _setLoading(true);
    _setError(null);

    try {
      await onSuccess(name);
      _setLoading(false);
      return true;
    } catch (e) {
      _setError(e.toString());
      _setLoading(false);
      onError(e);
      return false;
    }
  }

  void clearError() {
    _setError(null);
    notifyListeners();
  }
}

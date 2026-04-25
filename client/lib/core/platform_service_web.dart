import 'dart:js_interop';
import 'platform_service.dart';

@JS('window.location.host')
external JSString? get _host;

@JS('window.location.protocol')
external JSString? get _protocol;

class _PlatformServiceWeb implements PlatformService {
  @override
  String get host {
    final h = _host?.toDart;
    if (h != null && h.isNotEmpty) return h;
    return 'localhost:8088';
  }

  @override
  String get scheme {
    final p = _protocol?.toDart;
    if (p != null) return p == 'https:' ? 'https' : 'http';
    return 'http';
  }
}

PlatformService createPlatformService() => _PlatformServiceWeb();

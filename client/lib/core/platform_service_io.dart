import 'dart:io' show Platform;
import 'platform_service.dart';

class _PlatformServiceDefault implements PlatformService {
  @override
  String get host => Platform.isAndroid || Platform.isIOS ? '' : 'localhost:8088';

  @override
  String get scheme => 'http';
}

PlatformService createPlatformService() => _PlatformServiceDefault();

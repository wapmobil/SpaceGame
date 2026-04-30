import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'core/websocket_manager.dart';
import 'core/app_theme.dart';
import 'core/server_config.dart';
import 'providers/game_provider.dart';
import 'providers/auth_provider.dart';
import 'screens/landing_screen.dart';
import 'screens/home_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await ServerConfig().init();
  runApp(const SpaceGameApp());
}

class SpaceGameApp extends StatelessWidget {
  const SpaceGameApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => ServerConfig()),
        ChangeNotifierProvider(create: (_) => WebSocketManager()),
        ChangeNotifierProvider(
          create: (context) {
            final ws = context.read<WebSocketManager>();
            return GameProvider(websocket: ws);
          },
        ),
        ChangeNotifierProvider(create: (_) => AuthProvider()),
      ],
      child: MaterialApp(
        title: 'SpaceGame',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.theme,
        initialRoute: '/',
        routes: {
          '/': (context) => const LandingScreen(),
          '/home': (context) => const HomeScreen(),
        },
      ),
    );
  }
}

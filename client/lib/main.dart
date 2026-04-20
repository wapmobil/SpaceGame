import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'core/websocket_manager.dart';
import 'screens/landing_screen.dart';

void main() {
  runApp(const SpaceGameApp());
}

class SpaceGameApp extends StatelessWidget {
  const SpaceGameApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => WebSocketManager()),
      ],
      child: MaterialApp(
        title: 'SpaceGame',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(
          colorScheme: ColorScheme.fromSeed(
            seedColor: const Color(0xFF1a1a2e),
            brightness: Brightness.dark,
          ),
          useMaterial3: true,
        ),
        home: const LandingScreen(),
      ),
    );
  }
}

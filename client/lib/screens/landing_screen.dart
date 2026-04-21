import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../providers/game_provider.dart';
import '../providers/auth_provider.dart';
import '../core/app_theme.dart';
import '../screens/home_screen.dart';

class LandingScreen extends StatefulWidget {
  const LandingScreen({super.key});

  @override
  State<LandingScreen> createState() => _LandingScreenState();
}

class _LandingScreenState extends State<LandingScreen>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _fadeAnimation;
  final _nameController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    );
    _fadeAnimation = CurvedAnimation(
      parent: _controller,
      curve: Curves.easeInOut,
    );
    _controller.forward();
    _loadSavedName();
  }

  Future<void> _loadSavedName() async {
    final prefs = await SharedPreferences.getInstance();
    final savedName = prefs.getString('player_name');
    if (savedName != null && savedName.isNotEmpty) {
      _nameController.text = savedName;
    }
  }

  Future<void> _handleLogin() async {
    final name = _nameController.text.trim();
    if (name.isEmpty) return;

    final authProvider = context.read<AuthProvider>();
    final gameProvider = context.read<GameProvider>();

    final success = await authProvider.login(
      name,
      (n) => gameProvider.login(n),
      (e) => debugPrint('Login failed: $e'),
    );

    if (success && mounted) {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => const HomeScreen()),
      );
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    _nameController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.of(context).size;
    final isSmall = size.width < 600;

    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              AppTheme.backgroundColor,
              AppTheme.primaryColor,
              AppTheme.secondaryColor,
            ],
          ),
        ),
        child: SafeArea(
          child: Center(
            child: AnimatedBuilder(
              animation: _fadeAnimation,
              builder: (context, child) {
                return FadeTransition(
                  opacity: _fadeAnimation,
                  child: Transform.translate(
                    offset: Offset(0, (1 - _fadeAnimation.value) * 30),
                    child: child,
                  ),
                );
              },
              child: ConstrainedBox(
                constraints: BoxConstraints(
                  maxWidth: isSmall ? double.infinity : 420,
                  minWidth: isSmall ? size.width * 0.9 : 320,
                ),
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.space_dashboard, size: isSmall ? 80 : 100, color: AppTheme.accentColor),
                      const SizedBox(height: 16),
                      Text(
                        'SpaceGame',
                        style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                              color: Colors.white,
                              fontWeight: FontWeight.bold,
                              letterSpacing: 2,
                            ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Command your fleet across the galaxy',
                        style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                              color: Colors.white70,
                            ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 40),
                      TextField(
                        controller: _nameController,
                        decoration: const InputDecoration(
                          labelText: 'Player Name',
                          hintText: 'Enter your name',
                          prefixIcon: Icon(Icons.person_outline),
                        ),
                        onSubmitted: (_) => _handleLogin(),
                        style: const TextStyle(color: Colors.white),
                      ),
                      const SizedBox(height: 16),
                      SizedBox(
                        width: double.infinity,
                        child: ElevatedButton(
                          onPressed: _handleLogin,
                          child: const Text('Play'),
                        ),
                      ),
                      const SizedBox(height: 24),
                      const Text(
                        'v1.0.0',
                        style: TextStyle(color: Colors.white38, fontSize: 12),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

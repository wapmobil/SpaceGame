import 'package:flutter/material.dart';

class AppTheme {
  static const primaryColor = Color(0xFF0f3460);
  static const secondaryColor = Color(0xFFe94560);
  static const backgroundColor = Color(0xFF1a1a2e);
  static const cardColor = Color(0xFF16213e);
  static const accentColor = Color(0xFF00d2ff);
  static const successColor = Color(0xFF00c853);
  static const warningColor = Color(0xFFFFab00);
  static const dangerColor = Color(0xFFff1744);
  static const Map<String, Color> resourceColors = {
    'food': Color(0xFFff9800),
    'composite': Color(0xFF8bc34a),
    'mechanisms': Color(0xFF607d8b),
    'reagents': Color(0xFF9c27b0),
    'energy': Color(0xFFFFeb3b),
    'money': Color(0xFFffd700),
    'alien_tech': Color(0xFF00bcd4),
  };

  static const resourceIcons = {
    'food': '🍍',
    'composite': '🧬',
    'mechanisms': '⚙️',
    'reagents': '🧪',
    'energy': '⚡',
    'money': '💰',
    'alien_tech': '👾',
  };

  static ThemeData get theme => ThemeData(
        useMaterial3: true,
        brightness: Brightness.dark,
        colorScheme: const ColorScheme.dark(
          primary: primaryColor,
          secondary: secondaryColor,
          surface: cardColor,
          onSurface: Colors.white,
          error: dangerColor,
        ),
        scaffoldBackgroundColor: backgroundColor,
        appBarTheme: const AppBarTheme(
          backgroundColor: primaryColor,
          elevation: 0,
          centerTitle: true,
          titleTextStyle: TextStyle(
            color: Colors.white,
            fontSize: 20,
            fontWeight: FontWeight.w600,
          ),
          iconTheme: IconThemeData(color: Colors.white),
        ),
        cardTheme: CardThemeData(
          color: cardColor,
          elevation: 2,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: const BorderSide(color: Color(0xFF0f3460), width: 1),
          ),
        ),
        navigationBarTheme: NavigationBarThemeData(
          backgroundColor: cardColor,
          indicatorColor: secondaryColor.withValues(alpha: 0.3),
          labelTextStyle: WidgetStateProperty.resolveWith((states) {
            if (states.contains(WidgetState.selected)) {
              return const TextStyle(color: secondaryColor);
            }
            return const TextStyle(color: Colors.white70);
          }),
          iconTheme: WidgetStateProperty.resolveWith((states) {
            if (states.contains(WidgetState.selected)) {
              return const IconThemeData(color: secondaryColor);
            }
            return const IconThemeData(color: Colors.white54);
          }),
        ),
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: primaryColor,
            foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        ),
        textButtonTheme: TextButtonThemeData(
          style: TextButton.styleFrom(
            foregroundColor: accentColor,
          ),
        ),
        inputDecorationTheme: InputDecorationTheme(
          filled: true,
          fillColor: cardColor,
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: const BorderSide(color: Color(0xFF0f3460)),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: const BorderSide(color: Color(0xFF0f3460)),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: const BorderSide(color: accentColor, width: 2),
          ),
          labelStyle: const TextStyle(color: Colors.white70),
          hintStyle: const TextStyle(color: Colors.white38),
        ),
        floatingActionButtonTheme: const FloatingActionButtonThemeData(
          backgroundColor: secondaryColor,
          foregroundColor: Colors.white,
        ),
        dividerTheme: const DividerThemeData(
          color: Color(0xFF0f3460),
          thickness: 1,
        ),
      );
}

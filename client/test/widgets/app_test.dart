import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:spacegame/main.dart';
import 'package:spacegame/screens/landing_screen.dart';

void main() {
  testWidgets('SpaceGameApp renders LandingScreen', (WidgetTester tester) async {
    await tester.pumpWidget(const SpaceGameApp());

    expect(find.text('SpaceGame'), findsOneWidget);
    expect(find.byType(ElevatedButton), findsOneWidget);
  });

  testWidgets('LandingScreen displays content', (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: LandingScreen()));

    expect(find.text('SpaceGame'), findsOneWidget);
    expect(find.text('Play'), findsOneWidget);
  });
}

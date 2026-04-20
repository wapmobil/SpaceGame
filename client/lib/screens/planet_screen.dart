import 'package:flutter/material.dart';

class PlanetScreen extends StatelessWidget {
  const PlanetScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Planet')),
      body: const Center(child: Text('Planet View')),
    );
  }
}

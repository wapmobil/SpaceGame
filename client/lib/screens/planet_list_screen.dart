import 'package:flutter/material.dart';

class PlanetListScreen extends StatelessWidget {
  const PlanetListScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Planets')),
      body: const Center(child: Text('Select a planet')),
    );
  }
}

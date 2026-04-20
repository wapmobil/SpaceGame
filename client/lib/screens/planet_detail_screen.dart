import 'package:flutter/material.dart';

class PlanetDetailScreen extends StatelessWidget {
  final String planetId;

  const PlanetDetailScreen({super.key, required this.planetId});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Planet Details')),
      body: const Center(child: Text('Planet details')),
    );
  }
}

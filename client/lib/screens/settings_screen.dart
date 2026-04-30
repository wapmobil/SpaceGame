import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/game_provider.dart';
import '../core/server_config.dart';
import '../core/app_theme.dart';

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Настройки')),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, _) {
          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildProfileSection(context, gameProvider),
                const SizedBox(height: 16),
                _buildConnectionSection(context, gameProvider),
                const SizedBox(height: 16),
                _buildServerUrlSection(context, gameProvider),
                const SizedBox(height: 16),
                _buildLeaderboardSection(context, gameProvider),
                const SizedBox(height: 16),
                _buildStatsSection(context, gameProvider),
                const SizedBox(height: 16),
                _buildEventsSection(context, gameProvider),
                const SizedBox(height: 24),
                _buildLogoutButton(context, gameProvider),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildProfileSection(BuildContext context, GameProvider gameProvider) {
    final player = gameProvider.player;
    if (player == null) return const SizedBox.shrink();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const CircleAvatar(
                  radius: 24,
                  backgroundColor: AppTheme.secondaryColor,
                  child: Icon(Icons.person, size: 28, color: Colors.white),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        player.name,
                        style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white),
                      ),
                      Text(
                        player.id,
                        style: const TextStyle(fontSize: 11, color: Colors.white54),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildConnectionSection(BuildContext context, GameProvider gameProvider) {
    final ws = gameProvider.websocket;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Соединение', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 12),
            Row(
              children: [
                Container(
                  width: 12,
                  height: 12,
                  decoration: BoxDecoration(
                    color: ws.isConnected ? AppTheme.successColor : AppTheme.dangerColor,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  ws.isConnected ? 'Подключено' : 'Отключено',
                  style: TextStyle(
                    color: ws.isConnected ? AppTheme.successColor : AppTheme.dangerColor,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text('Server: ${gameProvider.baseUrl}', style: const TextStyle(fontSize: 12, color: Colors.white54)),
          ],
        ),
      ),
    );
  }

  Widget _buildServerUrlSection(BuildContext context, GameProvider gameProvider) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Сервер', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
            const SizedBox(height: 12),
            Consumer<ServerConfig>(
              builder: (context, config, _) {
                return OutlinedButton.icon(
                  onPressed: () async {
                    final controller = TextEditingController(text: config.url);
                    final result = await showDialog<String>(
                      context: context,
                      builder: (dialogContext) {
                        return AlertDialog(
                          backgroundColor: AppTheme.cardColor,
                          title: const Text('Адрес сервера'),
                          content: TextField(
                            controller: controller,
                            decoration: const InputDecoration(
                              hintText: 'localhost:8088',
                              prefixIcon: Icon(Icons.cloud_outlined),
                            ),
                            style: const TextStyle(color: Colors.white),
                            autofocus: true,
                          ),
                          actions: [
                            TextButton(
                              onPressed: () => Navigator.pop(dialogContext),
                              child: const Text('Отмена'),
                            ),
                            ElevatedButton(
                              onPressed: () => Navigator.pop(dialogContext, controller.text),
                              child: const Text('Сохранить'),
                            ),
                          ],
                        );
                      },
                    );
                    if (result != null && result.trim().isNotEmpty) {
                      await config.setUrl(result);
                      if (context.mounted) {
                        await gameProvider.logout();
                        if (context.mounted) {
                          Navigator.of(context).pushNamedAndRemoveUntil('/', (route) => false);
                        }
                      }
                    }
                  },
                  icon: const Icon(Icons.cloud_outlined),
                  label: Text(config.url),
                  style: OutlinedButton.styleFrom(foregroundColor: AppTheme.accentColor),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLeaderboardSection(BuildContext context, GameProvider gameProvider) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('Таблица лидеров', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                TextButton(
                  onPressed: () => gameProvider.loadRatings(),
                  child: const Text('Обновить'),
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (gameProvider.ratings.isEmpty)
              const Center(
                child: Padding(
                  padding: EdgeInsets.symmetric(vertical: 16),
                  child: Text('Нет данных', style: TextStyle(color: Colors.white38)),
                ),
              )
            else
              ...gameProvider.ratings.take(10).map((entry) => Padding(
                    padding: const EdgeInsets.only(bottom: 4),
                    child: ListTile(
                      dense: true,
                      leading: CircleAvatar(
                        radius: 12,
                        backgroundColor: entry.rank <= 3
                            ? (entry.rank == 1
                                ? AppTheme.warningColor
                                : entry.rank == 2
                                    ? Colors.grey
                                    : Colors.brown)
                            : AppTheme.primaryColor,
                        child: Text(
                          '#${entry.rank}',
                          style: const TextStyle(fontSize: 10, color: Colors.white, fontWeight: FontWeight.bold),
                        ),
                      ),
                      title: Text(entry.playerName, style: const TextStyle(fontSize: 13, color: Colors.white)),
                      subtitle: Text(entry.category, style: const TextStyle(fontSize: 10, color: Colors.white54)),
                      trailing: Text(
                        entry.value.toStringAsFixed(0),
                        style: const TextStyle(fontWeight: FontWeight.bold, color: AppTheme.accentColor),
                      ),
                    ),
                  )),
          ],
        ),
      ),
    );
  }

  Widget _buildStatsSection(BuildContext context, GameProvider gameProvider) {
    final stats = gameProvider.stats;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('Статистика', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                if (gameProvider.selectedPlanet != null)
                  TextButton(
                    onPressed: () => gameProvider.loadStats(gameProvider.selectedPlanet!.id),
                    child: const Text('Обновить'),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            if (stats == null)
              const Center(
                child: Padding(
                  padding: EdgeInsets.symmetric(vertical: 16),
                  child: Text('Нет данных', style: TextStyle(color: Colors.white38)),
                ),
              )
            else ...[
              if (stats['planet_name'] != null)
                Text(stats['planet_name'] as String, style: const TextStyle(fontSize: 14, color: AppTheme.accentColor)),
              if (stats['lifetime'] != null)
                ...((stats['lifetime'] as Map?)?.entries.map((e) {
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 4),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              (e.key as String).replaceAll('_', ' ').split(' ').map((w) => w[0].toUpperCase() + w.substring(1)).join(' '),
                              style: const TextStyle(fontSize: 12, color: Colors.white70),
                            ),
                            Text(
                              (e.value as num).toStringAsFixed(0),
                              style: const TextStyle(fontSize: 12, color: Colors.white),
                            ),
                          ],
                        ),
                      );
                    }).toList() ??
                    []),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildEventsSection(BuildContext context, GameProvider gameProvider) {
    final events = gameProvider.events;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('События', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white70)),
                if (gameProvider.selectedPlanet != null)
                  TextButton(
                    onPressed: () => gameProvider.loadEvents(gameProvider.selectedPlanet!.id),
                    child: const Text('Обновить'),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            if (events.isEmpty)
              const Center(
                child: Padding(
                  padding: EdgeInsets.symmetric(vertical: 16),
                  child: Text('Нет событий', style: TextStyle(color: Colors.white38)),
                ),
              )
            else
              ...events.where((e) => !(e['resolved'] as bool? ?? false)).take(5).map((event) {
                return Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: ListTile(
                    dense: true,
                    leading: const Icon(Icons.event_note, size: 16, color: AppTheme.warningColor),
                    title: Text(
                      event['description'] as String? ?? 'Неизвестное событие',
                      style: const TextStyle(fontSize: 12, color: Colors.white),
                    ),
                    subtitle: Text(
                      event['type'] as String? ?? '',
                      style: const TextStyle(fontSize: 10, color: Colors.white54),
                    ),
                    trailing: OutlinedButton(
                      onPressed: () => gameProvider.resolveEvent(event['type'] as String? ?? ''),
                      child: const Text('Решить', style: TextStyle(fontSize: 10)),
                    ),
                  ),
                );
              }),
          ],
        ),
      ),
    );
  }

  Widget _buildLogoutButton(BuildContext context, GameProvider gameProvider) {
    return SizedBox(
      width: double.infinity,
      child: OutlinedButton.icon(
        onPressed: () async {
          final confirmed = await showDialog<bool>(
            context: context,
            builder: (context) => AlertDialog(
              title: const Text('Выход'),
              content: const Text('Вы уверены, что хотите выйти?'),
              actions: [
                TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Отмена')),
                ElevatedButton(onPressed: () => Navigator.pop(context, true), child: const Text('Выйти')),
              ],
            ),
          );
          if (confirmed == true) {
            await gameProvider.logout();
            if (context.mounted) {
              Navigator.of(context).pushNamedAndRemoveUntil('/', (route) => false);
            }
          }
        },
        icon: const Icon(Icons.logout),
        label: const Text('Выйти'),
        style: OutlinedButton.styleFrom(foregroundColor: AppTheme.dangerColor),
      ),
    );
  }
}

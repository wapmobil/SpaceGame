class RatingEntry {
  final int rank;
  final String planetId;
  final String playerName;
  final String category;
  final double value;
  final String updated;

  RatingEntry({
    required this.rank,
    required this.planetId,
    required this.playerName,
    required this.category,
    required this.value,
    required this.updated,
  });

  factory RatingEntry.fromJson(Map<String, dynamic> json) {
    return RatingEntry(
      rank: json['rank'] as int? ?? 0,
      planetId: json['planet_id'] as String,
      playerName: json['player_name'] as String,
      category: json['category'] as String,
      value: (json['value'] as num?)?.toDouble() ?? 0,
      updated: json['updated'] as String,
    );
  }
}

class RatingResponse {
  final String category;
  final List<RatingEntry> entries;
  final int total;

  RatingResponse({
    this.category = '',
    this.entries = const [],
    this.total = 0,
  });

  factory RatingResponse.fromJson(Map<String, dynamic> json) {
    return RatingResponse(
      category: json['category'] as String? ?? '',
      entries: (json['entries'] as List?)
              ?.map((e) => RatingEntry.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      total: json['total'] as int? ?? 0,
    );
  }
}

class Player {
  final String id;
  final String authToken;
  final String name;

  Player({required this.id, required this.authToken, this.name = ''});

  factory Player.fromJson(Map<String, dynamic> json) {
    return Player(
      id: json['id'] as String,
      authToken: json['auth_token'] as String,
      name: json['name'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'auth_token': authToken,
      'name': name,
    };
  }
}

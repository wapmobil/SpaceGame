class Player {
  final String id;
  final String authToken;

  Player({required this.id, required this.authToken});

  factory Player.fromJson(Map<String, dynamic> json) {
    return Player(
      id: json['id'] as String,
      authToken: json['auth_token'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'auth_token': authToken,
    };
  }
}

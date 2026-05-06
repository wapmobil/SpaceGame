import 'planet_survey.dart';

class ExpeditionChain {
  final String id;
  final String planetId;
  final String ownerId;
  final String status;
  final int eventCount;
  final int currentEventIndex;
  final Map<String, double> inventory;
  final List<ExpeditionEvent> events;
  final Location? discoveredLocation;
  final DateTime createdAt;
  final DateTime updatedAt;

  ExpeditionChain({
    required this.id,
    required this.planetId,
    required this.ownerId,
    required this.status,
    required this.eventCount,
    required this.currentEventIndex,
    required this.inventory,
    required this.events,
    this.discoveredLocation,
    required this.createdAt,
    required this.updatedAt,
  });

  bool get isActive => status == 'active';
  bool get isGenerating => status == 'generating';
  bool get isCompleted => status == 'completed';
  bool get isFailed => status == 'failed';
  double get totalInventory =>
      inventory.values.fold(0.0, (sum, v) => sum + v);

  factory ExpeditionChain.fromJson(Map<String, dynamic> json) {
    final inventoryJson = json['inventory'] as Map<String, dynamic>? ?? {};
    final inventory = <String, double>{};
    inventoryJson.forEach((key, value) {
      inventory[key] = (value as num).toDouble();
    });

    final eventsJson = json['events'] as List? ?? [];
    final events = eventsJson
        .map((e) => ExpeditionEvent.fromJson(e as Map<String, dynamic>))
        .toList();

    return ExpeditionChain(
      id: json['id'] as String,
      planetId: json['planet_id'] as String,
      ownerId: json['owner_id'] as String,
      status: json['status'] as String? ?? 'active',
      eventCount: json['event_count'] as int? ?? 0,
      currentEventIndex: json['current_event_index'] as int? ?? 0,
      inventory: inventory,
      events: events,
      discoveredLocation: json['discovered_location'] != null
          ? Location.fromJson(json['discovered_location'] as Map<String, dynamic>)
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now().toUtc(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'] as String)
          : DateTime.now().toUtc(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'planet_id': planetId,
      'owner_id': ownerId,
      'status': status,
      'event_count': eventCount,
      'current_event_index': currentEventIndex,
      'inventory': inventory,
      'events': events.map((e) => e.toJson()).toList(),
      'discovered_location': discoveredLocation?.toJson(),
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }
}

class ExpeditionEvent {
  final String eventId;
  final String description;
  final Map<String, double> immediateReward;
  final List<ExpeditionChoice> choices;
  final bool isEnd;
  final String? locationReward;

  ExpeditionEvent({
    required this.eventId,
    required this.description,
    required this.immediateReward,
    required this.choices,
    required this.isEnd,
    this.locationReward,
  });

  factory ExpeditionEvent.fromJson(Map<String, dynamic> json) {
    final rewardJson = json['immediate_reward'] as Map<String, dynamic>? ?? {};
    final immediateReward = <String, double>{};
    rewardJson.forEach((key, value) {
      immediateReward[key] = (value as num).toDouble();
    });

    final choicesJson = json['choices'] as List? ?? [];
    final choices = choicesJson
        .map((c) => ExpeditionChoice.fromJson(c as Map<String, dynamic>))
        .toList();

    return ExpeditionEvent(
      eventId: json['event_id'] as String? ?? '',
      description: json['description'] as String? ?? '',
      immediateReward: immediateReward,
      choices: choices,
      isEnd: json['is_end'] as bool? ?? false,
      locationReward: json['location_reward'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'event_id': eventId,
      'description': description,
      'immediate_reward': immediateReward,
      'choices': choices.map((c) => c.toJson()).toList(),
      'is_end': isEnd,
      'location_reward': locationReward,
    };
  }
}

class ExpeditionChoice {
  final String label;
  final String description;
  final Map<String, double> reward;
  final String nextEventId;

  ExpeditionChoice({
    required this.label,
    required this.description,
    required this.reward,
    required this.nextEventId,
  });

  factory ExpeditionChoice.fromJson(Map<String, dynamic> json) {
    final rewardJson = json['reward'] as Map<String, dynamic>? ?? {};
    final reward = <String, double>{};
    rewardJson.forEach((key, value) {
      reward[key] = (value as num).toDouble();
    });

    return ExpeditionChoice(
      label: json['label'] as String? ?? '',
      description: json['description'] as String? ?? '',
      reward: reward,
      nextEventId: json['next_event_id'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'label': label,
      'description': description,
      'reward': reward,
      'next_event_id': nextEventId,
    };
  }
}

class ExpeditionEventLogEntry {
  final String eventId;
  final String description;
  final int playerChoice;
  final String choiceLabel;
  final Map<String, double> rewardsReceived;
  final DateTime createdAt;

  ExpeditionEventLogEntry({
    required this.eventId,
    required this.description,
    required this.playerChoice,
    required this.choiceLabel,
    required this.rewardsReceived,
    required this.createdAt,
  });

  factory ExpeditionEventLogEntry.fromJson(Map<String, dynamic> json) {
    final rewardsJson = json['rewards_received'] as Map<String, dynamic>? ?? {};
    final rewardsReceived = <String, double>{};
    rewardsJson.forEach((key, value) {
      rewardsReceived[key] = (value as num).toDouble();
    });

    return ExpeditionEventLogEntry(
      eventId: json['event_id'] as String? ?? '',
      description: json['description'] as String? ?? '',
      playerChoice: json['player_choice'] as int? ?? -1,
      choiceLabel: json['choice_label'] as String? ?? '',
      rewardsReceived: rewardsReceived,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now().toUtc(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'event_id': eventId,
      'description': description,
      'player_choice': playerChoice,
      'choice_label': choiceLabel,
      'rewards_received': rewardsReceived,
      'created_at': createdAt.toIso8601String(),
    };
  }
}

class ExpeditionChoiceResult {
  final ExpeditionEvent? event;
  final ExpeditionChain chain;
  final Map<String, double> inventory;
  final bool completed;
  final bool failed;
  final Location? location;
  final String? locationReward;
  final String? error;

  ExpeditionChoiceResult({
    this.event,
    required this.chain,
    required this.inventory,
    required this.completed,
    required this.failed,
    this.location,
    this.locationReward,
    this.error,
  });

  factory ExpeditionChoiceResult.fromJson(Map<String, dynamic> json) {
    final inventoryJson = json['inventory'] as Map<String, dynamic>? ?? {};
    final inventory = <String, double>{};
    inventoryJson.forEach((key, value) {
      inventory[key] = (value as num).toDouble();
    });

    return ExpeditionChoiceResult(
      event: json['event'] != null
          ? ExpeditionEvent.fromJson(json['event'] as Map<String, dynamic>)
          : null,
      chain: ExpeditionChain.fromJson(json['chain'] as Map<String, dynamic>),
      inventory: inventory,
      completed: json['completed'] as bool? ?? false,
      failed: json['failed'] as bool? ?? false,
      location: json['location'] != null
          ? Location.fromJson(json['location'] as Map<String, dynamic>)
          : null,
      locationReward: json['location_reward'] as String?,
      error: json['error'] as String?,
    );
  }
}

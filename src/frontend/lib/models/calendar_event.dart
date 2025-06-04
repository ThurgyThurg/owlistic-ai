class CalendarEvent {
  final String id;
  final String title;
  final String description;
  final DateTime date;
  final String? googleEventId;
  final String userId;
  final DateTime createdAt;
  final DateTime updatedAt;

  CalendarEvent({
    required this.id,
    required this.title,
    required this.description,
    required this.date,
    this.googleEventId,
    required this.userId,
    required this.createdAt,
    required this.updatedAt,
  });

  factory CalendarEvent.fromJson(Map<String, dynamic> json) {
    return CalendarEvent(
      id: json['id'],
      title: json['title'],
      description: json['description'] ?? '',
      date: DateTime.parse(json['date']),
      googleEventId: json['google_event_id'],
      userId: json['user_id'],
      createdAt: DateTime.parse(json['created_at']),
      updatedAt: DateTime.parse(json['updated_at']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'description': description,
      'date': date.toIso8601String(),
      'google_event_id': googleEventId,
      'user_id': userId,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  CalendarEvent copyWith({
    String? id,
    String? title,
    String? description,
    DateTime? date,
    String? googleEventId,
    String? userId,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return CalendarEvent(
      id: id ?? this.id,
      title: title ?? this.title,
      description: description ?? this.description,
      date: date ?? this.date,
      googleEventId: googleEventId ?? this.googleEventId,
      userId: userId ?? this.userId,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}
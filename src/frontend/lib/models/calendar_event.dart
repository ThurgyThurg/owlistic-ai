class CalendarEvent {
  final String id;
  final String title;
  final String description;
  final DateTime startTime;
  final DateTime endTime;
  final bool allDay;
  final String? location;
  final String? timeZone;
  final String status;
  final String visibility;
  final String? recurrence;
  final String source;
  final String? googleEventId;
  final String? googleCalendarId;
  final String? noteId;
  final String? taskId;
  final String userId;
  final Map<String, dynamic>? metadata;
  final DateTime createdAt;
  final DateTime updatedAt;

  CalendarEvent({
    required this.id,
    required this.title,
    required this.description,
    required this.startTime,
    required this.endTime,
    this.allDay = false,
    this.location,
    this.timeZone,
    this.status = 'confirmed',
    this.visibility = 'default',
    this.recurrence,
    this.source = 'owlistic',
    this.googleEventId,
    this.googleCalendarId,
    this.noteId,
    this.taskId,
    required this.userId,
    this.metadata,
    required this.createdAt,
    required this.updatedAt,
  });

  // Convenience getter for backward compatibility
  DateTime get date => startTime;

  factory CalendarEvent.fromJson(Map<String, dynamic> json) {
    return CalendarEvent(
      id: json['id'],
      title: json['title'],
      description: json['description'] ?? '',
      startTime: DateTime.parse(json['start_time']),
      endTime: DateTime.parse(json['end_time']),
      allDay: json['all_day'] ?? false,
      location: json['location'],
      timeZone: json['time_zone'],
      status: json['status'] ?? 'confirmed',
      visibility: json['visibility'] ?? 'default',
      recurrence: json['recurrence'],
      source: json['source'] ?? 'owlistic',
      googleEventId: json['google_event_id'],
      googleCalendarId: json['google_calendar_id'],
      noteId: json['note_id'],
      taskId: json['task_id'],
      userId: json['user_id'],
      metadata: json['metadata'] as Map<String, dynamic>?,
      createdAt: DateTime.parse(json['created_at']),
      updatedAt: DateTime.parse(json['updated_at']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'description': description,
      'start_time': startTime.toIso8601String(),
      'end_time': endTime.toIso8601String(),
      'all_day': allDay,
      'location': location,
      'time_zone': timeZone,
      'status': status,
      'visibility': visibility,
      'recurrence': recurrence,
      'source': source,
      'google_event_id': googleEventId,
      'google_calendar_id': googleCalendarId,
      'note_id': noteId,
      'task_id': taskId,
      'user_id': userId,
      'metadata': metadata,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  CalendarEvent copyWith({
    String? id,
    String? title,
    String? description,
    DateTime? startTime,
    DateTime? endTime,
    bool? allDay,
    String? location,
    String? timeZone,
    String? status,
    String? visibility,
    String? recurrence,
    String? source,
    String? googleEventId,
    String? googleCalendarId,
    String? noteId,
    String? taskId,
    String? userId,
    Map<String, dynamic>? metadata,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return CalendarEvent(
      id: id ?? this.id,
      title: title ?? this.title,
      description: description ?? this.description,
      startTime: startTime ?? this.startTime,
      endTime: endTime ?? this.endTime,
      allDay: allDay ?? this.allDay,
      location: location ?? this.location,
      timeZone: timeZone ?? this.timeZone,
      status: status ?? this.status,
      visibility: visibility ?? this.visibility,
      recurrence: recurrence ?? this.recurrence,
      source: source ?? this.source,
      googleEventId: googleEventId ?? this.googleEventId,
      googleCalendarId: googleCalendarId ?? this.googleCalendarId,
      noteId: noteId ?? this.noteId,
      taskId: taskId ?? this.taskId,
      userId: userId ?? this.userId,
      metadata: metadata ?? this.metadata,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}
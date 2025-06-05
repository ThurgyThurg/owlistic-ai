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
      id: json['id']?.toString() ?? '',
      title: json['title']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      startTime: json['start_time'] != null ? DateTime.parse(json['start_time']) : DateTime.now(),
      endTime: json['end_time'] != null ? DateTime.parse(json['end_time']) : DateTime.now().add(const Duration(hours: 1)),
      allDay: json['all_day'] ?? false,
      location: json['location']?.toString(),
      timeZone: json['time_zone']?.toString(),
      status: json['status']?.toString() ?? 'confirmed',
      visibility: json['visibility']?.toString() ?? 'default',
      recurrence: json['recurrence']?.toString(),
      source: json['source']?.toString() ?? 'owlistic',
      googleEventId: json['google_event_id']?.toString(),
      googleCalendarId: json['google_calendar_id']?.toString(),
      noteId: json['note_id']?.toString(),
      taskId: json['task_id']?.toString(),
      userId: json['user_id']?.toString() ?? '',
      metadata: json['metadata'] as Map<String, dynamic>?,
      createdAt: json['created_at'] != null ? DateTime.parse(json['created_at']) : DateTime.now(),
      updatedAt: json['updated_at'] != null ? DateTime.parse(json['updated_at']) : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'description': description,
      'start_time': _formatRFC3339(startTime),
      'end_time': _formatRFC3339(endTime),
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
      'created_at': _formatRFC3339(createdAt),
      'updated_at': _formatRFC3339(updatedAt),
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

  // Helper method to ensure proper RFC3339 format with timezone
  String _formatRFC3339(DateTime dateTime) {
    final utcTime = dateTime.toUtc();
    final iso8601 = utcTime.toIso8601String();
    // Ensure the string ends with 'Z' for UTC timezone
    return iso8601.endsWith('Z') ? iso8601 : '${iso8601}Z';
  }
}
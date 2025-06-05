import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/calendar_event.dart';
import '../services/calendar_service.dart';

final calendarServiceProvider = Provider<CalendarService>((ref) {
  return CalendarService();
});

final calendarProvider = StateNotifierProvider<CalendarNotifier, AsyncValue<List<CalendarEvent>>>((ref) {
  final service = ref.watch(calendarServiceProvider);
  return CalendarNotifier(service);
});

final googleCalendarConnectedProvider = FutureProvider<bool>((ref) async {
  final calendarService = ref.watch(calendarServiceProvider);
  try {
    return await calendarService.isGoogleCalendarConnected();
  } catch (e) {
    return false;
  }
});

class CalendarNotifier extends StateNotifier<AsyncValue<List<CalendarEvent>>> {
  final CalendarService _service;
  
  CalendarNotifier(this._service) : super(const AsyncValue.loading());
  
  Future<void> fetchEvents(DateTime month) async {
    try {
      state = const AsyncValue.loading();
      final events = await _service.getEvents(month);
      state = AsyncValue.data(events);
    } catch (e, stack) {
      state = AsyncValue.error(e, stack);
    }
  }
  
  Future<void> createEvent({
    required String title,
    required String description,
    required DateTime startTime,
    required DateTime endTime,
    bool allDay = false,
    String? location,
    String? timeZone,
    String? calendarId,
    String? noteId,
    String? taskId,
  }) async {
    try {
      await _service.createEvent(
        title: title,
        description: description,
        startTime: startTime,
        endTime: endTime,
        allDay: allDay,
        location: location,
        timeZone: timeZone,
        calendarId: calendarId,
        noteId: noteId,
        taskId: taskId,
      );
      // Refresh events
      await fetchEvents(startTime);
    } catch (e, stack) {
      state = AsyncValue.error(e, stack);
    }
  }
  
  Future<void> updateEvent(CalendarEvent event) async {
    try {
      await _service.updateEvent(event);
      // Refresh events
      await fetchEvents(event.startTime);
    } catch (e, stack) {
      state = AsyncValue.error(e, stack);
    }
  }
  
  Future<void> deleteEvent(String eventId) async {
    try {
      await _service.deleteEvent(eventId);
      // Update local state
      state.whenData((events) {
        state = AsyncValue.data(
          events.where((e) => e.id != eventId).toList(),
        );
      });
    } catch (e, stack) {
      state = AsyncValue.error(e, stack);
    }
  }
  
  Future<void> syncWithGoogle() async {
    try {
      await _service.syncWithGoogle();
      // Refresh events after sync
      await fetchEvents(DateTime.now());
    } catch (e, stack) {
      state = AsyncValue.error(e, stack);
    }
  }
}
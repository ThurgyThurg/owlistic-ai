import 'dart:convert';
import '../models/calendar_event.dart';
import '../utils/logger.dart';
import './base_service.dart';

class CalendarService extends BaseService {
  final Logger logger = Logger('CalendarService');

  Future<List<CalendarEvent>> getEvents(DateTime month) async {
    try {
      final startOfMonth = DateTime(month.year, month.month, 1);
      final endOfMonth = DateTime(month.year, month.month + 1, 0);
      
      final response = await authenticatedGet(
        '/api/calendar/events',
        queryParameters: {
          'start': startOfMonth.toIso8601String(),
          'end': endOfMonth.toIso8601String(),
        },
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = json.decode(response.body);
        return data.map((e) => CalendarEvent.fromJson(e)).toList();
      } else {
        throw Exception('Failed to load calendar events');
      }
    } catch (e) {
      logger.error('Error fetching calendar events: $e');
      throw e;
    }
  }

  Future<CalendarEvent> createEvent({
    required String title,
    required String description,
    required DateTime date,
  }) async {
    try {
      final response = await authenticatedPost(
        '/api/calendar/events',
        {
          'title': title,
          'description': description,
          'date': date.toIso8601String(),
        },
      );

      if (response.statusCode == 201) {
        return CalendarEvent.fromJson(json.decode(response.body));
      } else {
        throw Exception('Failed to create event');
      }
    } catch (e) {
      logger.error('Error creating event: $e');
      throw e;
    }
  }

  Future<CalendarEvent> updateEvent(CalendarEvent event) async {
    try {
      final response = await authenticatedPut(
        '/api/calendar/events/${event.id}',
        {
          'title': event.title,
          'description': event.description,
          'date': event.date.toIso8601String(),
        },
      );

      if (response.statusCode == 200) {
        return CalendarEvent.fromJson(json.decode(response.body));
      } else {
        throw Exception('Failed to update event');
      }
    } catch (e) {
      logger.error('Error updating event: $e');
      throw e;
    }
  }

  Future<void> deleteEvent(String eventId) async {
    try {
      final response = await authenticatedDelete('/api/calendar/events/$eventId');

      if (response.statusCode != 200) {
        throw Exception('Failed to delete event');
      }
    } catch (e) {
      logger.error('Error deleting event: $e');
      throw e;
    }
  }

  Future<String> getGoogleAuthUrl() async {
    try {
      final response = await authenticatedGet('/api/calendar/google/auth-url');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return data['auth_url'];
      } else {
        throw Exception('Failed to get Google auth URL');
      }
    } catch (e) {
      logger.error('Error getting Google auth URL: $e');
      throw e;
    }
  }

  Future<void> connectGoogleCalendar(String authCode) async {
    try {
      final response = await authenticatedPost(
        '/api/calendar/google/connect',
        {
          'auth_code': authCode,
        },
      );

      if (response.statusCode != 200) {
        throw Exception('Failed to connect Google Calendar');
      }
    } catch (e) {
      logger.error('Error connecting Google Calendar: $e');
      throw e;
    }
  }

  Future<void> disconnectGoogleCalendar() async {
    try {
      final response = await authenticatedPost('/api/calendar/google/disconnect', {});

      if (response.statusCode != 200) {
        throw Exception('Failed to disconnect Google Calendar');
      }
    } catch (e) {
      logger.error('Error disconnecting Google Calendar: $e');
      throw e;
    }
  }

  Future<void> syncWithGoogle() async {
    try {
      final response = await authenticatedPost('/api/calendar/google/sync', {});

      if (response.statusCode != 200) {
        throw Exception('Failed to sync with Google Calendar');
      }
    } catch (e) {
      logger.error('Error syncing with Google Calendar: $e');
      throw e;
    }
  }

  Future<bool> isGoogleCalendarConnected() async {
    try {
      final response = await authenticatedGet('/api/calendar/google/status');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return data['connected'] ?? false;
      } else {
        return false;
      }
    } catch (e) {
      logger.error('Error checking Google Calendar connection: $e');
      return false;
    }
  }
}
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
        '/api/v1/calendar/events',
        queryParameters: {
          'start': startOfMonth.toIso8601String(),
          'end': endOfMonth.toIso8601String(),
        },
      );

      if (response.statusCode == 200) {
        final Map<String, dynamic> responseData = json.decode(response.body);
        final List<dynamic> events = responseData['events'] ?? [];
        return events.map((e) {
          try {
            return CalendarEvent.fromJson(e as Map<String, dynamic>);
          } catch (error) {
            logger.error('Error parsing calendar event: $e, error: $error');
            // Return a placeholder event or skip this event
            rethrow;
          }
        }).toList();
      } else {
        throw Exception('Failed to load calendar events: ${response.statusCode}');
      }
    } catch (e) {
      logger.error('Error fetching calendar events: $e');
      throw e;
    }
  }

  Future<CalendarEvent> createEvent({
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
      final response = await authenticatedPost(
        '/api/v1/calendar/events',
        {
          'title': title,
          'description': description,
          'start_time': startTime.toIso8601String(),
          'end_time': endTime.toIso8601String(),
          'all_day': allDay,
          'location': location,
          'time_zone': timeZone,
          'calendar_id': calendarId,
          'note_id': noteId,
          'task_id': taskId,
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
        '/api/v1/calendar/events/${event.id}',
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
      final response = await authenticatedDelete('/api/v1/calendar/events/$eventId');

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
      final response = await authenticatedGet('/api/v1/calendar/oauth/authorize');

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
      // Note: The OAuth callback is handled automatically by the backend
      // This method is kept for compatibility but may not be needed
      // The actual connection happens when the user visits the auth URL
      // and Google redirects to the backend callback
      throw Exception('OAuth flow is handled automatically via callback URL. Please use the auth URL directly.');
    } catch (e) {
      logger.error('Error connecting Google Calendar: $e');
      throw e;
    }
  }

  Future<void> disconnectGoogleCalendar() async {
    try {
      final response = await authenticatedDelete('/api/v1/calendar/oauth/revoke');

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
      final response = await authenticatedPost('/api/v1/calendar/sync', {});

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
      final response = await authenticatedGet('/api/v1/calendar/oauth/status');

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        return data['has_access'] ?? false;
      } else {
        return false;
      }
    } catch (e) {
      logger.error('Error checking Google Calendar connection: $e');
      return false;
    }
  }

  Future<Map<String, dynamic>> getOAuthConfig() async {
    try {
      final response = await get('/api/v1/calendar/oauth/config');

      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('Failed to get OAuth config');
      }
    } catch (e) {
      logger.error('Error getting OAuth config: $e');
      throw e;
    }
  }
}
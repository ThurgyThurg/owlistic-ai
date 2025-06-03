import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../utils/logger.dart';

class SettingsViewModel extends ChangeNotifier {
  final Logger _logger = Logger('SettingsViewModel');
  
  bool _isLoading = false;
  bool _notificationsEnabled = true;
  bool _noteReminders = true;
  bool _taskReminders = true;
  bool _aiInsights = true;
  String? _errorMessage;
  
  // Getters
  bool get isLoading => _isLoading;
  bool get notificationsEnabled => _notificationsEnabled;
  bool get noteReminders => _noteReminders;
  bool get taskReminders => _taskReminders;
  bool get aiInsights => _aiInsights;
  String? get errorMessage => _errorMessage;
  
  SettingsViewModel() {
    _loadUserPreferences();
  }
  
  Future<void> _loadUserPreferences() async {
    try {
      _isLoading = true;
      _errorMessage = null;
      notifyListeners();
      
      final prefs = await SharedPreferences.getInstance();
      _notificationsEnabled = prefs.getBool('notifications_enabled') ?? true;
      _noteReminders = prefs.getBool('note_reminders') ?? true;
      _taskReminders = prefs.getBool('task_reminders') ?? true;
      _aiInsights = prefs.getBool('ai_insights') ?? true;
      
      _logger.info('User preferences loaded successfully');
    } catch (e) {
      _logger.error('Failed to load user preferences', e);
      _errorMessage = 'Failed to load preferences';
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }
  
  Future<void> toggleNotifications(bool value) async {
    try {
      _notificationsEnabled = value;
      notifyListeners();
      
      // If turning off notifications, disable all sub-options
      if (!value) {
        _noteReminders = false;
        _taskReminders = false;
        _aiInsights = false;
      }
      
      await _updatePreferences();
    } catch (e) {
      _logger.error('Failed to toggle notifications', e);
      _errorMessage = 'Failed to update notification settings';
      // Revert the change
      _notificationsEnabled = !value;
      notifyListeners();
    }
  }
  
  Future<void> toggleNoteReminders(bool value) async {
    try {
      _noteReminders = value;
      notifyListeners();
      await _updatePreferences();
    } catch (e) {
      _logger.error('Failed to toggle note reminders', e);
      _errorMessage = 'Failed to update settings';
      _noteReminders = !value;
      notifyListeners();
    }
  }
  
  Future<void> toggleTaskReminders(bool value) async {
    try {
      _taskReminders = value;
      notifyListeners();
      await _updatePreferences();
    } catch (e) {
      _logger.error('Failed to toggle task reminders', e);
      _errorMessage = 'Failed to update settings';
      _taskReminders = !value;
      notifyListeners();
    }
  }
  
  Future<void> toggleAIInsights(bool value) async {
    try {
      _aiInsights = value;
      notifyListeners();
      await _updatePreferences();
    } catch (e) {
      _logger.error('Failed to toggle AI insights', e);
      _errorMessage = 'Failed to update settings';
      _aiInsights = !value;
      notifyListeners();
    }
  }
  
  Future<void> _updatePreferences() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool('notifications_enabled', _notificationsEnabled);
      await prefs.setBool('note_reminders', _noteReminders);
      await prefs.setBool('task_reminders', _taskReminders);
      await prefs.setBool('ai_insights', _aiInsights);
      
      _logger.info('Preferences updated successfully');
    } catch (e) {
      _logger.error('Failed to update preferences', e);
      throw e;
    }
  }
  
  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }
}
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:owlistic/services/auth_service.dart';
import 'package:owlistic/services/websocket_service.dart';
import 'package:owlistic/viewmodel/login_viewmodel.dart';
import 'package:owlistic/models/user.dart';
import 'package:owlistic/utils/logger.dart';
import 'package:owlistic/config/app_config.dart';

class LoginProvider with ChangeNotifier implements LoginViewModel {
  final Logger _logger = Logger('LoginProvider');
  final AuthService _authService;
  final WebSocketService _webSocketService;
  
  // State
  bool _isLoading = false;
  bool _isActive = false;
  bool _isInitialized = false;
  String? _errorMessage;
  
  // Constructor with dependency injection
  LoginProvider({
    required AuthService authService,
    required WebSocketService webSocketService
  }) : _authService = authService,
       _webSocketService = webSocketService {
    _isInitialized = true;
    _initializeAuthState();
  }
  
  // Initialize auth state and websocket connection on startup
  Future<void> _initializeAuthState() async {
    try {
      _logger.info('Initializing auth state');
      
      // Set server URL from environment configuration
      _webSocketService.setServerUrl(AppConfig.serverUrl);
      
      // Initialize auth service
      await _authService.initialize();
      
      // If user is already logged in, setup websocket connection
      if (_authService.isLoggedIn) {
        _logger.info('User already logged in, setting up websocket connection');
        
        // Get the stored token from auth service
        final token = await _authService.getStoredToken();
        final userId = await _authService.getCurrentUserId();
        
        if (token != null) {
          // Set the token in WebSocketService
          _webSocketService.setAuthToken(token);
          if (userId != null) {
            _webSocketService.setUserId(userId);
          }
          
          // Establish WebSocket connection if not already connected
          if (!_webSocketService.isConnected) {
            await _webSocketService.connect();
          }
          
          _logger.info('WebSocket connection established from stored credentials');
        }
      }
    } catch (e) {
      _logger.error('Error initializing auth state', e);
    }
  }
  
  // LoginViewModel implementation
  @override
  Future<bool> login(String email, String password, bool rememberMe) async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();
    
    try {
      _logger.info('Attempting login for user: $email');
      
      if (rememberMe) {
        await saveEmail(email);
      }
      
      // Make sure auth service is initialized
      await _authService.initialize();
      
      // Call auth service to perform login
      final response = await _authService.login(email, password);
      final success = response['success'] == true;
      
      _logger.info('Login response: $response');
      
      if (success) {
        _logger.info('Login successful for user: $email');
        
        final token = response['token'] as String?;
        final userId = response['userId'] as String?;
        
        // Set auth token and user ID in WebSocketService
        if (token != null) {
          _webSocketService.setAuthToken(token);
        }
        
        if (userId != null) {
          _webSocketService.setUserId(userId);
        }

        // Ensure WebSocket connection is established after successful login
        if (!_webSocketService.isConnected) {
          await _webSocketService.connect();
        }
      } else {
        _errorMessage = "Authentication failed";
        _logger.error('Login unsuccessful: Authentication failed');
      }
      
      _isLoading = false;
      notifyListeners();
      
      return success;
    } catch (e) {
      _logger.error('Login error', e);
      _errorMessage = e.toString();
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }
  
  @override
  Future<String?> getSavedEmail() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      return prefs.getString('last_login_email');
    } catch (e) {
      _logger.error('Error getting saved email', e);
      return null;
    }
  }
  
  @override
  Future<void> saveEmail(String email) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('last_login_email', email);
    } catch (e) {
      _logger.error('Error saving email', e);
    }
  }
  
  // Server URL is now managed via environment configuration
  @override
  Future<void> saveServerUrl(String url) async {
    _logger.info('Server URL is configured via environment - ignoring save request');
    // No-op: Server URL is now configured via environment variables
  }
  
  // Get the current server URL from configuration
  @override
  String? getServerUrl() {
    return AppConfig.serverUrl;
  }
  
  @override
  Future<void> clearSavedEmail() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove('last_login_email');
    } catch (e) {
      _logger.error('Error clearing saved email', e);
    }
  }
  
  // Check WebSocket connection status
  bool get isConnected => _webSocketService.isConnected;
  
  @override
  bool get isLoggingIn => _isLoading;
  
  @override
  bool get isLoggedIn => _authService.isLoggedIn;
  
  @override
  Future<User?> get currentUser async {
    try {
      final token = await _authService.getStoredToken();
      if (token != null) {
        return await _authService.getCurrentUser();
      }
    } catch (e) {
      _logger.error('Error getting current user', e);
    }
    return null;
  }
  
  @override
  Stream<bool> get authStateChanges => _authService.authStateChanges;
  
  // navigateToRegister removed - single user system
  
  @override
  void navigateAfterSuccessfulLogin(BuildContext context) {
    _logger.info('Navigating after successful login');
    context.go('/'); // Navigate to home screen
  }
  
  @override
  void onLoginSuccess(BuildContext context) {
    _logger.info('Login successful, performing post-login actions');
    // Navigate to home screen
    navigateAfterSuccessfulLogin(context);
  }
  
  // BaseViewModel implementation
  @override
  bool get isLoading => _isLoading;
  
  @override
  bool get isInitialized => _isInitialized;
  
  @override
  bool get isActive => _isActive;
  
  @override
  String? get errorMessage => _errorMessage;
  
  @override
  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }
  
  @override
  void activate() {
    _isActive = true;
    _logger.info('LoginProvider activated');
    notifyListeners();
  }
  
  @override
  void deactivate() {
    _isActive = false;
    _logger.info('LoginProvider deactivated');
    notifyListeners();
  }
  
  @override
  void resetState() {
    _errorMessage = null;
    notifyListeners();
  }
}

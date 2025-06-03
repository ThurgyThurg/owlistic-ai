import 'dart:convert';
import 'dart:async';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:owlistic/models/user.dart';
import 'package:owlistic/utils/logger.dart';
import 'package:owlistic/config/app_config.dart';
import 'base_service.dart';

class AuthService extends BaseService {
  final Logger _logger = Logger('AuthService');
  static const String tokenKey = 'auth_token';
  final FlutterSecureStorage _secureStorage = const FlutterSecureStorage();
  
  // Stream controller for auth state changes
  final StreamController<bool> _authStateController = StreamController<bool>.broadcast();
  Stream<bool> get authStateChanges => _authStateController.stream;
  
  // Token management centralized in AuthService
  // Using instance property for local reference and a static for global access
  String? _token;
  static String? get token => _instance?._token;
  
  // Singleton pattern for global token access
  static AuthService? _instance;
  // Initialize synchronously - guaranteed to be true after constructor
  final bool _isInitialized = true;
  bool get isInitialized => true; // Always return true for safety
  
  // Constructor sets the instance for static access
  AuthService() {
    _instance = this;
    
    // For single-user mode, automatically set up authentication
    _initializeSingleUserMode();
    
    // Always mark initialized to avoid initialization errors
    _logger.info('AuthService initialized in single-user mode');
  }
  
  // Initialize single-user mode - skip authentication entirely
  void _initializeSingleUserMode() {
    try {
      // For single-user mode, just set a mock token and mark as authenticated
      _token = 'single-user-token';
      BaseService.setAuthToken(_token);
      _authStateController.add(true);
      _logger.info('Single-user mode initialized - authentication bypassed');
      
      // Immediately configure the server URL
      _configureServerUrl();
    } catch (e) {
      _logger.error('Error initializing single-user mode', e);
      // Fallback to mock token for development
      _token = 'fallback-token';
      BaseService.setAuthToken(_token);
      _authStateController.add(true);
    }
  }

  // Auto-login with single-user credentials
  Future<void> _autoLoginSingleUser() async {
    try {
      // First check if we already have a valid token stored
      final storedToken = await getStoredToken();
      if (storedToken != null && storedToken.isNotEmpty && storedToken != 'fallback-token') {
        _logger.info('Found existing valid token, skipping auto-login');
        return;
      }
      
      // Check if single-user credentials are configured
      if (!AppConfig.hasSingleUserCredentials) {
        _logger.error('Single-user credentials not configured. Please build with --dart-define=USER_EMAIL=your-email --dart-define=USER_PASSWORD=your-password');
        _logger.info('Or configure credentials in your Docker/build environment');
        // Skip auto-login if no credentials configured
        return;
      }
      
      // Use the single-user credentials from environment/config
      final email = AppConfig.userEmail!;
      final password = AppConfig.userPassword!;
      
      _logger.info('Attempting auto-login with configured single-user credentials');
      final result = await login(email, password);
      if (result['success'] == true) {
        _logger.info('Single-user auto-login successful');
        // Store the user ID from the login response
        if (result['userId'] != null) {
          final prefs = await SharedPreferences.getInstance();
          await prefs.setString('user_id', result['userId'].toString());
          _logger.info('Stored user ID: ${result['userId']}');
        }
      } else {
        _logger.info('Single-user auto-login failed, credentials may be incorrect');
        _logger.error('Please ensure frontend USER_EMAIL and USER_PASSWORD match backend .env configuration');
      }
    } catch (e) {
      _logger.error('Auto-login failed: $e');
      _logger.error('Please check your credentials configuration');
    }
  }

  // Synchronous loading of token using SharedPreferences
  void _loadTokenSync() {
    try {
      // Try loading from shared prefs first - this is synchronous
      final tokenFromPrefs = _getTokenFromPrefsSync();
      if (tokenFromPrefs != null) {
        _token = tokenFromPrefs;
        BaseService.setAuthToken(_token); // Important fix: set token in BaseService
        _logger.debug('Successfully loaded token synchronously from shared prefs');
        return;
      }
      
      // Fall back to empty token if we can't load one synchronously
      _token = null;
      _logger.debug('No token found in sync storage, defaulting to null token');
      
      // Start async token loading in background
      _loadTokenAsyncInBackground();
    } catch (e) {
      _logger.error('Error loading token synchronously', e);
      _token = null; // Default to no token on error
    }
  }
  
  // Try to get token from SharedPreferences synchronously
  String? _getTokenFromPrefsSync() {
    try {
      // This is actually async but we're setting it up to run in background
      SharedPreferences.getInstance().then((prefs) {
        final token = prefs.getString(tokenKey);
        if (token != null && token.isNotEmpty) {
          _token = token;
          BaseService.setAuthToken(_token); // Important fix: set token in BaseService
          _authStateController.add(true);
          _logger.debug('Token loaded from SharedPreferences in background');
        }
      });
      
      // Return null for now but it will be loaded in background
      return null;
    } catch (e) {
      _logger.error('Error reading token from SharedPreferences', e);
      return null;
    }
  }
  
  // Load token in background as backup
  void _loadTokenAsyncInBackground() {
    _secureStorage.read(key: tokenKey).then((value) {
      if (value != null && value.isNotEmpty) {
        _token = value;
        BaseService.setAuthToken(_token); // Important fix: set token in BaseService
        _authStateController.add(true);
        _logger.debug('Token loaded from secure storage in background');
        
        // Save to SharedPreferences for faster access next time
        SharedPreferences.getInstance().then((prefs) {
          prefs.setString(tokenKey, value);
        });
      }
    }).catchError((e) {
      _logger.error('Error loading token from secure storage', e);
    });
  }
  
  // Configure server URL for single-user mode
  void _configureServerUrl() {
    // Run asynchronously to avoid blocking initialization
    Future.delayed(Duration.zero, () async {
      // Clear any existing URL first
      try {
        final prefs = await SharedPreferences.getInstance();
        await prefs.remove('api_url');
        BaseService.resetCachedUrl();
        _logger.debug('Cleared existing API URL from SharedPreferences');
      } catch (e) {
        _logger.error('Error clearing API URL', e);
      }
      
      await initialize();
    });
  }
  
  // Fixed up initialize method - call explicitly from login/register to ensure token is set
  Future<void> initialize() async {
    _logger.debug('Initializing AuthService explicitly');
    
    // Debug: Print AppConfig values
    _logger.debug('AppConfig debug info: ${AppConfig.debugInfo}');
    
    // Set server URL in SharedPreferences for BaseService from AppConfig
    try {
      final prefs = await SharedPreferences.getInstance();
      final currentUrl = prefs.getString('api_url');
      _logger.debug('Current URL in SharedPreferences: $currentUrl');
      _logger.debug('AppConfig.serverUrl: ${AppConfig.serverUrl}');
      
      // Always use relative URLs to work with nginx proxy in production
      // This avoids CORS issues and works with any domain through cloudflared tunnel
      String serverUrl = '';
      _logger.debug('Using relative URLs for nginx proxy configuration');
      
      await prefs.setString('api_url', serverUrl);
      _logger.debug('Server URL updated in SharedPreferences: $serverUrl');
      print('🔥 AUTH SERVICE: Set API URL to: "$serverUrl" (empty = relative URLs)');
      BaseService.resetCachedUrl(); // Clear cached URL to force reload
    } catch (e) {
      _logger.error('Error setting server URL in SharedPreferences', e);
    }
    
    // Make sure token is loaded into BaseService
    if (_token != null) {
      BaseService.setAuthToken(_token);
      _logger.debug('Auth token set in BaseService: ${_token?.substring(0, 10)}...');
    }
    return Future.value();
  }
  
  bool get isLoggedIn => true; // Always logged in for single-user mode
  
  // Add back getStoredToken method that was accidentally removed
  Future<String?> getStoredToken() async {
    // If we already have a token in memory, just return it
    if (_token != null) return _token;
    
    try {
      // Try SharedPreferences first for speed
      final prefs = await SharedPreferences.getInstance();
      final storedToken = prefs.getString(tokenKey);
      
      if (storedToken != null && storedToken.isNotEmpty) {
        _token = storedToken;
        // Update BaseService token - critical fix
        BaseService.setAuthToken(_token);
        
        _authStateController.add(true);
        _logger.debug('Retrieved token from SharedPreferences');
        return storedToken;
      }
      
      // Fall back to secure storage if not in SharedPreferences
      _token = await _secureStorage.read(key: tokenKey);
      if (_token != null && _token!.isNotEmpty) {
        _logger.debug('Retrieved token from secure storage');
        // Update BaseService token - critical fix
        BaseService.setAuthToken(_token);
        
        _authStateController.add(true);
        
        // Save to SharedPreferences for faster access next time
        prefs.setString(tokenKey, _token!);
      } else {
        _token = null;
        _logger.debug('No token found in storage');
      }
      return _token;
    } catch (e) {
      _logger.error('Error reading token from storage', e);
      return null;
    }
  }

  // Authentication methods
  Future<Map<String, dynamic>> login(String email, String password) async {
    try {
      // Ensure auth service is properly initialized before login
      await initialize();
      
      final response = await createPostRequest(
        '/api/v1/login',
        {
          'email': email,
          'password': password,
        }
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        // Critical fix: Extract token properly
        final token = data['token'] as String?;
        if (token == null || token.isEmpty) {
          throw Exception('No token received from server');
        }
        
        _logger.debug('Login successful, token received');
        await _storeToken(token);
        return {'success': true, 'token': token, 'userId': data['userId'] ?? data['user_id']};
      } else {
        _logger.error('Login failed with status: ${response.statusCode}, body: ${response.body}');
        throw Exception('Failed to login: ${response.body}');
      }
    } catch (e) {
      _logger.error('Error during login', e);
      rethrow;
    }
  }
  
  // Helper method for unauthenticated POST - fixed to handle async URI creation
  Future<http.Response> createPostRequest(String path, dynamic body) async {
    // Properly await the URI creation
    final uri = await createUri(path);
    _logger.debug('Creating unauthenticated POST request to $uri');
    
    return http.post(
      uri,
      headers: getBaseHeaders(),
      body: jsonEncode(body),
    );
  }
  
  Future<bool> register(String email, String password) async {
    try {
      // Not using authenticatedPost here since we don't have a token yet
      final response = await createPostRequest(
        '/api/v1/register',
        {
          'email': email,
          'password': password,
        }
      );

      if (response.statusCode == 201) {
        _logger.info('Registration successful for: $email');
        return true;
      } else {
        _logger.error('Registration failed with status: ${response.statusCode}, body: ${response.body}');
        throw Exception('Registration failed: ${response.body}');
      }
    } catch (e) {
      _logger.error('Error during registration', e);
      rethrow;
    }
  }
  
  Future<bool> logout() async {
    try {
      // Clear token regardless of response
      await clearToken();
      await clearPreferences();
      _logger.info('Logged out successfully');
      return true;
    } catch (e) {
      _logger.error('Error during logout', e);
      await clearToken(); // Still clear token on error
      await clearPreferences();
      return false;
    }
  }
  
  // Token management - store in both secure storage and SharedPreferences
  Future<void> _storeToken(String token) async {
    if (token.isEmpty) return;
    
    _token = token;
    // Update the static token in BaseService
    BaseService.setAuthToken(token);
    
    _logger.debug('Storing auth token');
    
    try {
      // Store in secure storage
      await _secureStorage.write(key: tokenKey, value: token);
      
      // Also store in SharedPreferences for sync access next time
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(tokenKey, token);
      
      _authStateController.add(true);
    } catch (e) {
      _logger.error('Error storing token', e);
      rethrow;
    }
  }
  
  // Clear token from both storages
  Future<void> clearToken() async {
    _token = null;
    // Clear the static token in BaseService
    BaseService.setAuthToken(null);
    
    _logger.debug('Clearing auth token');
    
    try {
      // Clear from secure storage
      await _secureStorage.delete(key: tokenKey);
      
      // Also clear from SharedPreferences
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove(tokenKey);
      
      _authStateController.add(false);
    } catch (e) {
      _logger.error('Error clearing token', e);
    }
  }
  
  // Update token and notify systems
  Future<void> onTokenChanged(String? token) async {
    try {
      if (token == null || token.isEmpty) {
        await clearToken();
        _logger.info('Auth token cleared');
        return;
      }
      
      // Store token
      await _storeToken(token);
      _logger.info('Auth token updated successfully');
      
      // Attempt to fetch user info with new token
      await getUserProfile();
    } catch (e) {
      _logger.error('Error in onTokenChanged', e);
      await clearToken();
      rethrow;
    }
  }
  
  // Get user information from token or API
  Future<User?> getUserProfile() async {
    _token ??= await getStoredToken();
    
    if (_token == null) return null;
    
    try {
      // Extract user info from JWT payload
      final tokenParts = _token!.split('.');
      if (tokenParts.length != 3) {
        _logger.error("Invalid JWT token format");
        return null;
      }
      
      String normalized = base64Url.normalize(tokenParts[1]);
      final payloadJson = utf8.decode(base64Url.decode(normalized));
      final payload = jsonDecode(payloadJson);
      
      return User(
        id: payload['user_id'],
        email: payload['email'] ?? '',
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
      );
    } catch (e) {
      _logger.error('Error getting user profile', e);
      return null;
    }
  }
  
  // Get user info for single-user mode
  Future<User?> getCurrentUser() async {
    try {
      // For single-user mode, fetch the actual user from the backend
      // Use the dedicated current user endpoint
      final response = await http.get(
        Uri.parse('${await _getServerUrl()}/api/v1/users/current'),
        headers: {
          'Content-Type': 'application/json',
          if (_token != null) 'Authorization': 'Bearer $_token',
        },
      );
      
      if (response.statusCode == 200) {
        final userJson = json.decode(response.body) as Map<String, dynamic>;
        return User.fromJson(userJson);
      } else {
        _logger.warning('Failed to fetch current user from backend (${response.statusCode}), using fallback');
      }
    } catch (e) {
      _logger.error('Error getting current user from backend', e);
    }
    
    // Fallback to mock user if API fails or no users found
    return User(
      id: 'single-user',
      email: 'user@example.com',
      username: 'user',
      displayName: 'User',
      createdAt: DateTime.now(),
      updatedAt: DateTime.now(),
      preferences: {},
    );
  }
  

  // For single-user mode, return a consistent user ID
  Future<String?> getCurrentUserId() async {
    return 'single-user';
  }

  Future<void> clearPreferences() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.clear();
    } catch (e) {
      _logger.error('Error clearing preferences', e);
    }
  }
  
  // Safe login with better error handling
  Future<bool> loginSafe(String email, String password) async {
    try {
      final response = await login(email, password);
      return response['success'] == true;
    } catch (e) {
      _logger.error('Login error occurred', e);
      await clearToken();  // Ensure we clean up on error
      return false;
    }
  }
  
  // Clean up resources
  void dispose() {
    if (!_authStateController.isClosed) {
      _authStateController.close();
    }
  }
}

// Custom error class for initialization issues
class NotInitializedError extends Error {
  final String message;
  NotInitializedError(this.message);
  
  @override
  String toString() => message;
}

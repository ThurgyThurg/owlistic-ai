class AppConfig {
  // Server URL is configured via environment variables at build time
  // Set with --dart-define=SERVER_URL=http://your-server:8080
  // Or configure in your .env file and build system
  static const String _serverUrlFromEnv = String.fromEnvironment('SERVER_URL');
  
  // Fallback server URLs - these can be configured based on your setup
  static const String _fallbackServerUrl = String.fromEnvironment('FALLBACK_SERVER_URL', defaultValue: 'http://localhost');
  
  // Single-user credentials - configured via environment variables at build time
  // Set with --dart-define=USER_EMAIL=your-email@example.com --dart-define=USER_PASSWORD=your-password
  // These should match the values in your backend .env file
  static const String _userEmail = String.fromEnvironment('USER_EMAIL');
  static const String _userPassword = String.fromEnvironment('USER_PASSWORD');
  
  // Helper method to get the configured server URL
  static String get serverUrl {
    if (_serverUrlFromEnv.isNotEmpty) {
      return _serverUrlFromEnv;
    }
    if (_fallbackServerUrl.isNotEmpty) {
      return _fallbackServerUrl;
    }
    // Final fallback - should never happen with our setup
    return 'http://localhost';
  }
  
  // Helper methods to get single-user credentials
  static String? get userEmail => _userEmail.isNotEmpty ? _userEmail : null;
  static String? get userPassword => _userPassword.isNotEmpty ? _userPassword : null;
  
  // Check if single-user credentials are configured
  static bool get hasSingleUserCredentials => 
      userEmail != null && userEmail!.isNotEmpty && 
      userPassword != null && userPassword!.isNotEmpty;
  
  // Debug method to check what values we have
  static Map<String, String> get debugInfo => {
    'serverUrlFromEnv': _serverUrlFromEnv,
    'fallbackServerUrl': _fallbackServerUrl,
    'finalServerUrl': serverUrl,
    'hasUserEmail': (_userEmail.isNotEmpty).toString(),
    'hasUserPassword': (_userPassword.isNotEmpty).toString(),
  };
}
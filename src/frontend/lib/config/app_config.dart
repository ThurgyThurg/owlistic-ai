class AppConfig {
  // Server URL is configured via environment variables at build time
  // Set with --dart-define=SERVER_URL=http://your-server:8080
  // Or configure in your .env file and build system
  static const String _serverUrlFromEnv = String.fromEnvironment('SERVER_URL');
  
  // Fallback server URLs - these can be configured based on your setup
  static const String _fallbackServerUrl = String.fromEnvironment('FALLBACK_SERVER_URL', defaultValue: 'http://localhost');
  
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
  
  // Debug method to check what values we have
  static Map<String, String> get debugInfo => {
    'serverUrlFromEnv': _serverUrlFromEnv,
    'fallbackServerUrl': _fallbackServerUrl,
    'finalServerUrl': serverUrl,
  };
}
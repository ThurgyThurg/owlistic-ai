class AppConfig {
  // Server URL is configured via environment variables at build time
  // Set with --dart-define=SERVER_URL=http://your-server:8080
  // Or configure in your .env file and build system
  static const String defaultServerUrl = String.fromEnvironment('SERVER_URL');
  
  // Helper method to get the configured server URL
  static String get serverUrl => defaultServerUrl;
}
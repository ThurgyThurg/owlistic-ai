class AppConfig {
  // Server URL can be set at build time with --dart-define=SERVER_URL=http://your-server:8080
  static const String defaultServerUrl = String.fromEnvironment(
    'SERVER_URL',
    defaultValue: 'http://localhost:8080'
  );
}
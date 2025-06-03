# Server URL Configuration

The server URL is now configured via environment variables instead of manual input during login.

## Configuration Methods

### 1. Build-time Configuration
Set the server URL when building the Flutter app:

```bash
flutter build web --dart-define=SERVER_URL=http://your-server:8080
flutter run --dart-define=SERVER_URL=http://your-server:8080
```

### 2. Default Configuration
If no SERVER_URL is provided, the app defaults to `http://localhost:8080`.

### 3. Environment-based Configuration
For production deployments, you can set up your build pipeline to use environment variables:

```bash
export SERVER_URL=https://api.yourapp.com
flutter build web --dart-define=SERVER_URL=$SERVER_URL
```

## Changes Made

- **Removed**: Manual server URL input field from login screen
- **Updated**: `AppConfig` to use environment-configured URL
- **Modified**: `LoginProvider` to use configured URL instead of user input
- **Simplified**: Authentication flow without URL management

## Migration

If you were previously using a custom server URL:
1. Note your current server URL
2. Use the `--dart-define=SERVER_URL=your_url` flag when running/building the app
3. The login screen will no longer show the server URL field

## Benefits

- **Security**: Prevents users from accidentally connecting to wrong servers
- **Consistency**: Ensures all users connect to the intended backend
- **Simplicity**: Cleaner login interface
- **DevOps**: Better integration with CI/CD pipelines
# Network Access Setup Guide

This guide helps you configure Owlistic (with AI features) to be accessible from other computers on your local network.

## 🔧 Backend Configuration (Already Done)

The backend server has been configured to:
- Listen on `0.0.0.0:8080` instead of `localhost:8080`
- Accept connections from any device on your local network
- Display network access information on startup

## 🌐 Network Setup Steps

### 1. Find Your Computer's IP Address

**On Linux/macOS:**
```bash
# Find your local IP address
ip addr show | grep "inet " | grep -v 127.0.0.1
# OR
ifconfig | grep "inet " | grep -v 127.0.0.1
```

**On Windows:**
```cmd
ipconfig | find "IPv4"
```

Example output: `192.168.1.100` (your IP will be different)

### 2. Configure Firewall

**On Linux (Ubuntu/Debian):**
```bash
# Allow port 8080 through firewall
sudo ufw allow 8080
sudo ufw reload

# Check status
sudo ufw status
```

**On Windows:**
```
1. Go to Windows Defender Firewall
2. Click "Advanced settings"
3. Click "Inbound Rules" → "New Rule"
4. Select "Port" → Next
5. Select "TCP" and enter port "8080" → Next
6. Select "Allow the connection" → Next
7. Apply to all profiles → Next
8. Name it "Owlistic API" → Finish
```

**On macOS:**
```bash
# Usually no action needed, but if blocked:
# System Preferences → Security & Privacy → Firewall → Firewall Options
# Allow incoming connections for your app
```

### 3. Start the Backend

```bash
cd src/backend

# Set your AI API keys
export ANTHROPIC_API_KEY=your_anthropic_key
export OPENAI_API_KEY=your_openai_key

# Optional: Set custom port (default is 8080)
export APP_PORT=8080

# Start the server
go run cmd/main.go
```

You should see output like:
```
API server is running on http://0.0.0.0:8080
Access from other devices: http://YOUR_COMPUTER_IP:8080
```

### 4. Test Network Access

From another device on your network, test the API:

```bash
# Replace 192.168.1.100 with your actual IP
curl http://192.168.1.100:8080/api/v1/health

# Should return: {"status":"ok"}
```

## 📱 Flutter App Configuration

Your Flutter app is already configured to handle network URLs through the `base_service.dart` system.

### Option 1: Configure via Settings Screen

1. Open your Flutter app
2. Go to Settings (if you have a settings screen)
3. Set API URL to: `http://192.168.1.100:8080` (replace with your IP)

### Option 2: Configure Programmatically

Add this to your app startup (e.g., in `main.dart`):

```dart
import 'package:shared_preferences/shared_preferences.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  // Set the API URL for network access
  final prefs = await SharedPreferences.getInstance();
  await prefs.setString('api_url', 'http://192.168.1.100:8080'); // Replace with your IP
  
  runApp(MyApp());
}
```

### Option 3: Environment-Based Configuration

Create different configurations for development:

```dart
// In lib/config/api_config.dart
class ApiConfig {
  static const String localUrl = 'http://localhost:8080';
  static const String networkUrl = 'http://192.168.1.100:8080'; // Replace with your IP
  
  static String get baseUrl {
    // You can switch between local and network URLs
    return networkUrl; // or localUrl for local development
  }
}
```

## 🔗 Full Network URLs

Once configured, these URLs will be accessible from any device on your network:

**API Base:** `http://192.168.1.100:8080/api/v1/`

**AI Endpoints:**
- Process note: `POST http://192.168.1.100:8080/api/v1/ai/notes/{id}/process`
- Semantic search: `POST http://192.168.1.100:8080/api/v1/ai/notes/search/semantic`
- AI chat: `POST http://192.168.1.100:8080/api/v1/ai/chat`

**Regular Endpoints:**
- Notes: `GET http://192.168.1.100:8080/api/v1/notes`
- Notebooks: `GET http://192.168.1.100:8080/api/v1/notebooks`
- Tasks: `GET http://192.168.1.100:8080/api/v1/tasks`

## 🚨 Troubleshooting

### Backend Not Accessible

1. **Check if server is running:**
   ```bash
   netstat -tlnp | grep :8080
   # Should show: tcp6       0      0 :::8080                 :::*                    LISTEN
   ```

2. **Test local access first:**
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

3. **Check firewall:**
   ```bash
   sudo ufw status
   # Should show: 8080                       ALLOW       Anywhere
   ```

4. **Check the IP address:**
   ```bash
   ip route get 1.1.1.1 | grep -oP 'src \K\S+'
   # This shows your primary IP address
   ```

### Flutter App Connection Issues

1. **Verify API URL in SharedPreferences:**
   ```dart
   final prefs = await SharedPreferences.getInstance();
   final apiUrl = prefs.getString('api_url');
   print('Current API URL: $apiUrl');
   ```

2. **Test with manual HTTP request:**
   ```dart
   import 'package:http/http.dart' as http;
   
   final response = await http.get(
     Uri.parse('http://192.168.1.100:8080/api/v1/health')
   );
   print('Response: ${response.statusCode} ${response.body}');
   ```

3. **Check network connectivity:**
   - Ensure both devices are on the same WiFi network
   - Try pinging the server from client device

### Common Network Issues

**"Connection refused":**
- Server not running or firewall blocking
- Wrong IP address or port

**"Connection timeout":**
- Devices on different networks
- Router blocking internal communication

**"Host unreachable":**
- IP address changed (routers often reassign IPs)
- WiFi vs Ethernet different subnets

## 🔒 Security Considerations

**For Local Network Use:**
- This setup is safe for local networks
- Don't expose to the internet without additional security
- Consider using HTTPS in production

**For Internet Access:**
- Use a reverse proxy (nginx)
- Set up SSL/TLS certificates
- Configure proper CORS policies
- Add rate limiting

## 📝 Quick Reference

**Server Computer (where backend runs):**
```bash
# Get IP address
hostname -I | awk '{print $1}'

# Start server
cd src/backend && go run cmd/main.go
```

**Client Devices (Flutter app):**
```
API URL: http://[SERVER_IP]:8080
Example: http://192.168.1.100:8080
```

**Test Command:**
```bash
curl http://[SERVER_IP]:8080/api/v1/health
```

Replace `[SERVER_IP]` with your actual server IP address in all examples.

## 🎯 Next Steps

1. Start the backend server
2. Note the IP address from the startup message
3. Configure your Flutter app with the network URL
4. Test the connection
5. Start using AI features across your network!

Your Owlistic AI Second Brain is now accessible from any device on your local network! 🦉🧠
import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:owlistic/utils/logger.dart';
import 'package:owlistic/models/subscription.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:owlistic/config/app_config.dart';

class WebSocketService {
  static final WebSocketService _instance = WebSocketService._internal();
  final Logger _logger = Logger('WebSocketService');
  
  // WebSocket channel and streams
  WebSocketChannel? _channel;
  String? _authToken;
  String? _userId;
  String? _serverUrl;
  bool _isConnected = false;
  bool _isConnecting = false;
  bool _isInitialized = false; // Track initialization state
  
  // Subscription tracking
  final Set<String> _confirmedSubscriptions = {};
  final Set<String> _pendingSubscriptions = {};
  final Map<String, DateTime> _lastSubscriptionAttempt = {};
  final Duration _subscriptionThrottleTime = const Duration(seconds: 10);
  
  // Event handlers by type and event
  final Map<String, Map<String, List<Function(Map<String, dynamic>)>>> _eventHandlers = {};
  
  // Connection state stream
  final StreamController<bool> _connectionStateController = StreamController<bool>.broadcast();
  Stream<bool> get connectionStateStream => _connectionStateController.stream;
  
  // Message stream
  final StreamController<Map<String, dynamic>> _messageController = StreamController<Map<String, dynamic>>.broadcast();
  Stream<Map<String, dynamic>> get messageStream => _messageController.stream;
  
  // Private constructor
  WebSocketService._internal() {
    _logger.info('WebSocketService instance created - waiting for initialization');
    _loadServerUrl();
  }
  
  // Factory constructor for singleton
  factory WebSocketService() {
    return _instance;
  }
  
  // Explicit initialization method to be called in main.dart before any providers
  void initialize() {
    if (_isInitialized) return;
    _isInitialized = true;
    _loadServerUrl();
    _logger.info('WebSocketService explicitly initialized');
  }
  
  // Load server URL from SharedPreferences
  Future<void> _loadServerUrl() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      _serverUrl = prefs.getString('api_url');
      _logger.debug('Loaded server URL: $_serverUrl');
    } catch (e) {
      _logger.error('Error loading server URL from preferences', e);
    }
  }
  
  // Set server URL
  void setServerUrl(String? url) {
    _serverUrl = url;
    _logger.debug('WebSocket service server URL updated: $url');
  }
  
  // Get server URL
  String? get serverUrl => _serverUrl;
  
  // Getters
  bool get isConnected => _isConnected;
  bool get isConnecting => _isConnecting;
  String? get connectionState => _isConnected ? 'connected' : (_isConnecting ? 'connecting' : 'disconnected');
  bool get isInitialized => _isInitialized;
  
  // Set authentication information
  void setAuthToken(String? token) {
    _authToken = token;
  }
  
  void setUserId(String? id) {
    _userId = id;
  }
  
  // Connect to WebSocket with authentication
  Future<bool> connect() async {
    // Ensure we're initialized first
    if (!_isInitialized) {
      _logger.warning('Cannot connect WebSocket: Service not initialized');
      return false;
    }

    if (_isConnected) {
      _logger.debug('Already connected to WebSocket');
      return true;
    }
    
    if (_isConnecting) {
      _logger.debug('Connection already in progress');
      return false;
    }
    
    if (_authToken == null) {
      _logger.warning('Cannot connect: No auth token provided');
      return false;
    }
    
    _isConnecting = true;
    
    try {
      // Use the existing mechanism for determining the WebSocket URL
      // This preserves whatever URL determination logic was already in place
      final uri = _getWebsocketUri();
      
      _logger.info('Connecting to WebSocket...');
      
      _channel = WebSocketChannel.connect(uri);
      
      // Wait for the connection to be established
      await Future.delayed(const Duration(milliseconds: 300));
      
      _isConnected = true;
      _isConnecting = false;
      
      // Listen for messages
      _channel?.stream.listen(
        (message) {
          try {
            // Handle empty or whitespace-only messages
            if (message == null || (message is String && message.trim().isEmpty)) {
              _logger.debug('Received empty message, ignoring');
              return;
            }

            // Process message content - may be one or multiple concatenated JSON objects
            if (message is String) {
              // Split by newlines in case server sends multiple JSON objects
              final messages = message.split('\n')
                .where((m) => m.trim().isNotEmpty)
                .toList();
              
              _logger.debug('Processing ${messages.length} message segments');
              
              for (final msgPart in messages) {
                try {
                  final data = jsonDecode(msgPart);
                  if (data is Map<String, dynamic>) {
                    _handleMessage(data);
                    _messageController.add(data);
                  }
                } catch (e) {
                  _logger.error('Error parsing JSON segment: $e');
                  _logger.error('Problematic JSON: ${msgPart.substring(0, msgPart.length > 50 ? 50 : msgPart.length)}...');
                }
              }
            } else {
              _logger.warning('Received non-string message: $message');
            }
          } catch (e) {
            _logger.error('Error processing WebSocket message: $e');
          }
        },
        onError: (error) {
          _logger.error('WebSocket error: $error');
          _handleDisconnect();
        },
        onDone: () {
          _logger.info('WebSocket connection closed');
          _handleDisconnect();
        }
      );
      
      // Notify of connection state change
      _connectionStateController.add(true);
      _logger.info('WebSocket connected successfully');
      
      return true;
    } catch (e) {
      _logger.error('WebSocket connection failed: $e');
      _isConnected = false;
      _isConnecting = false;
      _connectionStateController.add(false);
      return false;
    }
  }
  
  // Get WebSocket URI with authentication parameters
  Uri _getWebsocketUri() {
    // Convert HTTP URL to WebSocket URL (ws:// or wss://)
    String wsUrl;
    
    if (_serverUrl != null && _serverUrl!.isNotEmpty) {
      if (_serverUrl!.startsWith('https://')) {
        wsUrl = 'wss://${_serverUrl!.substring(8)}/ws';
      } else if (_serverUrl!.startsWith('http://')) {
        wsUrl = 'ws://${_serverUrl!.substring(7)}/ws';
      } else {
        wsUrl = 'ws://$_serverUrl/ws';
      }
      _logger.debug('Using WebSocket URL from server URL: $wsUrl');
    } else {
      final defaultUrl = AppConfig.serverUrl;
      if (defaultUrl.isNotEmpty) {
        wsUrl = defaultUrl.replaceFirst('http', 'ws') + '/ws';
        _logger.warning('No server URL configured, using default: $wsUrl');
      } else {
        wsUrl = 'ws://localhost/ws'; // Final fallback
        _logger.error('No server URL available, using localhost fallback: $wsUrl');
      }
    }
    
    final uri = Uri.parse(wsUrl);
    
    // Add query parameters for authentication
    return uri.replace(
      queryParameters: {
        'token': _authToken!,
        if (_userId != null) 'user_id': _userId!
      }
    );
  }
  
  // Subscribe to a resource
  void subscribe(String resource, {String? id}) {
    if (!_isConnected) {
      _logger.warning('Cannot subscribe: Not connected');
      return;
    }
    
    final subscriptionKey = _buildSubscriptionKey(resource, id);
    
    // Check if already subscribed or pending
    if (_confirmedSubscriptions.contains(subscriptionKey)) {
      return;
    }
    
    if (_pendingSubscriptions.contains(subscriptionKey)) {
      // Check if recently attempted
      final lastAttempt = _lastSubscriptionAttempt[subscriptionKey];
      if (lastAttempt != null && 
          DateTime.now().difference(lastAttempt) < _subscriptionThrottleTime) {
        return;
      }
    }
    
    // Create subscription message matching server-expected format
    final message = {
      'type': 'subscribe',
      'payload': {
        'resource': resource,
        if (id != null) 'id': id
      }
    };
    
    _sendMessage(message);
    
    // Update subscription tracking
    _pendingSubscriptions.add(subscriptionKey);
    _lastSubscriptionAttempt[subscriptionKey] = DateTime.now();
  }
  
  // Subscribe to an event type (like note.created)
  void subscribeToEvent(String eventType) {
    if (!_isConnected) {
      _logger.warning('Cannot subscribe to event: Not connected');
      return;
    }
    
    final subscriptionKey = 'event:$eventType';
    
    // Check if already subscribed
    if (_confirmedSubscriptions.contains(subscriptionKey)) {
      return;
    }
    
    if (_pendingSubscriptions.contains(subscriptionKey)) {
      // Check if recently attempted
      final lastAttempt = _lastSubscriptionAttempt[subscriptionKey];
      if (lastAttempt != null && 
          DateTime.now().difference(lastAttempt) < _subscriptionThrottleTime) {
        return;
      }
    }
    
    // Create event subscription message matching server-expected format
    final message = {
      'type': 'subscribe',
      'payload': {
        'event_type': eventType
      }
    };
    
    _sendMessage(message);
    
    // Update subscription tracking
    _pendingSubscriptions.add(subscriptionKey);
    _lastSubscriptionAttempt[subscriptionKey] = DateTime.now();
  }
  
  // Batch subscribe to multiple resources
  void batchSubscribe(List<Subscription> subscriptions) {
    if (!_isConnected) {
      _logger.warning('Cannot batch subscribe: Not connected');
      return;
    }
    
    for (final subscription in subscriptions) {
      subscribe(subscription.resource, id: subscription.id);
    }
  }
  
  // Unsubscribe from a resource
  void unsubscribe(String resource, {String? id}) {
    final subscriptionKey = _buildSubscriptionKey(resource, id);
    
    // Send unsubscribe message if connected
    if (_isConnected) {
      final message = {
        'type': 'unsubscribe',
        'payload': {
          'resource': resource,
          if (id != null) 'id': id
        }
      };
      
      _sendMessage(message);
    }
    
    // Update subscription tracking
    _pendingSubscriptions.remove(subscriptionKey);
    _confirmedSubscriptions.remove(subscriptionKey);
    _lastSubscriptionAttempt.remove(subscriptionKey);
  }
  
  // Unsubscribe from an event
  void unsubscribeFromEvent(String eventType) {
    final subscriptionKey = 'event:$eventType';
    
    // Send unsubscribe message if connected
    if (_isConnected) {
      final message = {
        'type': 'unsubscribe',
        'payload': {
          'event_type': eventType
        }
      };
      
      _sendMessage(message);
    }
    
    // Update subscription tracking
    _pendingSubscriptions.remove(subscriptionKey);
    _confirmedSubscriptions.remove(subscriptionKey);
    _lastSubscriptionAttempt.remove(subscriptionKey);
  }
  
  // Add event listener
  void addEventListener(String type, String event, Function(Map<String, dynamic>) handler) {
    _eventHandlers[type] ??= {};
    _eventHandlers[type]![event] ??= [];
    
    // Check for duplicate
    if (!_eventHandlers[type]![event]!.contains(handler)) {
      _eventHandlers[type]![event]!.add(handler);
    }
  }
  
  // Remove event listener
  void removeEventListener(String type, String event, [Function(Map<String, dynamic>)? handler]) {
    if (_eventHandlers.containsKey(type) && _eventHandlers[type]!.containsKey(event)) {
      if (handler == null) {
        // Remove all handlers for this event
        _eventHandlers[type]!.remove(event);
      } else {
        // Remove specific handler
        _eventHandlers[type]![event]!.remove(handler);
      }
    }
  }
  
  // Check if subscribed to a resource
  bool isSubscribed(String resource, {String? id}) {
    final subscriptionKey = _buildSubscriptionKey(resource, id);
    return _confirmedSubscriptions.contains(subscriptionKey);
  }
  
  // Check if subscribed to an event
  bool isSubscribedToEvent(String eventType) {
    return _confirmedSubscriptions.contains('event:$eventType');
  }
  
  // Handle incoming messages with improved logging
  void _handleMessage(Map<String, dynamic> message) {
    try {
      _logger.debug('Received raw message: $message');

      final String type = message['type'] ?? 'unknown';
      final String event = message['event'] ?? 'unknown';
      
      _logger.debug('Received message: type=$type, event=$event');
      
      // Handle subscription confirmations
      if (type == 'subscription' && event == 'confirmed') {
        _handleSubscriptionConfirmation(message);
        return;
      }
      
      // Handle events
      if (_eventHandlers.containsKey(type) && _eventHandlers[type]!.containsKey(event)) {
        final handlers = _eventHandlers[type]![event]!;
        
        for (final handler in handlers) {
          try {
            handler(message);
          } catch (e) {
            _logger.error('Error in event handler for $type:$event', e);
          }
        }
      }
    } catch (e) {
      _logger.error('Error handling message: $e, message: $message');
    }
  }
  
  // Handle subscription confirmation with improved logging
  void _handleSubscriptionConfirmation(Map<String, dynamic> message) {
    try {
      final payload = message['payload'];
      if (payload == null) {
        _logger.warning('Received subscription confirmation with null payload');
        return;
      }
      
      _logger.debug('Processing subscription confirmation: $payload');
      
      String? subscriptionKey;
      
      // Resource subscription confirmation
      if (payload['resource'] != null) {
        final resource = payload['resource']?.toString();
        final id = payload['id']?.toString();
        subscriptionKey = _buildSubscriptionKey(resource, id);
      } 
      // Event subscription confirmation
      else if (payload['event_type'] != null) {
        final eventType = payload['event_type']?.toString();
        if (eventType != null) {
          subscriptionKey = 'event:$eventType';
        }
      }
      
      if (subscriptionKey != null) {
        _logger.info('Confirmed subscription: $subscriptionKey');
        _pendingSubscriptions.remove(subscriptionKey);
        _confirmedSubscriptions.add(subscriptionKey);
      } else {
        _logger.warning('Could not determine subscription key from payload: $payload');
      }
    } catch (e) {
      _logger.error('Error handling subscription confirmation: $e, message: $message');
    }
  }
  
  // Helper to build subscription key
  String _buildSubscriptionKey(String? resource, String? id) {
    if (resource == null) return '';
    return id != null ? '$resource:$id' : resource;
  }
  
  // Send a message to the WebSocket with improved debugging
  void _sendMessage(dynamic message) {
    if (!_isConnected || _channel == null) {
      _logger.warning('Cannot send message: Not connected');
      return;
    }
    
    try {
      final String jsonMessage = jsonEncode(message);
      _logger.debug('Sending WebSocket message: $jsonMessage');
      _channel!.sink.add(jsonMessage);
    } catch (e) {
      _logger.error('Error sending WebSocket message: $e');
    }
  }
  
  // Disconnect WebSocket
  void disconnect() {
    _logger.info('Disconnecting WebSocket');
    
    if (_isConnected && _channel != null) {
      _channel!.sink.close();
      _channel = null;
    }
    
    _handleDisconnect();
  }
  
  // Handle disconnect
  void _handleDisconnect() {
    _isConnected = false;
    _isConnecting = false;
    _connectionStateController.add(false);
  }
  
  // Clear state on logout
  void clearState() {
    _logger.info('Clearing WebSocket state');
    
    // Disconnect if connected
    if (_isConnected) {
      disconnect();
    }
    
    // Clear authentication
    _authToken = null;
    _userId = null;
    
    // Clear subscriptions
    _confirmedSubscriptions.clear();
    _pendingSubscriptions.clear();
    _lastSubscriptionAttempt.clear();
  }
  
  // Clean up resources
  void dispose() {
    disconnect();
    _messageController.close();
    _connectionStateController.close();
  }
}

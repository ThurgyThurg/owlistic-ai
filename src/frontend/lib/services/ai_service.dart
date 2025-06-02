import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:owlistic/utils/logger.dart';
import 'base_service.dart';

class AIService extends BaseService {
  static final Logger _logger = Logger('AIService');

  /// Process a note with AI to generate summary, tags, and embeddings
  Future<Map<String, dynamic>> processNoteWithAI(String noteId) async {
    try {
      _logger.info('Processing note $noteId with AI');
      
      // Use a longer timeout for AI processing (2 minutes)
      final uri = await createUri('/api/v1/ai/notes/$noteId/process');
      final response = await http.post(
        uri,
        headers: getAuthHeaders(),
        body: jsonEncode({}),
      ).timeout(const Duration(minutes: 2));
      
      // Check status and throw error if not successful
      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw Exception('HTTP ${response.statusCode}: ${response.body}');
      }
      
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      _logger.info('AI processing started for note $noteId');
      return data;
    } catch (e) {
      _logger.error('Failed to process note with AI: $e');
      rethrow;
    }
  }

  /// Get AI-enhanced note with metadata
  Future<Map<String, dynamic>> getEnhancedNote(String noteId) async {
    try {
      _logger.info('Fetching enhanced note $noteId');
      
      final response = await authenticatedGet('/api/v1/ai/notes/$noteId/enhanced');
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      return data;
    } catch (e) {
      _logger.error('Failed to fetch enhanced note: $e');
      rethrow;
    }
  }

  /// Perform semantic search using AI embeddings
  Future<Map<String, dynamic>> semanticSearch(String query, {int limit = 10}) async {
    try {
      _logger.info('Performing semantic search: $query');
      
      final response = await authenticatedPost('/api/v1/ai/notes/search/semantic', {
        'query': query,
        'limit': limit,
      });
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      _logger.info('Semantic search returned ${data['results']?.length ?? 0} results');
      return data;
    } catch (e) {
      _logger.error('Semantic search failed: $e');
      rethrow;
    }
  }

  /// Create an AI project
  Future<Map<String, dynamic>> createAIProject({
    required String name,
    String? description,
    List<String>? aiTags,
    Map<String, dynamic>? aiMetadata,
  }) async {
    try {
      _logger.info('Creating AI project: $name');
      
      final body = <String, dynamic>{
        'name': name,
      };
      if (description != null) body['description'] = description;
      if (aiTags != null) body['ai_tags'] = aiTags;
      if (aiMetadata != null) body['ai_metadata'] = aiMetadata;
      
      final response = await authenticatedPost('/api/v1/ai/projects', body);
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      _logger.info('AI project created: ${data['id']}');
      return data;
    } catch (e) {
      _logger.error('Failed to create AI project: $e');
      rethrow;
    }
  }

  /// Get all AI projects
  Future<List<dynamic>> getAIProjects() async {
    try {
      _logger.info('Fetching AI projects');
      
      final response = await authenticatedGet('/api/v1/ai/projects');
      final data = jsonDecode(response.body);
      
      return data is List ? data : [data];
    } catch (e) {
      _logger.error('Failed to fetch AI projects: $e');
      rethrow;
    }
  }

  /// Get specific AI project
  Future<Map<String, dynamic>> getAIProject(String projectId) async {
    try {
      _logger.info('Fetching AI project: $projectId');
      
      final response = await authenticatedGet('/api/v1/ai/projects/$projectId');
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      return data;
    } catch (e) {
      _logger.error('Failed to fetch AI project: $e');
      rethrow;
    }
  }

  /// Update AI project
  Future<Map<String, dynamic>> updateAIProject(
    String projectId, {
    String? name,
    String? description,
    String? status,
    List<String>? aiTags,
    Map<String, dynamic>? aiMetadata,
  }) async {
    try {
      _logger.info('Updating AI project: $projectId');
      
      final body = <String, dynamic>{};
      if (name != null) body['name'] = name;
      if (description != null) body['description'] = description;
      if (status != null) body['status'] = status;
      if (aiTags != null) body['ai_tags'] = aiTags;
      if (aiMetadata != null) body['ai_metadata'] = aiMetadata;
      
      final response = await authenticatedPut('/api/v1/ai/projects/$projectId', body);
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      _logger.info('AI project updated: $projectId');
      return data;
    } catch (e) {
      _logger.error('Failed to update AI project: $e');
      rethrow;
    }
  }

  /// Delete AI project
  Future<void> deleteAIProject(String projectId) async {
    try {
      _logger.info('Deleting AI project: $projectId');
      
      await authenticatedDelete('/api/v1/ai/projects/$projectId');
      
      _logger.info('AI project deleted: $projectId');
    } catch (e) {
      _logger.error('Failed to delete AI project: $e');
      rethrow;
    }
  }

  /// Run an AI agent
  Future<Map<String, dynamic>> runAgent(String agentType, Map<String, dynamic> inputData) async {
    try {
      _logger.info('Running AI agent: $agentType');
      
      final response = await authenticatedPost('/api/v1/ai/agents/run', {
        'agent_type': agentType,
        'input_data': inputData,
      });
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      _logger.info('AI agent started: ${data['id']}');
      return data;
    } catch (e) {
      _logger.error('Failed to run AI agent: $e');
      rethrow;
    }
  }

  /// Get agent runs
  Future<List<dynamic>> getAgentRuns({int limit = 20}) async {
    try {
      _logger.info('Fetching agent runs');
      
      final response = await authenticatedGet('/api/v1/ai/agents/runs', 
        queryParameters: {'limit': limit});
      final data = jsonDecode(response.body);
      
      return data is List ? data : [data];
    } catch (e) {
      _logger.error('Failed to fetch agent runs: $e');
      rethrow;
    }
  }

  /// Get specific agent run
  Future<Map<String, dynamic>> getAgentRun(String agentId) async {
    try {
      _logger.info('Fetching agent run: $agentId');
      
      final response = await authenticatedGet('/api/v1/ai/agents/runs/$agentId');
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      return data;
    } catch (e) {
      _logger.error('Failed to fetch agent run: $e');
      rethrow;
    }
  }

  /// Quick goal agent - plan a goal with AI
  Future<Map<String, dynamic>> runQuickGoal(String goal, {String context = ''}) async {
    try {
      _logger.info('Running quick goal agent: $goal');
      
      final response = await authenticatedPost('/api/v1/ai/agents/quick-goal', {
        'goal': goal,
        'context': context,
      });
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      _logger.info('Quick goal agent started: ${data['id']}');
      return data;
    } catch (e) {
      _logger.error('Failed to run quick goal agent: $e');
      rethrow;
    }
  }

  /// Chat with AI
  Future<Map<String, dynamic>> chatWithAI(String message, {String? sessionId}) async {
    try {
      _logger.info('Chatting with AI: ${message.substring(0, message.length > 50 ? 50 : message.length)}...');
      
      final body = <String, dynamic>{
        'message': message,
      };
      if (sessionId != null) body['session_id'] = sessionId;
      
      final response = await authenticatedPost('/api/v1/ai/chat', body);
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      _logger.info('AI chat response received');
      return data;
    } catch (e) {
      _logger.error('AI chat failed: $e');
      rethrow;
    }
  }

  /// Get chat history
  Future<Map<String, dynamic>> getChatHistory(String sessionId) async {
    try {
      _logger.info('Fetching chat history: $sessionId');
      
      final response = await authenticatedGet('/api/v1/ai/chat/history', 
        queryParameters: {'session_id': sessionId});
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      
      return data;
    } catch (e) {
      _logger.error('Failed to fetch chat history: $e');
      rethrow;
    }
  }
}
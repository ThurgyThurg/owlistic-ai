# AI Second Brain Integration Guide

This guide explains how the AI Second Brain features have been integrated into your existing Owlistic note-taking application.

## 🔗 Integration Overview

Instead of creating a separate system, the AI capabilities have been **integrated directly into your existing Go/Gin backend** and can be used with your **Flutter frontend**.

### What's Been Added

1. **AI Models** (in `models/ai_models.go`):
   - `AIEnhancedNote`: Extends your existing Note with AI metadata
   - `AIAgent`: Manages AI agent executions
   - `AIProject`: AI-powered project management
   - `ChatMemory`: Conversational AI history

2. **AI Service** (in `services/ai_service.go`):
   - Integrates with Anthropic Claude and OpenAI APIs
   - Automatic title generation, summarization, and tagging
   - Vector embeddings for semantic search
   - Related note discovery

3. **AI Routes** (in `routes/ai_routes.go`):
   - REST API endpoints for AI features
   - Integrated with your existing auth middleware
   - Compatible with your current user system

## 🚀 Quick Setup

### 1. Environment Variables

Add these to your backend environment:

```bash
# Required for AI features
ANTHROPIC_API_KEY=sk-ant-api03-your-anthropic-key
OPENAI_API_KEY=sk-your-openai-key

# Optional for vector search (ChromaDB)
CHROMA_BASE_URL=http://localhost:8000
```

### 2. Install and Run

```bash
cd src/backend

# Install new dependencies (if any)
go mod tidy

# Run database migrations (will create AI tables)
go run cmd/main.go
```

The AI models will be automatically migrated when you start the backend.

## 📱 Flutter Integration

### New API Endpoints Available

Your Flutter app can now call these new endpoints:

```dart
// AI-enhance a note
POST /api/v1/ai/notes/{id}/process

// Get enhanced note with AI metadata
GET /api/v1/ai/notes/{id}/enhanced

// Semantic search
POST /api/v1/ai/notes/search/semantic
{
  "query": "productivity techniques",
  "limit": 10
}

// Create AI project
POST /api/v1/ai/projects
{
  "name": "Learning Flutter",
  "description": "Track my Flutter learning progress"
}

// Run AI agent
POST /api/v1/ai/agents/quick-goal
{
  "goal": "Plan my week",
  "context": "I have 3 deadlines coming up"
}

// Chat with AI
POST /api/v1/ai/chat
{
  "message": "Help me organize my notes",
  "session_id": "optional-session-id"
}
```

### Flutter Service Example

Add this to your Flutter `services/ai_service.dart`:

```dart
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'base_service.dart';

class AIService extends BaseService {
  // Process note with AI
  Future<Map<String, dynamic>> processNoteWithAI(String noteId) async {
    final response = await makeRequest(
      'POST',
      '/ai/notes/$noteId/process',
    );
    return response;
  }

  // Get AI-enhanced note
  Future<Map<String, dynamic>> getEnhancedNote(String noteId) async {
    final response = await makeRequest(
      'GET', 
      '/ai/notes/$noteId/enhanced',
    );
    return response;
  }

  // Semantic search
  Future<List<dynamic>> semanticSearch(String query, {int limit = 10}) async {
    final response = await makeRequest(
      'POST',
      '/ai/notes/search/semantic',
      body: {
        'query': query,
        'limit': limit,
      },
    );
    return response['results'] ?? [];
  }

  // Quick goal agent
  Future<Map<String, dynamic>> runQuickGoal(String goal, {String context = ''}) async {
    final response = await makeRequest(
      'POST',
      '/ai/agents/quick-goal',
      body: {
        'goal': goal,
        'context': context,
      },
    );
    return response;
  }

  // Chat with AI
  Future<Map<String, dynamic>> chatWithAI(String message, {String? sessionId}) async {
    final response = await makeRequest(
      'POST',
      '/ai/chat',
      body: {
        'message': message,
        if (sessionId != null) 'session_id': sessionId,
      },
    );
    return response;
  }
}
```

### UI Integration Examples

#### 1. Add AI Processing Button to Note Editor

```dart
// In your note editor screen
FloatingActionButton(
  onPressed: () async {
    setState(() => _isProcessing = true);
    
    try {
      await AIService().processNoteWithAI(widget.noteId);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('AI processing started!')),
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('AI processing failed: $e')),
      );
    } finally {
      setState(() => _isProcessing = false);
    }
  },
  child: _isProcessing 
    ? CircularProgressIndicator(color: Colors.white)
    : Icon(Icons.auto_awesome),
  tooltip: 'Enhance with AI',
)
```

#### 2. Semantic Search Widget

```dart
class SemanticSearchWidget extends StatefulWidget {
  @override
  _SemanticSearchWidgetState createState() => _SemanticSearchWidgetState();
}

class _SemanticSearchWidgetState extends State<SemanticSearchWidget> {
  final _searchController = TextEditingController();
  List<dynamic> _results = [];
  bool _isSearching = false;

  Future<void> _performSearch() async {
    if (_searchController.text.isEmpty) return;

    setState(() => _isSearching = true);
    
    try {
      final results = await AIService().semanticSearch(_searchController.text);
      setState(() => _results = results);
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Search failed: $e')),
      );
    } finally {
      setState(() => _isSearching = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        TextField(
          controller: _searchController,
          decoration: InputDecoration(
            labelText: 'Semantic Search',
            suffixIcon: IconButton(
              icon: Icon(Icons.search),
              onPressed: _performSearch,
            ),
          ),
          onSubmitted: (_) => _performSearch(),
        ),
        if (_isSearching) CircularProgressIndicator(),
        Expanded(
          child: ListView.builder(
            itemCount: _results.length,
            itemBuilder: (context, index) {
              final note = _results[index];
              return ListTile(
                title: Text(note['title'] ?? 'Untitled'),
                subtitle: Text(note['summary'] ?? ''),
                onTap: () {
                  // Navigate to note
                  Navigator.pushNamed(context, '/note', arguments: note['id']);
                },
              );
            },
          ),
        ),
      ],
    );
  }
}
```

#### 3. AI Chat Interface

```dart
class AIChatScreen extends StatefulWidget {
  @override
  _AIChatScreenState createState() => _AIChatScreenState();
}

class _AIChatScreenState extends State<AIChatScreen> {
  final _messageController = TextEditingController();
  final List<Map<String, dynamic>> _messages = [];
  String? _sessionId;

  Future<void> _sendMessage() async {
    final message = _messageController.text.trim();
    if (message.isEmpty) return;

    setState(() {
      _messages.add({'role': 'user', 'content': message});
    });
    _messageController.clear();

    try {
      final response = await AIService().chatWithAI(message, sessionId: _sessionId);
      _sessionId = response['session_id'];
      
      setState(() {
        _messages.add({'role': 'assistant', 'content': response['response']});
      });
    } catch (e) {
      setState(() {
        _messages.add({'role': 'error', 'content': 'Failed to get AI response'});
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('AI Assistant')),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              itemCount: _messages.length,
              itemBuilder: (context, index) {
                final message = _messages[index];
                final isUser = message['role'] == 'user';
                
                return Align(
                  alignment: isUser ? Alignment.centerRight : Alignment.centerLeft,
                  child: Container(
                    margin: EdgeInsets.all(8),
                    padding: EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: isUser ? Colors.blue : Colors.grey[300],
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text(
                      message['content'],
                      style: TextStyle(
                        color: isUser ? Colors.white : Colors.black,
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
          Padding(
            padding: EdgeInsets.all(8),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _messageController,
                    decoration: InputDecoration(
                      hintText: 'Ask me anything about your notes...',
                      border: OutlineInputBorder(),
                    ),
                    onSubmitted: (_) => _sendMessage(),
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.send),
                  onPressed: _sendMessage,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
```

## 🔄 How It Works

### 1. Automatic Note Enhancement

When a user creates or edits a note:

1. **Manual Trigger**: User clicks "Enhance with AI" button
2. **API Call**: Flutter calls `POST /api/v1/ai/notes/{id}/process`
3. **Background Processing**: Go backend:
   - Extracts text from note blocks
   - Calls Anthropic API for title/summary/tags
   - Calls OpenAI API for embeddings
   - Finds related notes using similarity
   - Stores results in `ai_enhanced_notes` table
4. **UI Update**: Flutter can fetch enhanced data via `GET /api/v1/ai/notes/{id}/enhanced`

### 2. Semantic Search

Traditional text search vs. AI semantic search:

```dart
// Traditional search (existing)
final textResults = await NoteService().searchNotes('flutter widgets');

// AI semantic search (new)
final semanticResults = await AIService().semanticSearch('flutter UI components');
// This finds notes about widgets, components, UI elements even if they don't contain exact keywords
```

### 3. AI Agents

Run AI agents for complex tasks:

```dart
// Goal planning agent
final agent = await AIService().runQuickGoal(
  'Learn Flutter in 30 days',
  context: 'I have basic programming experience but new to mobile development'
);

// Agent will create:
// - Learning milestones
// - Practice projects  
// - Resource recommendations
// - Timeline with deadlines
```

## 🗄️ Database Changes

The integration adds these new tables to your existing database:

- `ai_enhanced_notes`: AI metadata for notes (summary, tags, embeddings)
- `ai_agents`: Agent execution tracking
- `ai_projects`: AI-powered project management
- `ai_task_enhancements`: AI features for tasks
- `chat_memories`: Conversation history

**No changes to existing tables** - your current data remains untouched.

## 🚀 Deployment

### Development
```bash
# Set environment variables
export ANTHROPIC_API_KEY=your_key
export OPENAI_API_KEY=your_key

# Start backend (will auto-migrate AI tables)
cd src/backend && go run cmd/main.go

# Start Flutter app
cd src/frontend && flutter run
```

### Production

Add the environment variables to your deployment environment and restart the backend. The AI features will be automatically available.

## 🎯 Next Steps

1. **Test Basic Integration**:
   - Add AI processing button to note editor
   - Test note enhancement API
   - Implement semantic search

2. **Advanced Features**:
   - Vector database (ChromaDB) for better semantic search
   - Background job processing for AI tasks
   - Real-time AI suggestions

3. **UI Enhancements**:
   - AI-generated tag suggestions
   - Related notes sidebar
   - Chat interface for note queries

## 🔧 Troubleshooting

### Common Issues

**AI API not working**:
- Check environment variables are set
- Verify API keys have sufficient credits
- Check network connectivity

**Database migration errors**:
- Ensure PostgreSQL version compatibility
- Check database permissions
- Review migration logs

**Flutter compilation errors**:
- Run `flutter clean && flutter pub get`
- Check Dart/Flutter version compatibility

This integration maintains your existing architecture while adding powerful AI capabilities. Your current notes, notebooks, and tasks continue to work exactly as before, with optional AI enhancements available when you want them.
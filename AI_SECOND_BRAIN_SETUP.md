# AI Second Brain System - Complete Setup Guide

This guide will help you set up the complete AI-powered personal knowledge management system with FastAPI backend and React frontend.

## 🎯 System Overview

The AI Second Brain System provides:
- **Intelligent Memory Management**: Auto-categorization, tagging, and summarization
- **Semantic Search**: Vector-based similarity search with ChromaDB
- **AI Agents**: Autonomous task planning, scheduling, and goal achievement
- **Project Organization**: Automatic linking of related content
- **Terminal-Themed UI**: Cyberpunk-inspired interface with CRT effects
- **Background Processing**: File monitoring, Telegram bot, calendar sync

## 🚀 Quick Start with Docker

### Prerequisites
- Docker and Docker Compose
- API keys for Anthropic and OpenAI
- (Optional) Telegram bot token and Google Calendar credentials

### 1. Environment Setup

Create a `.env` file in the project root:

```bash
# Required AI API Keys
ANTHROPIC_API_KEY=sk-ant-api03-your-anthropic-key-here
OPENAI_API_KEY=sk-your-openai-key-here

# Security
SECRET_KEY=your-super-secure-secret-key-change-this

# Optional Integrations
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
TELEGRAM_CHAT_ID=your-telegram-chat-id
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
```

### 2. Start the System

```bash
# Start core services (API + Database + Frontend)
docker-compose -f docker-compose.ai-brain.yml up -d

# Optional: Start with Telegram bot
docker-compose -f docker-compose.ai-brain.yml --profile telegram up -d

# Optional: Start with file watcher
docker-compose -f docker-compose.ai-brain.yml --profile file-watcher up -d

# Start everything
docker-compose -f docker-compose.ai-brain.yml --profile telegram --profile file-watcher up -d
```

### 3. Access the System

- **Frontend UI**: http://localhost:3000
- **API Documentation**: http://localhost:8000/docs
- **Database**: localhost:5432 (ai_user/ai_password)

## 🛠️ Manual Development Setup

### Backend Setup

```bash
cd src/ai_backend

# Create virtual environment
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Setup environment
cp .env.example .env
# Edit .env with your API keys

# Start PostgreSQL and Redis
# Ubuntu/Debian:
sudo systemctl start postgresql redis-server

# macOS with Homebrew:
brew services start postgresql redis

# Create database
createdb ai_second_brain

# Start the backend
python main.py
```

### Frontend Setup

```bash
cd src/ai_frontend

# Install dependencies
npm install

# Start development server
npm start
```

## 🔧 Configuration Guide

### API Keys Setup

1. **Anthropic (Claude)**:
   - Visit https://console.anthropic.com
   - Create an API key
   - Add to `.env` as `ANTHROPIC_API_KEY`

2. **OpenAI (Embeddings + Whisper)**:
   - Visit https://platform.openai.com
   - Create an API key
   - Add to `.env` as `OPENAI_API_KEY`

### Optional Integrations

#### Telegram Bot

1. Create bot with @BotFather on Telegram
2. Get your chat ID by messaging @userinfobot
3. Add `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` to `.env`

**Bot Commands:**
- `/goal "Plan my week"` - Run reasoning agent
- `/agent reasoning_loop {"goal": "Learn Python"}` - Run specific agent
- Send voice messages for transcription
- Send text messages to capture in memory

#### Google Calendar

1. Go to Google Cloud Console
2. Create OAuth 2.0 credentials
3. Add redirect URI: `http://localhost:8000/auth/google/callback`
4. Add `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` to `.env`

#### File Watcher

1. Set `WATCHED_FOLDER` in `.env`
2. Drop PDF, DOCX, or Markdown files in the folder
3. Files are automatically processed and added to memory

## 🎮 Usage Examples

### 1. Quick Memory Entry

Visit the dashboard and use the Quick Entry form:

```
Title: Meeting Notes
Content: Discussed Q4 roadmap with team. Key priorities: 
- Launch new AI features
- Improve performance by 20%
- Expand to European markets

AI will automatically:
- Generate tags: ["meeting", "roadmap", "AI", "performance", "expansion"]
- Create summary
- Find related entries
- Suggest project assignment
```

### 2. AI Agent Goal Planning

```
Goal: "Plan a productive week"
Context: "I have 3 project deadlines and want to learn React"

Agent will:
1. Break down the goal into specific tasks
2. Schedule tasks with realistic timeframes
3. Create calendar events
4. Set up progress tracking
```

### 3. Semantic Search

Search for: "productivity techniques"

The system will find:
- Entries about time management
- Notes on workflow optimization  
- Related project content
- Similar concepts even without exact keywords

### 4. Project Auto-Organization

When you add entries about "Machine Learning Course", the AI will:
- Create a project if it doesn't exist
- Link related entries automatically
- Suggest additional relevant content
- Track progress and deadlines

## 🏗️ Architecture Deep Dive

### Data Flow

```
Input → AI Processing → Storage → Search/Retrieval
  ↓         ↓            ↓           ↓
Text    Summarize    Database    Vector Search
Voice   Tag/Embed    ChromaDB    Text Search
File    Categorize   PostgreSQL  Agent Tools
```

### AI Processing Pipeline

1. **Content Analysis**: Anthropic Claude analyzes content
2. **Embedding Generation**: OpenAI creates vector embeddings
3. **Similarity Search**: ChromaDB finds related content
4. **Auto-Organization**: AI suggests projects and categories
5. **Background Processing**: Async tasks handle heavy operations

### Agent System

**Reasoning Loop Agent**:
- Takes high-level goals
- Breaks into actionable steps
- Uses tools (search, create, schedule)
- Provides step-by-step execution

**Scheduler Agent**:
- Reviews overdue tasks
- Suggests rescheduling or archiving
- Syncs with Google Calendar
- Sends notifications

**Memory Compression Agent**:
- Finds old entries (60+ days)
- Creates summarized versions
- Maintains original links
- Reduces storage overhead

## 🔍 API Usage Examples

### Create Memory Entry

```bash
curl -X POST "http://localhost:8000/api/memory/entries/" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Learned about FastAPI async patterns today. Very powerful for I/O intensive applications.",
    "tags": ["programming", "python", "fastapi"]
  }'
```

### Semantic Search

```bash
curl -X POST "http://localhost:8000/api/search/" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "python web development",
    "search_type": "vector",
    "limit": 10
  }'
```

### Run AI Agent

```bash
curl -X POST "http://localhost:8000/api/agents/quick-goal" \
  -H "Content-Type: application/json" \
  -d '{
    "goal": "Learn Docker containerization",
    "context": "I need this for my current project deployment"
  }'
```

## 🚨 Troubleshooting

### Common Issues

**Database Connection Error**:
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Create database if missing
createdb ai_second_brain
```

**AI API Errors**:
- Verify API keys in `.env`
- Check account credits/limits
- Ensure proper key format

**ChromaDB Permission Issues**:
```bash
# Fix permissions
sudo chown -R $USER:$USER ./chroma_db
chmod -R 755 ./chroma_db
```

**Frontend Build Errors**:
```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install
```

### Performance Optimization

1. **Database**: Use connection pooling
2. **Vector Search**: Batch embedding generation
3. **Caching**: Enable Redis for frequent queries
4. **Background Tasks**: Use Celery for heavy processing

## 🔐 Security Considerations

- Change default `SECRET_KEY` in production
- Use environment variables for all secrets
- Configure CORS properly for production
- Set up database access controls
- Monitor API usage and costs

## 📈 Monitoring & Logs

View logs:
```bash
# Docker logs
docker-compose -f docker-compose.ai-brain.yml logs -f ai-backend

# File logs (development)
tail -f logs/ai_backend.log
```

Health checks:
- Backend: http://localhost:8000/health
- Database: Check PostgreSQL connection
- Vector DB: ChromaDB collection status

## 🔄 Updates & Maintenance

Update the system:
```bash
# Pull latest changes
git pull origin main

# Rebuild containers
docker-compose -f docker-compose.ai-brain.yml build

# Restart services
docker-compose -f docker-compose.ai-brain.yml up -d
```

Backup data:
```bash
# Database backup
docker exec postgres pg_dump -U ai_user ai_second_brain > backup.sql

# Vector database backup
docker cp $(docker-compose ps -q ai-backend):/app/chroma_db ./chroma_backup
```

## 🎯 Next Steps

1. **Customize Agents**: Modify agent prompts for your workflow
2. **Add Integrations**: Connect additional services (Notion, Slack, etc.)
3. **Extend UI**: Add more visualization and interaction features
4. **Scale**: Deploy to cloud with proper scaling configuration
5. **Analytics**: Add usage tracking and insights

## 📚 Additional Resources

- **FastAPI Documentation**: https://fastapi.tiangolo.com
- **ChromaDB Guide**: https://docs.trychroma.com
- **Anthropic API**: https://docs.anthropic.com
- **React Query**: https://react-query.tanstack.com

For support or questions, check the project documentation or create an issue in the repository.
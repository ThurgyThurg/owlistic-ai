# Docker AI Setup Guide

This guide shows how to run Owlistic with AI features using Docker Compose, including network access configuration.

## 🚀 Quick Start

### 1. Set Up Environment Variables

**Option A: Create .env file (Recommended)**
```bash
# Copy the example file
cp .env.example .env

# Edit with your actual API keys
nano .env  # or vim, code, etc.
```

**Option B: Export environment variables**
```bash
export ANTHROPIC_API_KEY=sk-ant-api03-your-anthropic-key
export OPENAI_API_KEY=sk-your-openai-key
```

### 2. Build and Start the System

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f owlistic

# Check status
docker-compose ps
```

### 3. Find Your Network IP

**The backend will be accessible at:**
- **Local:** `http://localhost:8080`
- **Network:** `http://YOUR_IP:8080` (replace YOUR_IP with your actual IP)

```bash
# Find your IP address
ip addr show | grep "inet " | grep -v 127.0.0.1
# Example: 192.168.1.100
```

### 4. Configure Flutter App

The Flutter web app will be available at:
- **Local:** `http://localhost:80`
- **Network:** `http://YOUR_IP:80`

## 📋 Required Environment Variables

### **Required for AI Features:**
```bash
ANTHROPIC_API_KEY=sk-ant-api03-...    # Get from https://console.anthropic.com
OPENAI_API_KEY=sk-...                 # Get from https://platform.openai.com
```

### **Optional Integrations:**
```bash
TELEGRAM_BOT_TOKEN=...                # For Telegram bot integration
TELEGRAM_CHAT_ID=...                  # Your Telegram chat ID
GOOGLE_CLIENT_ID=...                  # For Google Calendar sync
GOOGLE_CLIENT_SECRET=...              # Google Calendar credentials
```

## 🔧 Configuration Methods

### Method 1: .env File (Recommended)

1. **Create .env file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit .env file:**
   ```
   ANTHROPIC_API_KEY=sk-ant-api03-your-actual-key
   OPENAI_API_KEY=sk-your-actual-key
   TELEGRAM_BOT_TOKEN=your-bot-token
   TELEGRAM_CHAT_ID=your-chat-id
   ```

3. **Start services:**
   ```bash
   docker-compose up -d
   ```

### Method 2: Environment Variables

```bash
# Set variables in your shell
export ANTHROPIC_API_KEY=sk-ant-api03-your-key
export OPENAI_API_KEY=sk-your-key

# Start services
docker-compose up -d
```

### Method 3: Docker Compose Override

Create `docker-compose.override.yml`:
```yaml
version: '3.8'
services:
  owlistic:
    environment:
      - ANTHROPIC_API_KEY=sk-ant-api03-your-key
      - OPENAI_API_KEY=sk-your-key
```

## 🌐 Network Access Configuration

### Backend Network Access

The Docker configuration is already set up for network access:
- Binds to `0.0.0.0:8080` inside container
- Maps to port `8080` on your host machine
- Accepts connections from any device on your network

### Frontend Network Access

The Flutter web app is configured for network access:
- Accessible at `http://YOUR_IP:80`
- Can connect to backend across network
- Supports mobile and desktop browsers

### Test Network Access

From another device on your network:
```bash
# Test backend API (replace with your IP)
curl http://192.168.1.100:8080/api/v1/health

# Expected response: {"status":"ok"}
```

## 🧪 Testing AI Features

### 1. Test Basic API

```bash
# Check if backend is running
curl http://localhost:8080/api/v1/health

# Should return: {"status":"ok"}
```

### 2. Test AI Endpoints (requires authentication)

```bash
# First, create a user and login to get JWT token
# Then test AI endpoints with the token

curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -X POST http://localhost:8080/api/v1/ai/notes/NOTE_ID/process
```

### 3. Test Through Frontend

1. Open `http://localhost:80` in browser
2. Register/login to create account
3. Create a note
4. Click the AI enhancement button
5. Check for AI-generated tags and summary

## 📊 Service Status and Logs

### Check Service Status
```bash
# View all services
docker-compose ps

# Should show:
# owlistic      - backend API
# owlistic-app  - frontend web app
# postgres      - database
# nats          - message broker
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f owlistic

# Follow AI processing logs
docker-compose logs -f owlistic | grep -i "ai\|anthropic\|openai"
```

### Restart Services
```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart owlistic

# Rebuild and restart
docker-compose up -d --build
```

## 🔒 Security Considerations

### For Local Network Use
- ✅ Safe for local development and testing
- ✅ Protected by your router's firewall
- ✅ API keys are securely passed via environment variables

### For Production Use
- 🔧 Use HTTPS with SSL certificates
- 🔧 Set up proper CORS policies
- 🔧 Add rate limiting
- 🔧 Use secrets management
- 🔧 Configure authentication properly

## 🚨 Troubleshooting

### Container Won't Start
```bash
# Check container logs
docker-compose logs owlistic

# Common issues:
# - Missing API keys
# - Port already in use
# - Database connection failed
```

### AI Features Not Working
```bash
# Check environment variables inside container
docker-compose exec owlistic env | grep -E "ANTHROPIC|OPENAI"

# Should show your API keys
```

### Network Access Issues
```bash
# Check if port is open
netstat -tlnp | grep :8080

# Check Docker port mapping
docker-compose port owlistic 8080
```

### API Key Issues
```bash
# Verify keys are loaded
docker-compose exec owlistic printenv | grep API_KEY

# Test API key manually
curl -H "Authorization: Bearer $ANTHROPIC_API_KEY" \
     https://api.anthropic.com/v1/messages \
     -X POST -d '{"model":"claude-3-sonnet-20240229","max_tokens":10,"messages":[{"role":"user","content":"Hi"}]}'
```

## 📱 Mobile Device Access

### Connect Mobile App
1. **Build Flutter app for mobile**
2. **Set API URL in app to:** `http://YOUR_IP:8080`
3. **Ensure both devices on same WiFi network**

### Web Browser Access
- **Any device browser:** `http://YOUR_IP:80`
- **Works on phones, tablets, laptops**

## 🎯 Available Endpoints

Once running, these endpoints are available:

### Core API
- **Health:** `GET /api/v1/health`
- **Notes:** `GET /api/v1/notes`
- **Notebooks:** `GET /api/v1/notebooks`
- **Tasks:** `GET /api/v1/tasks`

### AI Features
- **Process Note:** `POST /api/v1/ai/notes/{id}/process`
- **Enhanced Note:** `GET /api/v1/ai/notes/{id}/enhanced`
- **Semantic Search:** `POST /api/v1/ai/notes/search/semantic`
- **AI Projects:** `GET /api/v1/ai/projects`
- **Quick Goal:** `POST /api/v1/ai/agents/quick-goal`
- **AI Chat:** `POST /api/v1/ai/chat`

## 🔄 Updates and Maintenance

### Update System
```bash
# Pull latest code
git pull

# Rebuild and restart
docker-compose up -d --build
```

### Backup Data
```bash
# Backup database
docker-compose exec postgres pg_dump -U admin postgres > backup.sql

# Backup volumes
docker run --rm -v owlistic_postgres_data:/data -v $(pwd):/backup ubuntu tar czf /backup/postgres_backup.tar.gz /data
```

### Clean Up
```bash
# Stop and remove containers
docker-compose down

# Remove volumes (WARNING: deletes all data)
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

Your AI-enhanced Owlistic system is now ready to run in Docker with network access! 🦉🧠🐳
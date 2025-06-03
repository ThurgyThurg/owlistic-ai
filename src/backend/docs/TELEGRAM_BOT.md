# Telegram Bot Integration

The Owlistic AI Telegram Bot allows you to interact with your note-taking system directly through Telegram messages. The bot uses AI to automatically classify your messages and route them to the appropriate handlers.

## Setup

### Environment Variables

Make sure these variables are set in your `.env` file:

```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here
ANTHROPIC_API_KEY=your_anthropic_key_here
```

### Getting Your Bot Token

1. Start a chat with [@BotFather](https://t.me/botfather) on Telegram
2. Send `/newbot` and follow the instructions
3. Copy the bot token to your `.env` file

### Getting Your Chat ID

1. Start a chat with your bot
2. Send any message
3. Visit `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
4. Look for the `"chat":{"id":` field in the response
5. Copy the chat ID to your `.env` file

## Message Classification

The bot uses AI to classify your messages into four categories:

### 📅 Calendar Events
**Intent**: Adding events, meetings, appointments, or time-based activities

**Example messages**:
- "Meeting with John at 3pm tomorrow"
- "Doctor appointment on Friday at 10am"
- "Team standup every Monday 9am"
- "Call mom at 7pm"

**What happens**: Creates a task with calendar metadata (full calendar integration coming soon)

### ✅ Tasks
**Intent**: Creating to-do items, reminders, or action items

**Example messages**:
- "Remember to buy groceries"
- "Call the dentist to schedule appointment"
- "Fix the broken link on website"
- "Email the client about project update"

**What happens**: Creates a task linked to a note in your "📱 Telegram Messages" notebook

### 🚀 Projects
**Intent**: Starting complex goals that need to be broken down into steps

**Example messages**:
- "I want to build a mobile app for tracking expenses"
- "Learn machine learning and build a recommendation system"
- "Create a home automation system with IoT devices"
- "Plan and execute a marketing campaign for product launch"

**What happens**: 
- Uses AI to break down the goal into manageable steps
- Creates an AI Project with full task breakdown
- Generates a dedicated notebook with notes for each step
- Creates actual tasks for deliverables

### 📝 Notes
**Intent**: General information, thoughts, or miscellaneous content

**Example messages**:
- "Interesting article about quantum computing trends"
- "Recipe idea: pasta with garlic and olive oil"
- "Thoughts on today's team meeting"
- "Book recommendation: The Pragmatic Programmer"

**What happens**: 
- Creates a note in your "📱 Telegram Messages" notebook
- Triggers AI processing for enhanced insights, tags, and action steps

## API Endpoints

### Send Notification (Manual)
```http
POST /api/v1/telegram/send-notification
Authorization: Bearer <your_jwt_token>
Content-Type: application/json

{
  "message": "Your notification message here"
}
```

### Get Bot Status
```http
GET /api/v1/telegram/status
Authorization: Bearer <your_jwt_token>
```

### Webhook (for production deployment)
```http
POST /api/v1/telegram/webhook
Content-Type: application/json

{
  "update_id": 123456789,
  "message": {
    "message_id": 1,
    "from": {...},
    "chat": {...},
    "text": "Your message"
  }
}
```

## How It Works

1. **Message Received**: Bot receives your Telegram message
2. **AI Classification**: Uses Anthropic Claude to analyze intent and extract data
3. **Route to Handler**: Directs to appropriate handler based on classification
4. **Create Content**: Creates tasks, projects, notes, or calendar events
5. **AI Enhancement**: For notes and projects, triggers additional AI processing
6. **Confirmation**: Sends back confirmation with created item details

## Features

### Intelligent Classification
- High-accuracy intent detection using advanced AI
- Fallback classification using keyword matching
- Confidence scoring and reasoning provided

### Seamless Integration
- Creates content in existing notebook structure
- Links tasks to notes for full context
- Maintains relationships between projects, notes, and tasks

### AI-Powered Enhancements
- Automatic note processing with insights, tags, and action steps
- Project breakdown into manageable steps with deliverables
- Smart title generation and metadata extraction

### Security
- Chat ID verification to prevent unauthorized access
- User authentication for API endpoints
- Secure token handling

## Troubleshooting

### Bot Not Responding
1. Check that `TELEGRAM_BOT_TOKEN` is correct
2. Verify `TELEGRAM_CHAT_ID` matches your chat
3. Ensure the bot has been started (`/start` command)
4. Check server logs for error messages

### Wrong Classifications
The AI classification is generally accurate, but you can:
1. Be more specific in your messages
2. Use action words for tasks ("buy", "call", "email")
3. Include time references for calendar events
4. Describe complexity for projects

### Missing Environment Variables
```
Failed to initialize Telegram service: TELEGRAM_BOT_TOKEN environment variable not set
```
Ensure both `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` are set in your `.env` file.

## Example Workflow

1. **Send message**: "I want to learn React and build a portfolio website"
2. **AI classifies**: Project (high confidence)
3. **AI breaks down**: Creates 6-8 steps like "Learn React basics", "Set up development environment", etc.
4. **Creates project**: AI Project with dedicated notebook
5. **Generates tasks**: Creates actual tasks for each deliverable
6. **Confirmation**: "🚀 Project created: Learn React and build a portfolio website\n📊 Broken down into 7 steps\n📓 Notebook ID: abc-123"

The Telegram bot makes it incredibly easy to capture ideas, tasks, and goals on the go, with AI automatically organizing them into your productivity system.
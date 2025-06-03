# Google Calendar Integration

The Owlistic AI application now includes full Google Calendar integration, allowing users to sync their calendar events and create calendar entries directly from tasks, notes, or Telegram messages.

## Setup

### Environment Variables

Add these variables to your `.env` file:

```env
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret  
GOOGLE_REDIRECT_URI=http://localhost:8080/api/v1/calendar/oauth/callback
```

### Google Cloud Console Setup

1. **Create a Google Cloud Project**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select existing one

2. **Enable Google Calendar API**:
   - Navigate to "APIs & Services" > "Library"
   - Search for "Google Calendar API"
   - Click "Enable"

3. **Create OAuth 2.0 Credentials**:
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth 2.0 Client IDs"
   - Application type: "Web application"
   - Add redirect URI: `http://localhost:8080/api/v1/calendar/oauth/callback`
   - Copy Client ID and Client Secret to your `.env` file

## API Endpoints

### OAuth Authentication

#### Get Authorization URL
```http
GET /api/v1/calendar/oauth/authorize
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "auth_url": "https://accounts.google.com/oauth/authorize?...",
  "message": "Visit this URL to authorize Google Calendar access"
}
```

#### OAuth Callback (handled automatically)
```http
GET /api/v1/calendar/oauth/callback?code=...&state=...
```

#### Check OAuth Status
```http
GET /api/v1/calendar/oauth/status
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "status": "connected",
  "has_access": true
}
```

#### Revoke Access
```http
DELETE /api/v1/calendar/oauth/revoke
Authorization: Bearer <jwt_token>
```

### Calendar Management

#### List User's Calendars
```http
GET /api/v1/calendar/calendars
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "calendars": [
    {
      "id": "primary",
      "summary": "user@example.com",
      "description": "Primary calendar",
      "timeZone": "America/New_York"
    }
  ],
  "count": 1
}
```

#### Setup Calendar Sync
```http
POST /api/v1/calendar/calendars/{calendar_id}/sync
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "calendar_name": "Primary Calendar",
  "sync_direction": "bidirectional"
}
```

**Sync Directions**:
- `read_only`: Only sync events from Google Calendar to Owlistic
- `write_only`: Only sync events from Owlistic to Google Calendar  
- `bidirectional`: Full two-way sync (default)

#### Get Sync Status
```http
GET /api/v1/calendar/sync-status
Authorization: Bearer <jwt_token>
```

### Event Management

#### Create Calendar Event
```http
POST /api/v1/calendar/events
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "title": "Team Meeting",
  "description": "Weekly team standup",
  "location": "Conference Room A",
  "start_time": "2024-01-15T10:00:00Z",
  "end_time": "2024-01-15T11:00:00Z",
  "all_day": false,
  "time_zone": "America/New_York",
  "calendar_id": "primary",
  "note_id": "optional-note-uuid",
  "task_id": "optional-task-uuid"
}
```

#### Get Calendar Events
```http
GET /api/v1/calendar/events?start_time=2024-01-01T00:00:00Z&end_time=2024-01-31T23:59:59Z
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "events": [
    {
      "id": "event-uuid",
      "google_event_id": "google-event-id",
      "title": "Team Meeting",
      "description": "Weekly team standup",
      "start_time": "2024-01-15T10:00:00Z",
      "end_time": "2024-01-15T11:00:00Z",
      "all_day": false,
      "status": "confirmed",
      "source": "owlistic"
    }
  ],
  "count": 1
}
```

#### Update Calendar Event
```http
PUT /api/v1/calendar/events/{event_id}
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "title": "Updated Meeting Title",
  "start_time": "2024-01-15T10:30:00Z",
  "end_time": "2024-01-15T11:30:00Z"
}
```

#### Delete Calendar Event
```http
DELETE /api/v1/calendar/events/{event_id}
Authorization: Bearer <jwt_token>
```

#### Manual Sync
```http
POST /api/v1/calendar/sync
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "calendar_id": "primary"
}
```

Leave `calendar_id` empty to sync all calendars.

## Database Models

### GoogleCalendarCredentials
Stores OAuth tokens for Google Calendar access:
- `access_token`: Current access token
- `refresh_token`: Refresh token for obtaining new access tokens
- `expires_at`: Token expiration time
- Auto-refresh when tokens expire

### CalendarEvent
Local representation of calendar events:
- Links to Google Calendar events via `google_event_id`
- Can be linked to notes and tasks
- Stores metadata and sync information

### CalendarSync
Tracks sync configuration per calendar:
- `sync_direction`: read_only, write_only, bidirectional
- `sync_token`: For incremental sync with Google Calendar
- `last_sync_at`: Timestamp of last successful sync

## Telegram Integration

### Automatic Calendar Event Creation

When you send a message to the Telegram bot that's classified as a calendar event:

1. **With Google Calendar Connected**:
   ```
   "Meeting with John tomorrow at 3pm"
   ```
   Creates actual Google Calendar event:
   ```
   📅 Calendar event created successfully!

   **Meeting with John**
   📅 January 16, 2024 at 3:00 PM

   ✅ Added to your Google Calendar
   🔗 Event ID: abc-123-def
   ```

2. **Without Google Calendar**:
   ```
   📅 Calendar event detected, but you haven't connected your Google Calendar yet.

   Use /api/v1/calendar/oauth/authorize to connect your calendar, then try again.

   For now, I'll save this as a task:

   📅 Calendar event saved as task: "Meeting with John"
   📝 Note ID: xyz-789
   ```

### Smart Date/Time Parsing

The system intelligently parses date and time information:
- **"tomorrow at 3pm"** → Sets event for next day at 3:00 PM
- **"meeting Friday morning"** → Sets event for Friday at 9:00 AM
- **"lunch today"** → Sets event for today at default time
- **"all day conference Tuesday"** → Sets all-day event for Tuesday

## Features

### Bidirectional Sync
- Events created in Google Calendar appear in Owlistic
- Events created in Owlistic appear in Google Calendar
- Updates and deletions sync in both directions

### Intelligent Classification
- AI automatically detects calendar events from natural language
- Extracts title, date, time, and duration when possible
- Falls back to task creation when calendar isn't connected

### Task/Note Integration
- Calendar events can be linked to notes and tasks
- Create events from task due dates
- Generate tasks from calendar events

### Incremental Sync
- Uses Google Calendar sync tokens for efficient updates
- Only syncs changed events, not entire calendar
- Configurable sync frequency

## Security

### OAuth 2.0 Flow
- Secure OAuth 2.0 authorization with Google
- Tokens stored encrypted in database
- Automatic token refresh when expired

### User Isolation
- Each user's calendar data is completely isolated
- Users can only access their own events and calendars
- Proper authentication required for all endpoints

### Data Privacy
- Calendar data synchronized securely
- Can revoke access at any time
- Local calendar events deleted when access revoked

## Troubleshooting

### OAuth Issues
```
"missing Google OAuth credentials in environment variables"
```
- Ensure `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `GOOGLE_REDIRECT_URI` are set
- Verify redirect URI matches Google Cloud Console configuration

### Calendar Access
```
"no calendar credentials found for user"
```
- User needs to complete OAuth flow first
- Check OAuth status endpoint

### Sync Problems
```
"failed to sync calendar"
```
- Check if access token is still valid
- Verify calendar ID exists and user has access
- Check Google Calendar API quotas

### Telegram Calendar Events
- If calendar creation fails, events are saved as tasks
- Connect Google Calendar for full calendar integration
- Check that date/time parsing works correctly

## Example Usage Flow

1. **Setup Google Calendar**:
   - Configure environment variables
   - User visits authorization URL
   - Complete OAuth flow

2. **Enable Calendar Sync**:
   - List available calendars
   - Choose calendars to sync
   - Set sync direction

3. **Create Events**:
   - Via API endpoints
   - Through Telegram messages
   - From tasks or notes

4. **Sync Events**:
   - Automatic background sync
   - Manual sync triggers
   - Real-time updates

The Google Calendar integration provides a seamless way to manage calendar events within your Owlistic AI workflow, with intelligent AI-powered event creation and full bidirectional synchronization.
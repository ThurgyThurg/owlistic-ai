# Google Calendar OAuth Setup

## Environment Variables Required

Add these to your `.env` file in the backend directory:

```env
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
```

**Note**: `GOOGLE_REDIRECT_URI` is optional. If not provided, it defaults to `http://localhost:8080/api/calendar/oauth/callback` (or uses the PORT environment variable if set).

## Getting the Redirect URI

The redirect URI is automatically determined by your backend. To see what URI to use:

1. Start your backend server
2. Visit: `http://localhost:8080/api/calendar/oauth/config`
3. Use the `redirect_uri` from the response

Or check the backend logs when starting - it will show the redirect URI being used.

## Google Cloud Console Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **APIs & Services** → **Credentials**
3. Find your OAuth 2.0 Client ID (or create one if needed)
4. Click **Edit** on your OAuth 2.0 Client ID
5. In the **Authorized redirect URIs** section, add the redirect URI from above
6. Click **Save**

## Required APIs

Make sure these APIs are enabled in your Google Cloud Project:
- Google Calendar API
- Google+ API (for user info)

## OAuth Flow

1. User clicks "Connect Google Calendar" in app settings
2. App opens browser with Google OAuth URL
3. User authorizes the application
4. Google redirects to `http://localhost:8080/api/calendar/oauth/callback`
5. Backend automatically exchanges code for access tokens
6. User returns to app and sees connected status

## Troubleshooting

- **redirect_uri_mismatch**: The redirect URI in your Google Cloud Console doesn't match the one in your .env file
- **invalid_client**: Check your GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET
- **access_denied**: User cancelled the authorization process

## Production Setup

For production, you have two options:

### Option 1: Set DOMAIN environment variable (Recommended)
```env
DOMAIN=secondbrain.graham29.com
```
This will automatically use `https://secondbrain.graham29.com/api/calendar/oauth/callback`

### Option 2: Set GOOGLE_REDIRECT_URI explicitly
```env
GOOGLE_REDIRECT_URI=https://secondbrain.graham29.com/api/calendar/oauth/callback
```

### Your Google Cloud Console Setup
Add this URL to your Google Cloud Console authorized redirect URIs:
```
https://secondbrain.graham29.com/api/calendar/oauth/callback
```

## Checking Current Configuration

Visit `https://secondbrain.graham29.com/api/calendar/oauth/config` to see the current redirect URI configuration.
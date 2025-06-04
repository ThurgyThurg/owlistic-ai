package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"

	"owlistic-notes/owlistic/models"
	"github.com/google/uuid"
)

type CalendarService struct {
	db           *gorm.DB
	oauth2Config *oauth2.Config
}

type CalendarEventRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	AllDay      bool      `json:"all_day"`
	TimeZone    string    `json:"time_zone"`
	CalendarID  string    `json:"calendar_id"` // Google Calendar ID
	NoteID      *string   `json:"note_id,omitempty"`
	TaskID      *string   `json:"task_id,omitempty"`
}

func NewCalendarService(db *gorm.DB) (*CalendarService, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URI")

	log.Printf("Calendar Service Init - GOOGLE_CLIENT_ID length: %d", len(clientID))
	log.Printf("Calendar Service Init - GOOGLE_CLIENT_SECRET length: %d", len(clientSecret))
	
	// Allow service creation even without credentials for config endpoint
	// OAuth operations will fail but config endpoint will work
	var oauth2Config *oauth2.Config
	if clientID != "" && clientSecret != "" {
		oauth2Config = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				calendar.CalendarScope, // Full access to calendars
			},
			Endpoint: google.Endpoint,
		}
	} else {
		log.Printf("Warning: Google OAuth credentials not configured. Calendar OAuth features will not work.")
		log.Printf("Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET to enable Google Calendar integration.")
	}

	// Use default redirect URI if not provided
	if redirectURL == "" {
		// Check for production environment indicators
		domain := os.Getenv("DOMAIN")
		if domain != "" {
			// Production environment
			redirectURL = fmt.Sprintf("https://%s/api/calendar/oauth/callback", domain)
		} else {
			// Development environment
			serverPort := os.Getenv("PORT")
			if serverPort == "" {
				serverPort = "8080"
			}
			redirectURL = fmt.Sprintf("http://localhost:%s/api/calendar/oauth/callback", serverPort)
		}
		log.Printf("Using default Google OAuth redirect URI: %s", redirectURL)
		log.Printf("Make sure to add this URL to your Google Cloud Console OAuth 2.0 Client ID authorized redirect URIs")
		
		// Update oauth2Config with redirect URL if it exists
		if oauth2Config != nil {
			oauth2Config.RedirectURL = redirectURL
		}
	}

	return &CalendarService{
		db:           db,
		oauth2Config: oauth2Config,
	}, nil
}

// GetAuthURL generates the OAuth2 authorization URL
func (cs *CalendarService) GetAuthURL(userID uuid.UUID) string {
	if cs.oauth2Config == nil {
		log.Printf("Error: OAuth2 config not initialized. Missing Google OAuth credentials.")
		return ""
	}
	// Use user ID as state to prevent CSRF attacks
	state := userID.String()
	return cs.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCodeForTokens exchanges authorization code for access tokens
func (cs *CalendarService) ExchangeCodeForTokens(ctx context.Context, userID uuid.UUID, code string) error {
	if cs.oauth2Config == nil {
		return fmt.Errorf("OAuth2 config not initialized. Missing Google OAuth credentials")
	}
	token, err := cs.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Save credentials to database
	credentials := models.GoogleCalendarCredentials{
		UserID:       userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
		Scope:        "https://www.googleapis.com/auth/calendar",
	}

	// Delete any existing credentials for this user
	cs.db.Where("user_id = ?", userID).Delete(&models.GoogleCalendarCredentials{})

	// Save new credentials
	if err := cs.db.Create(&credentials).Error; err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	log.Printf("Successfully saved Google Calendar credentials for user %s", userID)
	return nil
}

// GetCalendarClient creates an authenticated Google Calendar client
func (cs *CalendarService) GetCalendarClient(ctx context.Context, userID uuid.UUID) (*calendar.Service, error) {
	credentials, err := cs.getValidCredentials(ctx, userID)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  credentials.AccessToken,
		RefreshToken: credentials.RefreshToken,
		TokenType:    credentials.TokenType,
		Expiry:       credentials.ExpiresAt,
	}

	if cs.oauth2Config == nil {
		return nil, fmt.Errorf("OAuth2 config not initialized. Missing Google OAuth credentials")
	}
	client := cs.oauth2Config.Client(ctx, token)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return service, nil
}

// getValidCredentials gets valid credentials, refreshing if necessary
func (cs *CalendarService) getValidCredentials(ctx context.Context, userID uuid.UUID) (*models.GoogleCalendarCredentials, error) {
	var credentials models.GoogleCalendarCredentials
	if err := cs.db.Where("user_id = ?", userID).First(&credentials).Error; err != nil {
		return nil, fmt.Errorf("no calendar credentials found for user: %w", err)
	}

	// Check if token needs refresh
	if credentials.NeedsRefresh() {
		if err := cs.refreshToken(ctx, &credentials); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	return &credentials, nil
}

// refreshToken refreshes the access token using the refresh token
func (cs *CalendarService) refreshToken(ctx context.Context, credentials *models.GoogleCalendarCredentials) error {
	token := &oauth2.Token{
		AccessToken:  credentials.AccessToken,
		RefreshToken: credentials.RefreshToken,
		TokenType:    credentials.TokenType,
		Expiry:       credentials.ExpiresAt,
	}

	if cs.oauth2Config == nil {
		return fmt.Errorf("OAuth2 config not initialized. Missing Google OAuth credentials")
	}
	tokenSource := cs.oauth2Config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return err
	}

	// Update credentials in database
	return credentials.UpdateTokens(cs.db, newToken.AccessToken, newToken.RefreshToken, newToken.Expiry)
}

// ListCalendars retrieves the user's Google Calendars
func (cs *CalendarService) ListCalendars(ctx context.Context, userID uuid.UUID) ([]*calendar.CalendarListEntry, error) {
	service, err := cs.GetCalendarClient(ctx, userID)
	if err != nil {
		return nil, err
	}

	calendarList, err := service.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}

	return calendarList.Items, nil
}

// SyncCalendar sets up sync for a specific Google Calendar
func (cs *CalendarService) SyncCalendar(ctx context.Context, userID uuid.UUID, googleCalendarID, calendarName string, syncDirection string) error {
	// Create or update calendar sync record
	sync := models.CalendarSync{
		UserID:           userID,
		GoogleCalendarID: googleCalendarID,
		CalendarName:     calendarName,
		SyncEnabled:      true,
		SyncDirection:    syncDirection,
	}

	// Try to find existing sync record
	var existing models.CalendarSync
	err := cs.db.Where("user_id = ? AND google_calendar_id = ?", userID, googleCalendarID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		// Create new sync record
		if err := cs.db.Create(&sync).Error; err != nil {
			return fmt.Errorf("failed to create calendar sync: %w", err)
		}
	} else if err != nil {
		return err
	} else {
		// Update existing sync record
		existing.CalendarName = calendarName
		existing.SyncDirection = syncDirection
		existing.SyncEnabled = true
		if err := cs.db.Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update calendar sync: %w", err)
		}
	}

	// Perform initial sync
	return cs.PerformSync(ctx, userID, googleCalendarID)
}

// PerformSync syncs events between Google Calendar and our database
func (cs *CalendarService) PerformSync(ctx context.Context, userID uuid.UUID, googleCalendarID string) error {
	service, err := cs.GetCalendarClient(ctx, userID)
	if err != nil {
		return err
	}

	// Get sync record
	var sync models.CalendarSync
	if err := cs.db.Where("user_id = ? AND google_calendar_id = ?", userID, googleCalendarID).First(&sync).Error; err != nil {
		return fmt.Errorf("calendar sync not found: %w", err)
	}

	// Sync events from the last 30 days to next 365 days
	timeMin := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	timeMax := time.Now().AddDate(1, 0, 0).Format(time.RFC3339)

	eventsCall := service.Events.List(googleCalendarID).
		TimeMin(timeMin).
		TimeMax(timeMax).
		SingleEvents(true).
		OrderBy("startTime")

	// Use sync token for incremental sync if available
	if sync.SyncToken != "" {
		eventsCall = eventsCall.SyncToken(sync.SyncToken)
	}

	events, err := eventsCall.Do()
	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	// Process each event
	for _, event := range events.Items {
		if err := cs.syncEvent(userID, googleCalendarID, event); err != nil {
			log.Printf("Failed to sync event %s: %v", event.Id, err)
			continue
		}
	}

	// Update sync record
	now := time.Now()
	sync.LastSyncAt = &now
	if events.NextSyncToken != "" {
		sync.SyncToken = events.NextSyncToken
	}
	cs.db.Save(&sync)

	log.Printf("Successfully synced %d events for calendar %s", len(events.Items), googleCalendarID)
	return nil
}

// syncEvent syncs a single event from Google Calendar
func (cs *CalendarService) syncEvent(userID uuid.UUID, googleCalendarID string, googleEvent *calendar.Event) error {
	// Skip cancelled events
	if googleEvent.Status == "cancelled" {
		// Delete from our database if it exists
		cs.db.Where("user_id = ? AND google_event_id = ?", userID, googleEvent.Id).Delete(&models.CalendarEvent{})
		return nil
	}

	// Parse start and end times
	startTime, err := cs.parseEventTime(googleEvent.Start)
	if err != nil {
		return fmt.Errorf("failed to parse start time: %w", err)
	}

	endTime, err := cs.parseEventTime(googleEvent.End)
	if err != nil {
		return fmt.Errorf("failed to parse end time: %w", err)
	}

	// Determine if it's an all-day event
	allDay := googleEvent.Start.Date != ""

	// Create or update calendar event
	event := models.CalendarEvent{
		UserID:           userID,
		GoogleEventID:    googleEvent.Id,
		GoogleCalendarID: googleCalendarID,
		Title:            googleEvent.Summary,
		Description:      googleEvent.Description,
		Location:         googleEvent.Location,
		StartTime:        startTime,
		EndTime:          endTime,
		AllDay:           allDay,
		TimeZone:         cs.getEventTimeZone(googleEvent),
		Status:           googleEvent.Status,
		Visibility:       googleEvent.Visibility,
		Source:           "google",
		Metadata: models.CalendarEventMetadata{
			"google_event":    googleEvent,
			"html_link":       googleEvent.HtmlLink,
			"ical_uid":        googleEvent.ICalUID,
			"sequence":        googleEvent.Sequence,
			"hangout_link":    googleEvent.HangoutLink,
			"conference_data": googleEvent.ConferenceData,
		},
	}

	return models.CreateOrUpdateEvent(cs.db, &event)
}

// parseEventTime parses Google Calendar event time
func (cs *CalendarService) parseEventTime(eventTime *calendar.EventDateTime) (time.Time, error) {
	if eventTime.DateTime != "" {
		return time.Parse(time.RFC3339, eventTime.DateTime)
	} else if eventTime.Date != "" {
		return time.Parse("2006-01-02", eventTime.Date)
	}
	return time.Time{}, fmt.Errorf("no valid time found")
}

// getEventTimeZone gets the timezone from the event or defaults to UTC
func (cs *CalendarService) getEventTimeZone(googleEvent *calendar.Event) string {
	if googleEvent.Start != nil && googleEvent.Start.TimeZone != "" {
		return googleEvent.Start.TimeZone
	}
	return "UTC"
}

// CreateEvent creates a new event in Google Calendar
func (cs *CalendarService) CreateEvent(ctx context.Context, userID uuid.UUID, req CalendarEventRequest) (*models.CalendarEvent, error) {
	service, err := cs.GetCalendarClient(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Default calendar ID if not provided
	calendarID := req.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}

	// Create Google Calendar event
	googleEvent := &calendar.Event{
		Summary:     req.Title,
		Description: req.Description,
		Location:    req.Location,
	}

	// Set event times
	if req.AllDay {
		googleEvent.Start = &calendar.EventDateTime{
			Date: req.StartTime.Format("2006-01-02"),
		}
		googleEvent.End = &calendar.EventDateTime{
			Date: req.EndTime.Format("2006-01-02"),
		}
	} else {
		timeZone := req.TimeZone
		if timeZone == "" {
			timeZone = "UTC"
		}
		googleEvent.Start = &calendar.EventDateTime{
			DateTime: req.StartTime.Format(time.RFC3339),
			TimeZone: timeZone,
		}
		googleEvent.End = &calendar.EventDateTime{
			DateTime: req.EndTime.Format(time.RFC3339),
			TimeZone: timeZone,
		}
	}

	// Create event in Google Calendar
	createdEvent, err := service.Events.Insert(calendarID, googleEvent).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Calendar event: %w", err)
	}

	// Parse UUIDs if provided
	var noteID, taskID *uuid.UUID
	if req.NoteID != nil {
		if parsed, err := uuid.Parse(*req.NoteID); err == nil {
			noteID = &parsed
		}
	}
	if req.TaskID != nil {
		if parsed, err := uuid.Parse(*req.TaskID); err == nil {
			taskID = &parsed
		}
	}

	// Create local calendar event record
	event := models.CalendarEvent{
		UserID:           userID,
		GoogleEventID:    createdEvent.Id,
		GoogleCalendarID: calendarID,
		Title:            req.Title,
		Description:      req.Description,
		Location:         req.Location,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		AllDay:           req.AllDay,
		TimeZone:         req.TimeZone,
		Status:           "confirmed",
		Source:           "owlistic",
		NoteID:           noteID,
		TaskID:           taskID,
		Metadata: models.CalendarEventMetadata{
			"created_via": "api",
			"html_link":   createdEvent.HtmlLink,
		},
	}

	if err := cs.db.Create(&event).Error; err != nil {
		return nil, fmt.Errorf("failed to save calendar event: %w", err)
	}

	return &event, nil
}

// GetEvents retrieves calendar events for a user within a date range
func (cs *CalendarService) GetEvents(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]models.CalendarEvent, error) {
	return models.GetCalendarEvents(cs.db, userID, startTime, endTime)
}

// UpdateEvent updates an existing calendar event
func (cs *CalendarService) UpdateEvent(ctx context.Context, userID uuid.UUID, eventID uuid.UUID, req CalendarEventRequest) (*models.CalendarEvent, error) {
	// Get existing event
	var event models.CalendarEvent
	if err := cs.db.Where("id = ? AND user_id = ?", eventID, userID).First(&event).Error; err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Get Google Calendar service
	service, err := cs.GetCalendarClient(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update Google Calendar event
	googleEvent := &calendar.Event{
		Summary:     req.Title,
		Description: req.Description,
		Location:    req.Location,
	}

	// Set event times
	if req.AllDay {
		googleEvent.Start = &calendar.EventDateTime{
			Date: req.StartTime.Format("2006-01-02"),
		}
		googleEvent.End = &calendar.EventDateTime{
			Date: req.EndTime.Format("2006-01-02"),
		}
	} else {
		timeZone := req.TimeZone
		if timeZone == "" {
			timeZone = event.TimeZone
		}
		googleEvent.Start = &calendar.EventDateTime{
			DateTime: req.StartTime.Format(time.RFC3339),
			TimeZone: timeZone,
		}
		googleEvent.End = &calendar.EventDateTime{
			DateTime: req.EndTime.Format(time.RFC3339),
			TimeZone: timeZone,
		}
	}

	// Update in Google Calendar
	_, err = service.Events.Update(event.GoogleCalendarID, event.GoogleEventID, googleEvent).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update Google Calendar event: %w", err)
	}

	// Update local record
	event.Title = req.Title
	event.Description = req.Description
	event.Location = req.Location
	event.StartTime = req.StartTime
	event.EndTime = req.EndTime
	event.AllDay = req.AllDay
	if req.TimeZone != "" {
		event.TimeZone = req.TimeZone
	}

	if err := cs.db.Save(&event).Error; err != nil {
		return nil, fmt.Errorf("failed to update calendar event: %w", err)
	}

	return &event, nil
}

// DeleteEvent deletes a calendar event
func (cs *CalendarService) DeleteEvent(ctx context.Context, userID uuid.UUID, eventID uuid.UUID) error {
	// Get existing event
	var event models.CalendarEvent
	if err := cs.db.Where("id = ? AND user_id = ?", eventID, userID).First(&event).Error; err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	// Get Google Calendar service
	service, err := cs.GetCalendarClient(ctx, userID)
	if err != nil {
		return err
	}

	// Delete from Google Calendar
	err = service.Events.Delete(event.GoogleCalendarID, event.GoogleEventID).Do()
	if err != nil {
		log.Printf("Failed to delete event from Google Calendar: %v", err)
		// Continue with local deletion even if Google deletion fails
	}

	// Delete local record
	return cs.db.Delete(&event).Error
}

// RevokeAccess revokes Google Calendar access for a user
func (cs *CalendarService) RevokeAccess(ctx context.Context, userID uuid.UUID) error {
	// Delete credentials from database
	if err := cs.db.Where("user_id = ?", userID).Delete(&models.GoogleCalendarCredentials{}).Error; err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	// Delete sync configurations
	if err := cs.db.Where("user_id = ?", userID).Delete(&models.CalendarSync{}).Error; err != nil {
		return fmt.Errorf("failed to delete sync configurations: %w", err)
	}

	log.Printf("Revoked Google Calendar access for user %s", userID)
	return nil
}

// HasCalendarAccess checks if a user has valid Google Calendar access
func (cs *CalendarService) HasCalendarAccess(ctx context.Context, userID uuid.UUID) bool {
	_, err := cs.getValidCredentials(ctx, userID)
	return err == nil
}
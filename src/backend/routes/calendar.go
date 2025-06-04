package routes

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"owlistic-notes/owlistic/models"
	"owlistic-notes/owlistic/services"
)

type CalendarRoutes struct {
	db              *gorm.DB
	calendarService *services.CalendarService
}

func NewCalendarRoutes(db *gorm.DB) (*CalendarRoutes, error) {
	calendarService, err := services.NewCalendarService(db)
	if err != nil {
		return nil, err
	}

	return &CalendarRoutes{
		db:              db,
		calendarService: calendarService,
	}, nil
}

func (cr *CalendarRoutes) RegisterRoutes(routerGroup *gin.RouterGroup) {
	calendarGroup := routerGroup.Group("/calendar")
	{
		// OAuth endpoints
		calendarGroup.GET("/oauth/authorize", cr.getAuthURL)
		calendarGroup.GET("/oauth/callback", cr.handleOAuthCallback)
		calendarGroup.DELETE("/oauth/revoke", cr.revokeAccess)
		calendarGroup.GET("/oauth/status", cr.getOAuthStatus)
		calendarGroup.GET("/oauth/config", cr.getOAuthConfig)

		// Calendar management
		calendarGroup.GET("/calendars", cr.listCalendars)
		calendarGroup.POST("/calendars/:id/sync", cr.syncCalendar)
		calendarGroup.GET("/sync-status", cr.getSyncStatus)

		// Event CRUD operations
		calendarGroup.POST("/events", cr.createEvent)
		calendarGroup.GET("/events", cr.getEvents)
		calendarGroup.GET("/events/:id", cr.getEvent)
		calendarGroup.PUT("/events/:id", cr.updateEvent)
		calendarGroup.DELETE("/events/:id", cr.deleteEvent)

		// Manual sync
		calendarGroup.POST("/sync", cr.performSync)
	}
}

// getAuthURL generates the OAuth2 authorization URL
func (cr *CalendarRoutes) getAuthURL(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	authURL := cr.calendarService.GetAuthURL(userUUID)
	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"message":  "Visit this URL to authorize Google Calendar access",
	})
}

// handleOAuthCallback handles the OAuth2 callback
func (cr *CalendarRoutes) handleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	if errorParam != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth authorization failed: " + errorParam})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	// Parse state as user ID
	userUUID, err := uuid.Parse(state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Exchange code for tokens
	if err := cr.calendarService.ExchangeCodeForTokens(c.Request.Context(), userUUID, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange authorization code: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Google Calendar access granted successfully",
		"user_id": userUUID,
	})
}

// revokeAccess revokes Google Calendar access
func (cr *CalendarRoutes) revokeAccess(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	if err := cr.calendarService.RevokeAccess(c.Request.Context(), userUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke access: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Google Calendar access revoked successfully"})
}

// getOAuthStatus checks the OAuth status for the user
func (cr *CalendarRoutes) getOAuthStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	hasAccess := cr.calendarService.HasCalendarAccess(c.Request.Context(), userUUID)
	
	status := "disconnected"
	if hasAccess {
		status = "connected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     status,
		"has_access": hasAccess,
	})
}

// listCalendars retrieves the user's Google Calendars
func (cr *CalendarRoutes) listCalendars(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	calendars, err := cr.calendarService.ListCalendars(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list calendars: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"calendars": calendars,
		"count":     len(calendars),
	})
}

// syncCalendar sets up sync for a specific calendar
func (cr *CalendarRoutes) syncCalendar(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	calendarID := c.Param("id")
	if calendarID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Calendar ID is required"})
		return
	}

	var request struct {
		CalendarName  string `json:"calendar_name" binding:"required"`
		SyncDirection string `json:"sync_direction"` // read_only, write_only, bidirectional
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default sync direction
	if request.SyncDirection == "" {
		request.SyncDirection = "bidirectional"
	}

	if err := cr.calendarService.SyncCalendar(c.Request.Context(), userUUID, calendarID, request.CalendarName, request.SyncDirection); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup calendar sync: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Calendar sync configured successfully",
		"calendar_id":    calendarID,
		"calendar_name":  request.CalendarName,
		"sync_direction": request.SyncDirection,
	})
}

// getSyncStatus gets the sync status for all calendars
func (cr *CalendarRoutes) getSyncStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var syncStatuses []models.CalendarSync
	if err := cr.db.Where("user_id = ?", userUUID).Find(&syncStatuses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sync status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sync_calendars": syncStatuses,
		"count":          len(syncStatuses),
	})
}

// createEvent creates a new calendar event
func (cr *CalendarRoutes) createEvent(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var request services.CalendarEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := cr.calendarService.CreateEvent(c.Request.Context(), userUUID, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// getEvents retrieves calendar events within a date range
func (cr *CalendarRoutes) getEvents(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Parse query parameters
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format. Use RFC3339 format."})
			return
		}
	} else {
		// Default to current month
		now := time.Now()
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format. Use RFC3339 format."})
			return
		}
	} else {
		// Default to end of current month
		startTime := startTime
		endTime = startTime.AddDate(0, 1, 0).Add(-time.Second)
	}

	events, err := cr.calendarService.GetEvents(c.Request.Context(), userUUID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get events: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":     events,
		"count":      len(events),
		"start_time": startTime,
		"end_time":   endTime,
	})
}

// getEvent retrieves a specific calendar event
func (cr *CalendarRoutes) getEvent(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.CalendarEvent
	if err := cr.db.Where("id = ? AND user_id = ?", eventID, userUUID).First(&event).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// updateEvent updates an existing calendar event
func (cr *CalendarRoutes) updateEvent(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var request services.CalendarEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := cr.calendarService.UpdateEvent(c.Request.Context(), userUUID, eventID, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// deleteEvent deletes a calendar event
func (cr *CalendarRoutes) deleteEvent(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := cr.calendarService.DeleteEvent(c.Request.Context(), userUUID, eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

// performSync manually triggers calendar sync
func (cr *CalendarRoutes) performSync(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var request struct {
		CalendarID string `json:"calendar_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.CalendarID == "" {
		// Sync all calendars
		var syncConfigs []models.CalendarSync
		if err := cr.db.Where("user_id = ? AND sync_enabled = true", userUUID).Find(&syncConfigs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sync configurations"})
			return
		}

		syncCount := 0
		for _, config := range syncConfigs {
			if err := cr.calendarService.PerformSync(c.Request.Context(), userUUID, config.GoogleCalendarID); err != nil {
				// Log error but continue with other calendars
				continue
			}
			syncCount++
		}

		c.JSON(http.StatusOK, gin.H{
			"message":           "Sync completed",
			"calendars_synced":  syncCount,
			"total_calendars":   len(syncConfigs),
		})
	} else {
		// Sync specific calendar
		if err := cr.calendarService.PerformSync(c.Request.Context(), userUUID, request.CalendarID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync calendar: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Calendar synced successfully",
			"calendar_id": request.CalendarID,
		})
	}
}

// getOAuthConfig returns the current OAuth configuration for setup purposes
func (cr *CalendarRoutes) getOAuthConfig(c *gin.Context) {
	// This endpoint doesn't require authentication as it's for setup purposes
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")
	
	// Use same default logic as calendar service
	if redirectURI == "" {
		// Check for production environment indicators
		domain := os.Getenv("DOMAIN")
		if domain != "" {
			// Production environment
			redirectURI = fmt.Sprintf("https://%s/api/calendar/oauth/callback", domain)
		} else {
			// Development environment
			serverPort := os.Getenv("PORT")
			if serverPort == "" {
				serverPort = "8080"
			}
			redirectURI = fmt.Sprintf("http://localhost:%s/api/calendar/oauth/callback", serverPort)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"redirect_uri": redirectURI,
		"setup_instructions": []string{
			"1. Go to Google Cloud Console → APIs & Services → Credentials",
			"2. Edit your OAuth 2.0 Client ID",
			"3. Add the redirect_uri above to 'Authorized redirect URIs'",
			"4. Save and ensure Google Calendar API is enabled",
		},
		"required_env_vars": []string{
			"GOOGLE_CLIENT_ID",
			"GOOGLE_CLIENT_SECRET", 
		},
		"optional_env_vars": map[string]string{
			"GOOGLE_REDIRECT_URI": "Override the default redirect URI",
			"PORT": "Server port (defaults to 8080)",
		},
	})
}
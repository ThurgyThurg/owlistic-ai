package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GoogleCalendarCredentials stores OAuth tokens for Google Calendar access
type GoogleCalendarCredentials struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"user_id"`
	AccessToken  string         `gorm:"type:text;not null" json:"access_token"`
	RefreshToken string         `gorm:"type:text;not null" json:"refresh_token"`
	TokenType    string         `gorm:"default:'Bearer'" json:"token_type"`
	ExpiresAt    time.Time      `gorm:"not null" json:"expires_at"`
	Scope        string         `gorm:"type:text" json:"scope"`
	CreatedAt    time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// CalendarEventMetadata stores additional metadata for calendar events
type CalendarEventMetadata map[string]interface{}

// Value implements the driver.Valuer interface for JSONB storage
func (cem CalendarEventMetadata) Value() (driver.Value, error) {
	if cem == nil {
		return nil, nil
	}
	return json.Marshal(cem)
}

// Scan implements the sql.Scanner interface for JSONB retrieval
func (cem *CalendarEventMetadata) Scan(value interface{}) error {
	if value == nil {
		*cem = make(CalendarEventMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, cem)
}

// CalendarEvent represents a calendar event in our system
type CalendarEvent struct {
	ID               uuid.UUID             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID             `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"user_id"`
	GoogleEventID    string                `gorm:"unique;not null" json:"google_event_id"`
	GoogleCalendarID string                `gorm:"not null" json:"google_calendar_id"`
	Title            string                `gorm:"not null" json:"title"`
	Description      string                `gorm:"type:text" json:"description"`
	Location         string                `json:"location"`
	StartTime        time.Time             `gorm:"not null" json:"start_time"`
	EndTime          time.Time             `gorm:"not null" json:"end_time"`
	AllDay           bool                  `gorm:"default:false" json:"all_day"`
	TimeZone         string                `gorm:"default:'UTC'" json:"time_zone"`
	Status           string                `gorm:"default:'confirmed'" json:"status"` // confirmed, tentative, cancelled
	Visibility       string                `gorm:"default:'default'" json:"visibility"` // default, public, private
	Recurrence       string                `gorm:"type:text" json:"recurrence,omitempty"` // RRULE string
	Source           string                `gorm:"default:'google'" json:"source"` // google, owlistic, telegram
	NoteID           *uuid.UUID            `gorm:"type:uuid" json:"note_id,omitempty"` // Link to related note
	TaskID           *uuid.UUID            `gorm:"type:uuid" json:"task_id,omitempty"` // Link to related task
	Metadata         CalendarEventMetadata `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time             `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time             `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt        gorm.DeletedAt        `gorm:"index" json:"deleted_at,omitempty"`
}

// CalendarSync tracks synchronization status with Google Calendar
type CalendarSync struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"user_id"`
	GoogleCalendarID string         `gorm:"not null" json:"google_calendar_id"`
	CalendarName     string         `gorm:"not null" json:"calendar_name"`
	SyncToken        string         `gorm:"type:text" json:"sync_token,omitempty"` // For incremental sync
	LastSyncAt       *time.Time     `json:"last_sync_at,omitempty"`
	SyncEnabled      bool           `gorm:"default:true" json:"sync_enabled"`
	SyncDirection    string         `gorm:"default:'bidirectional'" json:"sync_direction"` // read_only, write_only, bidirectional
	CreatedAt        time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// CalendarReminder represents reminders for calendar events
type CalendarReminder struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EventID   uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"event_id"`
	Method    string         `gorm:"not null" json:"method"` // email, popup, sms
	Minutes   int            `gorm:"not null" json:"minutes"` // minutes before event
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// CalendarAttendee represents attendees for calendar events
type CalendarAttendee struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EventID         uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"event_id"`
	Email           string         `gorm:"not null" json:"email"`
	DisplayName     string         `json:"display_name"`
	ResponseStatus  string         `gorm:"default:'needsAction'" json:"response_status"` // needsAction, declined, tentative, accepted
	Optional        bool           `gorm:"default:false" json:"optional"`
	Organizer       bool           `gorm:"default:false" json:"organizer"`
	Self            bool           `gorm:"default:false" json:"self"`
	CreatedAt       time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// GetActiveCredentials gets the active Google Calendar credentials for a user
func (gc *GoogleCalendarCredentials) GetActiveCredentials(db *gorm.DB, userID uuid.UUID) (*GoogleCalendarCredentials, error) {
	var credentials GoogleCalendarCredentials
	err := db.Where("user_id = ?", userID).First(&credentials).Error
	if err != nil {
		return nil, err
	}
	return &credentials, nil
}

// IsExpired checks if the access token is expired
func (gc *GoogleCalendarCredentials) IsExpired() bool {
	return time.Now().After(gc.ExpiresAt)
}

// NeedsRefresh checks if the token needs to be refreshed (expires within 5 minutes)
func (gc *GoogleCalendarCredentials) NeedsRefresh() bool {
	return time.Now().Add(5 * time.Minute).After(gc.ExpiresAt)
}

// UpdateTokens updates the access and refresh tokens
func (gc *GoogleCalendarCredentials) UpdateTokens(db *gorm.DB, accessToken, refreshToken string, expiresAt time.Time) error {
	gc.AccessToken = accessToken
	if refreshToken != "" {
		gc.RefreshToken = refreshToken
	}
	gc.ExpiresAt = expiresAt
	return db.Save(gc).Error
}

// GetUserCalendars gets all calendar sync configurations for a user
func GetUserCalendars(db *gorm.DB, userID uuid.UUID) ([]CalendarSync, error) {
	var calendars []CalendarSync
	err := db.Where("user_id = ? AND sync_enabled = true", userID).Find(&calendars).Error
	return calendars, err
}

// GetCalendarEvents gets calendar events for a user within a date range
func GetCalendarEvents(db *gorm.DB, userID uuid.UUID, startTime, endTime time.Time) ([]CalendarEvent, error) {
	var events []CalendarEvent
	err := db.Where("user_id = ? AND start_time >= ? AND start_time <= ?", userID, startTime, endTime).
		Order("start_time ASC").Find(&events).Error
	return events, err
}

// GetEventByGoogleID finds a calendar event by its Google event ID
func GetEventByGoogleID(db *gorm.DB, userID uuid.UUID, googleEventID string) (*CalendarEvent, error) {
	var event CalendarEvent
	err := db.Where("user_id = ? AND google_event_id = ?", userID, googleEventID).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// CreateOrUpdateEvent creates a new event or updates an existing one
func CreateOrUpdateEvent(db *gorm.DB, event *CalendarEvent) error {
	// Try to find existing event by Google ID
	var existing CalendarEvent
	err := db.Where("user_id = ? AND google_event_id = ?", event.UserID, event.GoogleEventID).First(&existing).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new event
		return db.Create(event).Error
	} else if err != nil {
		return err
	} else {
		// Update existing event
		event.ID = existing.ID
		event.CreatedAt = existing.CreatedAt
		return db.Save(event).Error
	}
}
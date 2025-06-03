package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// AIMetadata stores AI-generated metadata
type AIMetadata map[string]interface{}

// UUIDArray represents an array of UUIDs for PostgreSQL
type UUIDArray []uuid.UUID

func (ua UUIDArray) Value() (driver.Value, error) {
	if ua == nil {
		return nil, nil
	}
	strs := make([]string, len(ua))
	for i, u := range ua {
		strs[i] = u.String()
	}
	return pq.Array(strs).Value()
}

func (ua *UUIDArray) Scan(value interface{}) error {
	var strs pq.StringArray
	if err := strs.Scan(value); err != nil {
		return err
	}
	
	uuids := make([]uuid.UUID, len(strs))
	for i, s := range strs {
		u, err := uuid.Parse(s)
		if err != nil {
			return err
		}
		uuids[i] = u
	}
	*ua = uuids
	return nil
}

func (am AIMetadata) Value() (driver.Value, error) {
	if am == nil {
		return nil, nil
	}
	return json.Marshal(am)
}

func (am *AIMetadata) Scan(value interface{}) error {
	if value == nil {
		*am = make(AIMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, am)
}

// Embeddings stores vector embeddings as JSON
type Embeddings []float64

func (e Embeddings) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *Embeddings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, e)
}

// AIEnhancedNote extends the existing Note model with AI capabilities
type AIEnhancedNote struct {
	Note                    // Embed existing Note struct
	Summary        string   `json:"summary,omitempty"`
	AITags         pq.StringArray `gorm:"type:text[]" json:"ai_tags,omitempty"`
	ActionSteps    pq.StringArray `gorm:"type:text[]" json:"action_steps,omitempty"`
	LearningItems  pq.StringArray `gorm:"type:text[]" json:"learning_items,omitempty"`
	Embeddings     Embeddings `gorm:"type:jsonb" json:"embeddings,omitempty"`
	RelatedNoteIDs []uuid.UUID `gorm:"type:text[]" json:"related_note_ids,omitempty"`
	AIMetadata     AIMetadata `gorm:"type:jsonb;default:'{}'::jsonb" json:"ai_metadata,omitempty"`
	ProcessingStatus string `gorm:"default:'pending'" json:"processing_status"` // pending, processing, completed, failed
	LastProcessedAt *time.Time `json:"last_processed_at,omitempty"`
}

// AIAgent represents different types of AI agents
type AIAgent struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"user_id"`
	AgentType   string         `gorm:"not null" json:"agent_type"` // reasoning_loop, scheduler, etc.
	Status      string         `gorm:"default:'running'" json:"status"` // running, completed, failed
	InputData   AIMetadata     `gorm:"type:jsonb;default:'{}'::jsonb" json:"input_data"`
	OutputData  AIMetadata     `gorm:"type:jsonb;default:'{}'::jsonb" json:"output_data"`
	Steps       AIMetadata     `gorm:"type:jsonb;default:'[]'::jsonb" json:"steps"`
	ErrorMessage string        `json:"error_message,omitempty"`
	StartedAt   time.Time      `gorm:"not null;default:now()" json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// AIProject extends Notebook concept for AI project management
type AIProject struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"user_id"`
	Name            string         `gorm:"not null" json:"name"`
	Description     string         `json:"description"`
	Status          string         `gorm:"default:'active'" json:"status"` // active, completed, archived
	NotebookID      *uuid.UUID     `gorm:"type:uuid" json:"notebook_id,omitempty"` // Link to created notebook
	AITags          pq.StringArray `gorm:"type:text[]" json:"ai_tags,omitempty"`
	AIMetadata      AIMetadata     `gorm:"type:jsonb;default:'{}'::jsonb" json:"ai_metadata,omitempty"`
	RelatedNoteIDs  UUIDArray      `gorm:"type:text[]" json:"related_note_ids,omitempty"`
	RelatedTaskIDs  UUIDArray      `gorm:"type:text[]" json:"related_task_ids,omitempty"`
	CreatedAt       time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// AITaskEnhancement extends existing Task with AI capabilities
type AITaskEnhancement struct {
	TaskID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"task_id"`
	Task             Task       `gorm:"foreignKey:TaskID" json:"task"`
	AIGeneratedTitle string     `json:"ai_generated_title,omitempty"`
	AISuggestions    AIMetadata `gorm:"type:jsonb;default:'{}'::jsonb" json:"ai_suggestions,omitempty"`
	Priority         string     `gorm:"default:'medium'" json:"priority"` // low, medium, high
	EstimatedTime    int        `json:"estimated_time,omitempty"` // in minutes
	AIProjectID      *uuid.UUID `gorm:"type:uuid" json:"ai_project_id,omitempty"`
	CreatedAt        time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}

// ChatMemory stores conversation history for AI context
type ChatMemory struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"user_id"`
	SessionID string         `gorm:"not null" json:"session_id"`
	Role      string         `gorm:"not null" json:"role"` // user, assistant, system
	Content   string         `gorm:"type:text;not null" json:"content"`
	Metadata  AIMetadata     `gorm:"type:jsonb;default:'{}'::jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// PerplexicaSearch stores Perplexica search results for tracking and caching
type PerplexicaSearch struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"user_id"`
	Query         string         `gorm:"not null" json:"query"`
	FocusMode     string         `gorm:"not null" json:"focus_mode"`
	Answer        string         `gorm:"type:text" json:"answer"`
	Sources       AIMetadata     `gorm:"type:jsonb;default:'[]'::jsonb" json:"sources"`
	SearchContext AIMetadata     `gorm:"type:jsonb;default:'{}'::jsonb" json:"search_context"`
	Success       bool           `gorm:"default:false" json:"success"`
	ErrorMessage  string         `json:"error_message,omitempty"`
	ResponseTime  int            `json:"response_time"` // in milliseconds
	CreatedAt     time.Time      `gorm:"not null;default:now()" json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// AIToolUsage tracks when AI agents use external tools like Perplexica
type AIToolUsage struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"user_id"`
	AgentID       *uuid.UUID     `gorm:"type:uuid" json:"agent_id,omitempty"` // Links to AIAgent if used by agent
	ToolName      string         `gorm:"not null" json:"tool_name"`           // "perplexica", "anthropic", etc.
	ToolAction    string         `gorm:"not null" json:"tool_action"`         // "web_search", "academic_search", etc.
	InputData     AIMetadata     `gorm:"type:jsonb;default:'{}'::jsonb" json:"input_data"`
	OutputData    AIMetadata     `gorm:"type:jsonb;default:'{}'::jsonb" json:"output_data"`
	Success       bool           `gorm:"default:false" json:"success"`
	ErrorMessage  string         `json:"error_message,omitempty"`
	ResponseTime  int            `json:"response_time"` // in milliseconds
	CreatedAt     time.Time      `gorm:"not null;default:now()" json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Note struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"user_id"`
	NotebookID uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"notebook_id"`
	Title      string         `gorm:"not null" json:"title"`
	Blocks     []Block        `gorm:"foreignKey:NoteID" json:"blocks"`
	Tags       pq.StringArray `gorm:"type:text[]" json:"tags"`
	CreatedAt  time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (n *Note) FromJSON(data []byte) error {
	return json.Unmarshal(data, n)
}

func (n *Note) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

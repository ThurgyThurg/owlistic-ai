package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ZettelNode represents a node in the Zettelkasten graph
// It can reference notes, tasks, or projects
type ZettelNode struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	NodeType     string         `gorm:"not null" json:"node_type"` // "note", "task", "project"
	NodeID       uuid.UUID      `gorm:"type:uuid;not null" json:"node_id"`
	Title        string         `gorm:"not null" json:"title"`
	Summary      string         `gorm:"type:text" json:"summary"`
	Tags         []ZettelTag    `gorm:"many2many:zettel_node_tags;" json:"tags"`
	Connections  []ZettelEdge   `gorm:"foreignKey:SourceNodeID" json:"outgoing_connections"`
	BackLinks    []ZettelEdge   `gorm:"foreignKey:TargetNodeID" json:"incoming_connections"`
	Position     *NodePosition  `gorm:"embedded" json:"position"`
	CreatedAt    time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// ZettelEdge represents a connection between two nodes
type ZettelEdge struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SourceNodeID   uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"source_node_id"`
	TargetNodeID   uuid.UUID      `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE;" json:"target_node_id"`
	SourceNode     ZettelNode     `gorm:"foreignKey:SourceNodeID" json:"source_node"`
	TargetNode     ZettelNode     `gorm:"foreignKey:TargetNodeID" json:"target_node"`
	ConnectionType string         `gorm:"not null" json:"connection_type"` // "related", "depends_on", "references", "contradicts", "supports"
	Strength       float64        `gorm:"default:1.0" json:"strength"`     // Connection strength (0.0 to 1.0)
	Description    string         `gorm:"type:text" json:"description"`
	IsAutomatic    bool           `gorm:"default:true" json:"is_automatic"` // True if AI-generated
	CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// ZettelTag represents semantic tags for nodes
type ZettelTag struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"unique;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Color       string         `gorm:"default:'#3b82f6'" json:"color"` // Hex color for visualization
	Category    string         `json:"category"`                       // "concept", "topic", "project", "person", etc.
	Nodes       []ZettelNode   `gorm:"many2many:zettel_node_tags;" json:"nodes"`
	CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// NodePosition represents the 2D position of a node in the graph visualization
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ZettelGraph represents a saved graph layout or view
type ZettelGraph struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	FilterTags  []string       `gorm:"type:text[]" json:"filter_tags"` // Array of tag names to filter by
	Layout      string         `gorm:"default:'force'" json:"layout"`  // "force", "hierarchical", "circular"
	ViewState   GraphViewState `gorm:"embedded" json:"view_state"`
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// GraphViewState represents the current view state of the graph
type GraphViewState struct {
	CenterX float64 `json:"center_x"`
	CenterY float64 `json:"center_y"`
	Zoom    float64 `json:"zoom"`
}

// ZettelAnalysis represents AI analysis results for the knowledge graph
type ZettelAnalysis struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AnalysisType     string         `gorm:"not null" json:"analysis_type"` // "cluster", "gap", "recommendation"
	Title            string         `gorm:"not null" json:"title"`
	Description      string         `gorm:"type:text" json:"description"`
	Insights         []string       `gorm:"type:text[]" json:"insights"`
	Recommendations  []string       `gorm:"type:text[]" json:"recommendations"`
	AffectedNodeIDs  []uuid.UUID    `gorm:"type:uuid[]" json:"affected_node_ids"`
	Confidence       float64        `gorm:"default:0.0" json:"confidence"` // AI confidence in analysis
	CreatedAt        time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Input structures for API

// CreateZettelNodeInput represents input for creating a new node
type CreateZettelNodeInput struct {
	NodeType string        `json:"node_type" binding:"required,oneof=note task project"`
	NodeID   uuid.UUID     `json:"node_id" binding:"required"`
	Title    string        `json:"title" binding:"required"`
	Summary  string        `json:"summary"`
	Tags     []string      `json:"tags"`
	Position *NodePosition `json:"position"`
}

// CreateZettelEdgeInput represents input for creating a connection
type CreateZettelEdgeInput struct {
	SourceNodeID   uuid.UUID `json:"source_node_id" binding:"required"`
	TargetNodeID   uuid.UUID `json:"target_node_id" binding:"required"`
	ConnectionType string    `json:"connection_type" binding:"required,oneof=related depends_on references contradicts supports"`
	Strength       float64   `json:"strength"`
	Description    string    `json:"description"`
	IsAutomatic    bool      `json:"is_automatic"`
}

// CreateZettelTagInput represents input for creating a tag
type CreateZettelTagInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Category    string `json:"category"`
}

// UpdateNodePositionInput represents input for updating node positions
type UpdateNodePositionInput struct {
	NodeID   uuid.UUID    `json:"node_id" binding:"required"`
	Position NodePosition `json:"position" binding:"required"`
}

// ZettelSearchInput represents search parameters for the graph
type ZettelSearchInput struct {
	Query      string   `json:"query"`
	Tags       []string `json:"tags"`
	NodeTypes  []string `json:"node_types"`
	MaxDepth   int      `json:"max_depth"`
	MinStrength float64  `json:"min_strength"`
}

// GraphExportData represents the complete graph data for export/import
type GraphExportData struct {
	Nodes []ZettelNode `json:"nodes"`
	Edges []ZettelEdge `json:"edges"`
	Tags  []ZettelTag  `json:"tags"`
}

// Helper methods

// GetConnectedNodes returns all nodes connected to this node
func (zn *ZettelNode) GetConnectedNodes() []uuid.UUID {
	var connectedIDs []uuid.UUID
	
	// Add outgoing connections
	for _, edge := range zn.Connections {
		connectedIDs = append(connectedIDs, edge.TargetNodeID)
	}
	
	// Add incoming connections
	for _, edge := range zn.BackLinks {
		connectedIDs = append(connectedIDs, edge.SourceNodeID)
	}
	
	return connectedIDs
}

// GetTagNames returns the names of all tags associated with this node
func (zn *ZettelNode) GetTagNames() []string {
	var tagNames []string
	for _, tag := range zn.Tags {
		tagNames = append(tagNames, tag.Name)
	}
	return tagNames
}

// IsValidConnectionType checks if the connection type is valid
func IsValidConnectionType(connectionType string) bool {
	validTypes := map[string]bool{
		"related":     true,
		"depends_on":  true,
		"references":  true,
		"contradicts": true,
		"supports":    true,
	}
	return validTypes[connectionType]
}

// IsValidNodeType checks if the node type is valid
func IsValidNodeType(nodeType string) bool {
	validTypes := map[string]bool{
		"note":    true,
		"task":    true,
		"project": true,
	}
	return validTypes[nodeType]
}
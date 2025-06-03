package services

import (
	"context"
	"fmt"
	"log"

	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ZettelkastenService handles operations for the Zettelkasten knowledge graph
type ZettelkastenService struct {
	db       *gorm.DB
	aiService *ZettelkastenAIService
}

// NewZettelkastenService creates a new ZettelkastenService
func NewZettelkastenService(db *gorm.DB, aiService *AIService) *ZettelkastenService {
	zettelAI := NewZettelkastenAIService(db, aiService)
	return &ZettelkastenService{
		db:        db,
		aiService: zettelAI,
	}
}

// CreateNodeFromContent creates a Zettelkasten node from existing content (note, task, or project)
func (zs *ZettelkastenService) CreateNodeFromContent(ctx context.Context, nodeType string, nodeID uuid.UUID) (*models.ZettelNode, error) {
	// Check if node already exists
	var existingNode models.ZettelNode
	if err := zs.db.Where("node_type = ? AND node_id = ?", nodeType, nodeID).First(&existingNode).Error; err == nil {
		return &existingNode, nil // Node already exists
	}

	// Get content based on type
	title, content, err := zs.getContentByType(nodeType, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	// Use AI to analyze content and suggest tags
	aiResult, err := zs.aiService.AnalyzeContentForTags(ctx, content, title, nodeType)
	if err != nil {
		log.Printf("AI tagging failed, using basic extraction: %v", err)
		aiResult = &AITaggingResult{
			Tags: []TagSuggestion{{
				Name:       nodeType,
				Category:   "type",
				Confidence: 1.0,
				Reason:     "Content type",
				Color:      "#6b7280",
			}},
			Summary:    title,
			Confidence: 0.5,
		}
	}

	// Create or get tags
	var tags []models.ZettelTag
	for _, tagSugg := range aiResult.Tags {
		tag, err := zs.createOrGetTag(tagSugg)
		if err != nil {
			log.Printf("Failed to create tag %s: %v", tagSugg.Name, err)
			continue
		}
		tags = append(tags, *tag)
	}

	// Create the node
	node := &models.ZettelNode{
		NodeType: nodeType,
		NodeID:   nodeID,
		Title:    title,
		Summary:  aiResult.Summary,
		Tags:     tags,
	}

	if err := zs.db.Create(node).Error; err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	// Load the complete node with associations
	if err := zs.db.Preload("Tags").Preload("Connections").Preload("BackLinks").First(node, node.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load created node: %w", err)
	}

	// Analyze and create automatic connections
	go zs.analyzeAndCreateConnections(context.Background(), node)

	return node, nil
}

// GetOrCreateNode gets an existing node or creates one if it doesn't exist
func (zs *ZettelkastenService) GetOrCreateNode(ctx context.Context, nodeType string, nodeID uuid.UUID) (*models.ZettelNode, error) {
	var node models.ZettelNode
	
	// Try to find existing node
	err := zs.db.Preload("Tags").Preload("Connections.TargetNode").Preload("BackLinks.SourceNode").
		Where("node_type = ? AND node_id = ?", nodeType, nodeID).First(&node).Error
	
	if err == nil {
		return &node, nil // Found existing node
	}
	
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Create new node
	return zs.CreateNodeFromContent(ctx, nodeType, nodeID)
}

// GetAllNodes returns all nodes with optional filtering
func (zs *ZettelkastenService) GetAllNodes(ctx context.Context, filter *models.ZettelSearchInput) ([]models.ZettelNode, error) {
	query := zs.db.Preload("Tags").Preload("Connections.TargetNode").Preload("BackLinks.SourceNode")

	if filter != nil {
		if len(filter.NodeTypes) > 0 {
			query = query.Where("node_type IN ?", filter.NodeTypes)
		}
		
		if len(filter.Tags) > 0 {
			query = query.Joins("JOIN zettel_node_tags ON zettel_nodes.id = zettel_node_tags.zettel_node_id").
				Joins("JOIN zettel_tags ON zettel_node_tags.zettel_tag_id = zettel_tags.id").
				Where("zettel_tags.name IN ?", filter.Tags)
		}

		if filter.Query != "" {
			query = query.Where("title ILIKE ? OR summary ILIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
		}
	}

	var nodes []models.ZettelNode
	if err := query.Find(&nodes).Error; err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	return nodes, nil
}

// GetNodeByID returns a specific node with all its connections
func (zs *ZettelkastenService) GetNodeByID(ctx context.Context, nodeID uuid.UUID) (*models.ZettelNode, error) {
	var node models.ZettelNode
	
	err := zs.db.Preload("Tags").
		Preload("Connections.TargetNode.Tags").
		Preload("BackLinks.SourceNode.Tags").
		First(&node, nodeID).Error
		
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return &node, nil
}

// CreateConnection creates a connection between two nodes
func (zs *ZettelkastenService) CreateConnection(ctx context.Context, input *models.CreateZettelEdgeInput) (*models.ZettelEdge, error) {
	// Validate that both nodes exist
	var sourceExists, targetExists bool
	
	if err := zs.db.Model(&models.ZettelNode{}).Select("count(*) > 0").Where("id = ?", input.SourceNodeID).Find(&sourceExists).Error; err != nil {
		return nil, fmt.Errorf("failed to check source node: %w", err)
	}
	
	if err := zs.db.Model(&models.ZettelNode{}).Select("count(*) > 0").Where("id = ?", input.TargetNodeID).Find(&targetExists).Error; err != nil {
		return nil, fmt.Errorf("failed to check target node: %w", err)
	}

	if !sourceExists || !targetExists {
		return nil, fmt.Errorf("one or both nodes do not exist")
	}

	// Check for existing connection
	var existingEdge models.ZettelEdge
	err := zs.db.Where("source_node_id = ? AND target_node_id = ? AND connection_type = ?", 
		input.SourceNodeID, input.TargetNodeID, input.ConnectionType).First(&existingEdge).Error
	
	if err == nil {
		return &existingEdge, nil // Connection already exists
	}

	// Create new connection
	edge := &models.ZettelEdge{
		SourceNodeID:   input.SourceNodeID,
		TargetNodeID:   input.TargetNodeID,
		ConnectionType: input.ConnectionType,
		Strength:       input.Strength,
		Description:    input.Description,
		IsAutomatic:    input.IsAutomatic,
	}

	if edge.Strength == 0 {
		edge.Strength = 1.0 // Default strength
	}

	if err := zs.db.Create(edge).Error; err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	// Load the complete edge with nodes
	if err := zs.db.Preload("SourceNode").Preload("TargetNode").First(edge, edge.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load created edge: %w", err)
	}

	return edge, nil
}

// UpdateNodePosition updates the position of a node in the graph
func (zs *ZettelkastenService) UpdateNodePosition(ctx context.Context, nodeID uuid.UUID, position models.NodePosition) error {
	result := zs.db.Model(&models.ZettelNode{}).Where("id = ?", nodeID).
		Updates(map[string]interface{}{
			"position": position,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update node position: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("node not found")
	}

	return nil
}

// DeleteConnection removes a connection between nodes
func (zs *ZettelkastenService) DeleteConnection(ctx context.Context, edgeID uuid.UUID) error {
	result := zs.db.Delete(&models.ZettelEdge{}, edgeID)
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete connection: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("connection not found")
	}

	return nil
}

// GetAllTags returns all available tags
func (zs *ZettelkastenService) GetAllTags(ctx context.Context) ([]models.ZettelTag, error) {
	var tags []models.ZettelTag
	
	if err := zs.db.Preload("Nodes").Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return tags, nil
}

// CreateTag creates a new tag
func (zs *ZettelkastenService) CreateTag(ctx context.Context, input *models.CreateZettelTagInput) (*models.ZettelTag, error) {
	tag := &models.ZettelTag{
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
		Category:    input.Category,
	}

	if tag.Color == "" {
		tag.Color = "#3b82f6" // Default blue
	}

	if err := zs.db.Create(tag).Error; err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

// GetGraphExportData exports the complete graph data
func (zs *ZettelkastenService) GetGraphExportData(ctx context.Context) (*models.GraphExportData, error) {
	var nodes []models.ZettelNode
	var edges []models.ZettelEdge
	var tags []models.ZettelTag

	if err := zs.db.Preload("Tags").Find(&nodes).Error; err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	if err := zs.db.Preload("SourceNode").Preload("TargetNode").Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}

	if err := zs.db.Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return &models.GraphExportData{
		Nodes: nodes,
		Edges: edges,
		Tags:  tags,
	}, nil
}

// AnalyzeKnowledgeGraph performs AI analysis of the knowledge graph
func (zs *ZettelkastenService) AnalyzeKnowledgeGraph(ctx context.Context) (*models.ZettelAnalysis, error) {
	// Get all nodes and tags for analysis
	nodes, err := zs.GetAllNodes(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes for analysis: %w", err)
	}

	tags, err := zs.GetAllTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags for analysis: %w", err)
	}

	// Perform AI analysis
	analysis, err := zs.aiService.AnalyzeKnowledgeGaps(ctx, nodes, tags)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Save analysis to database
	if err := zs.db.Create(analysis).Error; err != nil {
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	return analysis, nil
}

// Helper methods

// getContentByType retrieves content based on the node type
func (zs *ZettelkastenService) getContentByType(nodeType string, nodeID uuid.UUID) (string, string, error) {
	switch nodeType {
	case "note":
		var note models.Note
		if err := zs.db.Preload("Blocks").First(&note, nodeID).Error; err != nil {
			return "", "", fmt.Errorf("note not found: %w", err)
		}
		
		content := ""
		for _, block := range note.Blocks {
			if text, ok := block.Content["text"]; ok {
				if textStr, ok := text.(string); ok {
					content += textStr + "\n"
				}
			}
		}
		
		return note.Title, content, nil

	case "task":
		var task models.Task
		if err := zs.db.First(&task, nodeID).Error; err != nil {
			return "", "", fmt.Errorf("task not found: %w", err)
		}
		return task.Title, task.Description, nil

	case "project":
		// Assuming you have a project model - adjust as needed
		return "Project", "Project content", nil

	default:
		return "", "", fmt.Errorf("unsupported node type: %s", nodeType)
	}
}

// createOrGetTag creates a new tag or returns existing one
func (zs *ZettelkastenService) createOrGetTag(suggestion TagSuggestion) (*models.ZettelTag, error) {
	var tag models.ZettelTag
	
	// Try to find existing tag
	err := zs.db.Where("name = ?", suggestion.Name).First(&tag).Error
	if err == nil {
		return &tag, nil // Found existing tag
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Create new tag
	tag = models.ZettelTag{
		Name:        suggestion.Name,
		Description: suggestion.Reason,
		Color:       suggestion.Color,
		Category:    suggestion.Category,
	}

	if tag.Color == "" {
		tag.Color = "#3b82f6"
	}

	if err := zs.db.Create(&tag).Error; err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return &tag, nil
}

// analyzeAndCreateConnections runs in background to find and create connections
func (zs *ZettelkastenService) analyzeAndCreateConnections(ctx context.Context, sourceNode *models.ZettelNode) {
	// Get candidate nodes for connection (limit to recent ones to avoid overwhelming AI)
	var candidates []models.ZettelNode
	if err := zs.db.Preload("Tags").Where("id != ?", sourceNode.ID).
		Order("created_at DESC").Limit(20).Find(&candidates).Error; err != nil {
		log.Printf("Failed to get candidate nodes for connection analysis: %v", err)
		return
	}

	if len(candidates) == 0 {
		return
	}

	// Use AI to analyze potential connections
	result, err := zs.aiService.AnalyzeConnections(ctx, sourceNode, candidates)
	if err != nil {
		log.Printf("Failed to analyze connections: %v", err)
		return
	}

	// Create suggested connections
	for _, suggestion := range result.Connections {
		if suggestion.Confidence < 0.6 { // Skip low confidence suggestions
			continue
		}

		input := &models.CreateZettelEdgeInput{
			SourceNodeID:   sourceNode.ID,
			TargetNodeID:   suggestion.TargetNodeID,
			ConnectionType: suggestion.ConnectionType,
			Strength:       suggestion.Strength,
			Description:    suggestion.Reason,
			IsAutomatic:    true,
		}

		_, err := zs.CreateConnection(ctx, input)
		if err != nil {
			log.Printf("Failed to create automatic connection: %v", err)
		}
	}
}
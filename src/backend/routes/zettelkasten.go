package routes

import (
	"log"
	"net/http"
	"strconv"

	"owlistic-notes/owlistic/models"
	"owlistic-notes/owlistic/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ZettelkastenRoutes handles HTTP routes for the Zettelkasten knowledge graph
type ZettelkastenRoutes struct {
	db      *gorm.DB
	service *services.ZettelkastenService
}

// NewZettelkastenRoutes creates a new ZettelkastenRoutes instance
func NewZettelkastenRoutes(db *gorm.DB) (*ZettelkastenRoutes, error) {
	aiService := services.NewAIService(db)
	zettelService := services.NewZettelkastenService(db, aiService)
	
	return &ZettelkastenRoutes{
		db:      db,
		service: zettelService,
	}, nil
}

// RegisterRoutes registers all Zettelkasten routes
func (zr *ZettelkastenRoutes) RegisterRoutes(router *gin.RouterGroup) {
	zettel := router.Group("/zettelkasten")
	{
		// Node operations
		zettel.POST("/nodes", zr.createNode)
		zettel.GET("/nodes", zr.getAllNodes)
		zettel.GET("/nodes/:id", zr.getNodeByID)
		zettel.PUT("/nodes/:id/position", zr.updateNodePosition)
		zettel.DELETE("/nodes/:id", zr.deleteNode)

		// Connection operations
		zettel.POST("/connections", zr.createConnection)
		zettel.DELETE("/connections/:id", zr.deleteConnection)

		// Tag operations
		zettel.GET("/tags", zr.getAllTags)
		zettel.POST("/tags", zr.createTag)

		// Graph operations
		zettel.GET("/graph", zr.getGraphData)
		zettel.GET("/graph/export", zr.exportGraph)
		zettel.POST("/graph/analyze", zr.analyzeGraph)

		// Search and discovery
		zettel.POST("/search", zr.searchNodes)
		zettel.GET("/discover/:nodeId", zr.discoverConnections)

		// Synchronization
		zettel.POST("/sync/notes", zr.syncNotes)
		zettel.POST("/sync/tasks", zr.syncTasks)
		zettel.POST("/sync/all", zr.syncAll)
	}
}

// createNode creates a new Zettelkasten node
func (zr *ZettelkastenRoutes) createNode(c *gin.Context) {
	var input models.CreateZettelNodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if !models.IsValidNodeType(input.NodeType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node type"})
		return
	}

	node, err := zr.service.CreateNodeFromContent(c.Request.Context(), input.NodeType, input.NodeID)
	if err != nil {
		log.Printf("Failed to create Zettelkasten node: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create node"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"node": node})
}

// getAllNodes returns all nodes with optional filtering
func (zr *ZettelkastenRoutes) getAllNodes(c *gin.Context) {
	var filter models.ZettelSearchInput

	// Parse query parameters
	if query := c.Query("query"); query != "" {
		filter.Query = query
	}

	if nodeTypes := c.QueryArray("node_types"); len(nodeTypes) > 0 {
		filter.NodeTypes = nodeTypes
	}

	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filter.Tags = tags
	}

	if maxDepthStr := c.Query("max_depth"); maxDepthStr != "" {
		if maxDepth, err := strconv.Atoi(maxDepthStr); err == nil {
			filter.MaxDepth = maxDepth
		}
	}

	if minStrengthStr := c.Query("min_strength"); minStrengthStr != "" {
		if minStrength, err := strconv.ParseFloat(minStrengthStr, 64); err == nil {
			filter.MinStrength = minStrength
		}
	}

	var filterPtr *models.ZettelSearchInput
	if filter.Query != "" || len(filter.NodeTypes) > 0 || len(filter.Tags) > 0 {
		filterPtr = &filter
	}

	nodes, err := zr.service.GetAllNodes(c.Request.Context(), filterPtr)
	if err != nil {
		log.Printf("Failed to get nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get nodes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

// getNodeByID returns a specific node with all its connections
func (zr *ZettelkastenRoutes) getNodeByID(c *gin.Context) {
	idStr := c.Param("id")
	nodeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	node, err := zr.service.GetNodeByID(c.Request.Context(), nodeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		log.Printf("Failed to get node: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get node"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"node": node})
}

// updateNodePosition updates the position of a node in the graph
func (zr *ZettelkastenRoutes) updateNodePosition(c *gin.Context) {
	idStr := c.Param("id")
	nodeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	var input models.UpdateNodePositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := zr.service.UpdateNodePosition(c.Request.Context(), nodeID, input.Position); err != nil {
		if err.Error() == "node not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		log.Printf("Failed to update node position: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position updated successfully"})
}

// deleteNode removes a node from the graph
func (zr *ZettelkastenRoutes) deleteNode(c *gin.Context) {
	idStr := c.Param("id")
	nodeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	// Delete node and its connections (handled by cascade)
	result := zr.db.Delete(&models.ZettelNode{}, nodeID)
	if result.Error != nil {
		log.Printf("Failed to delete node: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete node"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Node deleted successfully"})
}

// createConnection creates a connection between two nodes
func (zr *ZettelkastenRoutes) createConnection(c *gin.Context) {
	var input models.CreateZettelEdgeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if !models.IsValidConnectionType(input.ConnectionType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection type"})
		return
	}

	edge, err := zr.service.CreateConnection(c.Request.Context(), &input)
	if err != nil {
		log.Printf("Failed to create connection: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create connection"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"connection": edge})
}

// deleteConnection removes a connection between nodes
func (zr *ZettelkastenRoutes) deleteConnection(c *gin.Context) {
	idStr := c.Param("id")
	edgeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	if err := zr.service.DeleteConnection(c.Request.Context(), edgeID); err != nil {
		if err.Error() == "connection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
			return
		}
		log.Printf("Failed to delete connection: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection deleted successfully"})
}

// getAllTags returns all available tags
func (zr *ZettelkastenRoutes) getAllTags(c *gin.Context) {
	tags, err := zr.service.GetAllTags(c.Request.Context())
	if err != nil {
		log.Printf("Failed to get tags: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// createTag creates a new tag
func (zr *ZettelkastenRoutes) createTag(c *gin.Context) {
	var input models.CreateZettelTagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	tag, err := zr.service.CreateTag(c.Request.Context(), &input)
	if err != nil {
		log.Printf("Failed to create tag: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"tag": tag})
}

// getGraphData returns the complete graph data for visualization
func (zr *ZettelkastenRoutes) getGraphData(c *gin.Context) {
	data, err := zr.service.GetGraphExportData(c.Request.Context())
	if err != nil {
		log.Printf("Failed to get graph data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get graph data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": data.Nodes,
		"edges": data.Edges,
		"tags":  data.Tags,
	})
}

// exportGraph exports the complete graph data
func (zr *ZettelkastenRoutes) exportGraph(c *gin.Context) {
	data, err := zr.service.GetGraphExportData(c.Request.Context())
	if err != nil {
		log.Printf("Failed to export graph: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export graph"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=zettelkasten_graph.json")
	c.JSON(http.StatusOK, data)
}

// analyzeGraph performs AI analysis of the knowledge graph
func (zr *ZettelkastenRoutes) analyzeGraph(c *gin.Context) {
	analysis, err := zr.service.AnalyzeKnowledgeGraph(c.Request.Context())
	if err != nil {
		log.Printf("Failed to analyze graph: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze graph"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": analysis})
}

// searchNodes searches for nodes based on criteria
func (zr *ZettelkastenRoutes) searchNodes(c *gin.Context) {
	var input models.ZettelSearchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	nodes, err := zr.service.GetAllNodes(c.Request.Context(), &input)
	if err != nil {
		log.Printf("Failed to search nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search nodes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

// discoverConnections suggests potential connections for a node
func (zr *ZettelkastenRoutes) discoverConnections(c *gin.Context) {
	idStr := c.Param("nodeId")
	nodeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	// Get the source node
	sourceNode, err := zr.service.GetNodeByID(c.Request.Context(), nodeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		log.Printf("Failed to get source node: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get node"})
		return
	}

	// Get candidate nodes (excluding already connected ones)
	candidates, err := zr.service.GetAllNodes(c.Request.Context(), nil)
	if err != nil {
		log.Printf("Failed to get candidate nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get candidates"})
		return
	}

	// Filter out already connected nodes and the source node itself
	connectedIDs := sourceNode.GetConnectedNodes()
	connectedMap := make(map[uuid.UUID]bool)
	connectedMap[sourceNode.ID] = true
	for _, id := range connectedIDs {
		connectedMap[id] = true
	}

	var filteredCandidates []models.ZettelNode
	for _, candidate := range candidates {
		if !connectedMap[candidate.ID] {
			filteredCandidates = append(filteredCandidates, candidate)
		}
	}

	// Limit candidates to avoid overwhelming the AI
	if len(filteredCandidates) > 20 {
		filteredCandidates = filteredCandidates[:20]
	}

	c.JSON(http.StatusOK, gin.H{
		"source_node": sourceNode,
		"suggestions": filteredCandidates,
		"message":     "Use POST /zettelkasten/connections to create connections",
	})
}

// syncNotes creates nodes for all existing notes
func (zr *ZettelkastenRoutes) syncNotes(c *gin.Context) {
	var notes []models.Note
	if err := zr.db.Find(&notes).Error; err != nil {
		log.Printf("Failed to get notes for sync: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notes"})
		return
	}

	created := 0
	for _, note := range notes {
		_, err := zr.service.GetOrCreateNode(c.Request.Context(), "note", note.ID)
		if err != nil {
			log.Printf("Failed to sync note %s: %v", note.ID, err)
			continue
		}
		created++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Notes synchronized successfully",
		"total_notes": len(notes),
		"nodes_created": created,
	})
}

// syncTasks creates nodes for all existing tasks
func (zr *ZettelkastenRoutes) syncTasks(c *gin.Context) {
	var tasks []models.Task
	if err := zr.db.Find(&tasks).Error; err != nil {
		log.Printf("Failed to get tasks for sync: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	created := 0
	for _, task := range tasks {
		_, err := zr.service.GetOrCreateNode(c.Request.Context(), "task", task.ID)
		if err != nil {
			log.Printf("Failed to sync task %s: %v", task.ID, err)
			continue
		}
		created++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Tasks synchronized successfully",
		"total_tasks":  len(tasks),
		"nodes_created": created,
	})
}

// syncAll creates nodes for all existing content
func (zr *ZettelkastenRoutes) syncAll(c *gin.Context) {
	// Sync notes
	var notes []models.Note
	if err := zr.db.Find(&notes).Error; err != nil {
		log.Printf("Failed to get notes for sync: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notes"})
		return
	}

	notesCreated := 0
	for _, note := range notes {
		_, err := zr.service.GetOrCreateNode(c.Request.Context(), "note", note.ID)
		if err != nil {
			log.Printf("Failed to sync note %s: %v", note.ID, err)
			continue
		}
		notesCreated++
	}

	// Sync tasks
	var tasks []models.Task
	if err := zr.db.Find(&tasks).Error; err != nil {
		log.Printf("Failed to get tasks for sync: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	tasksCreated := 0
	for _, task := range tasks {
		_, err := zr.service.GetOrCreateNode(c.Request.Context(), "task", task.ID)
		if err != nil {
			log.Printf("Failed to sync task %s: %v", task.ID, err)
			continue
		}
		tasksCreated++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "All content synchronized successfully",
		"total_notes":    len(notes),
		"notes_created":  notesCreated,
		"total_tasks":    len(tasks),
		"tasks_created":  tasksCreated,
		"total_created":  notesCreated + tasksCreated,
	})
}
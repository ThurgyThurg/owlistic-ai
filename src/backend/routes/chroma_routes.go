package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"owlistic-notes/owlistic/models"
	"owlistic-notes/owlistic/services"
)

type ChromaRoutes struct {
	db        *gorm.DB
	aiService *services.AIService
}

func NewChromaRoutes(db *gorm.DB) *ChromaRoutes {
	return &ChromaRoutes{
		db:        db,
		aiService: services.NewAIService(db),
	}
}

func (cr *ChromaRoutes) RegisterRoutes(router *gin.RouterGroup) {
	chromaGroup := router.Group("/chroma")
	{
		// Collection management
		chromaGroup.GET("/stats", cr.getCollectionStats)
		chromaGroup.POST("/refresh", cr.refreshCollection)
		
		// Search endpoints
		chromaGroup.POST("/search", cr.semanticSearch)
		chromaGroup.GET("/notes/:id/related", cr.getRelatedNotes)
		
		// Note embedding management
		chromaGroup.POST("/notes/:id/embed", cr.embedNote)
		chromaGroup.DELETE("/notes/:id/embed", cr.removeNoteEmbedding)
	}
}

// getCollectionStats returns statistics about the ChromaDB collection
func (cr *ChromaRoutes) getCollectionStats(c *gin.Context) {
	stats, err := cr.aiService.GetChromaCollectionStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get collection stats",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
		"message": "Collection statistics retrieved successfully",
	})
}

// refreshCollection rebuilds the entire ChromaDB collection
func (cr *ChromaRoutes) refreshCollection(c *gin.Context) {
	// This could be a long operation, so we'll do it asynchronously
	go func() {
		if err := cr.aiService.RefreshChromaCollection(c.Request.Context()); err != nil {
			// Log error - in production, you might want to store this status somewhere
			// or use a job queue system
			println("Error refreshing ChromaDB collection:", err.Error())
		}
	}()
	
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Collection refresh started",
		"status": "processing",
	})
}

// semanticSearch performs semantic search across notes
func (cr *ChromaRoutes) semanticSearch(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	var req struct {
		Query string `json:"query" binding:"required"`
		Limit int    `json:"limit"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Default limit
	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 50 {
		req.Limit = 50 // Cap at 50
	}
	
	// Perform semantic search
	results, err := cr.aiService.SearchNotesByEmbedding(
		c.Request.Context(),
		req.Query,
		userID.(uuid.UUID),
		req.Limit,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to perform semantic search",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count": len(results),
		"query": req.Query,
	})
}

// getRelatedNotes finds notes related to a specific note
func (cr *ChromaRoutes) getRelatedNotes(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}
	
	// Get limit from query params
	limit := 5
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 20 {
			limit = parsed
		}
	}
	
	// Find related notes
	relatedNotes, err := cr.aiService.FindRelatedNotes(c.Request.Context(), noteID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find related notes",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"related_notes": relatedNotes,
		"count": len(relatedNotes),
		"note_id": noteID,
	})
}

// embedNote manually triggers embedding for a specific note
func (cr *ChromaRoutes) embedNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}
	
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	// Verify note belongs to user
	var note models.Note
	if err := cr.db.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	
	// Get enhanced note data if exists
	var enhanced models.AIEnhancedNote
	cr.db.Where("id = ?", noteID).First(&enhanced)
	
	// Add to ChromaDB
	if err := cr.aiService.AddNoteToChroma(c.Request.Context(), &note, &enhanced); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to embed note",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Note embedded successfully",
		"note_id": noteID,
	})
}

// removeNoteEmbedding removes a note from the vector database
func (cr *ChromaRoutes) removeNoteEmbedding(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}
	
	// Remove from ChromaDB
	if err := cr.aiService.RemoveNoteFromChroma(c.Request.Context(), noteID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove note embedding",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Note embedding removed successfully",
		"note_id": noteID,
	})
}
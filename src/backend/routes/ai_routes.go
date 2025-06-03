package routes

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"owlistic-notes/owlistic/models"
	"owlistic-notes/owlistic/services"
)

type AIRoutes struct {
	db                   *gorm.DB
	aiService            *services.AIService
	chatService          *services.ChatService
	reasoningAgentService *services.ReasoningAgentService
}

func NewAIRoutes(db *gorm.DB) *AIRoutes {
	// Initialize services
	aiService := services.NewAIService(db)
	noteService := services.NoteServiceInstance
	
	return &AIRoutes{
		db:                   db,
		aiService:            aiService,
		chatService:          services.NewChatService(db, aiService, noteService),
		reasoningAgentService: services.NewReasoningAgentService(db, aiService, noteService),
	}
}

func (ar *AIRoutes) RegisterRoutes(routerGroup *gin.RouterGroup) {
	aiGroup := routerGroup.Group("/ai")
	{
		// Note AI enhancements
		aiGroup.POST("/notes/:id/process", ar.processNoteWithAI)
		aiGroup.GET("/notes/:id/enhanced", ar.getEnhancedNote)
		aiGroup.POST("/notes/search/semantic", ar.semanticSearch)
		
		// AI Projects
		aiGroup.POST("/projects", ar.createAIProject)
		aiGroup.GET("/projects", ar.getAIProjects)
		aiGroup.GET("/projects/:id", ar.getAIProject)
		aiGroup.PUT("/projects/:id", ar.updateAIProject)
		aiGroup.DELETE("/projects/:id", ar.deleteAIProject)
		
		// AI Agents
		aiGroup.POST("/agents/run", ar.runAgent)
		aiGroup.GET("/agents/runs", ar.getAgentRuns)
		aiGroup.GET("/agents/runs/:id", ar.getAgentRun)
		aiGroup.POST("/agents/quick-goal", ar.quickGoalAgent)
		aiGroup.POST("/agents/task-breakdown", ar.breakDownTask)
		
		// AI Chat/Memory
		aiGroup.POST("/chat", ar.chatWithAI)
		aiGroup.GET("/chat/history", ar.getChatHistory)
		aiGroup.GET("/chat/sessions", ar.getChatSessions)
		aiGroup.DELETE("/chat/sessions/:id", ar.deleteChatSession)
		
		// Reasoning Agent
		aiGroup.POST("/agents/reasoning", ar.runReasoningAgent)
		aiGroup.GET("/agents/reasoning/:id", ar.getReasoningAgentResult)
	}
}

// processNoteWithAI triggers AI processing for a specific note
func (ar *AIRoutes) processNoteWithAI(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify note belongs to user
	var note models.Note
	if err := ar.db.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Process note with AI in background
	go func() {
		// Use background context to prevent cancellation when HTTP request completes
		ctx := context.Background()
		if err := ar.aiService.ProcessNoteWithAI(ctx, noteID); err != nil {
			// Log error - in production, you might want to store this in a job queue
			// and have proper error handling/retry logic
			println("AI processing error:", err.Error())
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "AI processing started",
		"note_id": noteID,
	})
}

// getEnhancedNote returns the AI-enhanced version of a note
func (ar *AIRoutes) getEnhancedNote(c *gin.Context) {
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

	var aiNote models.AIEnhancedNote
	if err := ar.db.Where("id = ? AND user_id = ?", noteID, userID).First(&aiNote).Error; err != nil {
		// Return regular note if AI enhancement doesn't exist
		var note models.Note
		if err := ar.db.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
			return
		}
		
		
		c.JSON(http.StatusOK, gin.H{
			"note": note,
			"ai_enhancement": gin.H{
				"processing_status": "not_processed",
				"message": "AI enhancement not available",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"note": aiNote.Note,
		"ai_enhancement": gin.H{
			"summary":            aiNote.Summary,
			"ai_tags":           aiNote.AITags,
			"action_steps":      aiNote.ActionSteps,
			"learning_items":    aiNote.LearningItems,
			"related_note_ids":  aiNote.RelatedNoteIDs,
			"processing_status": aiNote.ProcessingStatus,
			"last_processed_at": aiNote.LastProcessedAt,
		},
	})
}

// semanticSearch performs AI-powered semantic search
func (ar *AIRoutes) semanticSearch(c *gin.Context) {
	var request struct {
		Query string `json:"query" binding:"required"`
		Limit int    `json:"limit"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Limit == 0 {
		request.Limit = 10
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// For now, fall back to text search since semantic search needs vector DB
	var notes []models.Note
	searchTerm := "%" + request.Query + "%"
	
	if err := ar.db.Where("user_id = ? AND (title ILIKE ? OR EXISTS (SELECT 1 FROM blocks WHERE blocks.note_id = notes.id AND blocks.content ILIKE ?))", 
		userID, searchTerm, searchTerm).
		Limit(request.Limit).
		Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query": request.Query,
		"results": notes,
		"total_count": len(notes),
		"search_type": "text", // Will be "semantic" once vector DB is integrated
	})
}

// createAIProject creates a new AI-managed project
func (ar *AIRoutes) createAIProject(c *gin.Context) {
	var request struct {
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		AITags      []string               `json:"ai_tags"`
		AIMetadata  map[string]interface{} `json:"ai_metadata"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if this project was created from a task breakdown
	var notebookID *uuid.UUID
	var relatedNoteIDs []uuid.UUID

	if breakdown, hasBreakdown := request.AIMetadata["breakdown"]; hasBreakdown {
		if breakdownMap, ok := breakdown.(map[string]interface{}); ok {
			// Create notebook and notes for the project
			nbID, noteIDs, err := ar.aiService.CreateProjectNotebook(
				c.Request.Context(),
				userID.(uuid.UUID),
				request.Name,
				request.Description,
				breakdownMap,
			)
			if err != nil {
				log.Printf("Failed to create project notebook: %v", err)
				// Continue with project creation even if notebook creation fails
			} else {
				notebookID = nbID
				relatedNoteIDs = noteIDs
			}
		}
	}

	project := models.AIProject{
		UserID:         userID.(uuid.UUID),
		Name:           request.Name,
		Description:    request.Description,
		Status:         "active",
		NotebookID:     notebookID,
		AITags:         pq.StringArray(request.AITags),
		AIMetadata:     request.AIMetadata,
		RelatedNoteIDs: models.UUIDArray(relatedNoteIDs),
	}

	if err := ar.db.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// getAIProjects returns all AI projects for the user
func (ar *AIRoutes) getAIProjects(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var projects []models.AIProject
	if err := ar.db.Where("user_id = ?", userID).Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// getAIProject returns a specific AI project
func (ar *AIRoutes) getAIProject(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var project models.AIProject
	if err := ar.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// updateAIProject updates an AI project
func (ar *AIRoutes) updateAIProject(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request struct {
		Name        *string                 `json:"name"`
		Description *string                 `json:"description"`
		Status      *string                 `json:"status"`
		AITags      []string                `json:"ai_tags"`
		AIMetadata  map[string]interface{}  `json:"ai_metadata"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project models.AIProject
	if err := ar.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Update fields
	if request.Name != nil {
		project.Name = *request.Name
	}
	if request.Description != nil {
		project.Description = *request.Description
	}
	if request.Status != nil {
		project.Status = *request.Status
	}
	if request.AITags != nil {
		project.AITags = request.AITags
	}
	if request.AIMetadata != nil {
		project.AIMetadata = request.AIMetadata
	}

	if err := ar.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// deleteAIProject deletes an AI project
func (ar *AIRoutes) deleteAIProject(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	result := ar.db.Where("id = ? AND user_id = ?", projectID, userID).Delete(&models.AIProject{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// runAgent starts an AI agent
func (ar *AIRoutes) runAgent(c *gin.Context) {
	var request struct {
		AgentType string                 `json:"agent_type" binding:"required"`
		InputData map[string]interface{} `json:"input_data"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	agent := models.AIAgent{
		UserID:    userID.(uuid.UUID),
		AgentType: request.AgentType,
		Status:    "running",
		InputData: request.InputData,
	}

	if err := ar.db.Create(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent run"})
		return
	}

	// TODO: Implement actual agent execution in background
	
	c.JSON(http.StatusCreated, agent)
}

// getAgentRuns returns agent execution history
func (ar *AIRoutes) getAgentRuns(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	var agents []models.AIAgent
	if err := ar.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&agents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent runs"})
		return
	}

	c.JSON(http.StatusOK, agents)
}

// getAgentRun returns a specific agent run
func (ar *AIRoutes) getAgentRun(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var agent models.AIAgent
	if err := ar.db.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent run not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// quickGoalAgent runs a quick goal-planning agent
func (ar *AIRoutes) quickGoalAgent(c *gin.Context) {
	var request struct {
		Goal    string `json:"goal" binding:"required"`
		Context string `json:"context"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create agent run with goal planning type
	agentRequest := struct {
		AgentType string                 `json:"agent_type"`
		InputData map[string]interface{} `json:"input_data"`
	}{
		AgentType: "goal_planner",
		InputData: map[string]interface{}{
			"goal":    request.Goal,
			"context": request.Context,
		},
	}

	// Reuse the runAgent logic
	c.Set("agent_request", agentRequest)
	ar.runAgent(c)
}

// chatWithAI provides conversational AI interface with RAG
func (ar *AIRoutes) chatWithAI(c *gin.Context) {
	var request services.ChatRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Use the chat service to handle the request
	response, err := ar.chatService.Chat(c.Request.Context(), userID.(uuid.UUID), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process chat: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getChatHistory returns chat history for a session
func (ar *AIRoutes) getChatHistory(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var messages []models.ChatMemory
	if err := ar.db.Where("user_id = ? AND session_id = ?", userID, sessionID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"messages":   messages,
	})
}

// breakDownTask uses AI to break down a goal into manageable steps
func (ar *AIRoutes) breakDownTask(c *gin.Context) {
	var request struct {
		Goal        string                 `json:"goal" binding:"required"`
		Context     string                 `json:"context"`
		TimeFrame   string                 `json:"time_frame"`
		MaxSteps    int                    `json:"max_steps"`
		Priority    string                 `json:"priority"`
		Preferences map[string]interface{} `json:"preferences"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Call AI service to break down the task
	breakdown, err := ar.aiService.BreakDownTask(c.Request.Context(), request.Goal, request.Context, request.MaxSteps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to break down task: " + err.Error()})
		return
	}

	// Store the task breakdown as an AI agent run for tracking
	agent := models.AIAgent{
		UserID:    userID.(uuid.UUID),
		AgentType: "task_breakdown",
		Status:    "completed",
		InputData: map[string]interface{}{
			"goal":         request.Goal,
			"context":      request.Context,
			"time_frame":   request.TimeFrame,
			"max_steps":    request.MaxSteps,
			"priority":     request.Priority,
			"preferences":  request.Preferences,
		},
		OutputData: breakdown,
	}

	if err := ar.db.Create(&agent).Error; err != nil {
		// Log error but don't fail the request
		println("Failed to store agent run:", err.Error())
	}

	c.JSON(http.StatusOK, breakdown)
}

// getChatSessions returns all chat sessions for the user
func (ar *AIRoutes) getChatSessions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessions, err := ar.chatService.GetChatSessions(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
	})
}

// deleteChatSession deletes a chat session
func (ar *AIRoutes) deleteChatSession(c *gin.Context) {
	sessionID := c.Param("id")
	
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err := ar.chatService.DeleteChatSession(c.Request.Context(), userID.(uuid.UUID), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete chat session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat session deleted successfully"})
}

// runReasoningAgent starts a reasoning loop agent
func (ar *AIRoutes) runReasoningAgent(c *gin.Context) {
	var request struct {
		Goal     string `json:"goal" binding:"required"`
		Context  string `json:"context"`
		Strategy string `json:"strategy"` // methodical, exploratory, focused
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Execute reasoning loop in background
	agent, err := ar.reasoningAgentService.ExecuteReasoningLoop(
		c.Request.Context(),
		userID.(uuid.UUID),
		request.Goal,
		request.Context,
		request.Strategy,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start reasoning agent: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"agent_id": agent.ID,
		"status": agent.Status,
		"message": "Reasoning agent started successfully",
	})
}

// getReasoningAgentResult gets the result of a reasoning agent run
func (ar *AIRoutes) getReasoningAgentResult(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var agent models.AIAgent
	if err := ar.db.Where("id = ? AND user_id = ? AND agent_type = ?", agentID, userID, "reasoning_loop").
		First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reasoning agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}
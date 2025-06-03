package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"owlistic/services"
)

type AgentOrchestratorRoutes struct {
	db           *gorm.DB
	orchestrator *services.AgentOrchestrator
}

func NewAgentOrchestratorRoutes(db *gorm.DB) *AgentOrchestratorRoutes {
	return &AgentOrchestratorRoutes{
		db:           db,
		orchestrator: services.NewAgentOrchestrator(db),
	}
}

func (aor *AgentOrchestratorRoutes) RegisterRoutes(router *gin.RouterGroup) {
	agentGroup := router.Group("/agents/orchestrator")
	{
		// Chain execution
		agentGroup.POST("/chains/execute", aor.executeChain)
		agentGroup.GET("/executions", aor.getActiveExecutions)
		agentGroup.GET("/executions/:id", aor.getExecutionStatus)
		
		// Chain management
		agentGroup.GET("/chains", aor.listChains)
		agentGroup.GET("/chains/:id", aor.getChain)
		agentGroup.POST("/chains", aor.createChain)
		agentGroup.PUT("/chains/:id", aor.updateChain)
		agentGroup.DELETE("/chains/:id", aor.deleteChain)
		
		// Agent information
		agentGroup.GET("/agent-types", aor.getAgentTypes)
		
		// Predefined chain templates
		agentGroup.GET("/templates", aor.getChainTemplates)
		agentGroup.POST("/templates/:id/instantiate", aor.instantiateTemplate)
	}
}

// executeChain starts execution of an agent chain
func (aor *AgentOrchestratorRoutes) executeChain(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req services.ChainExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user ID
	req.UserID = userID.(uuid.UUID)

	// Execute chain
	result, err := aor.orchestrator.ExecuteChain(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"execution": result,
		"message":   "Chain execution started",
	})
}

// getActiveExecutions returns all active chain executions
func (aor *AgentOrchestratorRoutes) getActiveExecutions(c *gin.Context) {
	executions := aor.orchestrator.GetActiveExecutions()
	
	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"count":      len(executions),
	})
}

// getExecutionStatus returns the status of a specific execution
func (aor *AgentOrchestratorRoutes) getExecutionStatus(c *gin.Context) {
	executionID := c.Param("id")
	
	result, exists := aor.orchestrator.GetExecutionStatus(executionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"execution": result,
	})
}

// listChains returns all available chains
func (aor *AgentOrchestratorRoutes) listChains(c *gin.Context) {
	// In a real implementation, this would query the database
	// For now, return example chains
	chains := []map[string]interface{}{
		{
			"id":          "research-and-summarize",
			"name":        "Research and Summarize",
			"description": "Search web, analyze results, and create summary",
			"mode":        "sequential",
			"agents":      3,
		},
		{
			"id":          "note-enhancement-pipeline",
			"name":        "Note Enhancement Pipeline",
			"description": "Enhance notes with AI insights, tags, and related content",
			"mode":        "parallel",
			"agents":      3,
		},
		{
			"id":          "task-decomposition",
			"name":        "Task Decomposition",
			"description": "Break down complex goals into actionable tasks",
			"mode":        "sequential",
			"agents":      2,
		},
		{
			"id":          "content-creation",
			"name":        "Content Creation Pipeline",
			"description": "Research, outline, write, and polish content",
			"mode":        "sequential",
			"agents":      4,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"chains": chains,
		"count":  len(chains),
	})
}

// getChain returns details of a specific chain
func (aor *AgentOrchestratorRoutes) getChain(c *gin.Context) {
	chainID := c.Param("id")
	
	// In real implementation, load from database
	chain, err := aor.orchestrator.LoadChainDefinition(chainID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chain not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"chain": chain,
	})
}

// createChain creates a new agent chain
func (aor *AgentOrchestratorRoutes) createChain(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var chain services.AgentChain
	if err := c.ShouldBindJSON(&chain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user ID
	chain.UserID = userID.(uuid.UUID)

	// Create chain
	if err := aor.orchestrator.CreateCustomChain(&chain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"chain":   chain,
		"message": "Chain created successfully",
	})
}

// updateChain updates an existing chain
func (aor *AgentOrchestratorRoutes) updateChain(c *gin.Context) {
	chainID := c.Param("id")
	
	var chain services.AgentChain
	if err := c.ShouldBindJSON(&chain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	chain.ID = chainID
	
	// In real implementation, update in database
	c.JSON(http.StatusOK, gin.H{
		"chain":   chain,
		"message": "Chain updated successfully",
	})
}

// deleteChain deletes a chain
func (aor *AgentOrchestratorRoutes) deleteChain(c *gin.Context) {
	chainID := c.Param("id")
	
	// In real implementation, delete from database
	c.JSON(http.StatusOK, gin.H{
		"message": "Chain deleted successfully",
		"id":      chainID,
	})
}

// getAgentTypes returns all available agent types
func (aor *AgentOrchestratorRoutes) getAgentTypes(c *gin.Context) {
	agentTypes := []map[string]interface{}{
		{
			"type":        "reasoning",
			"name":        "Reasoning Agent",
			"description": "Multi-strategy reasoning and problem solving",
			"input_schema": map[string]interface{}{
				"problem":        "string (required)",
				"strategy":       "string (optional)",
				"max_iterations": "number (optional)",
			},
		},
		{
			"type":        "chat",
			"name":        "Chat Agent",
			"description": "Context-aware conversational agent",
			"input_schema": map[string]interface{}{
				"message": "string (required)",
				"context": "array<string> (optional)",
			},
		},
		{
			"type":        "web_search",
			"name":        "Web Search Agent",
			"description": "Search and retrieve information from the web",
			"input_schema": map[string]interface{}{
				"query":       "string (required)",
				"max_results": "number (optional)",
			},
		},
		{
			"type":        "note_analyzer",
			"name":        "Note Analyzer",
			"description": "Analyze notes and extract insights",
			"input_schema": map[string]interface{}{
				"note_id": "string (required)",
				"action":  "string (optional: analyze, find_related, extract_entities)",
			},
		},
		{
			"type":        "task_planner",
			"name":        "Task Planner",
			"description": "Create detailed task plans from goals",
			"input_schema": map[string]interface{}{
				"goal":         "string (required)",
				"create_tasks": "boolean (optional)",
			},
		},
		{
			"type":        "summarizer",
			"name":        "Summarizer",
			"description": "Create summaries of content",
			"input_schema": map[string]interface{}{
				"content":    "string (required)",
				"style":      "string (optional: concise, bullet, executive, technical)",
				"max_length": "number (optional)",
			},
		},
		{
			"type":        "code_generator",
			"name":        "Code Generator",
			"description": "Generate code based on specifications",
			"input_schema": map[string]interface{}{
				"specification": "string (required)",
				"language":      "string (optional)",
				"style":         "string (optional)",
			},
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"agent_types": agentTypes,
		"count":       len(agentTypes),
	})
}

// getChainTemplates returns predefined chain templates
func (aor *AgentOrchestratorRoutes) getChainTemplates(c *gin.Context) {
	templates := []map[string]interface{}{
		{
			"id":          "research-template",
			"name":        "Research Pipeline",
			"description": "Template for researching a topic comprehensively",
			"parameters": []map[string]string{
				{"name": "topic", "type": "string", "description": "The topic to research"},
				{"name": "depth", "type": "string", "description": "Research depth: shallow, medium, deep"},
			},
		},
		{
			"id":          "writing-template",
			"name":        "Writing Assistant",
			"description": "Template for creating written content",
			"parameters": []map[string]string{
				{"name": "topic", "type": "string", "description": "The topic to write about"},
				{"name": "style", "type": "string", "description": "Writing style: formal, casual, technical"},
				{"name": "length", "type": "number", "description": "Target word count"},
			},
		},
		{
			"id":          "learning-template",
			"name":        "Learning Path Generator",
			"description": "Create a structured learning path for a subject",
			"parameters": []map[string]string{
				{"name": "subject", "type": "string", "description": "The subject to learn"},
				{"name": "level", "type": "string", "description": "Current level: beginner, intermediate, advanced"},
				{"name": "timeframe", "type": "string", "description": "Learning timeframe"},
			},
		},
		{
			"id":          "project-planning",
			"name":        "Project Planning Assistant",
			"description": "Plan and organize a project from start to finish",
			"parameters": []map[string]string{
				{"name": "project_name", "type": "string", "description": "Name of the project"},
				{"name": "goals", "type": "string", "description": "Project goals and objectives"},
				{"name": "constraints", "type": "string", "description": "Time, budget, or resource constraints"},
			},
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"count":     len(templates),
	})
}

// instantiateTemplate creates a chain from a template
func (aor *AgentOrchestratorRoutes) instantiateTemplate(c *gin.Context) {
	templateID := c.Param("id")
	
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create chain based on template
	// In real implementation, this would use the template to create a customized chain
	
	chainID := uuid.New().String()
	chain := map[string]interface{}{
		"id":          chainID,
		"name":        params["name"],
		"description": "Chain created from template: " + templateID,
		"template_id": templateID,
		"parameters":  params,
		"created_at":  time.Now(),
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"chain":   chain,
		"message": "Chain created from template successfully",
	})
}
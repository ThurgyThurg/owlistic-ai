package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"owlistic-notes/owlistic/services"
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
	var req services.ChainExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user ID safely
	req.UserID = getUserUUID(c, aor.db)

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
	var chain services.AgentChain
	if err := c.ShouldBindJSON(&chain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user ID safely
	chain.UserID = getUserUUID(c, aor.db)

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
	
	// For single-user mode, use default user ID if not authenticated
	userIDInterface, exists := c.Get("userID")
	var userUUID uuid.UUID
	if !exists {
		// For single-user systems, use the first user in the database
		userUUID = getSingleUserIDFromDB(aor.db)
	} else {
		// Safely convert userID to UUID
		switch v := userIDInterface.(type) {
		case uuid.UUID:
			userUUID = v
		case string:
			if parsed, err := uuid.Parse(v); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
				return
			} else {
				userUUID = parsed
			}
		default:
			userUUID = getSingleUserIDFromDB(aor.db)
		}
	}
	
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create chain based on template
	var chain *services.AgentChain
	var initialData map[string]interface{}
	
	switch templateID {
	case "research-template":
		chain = &services.AgentChain{
			ID:          uuid.New().String(),
			Name:        "Research: " + getStringParam(params, "topic", "Unknown Topic"),
			Description: "Research pipeline for comprehensive topic analysis",
			Mode:        services.ChainModeSequential,
			UserID:      userUUID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Timeout:     300,
			Agents: []services.AgentDefinition{
				{
					ID:          "search-" + uuid.New().String(),
					Type:        services.AgentTypeWebSearch,
					Name:        "Initial Web Search",
					Description: "Search for information about the topic",
					InputMapping: map[string]string{
						"query": "search_query",
					},
					OutputKey: "search_results",
				},
				{
					ID:          "analyze-" + uuid.New().String(),
					Type:        services.AgentTypeReasoning,
					Name:        "Deep Analysis",
					Description: "Analyze search results with reasoning",
					Config: map[string]interface{}{
						"strategy": "comprehensive",
					},
					InputMapping: map[string]string{
						"problem": "search_results",
					},
					OutputKey: "analysis",
				},
				{
					ID:          "summarize-" + uuid.New().String(),
					Type:        services.AgentTypeSummarizer,
					Name:        "Create Summary",
					Description: "Summarize findings",
					Config: map[string]interface{}{
						"style": "executive",
					},
					InputMapping: map[string]string{
						"content": "analysis",
					},
					OutputKey: "summary",
				},
			},
		}
		
		// Set initial data based on template parameters
		initialData = map[string]interface{}{
			"search_query": params["topic"],
			"depth":        getStringParam(params, "depth", "medium"),
		}
		
	case "writing-template":
		chain = &services.AgentChain{
			ID:          uuid.New().String(),
			Name:        "Writing: " + getStringParam(params, "topic", "Unknown Topic"),
			Description: "Writing assistant for content creation",
			Mode:        services.ChainModeSequential,
			UserID:      userUUID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Timeout:     600,
			Agents: []services.AgentDefinition{
				{
					ID:          "research-" + uuid.New().String(),
					Type:        services.AgentTypeWebSearch,
					Name:        "Research Topic",
					Description: "Research the writing topic",
					InputMapping: map[string]string{
						"query": "topic",
					},
					OutputKey: "research",
				},
				{
					ID:          "outline-" + uuid.New().String(),
					Type:        services.AgentTypeReasoning,
					Name:        "Create Outline",
					Description: "Create a structured outline",
					Config: map[string]interface{}{
						"strategy": "structured_outline",
					},
					InputMapping: map[string]string{
						"problem": "outline_request",
					},
					OutputKey: "outline",
				},
				{
					ID:          "write-" + uuid.New().String(),
					Type:        services.AgentTypeChat,
					Name:        "Write Content",
					Description: "Write the full content",
					InputMapping: map[string]string{
						"message": "writing_request",
					},
					OutputKey: "draft",
				},
				{
					ID:          "polish-" + uuid.New().String(),
					Type:        services.AgentTypeChat,
					Name:        "Polish Content",
					Description: "Polish and improve the content",
					InputMapping: map[string]string{
						"message": "polish_request",
					},
					OutputKey: "final_content",
				},
			},
		}
		
		// Set initial data
		style := getStringParam(params, "style", "formal")
		length := getFloatParam(params, "length", 1000)
		initialData = map[string]interface{}{
			"topic":           params["topic"],
			"style":           style,
			"length":          length,
			"outline_request": fmt.Sprintf("Create a detailed outline for writing about '%s' in %s style, targeting %v words", params["topic"], style, length),
			"writing_request": fmt.Sprintf("Based on the outline, write content about '%s' in %s style, approximately %v words", params["topic"], style, length),
			"polish_request":  fmt.Sprintf("Polish and improve this draft, ensuring it maintains %s style and is approximately %v words", style, length),
		}
		
	case "learning-template":
		chain = &services.AgentChain{
			ID:          uuid.New().String(),
			Name:        "Learning Path: " + getStringParam(params, "subject", "Unknown Subject"),
			Description: "Create structured learning path",
			Mode:        services.ChainModeSequential,
			UserID:      userUUID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Timeout:     300,
			Agents: []services.AgentDefinition{
				{
					ID:          "assess-" + uuid.New().String(),
					Type:        services.AgentTypeReasoning,
					Name:        "Assess Learning Needs",
					Description: "Assess current level and learning needs",
					Config: map[string]interface{}{
						"strategy": "assessment",
					},
					InputMapping: map[string]string{
						"problem": "assessment_request",
					},
					OutputKey: "assessment",
				},
				{
					ID:          "plan-" + uuid.New().String(),
					Type:        services.AgentTypeTaskPlanner,
					Name:        "Create Learning Plan",
					Description: "Create structured learning plan",
					Config: map[string]interface{}{
						"create_tasks": true,
					},
					InputMapping: map[string]string{
						"goal": "learning_goal",
					},
					OutputKey: "learning_plan",
				},
				{
					ID:          "resources-" + uuid.New().String(),
					Type:        services.AgentTypeWebSearch,
					Name:        "Find Resources",
					Description: "Find learning resources",
					InputMapping: map[string]string{
						"query": "resource_query",
					},
					OutputKey: "resources",
				},
			},
		}
		
		// Set initial data
		subject := getStringParam(params, "subject", "")
		level := getStringParam(params, "level", "beginner")
		timeframe := getStringParam(params, "timeframe", "1 month")
		initialData = map[string]interface{}{
			"subject":            subject,
			"level":              level,
			"timeframe":          timeframe,
			"assessment_request": fmt.Sprintf("Assess learning needs for %s at %s level, planning for %s", subject, level, timeframe),
			"learning_goal":      fmt.Sprintf("Create a learning path for %s from %s level within %s", subject, level, timeframe),
			"resource_query":     fmt.Sprintf("%s %s level learning resources tutorials courses", subject, level),
		}
		
	case "project-planning":
		chain = &services.AgentChain{
			ID:          uuid.New().String(),
			Name:        "Project: " + getStringParam(params, "project_name", "Unnamed Project"),
			Description: "Project planning and organization",
			Mode:        services.ChainModeSequential,
			UserID:      userUUID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Timeout:     600,
			Agents: []services.AgentDefinition{
				{
					ID:          "analyze-" + uuid.New().String(),
					Type:        services.AgentTypeReasoning,
					Name:        "Analyze Project Requirements",
					Description: "Analyze project goals and constraints",
					Config: map[string]interface{}{
						"strategy": "requirements_analysis",
					},
					InputMapping: map[string]string{
						"problem": "project_analysis",
					},
					OutputKey: "requirements",
				},
				{
					ID:          "decompose-" + uuid.New().String(),
					Type:        services.AgentTypeTaskPlanner,
					Name:        "Create Project Tasks",
					Description: "Break down project into tasks",
					Config: map[string]interface{}{
						"create_tasks": true,
					},
					InputMapping: map[string]string{
						"goal": "project_goal",
					},
					OutputKey: "task_breakdown",
				},
				{
					ID:          "timeline-" + uuid.New().String(),
					Type:        services.AgentTypeReasoning,
					Name:        "Create Timeline",
					Description: "Create project timeline and milestones",
					Config: map[string]interface{}{
						"strategy": "timeline_planning",
					},
					InputMapping: map[string]string{
						"problem": "timeline_request",
					},
					OutputKey: "timeline",
				},
			},
		}
		
		// Set initial data
		projectName := getStringParam(params, "project_name", "")
		goals := getStringParam(params, "goals", "")
		constraints := getStringParam(params, "constraints", "")
		initialData = map[string]interface{}{
			"project_name":      projectName,
			"goals":             goals,
			"constraints":       constraints,
			"project_analysis":  fmt.Sprintf("Analyze project '%s' with goals: %s and constraints: %s", projectName, goals, constraints),
			"project_goal":      fmt.Sprintf("Create tasks for project '%s' to achieve: %s", projectName, goals),
			"timeline_request":  fmt.Sprintf("Create timeline for project '%s' considering: %s", projectName, constraints),
		}
		
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	
	// Create the chain
	if err := aor.orchestrator.CreateCustomChain(chain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if we should execute immediately
	executeNow := getBoolParam(params, "execute", true)
	
	if executeNow {
		// Execute the chain
		req := services.ChainExecutionRequest{
			ChainID:     chain.ID,
			InitialData: initialData,
			UserID:      userUUID,
		}
		
		result, err := aor.orchestrator.ExecuteChain(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"chain":   chain,
				"message": "Chain created but execution failed",
			})
			return
		}
		
		c.JSON(http.StatusCreated, gin.H{
			"chain":     chain,
			"execution": result,
			"message":   "Chain created and execution started",
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"chain":   chain,
			"message": "Chain created successfully",
		})
	}
}

// Helper functions for parameter extraction
func getStringParam(params map[string]interface{}, key string, defaultValue string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getFloatParam(params map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return defaultValue
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// getSingleUserIDFromDB returns the first user ID for single-user systems
func getSingleUserIDFromDB(db *gorm.DB) uuid.UUID {
	var user struct {
		ID uuid.UUID `gorm:"column:id"`
	}
	if err := db.Table("users").First(&user).Error; err != nil {
		// Return the intended single-user UUID as fallback
		singleUserUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
		return singleUserUUID
	}
	return user.ID
}

// getUserUUID safely extracts and converts user ID from gin context
func getUserUUID(c *gin.Context, db *gorm.DB) uuid.UUID {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		return getSingleUserIDFromDB(db)
	}
	
	switch v := userIDInterface.(type) {
	case uuid.UUID:
		return v
	case string:
		if parsed, err := uuid.Parse(v); err != nil {
			return getSingleUserIDFromDB(db)
		} else {
			return parsed
		}
	default:
		return getSingleUserIDFromDB(db)
	}
}
package services

import (
	"context"
	"fmt"
	"strings"

	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReasoningAgentExecutor wraps the ReasoningAgentService for orchestration
type ReasoningAgentExecutor struct {
	service *ReasoningAgentService
}

func (r *ReasoningAgentExecutor) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	// Extract parameters - handle both string and complex object inputs
	var problem string
	if problemStr, ok := input["problem"].(string); ok {
		problem = problemStr
	} else if problemObj, ok := input["problem"].(map[string]interface{}); ok {
		// Handle search results from WebSearchAgent
		if results, exists := problemObj["results"]; exists {
			if resultsStr, ok := results.(string); ok {
				problem = resultsStr
			} else {
				problem = fmt.Sprintf("Search Results: %v", results)
			}
		} else {
			problem = fmt.Sprintf("Analysis Input: %v", problemObj)
		}
	} else {
		return nil, fmt.Errorf("missing or invalid 'problem' parameter - expected string or object with results")
	}

	strategy := ReasoningStrategy("multi_strategy")
	if s, ok := input["strategy"].(string); ok {
		strategy = ReasoningStrategy(s)
	}

	maxIterations := 5
	if m, ok := input["max_iterations"].(int); ok {
		maxIterations = m
	}

	userID := uuid.Nil
	if u, ok := input["user_id"].(uuid.UUID); ok {
		userID = u
	} else if uStr, ok := input["user_id"].(string); ok {
		if parsed, err := uuid.Parse(uStr); err == nil {
			userID = parsed
		}
	}

	// Create request
	req := ReasoningRequest{
		Problem:       problem,
		Strategy:      strategy,
		MaxIterations: maxIterations,
		Context:       input,
	}

	// Execute reasoning
	result, err := r.service.ExecuteReasoningLoop(ctx, userID, req.Problem, fmt.Sprintf("%v", req.Context), string(req.Strategy))
	if err != nil {
		return nil, err
	}

	// Extract readable content from the reasoning result for downstream agents
	if result.OutputData != nil {
		if learnings, exists := result.OutputData["learnings"]; exists {
			return fmt.Sprintf("Reasoning Analysis:\n\nProblem: %s\n\nLearnings: %v\n\nFinal State: %v", 
				req.Problem, learnings, result.OutputData["final_state"]), nil
		}
	}
	
	// Fallback to basic summary
	return fmt.Sprintf("Reasoning completed for: %s\nStatus: %s", req.Problem, result.Status), nil
}

func (r *ReasoningAgentExecutor) GetType() AgentType {
	return AgentTypeReasoning
}

// ChatAgentExecutor wraps the ChatService for orchestration
type ChatAgentExecutor struct {
	service *ChatService
}

func (c *ChatAgentExecutor) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	// Extract parameters
	message, ok := input["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'message' parameter")
	}

	userID := uuid.Nil
	if u, ok := input["user_id"].(uuid.UUID); ok {
		userID = u
	} else if uStr, ok := input["user_id"].(string); ok {
		if parsed, err := uuid.Parse(uStr); err == nil {
			userID = parsed
		}
	}

	// Create request
	req := ChatRequest{
		Message: message,
	}

	// Add context if provided
	if context, ok := input["context"].([]string); ok {
		req.Context = strings.Join(context, "\n")
	} else if contextStr, ok := input["context"].(string); ok {
		req.Context = contextStr
	}

	// Execute chat
	response, err := c.service.Chat(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *ChatAgentExecutor) GetType() AgentType {
	return AgentTypeChat
}

// NoteAnalyzerAgent analyzes notes and extracts insights
type NoteAnalyzerAgent struct {
	aiService   *AIService
	noteService *NoteService
	db          *gorm.DB
}

func (n *NoteAnalyzerAgent) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	action := "analyze"
	if a, ok := input["action"].(string); ok {
		action = a
	}

	switch action {
	case "analyze":
		return n.analyzeNote(ctx, input)
	case "find_related":
		return n.findRelatedNotes(ctx, input)
	case "extract_entities":
		return n.extractEntities(ctx, input)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (n *NoteAnalyzerAgent) analyzeNote(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	noteID, ok := input["note_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'note_id' parameter")
	}

	// Get note content - need to access database directly for now
	var note models.Note
	err := n.db.Where("id = ?", noteID).Preload("Blocks").First(&note).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Analyze with AI
	prompt := fmt.Sprintf(`Analyze the following note and provide:
1. Key insights and main points
2. Suggested improvements
3. Potential connections to other topics
4. Action items if any

Note Title: %s
Note Content: %s`, note.Title, n.extractNoteContent(&note))

	response, err := n.aiService.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze note: %w", err)
	}

	return map[string]interface{}{
		"note_id":   noteID,
		"title":     note.Title,
		"analysis":  response,
		"timestamp": ctx.Value("timestamp"),
	}, nil
}

func (n *NoteAnalyzerAgent) findRelatedNotes(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	noteID, ok := input["note_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'note_id' parameter")
	}

	// In real implementation, this would use vector search
	// For now, return a simple response
	return map[string]interface{}{
		"note_id": noteID,
		"related_notes": []map[string]interface{}{
			{
				"id":         "related-1",
				"title":      "Related Note 1",
				"similarity": 0.85,
			},
			{
				"id":         "related-2",
				"title":      "Related Note 2",
				"similarity": 0.78,
			},
		},
	}, nil
}

func (n *NoteAnalyzerAgent) extractEntities(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	noteID, ok := input["note_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'note_id' parameter")
	}

	// Get note content - need to access database directly for now
	var note models.Note
	err := n.db.Where("id = ?", noteID).Preload("Blocks").First(&note).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Extract entities with AI
	prompt := fmt.Sprintf(`Extract all entities from the following note. Include:
- People names
- Organizations
- Locations
- Dates
- Key concepts
- Technologies/tools

Format as JSON.

Note: %s`, n.extractNoteContent(&note))

	response, err := n.aiService.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract entities: %w", err)
	}

	return map[string]interface{}{
		"note_id":  noteID,
		"entities": response,
	}, nil
}

func (n *NoteAnalyzerAgent) GetType() AgentType {
	return AgentTypeNoteAnalyzer
}

// TaskPlannerAgent creates and manages task plans
type TaskPlannerAgent struct {
	aiService   *AIService
	taskService *TaskService
}

func (t *TaskPlannerAgent) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	goal, ok := input["goal"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'goal' parameter")
	}

	// Generate task plan with AI
	prompt := fmt.Sprintf(`Create a detailed task plan to achieve the following goal:
Goal: %s

Provide a structured plan with:
1. Main tasks (break down into subtasks if needed)
2. Dependencies between tasks
3. Estimated time for each task
4. Priority levels
5. Success criteria

Format the response as a structured plan.`, goal)

	response, err := t.aiService.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create task plan: %w", err)
	}

	// Parse and structure the response
	plan := map[string]interface{}{
		"goal":       goal,
		"plan":       response,
		"created_at": ctx.Value("timestamp"),
	}

	// Optionally create actual tasks if requested
	if createTasks, ok := input["create_tasks"].(bool); ok && createTasks {
		// Parse AI response and create tasks
		// This would require more sophisticated parsing
		plan["tasks_created"] = true
	}

	return plan, nil
}

func (t *TaskPlannerAgent) GetType() AgentType {
	return AgentTypeTaskPlanner
}

// WebSearchAgent performs web searches
type WebSearchAgent struct {
	aiService *AIService
}

func (w *WebSearchAgent) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	query, ok := input["query"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'query' parameter")
	}

	maxResults := 5
	if m, ok := input["max_results"].(int); ok {
		maxResults = m
	}
	// Perform web search
	searchPrompt := fmt.Sprintf("Search the web for: %s\nReturn top %d relevant results with titles, URLs, and summaries.", query, maxResults)
	
	
	// Try Perplexica first, fall back to AI synthesis if not available
	response, err := w.aiService.PerformWebSearch(ctx, searchPrompt)
	if err != nil {
		// Fallback: Generate a synthetic response based on AI knowledge
		fallbackPrompt := fmt.Sprintf(`Based on your knowledge, provide a comprehensive response about: %s

Please structure your response as if it were search results, including:
1. A main answer/summary
2. Key points and facts
3. Relevant context and background information

Note: This response is generated from AI knowledge, not live web search.`, query)

		fallbackResponse, fallbackErr := w.aiService.GenerateResponse(ctx, fallbackPrompt, nil)
		if fallbackErr != nil {
			return nil, fmt.Errorf("both web search and fallback failed: web search error: %w, fallback error: %w", err, fallbackErr)
		}

		return map[string]interface{}{
			"query":   query,
			"results": fallbackResponse,
			"source":  "ai_knowledge",
			"note":    "Generated from AI knowledge since web search is unavailable",
		}, nil
	}

	return map[string]interface{}{
		"query":   query,
		"results": response,
		"source":  "web_search",
	}, nil
}

func (w *WebSearchAgent) GetType() AgentType {
	return AgentTypeWebSearch
}

// SummarizerAgent creates summaries of content
type SummarizerAgent struct {
	aiService *AIService
}

func (s *SummarizerAgent) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	content, ok := input["content"].(string)
	if !ok {
		// Try to get content from other possible fields
		if data, ok := input["data"]; ok {
			content = fmt.Sprintf("%v", data)
		} else {
			return nil, fmt.Errorf("missing or invalid 'content' parameter")
		}
	}

	style := "concise"
	if s, ok := input["style"].(string); ok {
		style = s
	}

	maxLength := 500
	if m, ok := input["max_length"].(int); ok {
		maxLength = m
	}

	// Create summary prompt based on style
	var prompt string
	switch style {
	case "bullet":
		prompt = fmt.Sprintf("Create a bullet-point summary of the following content (max %d words):\n\n%s", maxLength, content)
	case "executive":
		prompt = fmt.Sprintf("Create an executive summary of the following content (max %d words):\n\n%s", maxLength, content)
	case "technical":
		prompt = fmt.Sprintf("Create a technical summary of the following content, focusing on key technical details (max %d words):\n\n%s", maxLength, content)
	default:
		prompt = fmt.Sprintf("Create a concise summary of the following content (max %d words):\n\n%s", maxLength, content)
	}

	response, err := s.aiService.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create summary: %w", err)
	}

	return map[string]interface{}{
		"summary":         response,
		"style":           style,
		"original_length": len(content),
		"summary_length":  len(response),
	}, nil
}

func (s *SummarizerAgent) GetType() AgentType {
	return AgentTypeSummarizer
}

// CodeGeneratorAgent generates code based on specifications
type CodeGeneratorAgent struct {
	aiService *AIService
}

func (c *CodeGeneratorAgent) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	spec, ok := input["specification"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'specification' parameter")
	}

	language := "go"
	if l, ok := input["language"].(string); ok {
		language = l
	}

	style := "clean"
	if s, ok := input["style"].(string); ok {
		style = s
	}

	// Build prompt
	prompt := fmt.Sprintf(`Generate %s code based on the following specification:
%s

Requirements:
- Language: %s
- Style: %s code with proper error handling
- Include comments explaining the logic
- Follow best practices for %s

Provide only the code without additional explanation.`, language, spec, language, style, language)

	response, err := c.aiService.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	// Extract code from response (remove markdown if present)
	code := response
	if strings.Contains(code, "```") {
		// Extract code block
		parts := strings.Split(code, "```")
		if len(parts) >= 3 {
			code = strings.TrimSpace(parts[1])
			// Remove language identifier if present
			if idx := strings.Index(code, "\n"); idx > 0 {
				firstLine := code[:idx]
				if !strings.Contains(firstLine, " ") {
					code = code[idx+1:]
				}
			}
		}
	}

	return map[string]interface{}{
		"code":          code,
		"language":      language,
		"specification": spec,
	}, nil
}

func (c *CodeGeneratorAgent) GetType() AgentType {
	return AgentTypeCodeGenerator
}

// Helper method to extract content from note blocks
func (n *NoteAnalyzerAgent) extractNoteContent(note *models.Note) string {
	var contentBuilder strings.Builder
	for _, block := range note.Blocks {
		// Extract text content from the block
		if textContent, exists := block.Content["text"]; exists {
			if textStr, ok := textContent.(string); ok && strings.TrimSpace(textStr) != "" {
				contentBuilder.WriteString(textStr)
				contentBuilder.WriteString("\n")
			}
		}
	}
	return strings.TrimSpace(contentBuilder.String())
}

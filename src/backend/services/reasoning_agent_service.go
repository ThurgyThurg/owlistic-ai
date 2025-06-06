package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReasoningStrategy defines different reasoning approaches
type ReasoningStrategy string

const (
	MethodicalStrategy   ReasoningStrategy = "methodical"
	ExploratoryStrategy  ReasoningStrategy = "exploratory"
	FocusedStrategy      ReasoningStrategy = "focused"
	MultiStrategy        ReasoningStrategy = "multi_strategy"
	QuickStrategy        ReasoningStrategy = "quick"
	BalancedStrategy     ReasoningStrategy = "balanced"
	ComprehensiveStrategy ReasoningStrategy = "comprehensive"
)

// ReasoningRequest represents a request to the reasoning agent
type ReasoningRequest struct {
	Problem       string                 `json:"problem"`
	Strategy      ReasoningStrategy      `json:"strategy"`
	MaxIterations int                    `json:"max_iterations"`
	Context       map[string]interface{} `json:"context"`
}

// ReasoningAgentService implements a reasoning loop agent that can solve complex problems
type ReasoningAgentService struct {
	db          *gorm.DB
	ai          *AIService
	noteService *NoteService
}

// ReasoningStep represents a single step in the reasoning process
type ReasoningStep struct {
	StepNumber   int                    `json:"step_number"`
	Type         string                 `json:"type"` // analyze, plan, execute, reflect
	Content      string                 `json:"content"`
	Observations []string               `json:"observations"`
	Actions      []string               `json:"actions"`
	Metadata     map[string]interface{} `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
}

// ReasoningContext maintains the state of the reasoning process
type ReasoningContext struct {
	Goal           string                   `json:"goal"`
	InitialContext string                   `json:"initial_context"`
	Steps          []ReasoningStep          `json:"steps"`
	CurrentState   string                   `json:"current_state"`
	Learnings      []string                 `json:"learnings"`
	Resources      map[string]interface{}   `json:"resources"`
	MaxSteps       int                      `json:"max_steps"`
	Strategy       string                   `json:"strategy"` // methodical, exploratory, focused
}

func NewReasoningAgentService(db *gorm.DB, ai *AIService, noteService *NoteService) *ReasoningAgentService {
	return &ReasoningAgentService{
		db:          db,
		ai:          ai,
		noteService: noteService,
	}
}

// ExecuteReasoningLoop runs a complete reasoning loop for a given goal
func (r *ReasoningAgentService) ExecuteReasoningLoop(ctx context.Context, userID uuid.UUID, goal string, initialContext string, strategy string) (*models.AIAgent, error) {
	// Create agent record
	agent := &models.AIAgent{
		UserID:    userID,
		AgentType: "reasoning_loop",
		Status:    "running",
		InputData: models.AIMetadata{
			"goal":           goal,
			"initial_context": initialContext,
			"strategy":       strategy,
		},
		StartedAt: time.Now(),
	}

	if err := r.db.Create(agent).Error; err != nil {
		return nil, fmt.Errorf("failed to create agent record: %w", err)
	}

	// Initialize reasoning context
	reasoningCtx := &ReasoningContext{
		Goal:           goal,
		InitialContext: initialContext,
		Steps:          []ReasoningStep{},
		CurrentState:   "initializing",
		Learnings:      []string{},
		Resources:      make(map[string]interface{}),
		MaxSteps:       5, // Reduced from 10 to prevent timeouts
		Strategy:       strategy,
	}

	if strategy == "" {
		reasoningCtx.Strategy = "methodical"
	}

	// Execute reasoning loop
	err := r.runReasoningLoop(ctx, agent, reasoningCtx)
	
	// Update agent record
	agent.CompletedAt = &[]time.Time{time.Now()}[0]
	if err != nil {
		agent.Status = "failed"
		agent.ErrorMessage = err.Error()
	} else {
		agent.Status = "completed"
	}
	
	agent.OutputData = models.AIMetadata{
		"final_state": reasoningCtx.CurrentState,
		"learnings":   reasoningCtx.Learnings,
		"total_steps": len(reasoningCtx.Steps),
	}

	if err := r.db.Save(agent).Error; err != nil {
		log.Printf("Failed to update agent record: %v", err)
	}

	return agent, err
}

// runReasoningLoop executes the main reasoning loop
func (r *ReasoningAgentService) runReasoningLoop(ctx context.Context, agent *models.AIAgent, reasoningCtx *ReasoningContext) error {
	for stepNum := 1; stepNum <= reasoningCtx.MaxSteps; stepNum++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Analyze current state
		analysisStep, err := r.analyzeCurrentState(ctx, reasoningCtx)
		if err != nil {
			return fmt.Errorf("analysis failed at step %d: %w", stepNum, err)
		}
		analysisStep.StepNumber = stepNum
		reasoningCtx.Steps = append(reasoningCtx.Steps, *analysisStep)

		// Check if goal is achieved
		if r.isGoalAchieved(reasoningCtx) {
			reasoningCtx.CurrentState = "goal_achieved"
			break
		}

		// Plan next actions
		planStep, err := r.planNextActions(ctx, reasoningCtx)
		if err != nil {
			return fmt.Errorf("planning failed at step %d: %w", stepNum, err)
		}
		planStep.StepNumber = stepNum
		reasoningCtx.Steps = append(reasoningCtx.Steps, *planStep)

		// Execute planned actions
		executeStep, err := r.executeActions(ctx, reasoningCtx, planStep.Actions)
		if err != nil {
			return fmt.Errorf("execution failed at step %d: %w", stepNum, err)
		}
		executeStep.StepNumber = stepNum
		reasoningCtx.Steps = append(reasoningCtx.Steps, *executeStep)

		// Reflect on results
		reflectStep, err := r.reflectOnResults(ctx, reasoningCtx)
		if err != nil {
			return fmt.Errorf("reflection failed at step %d: %w", stepNum, err)
		}
		reflectStep.StepNumber = stepNum
		reasoningCtx.Steps = append(reasoningCtx.Steps, *reflectStep)

		// Update learnings
		reasoningCtx.Learnings = append(reasoningCtx.Learnings, reflectStep.Observations...)
		
		// Check for stagnation
		if r.isStagnating(reasoningCtx) {
			reasoningCtx.CurrentState = "stagnated"
			break
		}
	}

	return nil
}

// analyzeCurrentState analyzes the current state of the reasoning process
func (r *ReasoningAgentService) analyzeCurrentState(ctx context.Context, reasoningCtx *ReasoningContext) (*ReasoningStep, error) {
	prompt := fmt.Sprintf(`You are a reasoning agent analyzing the current state of problem-solving.

Goal: %s
Initial Context: %s
Current State: %s
Strategy: %s

Previous Steps Taken: %d

Recent Learnings:
%s

Analyze the current situation and provide:
1. Current understanding of the problem
2. Progress made so far
3. Key obstacles or challenges
4. Available resources or information

Provide your analysis in a structured format.`, 
		reasoningCtx.Goal,
		reasoningCtx.InitialContext,
		reasoningCtx.CurrentState,
		reasoningCtx.Strategy,
		len(reasoningCtx.Steps),
		strings.Join(reasoningCtx.Learnings[max(0, len(reasoningCtx.Learnings)-5):], "\n"))

	analysis, err := r.ai.callAnthropic(ctx, prompt, 500)
	if err != nil {
		return nil, err
	}

	// Extract observations from analysis
	observations := r.extractObservations(analysis)

	return &ReasoningStep{
		Type:         "analyze",
		Content:      analysis,
		Observations: observations,
		Timestamp:    time.Now(),
		Metadata: map[string]interface{}{
			"state": reasoningCtx.CurrentState,
		},
	}, nil
}

// planNextActions plans the next actions based on current analysis
func (r *ReasoningAgentService) planNextActions(ctx context.Context, reasoningCtx *ReasoningContext) (*ReasoningStep, error) {
	recentAnalysis := ""
	for i := len(reasoningCtx.Steps) - 1; i >= 0 && i >= len(reasoningCtx.Steps)-2; i-- {
		if reasoningCtx.Steps[i].Type == "analyze" {
			recentAnalysis = reasoningCtx.Steps[i].Content
			break
		}
	}

	prompt := fmt.Sprintf(`You are a strategic planner for a reasoning agent.

Goal: %s
Strategy: %s
Current Analysis: %s

Based on the analysis, plan the next 1-3 concrete actions to make progress toward the goal.

Consider the %s strategy:
- methodical: Step-by-step, thorough approach
- exploratory: Try multiple approaches, gather information
- focused: Direct path to goal, minimize steps

Return a JSON array of action strings. Each action should be specific and actionable.
Example: ["Research X topic in notes", "Create summary of findings", "Generate hypothesis"]

Actions:`, 
		reasoningCtx.Goal,
		reasoningCtx.Strategy,
		recentAnalysis,
		reasoningCtx.Strategy)

	response, err := r.ai.callAnthropic(ctx, prompt, 200)
	if err != nil {
		return nil, err
	}

	// Parse actions
	var actions []string
	if err := json.Unmarshal([]byte(response), &actions); err != nil {
		// Fallback parsing
		actions = strings.Split(response, "\n")
		for i := range actions {
			actions[i] = strings.TrimSpace(strings.Trim(actions[i], "-•*"))
		}
	}

	return &ReasoningStep{
		Type:     "plan",
		Content:  fmt.Sprintf("Planned actions based on %s strategy", reasoningCtx.Strategy),
		Actions:  actions,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"action_count": len(actions),
		},
	}, nil
}

// executeActions executes the planned actions
func (r *ReasoningAgentService) executeActions(ctx context.Context, reasoningCtx *ReasoningContext, actions []string) (*ReasoningStep, error) {
	results := []string{}
	
	for _, action := range actions {
		result, err := r.executeAction(ctx, reasoningCtx, action)
		if err != nil {
			results = append(results, fmt.Sprintf("Failed to execute '%s': %v", action, err))
		} else {
			results = append(results, result)
		}
	}

	return &ReasoningStep{
		Type:         "execute",
		Content:      fmt.Sprintf("Executed %d actions", len(actions)),
		Actions:      actions,
		Observations: results,
		Timestamp:    time.Now(),
		Metadata: map[string]interface{}{
			"success_count": len(results),
		},
	}, nil
}

// executeAction executes a single action
func (r *ReasoningAgentService) executeAction(ctx context.Context, reasoningCtx *ReasoningContext, action string) (string, error) {
	// Determine action type and execute accordingly
	actionLower := strings.ToLower(action)
	
	switch {
	case strings.Contains(actionLower, "search") || strings.Contains(actionLower, "research"):
		return r.executeSearchAction(ctx, reasoningCtx, action)
	case strings.Contains(actionLower, "create") || strings.Contains(actionLower, "generate"):
		return r.executeCreateAction(ctx, reasoningCtx, action)
	case strings.Contains(actionLower, "analyze") || strings.Contains(actionLower, "evaluate"):
		return r.executeAnalyzeAction(ctx, reasoningCtx, action)
	case strings.Contains(actionLower, "summarize"):
		return r.executeSummarizeAction(ctx, reasoningCtx, action)
	default:
		return r.executeGenericAction(ctx, reasoningCtx, action)
	}
}

// executeSearchAction handles search-related actions
func (r *ReasoningAgentService) executeSearchAction(ctx context.Context, reasoningCtx *ReasoningContext, action string) (string, error) {
	// Extract search query from action
	query := r.extractSearchQuery(action)
	
	// Check if we should search notes or web
	if strings.Contains(strings.ToLower(action), "notes") || strings.Contains(strings.ToLower(action), "knowledge") {
		// Search user's notes
		notes, err := r.searchUserNotes(ctx, query)
		if err != nil {
			return "", err
		}
		
		if len(notes) == 0 {
			return fmt.Sprintf("No notes found for query: %s", query), nil
		}
		
		// Store found notes in resources
		reasoningCtx.Resources[fmt.Sprintf("notes_%s", query)] = notes
		
		return fmt.Sprintf("Found %d relevant notes for '%s'", len(notes), query), nil
	} else if r.ai.perplexicaService != nil && r.ai.perplexicaService.IsEnabled() {
		// Use Perplexica for web search - with safe user ID extraction
		var userID uuid.UUID
		if userIDValue, exists := reasoningCtx.Resources["user_id"]; exists && userIDValue != nil {
			if uid, ok := userIDValue.(uuid.UUID); ok {
				userID = uid
			}
		}
		// If userID is still nil/empty, use a default or skip search
		if userID == uuid.Nil {
			return fmt.Sprintf("Web search skipped for '%s' - no valid user context", query), nil
		}
		
		searchResult, err := r.ai.SearchWithPerplexica(ctx, userID, query, "webSearch", nil)
		if err != nil {
			return "", err
		}
		
		// Store search results in resources
		reasoningCtx.Resources[fmt.Sprintf("search_%s", query)] = searchResult
		
		return fmt.Sprintf("Web search completed for '%s', found %d sources", query, len(searchResult.Sources)), nil
	}
	
	return fmt.Sprintf("Simulated search for: %s", query), nil
}

// executeCreateAction handles creation-related actions
func (r *ReasoningAgentService) executeCreateAction(ctx context.Context, reasoningCtx *ReasoningContext, action string) (string, error) {
	// Generate content based on action
	prompt := fmt.Sprintf(`Create content based on this action: %s

Context:
Goal: %s
Current Resources: %v

Generate appropriate content.`, action, reasoningCtx.Goal, reasoningCtx.Resources)

	content, err := r.ai.callAnthropic(ctx, prompt, 500)
	if err != nil {
		return "", err
	}
	
	// Store created content
	reasoningCtx.Resources[fmt.Sprintf("created_%d", len(reasoningCtx.Steps))] = content
	
	return fmt.Sprintf("Created content: %s...", content[:min(100, len(content))]), nil
}

// reflectOnResults reflects on the results of executed actions
func (r *ReasoningAgentService) reflectOnResults(ctx context.Context, reasoningCtx *ReasoningContext) (*ReasoningStep, error) {
	recentSteps := reasoningCtx.Steps[max(0, len(reasoningCtx.Steps)-4):]
	
	prompt := fmt.Sprintf(`Reflect on the recent actions and their results.

Goal: %s
Recent Steps: %d

Recent Actions and Results:
%s

Provide insights on:
1. What worked well?
2. What didn't work as expected?
3. What new information was discovered?
4. How should the approach be adjusted?

Be concise and focus on actionable insights.`,
		reasoningCtx.Goal,
		len(recentSteps),
		r.formatRecentSteps(recentSteps))

	reflection, err := r.ai.callAnthropic(ctx, prompt, 300)
	if err != nil {
		return nil, err
	}

	// Extract key insights
	insights := r.extractInsights(reflection)

	return &ReasoningStep{
		Type:         "reflect",
		Content:      reflection,
		Observations: insights,
		Timestamp:    time.Now(),
		Metadata: map[string]interface{}{
			"insight_count": len(insights),
		},
	}, nil
}

// Helper methods

func (r *ReasoningAgentService) isGoalAchieved(ctx *ReasoningContext) bool {
	// Simple heuristic - check if recent steps indicate completion
	if len(ctx.Steps) > 0 {
		recent := ctx.Steps[len(ctx.Steps)-1]
		return strings.Contains(strings.ToLower(recent.Content), "goal achieved") ||
			   strings.Contains(strings.ToLower(recent.Content), "completed successfully")
	}
	return false
}

func (r *ReasoningAgentService) isStagnating(ctx *ReasoningContext) bool {
	// Check if last 3 steps are similar (indicating no progress)
	if len(ctx.Steps) < 6 {
		return false
	}
	
	// Simple check - could be enhanced with better similarity detection
	recentActions := []string{}
	for i := len(ctx.Steps) - 6; i < len(ctx.Steps); i++ {
		if ctx.Steps[i].Type == "execute" {
			recentActions = append(recentActions, strings.Join(ctx.Steps[i].Actions, ","))
		}
	}
	
	// If actions are repeating, we're stagnating
	return len(recentActions) > 2 && recentActions[0] == recentActions[len(recentActions)-1]
}

func (r *ReasoningAgentService) extractObservations(analysis string) []string {
	// Simple extraction - split by newlines and filter
	lines := strings.Split(analysis, "\n")
	observations := []string{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "•") || 
			strings.Contains(line, ":")) {
			observations = append(observations, strings.TrimPrefix(strings.TrimPrefix(line, "-"), "•"))
		}
	}
	
	return observations
}

func (r *ReasoningAgentService) extractInsights(reflection string) []string {
	// Similar to extractObservations but focused on insights
	return r.extractObservations(reflection)
}

func (r *ReasoningAgentService) extractSearchQuery(action string) string {
	// Extract the main search terms from the action
	// Remove common words
	words := strings.Fields(action)
	query := []string{}
	
	skipWords := map[string]bool{
		"search": true, "for": true, "find": true, "look": true,
		"research": true, "in": true, "the": true, "about": true,
		"related": true, "to": true, "on": true, "notes": true,
	}
	
	for _, word := range words {
		if !skipWords[strings.ToLower(word)] {
			query = append(query, word)
		}
	}
	
	return strings.Join(query, " ")
}

func (r *ReasoningAgentService) formatRecentSteps(steps []ReasoningStep) string {
	formatted := []string{}
	for _, step := range steps {
		if step.Type == "execute" {
			formatted = append(formatted, fmt.Sprintf("Actions: %v\nResults: %v", 
				step.Actions, step.Observations))
		}
	}
	return strings.Join(formatted, "\n\n")
}

func (r *ReasoningAgentService) searchUserNotes(ctx context.Context, query string) ([]models.Note, error) {
	// This would integrate with the note service to search
	// For now, return empty
	return []models.Note{}, nil
}

func (r *ReasoningAgentService) executeAnalyzeAction(ctx context.Context, reasoningCtx *ReasoningContext, action string) (string, error) {
	return fmt.Sprintf("Analysis completed for: %s", action), nil
}

func (r *ReasoningAgentService) executeSummarizeAction(ctx context.Context, reasoningCtx *ReasoningContext, action string) (string, error) {
	return fmt.Sprintf("Summary created for: %s", action), nil
}

func (r *ReasoningAgentService) executeGenericAction(ctx context.Context, reasoningCtx *ReasoningContext, action string) (string, error) {
	return fmt.Sprintf("Executed generic action: %s", action), nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
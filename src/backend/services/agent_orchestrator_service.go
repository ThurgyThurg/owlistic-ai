package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"owlistic-notes/owlistic/database"
	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentType defines the type of agent
type AgentType string

const (
	AgentTypeReasoning      AgentType = "reasoning"
	AgentTypeChat           AgentType = "chat"
	AgentTypeWebSearch      AgentType = "web_search"
	AgentTypeNoteAnalyzer   AgentType = "note_analyzer"
	AgentTypeTaskPlanner    AgentType = "task_planner"
	AgentTypeCodeGenerator  AgentType = "code_generator"
	AgentTypeSummarizer     AgentType = "summarizer"
)

// ChainExecutionMode defines how agents in a chain are executed
type ChainExecutionMode string

const (
	ChainModeSequential ChainExecutionMode = "sequential" // Execute agents one after another
	ChainModeParallel   ChainExecutionMode = "parallel"   // Execute agents simultaneously
	ChainModeConditional ChainExecutionMode = "conditional" // Execute based on conditions
)

// AgentDefinition defines a single agent in a chain
type AgentDefinition struct {
	ID          string                 `json:"id"`
	Type        AgentType              `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	InputMapping map[string]string     `json:"input_mapping"` // Maps chain data to agent inputs
	OutputKey   string                 `json:"output_key"`    // Key to store agent output in chain data
	Conditions  []ChainCondition       `json:"conditions"`    // For conditional execution
	RetryPolicy RetryPolicy            `json:"retry_policy"`
}

// ChainCondition defines when an agent should be executed
type ChainCondition struct {
	Type     string      `json:"type"`     // "equals", "contains", "exists", "greater_than", etc.
	DataKey  string      `json:"data_key"` // Key in chain data to check
	Value    interface{} `json:"value"`    // Value to compare against
	Operator string      `json:"operator"` // "and", "or" for multiple conditions
}

// RetryPolicy defines how to retry failed agent executions
type RetryPolicy struct {
	MaxRetries     int           `json:"max_retries"`
	BackoffSeconds int           `json:"backoff_seconds"`
	RetryOnErrors  []string      `json:"retry_on_errors"` // Specific error types to retry
}

// AgentChain defines a chain of agents to execute
type AgentChain struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Mode        ChainExecutionMode `json:"mode"`
	Agents      []AgentDefinition  `json:"agents"`
	Timeout     int                `json:"timeout_seconds"`
	UserID      uuid.UUID          `json:"user_id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// ChainExecutionRequest represents a request to execute an agent chain
type ChainExecutionRequest struct {
	ChainID     string                 `json:"chain_id"`
	InitialData map[string]interface{} `json:"initial_data"`
	UserID      uuid.UUID              `json:"user_id"`
}

// ChainExecutionResult represents the result of a chain execution
type ChainExecutionResult struct {
	ID          string                  `json:"id"`
	ChainID     string                  `json:"chain_id"`
	Status      string                  `json:"status"` // "running", "completed", "failed", "timeout"
	StartTime   time.Time               `json:"start_time"`
	EndTime     *time.Time              `json:"end_time"`
	Results     map[string]interface{}  `json:"results"`
	Errors      []AgentExecutionError   `json:"errors"`
	ExecutionLog []AgentExecutionLog    `json:"execution_log"`
}

// AgentExecutionError represents an error during agent execution
type AgentExecutionError struct {
	AgentID   string    `json:"agent_id"`
	AgentName string    `json:"agent_name"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// AgentExecutionLog represents a log entry for agent execution
type AgentExecutionLog struct {
	AgentID   string                 `json:"agent_id"`
	AgentName string                 `json:"agent_name"`
	Status    string                 `json:"status"`
	Input     map[string]interface{} `json:"input"`
	Output    interface{}            `json:"output"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  float64                `json:"duration_seconds"`
}

// AgentOrchestrator manages agent chains and orchestration
type AgentOrchestrator struct {
	db                  *gorm.DB
	reasoningAgent      *ReasoningAgentService
	chatService         *ChatService
	noteService         *NoteService
	notebookService     *NotebookService
	taskService         *TaskService
	aiService           *AIService
	activeExecutions    map[string]*ChainExecutionResult
	registeredAgents    map[AgentType]AgentExecutor
}

// AgentExecutor interface that all agents must implement
type AgentExecutor interface {
	Execute(ctx context.Context, input map[string]interface{}) (interface{}, error)
	GetType() AgentType
}

// NewAgentOrchestrator creates a new agent orchestrator
func NewAgentOrchestrator(db *gorm.DB) *AgentOrchestrator {
	orchestrator := &AgentOrchestrator{
		db:               db,
		activeExecutions: make(map[string]*ChainExecutionResult),
		registeredAgents: make(map[AgentType]AgentExecutor),
	}

	// Initialize services
	orchestrator.aiService = NewAIService(db)
	orchestrator.noteService = NewNoteService().(*NoteService)
	orchestrator.notebookService = NewNotebookService().(*NotebookService)
	orchestrator.taskService = NewTaskService().(*TaskService)
	orchestrator.reasoningAgent = NewReasoningAgentService(db, orchestrator.aiService, orchestrator.noteService)
	orchestrator.chatService = NewChatService(db, orchestrator.aiService, orchestrator.noteService)

	// Register built-in agents
	orchestrator.registerBuiltInAgents()

	return orchestrator
}

// registerBuiltInAgents registers all built-in agent types
func (o *AgentOrchestrator) registerBuiltInAgents() {
	// Register reasoning agent
	o.registeredAgents[AgentTypeReasoning] = &ReasoningAgentExecutor{
		service: o.reasoningAgent,
	}

	// Register chat agent
	o.registeredAgents[AgentTypeChat] = &ChatAgentExecutor{
		service: o.chatService,
	}

	// Register note analyzer agent
	o.registeredAgents[AgentTypeNoteAnalyzer] = &NoteAnalyzerAgent{
		aiService:   o.aiService,
		noteService: o.noteService,
		db:          o.db,
	}

	// Register task planner agent
	o.registeredAgents[AgentTypeTaskPlanner] = &TaskPlannerAgent{
		aiService:   o.aiService,
		taskService: o.taskService,
	}

	// Register web search agent
	o.registeredAgents[AgentTypeWebSearch] = &WebSearchAgent{
		aiService: o.aiService,
	}

	// Register summarizer agent
	o.registeredAgents[AgentTypeSummarizer] = &SummarizerAgent{
		aiService: o.aiService,
	}

	// Register code generator agent
	o.registeredAgents[AgentTypeCodeGenerator] = &CodeGeneratorAgent{
		aiService: o.aiService,
	}
}

// GetAgent returns a registered agent executor by type
func (o *AgentOrchestrator) GetAgent(agentType AgentType) (AgentExecutor, bool) {
	agent, exists := o.registeredAgents[agentType]
	return agent, exists
}

// ExecuteChain executes an agent chain
func (o *AgentOrchestrator) ExecuteChain(ctx context.Context, req ChainExecutionRequest) (*ChainExecutionResult, error) {
	// Create execution result
	result := &ChainExecutionResult{
		ID:           uuid.New().String(),
		ChainID:      req.ChainID,
		Status:       "running",
		StartTime:    time.Now(),
		Results:      make(map[string]interface{}),
		Errors:       []AgentExecutionError{},
		ExecutionLog: []AgentExecutionLog{},
	}

	// Store active execution
	o.activeExecutions[result.ID] = result

	// Load chain definition (in real implementation, load from database)
	chain, err := o.LoadChainDefinition(req.ChainID)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, AgentExecutionError{
			Error:     fmt.Sprintf("Failed to load chain: %v", err),
			Timestamp: time.Now(),
		})
		return result, err
	}

	// Create timeout context
	timeoutDuration := time.Duration(chain.Timeout) * time.Second
	if timeoutDuration == 0 {
		timeoutDuration = 5 * time.Minute // Default timeout
	}
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	// Initialize chain data with request data
	chainData := make(map[string]interface{})
	for k, v := range req.InitialData {
		chainData[k] = v
	}

	// Execute based on mode
	fmt.Printf("Executing chain %s (%s) with %d agents in %s mode\n", chain.Name, chain.ID, len(chain.Agents), chain.Mode)
	switch chain.Mode {
	case ChainModeSequential:
		err = o.executeSequential(ctxWithTimeout, chain, chainData, result)
	case ChainModeParallel:
		err = o.executeParallel(ctxWithTimeout, chain, chainData, result)
	case ChainModeConditional:
		err = o.executeConditional(ctxWithTimeout, chain, chainData, result)
	default:
		err = fmt.Errorf("unsupported chain mode: %s", chain.Mode)
	}

	// Update final status
	endTime := time.Now()
	result.EndTime = &endTime
	if err != nil {
		result.Status = "failed"
		fmt.Printf("Chain %s failed: %v\n", chain.Name, err)
	} else {
		result.Status = "completed"
		fmt.Printf("Chain %s completed successfully in %.2fs\n", chain.Name, endTime.Sub(result.StartTime).Seconds())
	}

	// Store final results
	result.Results = chainData

	// Save execution result to database
	if saveErr := o.SaveExecutionResult(result); saveErr != nil {
		fmt.Printf("Failed to save execution result to database: %v\n", saveErr)
	}

	// Save execution as notebook and notes if successful
	if err == nil && req.UserID != uuid.Nil {
		go func() {
			if notebookID, noteIDs, saveErr := o.saveExecutionAsNotebook(context.Background(), req.UserID, chain, result); saveErr != nil {
				fmt.Printf("Failed to save execution as notebook: %v\n", saveErr)
			} else {
				fmt.Printf("Saved execution as notebook %s with %d notes\n", notebookID, len(noteIDs))
			}
		}()
	}

	// Clean up active execution
	delete(o.activeExecutions, result.ID)

	return result, err
}

// executeSequential executes agents one after another
func (o *AgentOrchestrator) executeSequential(ctx context.Context, chain *AgentChain, chainData map[string]interface{}, result *ChainExecutionResult) error {
	for _, agentDef := range chain.Agents {
		// Check if we should execute based on conditions
		if !o.checkConditions(agentDef.Conditions, chainData) {
			continue
		}

		// Execute agent
		if err := o.executeAgent(ctx, agentDef, chainData, result); err != nil {
			return fmt.Errorf("agent %s failed: %w", agentDef.Name, err)
		}
	}
	return nil
}

// executeParallel executes agents simultaneously
func (o *AgentOrchestrator) executeParallel(ctx context.Context, chain *AgentChain, chainData map[string]interface{}, result *ChainExecutionResult) error {
	// Create channels for results and errors
	type agentResult struct {
		agentID string
		output  interface{}
		err     error
	}
	
	resultChan := make(chan agentResult, len(chain.Agents))
	
	// Execute all agents in parallel
	for _, agentDef := range chain.Agents {
		go func(def AgentDefinition) {
			// Create a copy of chain data for thread safety
			localData := make(map[string]interface{})
			for k, v := range chainData {
				localData[k] = v
			}
			
			// Check conditions
			if !o.checkConditions(def.Conditions, localData) {
				resultChan <- agentResult{agentID: def.ID, output: nil, err: nil}
				return
			}
			
			// Execute agent
			output, err := o.executeSingleAgent(ctx, def, localData)
			resultChan <- agentResult{
				agentID: def.ID,
				output:  output,
				err:     err,
			}
		}(agentDef)
	}
	
	// Collect results
	hasErrors := false
	for i := 0; i < len(chain.Agents); i++ {
		select {
		case res := <-resultChan:
			if res.err != nil {
				hasErrors = true
				result.Errors = append(result.Errors, AgentExecutionError{
					AgentID:   res.agentID,
					Error:     res.err.Error(),
					Timestamp: time.Now(),
				})
			} else if res.output != nil {
				// Find agent definition and store output
				for _, agentDef := range chain.Agents {
					if agentDef.ID == res.agentID && agentDef.OutputKey != "" {
						chainData[agentDef.OutputKey] = res.output
						break
					}
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	if hasErrors {
		return fmt.Errorf("one or more agents failed in parallel execution")
	}
	
	return nil
}

// executeConditional executes agents based on conditions
func (o *AgentOrchestrator) executeConditional(ctx context.Context, chain *AgentChain, chainData map[string]interface{}, result *ChainExecutionResult) error {
	// For conditional execution, we evaluate each agent's conditions
	// and execute only those that meet the criteria
	for _, agentDef := range chain.Agents {
		if o.checkConditions(agentDef.Conditions, chainData) {
			if err := o.executeAgent(ctx, agentDef, chainData, result); err != nil {
				// In conditional mode, we might want to continue despite errors
				result.Errors = append(result.Errors, AgentExecutionError{
					AgentID:   agentDef.ID,
					AgentName: agentDef.Name,
					Error:     err.Error(),
					Timestamp: time.Now(),
				})
				// Continue with next agent instead of failing entire chain
				continue
			}
		}
	}
	return nil
}

// executeAgent executes a single agent with retry logic
func (o *AgentOrchestrator) executeAgent(ctx context.Context, agentDef AgentDefinition, chainData map[string]interface{}, result *ChainExecutionResult) error {
	var lastErr error
	maxRetries := 1
	if agentDef.RetryPolicy.MaxRetries > 0 {
		maxRetries = agentDef.RetryPolicy.MaxRetries + 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			backoff := time.Duration(agentDef.RetryPolicy.BackoffSeconds) * time.Second
			if backoff == 0 {
				backoff = time.Duration(attempt) * time.Second
			}
			time.Sleep(backoff)
		}

		output, err := o.executeSingleAgent(ctx, agentDef, chainData)
		if err == nil {
			// Success - store output if key is specified
			if agentDef.OutputKey != "" {
				chainData[agentDef.OutputKey] = output
			}
			return nil
		}

		lastErr = err
		
		// Check if we should retry this error
		if !o.shouldRetry(err, agentDef.RetryPolicy) {
			break
		}
	}

	return lastErr
}

// executeSingleAgent executes a single agent without retry logic
func (o *AgentOrchestrator) executeSingleAgent(ctx context.Context, agentDef AgentDefinition, chainData map[string]interface{}) (interface{}, error) {
	startTime := time.Now()

	// Get the agent executor
	executor, exists := o.registeredAgents[agentDef.Type]
	if !exists {
		return nil, fmt.Errorf("unknown agent type: %s", agentDef.Type)
	}

	// Prepare input by mapping chain data to agent inputs
	input := make(map[string]interface{})
	
	// Add config values
	for k, v := range agentDef.Config {
		input[k] = v
	}
	
	// Map chain data to agent inputs
	for agentKey, chainKey := range agentDef.InputMapping {
		if value, exists := chainData[chainKey]; exists {
			input[agentKey] = value
		}
	}

	// Execute the agent
	output, err := executor.Execute(ctx, input)

	// Log execution
	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds()
	
	log := AgentExecutionLog{
		AgentID:   agentDef.ID,
		AgentName: agentDef.Name,
		Status:    "completed",
		Input:     input,
		Output:    output,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  duration,
	}
	
	if err != nil {
		log.Status = "failed"
		log.Output = err.Error()
		// Log the error for debugging
		fmt.Printf("Agent %s (%s) failed: %v\n", agentDef.Name, agentDef.ID, err)
	} else {
		// Log successful execution
		fmt.Printf("Agent %s (%s) completed successfully in %.2fs\n", agentDef.Name, agentDef.ID, duration)
	}

	return output, err
}

// checkConditions evaluates conditions for agent execution
func (o *AgentOrchestrator) checkConditions(conditions []ChainCondition, chainData map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true // No conditions means always execute
	}

	// Evaluate conditions with AND logic by default
	for _, condition := range conditions {
		if !o.evaluateCondition(condition, chainData) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition
func (o *AgentOrchestrator) evaluateCondition(condition ChainCondition, chainData map[string]interface{}) bool {
	value, exists := chainData[condition.DataKey]
	
	switch condition.Type {
	case "exists":
		return exists
	case "not_exists":
		return !exists
	case "equals":
		return exists && value == condition.Value
	case "not_equals":
		return !exists || value != condition.Value
	case "contains":
		if str, ok := value.(string); ok {
			if searchStr, ok := condition.Value.(string); ok {
				return strings.Contains(str, searchStr)
			}
		}
		return false
	case "greater_than":
		return compareNumeric(value, condition.Value, ">")
	case "less_than":
		return compareNumeric(value, condition.Value, "<")
	case "in":
		if arr, ok := condition.Value.([]interface{}); ok {
			for _, item := range arr {
				if value == item {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

// compareNumeric compares two numeric values
func compareNumeric(a, b interface{}, op string) bool {
	// Convert to float64 for comparison
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	
	if !aOk || !bOk {
		return false
	}
	
	switch op {
	case ">":
		return aFloat > bFloat
	case "<":
		return aFloat < bFloat
	case ">=":
		return aFloat >= bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		return 0, false
	}
}

// shouldRetry determines if an error should trigger a retry
func (o *AgentOrchestrator) shouldRetry(err error, policy RetryPolicy) bool {
	if len(policy.RetryOnErrors) == 0 {
		return true // Retry all errors if no specific errors specified
	}

	errStr := err.Error()
	for _, retryErr := range policy.RetryOnErrors {
		if strings.Contains(errStr, retryErr) {
			return true
		}
	}

	return false
}

// saveChainAsAIAgent saves a chain definition as an AIAgent for execution tracking
func (o *AgentOrchestrator) saveChainAsAIAgent(chain *AgentChain, userID uuid.UUID) (*models.AIAgent, error) {
	// Convert chain to JSON for storage in input_data
	chainData := make(models.AIMetadata)
	chainBytes, _ := json.Marshal(chain)
	json.Unmarshal(chainBytes, &chainData)
	
	aiAgent := &models.AIAgent{
		ID:        uuid.MustParse(chain.ID),
		UserID:    userID,
		AgentType: "agent_chain",
		Status:    "running",
		InputData: chainData,
	}
	
	err := o.db.Create(aiAgent).Error
	if err != nil {
		return nil, fmt.Errorf("failed to save chain as AIAgent: %w", err)
	}
	
	return aiAgent, nil
}

// LoadChainDefinition loads a chain definition from database
func (o *AgentOrchestrator) LoadChainDefinition(chainID string) (*AgentChain, error) {
	// Try to load from AIAgent table first
	var aiAgent models.AIAgent
	err := o.db.Where("id = ? AND agent_type = ?", chainID, "agent_chain").First(&aiAgent).Error
	if err == nil {
		fmt.Printf("Found chain in AIAgent table: %s\n", chainID)
		// Convert back from JSON
		var chain AgentChain
		chainBytes, _ := json.Marshal(aiAgent.InputData)
		if err := json.Unmarshal(chainBytes, &chain); err == nil {
			return &chain, nil
		}
	}
	
	// If not found in database, check hardcoded chains for backward compatibility
	fmt.Printf("Chain not found in database, checking hardcoded chains: %s\n", chainID)
	
	switch chainID {
	case "research-and-summarize":
		return &AgentChain{
			ID:          chainID,
			Name:        "Research and Summarize",
			Description: "Search web, analyze results, and create summary",
			Mode:        ChainModeSequential,
			Timeout:     300,
			Agents: []AgentDefinition{
				{
					ID:          "search",
					Type:        AgentTypeWebSearch,
					Name:        "Web Search",
					Description: "Search for information on the web",
					InputMapping: map[string]string{
						"query": "search_query",
					},
					OutputKey: "search_results",
				},
				{
					ID:          "analyze",
					Type:        AgentTypeReasoning,
					Name:        "Analyze Results",
					Description: "Analyze search results",
					InputMapping: map[string]string{
						"problem": "search_results",
					},
					OutputKey: "analysis",
				},
				{
					ID:          "summarize",
					Type:        AgentTypeSummarizer,
					Name:        "Summarize Findings",
					Description: "Create a summary of findings",
					InputMapping: map[string]string{
						"content": "analysis",
					},
					OutputKey: "summary",
				},
			},
		}, nil
		
	case "note-enhancement-pipeline":
		return &AgentChain{
			ID:          chainID,
			Name:        "Note Enhancement Pipeline",
			Description: "Enhance notes with AI insights, tags, and related content",
			Mode:        ChainModeParallel,
			Timeout:     180,
			Agents: []AgentDefinition{
				{
					ID:          "analyze-note",
					Type:        AgentTypeNoteAnalyzer,
					Name:        "Note Analyzer",
					Description: "Extract key insights from note",
					InputMapping: map[string]string{
						"note_id": "note_id",
					},
					OutputKey: "note_analysis",
				},
				{
					ID:          "generate-tags",
					Type:        AgentTypeSummarizer,
					Name:        "Tag Generator",
					Description: "Generate relevant tags",
					InputMapping: map[string]string{
						"content": "note_content",
					},
					OutputKey: "tags",
				},
				{
					ID:          "find-related",
					Type:        AgentTypeNoteAnalyzer,
					Name:        "Related Content Finder",
					Description: "Find related notes and content",
					Config: map[string]interface{}{
						"action": "find_related",
					},
					InputMapping: map[string]string{
						"note_id": "note_id",
					},
					OutputKey: "related_content",
				},
			},
		}, nil
		
	default:
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}
}

// GetActiveExecutions returns all active chain executions
func (o *AgentOrchestrator) GetActiveExecutions() map[string]*ChainExecutionResult {
	results := make(map[string]*ChainExecutionResult)
	for k, v := range o.activeExecutions {
		results[k] = v
	}
	return results
}

// GetExecutionStatus returns the status of a specific execution
func (o *AgentOrchestrator) GetExecutionStatus(executionID string) (*ChainExecutionResult, bool) {
	result, exists := o.activeExecutions[executionID]
	return result, exists
}

// CreateCustomChain creates a custom agent chain
func (o *AgentOrchestrator) CreateCustomChain(chain *AgentChain) error {
	// Validate chain
	if chain.ID == "" {
		chain.ID = uuid.New().String()
	}
	
	if chain.Name == "" {
		return fmt.Errorf("chain name is required")
	}
	
	if len(chain.Agents) == 0 {
		return fmt.Errorf("chain must have at least one agent")
	}
	
	// Validate each agent
	for _, agent := range chain.Agents {
		if _, exists := o.registeredAgents[agent.Type]; !exists {
			return fmt.Errorf("unknown agent type: %s", agent.Type)
		}
	}
	
	// Save chain as AIAgent for tracking
	_, err := o.saveChainAsAIAgent(chain, chain.UserID)
	if err != nil {
		return fmt.Errorf("failed to save chain to database: %w", err)
	}
	
	fmt.Printf("Saved chain %s to database as AIAgent\n", chain.ID)
	return nil
}

// SaveExecutionResult saves an execution result to the database using AIAgent
func (o *AgentOrchestrator) SaveExecutionResult(result *ChainExecutionResult) error {
	// Update the AIAgent with completion status and results
	var aiAgent models.AIAgent
	err := o.db.Where("id = ?", result.ChainID).First(&aiAgent).Error
	if err != nil {
		return fmt.Errorf("failed to find AIAgent for chain: %w", err)
	}

	// Update status and output data
	aiAgent.Status = result.Status
	if result.EndTime != nil {
		aiAgent.CompletedAt = result.EndTime
	}
	
	// Store results in output_data
	if result.Results != nil {
		outputData := make(models.AIMetadata)
		for k, v := range result.Results {
			outputData[k] = v
		}
		aiAgent.OutputData = outputData
	}

	// Store errors if any
	if len(result.Errors) > 0 {
		if aiAgent.OutputData == nil {
			aiAgent.OutputData = make(models.AIMetadata)
		}
		aiAgent.OutputData["errors"] = result.Errors
	}

	// Save individual steps as AIAgentStep
	for i, log := range result.ExecutionLog {
		step := &models.AIAgentStep{
			AgentID:     aiAgent.ID,
			StepNumber:  i + 1,
			Name:        log.AgentName,
			Description: fmt.Sprintf("Agent: %s (%s)", log.AgentName, log.AgentID),
			Status:      log.Status,
			StartedAt:   &log.StartTime,
			CompletedAt: &log.EndTime,
		}

		// Store input/output data
		if log.Input != nil {
			inputData := make(models.AIMetadata)
			for k, v := range log.Input {
				inputData[k] = v
			}
			step.InputData = inputData
		}

		if log.Output != nil {
			outputData := make(models.AIMetadata)
			if outputMap, ok := log.Output.(map[string]interface{}); ok {
				for k, v := range outputMap {
					outputData[k] = v
				}
			} else {
				outputData["result"] = log.Output
			}
			step.OutputData = outputData
		}

		// Save the step
		if err := o.db.Create(step).Error; err != nil {
			fmt.Printf("Failed to save agent step %d: %v\n", i+1, err)
		}
	}

	// Update the agent
	err = o.db.Save(&aiAgent).Error
	if err != nil {
		return fmt.Errorf("failed to update AIAgent with results: %w", err)
	}

	fmt.Printf("Saved execution result for AIAgent %s\n", aiAgent.ID)
	return nil
}

// saveExecutionAsNotebook saves an agent chain execution as a notebook with notes for each step
func (o *AgentOrchestrator) saveExecutionAsNotebook(ctx context.Context, userID uuid.UUID, chain *AgentChain, result *ChainExecutionResult) (uuid.UUID, []uuid.UUID, error) {
	// Create notebook for this execution
	notebookTitle := fmt.Sprintf("Agent Chain: %s - %s", chain.Name, result.StartTime.Format("2006-01-02 15:04"))
	notebookData := map[string]interface{}{
		"name":        notebookTitle,
		"description": fmt.Sprintf("Results from %s execution (%s)", chain.Name, result.Status),
		"user_id":     userID.String(),
	}

	// Use database directly since we need to access the services
	dbWrapper := &database.Database{DB: o.db}
	notebook, err := o.notebookService.CreateNotebook(dbWrapper, notebookData)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("failed to create notebook: %w", err)
	}

	var noteIDs []uuid.UUID

	// Create overview note
	overviewNoteData := map[string]interface{}{
		"title":       "Execution Overview",
		"user_id":     userID.String(),
		"notebook_id": notebook.ID.String(),
	}

	overviewNote, err := o.noteService.CreateNote(dbWrapper, overviewNoteData)
	if err != nil {
		return notebook.ID, noteIDs, fmt.Errorf("failed to create overview note: %w", err)
	}
	noteIDs = append(noteIDs, overviewNote.ID)

	// Add overview content blocks
	blocks := []struct {
		blockType models.BlockType
		content   map[string]interface{}
		order     float64
	}{
		{
			blockType: models.HeadingBlock,
			content: map[string]interface{}{
				"text":  "Agent Chain Execution Results",
				"level": 1,
			},
			order: 1000.0,
		},
		{
			blockType: models.TextBlock,
			content: map[string]interface{}{
				"text": fmt.Sprintf("**Chain:** %s\n**Mode:** %s\n**Status:** %s\n**Duration:** %.2fs\n**Execution ID:** %s",
					chain.Name, chain.Mode, result.Status, 
					result.EndTime.Sub(result.StartTime).Seconds(), result.ID),
			},
			order: 2000.0,
		},
	}

	// Add error details if any
	if len(result.Errors) > 0 {
		var errorText strings.Builder
		errorText.WriteString("**Errors encountered:**\n")
		for _, err := range result.Errors {
			errorText.WriteString(fmt.Sprintf("- %s (%s): %s\n", err.AgentName, err.AgentID, err.Error))
		}
		
		blocks = append(blocks, struct {
			blockType models.BlockType
			content   map[string]interface{}
			order     float64
		}{
			blockType: models.TextBlock,
			content: map[string]interface{}{
				"text": errorText.String(),
			},
			order: 3000.0,
		})
	}

	// Create overview blocks
	for _, block := range blocks {
		blockModel := models.Block{
			ID:      uuid.New(),
			UserID:  userID,
			NoteID:  overviewNote.ID,
			Type:    block.blockType,
			Order:   block.order,
			Content: block.content,
			Metadata: models.BlockMetadata{},
		}
		o.db.Create(&blockModel)
	}

	// Create individual notes for each agent execution
	for i, log := range result.ExecutionLog {
		agentNoteData := map[string]interface{}{
			"title":       fmt.Sprintf("Step %d: %s", i+1, log.AgentName),
			"user_id":     userID.String(),
			"notebook_id": notebook.ID.String(),
		}

		agentNote, err := o.noteService.CreateNote(dbWrapper, agentNoteData)
		if err != nil {
			fmt.Printf("Failed to create note for agent %s: %v\n", log.AgentName, err)
			continue
		}
		noteIDs = append(noteIDs, agentNote.ID)

		// Create blocks for agent execution details
		agentBlocks := []struct {
			blockType models.BlockType
			content   map[string]interface{}
			order     float64
		}{
			{
				blockType: models.HeadingBlock,
				content: map[string]interface{}{
					"text":  fmt.Sprintf("Agent: %s", log.AgentName),
					"level": 1,
				},
				order: 1000.0,
			},
			{
				blockType: models.TextBlock,
				content: map[string]interface{}{
					"text": fmt.Sprintf("**Status:** %s\n**Duration:** %.2fs\n**Agent ID:** %s",
						log.Status, log.Duration, log.AgentID),
				},
				order: 2000.0,
			},
		}

		// Add input details if available
		if len(log.Input) > 0 {
			inputJSON, _ := json.MarshalIndent(log.Input, "", "  ")
			agentBlocks = append(agentBlocks, struct {
				blockType models.BlockType
				content   map[string]interface{}
				order     float64
			}{
				blockType: models.HeadingBlock,
				content: map[string]interface{}{
					"text":  "Input Parameters",
					"level": 2,
				},
				order: 3000.0,
			})
			agentBlocks = append(agentBlocks, struct {
				blockType models.BlockType
				content   map[string]interface{}
				order     float64
			}{
				blockType: models.TextBlock,
				content: map[string]interface{}{
					"text": fmt.Sprintf("```json\n%s\n```", string(inputJSON)),
				},
				order: 3100.0,
			})
		}

		// Add output details
		agentBlocks = append(agentBlocks, struct {
			blockType models.BlockType
			content   map[string]interface{}
			order     float64
		}{
			blockType: models.HeadingBlock,
			content: map[string]interface{}{
				"text":  "Output",
				"level": 2,
			},
			order: 4000.0,
		})

		// Format output based on its type
		var outputText string
		if log.Status == "failed" {
			outputText = fmt.Sprintf("**Error:** %v", log.Output)
		} else {
			if outputMap, ok := log.Output.(map[string]interface{}); ok {
				if outputJSON, err := json.MarshalIndent(outputMap, "", "  "); err == nil {
					outputText = fmt.Sprintf("```json\n%s\n```", string(outputJSON))
				} else {
					outputText = fmt.Sprintf("%v", log.Output)
				}
			} else {
				outputText = fmt.Sprintf("%v", log.Output)
			}
		}

		agentBlocks = append(agentBlocks, struct {
			blockType models.BlockType
			content   map[string]interface{}
			order     float64
		}{
			blockType: models.TextBlock,
			content: map[string]interface{}{
				"text": outputText,
			},
			order: 4100.0,
		})

		// Create all blocks for this agent
		for _, block := range agentBlocks {
			blockModel := models.Block{
				ID:      uuid.New(),
				UserID:  userID,
				NoteID:  agentNote.ID,
				Type:    block.blockType,
				Order:   block.order,
				Content: block.content,
				Metadata: models.BlockMetadata{},
			}
			o.db.Create(&blockModel)
		}
	}

	// Create final results note if there are results
	if len(result.Results) > 0 {
		resultsNoteData := map[string]interface{}{
			"title":       "Final Results",
			"user_id":     userID.String(),
			"notebook_id": notebook.ID.String(),
		}

		resultsNote, err := o.noteService.CreateNote(dbWrapper, resultsNoteData)
		if err == nil {
			noteIDs = append(noteIDs, resultsNote.ID)

			// Add results content
			resultsJSON, _ := json.MarshalIndent(result.Results, "", "  ")
			
			resultBlocks := []struct {
				blockType models.BlockType
				content   map[string]interface{}
				order     float64
			}{
				{
					blockType: models.HeadingBlock,
					content: map[string]interface{}{
						"text":  "Chain Results",
						"level": 1,
					},
					order: 1000.0,
				},
				{
					blockType: models.TextBlock,
					content: map[string]interface{}{
						"text": fmt.Sprintf("```json\n%s\n```", string(resultsJSON)),
					},
					order: 2000.0,
				},
			}

			for _, block := range resultBlocks {
				blockModel := models.Block{
					ID:      uuid.New(),
					UserID:  userID,
					NoteID:  resultsNote.ID,
					Type:    block.blockType,
					Order:   block.order,
					Content: block.content,
					Metadata: models.BlockMetadata{},
				}
				o.db.Create(&blockModel)
			}
		}
	}

	return notebook.ID, noteIDs, nil
}
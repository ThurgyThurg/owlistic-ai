package services

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatService implements a context-aware chat interface with RAG capabilities
type ChatService struct {
	db          *gorm.DB
	ai          *AIService
	noteService *NoteService
}

// ChatRequest represents a chat request from the user
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
	Context   string `json:"context"` // optional context override
}

// ChatResponse represents the AI's response
type ChatResponse struct {
	Message    string                 `json:"message"`
	Sources    []ChatSource           `json:"sources"`
	Metadata   map[string]interface{} `json:"metadata"`
	SessionID  string                 `json:"session_id"`
}

// ChatSource represents a source used to generate the response
type ChatSource struct {
	Type      string `json:"type"` // note, task, web
	ID        string `json:"id"`
	Title     string `json:"title"`
	Excerpt   string `json:"excerpt"`
	Relevance float64 `json:"relevance"`
}

func NewChatService(db *gorm.DB, ai *AIService, noteService *NoteService) *ChatService {
	return &ChatService{
		db:          db,
		ai:          ai,
		noteService: noteService,
	}
}

// Chat handles a chat request with RAG (Retrieval Augmented Generation)
func (c *ChatService) Chat(ctx context.Context, userID uuid.UUID, req ChatRequest) (*ChatResponse, error) {
	// Ensure we have a session ID
	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	// Store the user message in chat memory
	if err := c.storeChatMessage(ctx, userID, req.SessionID, "user", req.Message); err != nil {
		log.Printf("Failed to store user message: %v", err)
	}

	// Get conversation history
	history, err := c.getChatHistory(ctx, userID, req.SessionID, 10)
	if err != nil {
		log.Printf("Failed to get chat history: %v", err)
	}

	// Analyze the message to determine intent and extract key topics
	intent, topics := c.analyzeMessage(ctx, req.Message)
	
	// Retrieve relevant context based on intent
	sources := []ChatSource{}
	contextText := ""
	
	switch intent {
	case "search", "question":
		// Search for relevant notes and information
		noteSources, noteContext := c.searchNotes(ctx, userID, req.Message, topics)
		sources = append(sources, noteSources...)
		contextText += noteContext
		
		// If web search is needed and available
		if c.needsCurrentInfo(req.Message) && c.ai.perplexicaService != nil && c.ai.perplexicaService.IsEnabled() {
			webSources, webContext := c.searchWeb(ctx, userID, req.Message)
			sources = append(sources, webSources...)
			contextText += "\n\n" + webContext
		}
		
	case "task", "planning":
		// Search for relevant tasks and projects
		taskSources, taskContext := c.searchTasks(ctx, userID, topics)
		sources = append(sources, taskSources...)
		contextText += taskContext
		
	case "create", "generate":
		// Gather context for creation
		noteSources, noteContext := c.searchNotes(ctx, userID, req.Message, topics)
		sources = append(sources, noteSources...)
		contextText += noteContext
		
	default:
		// General conversation - include recent notes and general context
		recentSources, recentContext := c.getRecentContext(ctx, userID)
		sources = append(sources, recentSources...)
		contextText += recentContext
	}

	// Add custom context if provided
	if req.Context != "" {
		contextText = req.Context + "\n\n" + contextText
	}

	// Generate response using AI with retrieved context
	response, err := c.generateResponse(ctx, req.Message, contextText, history, intent)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Store the assistant's response
	if err := c.storeChatMessage(ctx, userID, req.SessionID, "assistant", response); err != nil {
		log.Printf("Failed to store assistant message: %v", err)
	}

	return &ChatResponse{
		Message:   response,
		Sources:   sources,
		SessionID: req.SessionID,
		Metadata: map[string]interface{}{
			"intent":       intent,
			"topics":       topics,
			"source_count": len(sources),
		},
	}, nil
}

// analyzeMessage analyzes the user's message to determine intent and extract topics
func (c *ChatService) analyzeMessage(ctx context.Context, message string) (string, []string) {
	messageLower := strings.ToLower(message)
	
	// Determine intent based on keywords
	intent := "general"
	if strings.Contains(messageLower, "?") || 
	   strings.Contains(messageLower, "what") || 
	   strings.Contains(messageLower, "how") || 
	   strings.Contains(messageLower, "why") || 
	   strings.Contains(messageLower, "when") || 
	   strings.Contains(messageLower, "where") {
		intent = "question"
	} else if strings.Contains(messageLower, "search") || 
	          strings.Contains(messageLower, "find") || 
	          strings.Contains(messageLower, "look for") {
		intent = "search"
	} else if strings.Contains(messageLower, "create") || 
	          strings.Contains(messageLower, "generate") || 
	          strings.Contains(messageLower, "write") || 
	          strings.Contains(messageLower, "make") {
		intent = "create"
	} else if strings.Contains(messageLower, "task") || 
	          strings.Contains(messageLower, "todo") || 
	          strings.Contains(messageLower, "plan") || 
	          strings.Contains(messageLower, "schedule") {
		intent = "task"
	}
	
	// Extract topics (simple keyword extraction)
	topics := c.extractTopics(message)
	
	return intent, topics
}

// extractTopics extracts key topics from the message
func (c *ChatService) extractTopics(message string) []string {
	// Remove common words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "is": true, "are": true, "was": true,
		"were": true, "been": true, "be": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
		"should": true, "may": true, "might": true, "must": true, "can": true, "what": true,
		"how": true, "why": true, "when": true, "where": true, "who": true, "which": true,
	}
	
	words := strings.Fields(strings.ToLower(message))
	topics := []string{}
	
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:'\"")
		if len(word) > 3 && !stopWords[word] {
			topics = append(topics, word)
		}
	}
	
	return topics
}

// searchNotes searches user's notes for relevant information
func (c *ChatService) searchNotes(ctx context.Context, userID uuid.UUID, query string, topics []string) ([]ChatSource, string) {
	sources := []ChatSource{}
	contextParts := []string{}
	
	// Search by query
	var notes []models.Note
	searchQuery := c.db.Where("user_id = ?", userID).
		Where("deleted_at IS NULL")
	
	// Add topic-based search
	if len(topics) > 0 {
		for _, topic := range topics {
			searchQuery = searchQuery.Where("title ILIKE ? OR tags::text ILIKE ?", 
				"%"+topic+"%", "%"+topic+"%")
		}
	}
	
	searchQuery.Limit(5).Find(&notes)
	
	// Process found notes
	for _, note := range notes {
		// Get note content
		content := c.extractNoteContent(&note)
		excerpt := content
		if len(excerpt) > 200 {
			excerpt = excerpt[:200] + "..."
		}
		
		source := ChatSource{
			Type:      "note",
			ID:        note.ID.String(),
			Title:     note.Title,
			Excerpt:   excerpt,
			Relevance: c.calculateRelevance(query, note.Title + " " + content),
		}
		sources = append(sources, source)
		
		contextParts = append(contextParts, fmt.Sprintf("Note '%s': %s", note.Title, content))
	}
	
	// Sort by relevance
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Relevance > sources[j].Relevance
	})
	
	context := "Relevant notes from user's knowledge base:\n" + strings.Join(contextParts, "\n\n")
	return sources, context
}

// searchTasks searches for relevant tasks
func (c *ChatService) searchTasks(ctx context.Context, userID uuid.UUID, topics []string) ([]ChatSource, string) {
	sources := []ChatSource{}
	contextParts := []string{}
	
	var tasks []models.Task
	query := c.db.Where("user_id = ?", userID).
		Where("deleted_at IS NULL")
	
	// Add topic-based search
	if len(topics) > 0 {
		for _, topic := range topics {
			query = query.Where("title ILIKE ? OR description ILIKE ?", 
				"%"+topic+"%", "%"+topic+"%")
		}
	}
	
	query.Limit(5).Find(&tasks)
	
	for _, task := range tasks {
		status := "pending"
		if task.IsCompleted {
			status = "completed"
		}
		
		source := ChatSource{
			Type:    "task",
			ID:      task.ID.String(),
			Title:   task.Title,
			Excerpt: fmt.Sprintf("[%s] %s", status, task.Description),
		}
		sources = append(sources, source)
		
		contextParts = append(contextParts, 
			fmt.Sprintf("Task '%s' (%s): %s", task.Title, status, task.Description))
	}
	
	context := "Relevant tasks:\n" + strings.Join(contextParts, "\n")
	return sources, context
}

// searchWeb searches the web for current information
func (c *ChatService) searchWeb(ctx context.Context, userID uuid.UUID, query string) ([]ChatSource, string) {
	sources := []ChatSource{}
	context := ""
	
	// Use Perplexica for web search
	result, err := c.ai.SearchWithPerplexica(ctx, userID, query, "webSearch", nil)
	if err != nil {
		log.Printf("Web search failed: %v", err)
		return sources, context
	}
	
	if result != nil && result.Success {
		// Convert Perplexica sources to chat sources
		for i, source := range result.Sources {
			if i >= 3 { // Limit to top 3 sources
				break
			}
			
			chatSource := ChatSource{
				Type:    "web",
				ID:      fmt.Sprintf("web_%d", i),
				Title:   fmt.Sprintf("Web source %d", i+1),
				Excerpt: source.PageContent[:min(200, len(source.PageContent))] + "...",
			}
			sources = append(sources, chatSource)
		}
		
		context = fmt.Sprintf("Current information from web search:\n%s", result.Answer)
	}
	
	return sources, context
}

// getRecentContext gets recent notes and activities for context
func (c *ChatService) getRecentContext(ctx context.Context, userID uuid.UUID) ([]ChatSource, string) {
	sources := []ChatSource{}
	contextParts := []string{}
	
	// Get recent notes
	var recentNotes []models.Note
	c.db.Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Order("updated_at DESC").
		Limit(3).
		Find(&recentNotes)
	
	for _, note := range recentNotes {
		source := ChatSource{
			Type:  "note",
			ID:    note.ID.String(),
			Title: note.Title,
			Excerpt: fmt.Sprintf("Recently updated: %s", 
				note.UpdatedAt.Format("Jan 2, 2006")),
		}
		sources = append(sources, source)
		
		contextParts = append(contextParts, 
			fmt.Sprintf("Recent note '%s' (updated %s)", 
				note.Title, note.UpdatedAt.Format("Jan 2, 2006")))
	}
	
	context := "Recent activity:\n" + strings.Join(contextParts, "\n")
	return sources, context
}

// generateResponse generates an AI response with context
func (c *ChatService) generateResponse(ctx context.Context, message, context string, history []models.ChatMemory, intent string) (string, error) {
	// Build conversation history
	historyText := ""
	for _, h := range history {
		if h.Role != "system" {
			historyText += fmt.Sprintf("%s: %s\n", strings.Title(h.Role), h.Content)
		}
	}
	
	// Build the prompt
	systemPrompt := `You are an intelligent assistant integrated with the user's personal knowledge management system. 
You have access to their notes, tasks, and can search for current information when needed.

Your responses should be:
1. Helpful and directly address the user's question or request
2. Based on the provided context when available
3. Clear about when you're using information from their notes vs general knowledge
4. Proactive in suggesting related information or next steps when appropriate`

	prompt := fmt.Sprintf(`%s

User Intent: %s

Available Context:
%s

Recent Conversation:
%s

Current Message: %s

Please provide a helpful response that addresses the user's message, incorporating relevant context when appropriate.`,
		systemPrompt,
		intent,
		context,
		historyText,
		message)

	// Generate response
	response, err := c.ai.callAnthropic(ctx, prompt, 1000)
	if err != nil {
		return "", err
	}
	
	return response, nil
}

// storeChatMessage stores a message in chat memory
func (c *ChatService) storeChatMessage(ctx context.Context, userID uuid.UUID, sessionID, role, content string) error {
	memory := models.ChatMemory{
		UserID:    userID,
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Metadata:  models.AIMetadata{},
	}
	
	return c.db.WithContext(ctx).Create(&memory).Error
}

// getChatHistory retrieves recent chat history for a session
func (c *ChatService) getChatHistory(ctx context.Context, userID uuid.UUID, sessionID string, limit int) ([]models.ChatMemory, error) {
	var history []models.ChatMemory
	err := c.db.WithContext(ctx).
		Where("user_id = ? AND session_id = ?", userID, sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	
	// Reverse to get chronological order
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}
	
	return history, err
}

// Helper methods

func (c *ChatService) extractNoteContent(note *models.Note) string {
	// Get blocks for the note
	var blocks []models.Block
	c.db.Where("note_id = ?", note.ID).Order("\"order\"").Find(&blocks)
	
	content := note.Title + "\n"
	for _, block := range blocks {
		if text, ok := block.Content["text"].(string); ok {
			content += text + "\n"
		}
	}
	
	return content
}

func (c *ChatService) calculateRelevance(query, content string) float64 {
	// Simple relevance calculation based on keyword matching
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)
	
	words := strings.Fields(queryLower)
	matches := 0
	
	for _, word := range words {
		if strings.Contains(contentLower, word) {
			matches++
		}
	}
	
	if len(words) == 0 {
		return 0
	}
	
	return float64(matches) / float64(len(words))
}

func (c *ChatService) needsCurrentInfo(message string) bool {
	// Check if the message asks about current events or recent information
	currentIndicators := []string{
		"today", "current", "latest", "recent", "now", "news",
		"2024", "2025", "this year", "this month", "this week",
	}
	
	messageLower := strings.ToLower(message)
	for _, indicator := range currentIndicators {
		if strings.Contains(messageLower, indicator) {
			return true
		}
	}
	
	return false
}

// GetChatSessions retrieves all chat sessions for a user
func (c *ChatService) GetChatSessions(ctx context.Context, userID uuid.UUID) ([]map[string]interface{}, error) {
	var sessions []map[string]interface{}
	
	// Get unique sessions with latest message
	rows, err := c.db.WithContext(ctx).
		Model(&models.ChatMemory{}).
		Select("session_id, MAX(created_at) as last_message_at, COUNT(*) as message_count").
		Where("user_id = ?", userID).
		Group("session_id").
		Order("last_message_at DESC").
		Rows()
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var sessionID string
		var lastMessageAt time.Time
		var messageCount int
		
		if err := rows.Scan(&sessionID, &lastMessageAt, &messageCount); err != nil {
			continue
		}
		
		// Get the first user message as preview
		var firstMessage models.ChatMemory
		c.db.Where("user_id = ? AND session_id = ? AND role = ?", userID, sessionID, "user").
			Order("created_at").
			First(&firstMessage)
		
		sessions = append(sessions, map[string]interface{}{
			"session_id":    sessionID,
			"preview":       firstMessage.Content,
			"last_message":  lastMessageAt,
			"message_count": messageCount,
		})
	}
	
	return sessions, nil
}

// DeleteChatSession deletes all messages in a chat session
func (c *ChatService) DeleteChatSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	return c.db.WithContext(ctx).
		Where("user_id = ? AND session_id = ?", userID, sessionID).
		Delete(&models.ChatMemory{}).Error
}
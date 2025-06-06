package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"owlistic-notes/owlistic/utils/logger"
)

type PerplexicaService struct {
	baseURL    string
	httpClient *http.Client
	logger     *logger.Logger
}

// PerplexicaRequest represents the request structure for Perplexica API
type PerplexicaRequest struct {
	ChatModel        *ChatModel        `json:"chatModel,omitempty"`
	EmbeddingModel   *EmbeddingModel   `json:"embeddingModel,omitempty"`
	OptimizationMode string            `json:"optimizationMode,omitempty"`
	FocusMode        string            `json:"focusMode"`
	Query            string            `json:"query"`
	History          [][]string        `json:"history,omitempty"`
	SystemInstructions string          `json:"systemInstructions,omitempty"`
	Stream           bool              `json:"stream"`
}

type ChatModel struct {
	Provider           string `json:"provider"`
	Name               string `json:"name"`
	CustomOpenAIBaseURL string `json:"customOpenAIBaseURL,omitempty"`
	CustomOpenAIKey    string `json:"customOpenAIKey,omitempty"`
}

type EmbeddingModel struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
}

// PerplexicaResponse represents the response from Perplexica API
type PerplexicaResponse struct {
	Message string              `json:"message"`
	Sources []PerplexicaSource  `json:"sources"`
}

type PerplexicaSource struct {
	PageContent string                 `json:"pageContent"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PerplexicaSearchResult represents the structured result for internal use
type PerplexicaSearchResult struct {
	Query       string              `json:"query"`
	Answer      string              `json:"answer"`
	Sources     []PerplexicaSource  `json:"sources"`
	FocusMode   string              `json:"focus_mode"`
	Timestamp   time.Time           `json:"timestamp"`
	Success     bool                `json:"success"`
	Error       string              `json:"error,omitempty"`
}

// NewPerplexicaService creates a new Perplexica service instance
func NewPerplexicaService() *PerplexicaService {
	baseURL := os.Getenv("PERPLEXICA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000" // Default Perplexica URL
	}

	return &PerplexicaService{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 25 * time.Second}, // Based on performance testing: avg 7.7s, max 11s observed
		logger:     logger.New("PerplexicaService"),
	}
}

// IsEnabled checks if Perplexica service is configured (without health check to prevent recursion)
func (p *PerplexicaService) IsEnabled() bool {
	// Check if PERPLEXICA_BASE_URL is configured
	if os.Getenv("PERPLEXICA_BASE_URL") == "" {
		return false
	}

	// Just check if URL is configured - don't do health check here to prevent infinite recursion
	return p.baseURL != ""
}

// Search performs a search using Perplexica with the specified focus mode
func (p *PerplexicaService) Search(ctx context.Context, query string, focusMode string, optimizationMode string) (*PerplexicaSearchResult, error) {
	if !p.IsEnabled() {
		return nil, fmt.Errorf("perplexica service is not enabled or configured")
	}

	// Validate focus mode
	validFocusModes := []string{"webSearch", "academicSearch", "writingAssistant", "wolframAlphaSearch", "youtubeSearch", "redditSearch"}
	if !contains(validFocusModes, focusMode) {
		focusMode = "webSearch" // Default to web search
	}

	// Set default optimization mode
	if optimizationMode == "" {
		optimizationMode = "balanced"
	}

	// Prepare request
	request := PerplexicaRequest{
		FocusMode:        focusMode,
		OptimizationMode: optimizationMode,
		Query:            query,
		Stream:           false, // We don't support streaming for now
	}

	// Add default models if configured
	if chatProvider := os.Getenv("PERPLEXICA_CHAT_PROVIDER"); chatProvider != "" {
		if chatModel := os.Getenv("PERPLEXICA_CHAT_MODEL"); chatModel != "" {
			request.ChatModel = &ChatModel{
				Provider: chatProvider,
				Name:     chatModel,
			}
		}
	}

	if embeddingProvider := os.Getenv("PERPLEXICA_EMBEDDING_PROVIDER"); embeddingProvider != "" {
		if embeddingModel := os.Getenv("PERPLEXICA_EMBEDDING_MODEL"); embeddingModel != "" {
			request.EmbeddingModel = &EmbeddingModel{
				Provider: embeddingProvider,
				Name:     embeddingModel,
			}
		}
	}

	p.logger.Info("Performing Perplexica search", map[string]interface{}{
		"query":      query,
		"focus_mode": focusMode,
		"optimization_mode": optimizationMode,
	})

	// Make the API call
	response, err := p.makeRequest(ctx, request)
	if err != nil {
		p.logger.Error("Perplexica search failed", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
		return &PerplexicaSearchResult{
			Query:     query,
			FocusMode: focusMode,
			Timestamp: time.Now(),
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	result := &PerplexicaSearchResult{
		Query:     query,
		Answer:    response.Message,
		Sources:   response.Sources,
		FocusMode: focusMode,
		Timestamp: time.Now(),
		Success:   true,
	}

	p.logger.Info("Perplexica search completed", map[string]interface{}{
		"query":        query,
		"sources_count": len(response.Sources),
		"answer_length": len(response.Message),
	})

	return result, nil
}

// WebSearch performs a web search using Perplexica
func (p *PerplexicaService) WebSearch(ctx context.Context, query string) (*PerplexicaSearchResult, error) {
	return p.Search(ctx, query, "webSearch", "balanced")
}

// AcademicSearch performs an academic search using Perplexica
func (p *PerplexicaService) AcademicSearch(ctx context.Context, query string) (*PerplexicaSearchResult, error) {
	return p.Search(ctx, query, "academicSearch", "balanced")
}

// RedditSearch performs a Reddit search using Perplexica
func (p *PerplexicaService) RedditSearch(ctx context.Context, query string) (*PerplexicaSearchResult, error) {
	return p.Search(ctx, query, "redditSearch", "speed")
}

// YouTubeSearch performs a YouTube search using Perplexica
func (p *PerplexicaService) YouTubeSearch(ctx context.Context, query string) (*PerplexicaSearchResult, error) {
	return p.Search(ctx, query, "youtubeSearch", "speed")
}

// SearchWithContext performs a contextual search with conversation history
func (p *PerplexicaService) SearchWithContext(ctx context.Context, query string, focusMode string, history [][]string, systemInstructions string) (*PerplexicaSearchResult, error) {
	if !p.IsEnabled() {
		return nil, fmt.Errorf("perplexica service is not enabled or configured")
	}

	// Validate focus mode
	validFocusModes := []string{"webSearch", "academicSearch", "writingAssistant", "wolframAlphaSearch", "youtubeSearch", "redditSearch"}
	if !contains(validFocusModes, focusMode) {
		focusMode = "webSearch" // Default to web search
	}

	// Prepare request with context
	request := PerplexicaRequest{
		FocusMode:          focusMode,
		OptimizationMode:   "balanced",
		Query:              query,
		History:            history,
		SystemInstructions: systemInstructions,
		Stream:             false,
	}

	// Add default models if configured
	if chatProvider := os.Getenv("PERPLEXICA_CHAT_PROVIDER"); chatProvider != "" {
		if chatModel := os.Getenv("PERPLEXICA_CHAT_MODEL"); chatModel != "" {
			request.ChatModel = &ChatModel{
				Provider: chatProvider,
				Name:     chatModel,
			}
		}
	}

	if embeddingProvider := os.Getenv("PERPLEXICA_EMBEDDING_PROVIDER"); embeddingProvider != "" {
		if embeddingModel := os.Getenv("PERPLEXICA_EMBEDDING_MODEL"); embeddingModel != "" {
			request.EmbeddingModel = &EmbeddingModel{
				Provider: embeddingProvider,
				Name:     embeddingModel,
			}
		}
	}

	p.logger.Info("Performing contextual Perplexica search", map[string]interface{}{
		"query":      query,
		"focus_mode": focusMode,
		"has_history": len(history) > 0,
		"has_instructions": systemInstructions != "",
	})

	// Make the API call
	response, err := p.makeRequest(ctx, request)
	if err != nil {
		p.logger.Error("Contextual Perplexica search failed", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
		return &PerplexicaSearchResult{
			Query:     query,
			FocusMode: focusMode,
			Timestamp: time.Now(),
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	result := &PerplexicaSearchResult{
		Query:     query,
		Answer:    response.Message,
		Sources:   response.Sources,
		FocusMode: focusMode,
		Timestamp: time.Now(),
		Success:   true,
	}

	p.logger.Info("Contextual Perplexica search completed", map[string]interface{}{
		"query":        query,
		"sources_count": len(response.Sources),
		"answer_length": len(response.Message),
	})

	return result, nil
}

// makeRequest performs the actual HTTP request to Perplexica API
func (p *PerplexicaService) makeRequest(ctx context.Context, request PerplexicaRequest) (*PerplexicaResponse, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/search", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("perplexica API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response PerplexicaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAvailableModels retrieves available models from Perplexica (if the API supports it)
func (p *PerplexicaService) GetAvailableModels(ctx context.Context) (map[string]interface{}, error) {
	if !p.IsEnabled() {
		return nil, fmt.Errorf("perplexica service is not enabled or configured")
	}

	url := fmt.Sprintf("%s/api/models", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("perplexica models API returned status %d: %s", resp.StatusCode, string(body))
	}

	var models map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("failed to parse models response: %w", err)
	}

	return models, nil
}

// HealthCheck checks if Perplexica service is healthy and reachable
func (p *PerplexicaService) HealthCheck(ctx context.Context) error {
	if !p.IsEnabled() {
		return fmt.Errorf("perplexica service is not enabled or configured")
	}

	// Try to get models as a health check
	_, err := p.GetAvailableModels(ctx)
	if err != nil {
		return fmt.Errorf("perplexica health check failed: %w", err)
	}

	return nil
}

// FormatSearchResultForAI formats the search result in a way that's useful for AI agents
func (p *PerplexicaService) FormatSearchResultForAI(result *PerplexicaSearchResult) string {
	if result == nil || !result.Success {
		return "Search failed or returned no results."
	}

	var formatted strings.Builder
	
	formatted.WriteString(fmt.Sprintf("Search Query: %s\n", result.Query))
	formatted.WriteString(fmt.Sprintf("Search Type: %s\n\n", result.FocusMode))
	
	formatted.WriteString("Answer:\n")
	formatted.WriteString(result.Answer)
	formatted.WriteString("\n\n")
	
	if len(result.Sources) > 0 {
		formatted.WriteString("Sources:\n")
		for i, source := range result.Sources {
			formatted.WriteString(fmt.Sprintf("%d. ", i+1))
			if title, ok := source.Metadata["title"].(string); ok {
				formatted.WriteString(fmt.Sprintf("%s\n", title))
			}
			if url, ok := source.Metadata["url"].(string); ok {
				formatted.WriteString(fmt.Sprintf("   URL: %s\n", url))
			}
			if source.PageContent != "" {
				// Truncate content if too long
				content := source.PageContent
				if len(content) > 200 {
					content = content[:200] + "..."
				}
				formatted.WriteString(fmt.Sprintf("   Content: %s\n", content))
			}
			formatted.WriteString("\n")
		}
	}
	
	return formatted.String()
}

// Helper function to check if a string slice contains a value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
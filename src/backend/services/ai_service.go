package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"owlistic-notes/owlistic/database"
	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Constants for ChromaDB collection
const (
	NoteEmbeddingsCollection = "note_embeddings"
	MaxDocumentLength       = 8000 // ChromaDB's default max length
)

type AIService struct {
	db                *gorm.DB
	anthropicKey      string
	anthropicModel    string
	chromaService     *ChromaService
	httpClient        *http.Client
	perplexicaService *PerplexicaService
}

type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func NewAIService(db *gorm.DB) *AIService {
	// Set default models if not specified
	anthropicModel := os.Getenv("ANTHROPIC_MODEL")
	if anthropicModel == "" {
		anthropicModel = "claude-3-5-sonnet-20241022" // Default Anthropic model
	}

	// Clean the API key of any whitespace
	anthropicKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	
	// Initialize ChromaDB service
	chromaBaseURL := os.Getenv("CHROMA_BASE_URL")
	chromaService := NewChromaService(chromaBaseURL, db)
	
	service := &AIService{
		db:                db,
		anthropicKey:      anthropicKey,
		anthropicModel:    anthropicModel,
		chromaService:     chromaService,
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		perplexicaService: NewPerplexicaService(),
	}
	
	// Initialize ChromaDB collection
	ctx := context.Background()
	if err := service.initializeChromaCollection(ctx); err != nil {
		log.Printf("Warning: Failed to initialize ChromaDB collection: %v", err)
		// Continue anyway - we'll handle errors gracefully
	}
	
	return service
}

// initializeChromaCollection ensures the note embeddings collection exists
func (ai *AIService) initializeChromaCollection(ctx context.Context) error {
	// Configuration for optimal note search
	config := &ChromaConfiguration{
		HNSW: &HNSWConfig{
			Space:          "cosine", // Cosine similarity works well for text
			EFConstruction: 200,      // Higher for better quality
			EFSearch:       100,      // Good balance of speed and accuracy
			MaxNeighbors:   32,       // More connections for better recall
		},
	}
	
	return ai.chromaService.GetOrCreateCollection(ctx, NoteEmbeddingsCollection, config)
}

// ProcessNoteWithAI enhances a note with AI-generated metadata
func (ai *AIService) ProcessNoteWithAI(ctx context.Context, noteID uuid.UUID) error {
	// Get the note
	var note models.Note
	if err := ai.db.WithContext(ctx).First(&note, noteID).Error; err != nil {
		return fmt.Errorf("failed to find note: %w", err)
	}

	// Get note content (combine title and blocks content)
	content := ai.extractNoteContent(&note)

	// Generate AI enhancements concurrently
	titleChan := make(chan string, 1)
	summaryChan := make(chan string, 1)
	tagsChan := make(chan []string, 1)
	actionStepsChan := make(chan []string, 1)
	learningItemsChan := make(chan []string, 1)
	errChan := make(chan error, 5)

	// Generate title if empty
	go func() {
		if note.Title == "" {
			title, err := ai.generateTitle(ctx, content)
			if err != nil {
				errChan <- err
				return
			}
			titleChan <- title
		} else {
			titleChan <- note.Title
		}
	}()

	// Generate summary
	go func() {
		summary, err := ai.generateSummary(ctx, content, note.Title)
		if err != nil {
			errChan <- err
			return
		}
		summaryChan <- summary
	}()

	// Extract tags
	go func() {
		tags, err := ai.extractTags(ctx, content, note.Title)
		if err != nil {
			errChan <- err
			return
		}
		tagsChan <- tags
	}()

	// Generate actionable steps
	go func() {
		steps, err := ai.extractActionableSteps(ctx, content, note.Title)
		if err != nil {
			errChan <- err
			return
		}
		actionStepsChan <- steps
	}()

	// Generate learning items
	go func() {
		items, err := ai.extractLearningItems(ctx, content, note.Title)
		if err != nil {
			errChan <- err
			return
		}
		learningItemsChan <- items
	}()

	// Collect results
	var finalTitle string
	var summary string
	var aiTags []string
	var actionSteps []string
	var learningItems []string
	
	completed := 0
	errors := 0

	for completed < 5 {
		select {
		case title := <-titleChan:
			finalTitle = title
			completed++
		case s := <-summaryChan:
			summary = s
			completed++
		case tags := <-tagsChan:
			aiTags = tags
			completed++
		case steps := <-actionStepsChan:
			actionSteps = steps
			completed++
		case items := <-learningItemsChan:
			learningItems = items
			completed++
		case err := <-errChan:
			log.Printf("AI processing error: %v", err)
			errors++
			completed++ // Count error as completed to avoid infinite loop
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Save AI enhancements to database
	enhancedNote := models.AIEnhancedNote{
		NoteID:           note.ID,
		Summary:          summary,
		AITags:           pq.StringArray(aiTags),
		ActionSteps:      pq.StringArray(actionSteps),
		LearningItems:    pq.StringArray(learningItems),
		ProcessingStatus: "completed",
		LastProcessedAt:  &[]time.Time{time.Now()}[0],
		AIMetadata: models.AIMetadata{
			"processing_errors": errors,
			"ai_model":         ai.anthropicModel,
		},
	}

	// Update the note title if it was empty
	if note.Title == "" && finalTitle != "" {
		note.Title = finalTitle
		if err := ai.db.Save(&note).Error; err != nil {
			log.Printf("Failed to update note title: %v", err)
		}
	}

	// Save enhanced note data (use Clauses for upsert behavior)
	if err := ai.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "note_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"summary", "ai_tags", "action_steps", "learning_items", 
			"processing_status", "last_processed_at", "ai_metadata", "updated_at",
		}),
	}).Create(&enhancedNote).Error; err != nil {
		log.Printf("Failed to save AI enhancements: %v", err)
		return fmt.Errorf("failed to save AI enhancements: %w", err)
	}

	// Add to ChromaDB for vector search
	if err := ai.AddNoteToChroma(ctx, &note, &enhancedNote); err != nil {
		log.Printf("Failed to add note to ChromaDB: %v", err)
		// Don't fail the whole operation if ChromaDB fails
	}

	// Find and store related notes
	go func() {
		if relatedNotes, err := ai.FindRelatedNotes(ctx, noteID, 5); err == nil && len(relatedNotes) > 0 {
			relatedIDs := make([]uuid.UUID, 0, len(relatedNotes))
			for _, rn := range relatedNotes {
				if rn.ID != noteID { // Don't include self
					relatedIDs = append(relatedIDs, rn.ID)
				}
			}
			
			if len(relatedIDs) > 0 {
				ai.db.Model(&enhancedNote).Update("related_note_ids", relatedIDs)
			}
		}
	}()

	return nil
}

// addNoteToChroma adds or updates a note in the ChromaDB collection
func (ai *AIService) AddNoteToChroma(ctx context.Context, note *models.Note, enhanced *models.AIEnhancedNote) error {
	// Prepare document text
	var docBuilder strings.Builder
	docBuilder.WriteString(note.Title)
	docBuilder.WriteString("\n\n")
	
	// Add summary if available
	if enhanced != nil && enhanced.Summary != "" {
		docBuilder.WriteString("Summary: ")
		docBuilder.WriteString(enhanced.Summary)
		docBuilder.WriteString("\n\n")
	}
	
	// Add content
	content := ai.extractNoteContent(note)
	docBuilder.WriteString(content)
	
	// Truncate if too long
	document := docBuilder.String()
	if len(document) > MaxDocumentLength {
		document = document[:MaxDocumentLength]
	}
	
	// Prepare metadata
	metadata := map[string]interface{}{
		"note_id":    note.ID.String(),
		"title":      note.Title,
		"created_at": note.CreatedAt.Format(time.RFC3339),
		"updated_at": note.UpdatedAt.Format(time.RFC3339),
		"user_id":    note.UserID.String(),
	}
	
	if note.NotebookID != uuid.Nil {
		metadata["notebook_id"] = note.NotebookID.String()
	}
	
	if enhanced != nil {
		if len(enhanced.AITags) > 0 {
			metadata["ai_tags"] = strings.Join(enhanced.AITags, ",")
		}
		if enhanced.ProcessingStatus != "" {
			metadata["processing_status"] = enhanced.ProcessingStatus
		}
	}
	
	// Upsert to ChromaDB
	ids := []string{NoteIDToChromaID(note.ID)}
	documents := []string{document}
	metadatas := []map[string]interface{}{metadata}
	
	return ai.chromaService.UpsertDocuments(ctx, NoteEmbeddingsCollection, ids, documents, metadatas)
}

// findRelatedNotes finds notes similar to the given note using vector search
func (ai *AIService) FindRelatedNotes(ctx context.Context, noteID uuid.UUID, limit int) ([]models.Note, error) {
	// Query ChromaDB for similar notes
	queryTexts := []string{}
	
	// Get the note content to use as query
	var note models.Note
	if err := ai.db.First(&note, noteID).Error; err != nil {
		return nil, err
	}
	
	// Use title and first part of content as query
	queryText := note.Title
	content := ai.extractNoteContent(&note)
	if len(content) > 500 {
		queryText += " " + content[:500]
	} else {
		queryText += " " + content
	}
	queryTexts = append(queryTexts, queryText)
	
	// Exclude the current note from results
	where := map[string]interface{}{
		"note_id": map[string]interface{}{
			"$ne": noteID.String(),
		},
	}
	
	// Query ChromaDB
	results, err := ai.chromaService.QueryByText(ctx, NoteEmbeddingsCollection, queryTexts, limit+1, where)
	if err != nil {
		return nil, fmt.Errorf("failed to query ChromaDB: %w", err)
	}
	
	// Convert results to notes
	var relatedNotes []models.Note
	if len(results.IDs) > 0 && len(results.IDs[0]) > 0 {
		for i, chromaID := range results.IDs[0] {
			if i >= limit {
				break
			}
			
			noteID, err := ChromaIDToNoteID(chromaID)
			if err != nil {
				log.Printf("Invalid ChromaDB ID: %s", chromaID)
				continue
			}
			
			var relatedNote models.Note
			if err := ai.db.First(&relatedNote, noteID).Error; err == nil {
				relatedNotes = append(relatedNotes, relatedNote)
			}
		}
	}
	
	return relatedNotes, nil
}

// SearchNotesByEmbedding performs semantic search across all notes
func (ai *AIService) SearchNotesByEmbedding(ctx context.Context, query string, userID uuid.UUID, limit int) ([]models.AIEnhancedNote, error) {
	// Filter by user ID
	where := map[string]interface{}{
		"user_id": userID.String(),
	}
	
	// Query ChromaDB
	results, err := ai.chromaService.QueryByText(ctx, NoteEmbeddingsCollection, []string{query}, limit, where)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}
	
	// Convert results to enhanced notes
	var enhancedNotes []models.AIEnhancedNote
	if len(results.IDs) > 0 && len(results.IDs[0]) > 0 {
		for i, chromaID := range results.IDs[0] {
			noteID, err := ChromaIDToNoteID(chromaID)
			if err != nil {
				continue
			}
			
			var enhancedNote models.AIEnhancedNote
			if err := ai.db.Where("note_id = ?", noteID).First(&enhancedNote).Error; err == nil {
				// Add relevance score from distance
				if len(results.Distances) > 0 && len(results.Distances[0]) > i {
					distance := results.Distances[0][i]
					enhancedNote.AIMetadata["relevance_score"] = 1.0 - distance // Convert distance to similarity
				}
				enhancedNotes = append(enhancedNotes, enhancedNote)
			}
		}
	}
	
	return enhancedNotes, nil
}

// RemoveNoteFromChroma removes a note from the ChromaDB collection
func (ai *AIService) RemoveNoteFromChroma(ctx context.Context, noteID uuid.UUID) error {
	ids := []string{NoteIDToChromaID(noteID)}
	return ai.chromaService.DeleteDocuments(ctx, NoteEmbeddingsCollection, ids)
}

// RefreshChromaCollection rebuilds the entire ChromaDB collection from database
func (ai *AIService) RefreshChromaCollection(ctx context.Context) error {
	log.Println("Starting ChromaDB collection refresh...")
	
	// Delete and recreate the collection
	if err := ai.chromaService.DeleteCollection(ctx, NoteEmbeddingsCollection); err != nil {
		log.Printf("Failed to delete collection (may not exist): %v", err)
	}
	
	// Reinitialize collection
	if err := ai.initializeChromaCollection(ctx); err != nil {
		return fmt.Errorf("failed to reinitialize collection: %w", err)
	}
	
	// Get all notes with AI enhancements
	var notes []models.Note
	if err := ai.db.Find(&notes).Error; err != nil {
		return fmt.Errorf("failed to fetch notes: %w", err)
	}
	
	// Batch process notes
	batchSize := 100
	for i := 0; i < len(notes); i += batchSize {
		end := i + batchSize
		if end > len(notes) {
			end = len(notes)
		}
		
		batch := notes[i:end]
		ids := make([]string, 0, len(batch))
		documents := make([]string, 0, len(batch))
		metadatas := make([]map[string]interface{}, 0, len(batch))
		
		for _, note := range batch {
			// Get enhanced data if exists
			var enhanced models.AIEnhancedNote
			ai.db.Where("note_id = ?", note.ID).First(&enhanced)
			
			// Prepare document
			var docBuilder strings.Builder
			docBuilder.WriteString(note.Title)
			if enhanced.Summary != "" {
				docBuilder.WriteString("\n\nSummary: ")
				docBuilder.WriteString(enhanced.Summary)
			}
			docBuilder.WriteString("\n\n")
			docBuilder.WriteString(ai.extractNoteContent(&note))
			
			document := docBuilder.String()
			if len(document) > MaxDocumentLength {
				document = document[:MaxDocumentLength]
			}
			
			// Prepare metadata
			metadata := map[string]interface{}{
				"note_id":    note.ID.String(),
				"title":      note.Title,
				"created_at": note.CreatedAt.Format(time.RFC3339),
				"updated_at": note.UpdatedAt.Format(time.RFC3339),
				"user_id":    note.UserID.String(),
			}
			
			if note.NotebookID != uuid.Nil {
				metadata["notebook_id"] = note.NotebookID.String()
			}
			
			ids = append(ids, NoteIDToChromaID(note.ID))
			documents = append(documents, document)
			metadatas = append(metadatas, metadata)
		}
		
		// Add batch to ChromaDB
		if err := ai.chromaService.AddDocuments(ctx, NoteEmbeddingsCollection, ids, documents, metadatas); err != nil {
			log.Printf("Failed to add batch %d-%d: %v", i, end, err)
			// Continue with next batch
		}
		
		log.Printf("Processed notes %d-%d of %d", i+1, end, len(notes))
	}
	
	log.Printf("ChromaDB collection refresh completed. Processed %d notes.", len(notes))
	return nil
}

// GetChromaCollectionStats returns statistics about the ChromaDB collection
func (ai *AIService) GetChromaCollectionStats(ctx context.Context) (map[string]interface{}, error) {
	count, err := ai.chromaService.CountDocuments(ctx, NoteEmbeddingsCollection)
	if err != nil {
		return nil, err
	}
	
	stats := map[string]interface{}{
		"collection_name": NoteEmbeddingsCollection,
		"document_count":  count,
		"embedding_model": "all-MiniLM-L6-v2", // ChromaDB default
		"last_updated":    time.Now().Format(time.RFC3339),
	}
	
	return stats, nil
}

// The rest of the methods remain the same as in the original ai_service.go...
// (Include all the other methods like generateTitle, generateSummary, extractTags, etc.)

// extractNoteContent extracts the full content from a note including blocks
func (ai *AIService) extractNoteContent(note *models.Note) string {
	// Get blocks for this note
	var blocks []models.Block
	ai.db.Where("note_id = ?", note.ID).Order("order_index").Find(&blocks)
	
	var contentBuilder strings.Builder
	for _, block := range blocks {
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

// GenerateResponse calls Anthropic's Claude API to generate a response to a prompt
func (ai *AIService) GenerateResponse(ctx context.Context, prompt string, context []string) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Add context if provided
	if len(context) > 0 {
		contextStr := strings.Join(context, "\n")
		messages = []Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Context:\n%s\n\nPrompt:\n%s", contextStr, prompt),
			},
		}
	}

	req := AnthropicRequest{
		Model:     ai.anthropicModel,
		MaxTokens: 4000,
		Messages:  messages,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ai.anthropicKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := ai.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return anthropicResp.Content[0].Text, nil
}

// PerformWebSearch performs a web search using the Perplexica service
func (ai *AIService) PerformWebSearch(ctx context.Context, query string) (interface{}, error) {
	if !ai.perplexicaService.IsEnabled() {
		return nil, fmt.Errorf("perplexica service is not enabled")
	}

	result, err := ai.perplexicaService.WebSearch(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("web search failed: %w", err)
	}

	return result, nil
}

// SearchWithPerplexica performs a search using the Perplexica service
func (ai *AIService) SearchWithPerplexica(ctx context.Context, userID uuid.UUID, query string, focusMode string, context []string) (*PerplexicaSearchResult, error) {
	if !ai.perplexicaService.IsEnabled() {
		return nil, fmt.Errorf("perplexica service is not enabled")
	}

	result, err := ai.perplexicaService.Search(ctx, query, focusMode, "balanced")
	if err != nil {
		return nil, fmt.Errorf("perplexica search failed: %w", err)
	}

	return result, nil
}

// callAnthropic makes a direct call to Anthropic's Claude API
func (ai *AIService) callAnthropic(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if maxTokens == 0 {
		maxTokens = 4000
	}

	messages := []Message{
		{
			Role:    "user", 
			Content: prompt,
		},
	}

	req := AnthropicRequest{
		Model:     ai.anthropicModel,
		MaxTokens: maxTokens,
		Messages:  messages,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ai.anthropicKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := ai.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return anthropicResp.Content[0].Text, nil
}

// generateTitle generates an AI-powered title for note content
func (ai *AIService) generateTitle(ctx context.Context, content string) (string, error) {
	prompt := fmt.Sprintf("Generate a concise, descriptive title for this content. Return only the title, no additional text:\n\n%s", content)
	
	response, err := ai.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(response), nil
}

// generateSummary generates an AI-powered summary for note content
func (ai *AIService) generateSummary(ctx context.Context, content, title string) (string, error) {
	prompt := fmt.Sprintf("Create a concise summary of this content. Focus on key points and main ideas:\n\nTitle: %s\nContent: %s", title, content)
	
	response, err := ai.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(response), nil
}

// extractTags extracts relevant tags from note content
func (ai *AIService) extractTags(ctx context.Context, content, title string) ([]string, error) {
	prompt := fmt.Sprintf("Extract 3-5 relevant tags for this content. Return as a comma-separated list:\n\nTitle: %s\nContent: %s", title, content)
	
	response, err := ai.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return nil, err
	}
	
	// Parse comma-separated tags
	tags := strings.Split(response, ",")
	var cleanTags []string
	for _, tag := range tags {
		cleanTag := strings.TrimSpace(tag)
		if cleanTag != "" {
			cleanTags = append(cleanTags, cleanTag)
		}
	}
	
	return cleanTags, nil
}

// extractActionableSteps extracts actionable steps from note content
func (ai *AIService) extractActionableSteps(ctx context.Context, content, title string) ([]string, error) {
	prompt := fmt.Sprintf("Extract actionable steps or tasks from this content. Return as a numbered list:\n\nTitle: %s\nContent: %s", title, content)
	
	response, err := ai.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return nil, err
	}
	
	// Parse numbered list into slice
	lines := strings.Split(response, "\n")
	var steps []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			steps = append(steps, trimmed)
		}
	}
	
	return steps, nil
}

// extractLearningItems extracts learning opportunities from note content
func (ai *AIService) extractLearningItems(ctx context.Context, content, title string) ([]string, error) {
	prompt := fmt.Sprintf("Extract learning opportunities, insights, or knowledge gaps from this content. Return as a bulleted list:\n\nTitle: %s\nContent: %s", title, content)
	
	response, err := ai.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return nil, err
	}
	
	// Parse bulleted list into slice
	lines := strings.Split(response, "\n")
	var items []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			items = append(items, trimmed)
		}
	}
	
	return items, nil
}

// BreakDownTask breaks down a complex task into smaller actionable steps
func (ai *AIService) BreakDownTask(ctx context.Context, title string, description string, maxSteps int) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`You are a project management AI. Break down the following goal into %d actionable steps.

Goal: %s
Context: %s

Please return ONLY a valid JSON object in this exact format:
{
  "goal": "%s",
  "steps": [
    {
      "step": 1,
      "title": "Step Title",
      "description": "Detailed description of what to do"
    },
    {
      "step": 2,
      "title": "Step Title",
      "description": "Detailed description of what to do"
    }
  ]
}

Return only the JSON, no additional text or formatting.`, maxSteps, title, description, title)
	
	response, err := ai.GenerateResponse(ctx, prompt, nil)
	if err != nil {
		return nil, err
	}
	
	// Clean the response - remove any markdown formatting
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)
	
	// Try to parse as JSON first
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If JSON parsing fails, fall back to manual parsing
		log.Printf("Failed to parse AI response as JSON: %v. Response: %s", err, response)
		
		// Parse the response into steps manually
		lines := strings.Split(response, "\n")
		var steps []map[string]interface{}
		stepNum := 1
		
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.Contains(trimmed, "{") && !strings.Contains(trimmed, "}") {
				steps = append(steps, map[string]interface{}{
					"step":        stepNum,
					"title":       fmt.Sprintf("Step %d", stepNum),
					"description": trimmed,
				})
				stepNum++
				if stepNum > maxSteps {
					break
				}
			}
		}
		
		// Return a properly structured result
		return map[string]interface{}{
			"goal":        title,
			"steps":       steps,
			"max_steps":   maxSteps,
		}, nil
	}
	
	// Validate and ensure proper structure
	if _, exists := result["goal"]; !exists {
		result["goal"] = title
	}
	if _, exists := result["max_steps"]; !exists {
		result["max_steps"] = maxSteps
	}
	
	// Ensure steps is an array of objects, not strings
	if stepsData, exists := result["steps"]; exists {
		if stepsArray, ok := stepsData.([]interface{}); ok {
			var properSteps []map[string]interface{}
			for i, step := range stepsArray {
				if stepMap, ok := step.(map[string]interface{}); ok {
					// Already a proper object
					properSteps = append(properSteps, stepMap)
				} else if stepStr, ok := step.(string); ok {
					// Convert string to proper step object
					properSteps = append(properSteps, map[string]interface{}{
						"step":        i + 1,
						"title":       fmt.Sprintf("Step %d", i+1),
						"description": stepStr,
					})
				}
			}
			result["steps"] = properSteps
		}
	}
	
	return result, nil
}

// CreateProjectNotebook creates a notebook structure for a project using AI
func (ai *AIService) CreateProjectNotebook(ctx context.Context, userID uuid.UUID, projectName, projectDescription string, breakdown map[string]interface{}) (*uuid.UUID, []uuid.UUID, error) {
	// Create the notebook using the notebook service
	notebookData := map[string]interface{}{
		"name":        projectName + " - Project Notebook",
		"description": projectDescription,
		"user_id":     userID.String(),
	}
	
	// Use the notebook service to create the notebook properly with roles
	notebook, err := NotebookServiceInstance.CreateNotebook(&database.Database{DB: ai.db}, notebookData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create project notebook: %w", err)
	}
	
	// Create initial project notes based on the breakdown if provided
	var noteIDs []uuid.UUID
	if breakdown != nil {
		if stepsData, exists := breakdown["steps"]; exists {
			if steps, ok := stepsData.([]interface{}); ok {
				// Create an introduction/overview note first
				introNoteData := map[string]interface{}{
					"title":       projectName + " - Project Overview",
					"user_id":     userID.String(),
					"notebook_id": notebook.ID.String(),
				}
				
				introNote, err := NoteServiceInstance.CreateNote(&database.Database{DB: ai.db}, introNoteData)
				if err != nil {
					log.Printf("Failed to create intro note: %v", err)
				} else {
					noteIDs = append(noteIDs, introNote.ID)
					
					// Create overview content blocks
					titleBlock := models.Block{
						ID:      uuid.New(),
						UserID:  userID,
						NoteID:  introNote.ID,
						Type:    "header",
						Order:   1000.0,
						Content: map[string]interface{}{
							"text":  "Project Overview",
							"level": 1,
						},
						Metadata: models.BlockMetadata{},
					}
					ai.db.Create(&titleBlock)
					
					descBlock := models.Block{
						ID:      uuid.New(),
						UserID:  userID,
						NoteID:  introNote.ID,
						Type:    "text",
						Order:   2000.0,
						Content: map[string]interface{}{
							"text": projectDescription,
						},
						Metadata: models.BlockMetadata{},
					}
					ai.db.Create(&descBlock)
					
					// Create steps overview
					stepsHeaderBlock := models.Block{
						ID:      uuid.New(),
						UserID:  userID,
						NoteID:  introNote.ID,
						Type:    "header",
						Order:   3000.0,
						Content: map[string]interface{}{
							"text":  "Steps Overview",
							"level": 2,
						},
						Metadata: models.BlockMetadata{},
					}
					ai.db.Create(&stepsHeaderBlock)
					
					// Create a bullet list of all steps
					var stepsList strings.Builder
					for i, step := range steps {
						if stepMap, ok := step.(map[string]interface{}); ok {
							stepTitle := "Untitled Step"
							if title, exists := stepMap["title"]; exists {
								if titleStr, ok := title.(string); ok {
									stepTitle = titleStr
								}
							}
							stepsList.WriteString(fmt.Sprintf("• Step %d: %s\n", i+1, stepTitle))
						}
					}
					
					stepsListBlock := models.Block{
						ID:      uuid.New(),
						UserID:  userID,
						NoteID:  introNote.ID,
						Type:    "text",
						Order:   4000.0,
						Content: map[string]interface{}{
							"text": stepsList.String(),
						},
						Metadata: models.BlockMetadata{},
					}
					ai.db.Create(&stepsListBlock)
				}
				
				// Create a separate note for each step
				log.Printf("Creating %d individual step notes", len(steps))
				for i, step := range steps {
					if stepMap, ok := step.(map[string]interface{}); ok {
						stepTitle := "Untitled Step"
						stepDesc := "No description"
						
						if title, exists := stepMap["title"]; exists {
							if titleStr, ok := title.(string); ok {
								stepTitle = titleStr
							}
						}
						
						if desc, exists := stepMap["description"]; exists {
							if descStr, ok := desc.(string); ok {
								stepDesc = descStr
							}
						}
						
						// Create individual note for this step
						stepNoteData := map[string]interface{}{
							"title":       fmt.Sprintf("Step %d: %s", i+1, stepTitle),
							"user_id":     userID.String(),
							"notebook_id": notebook.ID.String(),
						}
						
						stepNote, err := NoteServiceInstance.CreateNote(&database.Database{DB: ai.db}, stepNoteData)
						if err != nil {
							log.Printf("Failed to create step note %d: %v", i+1, err)
							continue
						}
						
						noteIDs = append(noteIDs, stepNote.ID)
						log.Printf("Created step note %d: %s", i+1, stepTitle)
						
						// Add content blocks to the step note
						stepHeaderBlock := models.Block{
							ID:      uuid.New(),
							UserID:  userID,
							NoteID:  stepNote.ID,
							Type:    "header",
							Order:   1000.0,
							Content: map[string]interface{}{
								"text":  fmt.Sprintf("Step %d: %s", i+1, stepTitle),
								"level": 1,
							},
							Metadata: models.BlockMetadata{},
						}
						ai.db.Create(&stepHeaderBlock)
						
						// Add description block
						stepDescBlock := models.Block{
							ID:      uuid.New(),
							UserID:  userID,
							NoteID:  stepNote.ID,
							Type:    "text",
							Order:   2000.0,
							Content: map[string]interface{}{
								"text": stepDesc,
							},
							Metadata: models.BlockMetadata{},
						}
						ai.db.Create(&stepDescBlock)
						
						// Add a status/progress section
						statusHeaderBlock := models.Block{
							ID:      uuid.New(),
							UserID:  userID,
							NoteID:  stepNote.ID,
							Type:    "header",
							Order:   3000.0,
							Content: map[string]interface{}{
								"text":  "Progress & Notes",
								"level": 2,
							},
							Metadata: models.BlockMetadata{},
						}
						ai.db.Create(&statusHeaderBlock)
						
						// Add a placeholder for progress notes
						progressBlock := models.Block{
							ID:      uuid.New(),
							UserID:  userID,
							NoteID:  stepNote.ID,
							Type:    "text",
							Order:   4000.0,
							Content: map[string]interface{}{
								"text": "Status: Not started\n\nAdd your progress notes here...",
							},
							Metadata: models.BlockMetadata{},
						}
						ai.db.Create(&progressBlock)
					}
				}
				log.Printf("Finished creating %d step notes", len(steps))
			}
		}
	}
	
	return &notebook.ID, noteIDs, nil
}